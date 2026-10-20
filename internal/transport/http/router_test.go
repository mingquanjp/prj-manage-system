package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type fakeHealth struct{ err error }

func (f fakeHealth) Ping(context.Context) error { return f.err }

func TestHealthResponseDoesNotExposeDriverError(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	for _, failed := range []bool{false, true} {
		var err error
		if failed {
			err = errors.New("driver-secret-example database-host-example")
		}
		router := NewRouter(fakeHealth{err: err})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/healthz", nil))
		wantCode, wantStatus, wantDB := http.StatusOK, "ok", "ok"
		if failed {
			wantCode, wantStatus, wantDB = http.StatusServiceUnavailable, "degraded", "unavailable"
		}
		var body struct {
			Status   string `json:"status"`
			Database string `json:"database"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if response.Code != wantCode || body.Status != wantStatus || body.Database != wantDB {
			t.Fatalf("health không đúng: code=%d body=%s", response.Code, response.Body.String())
		}
		if strings.Contains(response.Body.String()+logs.String(), "driver-secret-example") {
			t.Fatal("lỗi driver thô xuất hiện trong response hoặc log")
		}
	}
}

func TestRouterDoesNotTrustForwardedIP(t *testing.T) {
	router := NewRouter(fakeHealth{}).(chi.Router)
	router.Get("/test-remote", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.RemoteAddr))
	})
	request := httptest.NewRequest(http.MethodGet, "/test-remote", nil)
	request.RemoteAddr = "192.0.2.10:12345"
	request.Header.Set("X-Forwarded-For", "203.0.113.50")
	request.Header.Set("X-Real-IP", "203.0.113.51")
	request.Header.Set("True-Client-IP", "203.0.113.52")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Body.String() != request.RemoteAddr {
		t.Fatalf("router đã tin IP do client gửi: %s", response.Body.String())
	}
}
