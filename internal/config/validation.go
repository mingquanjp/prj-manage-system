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
// không kết nối mạng và không thay đổi input. LoadServer gọi sau khi parse.
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

func validateOrigin(raw, appEnv string) error {
	origin, err := ParseOrigin(raw)
	if err != nil {
		return err
	}
	if origin.Scheme == "http" && (appEnv != "dev" || !origin.IsLoopback()) {
		return errors.New("APP_ORIGIN phải dùng HTTPS, trừ HTTP loopback trong dev")
	}
	return nil
}

// Origin chứa bộ giá trị được so sánh khi kiểm nguồn request.
type Origin struct {
	Scheme string
	Host   string
	Port   int
}

func (o Origin) IsLoopback() bool {
	addr, _ := netip.ParseAddr(o.Host)
	return o.Host == "localhost" || addr.IsLoopback()
}

// ParseOrigin dùng chung cho cấu hình server và header Origin. Với tên miền
// quốc tế, cấu hình cần dùng dạng ASCII mà trình duyệt gửi (punycode).
func ParseOrigin(raw string) (Origin, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" ||
		u.Opaque != "" || u.User != nil || u.Path != "" || u.RawPath != "" ||
		strings.ContainsAny(raw, " \t\r\n,?#%\\") {
		// Lỗi từ net/url có thể chứa toàn bộ URL.
		return Origin{}, errors.New("APP_ORIGIN phải gồm http/https và host, không có userinfo, path, query hoặc fragment")
	}

	host := u.Hostname()
	addr, ipErr := netip.ParseAddr(host)
	if strings.HasPrefix(u.Host, "[") {
		if ipErr != nil || !addr.Is6() || addr.Zone() != "" {
			return Origin{}, errors.New("APP_ORIGIN có địa chỉ IPv6 không hợp lệ")
		}
	} else if strings.ContainsAny(host, "[]:") {
		return Origin{}, errors.New("APP_ORIGIN phải đặt địa chỉ IPv6 trong dấu ngoặc vuông")
	}
	if ipErr == nil {
		host = addr.String()
	} else if !isOriginDomain(host) {
		return Origin{}, errors.New("APP_ORIGIN cần tên miền ASCII (punycode nếu cần) hoặc địa chỉ IP đầy đủ")
	}

	port := 443
	if u.Scheme == "http" {
		port = 80
	}
	if value := u.Port(); value != "" {
		port, err = strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return Origin{}, errors.New("APP_ORIGIN có port ngoài khoảng 1 đến 65535")
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return Origin{}, errors.New("APP_ORIGIN có port rỗng")
	}
	return Origin{Scheme: u.Scheme, Host: strings.ToLower(host), Port: port}, nil
}

func isOriginDomain(host string) bool {
	const digits = "0123456789"
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	domain := strings.TrimSuffix(host, ".")
	if domain == "" || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") ||
		strings.Contains(domain, "..") || strings.Trim(domain, letters+digits+"-._") != "" {
		return false
	}
	// Trình duyệt diễn giải nhãn cuối toàn số hoặc dạng 0x… như IPv4.
	// IPv4 hợp lệ đã được netip.ParseAddr chấp nhận ở trên.
	lastLabel := strings.ToLower(domain[strings.LastIndexByte(domain, '.')+1:])
	if strings.Trim(lastLabel, digits) == "" {
		return false
	}
	hex, isHex := strings.CutPrefix(lastLabel, "0x")
	return !isHex || strings.Trim(hex, digits+"abcdef") != ""
}
