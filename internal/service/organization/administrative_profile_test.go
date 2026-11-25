package organization

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
)

func administrativeString(v string) AdministrativeField[string] {
	return AdministrativeField[string]{Present: true, Value: &v}
}
func administrativeInt(v int) AdministrativeField[int] {
	return AdministrativeField[int]{Present: true, Value: &v}
}
func administrativeCommand() EditAdministrativeProfileCommand {
	return EditAdministrativeProfileCommand{ActorID: 12, EmployeeID: 25,
		ExpectedBeforeValues: AdministrativeValues{Address: AdministrativeField[string]{Present: true}},
		Changes:              AdministrativeValues{Address: administrativeString("Địa chỉ mới")}}
}

func TestAdministrativePatchNormalizationPreservesBeforeAndCaller(t *testing.T) {
	legacy := " " + strings.Repeat("ế", 2100) + " "
	c := administrativeCommand()
	c.ExpectedBeforeValues = AdministrativeValues{
		Fullname: administrativeString(" Tên cũ "), Phone: administrativeString(strings.Repeat("p", 40)),
		Address: administrativeString(legacy), CompanyEmail: administrativeString("old@example.invalid"),
		PositionID: administrativeInt(3), DepartmentID: AdministrativeField[int]{Present: true}, JobLevelID: administrativeInt(5),
	}
	c.Changes = AdministrativeValues{
		Fullname: administrativeString("\u0085 Tên mới \u2003"), Phone: AdministrativeField[string]{Present: true},
		Address: administrativeString("  Địa chỉ mới  "), CompanyEmail: administrativeString("  NEW@EXAMPLE.INVALID  "),
		PositionID: administrativeInt(4), DepartmentID: administrativeInt(6), JobLevelID: AdministrativeField[int]{Present: true},
	}
	out, err := ValidateEditAdministrativeProfile(c)
	if err != nil {
		t.Fatal(err)
	}
	if *out.Changes.Fullname.Value != "Tên mới" || *out.Changes.CompanyEmail.Value != "new@example.invalid" || *out.Changes.Address.Value != "Địa chỉ mới" || out.Changes.Phone.Value != nil || out.Changes.JobLevelID.Value != nil {
		t.Fatal("chuẩn hóa proposed hoặc null sai")
	}
	if *out.ExpectedBeforeValues.Address.Value != legacy || *out.ExpectedBeforeValues.Fullname.Value != " Tên cũ " || len(*out.ExpectedBeforeValues.Phone.Value) != 40 {
		t.Fatal("before legacy bị chuẩn hóa/cắt")
	}
	*out.Changes.Fullname.Value = "changed"
	*out.ExpectedBeforeValues.Address.Value = "changed"
	*out.Changes.PositionID.Value = 9
	if *c.Changes.Fullname.Value != "\u0085 Tên mới \u2003" || *c.ExpectedBeforeValues.Address.Value != legacy || *c.Changes.PositionID.Value != 4 {
		t.Fatal("command còn dùng chung pointer caller")
	}
}

func TestAdministrativePatchUnicodeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name string
		n    int
		set  func(*AdministrativeValues, AdministrativeField[string])
	}{{"fullname", 100, func(v *AdministrativeValues, f AdministrativeField[string]) { v.Fullname = f }},
		{"phone", 20, func(v *AdministrativeValues, f AdministrativeField[string]) { v.Phone = f }},
		{"address", 2000, func(v *AdministrativeValues, f AdministrativeField[string]) { v.Address = f }}} {
		t.Run(tc.name, func(t *testing.T) {
			for _, n := range []int{tc.n, tc.n + 1} {
				c := EditAdministrativeProfileCommand{ActorID: 1, EmployeeID: 2}
				tc.set(&c.ExpectedBeforeValues, administrativeString("before"))
				tc.set(&c.Changes, administrativeString("  "+strings.Repeat("😀", n)+"  "))
				_, err := ValidateEditAdministrativeProfile(c)
				if (n == tc.n && err != nil) || (n > tc.n && !errors.Is(err, ErrValidation)) {
					t.Fatal("giới hạn code point sau trim sai")
				}
			}
		})
	}
}

func TestAdministrativePatchRejectsInvalidInput(t *testing.T) {
	for name, edit := range map[string]func(*EditAdministrativeProfileCommand){
		"empty": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues = AdministrativeValues{}
			c.Changes = AdministrativeValues{}
		},
		"different keys": func(c *EditAdministrativeProfileCommand) {
			c.Changes.Address.Present = false
			c.Changes.Address.Value = nil
		},
		"hidden value": func(c *EditAdministrativeProfileCommand) {
			c.Changes.Fullname = AdministrativeField[string]{Value: administrativeString("x").Value}
		},
		"null address":  func(c *EditAdministrativeProfileCommand) { c.Changes.Address.Value = nil },
		"blank address": func(c *EditAdministrativeProfileCommand) { c.Changes.Address = administrativeString("\u0085\u2003") },
		"invalid utf8": func(c *EditAdministrativeProfileCommand) {
			c.Changes.Address = administrativeString(string([]byte{0xff}))
		},
		"nul before": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.Address = administrativeString("\x00")
		},
		"nul proposed": func(c *EditAdministrativeProfileCommand) { c.Changes.Address = administrativeString("x\x00") },
		"position null": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.PositionID = administrativeInt(1)
			c.Changes.PositionID.Present = true
		},
		"department zero": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.DepartmentID.Present = true
			c.Changes.DepartmentID = administrativeInt(0)
		},
		"job overflow": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.JobLevelID.Present = true
			c.Changes.JobLevelID = administrativeInt(math.MaxInt32 + 1)
		},
		"null fullname before": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.Fullname.Present = true
			c.Changes.Fullname = administrativeString("Name")
		},
		"null email before": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.CompanyEmail.Present = true
			c.Changes.CompanyEmail = administrativeString("a@example.invalid")
		},
		"email malformed": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.CompanyEmail = administrativeString("a@example.invalid")
			c.Changes.CompanyEmail = administrativeString("broken")
		},
		"email display name": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.CompanyEmail = administrativeString("a@example.invalid")
			c.Changes.CompanyEmail = administrativeString("Name <a@example.invalid>")
		},
		"email newline": func(c *EditAdministrativeProfileCommand) {
			c.ExpectedBeforeValues.CompanyEmail = administrativeString("a@example.invalid")
			c.Changes.CompanyEmail = administrativeString("a@example.invalid\n")
		},
		"bad target": func(c *EditAdministrativeProfileCommand) { c.EmployeeID = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			c := administrativeCommand()
			edit(&c)
			if _, err := ValidateEditAdministrativeProfile(c); !errors.Is(err, ErrValidation) {
				t.Fatal("đầu vào sai không bị chặn")
			}
		})
	}
	for _, id := range []int{0, math.MaxInt32 + 1} {
		c := administrativeCommand()
		c.ActorID = id
		if _, err := ValidateEditAdministrativeProfile(c); !errors.Is(err, ErrUnauthenticated) {
			t.Fatal("actor không hợp lệ")
		}
	}
}

type administrativeRepoStub struct {
	p     AdministrativeProfile
	err   error
	calls int
	last  EditAdministrativeProfileCommand
	query ListAdministrativeEmployeesQuery
}

func (r *administrativeRepoStub) ListAdministrativeEmployees(_ context.Context, q ListAdministrativeEmployeesQuery) (AdministrativeEmployeePage, error) {
	r.calls++
	r.query = q
	return AdministrativeEmployeePage{Page: q.Page, PageSize: q.PageSize}, r.err
}
func (r *administrativeRepoStub) ReadAdministrativeProfile(context.Context, AdministrativeProfileQuery) (AdministrativeProfile, error) {
	r.calls++
	return r.p, r.err
}
func (r *administrativeRepoStub) EditAdministrativeProfile(_ context.Context, c EditAdministrativeProfileCommand) (AdministrativeProfile, error) {
	r.calls++
	r.last = c
	return r.p, r.err
}
func TestAdministrativeServicePurposeAndErrors(t *testing.T) {
	p := validSelfProfile()
	p.EmploymentStatus = "INACTIVE"
	phone := strings.Repeat("p", 40)
	p.Phone = &phone
	r := &administrativeRepoStub{p: p}
	s := NewAdministrativeProfileService(r)
	out, err := s.Read(t.Context(), AdministrativeProfileQuery{1, 25})
	if err != nil || out.UserID != 12 || out.CompanyEmail != "synthetic@example.invalid" || len(*out.Phone) != 40 || out.EmploymentStatus != "INACTIVE" {
		t.Fatal("projection HR không được áp quyền self/giới hạn nhập mới")
	}
	if _, err = s.Edit(t.Context(), administrativeCommand()); err != nil {
		t.Fatal(err)
	}
	page, err := s.List(t.Context(), ListAdministrativeEmployeesQuery{ActorID: 1, Query: "  %_  "})
	if err != nil || page.Items == nil || page.Page != 1 || page.PageSize != 20 || r.query.Query != "%_" {
		t.Fatal("default/query literal sai")
	}
	before := r.calls
	if _, err = s.Read(t.Context(), AdministrativeProfileQuery{0, 25}); !errors.Is(err, ErrUnauthenticated) || r.calls != before {
		t.Fatal("input sai vẫn gọi DB")
	}
	for _, want := range []error{ErrStale, ErrEmailConflict, ErrCatalogNotFound, ErrForbidden, ErrPasswordChangeRequired, ErrUnavailable, ErrPersistence, ErrCommitUnknown} {
		r.err = fmt.Errorf("raw provider error: %w", want)
		_, err = s.Edit(t.Context(), administrativeCommand())
		if !errors.Is(err, want) || strings.Contains(err.Error(), "raw provider") {
			t.Fatal("mất phân loại hoặc lộ lỗi adapter")
		}
	}
	r.err = errors.New("raw unexpected SQL")
	if _, err = s.Read(t.Context(), AdministrativeProfileQuery{1, 25}); err != ErrPersistence {
		t.Fatal("lỗi lạ phải thành persistence an toàn")
	}
}

func TestAdministrativeQueriesAndCorruptProjection(t *testing.T) {
	for _, q := range []ListAdministrativeEmployeesQuery{
		{ActorID: 1, Page: -1, PageSize: 20}, {ActorID: 1, Page: math.MaxInt32 + 1, PageSize: 20},
		{ActorID: 1, Page: 1, PageSize: 101}, {ActorID: 1, Page: 1, PageSize: 20, Query: strings.Repeat("ế", 101)},
	} {
		if _, err := ValidateListAdministrativeEmployees(q); !errors.Is(err, ErrValidation) {
			t.Fatal("query sai")
		}
	}
	for name, edit := range map[string]func(*AdministrativeProfile){
		"wrong target": func(p *AdministrativeProfile) { p.EmployeeID++ },
		"missing user": func(p *AdministrativeProfile) { p.UserID = 0 },
		"bad position": func(p *AdministrativeProfile) { p.Position.ID = 0 },
		"bad status":   func(p *AdministrativeProfile) { p.EmploymentStatus = "OTHER" },
		"bad email":    func(p *AdministrativeProfile) { p.CompanyEmail = "broken" },
	} {
		t.Run(name, func(t *testing.T) {
			p := validSelfProfile()
			edit(&p)
			_, err := NewAdministrativeProfileService(&administrativeRepoStub{p: p}).Read(t.Context(), AdministrativeProfileQuery{1, 25})
			if err != ErrPersistence {
				t.Fatal("projection hỏng trả success")
			}
		})
	}
}
