package postgres

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"time"

	org "ai-staffing/internal/service/organization"
	pc "ai-staffing/internal/service/profilechange"
	"github.com/jackc/pgx/v5"
)

type ProfileChangeRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ pc.Repository = (*ProfileChangeRepository)(nil)

func NewProfileChangeRepository(db catalogDB, timeout time.Duration) (*ProfileChangeRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, pc.ErrValidation
	}
	return &ProfileChangeRepository{db, timeout}, nil
}
func profileChangeError(err error) error {
	switch e := accountError(err); e {
	case nil:
		return nil
	case org.ErrUnavailable:
		return pc.ErrUnavailable
	case context.Canceled, context.DeadlineExceeded:
		return e
	default:
		return pc.ErrPersistence
	}
}
func profileChangeCommitError(err error) error {
	e := profileChangeError(err)
	if errors.Is(accountCommitError(err), org.ErrCommitUnknown) {
		return errors.Join(pc.ErrCommitUnknown, e)
	}
	return e
}
func profileChangeDeadline(ctx context.Context, err *error) {
	if errors.Is(*err, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), pc.ErrUnavailable) {
		*err = errors.Join(pc.ErrUnavailable, *err)
	}
}

type profileActor struct {
	userID, employeeID                  int
	active, limited, employeeActive, hr bool
	fullname                            string
	phone, address                      *string
}

func readProfileActor(ctx context.Context, tx pgx.Tx, id int) (profileActor, error) {
	var a profileActor
	err := tx.QueryRow(ctx, `SELECT u.user_id,e.employee_id,u.status='ACTIVE',u.must_change_password,
 e.employment_status='ACTIVE',EXISTS(SELECT 1 FROM public.user_roles r WHERE r.user_id=u.user_id AND r.role_code='HR'),
 e.fullname,e.phone,e.address FROM public.users u JOIN public.employees e ON e.user_id=u.user_id WHERE u.user_id=$1`, id).Scan(&a.userID, &a.employeeID, &a.active, &a.limited, &a.employeeActive, &a.hr, &a.fullname, &a.phone, &a.address)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, pc.ErrUnauthenticated
	}
	if err != nil {
		return a, profileChangeError(err)
	}
	if !a.active || !a.employeeActive {
		return a, pc.ErrUnauthenticated
	}
	if a.limited {
		return a, pc.ErrPasswordChangeRequired
	}
	return a, nil
}

// Chụp actor + tập HR hiện có; khóa toàn bộ users tăng dần rồi employees tăng
// dần. Không mở vòng thêm recipient sau snapshot hoặc khóa actor riêng trước.
func lockProfileParticipants(ctx context.Context, tx pgx.Tx, actorID int) (profileActor, []int, error) {
	rows, err := tx.Query(ctx, `SELECT u.user_id,e.employee_id,
 (u.status='ACTIVE' AND NOT u.must_change_password AND e.employment_status='ACTIVE'
 AND EXISTS(SELECT 1 FROM public.user_roles r WHERE r.user_id=u.user_id AND r.role_code='HR'))
 FROM public.users u JOIN public.employees e ON e.user_id=u.user_id
 WHERE u.user_id=$1 OR (u.status='ACTIVE' AND NOT u.must_change_password AND e.employment_status='ACTIVE'
 AND EXISTS(SELECT 1 FROM public.user_roles r WHERE r.user_id=u.user_id AND r.role_code='HR')) ORDER BY u.user_id`, actorID)
	if err != nil {
		return profileActor{}, nil, profileChangeError(err)
	}
	type pair struct {
		user, employee int
		notify         bool
	}
	var pairs []pair
	for rows.Next() {
		var p pair
		if err = rows.Scan(&p.user, &p.employee, &p.notify); err != nil {
			rows.Close()
			return profileActor{}, nil, profileChangeError(err)
		}
		pairs = append(pairs, p)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return profileActor{}, nil, profileChangeError(err)
	}
	actorEmployee := 0
	for _, p := range pairs {
		var id int
		err = tx.QueryRow(ctx, `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`, p.user).Scan(&id)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return profileActor{}, nil, profileChangeError(err)
		}
		if p.user == actorID {
			actorEmployee = p.employee
		}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].employee < pairs[j].employee })
	for _, p := range pairs {
		var id int
		err = tx.QueryRow(ctx, `SELECT employee_id FROM public.employees WHERE employee_id=$1 FOR UPDATE`, p.employee).Scan(&id)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return profileActor{}, nil, profileChangeError(err)
		}
	}
	a, err := readProfileActor(ctx, tx, actorID)
	if err != nil {
		return a, nil, err
	}
	if a.employeeID != actorEmployee {
		return a, nil, pc.ErrUnauthenticated
	}
	recipients := []int{}
	for _, p := range pairs {
		if !p.notify {
			continue
		}
		var eligible bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.users u JOIN public.employees e ON e.user_id=u.user_id
 WHERE u.user_id=$1 AND e.employee_id=$2 AND u.status='ACTIVE' AND NOT u.must_change_password
 AND e.employment_status='ACTIVE' AND EXISTS(SELECT 1 FROM public.user_roles r WHERE r.user_id=u.user_id AND r.role_code='HR'))`, p.user, p.employee).Scan(&eligible)
		if err != nil {
			return a, nil, profileChangeError(err)
		}
		if eligible {
			recipients = append(recipients, p.user)
		}
	}
	sort.Ints(recipients)
	return a, recipients, nil
}

func (r *ProfileChangeRepository) Submit(ctx context.Context, cmd pc.SubmitCommand) (result pc.Detail, resultErr error) {
	prepared, err := pc.PrepareSubmit(cmd.Input)
	if err != nil {
		return result, err
	}
	if prepared.Hash != cmd.Hash {
		return result, pc.ErrValidation
	}
	cmd = prepared
	in := cmd.Input
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, pc.ErrUnavailable)
	defer cancel()
	defer profileChangeDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return result, profileChangeError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	actor, recipients, err := lockProfileParticipants(ctx, tx, in.ActorID)
	if err != nil {
		return result, err
	}
	if actor.employeeID != in.EmployeeID {
		return result, pc.ErrNotFound
	}
	// Key cũ luôn được tra trước stale/current no-op; replay không tự gửi mutation.
	var oldHash, reference []byte
	var created time.Time
	var expires *time.Time
	err = tx.QueryRow(ctx, `SELECT request_hash,result_reference,created_at,expires_at FROM public.idempotency_records
 WHERE actor_id=$1 AND operation=$2 AND idempotency_key=$3 FOR UPDATE`, in.ActorID, pc.SubmitOperation, in.Key).Scan(&oldHash, &reference, &created, &expires)
	if err == nil {
		if len(oldHash) != 32 || expires == nil || !expires.Equal(created.Add(pc.ReplayWindow)) {
			return result, pc.ErrPersistence
		}
		if subtle.ConstantTimeCompare(oldHash, cmd.Hash[:]) != 1 {
			return result, pc.ErrKeyReused
		}
		var now time.Time
		if err = tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
			return result, profileChangeError(err)
		}
		if !now.Before(*expires) {
			return result, pc.ErrKeyExpired
		}
		id, err := decodeProfileReference(reference)
		if err != nil {
			return result, err
		}
		d, err := readProfileDetail(ctx, tx, id)
		if errors.Is(err, pc.ErrNotFound) {
			return result, pc.ErrPersistence
		}
		if err != nil {
			return result, err
		}
		if d.EmployeeID != in.EmployeeID || d.RequestedBy != in.ActorID || !reflect.DeepEqual(d.Before, in.Before) || !reflect.DeepEqual(d.Proposed, in.Proposed) || !reflect.DeepEqual(d.Explanation, in.Explanation) {
			return result, pc.ErrPersistence
		}
		if err = tx.Commit(ctx); err != nil {
			return result, profileChangeCommitError(err)
		}
		return d, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return result, profileChangeError(err)
	}
	official := pc.Values{"fullname": &actor.fullname, "phone": actor.phone, "address": actor.address}
	for k, v := range in.Before {
		if !reflect.DeepEqual(v, official[k]) {
			return result, pc.ErrStale
		}
	}
	before, err := pc.EncodeStoredValues(in.Before)
	if err != nil {
		return result, err
	}
	proposed, err := pc.EncodeStoredValues(in.Proposed)
	if err != nil {
		return result, err
	}
	var id int
	err = tx.QueryRow(ctx, `INSERT INTO public.employee_profile_change_requests(employee_id,requested_by,before_values,proposed_values,explanation)
 VALUES($1,$2,$3::jsonb,$4::jsonb,$5) RETURNING request_id`, in.EmployeeID, in.ActorID, before, proposed, in.Explanation).Scan(&id)
	if err != nil {
		return result, profileChangeError(err)
	}
	for _, recipient := range recipients {
		_, err = tx.Exec(ctx, `INSERT INTO public.notifications(recipient_id,notification_type,resource_type,resource_id,message)
 VALUES($1,'PROFILE_CHANGE_SUBMITTED','PROFILE_CHANGE_REQUEST',$2,'Có đề nghị cập nhật hồ sơ cần xem xét.')`, recipient, id)
		if err != nil {
			return result, profileChangeError(err)
		}
	}
	ref, err := json.Marshal(profileReference{1, id, "CREATED"})
	if err != nil {
		return result, pc.ErrPersistence
	}
	_, err = tx.Exec(ctx, `WITH stamp AS (SELECT clock_timestamp() AS at)
 INSERT INTO public.idempotency_records(actor_id,operation,idempotency_key,request_hash,result_reference,created_at,expires_at)
 SELECT $1,$2,$3,$4,$5::jsonb,at,at+interval '24 hours' FROM stamp`, in.ActorID, pc.SubmitOperation, in.Key, cmd.Hash[:], ref)
	if err != nil {
		return result, profileChangeError(err)
	}
	d, err := readProfileDetail(ctx, tx, id)
	if err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, profileChangeCommitError(err)
	}
	return d, nil
}

type profileReference struct {
	Version   int    `json:"schema_version"`
	RequestID int    `json:"request_id"`
	Code      string `json:"result_code"`
}

func decodeProfileReference(b []byte) (int, error) {
	var raw map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil || len(raw) != 3 {
		return 0, pc.ErrPersistence
	}
	for _, k := range []string{"schema_version", "request_id", "result_code"} {
		if _, ok := raw[k]; !ok {
			return 0, pc.ErrPersistence
		}
	}
	var ref profileReference
	if json.Unmarshal(b, &ref) != nil || ref.Version != 1 || ref.RequestID < 1 || int64(ref.RequestID) > 2147483647 || ref.Code != "CREATED" {
		return 0, pc.ErrPersistence
	}
	return ref.RequestID, nil
}

func readProfileDetail(ctx context.Context, tx pgx.Tx, id int) (pc.Detail, error) {
	var d pc.Detail
	var before, proposed []byte
	var fullname string
	var phone, address *string
	err := tx.QueryRow(ctx, `SELECT r.request_id,r.employee_id,e.employee_code,e.fullname,r.status,r.requested_at,r.resolved_at,
 r.before_values,r.proposed_values,r.explanation,r.requested_by,r.resolved_by,r.decision_reason,e.fullname,e.phone,e.address
 FROM public.employee_profile_change_requests r JOIN public.employees e ON e.employee_id=r.employee_id WHERE r.request_id=$1`, id).Scan(
		&d.RequestID, &d.EmployeeID, &d.EmployeeCode, &d.Fullname, &d.Status, &d.RequestedAt, &d.ResolvedAt,
		&before, &proposed, &d.Explanation, &d.RequestedBy, &d.ResolvedBy, &d.DecisionReason, &fullname, &phone, &address)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, pc.ErrNotFound
	}
	if err != nil {
		return d, profileChangeError(err)
	}
	d.Before, err = pc.DecodeStoredValues(before, false)
	if err != nil {
		return pc.Detail{}, err
	}
	d.Proposed, err = pc.DecodeStoredValues(proposed, true)
	if err != nil {
		return pc.Detail{}, err
	}
	d.ChangedFields = d.Before.Fields()
	d.Current = pc.Values{}
	official := pc.Values{"fullname": &fullname, "phone": phone, "address": address}
	for _, k := range d.ChangedFields {
		d.Current[k] = official[k]
	}
	if pc.ValidateDetail(d) != nil {
		return pc.Detail{}, pc.ErrPersistence
	}
	return d, nil
}
