package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordPolicy(t *testing.T) {
	for _, tc := range []struct {
		name, password string
		valid          bool
	}{
		{"14 code point", strings.Repeat("ắ", 14), false},
		{"15 code point", strings.Repeat("ắ", 15), true},
		{"72 byte", strings.Repeat("ắ", 24), true},
		{"73 byte", strings.Repeat("ắ", 24) + "a", false},
		{"khoảng trắng", " " + strings.Repeat("a", 13) + " ", true},
		{"trùng", "mật khẩu hiện tại", false},
		{"UTF-8 sai", strings.Repeat("a", 15) + "\xff", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := validNewPassword("mật khẩu hiện tại", tc.password); got != tc.valid {
				t.Fatalf("valid = %v, muốn %v", got, tc.valid)
			}
		})
	}
}

type passwordChangeTestRepo struct {
	basis             PasswordChangeBasis
	found             bool
	readErr, writeErr error
	readHash          SessionHash
	commands          []ChangeInitialPasswordCommand
}

func (r *passwordChangeTestRepo) ReadPasswordChangeBasis(_ context.Context, hash SessionHash) (PasswordChangeBasis, bool, error) {
	r.readHash = hash
	return r.basis, r.found, r.readErr
}

func (r *passwordChangeTestRepo) ChangeInitialPassword(_ context.Context, cmd ChangeInitialPasswordCommand) error {
	r.commands = append(r.commands, cmd)
	return r.writeErr
}

func TestPasswordChangeService(t *testing.T) {
	const current = " mật khẩu tạm "
	const next = " Mật khẩu mới có dấu "
	hash, err := bcrypt.GenerateFromPassword([]byte(current), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1800000000, 0)
	secret := SessionSecret{1, 2, 3}
	for _, tc := range []struct {
		name, current, next string
		change              func(*passwordChangeTestRepo)
		want                error
		writes              int
	}{
		{"thành công", current, next, nil, nil, 1},
		{"sai mật khẩu", "mật khẩu tạm", next, nil, ErrInvalidCurrentPassword, 0},
		{"current vượt 72 byte", strings.Repeat("a", 73), next, nil, ErrInvalidCurrentPassword, 0},
		{"policy sai", current, "ngắn", nil, ErrNewPasswordPolicy, 0},
		{"không tìm thấy", current, next, func(r *passwordChangeTestRepo) { r.found = false }, ErrUnauthenticated, 0},
		{"hết hạn tại biên", current, next, func(r *passwordChangeTestRepo) { r.basis.Session.ExpiresAt = now }, ErrUnauthenticated, 0},
		{"đã thu hồi", current, next, func(r *passwordChangeTestRepo) { r.basis.Session.RevokedAt = &now }, ErrUnauthenticated, 0},
		{"nhân viên inactive", current, next, func(r *passwordChangeTestRepo) { r.basis.Session.EmployeeActive = false }, ErrUnauthenticated, 0},
		{"tài khoản inactive", current, next, func(r *passwordChangeTestRepo) { r.basis.Session.UserActive = false }, ErrUnauthenticated, 0},
		{"cờ false", current, next, func(r *passwordChangeTestRepo) { r.basis.Session.Principal.MustChangePassword = false }, ErrPasswordChangeNotRequired, 0},
		{"hash lỗi", current, next, func(r *passwordChangeTestRepo) { r.basis.PasswordHash = "invalid" }, ErrPersistence, 0},
		{"lỗi đọc", current, next, func(r *passwordChangeTestRepo) { r.readErr = ErrUnavailable }, ErrUnavailable, 0},
		{"nền đổi dưới khóa", current, next, func(r *passwordChangeTestRepo) { r.writeErr = ErrUnauthenticated }, ErrUnauthenticated, 1},
		{"lỗi commit", current, next, func(r *passwordChangeTestRepo) { r.writeErr = ErrUnavailable }, ErrUnavailable, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &passwordChangeTestRepo{found: true, basis: PasswordChangeBasis{
				PasswordHash: string(hash), Session: SessionSnapshot{
					Principal:  Principal{UserID: 4, EmployeeID: 19, MustChangePassword: true},
					UserActive: true, EmployeeActive: true, ExpiresAt: now.Add(time.Hour),
				},
			}}
			if tc.change != nil {
				tc.change(repo)
			}
			service := NewPasswordChangeService(repo, sessionTestClock(func() time.Time { return now }))
			err := service.ChangeInitialPassword(t.Context(), secret, tc.current, tc.next)
			if !errors.Is(err, tc.want) || len(repo.commands) != tc.writes {
				t.Fatalf("error = %v, writes = %d; muốn %v, %d", err, len(repo.commands), tc.want, tc.writes)
			}
			tokenHash := sha256.Sum256(secret[:])
			if repo.readHash != tokenHash {
				t.Fatal("đọc sai phiên")
			}
			for _, cmd := range repo.commands {
				if cmd.UserID != 4 || cmd.TokenHash != tokenHash || cmd.VerifiedPasswordHash != string(hash) {
					t.Fatal("command không khớp nền xác minh")
				}
				if bcrypt.CompareHashAndPassword([]byte(cmd.NewPasswordHash), []byte(next)) != nil ||
					bcrypt.CompareHashAndPassword([]byte(cmd.NewPasswordHash), []byte(strings.TrimSpace(next))) == nil {
					t.Fatal("hash mới không giữ đúng nội dung mật khẩu")
				}
			}
		})
	}
}
