package postgres

import (
	"context"

	pc "ai-staffing/internal/service/profilechange"
	"github.com/jackc/pgx/v5"
)

func (r *ProfileChangeRepository) Read(ctx context.Context, q pc.DetailQuery) (result pc.Detail, resultErr error) {
	if err := pc.ValidateDetailQuery(q); err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, pc.ErrUnavailable)
	defer cancel()
	defer profileChangeDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, profileChangeError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	actor, err := readProfileActor(ctx, tx, q.ActorID)
	if err != nil {
		return result, err
	}
	// Kiểm quyền trước khi tải patch PII; Admin/Director không có bypass.
	var allowed bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.employee_profile_change_requests r
 JOIN public.employees e ON e.employee_id=r.employee_id WHERE r.request_id=$1 AND (e.user_id=$2 OR $3))`, q.RequestID, q.ActorID, actor.hr).Scan(&allowed)
	if err != nil {
		return result, profileChangeError(err)
	}
	if !allowed {
		return result, pc.ErrNotFound
	}
	d, err := readProfileDetail(ctx, tx, q.RequestID)
	if err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, profileChangeError(err)
	}
	return d, nil
}
func (r *ProfileChangeRepository) List(ctx context.Context, q pc.ListQuery) (result pc.Page, resultErr error) {
	q, err := pc.ValidateList(q)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, pc.ErrUnavailable)
	defer cancel()
	defer profileChangeDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, profileChangeError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	actor, err := readProfileActor(ctx, tx, q.ActorID)
	if err != nil {
		return result, err
	}
	if q.Scope == "review" && !actor.hr {
		return result, pc.ErrForbidden
	}
	p := pc.Page{Items: []pc.Summary{}, Page: q.Page, PageSize: q.PageSize}
	filter := ` FROM public.employee_profile_change_requests r JOIN public.employees e ON e.employee_id=r.employee_id
 WHERE ($1 OR e.user_id=$2) AND ($3='' OR r.status=$3)`
	if err = tx.QueryRow(ctx, `SELECT count(*)`+filter, q.Scope == "review", q.ActorID, q.Status).Scan(&p.Total); err != nil {
		return result, profileChangeError(err)
	}
	rows, err := tx.Query(ctx, `SELECT r.request_id,r.employee_id,e.employee_code,e.fullname,r.status,r.requested_at,r.resolved_at,
 ARRAY(SELECT jsonb_object_keys(r.proposed_values-'schema_version') ORDER BY 1),
 (public.pca_profile_values_valid(r.before_values,false) AND public.pca_profile_values_valid(r.proposed_values,true)
 AND public.pca_same_profile_keys(r.before_values,r.proposed_values) AND r.before_values<>r.proposed_values)`+filter+`
 ORDER BY r.requested_at DESC,r.request_id DESC LIMIT $4 OFFSET $5`, q.Scope == "review", q.ActorID, q.Status, q.PageSize, q.Offset())
	if err != nil {
		return result, profileChangeError(err)
	}
	for rows.Next() {
		var item pc.Summary
		var valid bool
		if err = rows.Scan(&item.RequestID, &item.EmployeeID, &item.EmployeeCode, &item.Fullname, &item.Status, &item.RequestedAt, &item.ResolvedAt, &item.ChangedFields, &valid); err != nil {
			rows.Close()
			return result, profileChangeError(err)
		}
		if !valid || pc.ValidateSummary(item) != nil {
			rows.Close()
			return result, pc.ErrPersistence
		}
		p.Items = append(p.Items, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return result, profileChangeError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return result, profileChangeError(err)
	}
	return p, nil
}
