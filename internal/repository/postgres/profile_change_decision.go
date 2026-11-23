package postgres

import (
	"context"
	"errors"
	"reflect"
	"sort"

	pc "ai-staffing/internal/service/profilechange"
	"github.com/jackc/pgx/v5"
)

func (r *ProfileChangeRepository) Approve(ctx context.Context, q pc.DetailQuery) (pc.Detail, error) {
	if err := pc.ValidateDetailQuery(q); err != nil {
		return pc.Detail{}, err
	}
	return r.decide(ctx, q, nil)
}

func (r *ProfileChangeRepository) Reject(ctx context.Context, in pc.RejectInput) (pc.Detail, error) {
	in, err := pc.PrepareReject(in)
	if err != nil {
		return pc.Detail{}, err
	}
	return r.decide(ctx, pc.DetailQuery{ActorID: in.ActorID, RequestID: in.RequestID}, &in.Reason)
}

// Đọc trước chỉ để lập bộ ID. Sau khóa phải kiểm lại liên kết, không tự bổ
// sung user/employee mới vào cuối bộ khóa khi liên kết đã thay đổi.
func lockProfileDecision(ctx context.Context, tx pgx.Tx, q pc.DetailQuery) (pc.Detail, int, bool, error) {
	var actorEmployee int
	err := tx.QueryRow(ctx, `SELECT employee_id FROM public.employees WHERE user_id=$1`, q.ActorID).Scan(&actorEmployee)
	if errors.Is(err, pgx.ErrNoRows) {
		return pc.Detail{}, 0, false, pc.ErrUnauthenticated
	}
	if err != nil {
		return pc.Detail{}, 0, false, profileChangeError(err)
	}
	var targetEmployee, targetUser int
	err = tx.QueryRow(ctx, `SELECT r.employee_id,e.user_id FROM public.employee_profile_change_requests r
 JOIN public.employees e ON e.employee_id=r.employee_id WHERE r.request_id=$1`, q.RequestID).Scan(&targetEmployee, &targetUser)
	missing := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !missing {
		return pc.Detail{}, 0, false, profileChangeError(err)
	}
	users := []int{q.ActorID}
	employees := []int{actorEmployee}
	if !missing {
		users = append(users, targetUser)
		employees = append(employees, targetEmployee)
	}
	sort.Ints(users)
	sort.Ints(employees)
	for _, group := range []struct {
		ids   []int
		query string
	}{
		{users, `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`},
		{employees, `SELECT employee_id FROM public.employees WHERE employee_id=$1 FOR UPDATE`},
	} {
		for i, id := range group.ids {
			if i > 0 && id == group.ids[i-1] {
				continue
			}
			var locked int
			if err = tx.QueryRow(ctx, group.query, id).Scan(&locked); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return pc.Detail{}, 0, false, pc.ErrPersistence
				}
				return pc.Detail{}, 0, false, profileChangeError(err)
			}
		}
	}
	a, err := readProfileActor(ctx, tx, q.ActorID)
	if err != nil {
		return pc.Detail{}, 0, false, err
	}
	if a.employeeID != actorEmployee {
		return pc.Detail{}, 0, false, pc.ErrUnauthenticated
	}
	if !a.hr {
		return pc.Detail{}, 0, false, pc.ErrForbidden
	}
	if missing {
		return pc.Detail{}, 0, false, pc.ErrNotFound
	}
	var lockedEmployee int
	err = tx.QueryRow(ctx, `SELECT employee_id FROM public.employee_profile_change_requests WHERE request_id=$1 FOR UPDATE`, q.RequestID).Scan(&lockedEmployee)
	if errors.Is(err, pgx.ErrNoRows) {
		return pc.Detail{}, 0, false, pc.ErrNotFound
	}
	if err != nil {
		return pc.Detail{}, 0, false, profileChangeError(err)
	}
	if lockedEmployee != targetEmployee {
		return pc.Detail{}, 0, false, pc.ErrPersistence
	}
	var currentUser int
	var active bool
	err = tx.QueryRow(ctx, `SELECT user_id,employment_status='ACTIVE' FROM public.employees WHERE employee_id=$1`, targetEmployee).Scan(&currentUser, &active)
	if err != nil {
		return pc.Detail{}, 0, false, profileChangeError(err)
	}
	if currentUser != targetUser {
		return pc.Detail{}, 0, false, pc.ErrPersistence
	}
	d, err := readProfileDetail(ctx, tx, q.RequestID)
	if err != nil {
		return pc.Detail{}, 0, false, err
	}
	if d.RequestedBy != targetUser {
		return pc.Detail{}, 0, false, pc.ErrPersistence
	}
	if d.Status != "PENDING" {
		return pc.Detail{}, 0, false, pc.ErrAlreadyResolved
	}
	return d, targetUser, active, nil
}

func (r *ProfileChangeRepository) decide(ctx context.Context, q pc.DetailQuery, reason *string) (result pc.Detail, resultErr error) {
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, pc.ErrUnavailable)
	defer cancel()
	defer profileChangeDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return result, profileChangeError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	d, owner, active, err := lockProfileDecision(ctx, tx, q)
	if err != nil {
		return result, err
	}
	status, notification, message := "REJECTED", "PROFILE_CHANGE_REJECTED", "Đề nghị cập nhật hồ sơ đã bị từ chối."
	if reason == nil {
		if !active {
			return result, pc.ErrEmployeeInactive
		}
		for k, before := range d.Before {
			if !reflect.DeepEqual(before, d.Current[k]) {
				return result, pc.ErrStale
			}
		}
		// Cờ hiện diện tách khỏi NULL: phone=NULL là xóa, khóa vắng là giữ.
		_, fullname := d.Proposed["fullname"]
		_, phone := d.Proposed["phone"]
		_, address := d.Proposed["address"]
		tag, e := tx.Exec(ctx, `UPDATE public.employees SET
 fullname=CASE WHEN $2 THEN $3::text ELSE fullname END,
 phone=CASE WHEN $4 THEN $5::text ELSE phone END,
 address=CASE WHEN $6 THEN $7::text ELSE address END WHERE employee_id=$1`,
			d.EmployeeID, fullname, d.Proposed["fullname"], phone, d.Proposed["phone"], address, d.Proposed["address"])
		if e != nil {
			return result, profileChangeError(e)
		}
		if tag.RowsAffected() != 1 {
			return result, pc.ErrPersistence
		}
		status, notification, message = "APPROVED", "PROFILE_CHANGE_APPROVED", "Đề nghị cập nhật hồ sơ đã được chấp thuận."
	}
	tag, err := tx.Exec(ctx, `UPDATE public.employee_profile_change_requests
 SET status=$2,resolved_by=$3,resolved_at=clock_timestamp(),decision_reason=$4
 WHERE request_id=$1 AND status='PENDING'`, q.RequestID, status, q.ActorID, reason)
	if err != nil {
		return result, profileChangeError(err)
	}
	if tag.RowsAffected() != 1 {
		return result, pc.ErrAlreadyResolved
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.notifications(recipient_id,notification_type,resource_type,resource_id,message)
 VALUES($1,$2,'PROFILE_CHANGE_REQUEST',$3,$4)`, owner, notification, q.RequestID, message)
	if err != nil {
		return result, profileChangeError(err)
	}
	updated, err := readProfileDetail(ctx, tx, q.RequestID)
	if err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, profileChangeCommitError(err)
	}
	return updated, nil
}
