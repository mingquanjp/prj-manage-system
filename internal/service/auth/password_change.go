package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordChangeNotRequired = errors.New("tài khoản không yêu cầu đổi mật khẩu lần đầu")
	ErrNewPasswordPolicy         = errors.New("mật khẩu mới không đạt chính sách")
	ErrInvalidCurrentPassword    = errors.New("mật khẩu hiện tại không hợp lệ")
)

type passwordChangeRepository interface {
	ReadPasswordChangeBasis(context.Context, SessionHash) (PasswordChangeBasis, bool, error)
	ChangeInitialPassword(context.Context, ChangeInitialPasswordCommand) error
}

type PasswordChangeService struct {
	repository passwordChangeRepository
	clock      Clock
}

func NewPasswordChangeService(repository passwordChangeRepository, clock Clock) *PasswordChangeService {
	return &PasswordChangeService{repository: repository, clock: clock}
}

func validNewPassword(current, next string) bool {
	return current != next && len(next) <= 72 && utf8.ValidString(next) && utf8.RuneCountInString(next) >= 15
}

func (s *PasswordChangeService) ChangeInitialPassword(ctx context.Context, secret SessionSecret, current, next string) error {
	tokenHash := sha256.Sum256(secret[:])
	basis, found, err := s.repository.ReadPasswordChangeBasis(ctx, tokenHash)
	if err != nil {
		return err
	}
	if !found {
		return ErrUnauthenticated
	}
	principal, err := evaluateSession(basis.Session, s.clock.Now())
	if err != nil {
		return err
	}
	if !principal.MustChangePassword {
		return ErrPasswordChangeNotRequired
	}
	if !validNewPassword(current, next) {
		return ErrNewPasswordPolicy
	}

	// Không xác minh theo tiền tố khi mật khẩu hiện tại vượt giới hạn bcrypt.
	if current == "" || len(current) > 72 {
		return ErrInvalidCurrentPassword
	}
	err = bcrypt.CompareHashAndPassword([]byte(basis.PasswordHash), []byte(current))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrInvalidCurrentPassword
	}
	if err != nil {
		return ErrPersistence
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(next), bcrypt.DefaultCost)
	if err != nil {
		return ErrPersistence
	}
	// Adapter kiểm lại nền dưới khóa trước khi thay hash và thu hồi phiên.
	return s.repository.ChangeInitialPassword(ctx, ChangeInitialPasswordCommand{
		TokenHash:            tokenHash,
		UserID:               principal.UserID,
		VerifiedPasswordHash: basis.PasswordHash,
		NewPasswordHash:      string(newHash),
	})
}
