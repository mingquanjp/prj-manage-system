//go:build integration

package postgres_test

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	n "ai-staffing/internal/service/notification"
	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
)

func notificationService(t *testing.T, db accountBeginner) *n.Service {
	t.Helper()
	r, e := postgres.NewNotificationRepository(db, 10*time.Second)
	authdb.Require(t, "adapter thông báo", e)
	return n.NewService(r)
}
func notificationList(t *testing.T, s *n.Service, a authdb.Account, filter string, page int) n.Page {
	t.Helper()
	p, e := s.List(t.Context(), n.ListQuery{ActorID: a.UserID, Filter: filter, Page: page, PageSize: 1})
	authdb.Require(t, "list chính chủ", e)
	return p
}
func TestNotificationOwnSnapshotFilterAndRead(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN", "DIRECTOR"}})
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid", Roles: []string{"ADMIN"}})
	if hr.UserID == hr.EmployeeID {
		t.Fatal("fixture ID chưa tách")
	}
	first := pcPending(t, f, actor, 1)
	second := pcPending(t, f, actor, 2)
	s := notificationService(t, f.DB.Conn)
	p := notificationList(t, s, hr, "all", 1)
	if p.Total != 2 || p.UnreadCount != 2 || len(p.Items) != 1 || p.Items[0].ResourceID != second.RequestID {
		t.Fatal("list/filter/thứ tự sai")
	}
	id := p.Items[0].ID
	empty := notificationList(t, s, other, "all", 1)
	if empty.Total != 0 || empty.UnreadCount != 0 || len(empty.Items) != 0 {
		t.Fatal("Admin thấy notification người khác")
	}
	if e := s.MarkRead(t.Context(), n.ReadCommand{ActorID: other.UserID, NotificationID: id}); !errors.Is(e, n.ErrNotFound) {
		t.Fatal("read ngoài own chưa 404")
	}
	// Mất quyền HR không mất quyền đọc/đánh dấu thông báo đã nhận.
	pcExec(t, f, `DELETE FROM user_roles WHERE user_id=$1`, hr.UserID)
	if _, e := pcService(t, f.DB.Conn).Read(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: second.RequestID}); !errors.Is(e, pc.ErrNotFound) {
		t.Fatal("quyền tài nguyên còn mở")
	}
	authdb.Require(t, "read own không phụ thuộc quyền resource", s.MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: id}))
	all := notificationList(t, s, hr, "all", 1)
	stamp := all.Items[0].ReadAt
	authdb.Require(t, "read lần hai", s.MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: id}))
	again := notificationList(t, s, hr, "all", 1)
	if stamp == nil || again.Items[0].ReadAt == nil || !stamp.Equal(*again.Items[0].ReadAt) {
		t.Fatal("đổi thời điểm read đầu")
	}
	unread := notificationList(t, s, hr, "unread", 1)
	if unread.Total != 1 || unread.UnreadCount != 1 || unread.Items[0].ResourceID != first.RequestID {
		t.Fatal("unread sai")
	}
	empty = notificationList(t, s, hr, "unread", 2147483647)
	if len(empty.Items) != 0 || empty.Total != 1 || empty.UnreadCount != 1 {
		t.Fatal("page rỗng mất count/overflow")
	}
	if e := s.MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: 2147483647}); !errors.Is(e, n.ErrNotFound) {
		t.Fatal("missing chưa 404")
	}
}
func TestNotificationListConcurrentSnapshot(t *testing.T) {
	for _, mutation := range []string{"insert", "read"} {
		t.Run(mutation, func(t *testing.T) {
			f := authdb.NewFixture(t)
			hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
			actor := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			pcPending(t, f, actor, 1)
			writer := notificationService(t, f.DB.Connect(t))
			initial := notificationList(t, writer, hr, "all", 1)
			var once sync.Once
			// Mutation giữa count và items qua connection khác, không sleep.
			reader := notificationService(t, catalogHookDB{db: f.DB.Connect(t), afterRow: func(sql string) {
				if strings.Contains(sql, "count(*) FILTER") {
					once.Do(func() {
						if mutation == "insert" {
							pcPending(t, f, actor, 2)
						} else {
							authdb.Require(t, "read xen snapshot", writer.MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: initial.Items[0].ID}))
						}
					})
				}
			}})
			p := notificationList(t, reader, hr, "unread", 1)
			if p.Total != 1 || p.UnreadCount != 1 || len(p.Items) != 1 || p.Items[0].ReadAt != nil {
				t.Fatal("count/items khác snapshot")
			}
			current := notificationList(t, writer, hr, "unread", 1)
			want := int64(2)
			if mutation == "read" {
				want = 0
			}
			if current.Total != want {
				t.Fatal("mutation không commit thật")
			}
		})
	}
}
func TestNotificationConcurrentReadAndActorRecheck(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	pcPending(t, f, actor, 1)
	id := notificationList(t, notificationService(t, f.DB.Conn), hr, "all", 1).Items[0].ID
	first, second := f.DB.Connect(t), f.DB.Connect(t)
	reached, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	var firstStamp time.Time
	s := notificationService(t, catalogHookDB{db: first, afterRow: func(sql string) {
		if strings.Contains(sql, "UPDATE public.notifications") {
			once.Do(func() {
				authdb.Require(t, "stamp ngay lần UPDATE đầu", first.QueryRow(t.Context(), `SELECT read_at FROM notifications WHERE notification_id=$1`, id).Scan(&firstStamp))
				close(reached)
				select {
				case <-release:
				case <-t.Context().Done():
				}
			})
		}
	}})
	done := make(chan error, 2)
	go func() { done <- s.MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: id}) }()
	select {
	case <-reached:
	case <-t.Context().Done():
		t.Fatal("barrier timeout")
	}
	go func() {
		done <- notificationService(t, second).MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: id})
	}()
	// DB lock evidence chờ user của lượt đầu, không chỉ hai goroutine tuần tự.
	monitor, e := f.DB.Conn.Begin(t.Context())
	authdb.Require(t, "monitor khóa", e)
	waitForAuthLock(t, t.Context(), monitor, first.PgConn().PID(), second.PgConn().PID())
	authdb.Require(t, "đóng monitor", monitor.Rollback(t.Context()))
	close(release)
	authdb.Require(t, "read đồng thời 1", <-done)
	authdb.Require(t, "read đồng thời 2", <-done)
	var stamp time.Time
	authdb.Require(t, "stamp DB", f.DB.Conn.QueryRow(t.Context(), `SELECT read_at FROM notifications WHERE notification_id=$1`, id).Scan(&stamp))
	if firstStamp.IsZero() || !stamp.Equal(firstStamp) {
		t.Fatal("lượt concurrent thứ hai thay timestamp của lượt đầu")
	}
	for _, mode := range []string{"limited", "locked", "inactive"} {
		pcExec(t, f, `UPDATE users SET status='ACTIVE',must_change_password=false WHERE user_id=$1`, hr.UserID)
		pcExec(t, f, `UPDATE employees SET employment_status='ACTIVE' WHERE user_id=$1`, hr.UserID)
		blocker, e := first.Begin(t.Context())
		authdb.Require(t, "actor lock", e)
		_, e = blocker.Exec(t.Context(), `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, hr.UserID)
		authdb.Require(t, "actor lock row", e)
		go func() {
			done <- notificationService(t, second).MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: id})
		}()
		waitForAuthLock(t, t.Context(), blocker, first.PgConn().PID(), second.PgConn().PID())
		sql := `UPDATE users SET must_change_password=true WHERE user_id=$1`
		want := n.ErrPasswordChangeRequired
		if mode == "locked" {
			sql = `UPDATE users SET status='LOCKED' WHERE user_id=$1`
			want = n.ErrUnauthenticated
		}
		if mode == "inactive" {
			sql = `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`
			want = n.ErrUnauthenticated
		}
		_, e = blocker.Exec(t.Context(), sql, hr.UserID)
		authdb.Require(t, "actor changed", e)
		authdb.Require(t, "commit actor", blocker.Commit(t.Context()))
		if !errors.Is(<-done, want) {
			t.Fatal("không recheck actor sau chờ khóa")
		}
	}
	var after time.Time
	authdb.Require(t, "stamp giữ nguyên", f.DB.Conn.QueryRow(t.Context(), `SELECT read_at FROM notifications WHERE notification_id=$1`, id).Scan(&after))
	if !stamp.Equal(after) {
		t.Fatal("concurrent read thay timestamp")
	}
}
func TestNotificationCommitFaults(t *testing.T) {
	for _, mode := range []string{"rollback", "lost"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
			actor := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			pcPending(t, f, actor, 1)
			s := notificationService(t, f.DB.Conn)
			id := notificationList(t, s, hr, "all", 1).Items[0].ID
			var db accountBeginner = accountCommitRollbackDB{f.DB.Conn}
			want := n.ErrPersistence
			if mode == "lost" {
				db = catalogHookDB{db: f.DB.Conn, commitLost: true}
				want = n.ErrCommitUnknown
			}
			if !errors.Is(notificationService(t, db).MarkRead(t.Context(), n.ReadCommand{ActorID: hr.UserID, NotificationID: id}), want) {
				t.Fatal("commit fault báo sai")
			}
			p := notificationList(t, s, hr, "all", 1)
			if (p.Items[0].ReadAt != nil) != (mode == "lost") {
				t.Fatal("fault không khớp DB thật")
			}
		})
	}
}
