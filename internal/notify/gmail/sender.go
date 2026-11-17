package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"ai-staffing/internal/service/organization"
	"golang.org/x/oauth2"
)

const (
	scope   = "https://www.googleapis.com/auth/gmail.send"
	sendURL = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"
)

var oauthEndpoint = oauth2.Endpoint{
	AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
	TokenURL:  "https://oauth2.googleapis.com/token",
	AuthStyle: oauth2.AuthStyleInParams,
}

var (
	ErrConfig         = errors.New("cấu hình Gmail không hợp lệ")
	ErrReconnect      = errors.New("cần cấp lại quyền gửi Gmail")
	ErrToken          = errors.New("không lấy được access token Gmail")
	ErrMessage        = errors.New("dữ liệu thư không hợp lệ")
	ErrRejected       = errors.New("Gmail từ chối yêu cầu gửi")
	ErrOutcomeUnknown = errors.New("chưa xác định kết quả gửi Gmail")
)

type Sender struct {
	config         oauth2.Config
	from           string
	client         *http.Client
	endpoint       string
	refreshTimeout time.Duration
	sendTimeout    time.Duration
	tokenLock      chan struct{}
	token          *oauth2.Token
}

var _ organization.MailSender = (*Sender)(nil)

func (*Sender) String() string         { return "[Gmail sender]" }
func (s *Sender) GoString() string     { return s.String() }
func (s *Sender) LogValue() slog.Value { return slog.StringValue(s.String()) }

func newSender(c credentials, client *http.Client, endpoint oauth2.Endpoint, sendEndpoint string) *Sender {
	return &Sender{
		config: oauth2.Config{ClientID: c.ClientID, ClientSecret: c.ClientSecret, Endpoint: endpoint, Scopes: []string{scope}},
		from:   c.From, client: client, endpoint: sendEndpoint,
		refreshTimeout: 3 * time.Second, sendTimeout: 7 * time.Second,
		tokenLock: make(chan struct{}, 1), token: &oauth2.Token{RefreshToken: c.RefreshToken},
	}
}

func providerClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// NewFromEnv đọc đường dẫn credential từ cấu hình tiến trình backend.
func NewFromEnv() (*Sender, error) {
	return NewFromFile(credentialPath())
}

func NewFromFile(path string) (*Sender, error) {
	var c credentials
	if readPrivateJSON(path, &c) != nil || !c.valid() {
		return nil, ErrConfig
	}
	return newSender(c, providerClient(), oauthEndpoint, sendURL), nil
}

func (s *Sender) tokenFor(ctx context.Context) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, s.refreshTimeout)
	defer cancel()
	select {
	case s.tokenLock <- struct{}{}:
		defer func() { <-s.tokenLock }()
	case <-ctx.Done():
		return nil, ErrToken
	}
	if ctx.Err() != nil {
		return nil, ErrToken
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.client)
	token, err := s.config.TokenSource(ctx, s.token).Token()
	if err != nil {
		var provider *oauth2.RetrieveError
		if errors.As(err, &provider) && provider.ErrorCode == "invalid_grant" {
			return nil, ErrReconnect
		}
		return nil, ErrToken
	}
	if !strings.EqualFold(token.Type(), "Bearer") || token.AccessToken == "" ||
		token.Expiry.IsZero() || !token.Valid() || ctx.Err() != nil {
		return nil, ErrToken
	}
	s.token = token
	return token, nil
}

// CheckConnection kiểm tra đường refresh, không gọi API gửi thư.
func (s *Sender) CheckConnection(ctx context.Context) error {
	_, err := s.tokenFor(ctx)
	return err
}

func (s *Sender) SendAccount(ctx context.Context, email string, password organization.TemporaryPassword) (organization.MailStatus, error) {
	value := password.RevealForDelivery()
	if value == "" {
		return organization.MailFailed, ErrMessage
	}
	return s.send(ctx, email, "Thông tin tài khoản TaskAssign",
		fmt.Sprintf("Chào bạn,\r\n\r\nEmail đăng nhập: %s\r\nMật khẩu tạm: %s\r\n\r\nBạn cần đổi mật khẩu ở lần đăng nhập đầu tiên trước khi sử dụng TaskAssign.\r\n", email, value))
}

// SendCheck gửi thư chẩn đoán có nội dung cố định.
func (s *Sender) SendCheck(ctx context.Context, to string) (organization.MailStatus, error) {
	return s.send(ctx, to, "Kiểm tra kết nối TaskAssign", "Đây là thư kiểm kết nối Gmail của TaskAssign.\r\n")
}

func (s *Sender) send(ctx context.Context, to, subject, body string) (organization.MailStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if !validAddress(to) {
		return organization.MailFailed, ErrMessage
	}
	payload := makePayload(s.from, to, subject, body)
	token, err := s.tokenFor(ctx)
	if err != nil {
		return organization.MailFailed, err
	}

	ctx, cancelSend := context.WithTimeout(ctx, s.sendTimeout)
	defer cancelSend()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return organization.MailFailed, ErrConfig
	}
	if ctx.Err() != nil {
		return organization.MailFailed, ErrRejected
	}
	// Bỏ khả năng đọc lại body để phương thức gửi chỉ thực hiện một lần.
	req.GetBody = nil
	req.Header.Set("Content-Type", "application/json")
	token.SetAuthHeader(req)
	resp, err := s.client.Do(req)
	if err != nil {
		return organization.MailUnknown, ErrOutcomeUnknown
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var reply struct{ ID string }
		if decodeLimitedJSON(resp.Body, &reply) != nil || strings.TrimSpace(reply.ID) == "" {
			return organization.MailUnknown, ErrOutcomeUnknown
		}
		return organization.MailAccepted, nil
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusTooManyRequests:
		return organization.MailFailed, ErrRejected
	default:
		return organization.MailUnknown, ErrOutcomeUnknown
	}
}

func validAddress(value string) bool {
	if len(value) > 320 || strings.ContainsAny(value, "\r\n\x00") {
		return false
	}
	address, err := mail.ParseAddress(value)
	return err == nil && address.Name == "" && address.Address == value
}

func makePayload(from, to, subject, body string) []byte {
	var message bytes.Buffer
	fmt.Fprintf(&message, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n",
		(&mail.Address{Address: from}).String(), (&mail.Address{Address: to}).String(),
		mime.QEncoding.Encode("UTF-8", subject), time.Now().UTC().Format(time.RFC1123Z))
	writer := quotedprintable.NewWriter(&message)
	_, _ = writer.Write([]byte(body))
	_ = writer.Close()
	data, _ := json.Marshal(struct {
		Raw string `json:"raw"`
	}{base64.RawURLEncoding.EncodeToString(message.Bytes())})
	return data
}
