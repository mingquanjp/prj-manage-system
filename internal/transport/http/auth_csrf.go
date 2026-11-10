package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"ai-staffing/internal/service/auth"

	"github.com/go-chi/chi/v5/middleware"
)

type csrfAuthenticator interface {
	Authenticate(context.Context, auth.SessionSecret) (auth.Principal, error)
}

func csrfHandler(authenticator csrfAuthenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret, ok := readSessionSecret(r)
		if !ok {
			writeAuthSessionError(w, r, auth.ErrUnauthenticated)
			return
		}
		if _, err := authenticator.Authenticate(r.Context(), secret); err != nil {
			writeAuthSessionError(w, r, err)
			return
		}
		// Phiên hạn chế hợp lệ vẫn cần token để đổi mật khẩu hoặc đăng xuất.
		writeAuthJSON(w, http.StatusOK, struct {
			Token string `json:"csrf_token"`
		}{Token: authEncoding.EncodeToString(csrfDigest(secret))})
	}
}

// serverAuthRequestID dùng mã do server cấp để phản hồi và log không lặp header từ client.
func serverAuthRequestID(next http.Handler) http.Handler {
	withRequestID := middleware.RequestID(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Router đã cấp mã nội bộ khi client không gửi mã.
		if r.Header.Get(middleware.RequestIDHeader) == "" {
			next.ServeHTTP(w, r)
			return
		}
		request := r.Clone(r.Context())
		request.Header.Del(middleware.RequestIDHeader)
		withRequestID.ServeHTTP(w, request)
	})
}

func recoverAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if cause := recover(); cause != nil {
				if cause == http.ErrAbortHandler {
					panic(cause)
				}
				// Nội dung panic có thể chứa dữ liệu xác thực.
				writeAuthSessionError(w, r, auth.ErrPersistence)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeAuthSessionError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Đã xảy ra lỗi khi xác thực."
	switch {
	case errors.Is(err, auth.ErrUnauthenticated):
		status, code, message = http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, auth.ErrUnavailable):
		status, code, message = http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Dịch vụ xác thực tạm thời không khả dụng."
	}
	requestID := middleware.GetReqID(r.Context())
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "xác thực phiên thất bại", "code", code, "request_id", requestID)
	}
	writeAuthJSON(w, status, struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	}{Code: code, Message: message, RequestID: requestID})
}

func writeAuthJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
