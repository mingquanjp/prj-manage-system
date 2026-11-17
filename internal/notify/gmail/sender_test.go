package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/organization"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status, Body: io.NopCloser(strings.NewReader(body)),
		Header: http.Header{"Content-Type": {"application/json"}},
	}
}

func testSender(t *testing.T, refresh, send roundTripFunc) *Sender {
	t.Helper()
	client := providerClient()
	client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.String() {
		case oauthEndpoint.TokenURL:
			if r.Method != http.MethodPost || r.ParseForm() != nil || r.PostForm.Get("grant_type") != "refresh_token" || r.PostForm.Get("refresh_token") != "test-refresh" {
				t.Fatal("Yêu cầu refresh sai contract")
			}
			return refresh(r)
		case sendURL:
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer test-access" ||
				r.Header.Get("Content-Type") != "application/json" || r.GetBody != nil {
				t.Fatal("Yêu cầu Gmail sai contract")
			}
			return send(r)
		default:
			t.Fatal("Đã phát yêu cầu sang đích ngoài dự kiến")
			return nil, nil
		}
	})
	return newSender(credentials{"test-client", "test-secret", "test-refresh", "sender@example.invalid"}, client, oauthEndpoint, sendURL)
}

func goodToken(*http.Request) (*http.Response, error) {
	return response(200, `{"access_token":"test-access","token_type":"Bearer","expires_in":3600}`), nil
}

type accountRepoFunc func(context.Context, organization.CreateAccountCommand) (organization.AccountProjection, error)

func (f accountRepoFunc) CreateAccount(ctx context.Context, command organization.CreateAccountCommand) (organization.AccountProjection, error) {
	return f(ctx, command)
}

func TestSendAccountMIMEAndTokenReuse(t *testing.T) {
	const to = "recipient@example.invalid"
	result, err := organization.NewAccountService(accountRepoFunc(func(context.Context, organization.CreateAccountCommand) (organization.AccountProjection, error) {
		return organization.AccountProjection{UserID: 1}, nil
	})).Create(context.Background(), organization.CreateAccountInput{ActorID: 2, Fullname: "Nhân viên thử", CompanyEmail: to, PositionID: 1})
	if err != nil {
		t.Fatal("Không tạo được dữ liệu thử nội bộ")
	}
	refreshes, sends := 0, 0
	s := testSender(t, func(r *http.Request) (*http.Response, error) {
		refreshes++
		return goodToken(r)
	}, func(r *http.Request) (*http.Response, error) {
		sends++
		var payload struct{ Raw string }
		if json.NewDecoder(r.Body).Decode(&payload) != nil {
			t.Fatal("JSON gửi sai")
		}
		raw, err := base64.RawURLEncoding.DecodeString(payload.Raw)
		if err != nil {
			t.Fatal("MIME không được mã hóa base64url")
		}
		message, err := mail.ReadMessage(bytes.NewReader(raw))
		if err != nil || message.Header.Get("To") != "<"+to+">" || message.Header.Get("From") != "<sender@example.invalid>" ||
			message.Header.Get("Content-Type") != "text/plain; charset=UTF-8" {
			t.Fatal("Header MIME sai")
		}
		body, err := io.ReadAll(quotedprintable.NewReader(message.Body))
		if err != nil || !strings.Contains(string(body), "Email đăng nhập: "+to) ||
			!strings.Contains(string(body), "Mật khẩu tạm: "+result.PasswordForDelivery().RevealForDelivery()) ||
			!strings.Contains(string(body), "đổi mật khẩu ở lần đăng nhập đầu tiên") {
			t.Fatal("Nội dung tài khoản sai")
		}
		return response(200, `{"id":"fake-message"}`), nil
	})
	for range 2 {
		status, err := s.SendAccount(context.Background(), to, result.PasswordForDelivery())
		if status != organization.MailAccepted || err != nil {
			t.Fatalf("kết quả %s, lỗi %v", status, err)
		}
	}
	s.token.Expiry = time.Now().Add(-time.Minute)
	if err := s.CheckConnection(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sends != 2 || refreshes != 2 {
		t.Fatalf("sends=%d refreshes=%d", sends, refreshes)
	}

	var log bytes.Buffer
	slog.New(slog.NewJSONHandler(&log, nil)).Info("kiểm", "sender", s)
	formatted := fmt.Sprintf("%v %+v %#v %v", s, s, s, &log)
	for _, secret := range []string{"test-access", "test-secret", "test-refresh", "sender@example.invalid"} {
		if strings.Contains(formatted, secret) {
			t.Fatal("Sender lộ dữ liệu khi format hoặc log")
		}
	}
}

func TestSendClassifiesFailureAndNeverReplays(t *testing.T) {
	for _, tc := range []struct {
		name    string
		refresh roundTripFunc
		send    roundTripFunc
		status  organization.MailStatus
		err     error
		calls   int
	}{
		{"invalid_grant", func(*http.Request) (*http.Response, error) {
			return response(400, `{"error":"invalid_grant","error_description":"private detail"}`), nil
		}, nil, organization.MailFailed, ErrReconnect, 0},
		{"refresh_timeout", blockUntilCanceled, nil, organization.MailFailed, ErrToken, 0},
		{"rejected", goodToken, func(*http.Request) (*http.Response, error) {
			return response(429, "private detail"), nil
		}, organization.MailFailed, ErrRejected, 1},
		{"server_error", goodToken, func(*http.Request) (*http.Response, error) {
			return response(500, "private detail"), nil
		}, organization.MailUnknown, ErrOutcomeUnknown, 1},
		{"lost_response", goodToken, func(*http.Request) (*http.Response, error) {
			return nil, errors.New("private detail")
		}, organization.MailUnknown, ErrOutcomeUnknown, 1},
		{"send_timeout", goodToken, blockUntilCanceled, organization.MailUnknown, ErrOutcomeUnknown, 1},
		{"missing_id", goodToken, func(*http.Request) (*http.Response, error) {
			return response(200, `{}`), nil
		}, organization.MailUnknown, ErrOutcomeUnknown, 1},
		{"redirect", goodToken, func(*http.Request) (*http.Response, error) {
			out := response(307, "")
			out.Header.Set("Location", "https://redirect.example.invalid/")
			return out, nil
		}, organization.MailUnknown, ErrOutcomeUnknown, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			s := testSender(t, tc.refresh, func(r *http.Request) (*http.Response, error) {
				calls++
				if tc.send == nil {
					t.Fatal("Đã gọi gửi khi token lỗi")
				}
				return tc.send(r)
			})
			s.refreshTimeout, s.sendTimeout = 50*time.Millisecond, 50*time.Millisecond
			status, err := s.SendCheck(context.Background(), "recipient@example.invalid")
			if status != tc.status || !errors.Is(err, tc.err) || calls != tc.calls {
				t.Fatalf("status=%s error=%v calls=%d", status, err, calls)
			}
		})
	}
}

func blockUntilCanceled(r *http.Request) (*http.Response, error) {
	<-r.Context().Done()
	return nil, r.Context().Err()
}
