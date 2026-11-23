// Package profilechange định nghĩa use case đề nghị hồ sơ, độc lập HTTP/DB.
package profilechange

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrValidation             = errors.New("profilechange: invalid input")
	ErrUnauthenticated        = errors.New("profilechange: unauthenticated")
	ErrPasswordChangeRequired = errors.New("profilechange: password change required")
	ErrForbidden              = errors.New("profilechange: forbidden")
	ErrNotFound               = errors.New("profilechange: not found")
	ErrStale                  = errors.New("profilechange: stale data")
	ErrAlreadyResolved        = errors.New("profilechange: request already resolved")
	ErrEmployeeInactive       = errors.New("profilechange: employee inactive")
	ErrKeyReused              = errors.New("profilechange: idempotency key reused")
	ErrKeyExpired             = errors.New("profilechange: idempotency key expired")
	ErrPersistence            = errors.New("profilechange: storage failure")
	ErrUnavailable            = errors.New("profilechange: storage unavailable")
	ErrCommitUnknown          = errors.New("profilechange: commit outcome unknown")
)

const SubmitOperation = "PROFILE_CHANGE_SUBMIT"
const ReplayWindow = 24 * time.Hour

// Values giữ cả sự hiện diện của khóa và NULL; không dùng omitempty để làm mất NULL.
type Values map[string]*string

func (v Values) Fields() []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type SubmitInput struct {
	ActorID, EmployeeID int
	Key                 string
	Before, Proposed    Values
	Explanation         *string
}
type SubmitCommand struct {
	Input SubmitInput
	Hash  [32]byte
}

func (SubmitCommand) String() string               { return "[profile submission redacted]" }
func (c SubmitCommand) GoString() string           { return c.String() }
func (SubmitCommand) LogValue() slog.Value         { return slog.StringValue("[profile submission redacted]") }
func (SubmitCommand) MarshalJSON() ([]byte, error) { return nil, ErrPersistence }

type Summary struct {
	RequestID, EmployeeID          int
	EmployeeCode, Fullname, Status string
	ChangedFields                  []string
	RequestedAt                    time.Time
	ResolvedAt                     *time.Time
}
type Detail struct {
	Summary
	Before, Proposed, Current Values
	Explanation               *string
	RequestedBy               int
	ResolvedBy                *int
	DecisionReason            *string
}
type ListQuery struct {
	ActorID       int
	Scope, Status string
	Page          int64
	PageSize      int
}

func (q ListQuery) Offset() int64 { return (q.Page - 1) * int64(q.PageSize) }

type Page struct {
	Items    []Summary
	Total    int64
	Page     int64
	PageSize int
}
type DetailQuery struct{ ActorID, RequestID int }

// ActorID chỉ nhận từ nguồn nội bộ đã xác thực. Middleware auth tại T05 kiểm
// session; adapter vẫn recheck trạng thái/cờ/quyền/liên kết hiện hành trong DB.
type Repository interface {
	Submit(context.Context, SubmitCommand) (Detail, error)
	List(context.Context, ListQuery) (Page, error)
	Read(context.Context, DetailQuery) (Detail, error)
	Approve(context.Context, DetailQuery) (Detail, error)
	Reject(context.Context, RejectInput) (Detail, error)
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Submit(ctx context.Context, in SubmitInput) (Detail, error) {
	cmd, err := PrepareSubmit(in)
	if err != nil {
		return Detail{}, err
	}
	d, err := s.repo.Submit(ctx, cmd)
	if err != nil {
		return Detail{}, err
	}
	if ValidateDetail(d) != nil || d.EmployeeID != in.EmployeeID || d.RequestedBy != in.ActorID {
		return Detail{}, ErrPersistence
	}
	return d, nil
}
func (s *Service) List(ctx context.Context, q ListQuery) (Page, error) {
	q, err := ValidateList(q)
	if err != nil {
		return Page{}, err
	}
	p, err := s.repo.List(ctx, q)
	if err != nil {
		return Page{}, err
	}
	if p.Items == nil || p.Total < 0 || p.Page != q.Page || p.PageSize != q.PageSize || len(p.Items) > q.PageSize || int64(len(p.Items)) > p.Total {
		return Page{}, ErrPersistence
	}
	for _, item := range p.Items {
		if ValidateSummary(item) != nil {
			return Page{}, ErrPersistence
		}
	}
	return p, nil
}
func (s *Service) Read(ctx context.Context, q DetailQuery) (Detail, error) {
	if err := ValidateDetailQuery(q); err != nil {
		return Detail{}, err
	}
	d, err := s.repo.Read(ctx, q)
	if err != nil {
		return Detail{}, err
	}
	if ValidateDetail(d) != nil || d.RequestID != q.RequestID {
		return Detail{}, ErrPersistence
	}
	return d, nil
}

var keyPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89aAbB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func validID(id int) bool     { return id > 0 && int64(id) <= math.MaxInt32 }
func validText(s string) bool { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }
func cloneText(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}
func NormalizeValues(v Values, proposed bool) (Values, error) {
	if len(v) < 1 || len(v) > 3 {
		return nil, ErrValidation
	}
	out := Values{}
	for k, p := range v {
		limit := 0
		switch k {
		case "fullname":
			limit = 100
		case "phone":
			limit = 20
		case "address":
			limit = 2000
		default:
			return nil, ErrValidation
		}
		if p == nil {
			if k == "fullname" || (proposed && k == "address") {
				return nil, ErrValidation
			}
			out[k] = nil
			continue
		}
		value := *p
		if !validText(value) {
			return nil, ErrValidation
		}
		if proposed {
			value = strings.TrimSpace(value)
			n := utf8.RuneCountInString(value)
			if n < 1 || n > limit {
				return nil, ErrValidation
			}
		} else if k == "fullname" && value == "" {
			return nil, ErrValidation
		}
		out[k] = &value
	}
	return out, nil
}
func PrepareSubmit(in SubmitInput) (SubmitCommand, error) {
	if !validID(in.ActorID) {
		return SubmitCommand{}, ErrUnauthenticated
	}
	if !validID(in.EmployeeID) || !keyPattern.MatchString(in.Key) {
		return SubmitCommand{}, ErrValidation
	}
	before, err := NormalizeValues(in.Before, false)
	if err != nil {
		return SubmitCommand{}, err
	}
	proposed, err := NormalizeValues(in.Proposed, true)
	if err != nil {
		return SubmitCommand{}, err
	}
	if !reflect.DeepEqual(before.Fields(), proposed.Fields()) || reflect.DeepEqual(before, proposed) {
		return SubmitCommand{}, ErrValidation
	}
	in.Before, in.Proposed, in.Key = before, proposed, strings.ToLower(in.Key)
	in.Explanation = cloneText(in.Explanation)
	if in.Explanation != nil {
		v := *in.Explanation
		if !validText(v) {
			return SubmitCommand{}, ErrValidation
		}
		v = strings.TrimSpace(v)
		if utf8.RuneCountInString(v) > 2000 {
			return SubmitCommand{}, ErrValidation
		}
		if v == "" {
			in.Explanation = nil
		} else {
			in.Explanation = &v
		}
	}
	// encoding/json sắp khóa map; tuple có version, ID đích và NULL tường minh.
	tuple := struct {
		Version          int
		EmployeeID       int
		Before, Proposed Values
		Explanation      *string
	}{1, in.EmployeeID, before, proposed, in.Explanation}
	payload, err := json.Marshal(tuple)
	if err != nil {
		return SubmitCommand{}, ErrPersistence
	}
	return SubmitCommand{in, sha256.Sum256(payload)}, nil
}
func ValidateList(q ListQuery) (ListQuery, error) {
	if !validID(q.ActorID) {
		return q, ErrUnauthenticated
	}
	if q.Scope == "" {
		q.Scope = "self"
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	if (q.Scope != "self" && q.Scope != "review") || (q.Status != "" && !validStatus(q.Status)) || q.Page < 1 || q.Page > math.MaxInt32 || q.PageSize < 1 || q.PageSize > 100 {
		return q, ErrValidation
	}
	return q, nil
}
func ValidateDetailQuery(q DetailQuery) error {
	if !validID(q.ActorID) {
		return ErrUnauthenticated
	}
	if !validID(q.RequestID) {
		return ErrValidation
	}
	return nil
}
func validStatus(v string) bool { return v == "PENDING" || v == "APPROVED" || v == "REJECTED" }
func ValidateSummary(v Summary) error {
	if !validID(v.RequestID) || !validID(v.EmployeeID) || !validStatus(v.Status) || !validText(v.EmployeeCode) || len(v.EmployeeCode) == 0 || utf8.RuneCountInString(v.EmployeeCode) > 30 || !validText(v.Fullname) || len(v.Fullname) == 0 || utf8.RuneCountInString(v.Fullname) > 100 || v.RequestedAt.IsZero() {
		return ErrPersistence
	}
	if len(v.ChangedFields) < 1 || len(v.ChangedFields) > 3 {
		return ErrPersistence
	}
	seen := map[string]bool{}
	for _, k := range v.ChangedFields {
		if (k != "fullname" && k != "phone" && k != "address") || seen[k] {
			return ErrPersistence
		}
		seen[k] = true
	}
	if v.Status == "PENDING" {
		if v.ResolvedAt != nil {
			return ErrPersistence
		}
	} else if v.ResolvedAt == nil || v.ResolvedAt.Before(v.RequestedAt) {
		return ErrPersistence
	}
	return nil
}
func ValidateDetail(d Detail) error {
	if ValidateSummary(d.Summary) != nil || !validID(d.RequestedBy) {
		return ErrPersistence
	}
	before, e1 := NormalizeValues(d.Before, false)
	proposed, e2 := NormalizeValues(d.Proposed, true)
	current, e3 := NormalizeValues(d.Current, false)
	if e1 != nil || e2 != nil || e3 != nil || !reflect.DeepEqual(proposed, d.Proposed) || !reflect.DeepEqual(before.Fields(), proposed.Fields()) || !reflect.DeepEqual(before.Fields(), current.Fields()) || !reflect.DeepEqual(before.Fields(), d.ChangedFields) || reflect.DeepEqual(before, proposed) {
		return ErrPersistence
	}
	if d.Explanation != nil && (!validText(*d.Explanation) || utf8.RuneCountInString(*d.Explanation) > 2000) {
		return ErrPersistence
	}
	if d.Status == "PENDING" {
		if d.ResolvedBy != nil || d.DecisionReason != nil {
			return ErrPersistence
		}
	} else {
		if d.ResolvedBy == nil || !validID(*d.ResolvedBy) {
			return ErrPersistence
		}
		if d.Status == "APPROVED" && d.DecisionReason != nil {
			return ErrPersistence
		}
		if d.Status == "REJECTED" && (d.DecisionReason == nil || !validText(*d.DecisionReason) || strings.TrimSpace(*d.DecisionReason) == "" || utf8.RuneCountInString(*d.DecisionReason) > 2000) {
			return ErrPersistence
		}
	}
	return nil
}

// EncodeStoredValues/DecodeStoredValues chỉ dùng nội bộ; transport phải dựng DTO
// từ Values đã validate, không xuất schema_version hoặc JSONB thô.
func EncodeStoredValues(v Values) ([]byte, error) {
	m := map[string]any{"schema_version": 1}
	for k, x := range v {
		m[k] = x
	}
	b, e := json.Marshal(m)
	if e != nil {
		return nil, ErrPersistence
	}
	return b, nil
}
func DecodeStoredValues(b []byte, proposed bool) (Values, error) {
	var raw map[string]json.RawMessage
	if !utf8.Valid(b) || json.Unmarshal(b, &raw) != nil || raw == nil {
		return nil, ErrPersistence
	}
	var version int
	if json.Unmarshal(raw["schema_version"], &version) != nil || version != 1 {
		return nil, ErrPersistence
	}
	delete(raw, "schema_version")
	v := Values{}
	for k, b := range raw {
		var x *string
		if json.Unmarshal(b, &x) != nil {
			return nil, ErrPersistence
		}
		v[k] = x
	}
	out, err := NormalizeValues(v, proposed)
	if err != nil || !reflect.DeepEqual(out, v) {
		return nil, ErrPersistence
	}
	return out, nil
}
