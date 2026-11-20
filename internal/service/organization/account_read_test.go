package organization

import (
	"context"
	"errors"
	"math"
	"testing"
)

type accountReadStub struct {
	q    ListAccountsQuery
	page AccountPage
	err  error
}

func (r *accountReadStub) ListAccounts(_ context.Context, q ListAccountsQuery) (AccountPage, error) {
	r.q = q
	return r.page, r.err
}
func (r *accountReadStub) PreviewAccountCode(context.Context, int) (string, error) {
	return "EMP000001", r.err
}

func TestAccountReadDefaultsAndValidation(t *testing.T) {
	repo := &accountReadStub{page: AccountPage{Items: []AccountProjection{{
		UserID: 1, EmployeeID: 2, EmployeeCode: "E001", Fullname: "Nhân viên", CompanyEmail: " EXAMPLE@EXAMPLE.INVALID ",
		Status: "LOCKED", SystemRoles: []string{"ADMIN", "HR"},
	}}}}
	s := NewAccountReadService(repo)
	page, err := s.List(t.Context(), ListAccountsQuery{ActorID: 3, Query: "  %_  "})
	if err != nil || repo.q.Page != 1 || repo.q.PageSize != 20 || repo.q.Query != "%_" || page.Items[0].CompanyEmail != "example@example.invalid" {
		t.Fatal("list không giữ default/normalize/projection")
	}
	for _, q := range []ListAccountsQuery{
		{ActorID: 3, Page: -1}, {ActorID: 3, Page: math.MaxInt32 + 1}, {ActorID: 3, PageSize: 101},
		{ActorID: 3, Query: string([]byte{255})}, {ActorID: 3, Query: "a\x00b"},
	} {
		if _, err = s.List(t.Context(), q); !errors.Is(err, ErrValidation) {
			t.Fatal("query sai phải từ chối")
		}
	}
	if _, err = s.Preview(t.Context(), 0); !errors.Is(err, ErrUnauthenticated) {
		t.Fatal("actor sai")
	}
	repo.err = ErrUnavailable
	if _, err = s.List(t.Context(), ListAccountsQuery{ActorID: 3}); !errors.Is(err, ErrUnavailable) {
		t.Fatal("lỗi port bị che")
	}
}

func TestAccountReadRejectsMalformedProjection(t *testing.T) {
	p := AccountProjection{UserID: 1, EmployeeID: 2, EmployeeCode: "E001", Fullname: "Name", CompanyEmail: "a@example.invalid", Status: "ACTIVE", SystemRoles: []string{}}
	for _, mutate := range []func(*AccountProjection){
		func(p *AccountProjection) { p.UserID = 0 },
		func(p *AccountProjection) { p.CompanyEmail = "A <a@example.invalid>" },
		func(p *AccountProjection) { p.SystemRoles = []string{"ADMIN", "ADMIN"} },
		func(p *AccountProjection) { p.SystemRoles = []string{"EMPLOYEE"} },
		func(p *AccountProjection) { p.Status = "UNKNOWN" },
	} {
		cp := p
		mutate(&cp)
		s := NewAccountReadService(&accountReadStub{page: AccountPage{Items: []AccountProjection{cp}}})
		if _, err := s.List(t.Context(), ListAccountsQuery{ActorID: 3}); !errors.Is(err, ErrPersistence) {
			t.Fatal("projection hỏng không bị từ chối")
		}
	}
}
