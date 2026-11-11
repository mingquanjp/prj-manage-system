package http

import (
	"context"
	"io"
	"net/http"

	"ai-staffing/internal/service/auth"
)

type sessionLogout interface {
	Logout(context.Context, auth.SessionSecret) error
}

func logoutHandler(service sessionLogout, protection authProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromRequest(r)
		if !ok {
			writeAuthSessionError(w, r, auth.ErrUnauthenticated)
			return
		}
		if !protection.allowsMutation(r, session.secret) {
			writeAuthError(w, r, http.StatusForbidden, "FORBIDDEN", "Yêu cầu đăng xuất không hợp lệ.")
			return
		}
		// Đọc một byte để phát hiện body có dữ liệu kể cả khi không có Content-Length.
		if r.Body != nil {
			var firstByte [1]byte
			if _, err := io.ReadFull(r.Body, firstByte[:]); err != io.EOF {
				writeAuthError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Yêu cầu đăng xuất phải có body rỗng.")
				return
			}
		}
		if err := service.Logout(r.Context(), session.secret); err != nil {
			writeAuthSessionError(w, r, err)
			return
		}
		protection.clearSessionCookie(w)
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusNoContent)
	}
}
