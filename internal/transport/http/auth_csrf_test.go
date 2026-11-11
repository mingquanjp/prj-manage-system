package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"

	"github.com/go-chi/chi/v5/middleware"
)

// Vector độc lập: key là 32 byte từ 0 đến 31, message là "csrf-v1".
const csrfVector = "nrFhTNM11DML5EQbTxAN7yOll570EIsnn_zlhuTux5Q"

func sequenceSecret() auth.SessionSecret {
	var secret auth.SessionSecret
	for i := range secret {
		secret[i] = byte(i)
	}
	return secret
}

type csrfTestClock struct{ now time.Time }

func (c csrfTestClock) Now() time.Time { return c.now }

type csrfTestRepo struct {
	snapshot  auth.SessionSnapshot
	err       error
	found     bool
	reads     int
	revokes   int
	hash      auth.SessionHash
	requestID string
	panicWith string
}

func (f *csrfTestRepo) ReadSession(ctx context.Context, hash auth.SessionHash) (auth.SessionSnapshot, bool, error) {
	f.reads++
	f.hash = hash
	f.requestID = middleware.GetReqID(ctx)
	if f.panicWith != "" {
		panic(f.panicWith)
	}
	return f.snapshot, f.found, f.err
}

func (f *csrfTestRepo) RevokeCurrentSession(context.Context, auth.SessionHash) error {
	f.revokes++
	return nil
}

func TestCSRFBootstrapThroughRouter(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	now := time.Date(2026, 11, 10, 5, 0, 0, 0, time.UTC)
	secret := sequenceSecret()
	encoded := base64.RawURLEncoding.EncodeToString(secret[:])
	session := "staffing_session=" + encoded
	for _, tt := range []struct {
		name    string
		cookies []string
		change  func(*csrfTestRepo)
		status  int
		code    string
		reads   int
	}{
		{"phiên thường cùng cookie khác", []string{"preferences=compact; " + session}, nil, 200, "", 1},
		{"phiên hạn chế", []string{session}, func(r *csrfTestRepo) { r.snapshot.Principal.MustChangePassword = true }, 200, "", 1},
		{"cookie thiếu", nil, nil, 401, "UNAUTHENTICATED", 0},
		{"cookie sai độ dài", []string{"staffing_session=short"}, nil, 401, "UNAUTHENTICATED", 0},
		{"cookie sai base64", []string{"staffing_session=" + strings.Repeat("!", 43)}, nil, 401, "UNAUTHENTICATED", 0},
		{"cookie padding", []string{session + "="}, nil, 401, "UNAUTHENTICATED", 0},
		{"cookie trùng", []string{session + "; " + session}, nil, 401, "UNAUTHENTICATED", 0},
		{"cookie trùng qua hai header", []string{session, session}, nil, 401, "UNAUTHENTICATED", 0},
		{"cookie lỗi cùng cookie hợp lệ", []string{session + "; staffing_session=\"bad"}, nil, 401, "UNAUTHENTICATED", 0},
		{"không tìm thấy", []string{session}, func(r *csrfTestRepo) { r.found = false }, 401, "UNAUTHENTICATED", 1},
		{"đúng biên hết hạn", []string{session}, func(r *csrfTestRepo) { r.snapshot.ExpiresAt = now }, 401, "UNAUTHENTICATED", 1},
		{"đã thu hồi", []string{session}, func(r *csrfTestRepo) { r.snapshot.RevokedAt = &now }, 401, "UNAUTHENTICATED", 1},
		{"tài khoản khóa", []string{session}, func(r *csrfTestRepo) { r.snapshot.UserActive = false }, 401, "UNAUTHENTICATED", 1},
		{"nhân viên ngừng hoạt động", []string{session}, func(r *csrfTestRepo) { r.snapshot.EmployeeActive = false }, 401, "UNAUTHENTICATED", 1},
		{"DB tạm lỗi dù snapshot hợp lệ", []string{session}, func(r *csrfTestRepo) { r.err = fmt.Errorf("DRIVER_MARKER: %w", auth.ErrUnavailable) }, 503, "SERVICE_UNAVAILABLE", 1},
		{"dữ liệu mâu thuẫn", []string{session}, func(r *csrfTestRepo) { r.err = auth.ErrIdentityConflict }, 500, "INTERNAL_ERROR", 1},
		{"lỗi lạ", []string{session}, func(r *csrfTestRepo) { r.err = errors.New("DRIVER_MARKER") }, 500, "INTERNAL_ERROR", 1},
		{"panic", []string{session}, func(r *csrfTestRepo) { r.panicWith = "DRIVER_MARKER" }, 500, "INTERNAL_ERROR", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			logs.Reset()
			repo := &csrfTestRepo{
				found: true,
				snapshot: auth.SessionSnapshot{
					Principal: auth.Principal{UserID: 13, EmployeeID: 23, SystemRoles: []string{"ADMIN"}},
					ExpiresAt: now.Add(time.Hour), UserActive: true, EmployeeActive: true,
				},
			}
			if tt.change != nil {
				tt.change(repo)
			}
			router := newRouter(fakeHealth{}, auth.NewService(repo, csrfTestClock{now}), authProtection{})
			r := httptest.NewRequest(http.MethodGet, "/api/auth/csrf", nil)
			r.Header.Set("Origin", "https://other.example")
			if tt.status != http.StatusServiceUnavailable {
				r.Header.Set(middleware.RequestIDHeader, "REQUEST_MARKER")
			}
			r.Header.Set("X-CSRF-Token", "TOKEN_MARKER")
			for _, cookie := range tt.cookies {
				r.Header.Add("Cookie", cookie)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			if w.Code != tt.status || w.Header().Get("Cache-Control") != "no-store" ||
				w.Header().Get("Content-Type") != "application/json; charset=utf-8" ||
				w.Header().Get("Set-Cookie") != "" || w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatalf("status/header ngoài contract, status=%d", w.Code)
			}
			if repo.reads != tt.reads || repo.revokes != 0 ||
				(repo.reads > 0 && repo.hash != sha256.Sum256(secret[:])) {
				t.Fatal("route phải đọc đúng hash và không thu hồi phiên")
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tt.status == 200 {
				if len(body) != 1 || body["csrf_token"] != csrfVector {
					t.Fatal("phản hồi thành công không khớp DTO hoặc vector HMAC")
				}
			} else if len(body) != 3 || body["code"] != tt.code || body["message"] == "" || body["request_id"] == "" ||
				(repo.reads > 0 && body["request_id"] != repo.requestID) {
				t.Fatal("DTO lỗi, mã lỗi hoặc request_id không đúng")
			}
			for _, marker := range []string{encoded, "REQUEST_MARKER", "TOKEN_MARKER", "DRIVER_MARKER"} {
				if strings.Contains(w.Body.String()+logs.String(), marker) {
					t.Fatal("phát hiện dữ liệu nhạy cảm trong phản hồi hoặc log")
				}
			}
			if strings.Contains(logs.String(), csrfVector) ||
				(tt.status >= 500 && !strings.Contains(logs.String(), body["request_id"])) {
				t.Fatal("log thiếu request_id hoặc có CSRF token")
			}
		})
	}
}

func TestCSRFRouteMethodAndDependency(t *testing.T) {
	repo := &csrfTestRepo{}
	configured := newRouter(fakeHealth{}, auth.NewService(repo, csrfTestClock{}), authProtection{})
	for _, tt := range []struct {
		method string
		router http.Handler
		status int
	}{
		{http.MethodPost, configured, http.StatusMethodNotAllowed},
		{http.MethodGet, NewRouter(fakeHealth{}), http.StatusNotFound},
	} {
		w := httptest.NewRecorder()
		tt.router.ServeHTTP(w, httptest.NewRequest(tt.method, "/api/auth/csrf", nil))
		if w.Code != tt.status || repo.reads != 0 {
			t.Fatalf("%s /api/auth/csrf: nhận %d, cần %d và không đọc phiên", tt.method, w.Code, tt.status)
		}
	}
}
