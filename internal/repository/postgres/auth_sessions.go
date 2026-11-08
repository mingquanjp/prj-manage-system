package postgres

import (
	"context"
	"errors"
	"time"

	"ai-staffing/internal/service/auth"
	"github.com/jackc/pgx/v5"
)

func (r *AuthRepository) FinalizeLogin(ctx context.Context, cmd auth.FinalizeLoginCommand) (auth.SessionSnapshot, error) {
	// TTL là cấu hình server đã duyệt, sai đầu vào nội bộ phải dừng trước khi ghi.
	if cmd.TTL != 8*time.Hour {
		return auth.SessionSnapshot{}, auth.ErrPersistence
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return auth.SessionSnapshot{}, authPersistenceError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)

	identity, principal, err := lockAuthIdentity(ctx, tx, cmd.VerifiedIdentity.UserID)
	if errors.Is(err, auth.ErrUnauthenticated) {
		err = auth.ErrInvalidCredentials
	}
	if err != nil {
		return auth.SessionSnapshot{}, err
	}
	verified := cmd.VerifiedIdentity
	if identity.EmployeeID != verified.EmployeeID || identity.Email != verified.Email ||
		identity.PasswordHash != verified.PasswordHash || !identity.UserActive || !identity.EmployeeActive {
		return auth.SessionSnapshot{}, auth.ErrInvalidCredentials
	}

	var previousExpiry time.Time
	var previousRevoked *time.Time
	var previousFound bool
	if cmd.PreviousTokenHash != nil {
		err = tx.QueryRow(ctx, `SELECT expires_at, revoked_at FROM public.auth_sessions
 WHERE token_hash=$1 AND user_id=$2 FOR UPDATE`, cmd.PreviousTokenHash[:], identity.UserID).
			Scan(&previousExpiry, &previousRevoked)
		previousFound = err == nil
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return auth.SessionSnapshot{}, authPersistenceError(err)
		}
	}

	// Chỉ lấy giờ sau khi đã chờ đủ khóa, dùng chung cho thu hồi và cấp phiên.
	now := r.clock.Now()
	if previousFound && previousRevoked == nil && now.Before(previousExpiry) {
		_, err = tx.Exec(ctx, `UPDATE public.auth_sessions SET revoked_at=$1
 WHERE token_hash=$2 AND user_id=$3`, now, cmd.PreviousTokenHash[:], identity.UserID)
		if err != nil {
			return auth.SessionSnapshot{}, authPersistenceError(err)
		}
	}
	snapshot := auth.SessionSnapshot{Principal: principal, UserActive: true, EmployeeActive: true}
	// RETURNING giữ chính xác timestamp thực sự lưu với độ chính xác của PostgreSQL.
	err = tx.QueryRow(ctx, `INSERT INTO public.auth_sessions(user_id,token_hash,created_at,expires_at)
 VALUES ($1,$2,$3,$4) RETURNING created_at,expires_at`, identity.UserID, cmd.NewTokenHash[:], now, now.Add(cmd.TTL)).
		Scan(&snapshot.CreatedAt, &snapshot.ExpiresAt)
	if err != nil {
		return auth.SessionSnapshot{}, authPersistenceError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return auth.SessionSnapshot{}, authPersistenceError(err)
	}
	return snapshot, nil
}

func (r *AuthRepository) RevokeCurrentSession(ctx context.Context, hash auth.SessionHash) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return authPersistenceError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)

	// Bước đầu chỉ tìm chủ phiên, chưa chấp thuận hiệu lực từ dữ liệu chưa khóa.
	var userID int
	err = tx.QueryRow(ctx, `SELECT user_id FROM public.auth_sessions WHERE token_hash=$1`, hash[:]).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrUnauthenticated
	}
	if err != nil {
		return authPersistenceError(err)
	}
	identity, _, err := lockAuthIdentity(ctx, tx, userID)
	if err != nil {
		return err
	}
	var expires time.Time
	var revoked *time.Time
	err = tx.QueryRow(ctx, `SELECT expires_at, revoked_at FROM public.auth_sessions
 WHERE token_hash=$1 AND user_id=$2 FOR UPDATE`, hash[:], userID).Scan(&expires, &revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.ErrUnauthenticated
	}
	if err != nil {
		return authPersistenceError(err)
	}
	now := r.clock.Now()
	if !identity.UserActive || !identity.EmployeeActive || revoked != nil || !now.Before(expires) {
		return auth.ErrUnauthenticated
	}
	_, err = tx.Exec(ctx, `UPDATE public.auth_sessions SET revoked_at=$1 WHERE token_hash=$2 AND user_id=$3`, now, hash[:], userID)
	if err != nil {
		return authPersistenceError(err)
	}
	return authPersistenceError(tx.Commit(ctx))
}

// Mọi mutation auth tuân cùng thứ tự user → employee → session để tránh đảo khóa.
func lockAuthIdentity(ctx context.Context, tx pgx.Tx, userID int) (auth.LoginIdentity, auth.Principal, error) {
	identity := auth.LoginIdentity{UserID: userID}
	principal := auth.Principal{UserID: userID}
	err := tx.QueryRow(ctx, `SELECT password_hash,status='ACTIVE',must_change_password
 FROM public.users WHERE user_id=$1 FOR UPDATE`, userID).
		Scan(&identity.PasswordHash, &identity.UserActive, &principal.MustChangePassword)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.LoginIdentity{}, auth.Principal{}, auth.ErrUnauthenticated
	}
	if err != nil {
		return auth.LoginIdentity{}, auth.Principal{}, authPersistenceError(err)
	}
	err = tx.QueryRow(ctx, `SELECT employee_id,email,employment_status='ACTIVE',employee_code,fullname
 FROM public.employees WHERE user_id=$1 FOR UPDATE`, userID).
		Scan(&identity.EmployeeID, &identity.Email, &identity.EmployeeActive, &principal.EmployeeCode, &principal.Fullname)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.LoginIdentity{}, auth.Principal{}, auth.ErrUnauthenticated
	}
	if err != nil {
		return auth.LoginIdentity{}, auth.Principal{}, authPersistenceError(err)
	}
	principal.EmployeeID = identity.EmployeeID
	err = tx.QueryRow(ctx, `SELECT ARRAY(SELECT role_code FROM public.user_roles WHERE user_id=$1 ORDER BY role_code)`, userID).
		Scan(&principal.SystemRoles)
	if err != nil {
		return auth.LoginIdentity{}, auth.Principal{}, authPersistenceError(err)
	}
	return identity, principal, nil
}

func rollbackAuthTransaction(ctx context.Context, tx pgx.Tx) {
	// Context của request có thể đã hủy; cleanup vẫn có thời gian hữu hạn riêng.
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(cleanup)
}
