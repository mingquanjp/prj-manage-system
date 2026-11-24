package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
	n "ai-staffing/internal/service/notification"
)

type notificationStub struct {
	err   error
	calls int
	actor int
}

func (s *notificationStub) List(_ context.Context, q n.ListQuery) (n.Page, error) {
	s.calls++
	s.actor = q.ActorID
	return n.Page{Items: []n.Item{}, Page: q.Page, PageSize: q.PageSize}, s.err
}
func (s *notificationStub) MarkRead(_ context.Context, q n.ReadCommand) error {
	s.calls++
	s.actor = q.ActorID
	return s.err
}
func notificationRouter(t *testing.T, s *notificationStub, env string) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	session := catalogHR()
	h, e := NewAuthRouter(fakeHealth{}, env, loginTestOrigin, AuthServices{Sessions: session, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: session, Clock: clock, Notifications: s})
	if e != nil {
		t.Fatal(e)
	}
	return h
}
func TestNotificationHTTPErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{n.ErrUnauthenticated, 401, "UNAUTHENTICATED"}, {n.ErrPasswordChangeRequired, 403, "PASSWORD_CHANGE_REQUIRED"}, {n.ErrValidation, 400, "VALIDATION_ERROR"}, {n.ErrNotFound, 404, "NOT_FOUND"}, {n.ErrPersistence, 500, "INTERNAL_ERROR"}, {n.ErrUnavailable, 503, "SERVICE_UNAVAILABLE"}, {context.DeadlineExceeded, 503, "SERVICE_UNAVAILABLE"}, {errors.Join(n.ErrCommitUnknown, n.ErrUnavailable), 503, "SERVICE_UNAVAILABLE"}, {n.ErrCommitUnknown, 500, "INTERNAL_ERROR"},
	} {
		t.Run(tc.code+tc.err.Error(), func(t *testing.T) {
			w := httptest.NewRecorder()
			writeNotificationError(w, httptest.NewRequest("POST", "/api/notifications/1/read", nil), tc.err)
			if w.Code != tc.status || !strings.Contains(w.Body.String(), `"code":"`+tc.code+`"`) {
				t.Fatal("phân loại lỗi sai")
			}
		})
	}
}
func TestNotificationSwaggerOperationInventory(t *testing.T) {
	for _, env := range []string{"dev", "production"} {
		h := notificationRouter(t, &notificationStub{}, env)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/docs/notifications.openapi.json", nil))
		if env != "dev" {
			if w.Code != 404 {
				t.Fatal("production mở docs")
			}
			continue
		}
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("dev spec không chạy")
		}
		var spec struct {
			Paths map[string]map[string]json.RawMessage `json:"paths"`
		}
		if json.Unmarshal(w.Body.Bytes(), &spec) != nil || len(spec.Paths) != 2 || len(spec.Paths["/api/notifications"]) != 1 || len(spec.Paths["/api/notifications/{id}/read"]) != 1 || spec.Paths["/api/notifications"]["get"] == nil || spec.Paths["/api/notifications/{id}/read"]["post"] == nil {
			t.Fatal("Swagger quảng cáo route chưa chạy")
		}
	}
	w := httptest.NewRecorder()
	swaggerWithNotifications(AuthServices{}, false)(w, httptest.NewRequest("GET", "/docs/notifications.openapi.json", nil))
	if w.Code != 404 {
		t.Fatal("spec không tắt theo dependency")
	}
}
