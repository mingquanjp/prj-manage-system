// Package app lắp ghép adapter, service và HTTP cho auth/danh mục; không chứa quy tắc nghiệp vụ.
package app

import (
	"net/http"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/service/organization"
	httptransport "ai-staffing/internal/transport/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthHandler lắp auth và danh mục, dùng chung giữa server và integration.
func AuthHandler(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock) (http.Handler, error) {

	services, err := applicationServices(cfg, pool, clock)
	if err != nil {
		return nil, err
	}
	return httptransport.NewApplicationRouter(pool, cfg.AppEnv, cfg.AppOrigin, services)
}

// LocalAuthHandler chỉ mở hai Origin loopback cố định cho FE và Swagger local.
func LocalAuthHandler(pool *pgxpool.Pool, clock auth.Clock) (http.Handler, error) {
	cfg := &config.Config{SessionTTL: 8 * time.Hour}
	services, err := applicationServices(cfg, pool, clock)
	if err != nil {
		return nil, err
	}
	return httptransport.NewLocalDevelopmentApplicationRouter(pool, services)
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

func applicationServices(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock) (httptransport.AuthServices, error) {
	services := authServices(cfg, pool, clock)
	// Nằm trong timeout HTTP 30s, không thêm cấu hình môi trường riêng.
	repository, err := postgres.NewCatalogRepository(pool, 10*time.Second)
	if err != nil {
		return httptransport.AuthServices{}, err
	}
	services.Catalogs = organization.NewCatalogService(repository)
	return services, nil
}
