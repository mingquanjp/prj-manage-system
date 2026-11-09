package postgres

import (
	"context"
	"time"

	"ai-staffing/internal/service/auth"
	"github.com/jackc/pgx/v5"
)

func (r *AuthRepository) ChangeInitialPassword(ctx context.Context, cmd auth.ChangeInitialPasswordCommand) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return authPersistenceError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)

	identity, principal, err := lockAuthIdentity(ctx, tx, cmd.UserID)
	if err != nil {
		return err
	}
	if !identity.UserActive || !identity.EmployeeActive || !principal.MustChangePassword ||
		identity.PasswordHash != cmd.VerifiedPasswordHash {
		return auth.ErrUnauthenticated
	}

	expires, revoked, found, err := lockPasswordChangeSessions(ctx, tx, cmd)
	if err != nil {
		return err
	}
	// Các phiên đều đã khóa, kể cả khi phiên gọi đứng trước phiên khác theo ID.
	now := r.clock.Now()
	if !found || revoked != nil || !now.Before(expires) {
		return auth.ErrUnauthenticated
	}

	_, err = tx.Exec(ctx, `UPDATE public.users SET password_hash=$1, must_change_password=false
 WHERE user_id=$2`, cmd.NewPasswordHash, cmd.UserID)
	if err != nil {
		return authPersistenceError(err)
	}
	_, err = tx.Exec(ctx, `UPDATE public.auth_sessions SET revoked_at=$1
 WHERE user_id=$2 AND revoked_at IS NULL`, now, cmd.UserID)
	if err != nil {
		return authPersistenceError(err)
	}
	return authPersistenceError(tx.Commit(ctx))
}

func lockPasswordChangeSessions(ctx context.Context, tx pgx.Tx, cmd auth.ChangeInitialPasswordCommand) (expires time.Time, revoked *time.Time, found bool, err error) {
	rows, err := tx.Query(ctx, `SELECT token_hash=$1, expires_at, revoked_at
 FROM public.auth_sessions WHERE user_id=$2 ORDER BY session_id FOR UPDATE`, cmd.TokenHash[:], cmd.UserID)
	if err != nil {
		return time.Time{}, nil, false, authPersistenceError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var current bool
		var expiry time.Time
		var revocation *time.Time
		if err := rows.Scan(&current, &expiry, &revocation); err != nil {
			return time.Time{}, nil, false, authPersistenceError(err)
		}
		if current {
			expires, revoked, found = expiry, revocation, true
		}
	}
	return expires, revoked, found, authPersistenceError(rows.Err())
}
