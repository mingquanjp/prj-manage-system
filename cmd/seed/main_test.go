package main

import (
	"strings"
	"testing"
)

func TestRunRejectsSeedBeforeConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	for _, env := range []string{"", "prod"} {
		t.Setenv("APP_ENV", env)
		if err := run(); err == nil || !strings.Contains(err.Error(), "APP_ENV=dev") {
			t.Fatalf("phải từ chối seed trước khi đọc cấu hình DB: %v", err)
		}
	}
}
