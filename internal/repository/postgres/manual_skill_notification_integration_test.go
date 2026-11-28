//go:build integration

package postgres_test

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	n "ai-staffing/internal/service/notification"
	"ai-staffing/internal/testutil/authdb"
)

// Các dòng kỹ năng chỉ là fixture cho phép đọc/read, không kiểm writer gửi hoặc duyệt.
func skillNotificationFixture(t *testing.T, f *authdb.Fixture, recipient int, notificationType, resourceType string) int {
	t.Helper()
	var id int
	authdb.Require(t, "thông báo kỹ năng tổng hợp", f.DB.Conn.QueryRow(t.Context(), `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message)
 VALUES($1,$2,$3,701,'Có cập nhật đề nghị.') RETURNING notification_id`, recipient, notificationType, resourceType).Scan(&id))
	return id
}

func TestSkillNotificationMixedOwnRead(t *testing.T) {
	f := authdb.NewFixture(t)
	owner := f.AddAccount(t, authdb.AccountOptions{Email: "director@example.invalid", Roles: []string{"DIRECTOR", "HR", "ADMIN"}})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid", Roles: []string{"DIRECTOR"}})
	if owner.UserID == owner.EmployeeID {
		t.Fatal("fixture phải phân biệt user và employee")
	}
	for _, family := range []string{"PROFILE", "SKILL"} {
		for _, event := range []string{"SUBMITTED", "APPROVED", "REJECTED"} {
			skillNotificationFixture(t, f, owner.UserID, family+"_CHANGE_"+event, family+"_CHANGE_REQUEST")
		}
	}
	s := notificationService(t, f.DB.Conn)
	p, err := s.List(t.Context(), n.ListQuery{ActorID: owner.UserID, PageSize: 20})
	authdb.Require(t, "list trộn", err)
	if p.Total != 6 || p.UnreadCount != 6 || len(p.Items) != 6 {
		t.Fatal("thiếu loại hợp lệ hoặc nhân dòng do role")
	}
	id := p.Items[0].ID
	if err = s.MarkRead(t.Context(), n.ReadCommand{ActorID: other.UserID, NotificationID: id}); !errors.Is(err, n.ErrNotFound) {
		t.Fatal("read người khác phải bị chặn")
	}
	// Quyền tài nguyên đã mất và ID fixture không có request thật vẫn không cản own-read.
	pcExec(t, f, `DELETE FROM user_roles WHERE user_id=$1`, owner.UserID)
	authdb.Require(t, "own-read sau mất role", s.MarkRead(t.Context(), n.ReadCommand{ActorID: owner.UserID, NotificationID: id}))
	var firstRead, againRead time.Time
	authdb.Require(t, "thời điểm read", f.DB.Conn.QueryRow(t.Context(), `SELECT read_at FROM notifications WHERE notification_id=$1`, id).Scan(&firstRead))
	authdb.Require(t, "read lặp", s.MarkRead(t.Context(), n.ReadCommand{ActorID: owner.UserID, NotificationID: id}))
	authdb.Require(t, "thời điểm không đổi", f.DB.Conn.QueryRow(t.Context(), `SELECT read_at FROM notifications WHERE notification_id=$1`, id).Scan(&againRead))
	if !firstRead.Equal(againRead) {
		t.Fatal("read lặp thay thời điểm đầu")
	}
	unread := notificationList(t, s, owner, "unread", 1)
	if unread.Total != 5 || unread.UnreadCount != 5 || len(unread.Items) != 1 {
		t.Fatal("unread filter/count sai")
	}
	empty := notificationList(t, s, owner, "all", 2147483647)
	if empty.Total != 6 || empty.UnreadCount != 5 || len(empty.Items) != 0 {
		t.Fatal("page rỗng mất total hoặc unread_count")
	}
}

func TestSkillNotificationMixedSnapshot(t *testing.T) {
	for _, mutation := range []string{"insert", "read"} {
		t.Run(mutation, func(t *testing.T) {
			f := authdb.NewFixture(t)
			owner := f.AddAccount(t, authdb.AccountOptions{Email: "owner@example.invalid"})
			skillNotificationFixture(t, f, owner.UserID, "PROFILE_CHANGE_APPROVED", "PROFILE_CHANGE_REQUEST")
			id := skillNotificationFixture(t, f, owner.UserID, "SKILL_CHANGE_SUBMITTED", "SKILL_CHANGE_REQUEST")
			writer := notificationService(t, f.DB.Connect(t))
			var once sync.Once
			// Chèn/read đã commit giữa count và items qua connection khác để kiểm snapshot.
			reader := notificationService(t, catalogHookDB{db: f.DB.Connect(t), afterRow: func(sql string) {
				if strings.Contains(sql, "count(*) FILTER") {
					once.Do(func() {
						if mutation == "insert" {
							skillNotificationFixture(t, f, owner.UserID, "SKILL_CHANGE_REJECTED", "SKILL_CHANGE_REQUEST")
						} else {
							authdb.Require(t, "read xen giữa", writer.MarkRead(t.Context(), n.ReadCommand{ActorID: owner.UserID, NotificationID: id}))
						}
					})
				}
			}})
			p, err := reader.List(t.Context(), n.ListQuery{ActorID: owner.UserID, Filter: "unread", PageSize: 20})
			authdb.Require(t, "snapshot trộn", err)
			if p.Total != 2 || p.UnreadCount != 2 || len(p.Items) != 2 || p.Items[0].ReadAt != nil || p.Items[1].ReadAt != nil {
				t.Fatal("count/unread/items khác snapshot")
			}
			after := notificationList(t, writer, owner, "unread", 1)
			want := int64(3)
			if mutation == "read" {
				want = 1
			}
			if after.Total != want || after.UnreadCount != want {
				t.Fatal("mutation snapshot chưa commit thật")
			}
		})
	}
}

func TestSkillNotificationInvalidPersistedPair(t *testing.T) {
	for _, tc := range []struct{ notificationType, resourceType string }{
		{"SKILL_CHANGE_APPROVED", "PROFILE_CHANGE_REQUEST"},
		{"PROFILE_CHANGE_REJECTED", "SKILL_CHANGE_REQUEST"},
		{"SKILL_CHANGE_UNKNOWN", "SKILL_CHANGE_REQUEST"},
	} {
		t.Run(tc.notificationType, func(t *testing.T) {
			f := authdb.NewFixture(t)
			owner := f.AddAccount(t, authdb.AccountOptions{Email: "owner@example.invalid"})
			skillNotificationFixture(t, f, owner.UserID, tc.notificationType, tc.resourceType)
			p, err := notificationService(t, f.DB.Conn).List(t.Context(), n.ListQuery{ActorID: owner.UserID})
			if !errors.Is(err, n.ErrPersistence) || len(p.Items) != 0 {
				t.Fatal("cặp persisted sai không được trả projection")
			}
		})
	}
}

func TestSkillNotificationCommitFault(t *testing.T) {
	for _, mode := range []string{"rollback", "lost"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			owner := f.AddAccount(t, authdb.AccountOptions{Email: "owner@example.invalid"})
			id := skillNotificationFixture(t, f, owner.UserID, "SKILL_CHANGE_APPROVED", "SKILL_CHANGE_REQUEST")
			var db accountBeginner = accountCommitRollbackDB{f.DB.Conn}
			want := n.ErrPersistence
			if mode == "lost" {
				db = catalogHookDB{db: f.DB.Conn, commitLost: true}
				want = n.ErrCommitUnknown
			}
			if !errors.Is(notificationService(t, db).MarkRead(t.Context(), n.ReadCommand{ActorID: owner.UserID, NotificationID: id}), want) {
				t.Fatal("commit fault không được giả thành công")
			}
			p := notificationList(t, notificationService(t, f.DB.Conn), owner, "all", 1)
			if (p.Items[0].ReadAt != nil) != (mode == "lost") {
				t.Fatal("kết quả fault không khớp trạng thái DB")
			}
		})
	}
}
