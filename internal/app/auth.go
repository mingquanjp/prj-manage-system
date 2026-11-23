// Package app lắp ghép adapter, service và HTTP cho auth/danh mục/hồ sơ/tài khoản; không chứa quy tắc nghiệp vụ.
package app

import (
	"errors"
	"net/http"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/notify/gmail"
	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/service/notification"
	"ai-staffing/internal/service/organization"
	"ai-staffing/internal/service/profilechange"
	httptransport "ai-staffing/internal/transport/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthHandler lắp auth, danh mục, hồ sơ và tài khoản, dùng chung giữa server và integration.
func AuthHandler(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock) (http.Handler, error) {

	sender, err := gmail.NewFromEnv()
	if err != nil {
		return nil, err
	}
	return AuthHandlerWithMail(cfg, pool, clock, sender)
}

// AuthHandlerWithMail cho phép test và composition cung cấp port gửi thư cụ thể.
func AuthHandlerWithMail(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock, sender organization.MailSender) (http.Handler, error) {
	services, err := applicationServices(cfg, pool, clock, sender)
	if err != nil {
		return nil, err
	}
	return httptransport.NewApplicationRouter(pool, cfg.AppEnv, cfg.AppOrigin, services)
}

// LocalAuthHandler chỉ mở hai Origin loopback cố định cho FE và Swagger local.
func LocalAuthHandler(pool *pgxpool.Pool, clock auth.Clock) (http.Handler, error) {
	sender, err := gmail.NewFromEnv()
	if err != nil {
		return nil, err
	}
	cfg := &config.Config{SessionTTL: 8 * time.Hour}
	services, err := applicationServices(cfg, pool, clock, sender)
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

func applicationServices(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock, sender organization.MailSender) (httptransport.AuthServices, error) {
	if sender == nil {
		return httptransport.AuthServices{}, errors.New("thiếu cấu hình gửi thư tài khoản")
	}
	services := authServices(cfg, pool, clock)
	// Mỗi thao tác DB có tối đa 10s; gửi thư diễn ra sau khi transaction đóng.
	repository, err := postgres.NewCatalogRepository(pool, 10*time.Second)
	if err != nil {
		return httptransport.AuthServices{}, err
	}
	services.Catalogs = organization.NewCatalogService(repository)
	profile, err := postgres.NewProfileRepository(pool, 10*time.Second)
	if err != nil {
		return httptransport.AuthServices{}, err
	}
	services.Profiles = organization.NewProfileService(profile)
	accounts, err := postgres.NewAccountRepository(pool, 10*time.Second)
	if err != nil {
		return httptransport.AuthServices{}, err
	}
	reads, err := postgres.NewAccountReadRepository(pool, 10*time.Second)
	if err != nil {
		return httptransport.AuthServices{}, err
	}
	services.Accounts = organization.NewAccountAPI(organization.NewAccountService(accounts), organization.NewAccountReadService(reads), sender)
	changes, err := postgres.NewProfileChangeRepository(pool, 10*time.Second)
	if err != nil {
		return httptransport.AuthServices{}, err
	}
	services.ProfileChanges = profilechange.NewService(changes)
	notifications, err := postgres.NewNotificationRepository(pool, 10*time.Second)
	if err != nil {
		return httptransport.AuthServices{}, err
	}
	services.Notifications = notification.NewService(notifications)
	return services, nil
}
