package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
)

func TestSessionCookieHeaders(t *testing.T) {
	secret := auth.SessionSecret{1, 2, 3}
	now := time.Date(2026, 11, 10, 5, 0, 0, 500_000_000, time.UTC)
	expires := now.Add(8 * time.Hour)
	for _, tt := range []struct {
		name, env, origin string
		secure            bool
	}{
		{"dev localhost", "dev", "http://localhost:5173", false},
		{"dev IPv4 loopback", "dev", "http://127.0.0.1:5173", false},
		{"dev IPv6 loopback", "dev", "http://[::1]:5173", false},
		{"dev HTTPS", "dev", "https://localhost:5173", true},
		{"prod HTTPS", "prod", "https://staffing.example", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := newAuthProtection(tt.env, tt.origin)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			if !policy.setSessionCookie(w, secret, expires, now) {
				t.Fatal("không đặt cookie cho phiên còn hạn")
			}
			set := checkedSessionCookie(t, w, tt.secure)
			if set.Value == "" || !set.Expires.Equal(expires.Truncate(time.Second)) || set.MaxAge != 0 {
				t.Fatal("cookie phải hết hạn theo mốc tuyệt đối đã commit")
			}
			r := httptest.NewRequest(http.MethodGet, "/api/auth/csrf", nil)
			r.AddCookie(set)
			if read, ok := readSessionSecret(r); !ok || read != secret {
				t.Fatal("cookie không đọc lại đúng secret")
			}

			w = httptest.NewRecorder()
			policy.clearSessionCookie(w)
			cleared := checkedSessionCookie(t, w, tt.secure)
			if cleared.Value != "" || !cleared.Expires.Before(now) ||
				!strings.Contains(w.Header().Get("Set-Cookie"), "; Max-Age=0") {
				t.Fatal("cookie xóa thiếu giá trị rỗng, Max-Age=0 hoặc hạn quá khứ")
			}

			w = httptest.NewRecorder()
			if policy.setSessionCookie(w, secret, now, now) || w.Header().Get("Set-Cookie") != "" {
				t.Fatal("đã đặt cookie tại biên hết hạn")
			}
		})
	}
}

func checkedSessionCookie(t *testing.T, w *httptest.ResponseRecorder, secure bool) *http.Cookie {
	t.Helper()
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cần đúng một Set-Cookie, nhận %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "staffing_session" || c.Path != "/api" || c.Domain != "" ||
		strings.Contains(w.Header().Get("Set-Cookie"), "; Domain=") ||
		!c.HttpOnly || c.Secure != secure || c.SameSite != http.SameSiteLaxMode {
		t.Fatal("thuộc tính bảo mật cookie không đúng contract")
	}
	return c
}

func TestAuthOrigin(t *testing.T) {
	for _, tt := range []struct {
		name, appOrigin string
		values          []string
		want            bool
	}{
		{"đúng origin proxy", "http://localhost:5173", []string{"http://localhost:5173"}, true},
		{"port mặc định", "https://staffing.example:443", []string{"https://staffing.example"}, true},
		{"punycode đúng host", "https://xn--i-9bb.example", []string{"https://xn--i-9bb.example"}, true},
		{"punycode khác host ASCII", "https://xn--i-9bb.example", []string{"https://i.example"}, false},
		{"Origin Unicode không được đổi thành host khác", "https://i.example", []string{"https://İ.example"}, false},
		{"host hoa thường", "https://STAFFING.example", []string{"https://staffing.example"}, true},
		{"IPv6 cùng địa chỉ", "http://[0:0:0:0:0:0:0:1]:5173", []string{"http://[::1]:5173"}, true},
		{"khác scheme", "http://localhost:5173", []string{"https://localhost:5173"}, false},
		{"khác port", "http://localhost:5173", []string{"http://localhost:8080"}, false},
		{"khác host", "http://localhost:5173", []string{"http://127.0.0.1:5173"}, false},
		{"IPv4 trong ngoặc", "http://127.0.0.1:5173", []string{"http://[127.0.0.1]:5173"}, false},
		{"thiếu", "http://localhost:5173", nil, false},
		{"null", "http://localhost:5173", []string{"null"}, false},
		{"hai header giống nhau", "http://localhost:5173", []string{"http://localhost:5173", "http://localhost:5173"}, false},
		{"danh sách", "http://localhost:5173", []string{"http://localhost:5173,http://localhost:5173"}, false},
		{"path", "http://localhost:5173", []string{"http://localhost:5173/"}, false},
		{"userinfo", "http://localhost:5173", []string{"http://user@localhost:5173"}, false},
		{"query", "http://localhost:5173", []string{"http://localhost:5173?"}, false},
		{"fragment", "http://localhost:5173", []string{"http://localhost:5173#"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := newAuthProtection("dev", tt.appOrigin)
			if err != nil {
				t.Fatal(err)
			}
			// Giá trị Host và Forwarded không được cấp hoặc chặn quyền thay Origin.
			r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/auth/logout", nil)
			r.Header.Set("Forwarded", "host=localhost:5173;proto=http")
			r.Header.Set("X-Forwarded-Host", "localhost:5173")
			r.Header.Set("X-Forwarded-Proto", "http")
			for _, value := range tt.values {
				r.Header.Add("Origin", value)
			}
			if got := policy.allowsOrigin(r); got != tt.want {
				t.Fatalf("allowsOrigin = %t, cần %t", got, tt.want)
			}
		})
	}
}

func TestCSRFMutationProtection(t *testing.T) {
	secret := sequenceSecret()
	rotated := secret
	rotated[0]++
	policy, err := newAuthProtection("prod", "https://staffing.example")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, origin string
		tokens       []string
		cookieSecret auth.SessionSecret
		want         bool
	}{
		{"cùng phiên", "https://staffing.example", []string{csrfVector}, secret, true},
		{"token cũ cùng cookie mới", "https://staffing.example", []string{csrfVector}, rotated, false},
		{"token đúng origin sai", "https://staffing.example:8443", []string{csrfVector}, secret, false},
		{"token thiếu", "https://staffing.example", nil, secret, false},
		{"token sai định dạng", "https://staffing.example", []string{strings.Repeat("!", 43)}, secret, false},
		{"token trùng", "https://staffing.example", []string{csrfVector, csrfVector}, secret, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			r.Header.Set("Origin", tt.origin)
			for _, token := range tt.tokens {
				r.Header.Add("X-CSRF-Token", token)
			}
			if got := policy.allowsMutation(r, tt.cookieSecret); got != tt.want {
				t.Fatalf("allowsMutation = %t, cần %t", got, tt.want)
			}
		})
	}
}
