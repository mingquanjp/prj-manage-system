package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"unicode/utf8"

	"ai-staffing/internal/service/auth"
)

const authBodyLimit = 8192

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
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" || len(r.Header.Values("Content-Type")) != 1 {
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
	if !utf8.Valid(body) {
		return "", "", false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return "", "", false
	}
	fields := make(map[string]string, 2)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil || (token != "email" && token != "password") {
			return "", "", false
		}
		key := token.(string)
		if _, duplicate := fields[key]; duplicate {
			return "", "", false
		}
		var value string
		if err := decoder.Decode(&value); err != nil {
			return "", "", false
		}
		fields[key] = value
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return "", "", false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return "", "", false
	}

	email = auth.NormalizeLoginEmail(fields["email"])
	password = fields["password"]
	address, err := mail.ParseAddress(email)
	valid = err == nil && address.Address == email && utf8.RuneCountInString(email) <= 100 && password != ""
	return email, password, valid
}
