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
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Cố ý KHÔNG dùng middleware.RealIP. Nó ghi đè r.RemoteAddr bằng giá trị
	// X-Forwarded-For do client gửi mà không kiểm tra có proxy tin cậy hay không,
	// nên bất kỳ ai cũng giả được địa chỉ nguồn (GHSA-3fxj-6jh8-hvhx). Việc giới
	// hạn tần suất đăng nhập ở M1 dựa vào địa chỉ nguồn, nên phải giữ RemoteAddr
	// thật. Nếu về sau triển khai sau reverse proxy thì mới thêm lại, kèm danh
	// sách trắng các proxy được phép.

	r.Route("/api", func(api chi.Router) {
		api.Get("/healthz", healthHandler(pool))
	})

	return r
}

// healthHandler kiểm tra cả tiến trình lẫn kết nối cơ sở dữ liệu. Trên Neon
// free tier, lần gọi đầu sau khi compute ngủ có thể mất khoảng một giây.
//
// Endpoint này không yêu cầu đăng nhập nên chỉ trả về trạng thái chung chung.
// Lỗi thật của pgx thường chứa hostname, tên người dùng và tên cơ sở dữ liệu,
// vì vậy chi tiết chỉ được ghi vào log phía máy chủ.
func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		body := map[string]any{"status": "ok", "database": "ok"}
		code := http.StatusOK

		start := time.Now()
		if err := pool.Ping(ctx); err != nil {
			slog.ErrorContext(ctx, "healthz: không ping được cơ sở dữ liệu",
				"err", err, "request_id", middleware.GetReqID(ctx))

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
