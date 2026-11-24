package postgres

import (
	n "ai-staffing/internal/service/notification"
	org "ai-staffing/internal/service/organization"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

type NotificationRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ n.Repository = (*NotificationRepository)(nil)

func NewNotificationRepository(db catalogDB, timeout time.Duration) (*NotificationRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, n.ErrValidation
	}
	return &NotificationRepository{db, timeout}, nil
}
func notificationError(err error) error {
	switch e := accountError(err); e {
	case nil:
		return nil
	case org.ErrUnavailable:
		return n.ErrUnavailable
	case context.Canceled, context.DeadlineExceeded:
		return e
	default:
		return n.ErrPersistence
	}
}
func notificationDeadline(ctx context.Context, err *error) {
	if errors.Is(*err, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), n.ErrUnavailable) {
		*err = errors.Join(n.ErrUnavailable, *err)
	}
}

// Mutation khóa user rồi employee trước notification, cùng chiều với writer hồ sơ.
// List dùng snapshot chỉ đọc; không khóa và không thay read_at khi mở khung.
func notificationActor(ctx context.Context, tx pgx.Tx, id int, lock bool) error {
	if lock {
		var user int
		err := tx.QueryRow(ctx, `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`, id).Scan(&user)
		if errors.Is(err, pgx.ErrNoRows) {
			return n.ErrUnauthenticated
		}
		if err != nil {
			return notificationError(err)
		}
	}
	query := `SELECT u.status='ACTIVE',u.must_change_password,e.employment_status='ACTIVE'
 FROM public.users u JOIN public.employees e ON e.user_id=u.user_id WHERE u.user_id=$1`
	if lock {
		query += ` FOR UPDATE OF e`
	}
	var active, limited, employeeActive bool
	err := tx.QueryRow(ctx, query, id).Scan(&active, &limited, &employeeActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return n.ErrUnauthenticated
	}
	if err != nil {
		return notificationError(err)
	}
	if !active || !employeeActive {
		return n.ErrUnauthenticated
	}
	if limited {
		return n.ErrPasswordChangeRequired
	}
	return nil
}
func (r *NotificationRepository) List(ctx context.Context, q n.ListQuery) (result n.Page, resultErr error) {
	q, err := n.ValidateList(q)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, n.ErrUnavailable)
	defer cancel()
	defer notificationDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, notificationError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = notificationActor(ctx, tx, q.ActorID, false); err != nil {
		return result, err
	}
	p := n.Page{Items: []n.Item{}, Page: q.Page, PageSize: q.PageSize}
	err = tx.QueryRow(ctx, `SELECT count(*) FILTER (WHERE $2='all' OR read_at IS NULL),count(*) FILTER (WHERE read_at IS NULL) FROM public.notifications WHERE recipient_id=$1`, q.ActorID, q.Filter).Scan(&p.Total, &p.UnreadCount)
	if err != nil {
		return result, notificationError(err)
	}
	rows, err := tx.Query(ctx, `SELECT notification_id,notification_type,resource_type,resource_id,message,created_at,read_at FROM public.notifications WHERE recipient_id=$1 AND ($2='all' OR read_at IS NULL) ORDER BY created_at DESC,notification_id DESC LIMIT $3 OFFSET $4`, q.ActorID, q.Filter, q.PageSize, q.Offset())
	if err != nil {
		return result, notificationError(err)
	}
	for rows.Next() {
		var item n.Item
		if err = rows.Scan(&item.ID, &item.Type, &item.ResourceType, &item.ResourceID, &item.Message, &item.CreatedAt, &item.ReadAt); err != nil {
			rows.Close()
			return result, notificationError(err)
		}
		p.Items = append(p.Items, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return result, notificationError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return result, notificationError(err)
	}
	return p, nil
}
func (r *NotificationRepository) MarkRead(ctx context.Context, q n.ReadCommand) (resultErr error) {
	if err := n.ValidateRead(q); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, n.ErrUnavailable)
	defer cancel()
	defer notificationDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return notificationError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = notificationActor(ctx, tx, q.ActorID, true); err != nil {
		return err
	}
	var id int
	err = tx.QueryRow(ctx, `UPDATE public.notifications SET read_at=COALESCE(read_at,clock_timestamp()) WHERE notification_id=$1 AND recipient_id=$2 RETURNING notification_id`, q.NotificationID, q.ActorID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return n.ErrNotFound
	}
	if err != nil {
		return notificationError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		if errors.Is(accountCommitError(err), org.ErrCommitUnknown) {
			return errors.Join(n.ErrCommitUnknown, notificationError(err))
		}
		return notificationError(err)
	}
	return nil
}
