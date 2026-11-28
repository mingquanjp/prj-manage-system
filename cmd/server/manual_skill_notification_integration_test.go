//go:build integration

package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/testutil/authdb"
)

func TestSkillNotificationApplicationPostgres(t *testing.T) {
	a := newAuthApp(t)
	f := a.fixture
	owner := f.AddAccount(t, authdb.AccountOptions{Email: "skills-director@example.invalid", Roles: []string{"DIRECTOR", "HR", "ADMIN"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "skills-other@example.invalid", Roles: []string{"ADMIN"}})
	if owner.UserID == owner.EmployeeID {
		t.Fatal("fixture notification phải dùng user_id khác employee_id")
	}
	var lastID int
	// Fixture chỉ kiểm API đọc/read; các resource ID này không đại diện writer kỹ năng thật.
	for _, family := range []string{"PROFILE", "SKILL"} {
		for _, event := range []string{"SUBMITTED", "APPROVED", "REJECTED"} {
			authdb.Require(t, "fixture thông báo trộn", f.DB.Conn.QueryRow(t.Context(), `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message)
 VALUES($1,$2,$3,701,'Có cập nhật đề nghị.') RETURNING notification_id`, owner.UserID, family+"_CHANGE_"+event, family+"_CHANGE_REQUEST").Scan(&lastID))
		}
	}
	cookie := a.login(t, "skills-director@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	_, body := catalogCall(t, a, "GET", "/api/notifications", cookie, "", nil, 200)
	var page struct {
		Items       []map[string]json.RawMessage `json:"items"`
		Total       int                          `json:"total"`
		UnreadCount int                          `json:"unread_count"`
	}
	authdb.Require(t, "projection trộn HTTP", json.Unmarshal(body, &page))
	if len(page.Items) != 6 || page.Total != 6 || page.UnreadCount != 6 {
		t.Fatal("HTTP không đọc đủ sáu loại hoặc nhân dòng do role")
	}
	types := map[string]bool{}
	for _, item := range page.Items {
		if len(item) != 7 || string(item["read_at"]) != "null" {
			t.Fatal("DTO allowlist hoặc read_at sai")
		}
		for _, key := range []string{"notification_id", "notification_type", "resource_type", "resource_id", "message", "created_at", "read_at"} {
			if item[key] == nil {
				t.Fatal("DTO thiếu trường contract")
			}
		}
		var notificationType string
		authdb.Require(t, "loại notification", json.Unmarshal(item["notification_type"], &notificationType))
		types[notificationType] = true
	}
	if len(types) != 6 {
		t.Fatal("projection thiếu loại notification")
	}
	for _, key := range []string{"note", "decision_reason", "phone", "address", "recipient_id", "before_values", "proposed_values", "password"} {
		if strings.Contains(string(body), `"`+key+`"`) {
			t.Fatal("DTO chứa trường ngoài mục đích thông báo")
		}
	}
	var readCount int
	authdb.Require(t, "mở list không tự read", f.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM notifications WHERE recipient_id=$1 AND read_at IS NOT NULL`, owner.UserID).Scan(&readCount))
	if readCount != 0 {
		t.Fatal("GET đã tự đánh dấu đọc")
	}
	_, body = catalogCall(t, a, "GET", "/api/notifications?page=2&page_size=2", cookie, "", nil, 200)
	authdb.Require(t, "phân trang trộn", json.Unmarshal(body, &page))
	if len(page.Items) != 2 || page.Total != 6 || page.UnreadCount != 6 {
		t.Fatal("phân trang làm thay total/unread")
	}
	path := "/api/notifications/" + strconv.Itoa(lastID) + "/read"
	other := a.login(t, "skills-other@example.invalid", authdb.FixturePassword, nil)
	catalogCall(t, a, "POST", path, other, a.csrf(t, other), map[string]any{}, 404)
	_, body = catalogCall(t, a, "GET", "/api/notifications", other, "", nil, 200)
	authdb.Require(t, "thông báo ngoài owner", json.Unmarshal(body, &page))
	if len(page.Items) != 0 || page.Total != 0 || page.UnreadCount != 0 {
		t.Fatal("Admin thấy thông báo người khác")
	}
	_, err := f.DB.Conn.Exec(t.Context(), `DELETE FROM user_roles WHERE user_id=$1`, owner.UserID)
	authdb.Require(t, "gỡ quyền tài nguyên", err)
	catalogCall(t, a, "POST", path, cookie, token, map[string]any{}, 204)
	var firstRead, secondRead time.Time
	authdb.Require(t, "read kỹ năng đã lưu", f.DB.Conn.QueryRow(t.Context(), `SELECT read_at FROM notifications WHERE notification_id=$1`, lastID).Scan(&firstRead))
	catalogCall(t, a, "POST", path, cookie, token, map[string]any{}, 204)
	authdb.Require(t, "read lặp kỹ năng", f.DB.Conn.QueryRow(t.Context(), `SELECT read_at FROM notifications WHERE notification_id=$1`, lastID).Scan(&secondRead))
	if !firstRead.Equal(secondRead) {
		t.Fatal("read lặp thay thời điểm lần đầu")
	}
	_, body = catalogCall(t, a, "GET", "/api/notifications?filter=unread", cookie, "", nil, 200)
	authdb.Require(t, "unread sau read", json.Unmarshal(body, &page))
	if len(page.Items) != 5 || page.Total != 5 || page.UnreadCount != 5 {
		t.Fatal("unread/count trộn sai")
	}
	_, body = catalogCall(t, a, "GET", "/api/notifications?page=2147483647", cookie, "", nil, 200)
	authdb.Require(t, "page rỗng trộn", json.Unmarshal(body, &page))
	if len(page.Items) != 0 || page.Total != 6 || page.UnreadCount != 5 {
		t.Fatal("page rỗng làm mất count")
	}
	for _, tc := range []struct{ notificationType, resourceType string }{
		{"SKILL_CHANGE_REJECTED", "PROFILE_CHANGE_REQUEST"},
		{"PROFILE_CHANGE_REJECTED", "SKILL_CHANGE_REQUEST"},
		{"SKILL_CHANGE_UNKNOWN", "SKILL_CHANGE_REQUEST"},
	} {
		_, err = f.DB.Conn.Exec(t.Context(), `UPDATE notifications SET notification_type=$1,resource_type=$2,message='Nội dung fixture không được trả trong lỗi.' WHERE notification_id=$3`, tc.notificationType, tc.resourceType, lastID)
		authdb.Require(t, "fault cặp notification", err)
		_, body = catalogCall(t, a, "GET", "/api/notifications", cookie, "", nil, 500)
		assertErrorCode(t, body, "INTERNAL_ERROR")
		if strings.Contains(string(body), "Nội dung fixture") || strings.Contains(string(body), tc.notificationType) || strings.Contains(string(body), "notification_type") {
			t.Fatal("lỗi persisted làm lộ nội dung hoặc chi tiết provider")
		}
	}
	_, err = f.DB.Conn.Exec(t.Context(), `UPDATE notifications SET notification_type='SKILL_CHANGE_REJECTED',resource_type='SKILL_CHANGE_REQUEST',message='Có cập nhật đề nghị.' WHERE notification_id=$1`, lastID)
	authdb.Require(t, "khôi phục cặp notification", err)
	_, body = catalogCall(t, a, "GET", "/api/notifications", cookie, "", nil, 200)
	authdb.Require(t, "projection sau khôi phục", json.Unmarshal(body, &page))
	if page.Total != 6 || page.UnreadCount != 5 {
		t.Fatal("lỗi list làm mất dữ liệu hoặc đổi read")
	}
}
