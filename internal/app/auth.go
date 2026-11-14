// Package app lắp ghép adapter, service và HTTP cho auth; không chứa quy tắc nghiệp vụ.
package app

import (
	"net/http"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	httptransport "ai-staffing/internal/transport/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthHandler là composition root dùng chung giữa server, local dev và integration.
func AuthHandler(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock) (http.Handler, error) {

	return httptransport.NewAuthRouter(pool, cfg.AppEnv, cfg.AppOrigin, authServices(cfg, pool, clock))
}

// LocalAuthHandler chỉ mở hai Origin loopback cố định cho FE và Swagger local.
func LocalAuthHandler(pool *pgxpool.Pool, clock auth.Clock) (http.Handler, error) {
	cfg := &config.Config{SessionTTL: 8 * time.Hour}
	return httptransport.NewLocalDevelopmentAuthRouter(pool, authServices(cfg, pool, clock))
}

func authServices(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock) httptransport.AuthServices {
	repository := postgres.NewAuthRepository(pool, clock)
	return httptransport.AuthServices{
		Sessions:       auth.NewService(repository, clock),
		Login:          auth.NewLoginService(repository, clock, cfg.SessionTTL),
		PasswordChange: auth.NewPasswordChangeService(repository, clock),
		Clock:          clock,
	}
}
