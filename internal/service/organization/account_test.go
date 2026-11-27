package organization

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func validAccountInput() CreateAccountInput {
	return CreateAccountInput{ActorID: 12, Fullname: " Nhân viên thử nghiệm ", CompanyEmail: " New.Employee@EXAMPLE.invalid ", PositionID: 2}
}
func TestValidateAccountInput(t *testing.T) {
	// Tất cả tổ hợp của ba role đều hợp lệ; không có hierarchy Admin superset.
	for mask := 0; mask < 8; mask++ {
		input := validAccountInput()
		for i, role := range []string{"DIRECTOR", "HR", "ADMIN"} {
			if mask&(1<<i) != 0 {
				input.SystemRoles = append(input.SystemRoles, role)
			}
		}
		got, err := ValidateAccountInput(input)
		if err != nil || len(got.SystemRoles) != len(input.SystemRoles) {
			t.Fatal("tổ hợp role hợp lệ bị từ chối")
		}
	}
	for _, tc := range []struct {
		name string
		edit func(*CreateAccountInput)
	}{
		{"actor missing", func(c *CreateAccountInput) { c.ActorID = 0 }},
		{"name blank", func(c *CreateAccountInput) { c.Fullname = "\u00a0\t" }},
		{"name101", func(c *CreateAccountInput) { c.Fullname = strings.Repeat("ệ", 101) }},
		{"invalid utf8", func(c *CreateAccountInput) { c.Fullname = string([]byte{255}) }},
		{"name nul", func(c *CreateAccountInput) { c.Fullname = "a\x00b" }},
		{"email list", func(c *CreateAccountInput) { c.CompanyEmail = "a@example.invalid,b@example.invalid" }},
		{"email display", func(c *CreateAccountInput) { c.CompanyEmail = "Name <a@example.invalid>" }},
		{"email malformed", func(c *CreateAccountInput) { c.CompanyEmail = "a@@example.invalid" }},
		{"email CRLF", func(c *CreateAccountInput) { c.CompanyEmail = "a@example.invalid\r\n" }},
		{"email long", func(c *CreateAccountInput) { c.CompanyEmail = strings.Repeat("a", 90) + "@example.invalid" }},
		{"position missing", func(c *CreateAccountInput) { c.PositionID = 0 }},
		{"id overflow", func(c *CreateAccountInput) { c.PositionID = math.MaxInt32 + 1 }},
		{"department zero", func(c *CreateAccountInput) { n := 0; c.DepartmentID = &n }},
		{"department overflow", func(c *CreateAccountInput) { n := math.MaxInt32 + 1; c.DepartmentID = &n }},
		{"level negative", func(c *CreateAccountInput) { n := -1; c.JobLevelID = &n }},
		{"level overflow", func(c *CreateAccountInput) { n := math.MaxInt32 + 1; c.JobLevelID = &n }},
		{"phone blank", func(c *CreateAccountInput) { s := "\t "; c.Phone = &s }},
		{"phone21", func(c *CreateAccountInput) { s := strings.Repeat("1", 21); c.Phone = &s }},
		{"phone invalid utf8", func(c *CreateAccountInput) { s := string([]byte{255}); c.Phone = &s }},
		{"phone nul", func(c *CreateAccountInput) { s := "1\x00"; c.Phone = &s }},
		{"address blank", func(c *CreateAccountInput) { s := "\u00a0"; c.Address = &s }},
		{"address invalid utf8", func(c *CreateAccountInput) { s := string([]byte{255}); c.Address = &s }},
		{"address nul", func(c *CreateAccountInput) { s := "a\x00"; c.Address = &s }},
		{"role PM", func(c *CreateAccountInput) { c.SystemRoles = []string{"PM"} }},
		{"role employee", func(c *CreateAccountInput) { c.SystemRoles = []string{"EMPLOYEE"} }},
		{"duplicate role", func(c *CreateAccountInput) { c.SystemRoles = []string{"HR", "HR"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := validAccountInput()
			tc.edit(&c)
			if _, err := ValidateAccountInput(c); err == nil {
				t.Fatal("phải từ chối input sai")
			}
		})
	}
	c := validAccountInput()
	c.Fullname = " " + strings.Repeat("ệ", 100) + " "
	phone := " " + strings.Repeat("1", 20) + " "
	address := "  Địa chỉ tổng hợp  "
	id := 3
	c.Phone = &phone
	c.Address = &address
	c.DepartmentID = &id
	c.JobLevelID = &id
	c.SystemRoles = []string{"DIRECTOR", "ADMIN", "HR"}
	got, err := ValidateAccountInput(c)
	if err != nil || got.CompanyEmail != "new.employee@example.invalid" || got.Fullname != strings.Repeat("ệ", 100) || *got.Phone != strings.Repeat("1", 20) || *got.Address != "Địa chỉ tổng hợp" || !reflect.DeepEqual(got.SystemRoles, []string{"ADMIN", "HR", "DIRECTOR"}) {
		t.Fatal("chuẩn hóa/biên input sai")
	}
	if phone != " "+strings.Repeat("1", 20)+" " || address != "  Địa chỉ tổng hợp  " {
		t.Fatal("validation không được sửa dữ liệu caller")
	}
	id = 8
	phone = "changed"
	address = "changed"
	c.SystemRoles[0] = "changed"
	if *got.DepartmentID != 3 || *got.JobLevelID != 3 || *got.Phone != strings.Repeat("1", 20) || *got.Address != "Địa chỉ tổng hợp" || got.SystemRoles[2] != "DIRECTOR" {
		t.Fatal("command không được giữ pointer/slice caller")
	}
	got, err = ValidateAccountInput(validAccountInput())
	if err != nil || got.Phone != nil || got.Address != nil || got.DepartmentID != nil || got.JobLevelID != nil || got.SystemRoles == nil || len(got.SystemRoles) != 0 {
		t.Fatal("nullable/roles rỗng sai")
	}
}

type accountRepoFunc func(context.Context, CreateAccountCommand) (AccountProjection, error)

func (f accountRepoFunc) CreateAccount(ctx context.Context, c CreateAccountCommand) (AccountProjection, error) {
	return f(ctx, c)
}
func TestAccountServiceBcryptAndSecretBoundary(t *testing.T) {
	var hash string
	calls := 0
	s := NewAccountService(accountRepoFunc(func(_ context.Context, c CreateAccountCommand) (AccountProjection, error) {
		calls++
		hash = c.PasswordHash
		if c.Input.CompanyEmail != "new.employee@example.invalid" {
			t.Fatal("repository cần email canonical")
		}
		return AccountProjection{UserID: 101, EmployeeID: 201, EmployeeCode: "EMP000001", Fullname: c.Input.Fullname, CompanyEmail: c.Input.CompanyEmail, Status: "ACTIVE", SystemRoles: c.Input.SystemRoles, MustChangePassword: true}, nil
	}))
	entropy := bytes.Repeat([]byte{0xA5}, 32)
	s.random = bytes.NewReader(entropy)
	result, err := s.Create(t.Context(), validAccountInput())
	if err != nil || calls != 1 || result.Account.UserID == result.Account.EmployeeID {
		t.Fatal("create service không thành công đúng liên kết")
	}
	password := result.PasswordForDelivery().RevealForDelivery()
	if password != base64.RawURLEncoding.EncodeToString(entropy) || len(password) != 43 || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		t.Fatal("32 byte entropy hoặc bcrypt không khớp")
	}
	for _, v := range []any{result, result.PasswordForDelivery(), CreateAccountCommand{Input: validAccountInput(), PasswordHash: hash}} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			text := fmt.Sprintf(format, v)
			if strings.Contains(text, password) || strings.Contains(text, hash) {
				t.Fatal("format làm lộ credential")
			}
		}
		if _, err := json.Marshal(v); err == nil {
			t.Fatal("kết quả nội bộ không được serialize JSON")
		}
		var log bytes.Buffer
		slog.New(slog.NewJSONHandler(&log, nil)).Info("test", "value", v)
		if strings.Contains(log.String(), password) || strings.Contains(log.String(), hash) || strings.Contains(log.String(), "EXAMPLE.invalid") {
			t.Fatal("log làm lộ dữ liệu nội bộ")
		}
	}
}

func TestAccountServiceFailureDoesNotReturnCredential(t *testing.T) {
	for _, kind := range []error{ErrValidation, ErrUnauthenticated, ErrPasswordChangeRequired, ErrForbidden, ErrEmailConflict, ErrUnavailable, ErrPersistence, ErrCommitUnknown, context.Canceled, errors.New("raw driver private detail")} {
		calls := 0
		s := NewAccountService(accountRepoFunc(func(context.Context, CreateAccountCommand) (AccountProjection, error) {
			calls++
			return AccountProjection{UserID: 999}, errors.Join(kind, errors.New("raw driver private detail"))
		}))
		result, err := s.Create(t.Context(), validAccountInput())
		if err == nil || calls != 1 || result.Account.UserID != 0 || result.PasswordForDelivery().RevealForDelivery() != "" || strings.Contains(err.Error(), "private detail") {
			t.Fatal("lỗi không được trả credential/projection/rawerror")
		}
		if kind == ErrCommitUnknown && !errors.Is(err, ErrCommitUnknown) {
			t.Fatal("mất loại commit chưa xác định")
		}
	}
}

func TestAccountServiceRejectsBeforeRepository(t *testing.T) {
	for _, stage := range []string{"validation", "random short", "random error", "bcrypt", "cancel before", "cancel after hash"} {
		t.Run(stage, func(t *testing.T) {
			calls := 0
			hashCalls := 0
			s := NewAccountService(accountRepoFunc(func(context.Context, CreateAccountCommand) (AccountProjection, error) {
				calls++
				return AccountProjection{}, nil
			}))
			input := validAccountInput()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			s.hash = func([]byte, int) ([]byte, error) {
				hashCalls++
				if stage == "cancel after hash" {
					cancel()
				}
				if stage == "bcrypt" {
					return nil, errors.New("bcrypt internal")
				}
				return []byte("synthetic internal hash"), nil
			}
			switch stage {
			case "validation":
				input.PositionID = 0
			case "random short":
				s.random = bytes.NewReader([]byte{1})
			case "random error":
				s.random = errorRandom{}
			case "cancel before":
				cancel()
			}
			result, err := s.Create(ctx, input)
			if err == nil || calls != 0 || result.PasswordForDelivery().RevealForDelivery() != "" {
				t.Fatal("lỗi chuẩn bị không được mở transaction/trả secret")
			}
			if (stage == "random short" || stage == "random error" || stage == "validation" || stage == "cancel before") && hashCalls != 0 {
				t.Fatal("không bcrypt sau cổng lỗi")
			}
		})
	}
}

type errorRandom struct{}

func (errorRandom) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
