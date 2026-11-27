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
	if c.DepartmentID != nil {
		departmentID := *c.DepartmentID
		if departmentID < 1 || int64(departmentID) > math.MaxInt32 {
			return CreateAccountInput{}, ErrValidation
		}
		c.DepartmentID = &departmentID
	}
	if c.JobLevelID != nil {
		jobLevelID := *c.JobLevelID
		if jobLevelID < 1 || int64(jobLevelID) > math.MaxInt32 {
			return CreateAccountInput{}, ErrValidation
		}
		c.JobLevelID = &jobLevelID
	}
	if c.Phone != nil {
		phone := *c.Phone
		if !utf8.ValidString(phone) || strings.ContainsRune(phone, 0) {
			return CreateAccountInput{}, ErrValidation
		}
		phone = strings.TrimSpace(phone)
		if phone == "" || utf8.RuneCountInString(phone) > 20 {
			return CreateAccountInput{}, ErrValidation
		}
		c.Phone = &phone
	}
	if c.Address != nil {
		address := *c.Address
		if !utf8.ValidString(address) || strings.ContainsRune(address, 0) {
			return CreateAccountInput{}, ErrValidation
		}
		address = strings.TrimSpace(address)
		if address == "" {
			return CreateAccountInput{}, ErrValidation
		}
		c.Address = &address
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
