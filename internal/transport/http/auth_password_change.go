package http

import (
	"context"
	"errors"
	"io"
	"net/http"

	"ai-staffing/internal/service/auth"
)

type passwordChangeUseCase interface {
	ChangeInitialPassword(context.Context, auth.SessionSecret, string, string) error
}

type passwordChangeRoute struct {
	service passwordChangeUseCase
	limiter *authRateLimiter
}

func (change passwordChangeRoute) handler(protection authProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromRequest(r)
		if !ok {
			writeAuthSessionError(w, r, auth.ErrUnauthenticated)
			return
		}
		if !protection.allowsMutation(r, session.secret) {
			writeAuthError(w, r, http.StatusForbidden, "FORBIDDEN", "Yêu cầu đổi mật khẩu không hợp lệ.")
			return
		}
		ip, ok := authPeerIP(r)
		if !ok {
			writeAuthSessionError(w, r, auth.ErrPersistence)
			return
		}
		if wait := change.limiter.allowPasswordChangeIP(ip); wait > 0 {
			writeAuthRateLimited(w, r, wait)
			return
		}
		if !hasAuthJSONContentType(r) {
			writeAuthError(w, r, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Yêu cầu đổi mật khẩu phải dùng JSON.")
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, authBodyLimit))
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeAuthError(w, r, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Dữ liệu đổi mật khẩu vượt quá 8 KiB.")
			return
		}
		fields, valid := decodeAuthStringPair(body, "current_password", "new_password")
		if err != nil || !valid || fields["current_password"] == "" {
			writeAuthError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Dữ liệu đổi mật khẩu không hợp lệ.")
			return
		}
		finish, wait := change.limiter.beginPasswordChangeAttempt(session.principal)
		if wait > 0 {
			writeAuthRateLimited(w, r, wait)
			return
		}
		defer finish(false)

		err = change.service.ChangeInitialPassword(r.Context(), session.secret, fields["current_password"], fields["new_password"])
		switch {
		case err == nil:
			protection.clearSessionCookie(w)
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, auth.ErrInvalidCurrentPassword):
			finish(true)
			writeAuthError(w, r, http.StatusBadRequest, "INVALID_CURRENT_PASSWORD", "Mật khẩu hiện tại không đúng.")
		case errors.Is(err, auth.ErrNewPasswordPolicy):
			writeAuthError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Mật khẩu mới phải khác mật khẩu hiện tại, có ít nhất 15 ký tự và không quá 72 byte UTF-8.")
		case errors.Is(err, auth.ErrPasswordChangeNotRequired):
			writeAuthError(w, r, http.StatusForbidden, "PASSWORD_CHANGE_NOT_REQUIRED", "Tài khoản không cần đổi mật khẩu lần đầu.")
		default:
			writeAuthSessionError(w, r, err)
		}
	}
}
