package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"sync"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"

	"github.com/go-chi/chi/v5/middleware"
)

type rateTestClock struct{ now time.Time }

func (c *rateTestClock) Now() time.Time { return c.now }

func TestAuthRateLimiterPolicyBoundaries(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.10")
	principal := auth.Principal{UserID: 10, EmployeeID: 20}
	for _, tc := range []struct {
		name   string
		limit  int
		window time.Duration
		take   func(*authRateLimiter) time.Duration
	}{
		{"login IP", 30, 5 * time.Minute, func(l *authRateLimiter) time.Duration { return l.allowLoginIP(ip) }},
		{"change IP", 30, 5 * time.Minute, func(l *authRateLimiter) time.Duration { return l.allowPasswordChangeIP(ip) }},
		{"login email", 5, 5 * time.Minute, func(l *authRateLimiter) time.Duration {
			finish, wait := l.beginLoginAttempt("an@example.com")
			if finish != nil {
				finish(true)
			}
			return wait
		}},
		{"change user", 5, 15 * time.Minute, func(l *authRateLimiter) time.Duration {
			finish, wait := l.beginPasswordChangeAttempt(principal)
			if finish != nil {
				finish(true)
			}
			return wait
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Unix(1000, 0)
			clock := &rateTestClock{now: start}
			limiter := newAuthRateLimiter(clock)
			for i := range tc.limit {
				clock.now = start.Add(time.Duration(i) * time.Second)
				if wait := tc.take(limiter); wait != 0 {
					t.Fatalf("lượt %d bị chặn %v", i+1, wait)
				}
			}
			if wait, want := tc.take(limiter), tc.window-time.Duration(tc.limit-1)*time.Second; wait != want {
				t.Fatalf("lượt %d chờ %v, muốn %v", tc.limit+1, wait, want)
			}
			clock.now = start.Add(tc.window - time.Nanosecond)
			if wait := tc.take(limiter); wait != time.Nanosecond {
				t.Fatalf("ngay trước khi hết hạn chờ %v", wait)
			}
			clock.now = start.Add(tc.window)
			if wait := tc.take(limiter); wait != 0 {
				t.Fatalf("tại mốc hết hạn vẫn bị chặn %v", wait)
			}
			if wait := tc.take(limiter); wait != time.Second {
				t.Fatalf("cửa sổ trượt phải chỉ giải phóng một lượt, chờ %v", wait)
			}
		})
	}
}

func TestAuthRateLimiterSeparatesFlowsAndKeys(t *testing.T) {
	limiter := newAuthRateLimiter(&rateTestClock{now: time.Unix(1000, 0)})
	ip := netip.MustParseAddr("192.0.2.10")
	for i := range 30 {
		if wait := limiter.allowLoginIP(ip); wait != 0 {
			t.Fatalf("IP login lượt %d chờ %v", i+1, wait)
		}
	}
	if limiter.allowLoginIP(ip) == 0 || limiter.allowPasswordChangeIP(ip) != 0 ||
		limiter.allowLoginIP(netip.MustParseAddr("192.0.2.11")) != 0 {
		t.Fatal("ngân sách IP phải tách theo luồng và địa chỉ")
	}

	for range 5 {
		finish, wait := limiter.beginLoginAttempt("  An@Example.COM  ")
		if wait != 0 || finish == nil {
			t.Fatalf("không giữ được lượt login, chờ %v", wait)
		}
		finish(true)
	}
	if finish, wait := limiter.beginLoginAttempt("an@example.com"); finish != nil || wait != 5*time.Minute {
		t.Fatalf("hai biến thể email phải dùng chung bucket, chờ %v", wait)
	}
	if finish, wait := limiter.beginLoginAttempt("other@example.com"); wait != 0 || finish == nil {
		t.Fatalf("email khác bị chặn %v", wait)
	} else {
		finish(false)
	}

	for range 5 {
		finish, wait := limiter.beginPasswordChangeAttempt(auth.Principal{UserID: 10, EmployeeID: 20})
		if wait != 0 || finish == nil {
			t.Fatalf("không giữ được lượt đổi mật khẩu, chờ %v", wait)
		}
		finish(true)
	}
	if finish, wait := limiter.beginPasswordChangeAttempt(auth.Principal{UserID: 10, EmployeeID: 30}); finish != nil || wait != 15*time.Minute {
		t.Fatalf("cùng user phải dùng chung bucket, chờ %v", wait)
	}
	if finish, wait := limiter.beginPasswordChangeAttempt(auth.Principal{UserID: 11, EmployeeID: 20}); wait != 0 || finish == nil {
		t.Fatalf("user khác bị chặn %v", wait)
	} else {
		finish(false)
	}
}

func TestRateWindowConcurrentReservation(t *testing.T) {
	clock := &rateTestClock{now: time.Unix(1000, 0)}
	window := newRateWindow[string](clock, 5, 5*time.Minute, 2)
	for range 4 {
		finish, _ := window.begin("same")
		finish(true)
	}

	const contenders = 20
	start := make(chan struct{})
	ready := sync.WaitGroup{}
	ready.Add(contenders)
	type result struct {
		finish func(bool)
		wait   time.Duration
	}
	results := make(chan result, contenders)
	for range contenders {
		go func() {
			ready.Done()
			<-start
			finish, wait := window.begin("same")
			results <- result{finish, wait}
		}()
	}
	ready.Wait()
	close(start)
	var winners []func(bool)
	for range contenders {
		res := <-results
		if res.finish != nil && res.wait == 0 {
			winners = append(winners, res.finish)
		} else if res.finish != nil || res.wait != time.Second {
			t.Fatalf("kết quả giữ lượt không hợp lệ, chờ %v", res.wait)
		}
	}
	if len(winners) != 1 {
		t.Fatalf("có %d request giữ được lượt cuối", len(winners))
	}
	other, wait := window.begin("other")
	if other == nil || wait != 0 {
		t.Fatalf("xác minh đang chạy chặn định danh khác %v", wait)
	}
	other(false)

	// Lỗi DB hoặc hủy request dùng cùng đường giải phóng như xác minh đúng.
	winners[0](false)
	winners[0](true)
	finish, wait := window.begin("same")
	if finish == nil || wait != 0 {
		t.Fatalf("lượt không được trả lại hoặc bị ghi sai khi kết thúc lặp, chờ %v", wait)
	}
	// Ghi sai tại thời điểm có kết quả, kể cả khi cửa sổ cũ đã hết.
	clock.now = clock.now.Add(5 * time.Minute)
	finish(true)
	for range 4 {
		next, _ := window.begin("same")
		next(true)
	}
	if _, wait := window.begin("same"); wait != 5*time.Minute {
		t.Fatalf("lần sai không nằm trong cửa sổ hiện hành, chờ %v", wait)
	}
}

func TestAuthRateLimiterConcurrentIP(t *testing.T) {
	limiter := newAuthRateLimiter(&rateTestClock{now: time.Unix(1000, 0)})
	ip := netip.MustParseAddr("192.0.2.10")
	const requests = 64
	start := make(chan struct{})
	results := make(chan time.Duration, requests)
	for range requests {
		go func() {
			<-start
			results <- limiter.allowLoginIP(ip)
		}()
	}
	close(start)
	allowed := 0
	for range requests {
		switch wait := <-results; wait {
		case 0:
			allowed++
		case 5 * time.Minute:
		default:
			t.Fatalf("thời gian chờ IP = %v", wait)
		}
	}
	if allowed != 30 {
		t.Fatalf("cho qua %d request cùng IP, muốn 30", allowed)
	}
}

func TestRateWindowCapacityKeepsActiveBucketsAndReclaimsExpired(t *testing.T) {
	start := time.Unix(1000, 0)
	clock := &rateTestClock{now: start}
	window := newRateWindow[string](clock, 2, 5*time.Minute, 2)
	window.take("first")
	finish, _ := window.begin("pending")
	if window.take("overflow") != time.Second || len(window.buckets) != 2 {
		t.Fatal("vượt trần bucket")
	}
	clock.now = start.Add(250 * time.Millisecond)
	if window.take("overflow") != 750*time.Millisecond {
		t.Fatal("yêu cầu khi bão hòa phải đợi tới lượt dọn tiếp theo")
	}
	if window.take("first") != 0 || window.take("first") != 5*time.Minute-250*time.Millisecond {
		t.Fatal("bão hòa bộ nhớ làm mất bộ đếm hiện hành")
	}
	clock.now = start.Add(5 * time.Minute)
	if window.take("overflow") != time.Second {
		t.Fatal("bucket vẫn còn sự kiện mới hơn mốc hết hạn")
	}
	clock.now = clock.now.Add(time.Second)
	if window.take("overflow") != 0 || len(window.buckets) != 2 || window.buckets["pending"] == nil {
		t.Fatal("không thu hồi đúng bucket hết hạn hoặc làm mất lượt đang giữ")
	}
	finish(false)
	if window.take("next") != 0 || len(window.buckets) != 2 {
		t.Fatal("bucket rỗng chưa được thu hồi khi trả lượt")
	}
}

func TestAuthPeerIPIgnoresForwardedHeaders(t *testing.T) {
	for _, tc := range []struct {
		remote, want string
	}{
		{"192.0.2.10:12345", "192.0.2.10"},
		{"192.0.2.10:54321", "192.0.2.10"},
		{"[::ffff:192.0.2.10]:12345", "192.0.2.10"},
		{"[2001:0db8::1]:12345", "2001:db8::1"},
		{"invalid-peer", ""},
	} {
		t.Run(tc.remote, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
			request.RemoteAddr = tc.remote
			request.Header.Set("X-Forwarded-For", "198.51.100.2")
			request.Header.Set("X-Real-IP", "198.51.100.3")
			request.Header.Set("Forwarded", "for=198.51.100.4")
			ip, ok := authPeerIP(request)
			if tc.want == "" {
				if ok {
					t.Fatalf("peer sai lại được đọc là %v", ip)
				}
				return
			}
			if !ok || ip.String() != tc.want {
				t.Fatalf("IP = %v, muốn %s", ip, tc.want)
			}
		})
	}
}

func TestWriteAuthRateLimited(t *testing.T) {
	for _, tc := range []struct {
		wait time.Duration
		want string
	}{{time.Nanosecond, "1"}, {2 * time.Second, "2"}, {2*time.Second + time.Nanosecond, "3"}} {
		t.Run(tc.wait.String(), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
			request.Header.Set(middleware.RequestIDHeader, "client-marker")
			recorder := httptest.NewRecorder()
			handler := serverAuthRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeAuthRateLimited(w, r, tc.wait)
			}))
			handler.ServeHTTP(recorder, request)
			if recorder.Code != 429 || recorder.Header().Get("Retry-After") != tc.want ||
				recorder.Header().Get("Cache-Control") != "no-store" ||
				recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Fatalf("status = %d, header = %v", recorder.Code, recorder.Header())
			}
			var body authErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != "RATE_LIMITED" || body.Message == "" || body.RequestID == "" || body.RequestID == "client-marker" {
				t.Fatalf("response không khớp contract: %+v", body)
			}
		})
	}
}
