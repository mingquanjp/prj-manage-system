// Package skillcatalog cung cấp danh mục chọn kỹ năng và quyền quản trị Giám đốc.
package skillcatalog

import (
	"context"
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

// Kind giới hạn danh mục ở nhóm kỹ năng và kỹ năng.
type Kind string

const (
	Categories Kind = "skill-categories"
	Skills     Kind = "skills"
)

var (
	ErrValidation             = errors.New("skill catalog: invalid input")
	ErrUnauthenticated        = errors.New("skill catalog: unauthenticated")
	ErrPasswordChangeRequired = errors.New("skill catalog: password change required")
	ErrForbidden              = errors.New("skill catalog: forbidden")
	ErrNotFound               = errors.New("skill catalog: not found")
	ErrStale                  = errors.New("skill catalog: stale data")
	ErrDuplicateName          = errors.New("skill catalog: duplicate name")
	ErrCatalogInUse           = errors.New("skill catalog: catalog in use")
	ErrUnavailable            = errors.New("skill catalog: unavailable")
	ErrPersistence            = errors.New("skill catalog: persistence")
	// ErrCommitUnknown yêu cầu caller đối soát GET trước thao tác tiếp, không tự gửi lại.
	ErrCommitUnknown = errors.New("skill catalog: commit outcome unknown")
)

// Category là nhận diện nhóm hiện hành, không mang quyền hoặc trạng thái nghiệp vụ.
type Category struct {
	ID   int
	Name string
}

// Item chứa mục danh mục; Category chỉ có giá trị khi Kind là Skills.
type Item struct {
	ID       int
	Name     string
	Category Category
}

// Page giữ count và items của cùng snapshot, kể cả khi trang không có mục.
type Page struct {
	Items          []Item
	Page, PageSize int
	Total          int64
}

// ListQuery lấy ActorID từ phiên; CategoryID bằng 0 nghĩa không lọc nhóm.
type ListQuery struct {
	ActorID                    int
	Kind                       Kind
	Query                      string
	CategoryID, Page, PageSize int
}

// Offset dùng số nguyên 64 bit để trang lớn không làm tràn phép nhân.
func (q ListQuery) Offset() int64 { return (int64(q.Page) - 1) * int64(q.PageSize) }

// Before giữ nguyên tên và nhóm đã đọc, không normalize lại khi so stale.
type Before struct {
	Name       string
	CategoryID int
}

// Command lấy ActorID từ phiên và giữ riêng nhóm đích/ExpectedBefore khi chuyển nhóm.
// Caller không gửi lại mutation khi commit chưa xác định.
type Command struct {
	ActorID        int
	Kind           Kind
	ID             int
	Name           string
	CategoryID     int
	ExpectedBefore Before
}

// Repository kiểm quyền hiện hành và ghi audit cùng transaction với mutation.
// List đọc actor, count và items trong cùng snapshot; không suy quyền từ principal client.
type Repository interface {
	List(context.Context, ListQuery) (Page, error)
	Create(context.Context, Command) (Item, error)
	Update(context.Context, Command) (Item, error)
	Delete(context.Context, Command) error
}

// Service kiểm input và điều phối danh mục; repository quyết định quyền hiện hành.
type Service struct{ repo Repository }

// NewService nhận adapter chịu trách nhiệm quyền hiện hành và transaction qua Repository.
func NewService(repo Repository) *Service { return &Service{repo: repo} }
func validID(id int) bool                 { return id > 0 && int64(id) <= math.MaxInt32 }
func validText(s string) bool             { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }
func validKind(k Kind) bool               { return k == Categories || k == Skills }

// ValidateList chuẩn hóa query, giữ category_id chỉ cho danh sách kỹ năng.
func ValidateList(q ListQuery) (ListQuery, error) {
	if !validID(q.ActorID) {
		return q, ErrUnauthenticated
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	if !validKind(q.Kind) || !validID(q.Page) || q.PageSize < 1 || q.PageSize > 100 || q.CategoryID < 0 || int64(q.CategoryID) > math.MaxInt32 || (q.Kind == Categories && q.CategoryID != 0) || !validText(q.Query) {
		return q, ErrValidation
	}
	q.Query = strings.TrimSpace(q.Query)
	if utf8.RuneCountInString(q.Query) > 100 {
		return q, ErrValidation
	}
	return q, nil
}

// ValidateCommand kiểm ID/before khi existing (UPDATE/DELETE) và tên/nhóm đích khi named (CREATE/UPDATE).
// Chỉ tên đích được normalize; before giữ nguyên để phát hiện màn hình cũ.
func ValidateCommand(c Command, existing, named bool) (Command, error) {
	if !validID(c.ActorID) {
		return c, ErrUnauthenticated
	}
	if !validKind(c.Kind) {
		return c, ErrValidation
	}
	if existing && (!validID(c.ID) || !validText(c.ExpectedBefore.Name) || (c.Kind == Skills && !validID(c.ExpectedBefore.CategoryID)) || (c.Kind == Categories && c.ExpectedBefore.CategoryID != 0)) {
		return c, ErrValidation
	}
	if named {
		if !validText(c.Name) {
			return c, ErrValidation
		}
		c.Name = strings.TrimSpace(c.Name)
		if n := utf8.RuneCountInString(c.Name); n < 1 || n > 100 {
			return c, ErrValidation
		}
		if (c.Kind == Skills && !validID(c.CategoryID)) || (c.Kind == Categories && c.CategoryID != 0) {
			return c, ErrValidation
		}
	}
	return c, nil
}

// List đọc danh mục cho mọi phiên thường; không yêu cầu vai trò quản trị.
func (s *Service) List(ctx context.Context, q ListQuery) (Page, error) {
	q, e := ValidateList(q)
	if e != nil {
		return Page{}, e
	}
	return s.repo.List(ctx, q)
}

// Create trả mục sau commit đã xác nhận; chỉ Giám đốc được ghi.
func (s *Service) Create(ctx context.Context, c Command) (Item, error) {
	c, e := ValidateCommand(c, false, true)
	if e != nil {
		return Item{}, e
	}
	return s.repo.Create(ctx, c)
}

// Update giữ ID và so ExpectedBefore; no-op không sinh audit.
func (s *Service) Update(ctx context.Context, c Command) (Item, error) {
	c, e := ValidateCommand(c, true, true)
	if e != nil {
		return Item{}, e
	}
	return s.repo.Update(ctx, c)
}

// Delete chỉ xóa khi ExpectedBefore khớp và không còn tham chiếu.
func (s *Service) Delete(ctx context.Context, c Command) error {
	c, e := ValidateCommand(c, true, false)
	if e != nil {
		return e
	}
	return s.repo.Delete(ctx, c)
}
