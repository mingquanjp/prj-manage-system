package config

import (
	"errors"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

// ErrDatabaseURLInvalid không chứa URL hoặc nguyên nhân parse có thể lộ credential.
var ErrDatabaseURLInvalid = errors.New("DB_URL_INVALID: DATABASE_URL phải là URL PostgreSQL một host được hỗ trợ")

// DatabaseEndpoint được tạo bởi ParseDatabaseURL; các URL chứa bí mật, không log struct này.
type DatabaseEndpoint struct {
	RuntimeURL string
	DirectURL  string
	Pooled     bool
}

// ParseDatabaseURL chỉ phân tích cú pháp, không đọc file/môi trường, DNS hoặc mở DB.
func ParseDatabaseURL(raw string) (DatabaseEndpoint, error) {
	invalid := func() (DatabaseEndpoint, error) { return DatabaseEndpoint{}, ErrDatabaseURLInvalid }
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Opaque != "" || u.Hostname() == "" || strings.Contains(raw, "#") {
		return invalid()
	}
	host := u.Hostname()
	if strings.ContainsAny(host, ",/\\ \t\r\n") {
		return invalid()
	}
	addr, ipErr := netip.ParseAddr(host)
	if ipErr != nil {
		for _, label := range strings.Split(strings.TrimSuffix(strings.ToLower(host), "."), ".") {
			if !validDNSLabel(label) {
				return invalid()
			}
		}
	}
	if strings.HasPrefix(u.Host, "[") {
		if ipErr != nil || !addr.Is6() || addr.Zone() != "" {
			return invalid()
		}
	} else if strings.Contains(host, ":") {
		return invalid()
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return invalid()
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return invalid()
	}
	// Không cho query đổi endpoint/định danh hoặc chuyển sang service/socket/multi-host.
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return invalid()
	}
	for key := range query {
		switch strings.ToLower(key) {
		case "host", "hostaddr", "port", "user", "password", "dbname", "database", "service", "servicefile":
			return invalid()
		}
	}
	result := DatabaseEndpoint{RuntimeURL: raw, DirectURL: raw}
	labels := strings.Split(strings.ToLower(host), ".")
	// Chỉ hỗ trợ mẫu hostname AWS Neon hiện dùng; không suy mọi miền có chữ pooler.
	if len(labels) != 5 || labels[2] != "aws" || labels[3] != "neon" || labels[4] != "tech" || !validDNSLabel(labels[1]) {
		return result, nil
	}
	endpoint := labels[0]
	if !strings.HasPrefix(endpoint, "ep-") || !strings.HasSuffix(endpoint, "-pooler") {
		return result, nil
	}
	directLabel := strings.TrimSuffix(endpoint, "-pooler")
	if len(directLabel) <= 3 || !validDNSLabel(directLabel) {
		return result, nil
	}
	directHost := host[:len(directLabel)] + host[len(endpoint):]
	if port := u.Port(); port != "" {
		directHost = net.JoinHostPort(directHost, port)
	}
	// Thay đúng span host đã parse; giữ nguyên từng byte userinfo/path/query.
	authorityStart := strings.Index(raw, "://") + 3
	authorityEnd := len(raw)
	if i := strings.IndexAny(raw[authorityStart:], "/?"); i >= 0 {
		authorityEnd = authorityStart + i
	}
	authority := raw[authorityStart:authorityEnd]
	hostStart := authorityStart
	if i := strings.LastIndex(authority, "@"); i >= 0 {
		hostStart += i + 1
	}
	if raw[hostStart:authorityEnd] != u.Host {
		return invalid()
	}
	result.DirectURL = raw[:hostStart] + directHost + raw[authorityEnd:]
	result.Pooled = true
	return result, nil
}

func validDNSLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for _, c := range label {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}
