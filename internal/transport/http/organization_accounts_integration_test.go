//go:build integration

package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type accountIntegrationMailer struct {
	calls  int
	status org.MailStatus
	err    error
	check  func(context.Context, string, org.TemporaryPassword)
}

func (m *accountIntegrationMailer) SendAccount(ctx context.Context, email string, password org.TemporaryPassword) (org.MailStatus, error) {
	m.calls++
	if m.check != nil {
		m.check(ctx, email, password)
	}
	return m.status, m.err
}

func accountIntegrationRouter(t *testing.T, f *authdb.Fixture, actor authdb.Account, creator org.AccountRepository, mail org.MailSender) http.Handler {
	t.Helper()
	reader, err := postgres.NewAccountReadRepository(f.DB.Conn, 5*time.Second)
	authdb.Require(t, "dựng adapter đọc tài khoản", err)
	api := org.NewAccountAPI(org.NewAccountService(creator), org.NewAccountReadService(reader), mail)
	session := catalogHR()
	session.principal.UserID = actor.UserID
	session.principal.EmployeeID = actor.EmployeeID
	session.principal.SystemRoles = []string{"ADMIN"}
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	router, err := NewAuthRouter(fakeHealth{}, "dev", loginTestOrigin, AuthServices{
		Sessions: session, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour),
		PasswordChange: session, Clock: clock, Accounts: api,
	})
	authdb.Require(t, "dựng router tài khoản", err)
	return router
}

func accountIntegrationCreator(t *testing.T, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) *postgres.AccountRepository {
	t.Helper()
	repo, err := postgres.NewAccountRepository(db, 5*time.Second)
	authdb.Require(t, "dựng create adapter", err)
	return repo
}

func accountIntegrationBody(t *testing.T, f *authdb.Fixture, email string) string {
	t.Helper()
	var position int
	authdb.Require(t, "đọc ID vị trí fixture", f.DB.Conn.QueryRow(t.Context(), `SELECT position_id FROM public.positions ORDER BY position_id LIMIT 1`).Scan(&position))
	body, err := json.Marshal(map[string]any{
		"fullname": "Nhân viên tổng hợp", "company_email": email, "position_id": position,
		"department_id": nil, "job_level_id": nil, "phone": "0123456789", "address": "Địa chỉ tổng hợp",
		"system_roles": []string{"HR", "DIRECTOR"},
	})
	authdb.Require(t, "dựng body tổng hợp", err)
	return string(body)
}

func TestAccountHTTPPersistsBeforeMailAndKeeps201ForMailOutcomes(t *testing.T) {
	for _, status := range []org.MailStatus{org.MailAccepted, org.MailFailed, org.MailUnknown} {
		t.Run(string(status), func(t *testing.T) {
			f := authdb.NewFixture(t)
			actor := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
			observer := f.DB.Connect(t)
			const email = "new-account@example.invalid"
			mailer := &accountIntegrationMailer{status: status}
			if status != org.MailAccepted {
				mailer.err = errors.New("synthetic private provider error")
			}
			mailer.check = func(ctx context.Context, gotEmail string, password org.TemporaryPassword) {
				if gotEmail != email || password.RevealForDelivery() == "" {
					t.Fatal("mail không nhận credential nội bộ đúng account")
				}
				var hash string
				var limited bool
				var audit, roleCount int
				authdb.Require(t, "connection độc lập đọc account đã commit trước mail", observer.QueryRow(ctx, `SELECT u.password_hash,u.must_change_password,
 (SELECT count(*) FROM public.audit_events a WHERE a.entity_id=e.employee_id AND a.operation='CREATE_ACCOUNT'),
 (SELECT count(*) FROM public.user_roles r WHERE r.user_id=u.user_id)
 FROM public.users u JOIN public.employees e ON e.user_id=u.user_id WHERE e.email=$1`, email).Scan(&hash, &limited, &audit, &roleCount))
				if !limited || audit != 1 || roleCount != 2 || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password.RevealForDelivery())) != nil {
					t.Fatal("user/employee/roles/audit chưa commit đầy đủ trước mail")
				}
			}
			h := accountIntegrationRouter(t, f, actor, accountIntegrationCreator(t, f.DB.Conn), mailer)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", accountIntegrationBody(t, f, email)))
			assertCatalogHTTP(t, w, 201, "")
			var result struct {
				Created bool           `json:"account_created"`
				Account map[string]any `json:"account"`
				Mail    struct {
					Status org.MailStatus `json:"status"`
				} `json:"mail"`
			}
			authdb.Require(t, "đọc create response", json.Unmarshal(w.Body.Bytes(), &result))
			if !result.Created || result.Mail.Status != status || result.Account["company_email"] != email || len(result.Account) != 8 || mailer.calls != 1 {
				t.Fatal("201 hoặc projection không đúng contract")
			}
			for _, hidden := range []string{"password", "hash", "phone", "address", "synthetic private", "0123456789", "Địa chỉ tổng hợp"} {
				if strings.Contains(w.Body.String(), `"`+hidden+`"`) || strings.Contains(w.Body.String(), "synthetic private") {
					t.Fatal("create response lộ dữ liệu ngoài allowlist")
				}
			}
			var sessions int
			authdb.Require(t, "kiểm không tạo session", observer.QueryRow(t.Context(), `SELECT count(*) FROM public.auth_sessions`).Scan(&sessions))
			if sessions != 0 {
				t.Fatal("create đã sinh session")
			}
		})
	}
}

func TestAccountHTTPRejectsDatabaseFailuresBeforeMail(t *testing.T) {
	for _, mode := range []string{"missing FK", "duplicate email", "revoked admin", "limited", "inactive"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			actor := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
			body := accountIntegrationBody(t, f, "candidate@example.invalid")
			want, code := 400, "VALIDATION_ERROR"
			switch mode {
			case "missing FK":
				var input map[string]any
				authdb.Require(t, "đọc input", json.Unmarshal([]byte(body), &input))
				input["position_id"] = 2147483647
				b, err := json.Marshal(input)
				authdb.Require(t, "encode FK sai", err)
				body = string(b)
			case "duplicate email":
				f.AddAccount(t, authdb.AccountOptions{Email: "candidate@example.invalid"})
				want, code = 409, "EMAIL_CONFLICT"
			case "revoked admin":
				_, err := f.DB.Conn.Exec(t.Context(), `DELETE FROM public.user_roles WHERE user_id=$1`, actor.UserID)
				authdb.Require(t, "đổi quyền sau snapshot middleware", err)
				want, code = 403, "FORBIDDEN"
			case "limited":
				_, err := f.DB.Conn.Exec(t.Context(), `UPDATE public.users SET must_change_password=true WHERE user_id=$1`, actor.UserID)
				authdb.Require(t, "đổi cờ sau snapshot middleware", err)
				want, code = 403, "PASSWORD_CHANGE_REQUIRED"
			case "inactive":
				_, err := f.DB.Conn.Exec(t.Context(), `UPDATE public.employees SET employment_status='INACTIVE' WHERE user_id=$1`, actor.UserID)
				authdb.Require(t, "đổi active sau snapshot middleware", err)
				want, code = 401, "UNAUTHENTICATED"
			}
			mailer := &accountIntegrationMailer{status: org.MailAccepted}
			h := accountIntegrationRouter(t, f, actor, accountIntegrationCreator(t, f.DB.Conn), mailer)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", body))
			assertCatalogHTTP(t, w, want, code)
			var count int
			authdb.Require(t, "đếm audit sau create lỗi", f.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM public.audit_events WHERE operation='CREATE_ACCOUNT'`).Scan(&count))
			if mailer.calls != 0 || count != 0 {
				t.Fatal("lỗi create đã gửi mail hoặc ghi audit")
			}
		})
	}
}

// Chỉ thay phản hồi sau PostgreSQL commit thật để kiểm kết quả HTTP chưa xác định.
type accountHTTPLostReplyDB struct{ *pgx.Conn }

func (db accountHTTPLostReplyDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.Conn.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return accountHTTPLostReplyTx{tx}, nil
}

type accountHTTPLostReplyTx struct{ pgx.Tx }

func (tx accountHTTPLostReplyTx) Commit(ctx context.Context) error {
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	return io.EOF
}

func TestAccountHTTPUnknownCommitSkipsMailAndCanReconcileByList(t *testing.T) {
	f := authdb.NewFixture(t)
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
	mailer := &accountIntegrationMailer{status: org.MailAccepted}
	h := accountIntegrationRouter(t, f, actor, accountIntegrationCreator(t, accountHTTPLostReplyDB{f.DB.Conn}), mailer)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", accountIntegrationBody(t, f, "uncertain@example.invalid")))
	assertCatalogHTTP(t, w, 503, "SERVICE_UNAVAILABLE")
	if mailer.calls != 0 {
		t.Fatal("commit chưa xác định vẫn gửi mail")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("GET", "/api/users?q=uncertain%40example.invalid", ""))
	assertCatalogHTTP(t, w, 200, "")
	var page struct {
		Items []map[string]any `json:"items"`
	}
	authdb.Require(t, "đọc account đối soát", json.Unmarshal(w.Body.Bytes(), &page))
	if len(page.Items) != 1 || page.Items[0]["company_email"] != "uncertain@example.invalid" || len(page.Items[0]) != 8 {
		t.Fatal("list không đối soát được commit thực")
	}
	if strings.Contains(w.Body.String(), `"mail"`) || strings.Contains(w.Body.String(), `"password_hash"`) {
		t.Fatal("list suy delivery hoặc lộ credential")
	}
}

func TestAccountHTTPPreviewDoesNotConsumeAndListUsesAllowedProjection(t *testing.T) {
	f := authdb.NewFixture(t)
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
	mailer := &accountIntegrationMailer{status: org.MailAccepted}
	h := accountIntegrationRouter(t, f, actor, accountIntegrationCreator(t, f.DB.Conn), mailer)
	preview := func() string {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, passwordChangeRequest("GET", "/api/employees/next-code", ""))
		assertCatalogHTTP(t, w, 200, "")
		var body struct {
			Code    string `json:"employee_code"`
			Preview bool   `json:"is_preview"`
		}
		authdb.Require(t, "đọc preview", json.Unmarshal(w.Body.Bytes(), &body))
		if !body.Preview || body.Code == "" {
			t.Fatal("preview sai contract")
		}
		return body.Code
	}
	first := preview()
	if preview() != first {
		t.Fatal("preview đã tiêu số")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", accountIntegrationBody(t, f, "created@example.invalid")))
	assertCatalogHTTP(t, w, 201, "")
	var created struct {
		Account struct {
			Code string `json:"employee_code"`
		} `json:"account"`
	}
	authdb.Require(t, "đọc mã commit", json.Unmarshal(w.Body.Bytes(), &created))
	if created.Account.Code != first || preview() == first {
		t.Fatal("preview/cấp số không khớp tuần tự")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("GET", "/api/users?page=1&page_size=1&q=created", ""))
	assertCatalogHTTP(t, w, 200, "")
	var page struct {
		Items      []map[string]any `json:"items"`
		Pagination struct {
			Total int64 `json:"total"`
		} `json:"pagination"`
	}
	authdb.Require(t, "đọc trang account", json.Unmarshal(w.Body.Bytes(), &page))
	if page.Pagination.Total != 1 || len(page.Items) != 1 || fmt.Sprint(page.Items[0]["system_roles"]) != "[HR DIRECTOR]" {
		t.Fatal("list roles/search/pagination sai")
	}
}

func TestAccountHTTPActualBodyLimitBeforeDatabaseWrite(t *testing.T) {
	f := authdb.NewFixture(t)
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
	mailer := &accountIntegrationMailer{status: org.MailAccepted}
	h := accountIntegrationRouter(t, f, actor, accountIntegrationCreator(t, f.DB.Conn), mailer)
	body := accountIntegrationBody(t, f, "boundary@example.invalid")
	body += strings.Repeat(" ", 8192-len(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", body+" "))
	assertCatalogHTTP(t, w, 413, "PAYLOAD_TOO_LARGE")
	var count int
	authdb.Require(t, "đếm account trước body hợp lệ", f.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM public.employees`).Scan(&count))
	if count != 1 || mailer.calls != 0 {
		t.Fatal("body 8193 byte đã ghi DB hoặc gửi mail")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", body))
	assertCatalogHTTP(t, w, 201, "")
	if mailer.calls != 1 {
		t.Fatal("body 8192 byte không đi tới mail đúng một lần")
	}
}
