package organization

import (
	"math"
	"net/mail"
	"strings"
	"unicode/utf8"

	"ai-staffing/internal/service/auth"
)

// ValidateAccountInput dùng lại chuẩn hóa email auth, không suy role từ chức danh.
// Copy pointer/slice để caller không làm đổi command trong lúc chờ DB.
func ValidateAccountInput(c CreateAccountInput) (CreateAccountInput, error) {
	if c.ActorID < 1 || int64(c.ActorID) > math.MaxInt32 {
		return CreateAccountInput{}, ErrUnauthenticated
	}
	var err error
	c.Fullname, err = NormalizeCatalogName(c.Fullname)
	if err != nil {
		return CreateAccountInput{}, err
	}
	if !utf8.ValidString(c.CompanyEmail) || strings.ContainsAny(c.CompanyEmail, "\r\n\x00") {
		return CreateAccountInput{}, ErrValidation
	}
	c.CompanyEmail = auth.NormalizeLoginEmail(c.CompanyEmail)
	address, err := mail.ParseAddress(c.CompanyEmail)
	if err != nil || address.Name != "" || address.Address != c.CompanyEmail || utf8.RuneCountInString(c.CompanyEmail) > 100 {
		return CreateAccountInput{}, ErrValidation
	}
	if c.PositionID < 1 || int64(c.PositionID) > math.MaxInt32 {
		return CreateAccountInput{}, ErrValidation
	}
	for _, p := range []**int{&c.DepartmentID, &c.JobLevelID} {
		if *p != nil {
			n := **p
			if n < 1 || int64(n) > math.MaxInt32 {
				return CreateAccountInput{}, ErrValidation
			}
			*p = &n
		}
	}
	for _, field := range []struct {
		p     **string
		limit int
	}{{&c.Phone, 20}, {&c.Address, 0}} {
		if *field.p == nil {
			continue
		}
		raw := **field.p
		if !utf8.ValidString(raw) || strings.ContainsRune(raw, 0) {
			return CreateAccountInput{}, ErrValidation
		}
		value := strings.TrimSpace(raw)
		if value == "" || (field.limit > 0 && utf8.RuneCountInString(value) > field.limit) {
			return CreateAccountInput{}, ErrValidation
		}
		*field.p = &value
	}
	seen := map[string]bool{}
	for _, role := range c.SystemRoles {
		if seen[role] || (role != "ADMIN" && role != "HR" && role != "DIRECTOR") {
			return CreateAccountInput{}, ErrValidation
		}
		seen[role] = true
	}
	c.SystemRoles = []string{}
	for _, role := range []string{"ADMIN", "HR", "DIRECTOR"} {
		if seen[role] {
			c.SystemRoles = append(c.SystemRoles, role)
		}
	}
	return c, nil
}
