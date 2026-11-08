package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrIdentityConflict   = errors.New("danh tính đăng nhập mâu thuẫn")
	ErrUnavailable        = errors.New("kho dữ liệu xác thực tạm thời không khả dụng")
	ErrPersistence        = errors.New("lỗi lưu trữ dữ liệu xác thực")
	ErrInvalidCredentials = errors.New("thông tin đăng nhập không còn hợp lệ")
	ErrUnauthenticated    = errors.New("phiên xác thực không còn hợp lệ")
)

type Clock interface {
	Now() time.Time
}

type FinalizeLoginCommand struct {
	VerifiedIdentity  LoginIdentity
	NewTokenHash      SessionHash
	PreviousTokenHash *SessionHash
	TTL               time.Duration
}

// SessionWriter hoàn tất từng thao tác nguyên vẹn trước khi trả thành công.
type SessionWriter interface {
	FinalizeLogin(context.Context, FinalizeLoginCommand) (SessionSnapshot, error)
	RevokeCurrentSession(context.Context, SessionHash) error
}

type SessionHash [32]byte

// LoginIdentity là dữ liệu nội bộ để xác minh mật khẩu và đối chiếu lại khi cấp phiên.
type LoginIdentity struct {
	UserID         int
	EmployeeID     int
	Email          string
	PasswordHash   string
	UserActive     bool
	EmployeeActive bool
}

// Principal giữ role lưu trong DB; service xác định quyền hiệu lực của phiên.
type Principal struct {
	UserID             int
	EmployeeID         int
	EmployeeCode       string
	Fullname           string
	SystemRoles        []string
	MustChangePassword bool
}

type SessionSnapshot struct {
	Principal      Principal
	CreatedAt      time.Time
	ExpiresAt      time.Time
	RevokedAt      *time.Time
	UserActive     bool
	EmployeeActive bool
}

// IdentityReader trả found=false, err=nil khi không tìm được danh tính đầy đủ.
// Caller kiểm tra hiệu lực từ snapshot sau khi đọc.
type IdentityReader interface {
	FindLoginIdentity(ctx context.Context, email string) (LoginIdentity, bool, error)
	ReadSession(ctx context.Context, hash SessionHash) (SessionSnapshot, bool, error)
}
