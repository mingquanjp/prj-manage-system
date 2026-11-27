package organization

import (
	"context"
	"math"
	"net/mail"
	"strings"
	"unicode/utf8"

	"ai-staffing/internal/service/auth"
)

type ListAccountsQuery struct {
	ActorID  int
	Page     int64
	PageSize int
	Query    string
}

func (q ListAccountsQuery) Offset() int64 { return (q.Page - 1) * int64(q.PageSize) }

type AccountPage struct {
	Items    []AccountProjection
	Page     int64
	PageSize int
	Total    int64
}

type AccountReadRepository interface {
	ListAccounts(context.Context, ListAccountsQuery) (AccountPage, error)
	PreviewAccountCode(context.Context, int) (string, error)
}

type AccountReadService struct{ repo AccountReadRepository }

func NewAccountReadService(repo AccountReadRepository) *AccountReadService {
	return &AccountReadService{repo: repo}
}

func ValidateListAccounts(q ListAccountsQuery) (ListAccountsQuery, error) {
	if q.ActorID < 1 || int64(q.ActorID) > math.MaxInt32 {
		return q, ErrUnauthenticated
	}
	if q.Page < 1 || q.Page > math.MaxInt32 || q.PageSize < 1 || q.PageSize > 100 {
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

func (s *AccountReadService) List(ctx context.Context, q ListAccountsQuery) (AccountPage, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	q, err := ValidateListAccounts(q)
	if err != nil {
		return AccountPage{}, err
	}
	result, err := s.repo.ListAccounts(ctx, q)
	if err != nil {
		return AccountPage{}, err
	}
	for i := range result.Items {
		result.Items[i].CompanyEmail = auth.NormalizeLoginEmail(result.Items[i].CompanyEmail)
		if !validAccountProjection(result.Items[i]) {
			return AccountPage{}, ErrPersistence
		}
	}
	return result, nil
}

func (s *AccountReadService) Preview(ctx context.Context, actorID int) (string, error) {
	if actorID < 1 || int64(actorID) > math.MaxInt32 {
		return "", ErrUnauthenticated
	}
	return s.repo.PreviewAccountCode(ctx, actorID)
}

func validAccountProjection(p AccountProjection) bool {
	if p.UserID < 1 || int64(p.UserID) > math.MaxInt32 || p.EmployeeID < 1 || int64(p.EmployeeID) > math.MaxInt32 ||
		!validProfileText(p.EmployeeCode, 1, 30) || !validProfileText(p.Fullname, 1, 100) {
		return false
	}
	switch p.Status {
	case "ACTIVE", "LOCKED", "DISABLED":
	default:
		return false
	}
	address, err := mail.ParseAddress(p.CompanyEmail)
	if err != nil || address.Name != "" || address.Address != p.CompanyEmail ||
		!validProfileText(p.CompanyEmail, 3, 100) || strings.ContainsAny(p.CompanyEmail, "\r\n") {
		return false
	}

	seen := map[string]bool{}
	for _, role := range p.SystemRoles {
		if (role != "ADMIN" && role != "HR" && role != "DIRECTOR") || seen[role] {
			return false
		}
		seen[role] = true
	}
	return true
}
