package organization

import (
	"context"
	"errors"
	"testing"
)

type creatorFunc func(context.Context, CreateAccountInput) (CreateAccountResult, error)

func (f creatorFunc) Create(ctx context.Context, input CreateAccountInput) (CreateAccountResult, error) {
	return f(ctx, input)
}

type mailFunc func(context.Context, string, TemporaryPassword) (MailStatus, error)

func (f mailFunc) SendAccount(ctx context.Context, email string, password TemporaryPassword) (MailStatus, error) {
	return f(ctx, email, password)
}

func TestAccountAPISendsOnlyAfterConfirmedCreation(t *testing.T) {
	for _, creationErr := range []error{nil, ErrValidation, ErrEmailConflict, ErrForbidden, ErrUnavailable, errors.Join(ErrCommitUnknown, ErrPersistence)} {
		for _, status := range []MailStatus{MailAccepted, MailFailed, MailUnknown, ""} {
			calls := 0
			committed := false
			api := NewAccountAPI(creatorFunc(func(context.Context, CreateAccountInput) (CreateAccountResult, error) {
				if creationErr != nil {
					return CreateAccountResult{}, creationErr
				}
				committed = true
				return CreateAccountResult{Account: AccountProjection{UserID: 23, CompanyEmail: "test@example.invalid"}, password: TemporaryPassword{"internal-only"}}, nil
			}), nil, mailFunc(func(ctx context.Context, email string, password TemporaryPassword) (MailStatus, error) {
				calls++
				if !committed || email != "test@example.invalid" || password.value != "internal-only" {
					t.Fatal("Thư sai dữ liệu hoặc được gọi trước kết quả tạo")
				}
				if status == MailAccepted {
					return status, nil
				}
				return status, errors.New("private provider detail")
			}))
			result, err := api.Create(t.Context(), CreateAccountInput{})
			if creationErr != nil {
				if !errors.Is(err, creationErr) || calls != 0 || result.Account.UserID != 0 {
					// Marker lỗi join phải được bảo toàn qua safeAccountError.
					if !(errors.Is(creationErr, ErrCommitUnknown) && errors.Is(err, ErrCommitUnknown) && calls == 0 && result.Account.UserID == 0) {
						t.Fatal("Lỗi tạo đã bị che hoặc vẫn gửi thư")
					}
				}
				continue
			}
			expected := status
			if expected == "" {
				expected = MailUnknown
			}
			if err != nil || calls != 1 || result.Account.UserID != 23 || result.Mail != expected {
				t.Fatal("Không giữ đúng kết quả tạo sau commit")
			}
		}
	}
}

func TestAccountAPIPreservesCommitAfterMailPanicOrContradiction(t *testing.T) {
	for _, panicMail := range []bool{false, true} {
		calls := 0
		api := NewAccountAPI(creatorFunc(func(context.Context, CreateAccountInput) (CreateAccountResult, error) {
			return CreateAccountResult{Account: AccountProjection{UserID: 4}}, nil
		}), nil, mailFunc(func(context.Context, string, TemporaryPassword) (MailStatus, error) {
			calls++
			if panicMail {
				panic("private provider detail")
			}
			return MailAccepted, errors.New("private provider detail")
		}))
		result, err := api.Create(t.Context(), CreateAccountInput{})
		if err != nil || result.Account.UserID != 4 || result.Mail != MailUnknown || calls != 1 {
			t.Fatal("Đã mất kết quả commit sau lỗi mail")
		}
	}
}
