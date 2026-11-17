package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type AccountRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ org.AccountRepository = (*AccountRepository)(nil)

// Timeout giới hạn cả thời gian chờ khóa, ghi và xác nhận commit.
func NewAccountRepository(db catalogDB, timeout time.Duration) (*AccountRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, org.ErrValidation
	}
	return &AccountRepository{db: db, timeout: timeout}, nil
}

// CreateAccount không trả account trước khi commit được xác nhận. Mọi raw error
// từ driver bị thay bằng phân loại không chứa SQL, credential hoặc dữ liệu hồ sơ.
func (r *AccountRepository) CreateAccount(ctx context.Context, cmd org.CreateAccountCommand) (result org.AccountProjection, resultErr error) {
	input, err := org.ValidateAccountInput(cmd.Input)
	if err != nil {
		return org.AccountProjection{}, err
	}
	if cost, err := bcrypt.Cost([]byte(cmd.PasswordHash)); err != nil || cost != bcrypt.DefaultCost {
		return org.AccountProjection{}, org.ErrPersistence
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer func() {
		if errors.Is(resultErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), org.ErrUnavailable) {
			resultErr = errors.Join(org.ErrUnavailable, resultErr)
		}
	}()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return org.AccountProjection{}, accountError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = checkAccountActor(ctx, tx, input.ActorID); err != nil {
		return org.AccountProjection{}, err
	}

	// Input mới phải nhất quán với biểu thức uniqueness và lookup auth của DB.
	// Không đổi auth hoặc sửa email để vượt một collation không tương thích.
	var pgEmail string
	if err = tx.QueryRow(ctx, `SELECT lower(btrim($1::text))`, input.CompanyEmail).Scan(&pgEmail); err != nil {
		return org.AccountProjection{}, accountError(err)
	}
	if pgEmail != input.CompanyEmail {
		return org.AccountProjection{}, org.ErrValidation
	}
	// Thứ tự cố định department → position → job_level, giữ KEY SHARE đến commit.
	for _, catalog := range []struct {
		table, column string
		id            *int
	}{
		{"public.departments", "department_id", input.DepartmentID},
		{"public.positions", "position_id", &input.PositionID},
		{"public.job_levels", "job_level_id", input.JobLevelID},
	} {
		if catalog.id == nil {
			continue
		}
		var id int
		err = tx.QueryRow(ctx, "SELECT "+catalog.column+" FROM "+catalog.table+" WHERE "+catalog.column+"=$1 FOR KEY SHARE", *catalog.id).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return org.AccountProjection{}, org.ErrValidation
		}
		if err != nil {
			return org.AccountProjection{}, accountError(err)
		}
	}
	var number int64
	if err = tx.QueryRow(ctx, `SELECT nextval('public.employee_code_seq')`).Scan(&number); err != nil {
		return org.AccountProjection{}, accountError(err)
	}
	code, err := FormatEmployeeCode(number)
	if err != nil {
		return org.AccountProjection{}, org.ErrPersistence
	}
	account := org.AccountProjection{
		EmployeeCode: code, Fullname: input.Fullname, CompanyEmail: input.CompanyEmail,
		Status: "ACTIVE", MustChangePassword: true, SystemRoles: append([]string{}, input.SystemRoles...),
	}
	err = tx.QueryRow(ctx, `INSERT INTO public.users(password_hash,system_role,status,must_change_password)
 VALUES ($1,'EMPLOYEE','ACTIVE',true) RETURNING user_id`, cmd.PasswordHash).Scan(&account.UserID)
	if err != nil {
		return org.AccountProjection{}, accountError(err)
	}
	err = tx.QueryRow(ctx, `INSERT INTO public.employees
 (user_id,employee_code,fullname,email,position_id,department_id,job_level_id,phone,address,employment_status)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'ACTIVE') RETURNING employee_id`, account.UserID,
		code, input.Fullname, input.CompanyEmail, input.PositionID, input.DepartmentID,
		input.JobLevelID, input.Phone, input.Address).Scan(&account.EmployeeID)
	if err != nil {
		return org.AccountProjection{}, accountError(err)
	}
	for _, role := range input.SystemRoles {
		if _, err = tx.Exec(ctx, `INSERT INTO public.user_roles(user_id,role_code) VALUES ($1,$2)`, account.UserID, role); err != nil {
			return org.AccountProjection{}, accountError(err)
		}
	}
	payload, err := json.Marshal(struct {
		SchemaVersion int      `json:"schema_version"`
		UserID        int      `json:"user_id"`
		EmployeeID    int      `json:"employee_id"`
		EmployeeCode  string   `json:"employee_code"`
		PositionID    int      `json:"position_id"`
		DepartmentID  *int     `json:"department_id"`
		JobLevelID    *int     `json:"job_level_id"`
		SystemRoles   []string `json:"system_roles"`
	}{1, account.UserID, account.EmployeeID, account.EmployeeCode, input.PositionID, input.DepartmentID, input.JobLevelID, account.SystemRoles})
	if err != nil {
		return org.AccountProjection{}, org.ErrPersistence
	}
	if _, err = tx.Exec(ctx, `INSERT INTO public.audit_events
 (actor_id,entity_type,entity_id,operation,before_values,after_values)
 VALUES ($1,'employee',$2,'CREATE_ACCOUNT',NULL,$3::jsonb)`, input.ActorID, account.EmployeeID, payload); err != nil {
		return org.AccountProjection{}, accountError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return org.AccountProjection{}, accountCommitError(err)
	}
	return account, nil
}

func checkAccountActor(ctx context.Context, tx pgx.Tx, actorID int) error {
	var active, limited, employeeActive bool
	err := tx.QueryRow(ctx, `SELECT status='ACTIVE',must_change_password FROM public.users WHERE user_id=$1 FOR UPDATE`, actorID).Scan(&active, &limited)
	if errors.Is(err, pgx.ErrNoRows) {
		return org.ErrUnauthenticated
	}
	if err != nil {
		return accountError(err)
	}
	err = tx.QueryRow(ctx, `SELECT employment_status='ACTIVE' FROM public.employees WHERE user_id=$1 FOR UPDATE`, actorID).Scan(&employeeActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return org.ErrUnauthenticated
	}
	if err != nil {
		return accountError(err)
	}
	if !active || !employeeActive {
		return org.ErrUnauthenticated
	}
	if limited {
		return org.ErrPasswordChangeRequired
	}
	var admin bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.user_roles WHERE user_id=$1 AND role_code='ADMIN')`, actorID).Scan(&admin)
	if err != nil {
		return accountError(err)
	}
	if !admin {
		return org.ErrForbidden
	}
	return nil
}

func accountError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "uq_employees_normalized_email" || pgErr.ConstraintName == "employees_email_key" {
				return org.ErrEmailConflict
			}
		case "23503":
			switch pgErr.ConstraintName {
			case "employees_position_id_fkey", "employees_department_id_fkey", "employees_job_level_id_fkey":
				return org.ErrValidation
			}
		case "55P03", "40P01", "57014":
			return org.ErrUnavailable
		}
	}
	switch e := authPersistenceError(err); e {
	case context.Canceled, context.DeadlineExceeded:
		return e
	case auth.ErrUnavailable:
		return org.ErrUnavailable
	default:
		return org.ErrPersistence
	}
}

func accountCommitError(err error) error {
	var pgErr *pgconn.PgError
	// Chỉ các lỗi xác nhận transaction bị hủy mới là thất bại chắc chắn.
	// Có PgError không đủ chứng minh: FATAL, class08 hoặc completion unknown
	// vẫn có thể làm mất xác nhận commit, nên không tự gửi lại mutation.
	if errors.Is(err, pgx.ErrTxCommitRollback) {
		return accountError(err)
	}
	if errors.As(err, &pgErr) {
		severity := pgErr.SeverityUnlocalized
		if severity == "" {
			severity = pgErr.Severity
		}
		if severity == "ERROR" && (strings.HasPrefix(pgErr.Code, "23") || pgErr.Code == "40001" || pgErr.Code == "40P01" || pgErr.Code == "25P02") {
			return accountError(err)
		}
	}
	return errors.Join(org.ErrCommitUnknown, accountError(err))
}
