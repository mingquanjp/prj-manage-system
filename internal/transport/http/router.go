// Package http là tầng vận chuyển: định tuyến, middleware, chuyển đổi DTO.
// Tầng này gọi xuống service, không truy cập cơ sở dữ liệu trực tiếp.
package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// healthChecker cho phép kiểm tra health bằng fake mà không mở kết nối DB.
type healthChecker interface {
	Ping(context.Context) error
}

func NewRouter(pool healthChecker) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	// Chưa cấu hình proxy tin cậy: giữ RemoteAddr, không tin forwarded IP từ client.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api", func(api chi.Router) {
		api.Get("/healthz", healthHandler(pool))
	})

	return r
}

// healthHandler kiểm tra cả tiến trình lẫn kết nối cơ sở dữ liệu. Trên Neon
// free tier, lần gọi đầu sau khi compute ngủ có thể mất khoảng một giây.
func healthHandler(pool healthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		body := map[string]any{"status": "ok", "database": "ok"}
		code := http.StatusOK

		start := time.Now()
		if err := pool.Ping(ctx); err != nil {
			// Không ghi lỗi driver thô vì có thể chứa thông tin kết nối nhạy cảm.
			slog.ErrorContext(ctx, "healthz: không kết nối được cơ sở dữ liệu",
				"request_id", middleware.GetReqID(ctx))
			body["status"] = "degraded"
			body["database"] = "unavailable"
			code = http.StatusServiceUnavailable
		}
		body["database_latency_ms"] = time.Since(start).Milliseconds()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(body)
	}
}
