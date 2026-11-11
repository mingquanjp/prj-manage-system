package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"unicode/utf8"

	"ai-staffing/internal/service/auth"
)

type loginUseCase interface {
	Login(context.Context, string, string, *auth.SessionSecret) (auth.LoginResult, error)
}

type loginRoute struct {
	service loginUseCase
	limiter *authRateLimiter
	clock   auth.Clock
}

func (login loginRoute) handler(protection authProtection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !protection.allowsOrigin(r) {
			writeAuthError(w, r, http.StatusForbidden, "FORBIDDEN", "Nguồn yêu cầu đăng nhập không hợp lệ.")
			return
		}
		ip, ok := authPeerIP(r)
		if !ok {
			writeAuthSessionError(w, r, auth.ErrPersistence)
			return
		}
		if wait := login.limiter.allowLoginIP(ip); wait > 0 {
			writeAuthRateLimited(w, r, wait)
			return
		}
		if !hasAuthJSONContentType(r) {
			writeAuthError(w, r, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Yêu cầu đăng nhập phải dùng JSON.")
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, authBodyLimit))
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeAuthError(w, r, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Dữ liệu đăng nhập vượt quá 8 KiB.")
			return
		}
		email, password, valid := decodeLoginBody(body)
		if err != nil || !valid {
			writeAuthError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Email công ty hoặc dữ liệu đăng nhập không hợp lệ.")
			return
		}
		finish, wait := login.limiter.beginLoginAttempt(email)
		if wait > 0 {
			writeAuthRateLimited(w, r, wait)
			return
		}
		defer finish(false)

		var previous *auth.SessionSecret
		if secret, ok := readSessionSecret(r); ok {
			previous = &secret
		}
		result, err := login.service.Login(r.Context(), email, password, previous)
		if errors.Is(err, auth.ErrLoginVerificationFailed) {
			finish(true)
		}
		if errors.Is(err, auth.ErrLoginVerificationFailed) || errors.Is(err, auth.ErrInvalidCredentials) {
			writeAuthError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email hoặc mật khẩu không hợp lệ.")
			return
		}
		if err != nil {
			writeAuthSessionError(w, r, err)
			return
		}
		if !protection.setSessionCookie(w, result.Secret, result.ExpiresAt, login.clock.Now()) {
			writeAuthSessionError(w, r, auth.ErrPersistence)
			return
		}
		writeAuthJSON(w, http.StatusOK, newAuthIdentityResponse(result.Principal))
	}
}

func decodeLoginBody(body []byte) (email, password string, valid bool) {
	fields, ok := decodeAuthStringPair(body, "email", "password")
	if !ok {
		return "", "", false
	}

	email = auth.NormalizeLoginEmail(fields["email"])
	password = fields["password"]
	address, err := mail.ParseAddress(email)
	valid = err == nil && address.Address == email && utf8.RuneCountInString(email) <= 100 && password != ""
	return email, password, valid
}
