package config

import (
	"errors"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const sessionTTL = 8 * time.Hour

// ValidateServer kiểm tra cấu hình đã có kiểu dữ liệu, không đọc môi trường/file,
// không kết nối mạng và không thay đổi input. Load chưa gọi hàm này ở DATN-3.
func ValidateServer(cfg Config) error {
	if cfg.AppEnv != "dev" && cfg.AppEnv != "prod" {
		return errors.New("APP_ENV phải là dev hoặc prod")
	}
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return errors.New("HTTP_PORT phải từ 1 đến 65535")
	}
	if err := validateOrigin(cfg.AppOrigin, cfg.AppEnv); err != nil {
		return err
	}
	if cfg.SessionTTL != sessionTTL {
		return errors.New("SESSION_TTL phải bằng 8h")
	}
	return nil
}

func validateOrigin(origin, appEnv string) error {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "http" && u.Scheme != "https" || u.Hostname() == "" ||
		u.Opaque != "" || u.User != nil || u.Path != "" || u.RawPath != "" ||
		strings.ContainsAny(origin, "?#") {
		// Lỗi từ net/url có thể chứa toàn bộ URL, nên không trả hoặc wrap lỗi đó.
		return errors.New("APP_ORIGIN phải gồm http/https và host, không có userinfo, path, query hoặc fragment")
	}

	host := u.Hostname()
	addr, ipErr := netip.ParseAddr(host)
	if strings.HasPrefix(u.Host, "[") {
		if ipErr != nil || !addr.Is6() || addr.Zone() != "" {
			return errors.New("APP_ORIGIN có địa chỉ IPv6 không hợp lệ")
		}
	} else if strings.Contains(host, ":") {
		return errors.New("APP_ORIGIN phải đặt địa chỉ IPv6 trong dấu ngoặc vuông")
	}

	if port := u.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return errors.New("APP_ORIGIN có port ngoài khoảng 1 đến 65535")
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return errors.New("APP_ORIGIN có port rỗng")
	}

	// Chỉ nhận loopback theo giá trị địa chỉ; không phân giải DNS để cấp ngoại lệ.
	loopback := strings.EqualFold(host, "localhost") || (ipErr == nil && addr.IsLoopback())
	if u.Scheme == "http" && (appEnv != "dev" || !loopback) {
		return errors.New("APP_ORIGIN phải dùng HTTPS, trừ HTTP loopback trong dev")
	}
	return nil
}
