package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type loginTestRepo struct {
	identity LoginIdentity
	found    bool
	readErr  error
	writeErr error
	snapshot SessionSnapshot
	email    string
	command  *FinalizeLoginCommand
}

func (r *loginTestRepo) FindLoginIdentity(_ context.Context, email string) (LoginIdentity, bool, error) {
	r.email = email
	return r.identity, r.found, r.readErr
}

func (r *loginTestRepo) FinalizeLogin(_ context.Context, command FinalizeLoginCommand) (SessionSnapshot, error) {
	r.command = &command
	return r.snapshot, r.writeErr
}

type loginTestClock struct{ now time.Time }

func (c loginTestClock) Now() time.Time { return c.now }

func TestLogin(t *testing.T) {
	const password = " pass "
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if cost, err := bcrypt.Cost([]byte(dummyLoginHash)); err != nil || cost != bcrypt.DefaultCost {
		t.Fatal("hash giả cần có cùng cost với bcrypt mặc định")
	}
	now := time.Unix(1800000000, 0)
	cases := []struct {
		name     string
		password string
		change   func(*loginTestRepo, *LoginService)
		want     error
		finalize bool
		roles    []string
	}{
		{"phiên thường", password, nil, nil, true, []string{"HR", "ADMIN"}},
		{"phiên hạn chế", password, func(r *loginTestRepo, _ *LoginService) {
			r.snapshot.Principal.MustChangePassword = true
		}, nil, true, []string{}},
		{"mật khẩu sai", "pass", nil, ErrLoginVerificationFailed, false, nil},
		{"không có tài khoản dù khớp hash giả", "public placeholder for DATN dummy verification", func(r *loginTestRepo, _ *LoginService) {
			r.found = false
		}, ErrLoginVerificationFailed, false, nil},
		{"tài khoản khóa", password, func(r *loginTestRepo, _ *LoginService) {
			r.identity.UserActive = false
		}, ErrLoginVerificationFailed, false, nil},
		{"nhân viên không hoạt động", password, func(r *loginTestRepo, _ *LoginService) {
			r.identity.EmployeeActive = false
		}, ErrLoginVerificationFailed, false, nil},
		{"hash hỏng", password, func(r *loginTestRepo, _ *LoginService) {
			r.identity.PasswordHash = "invalid"
		}, ErrPersistence, false, nil},
		{"email mâu thuẫn", password, func(r *loginTestRepo, _ *LoginService) {
			r.readErr = ErrIdentityConflict
		}, ErrIdentityConflict, false, nil},
		{"random lỗi", password, func(_ *loginTestRepo, s *LoginService) {
			s.random = iotest.ErrReader(errors.New("random failure"))
		}, ErrPersistence, false, nil},
		{"nền đã thay đổi", password, func(r *loginTestRepo, _ *LoginService) {
			r.writeErr = ErrInvalidCredentials
		}, ErrInvalidCredentials, true, nil},
		{"lỗi khi hoàn tất", password, func(r *loginTestRepo, _ *LoginService) {
			r.writeErr = ErrUnavailable
		}, ErrUnavailable, true, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &loginTestRepo{
				found:    true,
				identity: LoginIdentity{UserID: 2, EmployeeID: 7, Email: "An@example.test", PasswordHash: string(hash), UserActive: true, EmployeeActive: true},
				snapshot: SessionSnapshot{Principal: Principal{UserID: 2, EmployeeID: 7, SystemRoles: []string{"HR", "ADMIN"}}, UserActive: true, EmployeeActive: true, ExpiresAt: now.Add(8 * time.Hour)},
			}
			service := NewLoginService(repo, loginTestClock{now}, 8*time.Hour)
			secret := SessionSecret{3}
			previous := SessionSecret{9}
			service.random = bytes.NewReader(secret[:])
			if tc.change != nil {
				tc.change(repo, service)
			}

			got, err := service.Login(context.Background(), " An@Example.Test ", tc.password, &previous)
			if !errors.Is(err, tc.want) || repo.email != "an@example.test" || (repo.command != nil) != tc.finalize {
				t.Fatalf("sai lỗi, email chuẩn hóa hoặc số lần gọi finalize: %v", err)
			}
			if tc.want != nil {
				if !reflect.DeepEqual(got, LoginResult{}) {
					t.Fatal("không được trả dữ liệu phiên khi lỗi")
				}
				return
			}
			cmd := repo.command
			if cmd.VerifiedIdentity != repo.identity || cmd.NewTokenHash != sha256.Sum256(secret[:]) ||
				cmd.PreviousTokenHash == nil || *cmd.PreviousTokenHash != sha256.Sum256(previous[:]) || cmd.TTL != 8*time.Hour {
				t.Fatal("lệnh finalize sai hash hoặc nền xác thực")
			}
			if got.Secret != secret || !got.ExpiresAt.Equal(repo.snapshot.ExpiresAt) ||
				got.Principal.MustChangePassword != repo.snapshot.Principal.MustChangePassword ||
				!reflect.DeepEqual(got.Principal.SystemRoles, tc.roles) {
				t.Fatal("kết quả phiên không khớp snapshot")
			}
		})
	}
}

func TestLoginDoesNotAcceptOverlongBcryptPrefix(t *testing.T) {
	password := strings.Repeat("a", 72)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &loginTestRepo{
		found:    true,
		identity: LoginIdentity{PasswordHash: string(hash), UserActive: true, EmployeeActive: true},
	}
	service := NewLoginService(repo, loginTestClock{}, 8*time.Hour)
	_, err = service.Login(context.Background(), "an@example.test", password+"b", nil)
	if !errors.Is(err, ErrLoginVerificationFailed) || repo.command != nil {
		t.Fatal("mật khẩu vượt giới hạn byte của bcrypt phải bị từ chối trước finalize")
	}
}
