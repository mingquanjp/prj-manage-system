package http

import (
	"errors"
	"net/http"

	"ai-staffing/internal/config"
	"ai-staffing/internal/service/auth"
)

// AuthServices là các use case được composition root lắp trước khi mở server.
type AuthServices struct {
	Sessions       authSessionService
	Login          loginUseCase
	PasswordChange passwordChangeUseCase
	Clock          auth.Clock
}

// NewAuthRouter yêu cầu đủ năm route; dependency thiếu phải chặn startup.
func NewAuthRouter(health healthChecker, appEnv, appOrigin string, services AuthServices) (http.Handler, error) {
	return newAuthRouter(health, appEnv, appOrigin, services, false)
}

// NewLocalDevelopmentAuthRouter chỉ dùng command localdev với hai origin loopback cố định.
// Không đọc Host/Forwarded để cấp Origin; production vẫn chỉ nhận APP_ORIGIN duy nhất.
func NewLocalDevelopmentAuthRouter(health healthChecker, services AuthServices) (http.Handler, error) {
	return newAuthRouter(health, "dev", "http://127.0.0.1:5174", services, true)
}

func newAuthRouter(health healthChecker, appEnv, appOrigin string, services AuthServices, localDocs bool) (http.Handler, error) {
	if services.Sessions == nil || services.Login == nil || services.PasswordChange == nil || services.Clock == nil {
		return nil, errors.New("thiếu dependency auth khi dựng router")
	}
	protection, err := newAuthProtection(appEnv, appOrigin)
	if err != nil {
		return nil, errors.New("origin auth không hợp lệ khi dựng router")
	}
	if localDocs {
		origin, err := config.ParseOrigin("http://127.0.0.1:8080")
		if err != nil {
			return nil, errors.New("origin tài liệu local không hợp lệ")
		}
		protection.localDocsOrigin = &origin
	}
	limiter := newAuthRateLimiter(services.Clock)
	router := newRouter(health, services.Sessions, protection,
		&loginRoute{service: services.Login, limiter: limiter, clock: services.Clock},
		&passwordChangeRoute{service: services.PasswordChange, limiter: limiter})
	// Dev tường minh mới mở tài liệu; không nới allowlist của phiên hạn chế.
	if appEnv == "dev" {
		router.HandleFunc("/docs", swaggerHandler)
		router.HandleFunc("/docs/*", swaggerHandler)
	}
	return router, nil
}
