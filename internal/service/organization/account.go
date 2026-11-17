package organization

import (
	"context"
	"errors"
	"log/slog"
)

var ErrEmailConflict = errors.New("organization: email conflict")

// ActorID là nguồn nội bộ đã xác thực, không phải field HTTP của hồ sơ mới.
type CreateAccountInput struct {
	ActorID      int
	Fullname     string
	CompanyEmail string
	PositionID   int
	DepartmentID *int
	JobLevelID   *int
	Phone        *string
	Address      *string
	SystemRoles  []string
}

// AccountProjection không chứa credential hoặc PII ngoài allowlist tài khoản.
type AccountProjection struct {
	UserID             int
	EmployeeID         int
	EmployeeCode       string
	Fullname           string
	CompanyEmail       string
	Status             string
	SystemRoles        []string
	MustChangePassword bool
}

// CreateAccountCommand chỉ dành cho port nội bộ. Hash không được serialize/log.
type CreateAccountCommand struct {
	Input        CreateAccountInput
	PasswordHash string `json:"-"`
}

func (CreateAccountCommand) String() string     { return "[create account command redacted]" }
func (c CreateAccountCommand) GoString() string { return c.String() }
func (CreateAccountCommand) LogValue() slog.Value {
	return slog.StringValue("[create account command redacted]")
}
func (CreateAccountCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("command nội bộ không được xuất JSON")
}

// CreateAccount chỉ trả projection khi DB đã xác nhận commit. ErrCommitUnknown
// có nghĩa phải đối soát, không gửi mail hoặc tự retry mutation.
type AccountRepository interface {
	CreateAccount(context.Context, CreateAccountCommand) (AccountProjection, error)
}
