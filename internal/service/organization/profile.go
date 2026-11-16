package organization

import (
	"context"
	"math"
	"net/mail"
	"strings"
	"unicode/utf8"

	"ai-staffing/internal/service/auth"
)

// SelfProfile chỉ chứa hồ sơ hành chính chính chủ, không chứa quyền hay credential.
type SelfProfile struct {
	UserID           int
	EmployeeID       int
	EmployeeCode     string
	Fullname         string
	CompanyEmail     string
	Phone            *string
	Address          *string
	Position         CatalogItem
	Department       *CatalogItem
	JobLevel         *CatalogItem
	EmploymentStatus string
}

type SelfProfileQuery struct{ ActorID, EmployeeID int }

// Adapter kiểm lại trạng thái/cờ và liên kết từ DB trong cùng snapshot đọc.
type ProfileRepository interface {
	ReadSelf(context.Context, SelfProfileQuery) (SelfProfile, error)
}
type ProfileService struct{ repo ProfileRepository }

func NewProfileService(repo ProfileRepository) *ProfileService { return &ProfileService{repo: repo} }
func ValidateSelfProfileQuery(q SelfProfileQuery) error {
	if q.ActorID < 1 || int64(q.ActorID) > math.MaxInt32 {
		return ErrUnauthenticated
	}
	if q.EmployeeID < 1 || int64(q.EmployeeID) > math.MaxInt32 {
		return ErrValidation
	}
	return nil
}
func (s *ProfileService) ReadSelf(ctx context.Context, q SelfProfileQuery) (SelfProfile, error) {
	if err := ValidateSelfProfileQuery(q); err != nil {
		return SelfProfile{}, err
	}
	p, err := s.repo.ReadSelf(ctx, q)
	if err != nil {
		return SelfProfile{}, err
	}
	// Projection chuẩn hóa khi đọc; không sửa email legacy được lưu trong DB.
	p.CompanyEmail = auth.NormalizeLoginEmail(p.CompanyEmail)
	if !validSelfProfileProjection(p, q) {
		return SelfProfile{}, ErrPersistence
	}
	return p, nil
}

func validSelfProfileProjection(p SelfProfile, q SelfProfileQuery) bool {
	if p.UserID != q.ActorID || p.EmployeeID != q.EmployeeID || p.EmploymentStatus != "ACTIVE" {
		return false
	}
	if !validProfileText(p.EmployeeCode, 1, 30) || !validProfileText(p.Fullname, 1, 100) {
		return false
	}
	email, err := mail.ParseAddress(p.CompanyEmail)
	if !validProfileText(p.CompanyEmail, 3, 100) || err != nil || email.Address != p.CompanyEmail || strings.ContainsAny(p.CompanyEmail, "\r\n") {
		return false
	}
	validItem := func(v CatalogItem) bool {
		return v.ID > 0 && int64(v.ID) <= math.MaxInt32 && validProfileText(v.Name, 1, 100)
	}
	if !validItem(p.Position) || (p.Department != nil && !validItem(*p.Department)) || (p.JobLevel != nil && !validItem(*p.JobLevel)) {
		return false
	}
	if p.Phone != nil && !validProfileText(*p.Phone, 0, 20) {
		return false
	}
	return p.Address == nil || (utf8.ValidString(*p.Address) && !strings.ContainsRune(*p.Address, 0))
}
func validProfileText(value string, min, max int) bool {
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return false
	}
	n := utf8.RuneCountInString(value)
	return n >= min && n <= max
}
