package postgres

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"

	"ai-staffing/internal/service/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authQueries interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type AuthRepository struct {
	db authQueries
}

var (
	_ auth.IdentityReader = (*AuthRepository)(nil)
	_ authQueries         = (*pgxpool.Pool)(nil)
)

func NewAuthRepository(db authQueries) *AuthRepository {
	return &AuthRepository{db: db}
}

const findLoginIdentitySQL = `
SELECT u.user_id, e.employee_id, e.email, u.password_hash,
       u.status = 'ACTIVE', e.employment_status = 'ACTIVE'
FROM public.employees AS e
JOIN public.users AS u ON u.user_id = e.user_id
WHERE lower(btrim(e.email)) = lower(btrim($1::text))
LIMIT 2`

func (r *AuthRepository) FindLoginIdentity(ctx context.Context, email string) (auth.LoginIdentity, bool, error) {
	rows, err := r.db.Query(ctx, findLoginIdentitySQL, strings.TrimSpace(email))
	if err != nil {
		return auth.LoginIdentity{}, false, authReadError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return auth.LoginIdentity{}, false, authReadError(rows.Err())
	}
	var identity auth.LoginIdentity
	if err := rows.Scan(&identity.UserID, &identity.EmployeeID, &identity.Email, &identity.PasswordHash,
		&identity.UserActive, &identity.EmployeeActive); err != nil {
		return auth.LoginIdentity{}, false, authReadError(err)
	}
	// Bản ghi thứ hai đủ để phát hiện xung đột, kể cả khi một tài khoản bị khóa.
	if rows.Next() {
		return auth.LoginIdentity{}, false, auth.ErrIdentityConflict
	}
	if err := rows.Err(); err != nil {
		return auth.LoginIdentity{}, false, authReadError(err)
	}
	return identity, true, nil
}

const readSessionSQL = `
SELECT u.user_id, e.employee_id, e.employee_code, e.fullname, u.must_change_password,
       ARRAY(SELECT ur.role_code FROM public.user_roles AS ur
             WHERE ur.user_id = u.user_id ORDER BY ur.role_code),
       s.created_at, s.expires_at, s.revoked_at,
       u.status = 'ACTIVE', e.employment_status = 'ACTIVE'
FROM public.auth_sessions AS s
JOIN public.users AS u ON u.user_id = s.user_id
JOIN public.employees AS e ON e.user_id = u.user_id
WHERE s.token_hash = $1`

func (r *AuthRepository) ReadSession(ctx context.Context, hash auth.SessionHash) (auth.SessionSnapshot, bool, error) {
	var snapshot auth.SessionSnapshot
	p := &snapshot.Principal
	err := r.db.QueryRow(ctx, readSessionSQL, hash[:]).Scan(
		&p.UserID, &p.EmployeeID, &p.EmployeeCode, &p.Fullname, &p.MustChangePassword, &p.SystemRoles,
		&snapshot.CreatedAt, &snapshot.ExpiresAt, &snapshot.RevokedAt,
		&snapshot.UserActive, &snapshot.EmployeeActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.SessionSnapshot{}, false, nil
	}
	if err != nil {
		return auth.SessionSnapshot{}, false, authReadError(err)
	}
	return snapshot, true, nil
}

// Trả lỗi ổn định để tầng trên phân loại mà không mang theo SQL hay dữ liệu kết nối.
func authReadError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return auth.ErrUnavailable
	}
	var connectErr *pgconn.ConnectError
	var networkErr *net.OpError
	if errors.As(err, &connectErr) || errors.As(err, &networkErr) {
		return auth.ErrUnavailable
	}
	var dbErr *pgconn.PgError
	if errors.As(err, &dbErr) {
		switch dbErr.Code {
		case "57P01", "57P02", "57P03": // Tắt server, server gặp sự cố hoặc chưa nhận kết nối.
			return auth.ErrUnavailable
		}
		if strings.HasPrefix(dbErr.Code, "08") {
			return auth.ErrUnavailable
		}
	}
	return auth.ErrPersistence
}
