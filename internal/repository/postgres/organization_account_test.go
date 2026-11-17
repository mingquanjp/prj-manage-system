package postgres

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestAccountErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"canonical email", &pgconn.PgError{Code: "23505", ConstraintName: "uq_employees_normalized_email"}, org.ErrEmailConflict},
		{"legacy email", &pgconn.PgError{Code: "23505", ConstraintName: "employees_email_key"}, org.ErrEmailConflict},
		{"code collision", &pgconn.PgError{Code: "23505", ConstraintName: "employees_employee_code_key"}, org.ErrPersistence},
		{"position", &pgconn.PgError{Code: "23503", ConstraintName: "employees_position_id_fkey"}, org.ErrValidation},
		{"department", &pgconn.PgError{Code: "23503", ConstraintName: "employees_department_id_fkey"}, org.ErrValidation},
		{"level", &pgconn.PgError{Code: "23503", ConstraintName: "employees_job_level_id_fkey"}, org.ErrValidation},
		{"audit fk", &pgconn.PgError{Code: "23503", ConstraintName: "audit_events_actor_id_fkey"}, org.ErrPersistence},
		{"deadlock", &pgconn.PgError{Code: "40P01"}, org.ErrUnavailable},
		{"lock timeout", &pgconn.PgError{Code: "55P03"}, org.ErrUnavailable},
		{"query timeout", &pgconn.PgError{Code: "57014"}, org.ErrUnavailable},
		{"context canceled", context.Canceled, context.Canceled},
		{"context deadline", context.DeadlineExceeded, context.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if pg, ok := tc.err.(*pgconn.PgError); ok {
				pg.Message = "synthetic private database detail"
				pg.Detail = "synthetic private database detail"
			}
			got := accountError(tc.err)
			if !errors.Is(got, tc.want) || strings.Contains(got.Error(), "private") {
				t.Fatal("phân loại không đúng hoặc lộ lỗi thô")
			}
		})
	}
}

func TestAccountCommitClassification(t *testing.T) {
	for _, tc := range []struct {
		name    string
		err     error
		unknown bool
	}{
		{"known rollback", pgx.ErrTxCommitRollback, false},
		{"known server abort", &pgconn.PgError{Code: "23514", Severity: "ERROR", Message: "private"}, false},
		{"serialization abort", &pgconn.PgError{Code: "40001", Severity: "ERROR", Message: "private"}, false},
		{"deadlock abort", &pgconn.PgError{Code: "40P01", Severity: "ERROR", Message: "private"}, false},
		{"missing severity", &pgconn.PgError{Code: "23514", Message: "private"}, true},
		{"unrecognized server error", &pgconn.PgError{Code: "XX000", Message: "private"}, true},
		{"connection failure reply", &pgconn.PgError{Code: "08006", Message: "private"}, true},
		{"fatal server error", &pgconn.PgError{Code: "23514", Severity: "FATAL", Message: "private"}, true},
		{"commit reply lost", io.ErrUnexpectedEOF, true},
		{"commit deadline", context.DeadlineExceeded, true},
		{"statement completion unknown", &pgconn.PgError{Code: "40003", Message: "private"}, true},
		{"transaction resolution unknown", &pgconn.PgError{Code: "08007", Message: "private"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := accountCommitError(tc.err)
			if errors.Is(got, org.ErrCommitUnknown) != tc.unknown || strings.Contains(got.Error(), "private") {
				t.Fatal("commit outcome sai hoặc lộ lỗi thô")
			}
		})
	}
}

type accountUnavailableDB struct{ calls int }

func (d *accountUnavailableDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.calls++
	return nil, &pgconn.PgError{Code: "53300", Message: "private connection detail"}
}

func TestAccountRepositoryConstructorAndEarlyValidation(t *testing.T) {
	if _, err := NewAccountRepository(nil, time.Second); err == nil {
		t.Fatal("không nhận DB nil")
	}
	db := &accountUnavailableDB{}
	if _, err := NewAccountRepository(db, 0); err == nil {
		t.Fatal("timeout phải dương")
	}
	repo, err := NewAccountRepository(db, time.Second)
	if err != nil {
		t.Fatal("constructor hợp lệ bị từ chối")
	}
	if _, err = repo.CreateAccount(t.Context(), org.CreateAccountCommand{}); !errors.Is(err, org.ErrUnauthenticated) {
		t.Fatal("input sai phải dừng trước DB")
	}
	if db.calls != 0 {
		t.Fatal("input sai vẫn mở transaction")
	}
	input := org.CreateAccountInput{ActorID: 1, Fullname: "Synthetic", CompanyEmail: "synthetic@example.invalid", PositionID: 1}
	if _, err = repo.CreateAccount(t.Context(), org.CreateAccountCommand{Input: input, PasswordHash: "not a hash"}); !errors.Is(err, org.ErrPersistence) {
		t.Fatal("hash nội bộ sai phải bị từ chối")
	}
	if db.calls != 0 {
		t.Fatal("hash sai vẫn mở transaction")
	}
}
