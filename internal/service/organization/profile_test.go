package organization

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type profileRepoStub struct {
	p     SelfProfile
	err   error
	calls int
}

func (r *profileRepoStub) ReadSelf(context.Context, SelfProfileQuery) (SelfProfile, error) {
	r.calls++
	return r.p, r.err
}
func validSelfProfile() SelfProfile {
	return SelfProfile{UserID: 12, EmployeeID: 25, EmployeeCode: "LEGACY-25", Fullname: "Nhân viên tổng hợp", CompanyEmail: " Synthetic@Example.invalid ", Position: CatalogItem{ID: 3, Name: "Vị trí"}, EmploymentStatus: "ACTIVE"}
}
func TestReadSelfService(t *testing.T) {
	q := SelfProfileQuery{12, 25}
	r := &profileRepoStub{p: validSelfProfile()}
	s := NewProfileService(r)
	p, err := s.ReadSelf(t.Context(), q)
	if err != nil || p.CompanyEmail != "synthetic@example.invalid" || p.Department != nil || p.JobLevel != nil || p.Address != nil || p.Phone != nil {
		t.Fatal("projection nullable/email sai")
	}
	for _, tc := range []struct {
		q    SelfProfileQuery
		want error
	}{{SelfProfileQuery{0, 25}, ErrUnauthenticated}, {SelfProfileQuery{12, 0}, ErrValidation}, {SelfProfileQuery{12, 2147483648}, ErrValidation}} {
		before := r.calls
		_, err = s.ReadSelf(t.Context(), tc.q)
		if !errors.Is(err, tc.want) || r.calls != before {
			t.Fatal("validation đã gọi repository")
		}
	}
	for _, want := range []error{ErrNotFound, ErrUnauthenticated, ErrPasswordChangeRequired, ErrUnavailable, ErrPersistence} {
		r.err = want
		_, err = s.ReadSelf(t.Context(), q)
		if !errors.Is(err, want) {
			t.Fatal("mất phân loại lỗi")
		}
	}
}
func TestReadSelfRejectsCorruptProjection(t *testing.T) {
	for name, edit := range map[string]func(*SelfProfile){
		"wrong user": func(p *SelfProfile) { p.UserID++ }, "wrong employee": func(p *SelfProfile) { p.EmployeeID++ },
		"position missing": func(p *SelfProfile) { p.Position = CatalogItem{} }, "email malformed": func(p *SelfProfile) { p.CompanyEmail = "broken" },
		"email display name": func(p *SelfProfile) { p.CompanyEmail = "Name <valid@example.invalid>" }, "email newline": func(p *SelfProfile) { p.CompanyEmail = "one\r\n@example.invalid" },
		"fullname too long": func(p *SelfProfile) { p.Fullname = strings.Repeat("ế", 101) }, "inactive": func(p *SelfProfile) { p.EmploymentStatus = "INACTIVE" },
	} {
		t.Run(name, func(t *testing.T) {
			p := validSelfProfile()
			edit(&p)
			_, err := NewProfileService(&profileRepoStub{p: p}).ReadSelf(t.Context(), SelfProfileQuery{12, 25})
			if !errors.Is(err, ErrPersistence) {
				t.Fatal("dữ liệu hỏng không được trả thành công")
			}
		})
	}
}
