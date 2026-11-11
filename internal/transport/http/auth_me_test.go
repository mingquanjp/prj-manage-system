package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
)

func TestMeThroughRouter(t *testing.T) {
	now := time.Unix(1000, 0)
	for _, tt := range []struct {
		name   string
		cookie bool
		update func(*csrfTestRepo)
		status int
		roles  []string
	}{
		{name: "phiên thường", cookie: true, status: 200, roles: []string{"ADMIN"}},
		{name: "phiên thường không có quyền đặc biệt", cookie: true, status: 200, roles: []string{}, update: func(r *csrfTestRepo) { r.snapshot.Principal.SystemRoles = nil }},
		{name: "phiên buộc đổi", cookie: true, status: 200, roles: []string{}, update: func(r *csrfTestRepo) { r.snapshot.Principal.MustChangePassword = true }},
		{name: "thiếu cookie", status: 401},
		{name: "hết hạn", cookie: true, status: 401, update: func(r *csrfTestRepo) { r.snapshot.ExpiresAt = now }},
		{name: "đã thu hồi", cookie: true, status: 401, update: func(r *csrfTestRepo) { r.snapshot.RevokedAt = &now }},
		{name: "tài khoản khóa", cookie: true, status: 401, update: func(r *csrfTestRepo) { r.snapshot.UserActive = false }},
		{name: "nhân viên ngừng hoạt động", cookie: true, status: 401, update: func(r *csrfTestRepo) { r.snapshot.EmployeeActive = false }},
		{name: "DB tạm lỗi", cookie: true, status: 503, update: func(r *csrfTestRepo) { r.err = auth.ErrUnavailable }},
		{name: "lỗi lưu trữ", cookie: true, status: 500, update: func(r *csrfTestRepo) { r.err = auth.ErrPersistence }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &csrfTestRepo{
				found: true,
				snapshot: auth.SessionSnapshot{
					Principal: auth.Principal{
						UserID:       12,
						EmployeeID:   37,
						EmployeeCode: "E037",
						Fullname:     "Nguyễn Văn An",
						SystemRoles:  []string{"ADMIN"},
					},
					ExpiresAt:      now.Add(time.Hour),
					UserActive:     true,
					EmployeeActive: true,
				},
			}
			if tt.update != nil {
				tt.update(repo)
			}
			service := auth.NewService(repo, csrfTestClock{now})
			handler := newRouter(fakeHealth{}, service, authProtection{}, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if tt.cookie {
				secret := sequenceSecret()
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: authEncoding.EncodeToString(secret[:])})
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)

			if response.Code != tt.status || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Set-Cookie") != "" {
				t.Fatalf("status = %d, muốn %d, headers = %v", response.Code, tt.status, response.Header())
			}
			wantReads := 0
			if tt.cookie {
				wantReads = 1
			}
			if repo.reads != wantReads {
				t.Errorf("số lần đọc phiên = %d, muốn %d", repo.reads, wantReads)
			}
			if tt.status != http.StatusOK {
				var body authErrorResponse
				wantCode := map[int]string{401: "UNAUTHENTICATED", 503: "SERVICE_UNAVAILABLE", 500: "INTERNAL_ERROR"}[tt.status]
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Code != wantCode {
					t.Fatalf("code = %q, muốn %q, lỗi = %v", body.Code, wantCode, err)
				}
				return
			}
			var body authIdentityResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			want := authIdentityResponse{12, 37, "E037", "Nguyễn Văn An", tt.roles, repo.snapshot.Principal.MustChangePassword}
			if !reflect.DeepEqual(body, want) {
				t.Errorf("response = %+v, muốn %+v", body, want)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(response.Body.Bytes(), &fields); err != nil || len(fields) != 6 {
				t.Fatalf("DTO phải có đúng sáu trường, lỗi = %v", err)
			}
		})
	}
}
