//go:build integration

package http

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
)

func TestSelfProfileRejectsStaleMiddlewarePrincipal(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "stale-principal@example.invalid"})
	r, err := postgres.NewProfileRepository(f.DB.Conn, 5*time.Second)
	authdb.Require(t, "profile adapter thật", err)
	// Giữ principal cũ như thời điểm Authenticate đã xong; nguồn DB đổi trước ReadSelf.
	session := catalogHR()
	session.principal.UserID = a.UserID
	session.principal.EmployeeID = a.EmployeeID
	h := profileHTTPRouter(t, session, org.NewProfileService(r), "test")
	path := "/api/employees/" + strconv.Itoa(a.EmployeeID)
	for _, tc := range []struct {
		sql    string
		status int
		code   string
	}{
		{`UPDATE users SET must_change_password=true WHERE user_id=$1`, 403, "PASSWORD_CHANGE_REQUIRED"},
		{`UPDATE users SET must_change_password=false,status='LOCKED' WHERE user_id=$1`, 401, "UNAUTHENTICATED"},
		{`UPDATE users SET status='ACTIVE' WHERE user_id=$1`, 200, ""},
		{`UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, 401, "UNAUTHENTICATED"},
	} {
		_, err = f.DB.Conn.Exec(t.Context(), tc.sql, a.UserID)
		authdb.Require(t, "đổi nền trước handler", err)
		req := passwordChangeRequest("GET", path, "")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assertCatalogHTTP(t, w, tc.status, tc.code)
	}
}

func TestSelfProfileStorageTimeoutIs503(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "timeout-profile@example.invalid"})
	pool := f.DB.Pool(t)
	defer pool.Close()
	locker, err := f.DB.Conn.Begin(t.Context())
	authdb.Require(t, "mở khóa DB case", err)
	defer locker.Rollback(t.Context())
	_, err = locker.Exec(t.Context(), `LOCK TABLE departments IN ACCESS EXCLUSIVE MODE`)
	authdb.Require(t, "barrier khóa trước query", err)
	repository, err := postgres.NewProfileRepository(pool, 100*time.Millisecond)
	authdb.Require(t, "profile có deadline", err)
	session := catalogHR()
	session.principal.UserID = a.UserID
	session.principal.EmployeeID = a.EmployeeID
	h := profileHTTPRouter(t, session, org.NewProfileService(repository), "test")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("GET", "/api/employees/"+strconv.Itoa(a.EmployeeID), ""))
	assertCatalogHTTP(t, w, 503, "SERVICE_UNAVAILABLE")
}
