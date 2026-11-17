package organization

import "context"

type MailStatus string

const (
	MailAccepted MailStatus = "accepted"
	MailFailed   MailStatus = "failed"
	MailUnknown  MailStatus = "unknown"
)

// MailSender nhận mật khẩu tạm qua luồng nội bộ sau khi commit được xác nhận.
type MailSender interface {
	SendAccount(ctx context.Context, email string, password TemporaryPassword) (MailStatus, error)
}
