package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/service/auth"
)

const (
	sessionCookieName = "staffing_session"
	sessionCookiePath = "/api"
	csrfHeaderName    = "X-CSRF-Token"
	csrfMessage       = "csrf-v1"
)

var authEncoding = base64.RawURLEncoding.Strict()

type authProtection struct {
	origin       config.Origin
	secureCookie bool
}

// newAuthProtection dùng APP_ENV và APP_ORIGIN đã qua validation của server.
func newAuthProtection(appEnv, appOrigin string) (authProtection, error) {
	origin, err := config.ParseOrigin(appOrigin)
	if err != nil {
		return authProtection{}, err
	}
	localHTTP := appEnv == "dev" && origin.Scheme == "http" && origin.IsLoopback()
	return authProtection{origin: origin, secureCookie: !localHTTP}, nil
}

func (p authProtection) allowsOrigin(r *http.Request) bool {
	values := r.Header.Values("Origin")
	if len(values) != 1 {
		return false
	}
	origin, err := config.ParseOrigin(values[0])
	return err == nil && origin == p.origin
}

func (p authProtection) allowsMutation(r *http.Request, secret auth.SessionSecret) bool {
	return p.allowsOrigin(r) && validCSRFToken(r, secret)
}

func readSessionSecret(r *http.Request) (auth.SessionSecret, bool) {
	var secret auth.SessionSecret
	found := false
	for _, line := range r.Header.Values("Cookie") {
		cookies, err := http.ParseCookie(line)
		if err != nil {
			return auth.SessionSecret{}, false
		}
		for _, cookie := range cookies {
			if cookie.Name != sessionCookieName {
				continue
			}
			if found || !decodeAuthValue(secret[:], cookie.Value) {
				return auth.SessionSecret{}, false
			}
			found = true
		}
	}
	return secret, found
}

// setSessionCookie dùng hạn đã commit. Expires là mốc tuyệt đối của phiên.
func (p authProtection) setSessionCookie(w http.ResponseWriter, secret auth.SessionSecret, expiresAt, now time.Time) bool {
	expires := expiresAt.UTC().Truncate(time.Second)
	if !expires.After(now) {
		return false
	}
	cookie := p.sessionCookie()
	cookie.Value = authEncoding.EncodeToString(secret[:])
	cookie.Expires = expires
	http.SetCookie(w, &cookie)
	return true
}

func (p authProtection) clearSessionCookie(w http.ResponseWriter) {
	cookie := p.sessionCookie()
	cookie.Expires = time.Unix(1, 0).UTC()
	cookie.MaxAge = -1 // net/http gửi Max-Age=0.
	http.SetCookie(w, &cookie)
}

func (p authProtection) sessionCookie() http.Cookie {
	return http.Cookie{
		Name:     sessionCookieName,
		Path:     sessionCookiePath,
		HttpOnly: true,
		Secure:   p.secureCookie,
		SameSite: http.SameSiteLaxMode,
	}
}

func csrfDigest(secret auth.SessionSecret) []byte {
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(csrfMessage))
	return mac.Sum(nil)
}

func validCSRFToken(r *http.Request, secret auth.SessionSecret) bool {
	values := r.Header.Values(csrfHeaderName)
	var supplied [sha256.Size]byte
	return len(values) == 1 && decodeAuthValue(supplied[:], values[0]) &&
		hmac.Equal(supplied[:], csrfDigest(secret))
}

func decodeAuthValue(dst []byte, value string) bool {
	// Kiểm chiều dài trước để loại biến thể có padding hoặc ký tự xuống dòng.
	if len(value) != authEncoding.EncodedLen(len(dst)) {
		return false
	}
	n, err := authEncoding.Decode(dst, []byte(value))
	return err == nil && n == len(dst)
}
