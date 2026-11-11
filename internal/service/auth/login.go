package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Hash công khai của mật khẩu giả, cùng cost với bcrypt mặc định hiện dùng.
const dummyLoginHash = "$2a$10$5wjelFvyTfb3hZsbVPR7dOTAv7sowg7dZuU38M921/CqvJGJGhFsW"

// ErrLoginVerificationFailed cho biết limiter cần ghi nhận một lần xác minh thất bại.
var ErrLoginVerificationFailed = errors.New("xác minh thông tin đăng nhập thất bại")

type loginRepository interface {
	FindLoginIdentity(context.Context, string) (LoginIdentity, bool, error)
	FinalizeLogin(context.Context, FinalizeLoginCommand) (SessionSnapshot, error)
}

type LoginService struct {
	repository loginRepository
	clock      Clock
	ttl        time.Duration
	random     io.Reader
}

type LoginResult struct {
	Principal Principal
	Secret    SessionSecret
	ExpiresAt time.Time
}

func NewLoginService(repository loginRepository, clock Clock, ttl time.Duration) *LoginService {
	return &LoginService{repository: repository, clock: clock, ttl: ttl, random: rand.Reader}
}

func NormalizeLoginEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *LoginService) Login(ctx context.Context, email, password string, previous *SessionSecret) (LoginResult, error) {
	identity, found, err := s.repository.FindLoginIdentity(ctx, NormalizeLoginEmail(email))
	if err != nil {
		return LoginResult{}, err
	}
	active := found && identity.UserActive && identity.EmployeeActive
	hash := dummyLoginHash
	if active {
		hash = identity.PasswordHash
	}
	// bcrypt dùng tối đa 72 byte khi tạo hash. Từ chối chuỗi dài ở nhánh credential,
	// đồng thời vẫn thực hiện phép so sánh giả với đầu vào trong giới hạn.
	if len(password) > 72 {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyLoginHash), []byte("mật khẩu vượt giới hạn"))
		return LoginResult{}, ErrLoginVerificationFailed
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil && !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return LoginResult{}, ErrPersistence
	}
	if !active || err != nil {
		return LoginResult{}, ErrLoginVerificationFailed
	}

	var secret SessionSecret
	if _, err := io.ReadFull(s.random, secret[:]); err != nil {
		return LoginResult{}, ErrPersistence
	}
	command := FinalizeLoginCommand{
		VerifiedIdentity: identity,
		NewTokenHash:     sha256.Sum256(secret[:]),
		TTL:              s.ttl,
	}
	if previous != nil {
		hash := SessionHash(sha256.Sum256(previous[:]))
		command.PreviousTokenHash = &hash
	}
	snapshot, err := s.repository.FinalizeLogin(ctx, command)
	if err != nil {
		return LoginResult{}, err
	}
	principal, err := evaluateSession(snapshot, s.clock.Now())
	if err != nil {
		return LoginResult{}, ErrPersistence
	}
	return LoginResult{Principal: principal, Secret: secret, ExpiresAt: snapshot.ExpiresAt}, nil
}
