package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
)

type AccountReadRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ org.AccountReadRepository = (*AccountReadRepository)(nil)

func NewAccountReadRepository(db catalogDB, timeout time.Duration) (*AccountReadRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, org.ErrValidation
	}
	return &AccountReadRepository{db, timeout}, nil
}

func checkAccountReader(ctx context.Context, tx pgx.Tx, actorID int) error {
	var active, limited, admin bool
	var employeeActive *bool
	err := tx.QueryRow(ctx, `SELECT u.status='ACTIVE',u.must_change_password,e.employment_status='ACTIVE',
 EXISTS(SELECT 1 FROM public.user_roles r WHERE r.user_id=u.user_id AND r.role_code='ADMIN')
 FROM public.users u LEFT JOIN public.employees e ON e.user_id=u.user_id WHERE u.user_id=$1`, actorID).Scan(&active, &limited, &employeeActive, &admin)
	if errors.Is(err, pgx.ErrNoRows) {
		return org.ErrUnauthenticated
	}
	if err != nil {
		return accountError(err)
	}
	if !active || employeeActive == nil || !*employeeActive {
		return org.ErrUnauthenticated
	}
	if limited {
		return org.ErrPasswordChangeRequired
	}
	if !admin {
		return org.ErrForbidden
	}
	return nil
}

func (r *AccountReadRepository) ListAccounts(ctx context.Context, q org.ListAccountsQuery) (result org.AccountPage, resultErr error) {
	q, err := org.ValidateListAccounts(q)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer func() {
		if errors.Is(resultErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), org.ErrUnavailable) {
			resultErr = org.ErrUnavailable
		}
	}()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, accountError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = checkAccountReader(ctx, tx, q.ActorID); err != nil {
		return result, err
	}
	// User thiếu liên kết employee là dữ liệu nền hỏng theo SQL plan.
	var orphan bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.users u LEFT JOIN public.employees e ON e.user_id=u.user_id WHERE e.employee_id IS NULL)`).Scan(&orphan); err != nil {
		return result, accountError(err)
	}
	if orphan {
		return result, org.ErrPersistence
	}
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(q.Query) + "%"
	filter := ` FROM public.users u JOIN public.employees e ON e.user_id=u.user_id
 WHERE e.employee_code ILIKE $1 ESCAPE '\' OR e.fullname ILIKE $1 ESCAPE '\' OR e.email ILIKE $1 ESCAPE '\'`
	page := org.AccountPage{Items: []org.AccountProjection{}, Page: q.Page, PageSize: q.PageSize}
	if err = tx.QueryRow(ctx, "SELECT count(*)"+filter, pattern).Scan(&page.Total); err != nil {
		return result, accountError(err)
	}
	rows, err := tx.Query(ctx, `SELECT u.user_id,e.employee_id,e.employee_code,e.fullname,e.email,u.status,u.must_change_password,
 ARRAY(SELECT role_code FROM public.user_roles r WHERE r.user_id=u.user_id
 ORDER BY CASE role_code WHEN 'ADMIN' THEN 1 WHEN 'HR' THEN 2 WHEN 'DIRECTOR' THEN 3 ELSE 4 END)`+filter+` ORDER BY u.user_id LIMIT $2 OFFSET $3`, pattern, q.PageSize, q.Offset())
	if err != nil {
		return result, accountError(err)
	}
	for rows.Next() {
		var item org.AccountProjection
		if err = rows.Scan(&item.UserID, &item.EmployeeID, &item.EmployeeCode, &item.Fullname, &item.CompanyEmail, &item.Status, &item.MustChangePassword, &item.SystemRoles); err != nil {
			rows.Close()
			return result, accountError(err)
		}
		page.Items = append(page.Items, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return result, accountError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return result, accountError(err)
	}
	return page, nil
}

func (r *AccountReadRepository) PreviewAccountCode(ctx context.Context, actorID int) (result string, resultErr error) {
	if _, err := org.ValidateListAccounts(org.ListAccountsQuery{ActorID: actorID, Page: 1, PageSize: 1}); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer func() {
		if errors.Is(resultErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), org.ErrUnavailable) {
			resultErr = org.ErrUnavailable
		}
	}()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return "", accountError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = checkAccountReader(ctx, tx, actorID); err != nil {
		return "", err
	}
	reader := &previewSequenceReader{tx: tx}
	code, err := PreviewEmployeeCode(ctx, reader)
	if err != nil {
		if reader.err != nil {
			return "", accountError(reader.err)
		}
		return "", org.ErrPersistence
	}
	if err = tx.Commit(ctx); err != nil {
		return "", accountError(err)
	}
	return code, nil
}

// Giữ loại lỗi QueryRow khi dùng helper preview vốn trả lỗi chung của allocator.
type previewSequenceReader struct {
	tx  sequenceReader
	err error
}

func (r *previewSequenceReader) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return previewSequenceRow{row: r.tx.QueryRow(ctx, sql, args...), capture: &r.err}
}

type previewSequenceRow struct {
	row     pgx.Row
	capture *error
}

func (r previewSequenceRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	*r.capture = err
	return err
}
