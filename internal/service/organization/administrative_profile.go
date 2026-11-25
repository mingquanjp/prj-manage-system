package organization

import (
	"context"
	"errors"
	"math"
	"net/mail"
	"strings"
	"unicode/utf8"

	"ai-staffing/internal/service/auth"
)

var ErrCatalogNotFound = errors.New("organization: catalog not found")

// AdministrativeField phân biệt không gửi trường với chủ ý xóa bằng null.
type AdministrativeField[T any] struct {
	Present bool
	Value   *T
}

type AdministrativeValues struct {
	Fullname, CompanyEmail, Phone, Address AdministrativeField[string]
	PositionID, DepartmentID, JobLevelID   AdministrativeField[int]
}

type AdministrativeProfile = SelfProfile
type AdministrativeProfileQuery struct{ ActorID, EmployeeID int }
type EditAdministrativeProfileCommand struct {
	ActorID, EmployeeID           int
	ExpectedBeforeValues, Changes AdministrativeValues
}
type ListAdministrativeEmployeesQuery struct {
	ActorID  int
	Page     int64
	PageSize int
	Query    string
}

func (q ListAdministrativeEmployeesQuery) Offset() int64 { return (q.Page - 1) * int64(q.PageSize) }

// EmployeeSummary không chứa phone/address hoặc thông tin quyền tài khoản.
type EmployeeSummary struct {
	EmployeeID                           int
	EmployeeCode, Fullname, CompanyEmail string
	Position                             CatalogItem
	Department, JobLevel                 *CatalogItem
	EmploymentStatus                     string
}
type AdministrativeEmployeePage struct {
	Items    []EmployeeSummary
	Page     int64
	PageSize int
	Total    int64
}

// Adapter kiểm quyền hiện hành và bảo đảm update/audit/revoke cùng transaction.
type AdministrativeProfileRepository interface {
	ListAdministrativeEmployees(context.Context, ListAdministrativeEmployeesQuery) (AdministrativeEmployeePage, error)
	ReadAdministrativeProfile(context.Context, AdministrativeProfileQuery) (AdministrativeProfile, error)
	EditAdministrativeProfile(context.Context, EditAdministrativeProfileCommand) (AdministrativeProfile, error)
}
type AdministrativeProfileService struct {
	repo AdministrativeProfileRepository
}

func NewAdministrativeProfileService(repo AdministrativeProfileRepository) *AdministrativeProfileService {
	return &AdministrativeProfileService{repo: repo}
}

func ValidateAdministrativeProfileQuery(q AdministrativeProfileQuery) error {
	return ValidateSelfProfileQuery(SelfProfileQuery{ActorID: q.ActorID, EmployeeID: q.EmployeeID})
}
func ValidateListAdministrativeEmployees(q ListAdministrativeEmployeesQuery) (ListAdministrativeEmployeesQuery, error) {
	if q.ActorID < 1 || int64(q.ActorID) > math.MaxInt32 {
		return q, ErrUnauthenticated
	}
	if q.Page < 1 || q.Page > math.MaxInt32 || q.PageSize < 1 || q.PageSize > 100 || !administrativeText(q.Query) {
		return q, ErrValidation
	}
	q.Query = strings.TrimSpace(q.Query)
	if utf8.RuneCountInString(q.Query) > 100 {
		return q, ErrValidation
	}
	return q, nil
}

func ValidateEditAdministrativeProfile(c EditAdministrativeProfileCommand) (EditAdministrativeProfileCommand, error) {
	if err := ValidateAdministrativeProfileQuery(AdministrativeProfileQuery{c.ActorID, c.EmployeeID}); err != nil {
		return c, err
	}
	count := 0
	for _, field := range []struct {
		before *AdministrativeField[string]
		after  *AdministrativeField[string]
		null   bool
		limit  int
		email  bool
	}{{&c.ExpectedBeforeValues.Fullname, &c.Changes.Fullname, false, 100, false},
		{&c.ExpectedBeforeValues.CompanyEmail, &c.Changes.CompanyEmail, false, 100, true},
		{&c.ExpectedBeforeValues.Phone, &c.Changes.Phone, true, 20, false},
		{&c.ExpectedBeforeValues.Address, &c.Changes.Address, false, 2000, false}} {
		if field.before.Present != field.after.Present || (!field.before.Present && (field.before.Value != nil || field.after.Value != nil)) {
			return c, ErrValidation
		}
		if !field.before.Present {
			continue
		}
		count++
		// Before giữ raw/NULL đã đọc, kể cả legacy dài hơn giới hạn nhập mới.
		if field.before.Value == nil {
			if field.before == &c.ExpectedBeforeValues.Fullname || field.email {
				return c, ErrValidation
			}
		} else {
			value := *field.before.Value
			if !administrativeText(value) || (field.before == &c.ExpectedBeforeValues.Fullname && value == "") {
				return c, ErrValidation
			}
			field.before.Value = &value
		}
		if field.after.Value == nil {
			if !field.null {
				return c, ErrValidation
			}
			continue
		}
		value := *field.after.Value
		if !administrativeText(value) || (field.email && strings.ContainsAny(value, "\r\n")) {
			return c, ErrValidation
		}
		value = strings.TrimSpace(value)
		if field.email {
			value = auth.NormalizeLoginEmail(value)
		}
		if n := utf8.RuneCountInString(value); n < 1 || n > field.limit {
			return c, ErrValidation
		}
		if field.email && !validAdministrativeEmail(value) {
			return c, ErrValidation
		}
		field.after.Value = &value
	}
	for _, field := range []struct {
		before, after *AdministrativeField[int]
		null          bool
	}{{&c.ExpectedBeforeValues.PositionID, &c.Changes.PositionID, false},
		{&c.ExpectedBeforeValues.DepartmentID, &c.Changes.DepartmentID, true},
		{&c.ExpectedBeforeValues.JobLevelID, &c.Changes.JobLevelID, true}} {
		if field.before.Present != field.after.Present || (!field.before.Present && (field.before.Value != nil || field.after.Value != nil)) {
			return c, ErrValidation
		}
		if !field.before.Present {
			continue
		}
		count++
		for _, f := range []*AdministrativeField[int]{field.before, field.after} {
			if f.Value == nil {
				if !field.null {
					return c, ErrValidation
				}
				continue
			}
			value := *f.Value
			if value < 1 || int64(value) > math.MaxInt32 {
				return c, ErrValidation
			}
			f.Value = &value
		}
	}
	if count == 0 {
		return c, ErrValidation
	}
	return c, nil
}

func (s *AdministrativeProfileService) List(ctx context.Context, q ListAdministrativeEmployeesQuery) (AdministrativeEmployeePage, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	q, err := ValidateListAdministrativeEmployees(q)
	if err != nil {
		return AdministrativeEmployeePage{}, err
	}
	if s.repo == nil {
		return AdministrativeEmployeePage{}, ErrPersistence
	}
	p, err := s.repo.ListAdministrativeEmployees(ctx, q)
	if err != nil {
		return AdministrativeEmployeePage{}, safeAdministrativeError(err)
	}
	if p.Page != q.Page || p.PageSize != q.PageSize || p.Total < 0 || len(p.Items) > p.PageSize || int64(len(p.Items)) > p.Total {
		return AdministrativeEmployeePage{}, ErrPersistence
	}
	if p.Items == nil {
		p.Items = []EmployeeSummary{}
	}
	for i := range p.Items {
		v := &p.Items[i]
		v.CompanyEmail = auth.NormalizeLoginEmail(v.CompanyEmail)
		profile := AdministrativeProfile{UserID: 1, EmployeeID: v.EmployeeID, EmployeeCode: v.EmployeeCode, Fullname: v.Fullname, CompanyEmail: v.CompanyEmail, Position: v.Position, Department: v.Department, JobLevel: v.JobLevel, EmploymentStatus: v.EmploymentStatus}
		if !validAdministrativeProjection(profile, v.EmployeeID) {
			return AdministrativeEmployeePage{}, ErrPersistence
		}
	}
	return p, nil
}

func (s *AdministrativeProfileService) Read(ctx context.Context, q AdministrativeProfileQuery) (AdministrativeProfile, error) {
	if err := ValidateAdministrativeProfileQuery(q); err != nil {
		return AdministrativeProfile{}, err
	}
	if s.repo == nil {
		return AdministrativeProfile{}, ErrPersistence
	}
	p, err := s.repo.ReadAdministrativeProfile(ctx, q)
	return administrativeResult(p, q.EmployeeID, err)
}
func (s *AdministrativeProfileService) Edit(ctx context.Context, c EditAdministrativeProfileCommand) (AdministrativeProfile, error) {
	c, err := ValidateEditAdministrativeProfile(c)
	if err != nil {
		return AdministrativeProfile{}, err
	}
	if s.repo == nil {
		return AdministrativeProfile{}, ErrPersistence
	}
	p, err := s.repo.EditAdministrativeProfile(ctx, c)
	return administrativeResult(p, c.EmployeeID, err)
}
func administrativeResult(p AdministrativeProfile, employeeID int, err error) (AdministrativeProfile, error) {
	if err != nil {
		return AdministrativeProfile{}, safeAdministrativeError(err)
	}
	p.CompanyEmail = auth.NormalizeLoginEmail(p.CompanyEmail)
	if !validAdministrativeProjection(p, employeeID) {
		return AdministrativeProfile{}, ErrPersistence
	}
	return p, nil
}
func administrativeText(v string) bool { return utf8.ValidString(v) && !strings.ContainsRune(v, 0) }
func validAdministrativeEmail(v string) bool {
	a, err := mail.ParseAddress(v)
	return err == nil && a.Name == "" && a.Address == v && validProfileText(v, 3, 100) && !strings.ContainsAny(v, "\r\n")
}
func validAdministrativeProjection(p AdministrativeProfile, id int) bool {
	if p.EmployeeID != id || id < 1 || int64(id) > math.MaxInt32 || p.UserID < 1 || int64(p.UserID) > math.MaxInt32 || !validProfileText(p.EmployeeCode, 1, 30) || !validProfileText(p.Fullname, 1, 100) || !validAdministrativeEmail(p.CompanyEmail) {
		return false
	}
	if p.EmploymentStatus != "ACTIVE" && p.EmploymentStatus != "INACTIVE" && p.EmploymentStatus != "RESIGNED" {
		return false
	}
	for _, item := range []*CatalogItem{&p.Position, p.Department, p.JobLevel} {
		if item != nil && (item.ID < 1 || int64(item.ID) > math.MaxInt32 || !validProfileText(item.Name, 1, 100)) {
			return false
		}
	}
	return (p.Phone == nil || administrativeText(*p.Phone)) && (p.Address == nil || administrativeText(*p.Address))
}
func safeAdministrativeError(err error) error {
	if errors.Is(err, ErrCommitUnknown) {
		if errors.Is(err, ErrUnavailable) {
			return errors.Join(ErrCommitUnknown, ErrUnavailable)
		}
		return errors.Join(ErrCommitUnknown, ErrPersistence)
	}
	for _, kind := range []error{ErrValidation, ErrUnauthenticated, ErrPasswordChangeRequired, ErrForbidden, ErrNotFound, ErrStale, ErrCatalogNotFound, ErrEmailConflict, ErrUnavailable, ErrPersistence, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, kind) {
			return kind
		}
	}
	return ErrPersistence
}
