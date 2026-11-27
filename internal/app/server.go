package app

import (
	"net/http"
	"time"
)

// NewHTTPServer dùng cùng giới hạn kết nối cho server và localdev.
// Hạn đọc toàn request ngăn body dang dở giữ kết nối; xử lý DB/mail có deadline riêng.
func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
