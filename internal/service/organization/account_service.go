package organization

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

// TemporaryPassword chỉ dành cho orchestration gửi mail sau commit. Format/JSON
// mặc định không được đưa credential vào transport hoặc log.
type TemporaryPassword struct{ value string }

func (TemporaryPassword) String() string     { return "[temporary password redacted]" }
func (p TemporaryPassword) GoString() string { return p.String() }
func (TemporaryPassword) LogValue() slog.Value {
	return slog.StringValue("[temporary password redacted]")
}
func (TemporaryPassword) MarshalJSON() ([]byte, error) {
	return nil, errors.New("credential không được xuất JSON")
}

// RevealForDelivery chỉ dùng trong adapter mail; không dùng trong DTO/log/audit.
func (p TemporaryPassword) RevealForDelivery() string { return p.value }

type CreateAccountResult struct {
	Account  AccountProjection
	password TemporaryPassword
}

func (CreateAccountResult) String() string     { return "[create account result redacted]" }
func (r CreateAccountResult) GoString() string { return r.String() }
func (CreateAccountResult) LogValue() slog.Value {
	return slog.StringValue("[create account result redacted]")
}
func (CreateAccountResult) MarshalJSON() ([]byte, error) {
	return nil, errors.New("kết quả nội bộ không được xuất JSON")
}
func (r CreateAccountResult) PasswordForDelivery() TemporaryPassword { return r.password }

type AccountService struct {
	repo   AccountRepository
	random io.Reader
	hash   func([]byte, int) ([]byte, error)
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo, random: rand.Reader, hash: bcrypt.GenerateFromPassword}
}

func (s *AccountService) Create(ctx context.Context, input CreateAccountInput) (CreateAccountResult, error) {
	input, err := ValidateAccountInput(input)
	if err != nil {
		return CreateAccountResult{}, err
	}
	if err = ctx.Err(); err != nil {
		return CreateAccountResult{}, err
	}
	if s.repo == nil || s.random == nil || s.hash == nil {
		return CreateAccountResult{}, ErrPersistence
	}
	var raw [32]byte
	defer clear(raw[:])
	if _, err = io.ReadFull(s.random, raw[:]); err != nil {
		return CreateAccountResult{}, ErrPersistence
	}
	password := base64.RawURLEncoding.EncodeToString(raw[:])
	hash, err := s.hash([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return CreateAccountResult{}, ErrPersistence
	}
	defer clear(hash)
	if err = ctx.Err(); err != nil {
		return CreateAccountResult{}, err
	}
	account, err := s.repo.CreateAccount(ctx, CreateAccountCommand{Input: input, PasswordHash: string(hash)})
	if err != nil {
		return CreateAccountResult{}, safeAccountError(err)
	}
	return CreateAccountResult{Account: account, password: TemporaryPassword{value: password}}, nil
}

// Port không được wrap lỗi driver vào use case: giữ loại lỗi, bỏ dữ liệu thô.
func safeAccountError(err error) error {
	for _, kind := range []error{ErrCommitUnknown, ErrValidation, ErrUnauthenticated, ErrPasswordChangeRequired, ErrForbidden, ErrEmailConflict, ErrUnavailable, context.Canceled, context.DeadlineExceeded, ErrPersistence} {
		if errors.Is(err, kind) {
			if kind == ErrCommitUnknown {
				if errors.Is(err, ErrUnavailable) {
					return errors.Join(ErrCommitUnknown, ErrUnavailable)
				}
				return errors.Join(ErrCommitUnknown, ErrPersistence)
			}
			return kind
		}
	}
	return ErrPersistence
}
