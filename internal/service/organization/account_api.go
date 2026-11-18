package organization

import "context"

type AccountCreator interface {
	Create(context.Context, CreateAccountInput) (CreateAccountResult, error)
}

type AccountReader interface {
	List(context.Context, ListAccountsQuery) (AccountPage, error)
	Preview(context.Context, int) (string, error)
}

type AccountCreation struct {
	Account AccountProjection
	Mail    MailStatus
}

// AccountAPI điều phối gửi thư sau khi adapter xác nhận commit.
type AccountAPI struct {
	creator AccountCreator
	reader  AccountReader
	mail    MailSender
}

func NewAccountAPI(creator AccountCreator, reader AccountReader, mail MailSender) *AccountAPI {
	return &AccountAPI{creator: creator, reader: reader, mail: mail}
}

func (s *AccountAPI) Create(ctx context.Context, input CreateAccountInput) (AccountCreation, error) {
	if s.creator == nil || s.mail == nil {
		return AccountCreation{}, ErrPersistence
	}
	result, err := s.creator.Create(ctx, input)
	if err != nil {
		return AccountCreation{}, safeAccountError(err)
	}
	status := s.sendCommitted(ctx, result)
	// Trạng thái không hợp lệ từ adapter giữ kết quả tạo và báo chưa xác định.
	switch status {
	case MailAccepted, MailFailed, MailUnknown:
	default:
		status = MailUnknown
	}
	return AccountCreation{Account: result.Account, Mail: status}, nil
}

func (s *AccountAPI) List(ctx context.Context, q ListAccountsQuery) (AccountPage, error) {
	return s.reader.List(ctx, q)
}

func (s *AccountAPI) Preview(ctx context.Context, actorID int) (string, error) {
	return s.reader.Preview(ctx, actorID)
}

// Đã commit: giữ kết quả tạo cả khi adapter gửi thư gặp lỗi ngoài dự kiến.
func (s *AccountAPI) sendCommitted(ctx context.Context, result CreateAccountResult) (status MailStatus) {
	status = MailUnknown
	defer func() {
		if recover() != nil {
			status = MailUnknown
		}
	}()
	value, err := s.mail.SendAccount(ctx, result.Account.CompanyEmail, result.PasswordForDelivery())
	if value == MailAccepted && err != nil {
		return MailUnknown
	}
	return value
}
