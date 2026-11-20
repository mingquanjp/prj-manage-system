package app

import (
	"testing"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/notify/gmail"
	"ai-staffing/internal/service/auth"
)

type testClock struct{}

func (testClock) Now() time.Time { return time.Now() }

var _ auth.Clock = testClock{}

func TestStartupRejectsMissingMailConfiguration(t *testing.T) {
	t.Setenv("GMAIL_OAUTH_CREDENTIALS_FILE", "")
	if _, err := AuthHandler(&config.Config{}, nil, testClock{}); err != gmail.ErrConfig {
		t.Fatal("Startup phải từ chối cấu hình mail thiếu trước khi mở router")
	}
	if _, err := LocalAuthHandler(nil, testClock{}); err != gmail.ErrConfig {
		t.Fatal("Local startup phải kiểm cấu hình mail")
	}
	if _, err := AuthHandlerWithMail(&config.Config{}, nil, testClock{}, nil); err == nil {
		t.Fatal("Composition phải yêu cầu port gửi thư")
	}
}
