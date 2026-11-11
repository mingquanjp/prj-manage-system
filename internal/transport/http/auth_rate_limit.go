package http

import (
	"crypto/sha256"
	"net/http"
	"net/netip"
	"slices"
	"strconv"
	"sync"
	"time"

	"ai-staffing/internal/service/auth"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	loginIPLimit          = 30
	loginIPWindow         = 5 * time.Minute
	loginFailureLimit     = 5
	loginFailureWindow    = 5 * time.Minute
	passwordIPLimit       = 30
	passwordIPWindow      = 5 * time.Minute
	passwordFailureLimit  = 5
	passwordFailureWindow = 15 * time.Minute
	authBucketCapacity    = 4096
	authBucketSweepPeriod = time.Second
)

// authRateLimiter được tạo một lần và dùng chung giữa các request của hai handler.
type authRateLimiter struct {
	loginIPs       *rateWindow[netip.Addr]
	loginFailures  *rateWindow[[sha256.Size]byte]
	changeIPs      *rateWindow[netip.Addr]
	changeFailures *rateWindow[int]
}

func newAuthRateLimiter(clock auth.Clock) *authRateLimiter {
	return &authRateLimiter{
		loginIPs:       newRateWindow[netip.Addr](clock, loginIPLimit, loginIPWindow, authBucketCapacity),
		loginFailures:  newRateWindow[[sha256.Size]byte](clock, loginFailureLimit, loginFailureWindow, authBucketCapacity),
		changeIPs:      newRateWindow[netip.Addr](clock, passwordIPLimit, passwordIPWindow, authBucketCapacity),
		changeFailures: newRateWindow[int](clock, passwordFailureLimit, passwordFailureWindow, authBucketCapacity),
	}
}

// authPeerIP chỉ dùng địa chỉ peer do server cung cấp.
func authPeerIP(r *http.Request) (netip.Addr, bool) {
	peer, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, false
	}
	return peer.Addr().WithZone("").Unmap(), true
}

func (l *authRateLimiter) allowLoginIP(ip netip.Addr) time.Duration {
	return l.loginIPs.take(ip)
}

func (l *authRateLimiter) allowPasswordChangeIP(ip netip.Addr) time.Duration {
	return l.changeIPs.take(ip)
}

func (l *authRateLimiter) beginLoginAttempt(email string) (finish func(failed bool), wait time.Duration) {
	key := sha256.Sum256([]byte(auth.NormalizeLoginEmail(email)))
	return l.loginFailures.begin(key)
}

func (l *authRateLimiter) beginPasswordChangeAttempt(principal auth.Principal) (finish func(failed bool), wait time.Duration) {
	return l.changeFailures.begin(principal.UserID)
}

func writeAuthRateLimited(w http.ResponseWriter, r *http.Request, wait time.Duration) {
	seconds := int64(wait / time.Second)
	if wait%time.Second > 0 {
		seconds++
	}
	w.Header().Set("Retry-After", strconv.FormatInt(max(1, seconds), 10))
	writeAuthJSON(w, http.StatusTooManyRequests, authErrorResponse{
		Code:      "RATE_LIMITED",
		Message:   "Bạn đã thử quá nhiều lần. Vui lòng thử lại sau.",
		RequestID: middleware.GetReqID(r.Context()),
	})
}

type rateWindow[K comparable] struct {
	mu        sync.Mutex
	clock     auth.Clock
	limit     int
	window    time.Duration
	capacity  int
	nextSweep time.Time
	buckets   map[K]*rateBucket
}

type rateBucket struct {
	events  []time.Time
	pending int
}

func newRateWindow[K comparable](clock auth.Clock, limit int, window time.Duration, capacity int) *rateWindow[K] {
	return &rateWindow[K]{
		clock: clock, limit: limit, window: window, capacity: capacity,
		buckets: make(map[K]*rateBucket),
	}
}

func (w *rateWindow[K]) take(key K) time.Duration {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.clock.Now()
	bucket, wait := w.availableBucket(key, now)
	if bucket != nil {
		bucket.events = append(bucket.events, now)
	}
	return wait
}

// begin giữ lượt đến khi caller gọi finish. Caller phải defer finish(false)
// ngay sau khi được phép và gọi finish(true) nếu xác minh thực sự thất bại.
// Chỉ lần gọi finish đầu tiên có hiệu lực.
func (w *rateWindow[K]) begin(key K) (finish func(failed bool), wait time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()

	bucket, wait := w.availableBucket(key, w.clock.Now())
	if bucket == nil {
		return nil, wait
	}
	bucket.pending++
	finished := false
	return func(failed bool) {
		w.mu.Lock()
		defer w.mu.Unlock()
		if finished {
			return
		}
		finished = true
		bucket.pending--
		now := w.clock.Now()
		bucket.expire(now.Add(-w.window))
		if failed {
			bucket.events = append(bucket.events, now)
		}
		if len(bucket.events) == 0 && bucket.pending == 0 {
			delete(w.buckets, key)
		}
	}, 0
}

// availableBucket chỉ được gọi khi đang giữ mu.
func (w *rateWindow[K]) availableBucket(key K, now time.Time) (*rateBucket, time.Duration) {
	cutoff := now.Add(-w.window)
	if !now.Before(w.nextSweep) {
		for id, bucket := range w.buckets {
			bucket.expire(cutoff)
			if len(bucket.events) == 0 && bucket.pending == 0 {
				delete(w.buckets, id)
			}
		}
		w.nextSweep = now.Add(authBucketSweepPeriod)
	}
	bucket, exists := w.buckets[key]
	if !exists {
		if len(w.buckets) >= w.capacity {
			return nil, w.nextSweep.Sub(now)
		}
		bucket = &rateBucket{}
		w.buckets[key] = bucket
	}
	bucket.expire(cutoff)
	if len(bucket.events)+bucket.pending < w.limit {
		return bucket, 0
	}
	if bucket.pending > 0 {
		// Lượt đang xác minh có thể được trả lại bất kỳ lúc nào.
		return nil, time.Second
	}
	return nil, bucket.events[0].Add(w.window).Sub(now)
}

func (b *rateBucket) expire(cutoff time.Time) {
	expired := 0
	for expired < len(b.events) && !b.events[expired].After(cutoff) {
		expired++
	}
	b.events = slices.Delete(b.events, 0, expired)
}
