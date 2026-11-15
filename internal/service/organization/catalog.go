// Package organization định nghĩa use case danh mục, độc lập HTTP và PostgreSQL.
package organization

import (
	"context"
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

type CatalogKind string

const (
	Departments CatalogKind = "departments"
	Positions   CatalogKind = "positions"
	JobLevels   CatalogKind = "job-levels"
)

func (k CatalogKind) Valid() bool { return k == Departments || k == Positions || k == JobLevels }

var (
	ErrValidation             = errors.New("organization: invalid input")
	ErrUnauthenticated        = errors.New("organization: unauthenticated")
	ErrPasswordChangeRequired = errors.New("organization: password change required")
	ErrForbidden              = errors.New("organization: forbidden")
	ErrNotFound               = errors.New("organization: not found")
	ErrStale                  = errors.New("organization: stale data")
	ErrDuplicateName          = errors.New("organization: duplicate name")
	ErrCatalogInUse           = errors.New("organization: catalog in use")
	ErrUnavailable            = errors.New("organization: storage unavailable")
	ErrPersistence            = errors.New("organization: storage failure")
	// Commit có thể đã tới DB: caller không tự gửi lại mutation.
	ErrCommitUnknown = errors.New("organization: commit outcome unknown")
)

type CatalogItem struct {
	ID   int
	Name string
}
type CatalogPage struct {
	Items    []CatalogItem
	Page     int64
	PageSize int
	Total    int64
}
type ListCatalogQuery struct {
	ActorID  int
	Kind     CatalogKind
	Page     int64
	PageSize int
	Query    string
}

func (q ListCatalogQuery) Offset() int64 { return (q.Page - 1) * int64(q.PageSize) }

type CatalogCommand struct {
	ActorID            int
	Kind               CatalogKind
	ID                 int
	Name               string
	ExpectedBeforeName string
}

// Adapter bảo đảm kiểm quyền hiện hành và audit cùng transaction với mutation.
type CatalogRepository interface {
	ListCatalog(context.Context, ListCatalogQuery) (CatalogPage, error)
	CreateCatalog(context.Context, CatalogCommand) (CatalogItem, error)
	RenameCatalog(context.Context, CatalogCommand) (CatalogItem, error)
	DeleteCatalog(context.Context, CatalogCommand) error
}
type CatalogService struct{ repo CatalogRepository }

func NewCatalogService(repo CatalogRepository) *CatalogService { return &CatalogService{repo: repo} }

func NormalizeCatalogName(name string) (string, error) {
	if !utf8.ValidString(name) || strings.ContainsRune(name, 0) {
		return "", ErrValidation
	}
	name = strings.TrimSpace(name)
	if n := utf8.RuneCountInString(name); n < 1 || n > 100 {
		return "", ErrValidation
	}
	return name, nil
}
func ValidateListCatalog(q ListCatalogQuery) (ListCatalogQuery, error) {
	if q.ActorID < 1 {
		return q, ErrUnauthenticated
	}
	if !q.Kind.Valid() || q.Page < 1 || q.Page > math.MaxInt32 || q.PageSize < 1 || q.PageSize > 100 || (q.Page-1) > math.MaxInt64/int64(q.PageSize) {
		return q, ErrValidation
	}
	if !utf8.ValidString(q.Query) || strings.ContainsRune(q.Query, 0) {
		return q, ErrValidation
	}
	q.Query = strings.TrimSpace(q.Query)
	if utf8.RuneCountInString(q.Query) > 100 {
		return q, ErrValidation
	}
	return q, nil
}
func ValidateCatalogCommand(c CatalogCommand, existing bool, name bool) (CatalogCommand, error) {
	if c.ActorID < 1 {
		return c, ErrUnauthenticated
	}
	if !c.Kind.Valid() {
		return c, ErrValidation
	}
	if existing && (c.ID < 1 || int64(c.ID) > math.MaxInt32 || !utf8.ValidString(c.ExpectedBeforeName) || strings.ContainsRune(c.ExpectedBeforeName, 0) || utf8.RuneCountInString(c.ExpectedBeforeName) < 1 || utf8.RuneCountInString(c.ExpectedBeforeName) > 100) {
		return c, ErrValidation
	}
	if name {
		var err error
		c.Name, err = NormalizeCatalogName(c.Name)
		if err != nil {
			return c, err
		}
	}
	return c, nil
}
func (s *CatalogService) List(ctx context.Context, q ListCatalogQuery) (CatalogPage, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	q, err := ValidateListCatalog(q)
	if err != nil {
		return CatalogPage{}, err
	}
	return s.repo.ListCatalog(ctx, q)
}
func (s *CatalogService) Create(ctx context.Context, c CatalogCommand) (CatalogItem, error) {
	c, err := ValidateCatalogCommand(c, false, true)
	if err != nil {
		return CatalogItem{}, err
	}
	return s.repo.CreateCatalog(ctx, c)
}
func (s *CatalogService) Rename(ctx context.Context, c CatalogCommand) (CatalogItem, error) {
	c, err := ValidateCatalogCommand(c, true, true)
	if err != nil {
		return CatalogItem{}, err
	}
	return s.repo.RenameCatalog(ctx, c)
}
func (s *CatalogService) Delete(ctx context.Context, c CatalogCommand) error {
	c, err := ValidateCatalogCommand(c, true, false)
	if err != nil {
		return err
	}
	return s.repo.DeleteCatalog(ctx, c)
}
