//go:build integration

package postgres_test

import (
	"crypto/sha256"
	"errors"
	"reflect"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
	"golang.org/x/crypto/bcrypt"
)

func addPasswordChangeSession(t *testing.T, f *authdb.Fixture, userID int, token string, expiry time.Time, revoked *time.Time) (auth.SessionHash, int) {
	t.Helper()
	hash := auth.SessionHash(sha256.Sum256([]byte(token)))
	id := f.AddSession(t, userID, hash, sessionTestTime.Add(-2*time.Hour), expiry, revoked)
	return hash, id
}

func replacementHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	authdb.Require(t, "tạo hash thay thế cho fixture", err)
	return string(hash)
}

func assertPasswordChangeState(t *testing.T, f *authdb.Fixture, userID int, hash string, mustChange bool, sessionCount int) {
	t.Helper()
	var storedHash string
	var storedFlag bool
	ctx := t.Context()
	err := f.DB.Conn.QueryRow(ctx, `SELECT password_hash, must_change_password
 FROM public.users WHERE user_id=$1`, userID).Scan(&storedHash, &storedFlag)
	authdb.Require(t, "đối chiếu user sau transaction", err)
	if storedHash != hash || storedFlag != mustChange {
		t.Fatal("hash hoặc cờ buộc đổi trong DB khác kết quả mong đợi")
	}
	var count int
	err = f.DB.Conn.QueryRow(ctx, `SELECT count(*) FROM public.auth_sessions WHERE user_id=$1`, userID).Scan(&count)
	authdb.Require(t, "đếm phiên trực tiếp trong DB", err)
	if count != sessionCount {
		t.Fatal("số phiên của user thay đổi ngoài dự kiến")
	}
}

func TestReadPasswordChangeBasis(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{
		Email: "basis@example.invalid", MustChangePassword: true, Roles: []string{"DIRECTOR", "ADMIN"},
		EmploymentStatus: "INACTIVE",
	})
	token, _ := addPasswordChangeSession(t, f, a.UserID, "basis", sessionTestTime.Add(time.Hour), nil)
	repo := postgres.NewAuthRepository(f.DB.Conn, nil)
	basis, found, err := repo.ReadPasswordChangeBasis(t.Context(), token)
	authdb.Require(t, "đọc nền đổi mật khẩu", err)
	snapshot := sessionSnapshot(t, repo, token)
	principal := basis.Session.Principal
	if !found || basis.PasswordHash != f.PasswordHash || !reflect.DeepEqual(basis.Session, snapshot) ||
		principal.UserID != a.UserID || principal.EmployeeID != a.EmployeeID ||
		!principal.MustChangePassword || !basis.Session.UserActive || basis.Session.EmployeeActive ||
		!reflect.DeepEqual(principal.SystemRoles, []string{"ADMIN", "DIRECTOR"}) {
		t.Fatal("nền đổi mật khẩu không khớp đúng phiên, định danh và trạng thái")
	}

	missing := auth.SessionHash{}
	if got, found, err := repo.ReadPasswordChangeBasis(t.Context(), missing); found || err != nil ||
		!reflect.DeepEqual(got, auth.PasswordChangeBasis{}) {
		t.Fatal("token không tồn tại phải trả nền rỗng và found=false")
	}
	emptyDB := authdb.NewCase(t)
	brokenRepo := postgres.NewAuthRepository(emptyDB.Conn, nil)
	if got, found, err := brokenRepo.ReadPasswordChangeBasis(t.Context(), token); found ||
		!errors.Is(err, auth.ErrPersistence) || !reflect.DeepEqual(got, auth.PasswordChangeBasis{}) {
		t.Fatal("lỗi DB phải phân biệt với token không tồn tại và không trả nền")
	}
}

func TestChangeInitialPassword(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "change@example.invalid", MustChangePassword: true})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid", MustChangePassword: true})
	expiry := sessionTestTime.Add(time.Hour)
	earlier := sessionTestTime.Add(-time.Minute)
	current, _ := addPasswordChangeSession(t, f, a.UserID, "current", expiry, nil)
	second, _ := addPasswordChangeSession(t, f, a.UserID, "second", expiry, nil)
	expired, _ := addPasswordChangeSession(t, f, a.UserID, "expired", earlier, nil)
	alreadyRevoked, _ := addPasswordChangeSession(t, f, a.UserID, "already revoked", expiry, &earlier)
	otherToken, _ := addPasswordChangeSession(t, f, other.UserID, "other", expiry, nil)
	repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(func() time.Time { return sessionTestTime }))
	cmd := auth.ChangeInitialPasswordCommand{
		TokenHash: current, UserID: a.UserID, VerifiedPasswordHash: f.PasswordHash,
		NewPasswordHash: replacementHash(t, "Synthetic replacement password"),
	}
	authdb.Require(t, "commit đổi mật khẩu lần đầu", repo.ChangeInitialPassword(t.Context(), cmd))
	assertPasswordChangeState(t, f, a.UserID, cmd.NewPasswordHash, false, 4)
	assertPasswordChangeState(t, f, other.UserID, f.PasswordHash, true, 1)
	for _, token := range []auth.SessionHash{current, second, expired} {
		if !sameOptionalTime(sessionSnapshot(t, repo, token).RevokedAt, &sessionTestTime) {
			t.Fatal("phiên chưa thu hồi của user phải cùng thời điểm thu hồi")
		}
	}
	if !sameOptionalTime(sessionSnapshot(t, repo, alreadyRevoked).RevokedAt, &earlier) ||
		sessionSnapshot(t, repo, otherToken).RevokedAt != nil {
		t.Fatal("phiên đã thu hồi và phiên user khác phải được giữ nguyên")
	}
}

func TestChangeInitialPasswordRejectsChangedBasis(t *testing.T) {
	for _, scenario := range []string{"wrong user", "missing token", "revoked", "expired", "user locked",
		"employee inactive", "no employee", "flag cleared", "hash changed"} {
		t.Run(scenario, func(t *testing.T) {
			f := authdb.NewFixture(t)
			opts := authdb.AccountOptions{Email: "reject-change@example.invalid", MustChangePassword: scenario != "flag cleared"}
			switch scenario {
			case "user locked":
				opts.UserStatus = "LOCKED"
			case "employee inactive":
				opts.EmploymentStatus = "INACTIVE"
			case "no employee":
				opts.WithoutEmployee = true
			}
			a := f.AddAccount(t, opts)
			expiry := sessionTestTime.Add(time.Hour)
			var revoked *time.Time
			if scenario == "revoked" {
				revoked = &sessionTestTime
			}
			if scenario == "expired" {
				expiry = sessionTestTime
			}
			token, _ := addPasswordChangeSession(t, f, a.UserID, "reject-change", expiry, revoked)
			cmd := auth.ChangeInitialPasswordCommand{
				TokenHash: token, UserID: a.UserID, VerifiedPasswordHash: f.PasswordHash,
				NewPasswordHash: replacementHash(t, "Synthetic rejected password"),
			}
			switch scenario {
			case "wrong user":
				other := f.AddAccount(t, authdb.AccountOptions{Email: "wrong-user@example.invalid", MustChangePassword: true})
				cmd.UserID = other.UserID
			case "missing token":
				cmd.TokenHash = auth.SessionHash{}
			case "hash changed":
				cmd.VerifiedPasswordHash = cmd.NewPasswordHash
			}
			repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(func() time.Time { return sessionTestTime }))
			if err := repo.ChangeInitialPassword(t.Context(), cmd); !errors.Is(err, auth.ErrUnauthenticated) {
				t.Fatal("nền hoặc phiên không còn hợp lệ phải trả lỗi mất xác thực")
			}
			assertPasswordChangeState(t, f, a.UserID, f.PasswordHash, opts.MustChangePassword, 1)
			if cmd.UserID != a.UserID {
				assertPasswordChangeState(t, f, cmd.UserID, f.PasswordHash, true, 0)
			}
			var storedRevoked *time.Time
			err := f.DB.Conn.QueryRow(t.Context(), `SELECT revoked_at FROM public.auth_sessions WHERE token_hash=$1`, token[:]).Scan(&storedRevoked)
			authdb.Require(t, "kiểm phiên bị từ chối trong DB", err)
			if !sameOptionalTime(storedRevoked, revoked) {
				t.Fatal("request bị từ chối đã thay đổi phiên")
			}
		})
	}
}

func TestChangeInitialPasswordRollback(t *testing.T) {
	for _, failure := range []struct{ name, trigger string }{
		{"user write", `CREATE TRIGGER reject_change BEFORE UPDATE ON public.users
 FOR EACH ROW EXECUTE FUNCTION reject_password_change_fixture()`},
		{"session write", `CREATE TRIGGER reject_change BEFORE UPDATE ON public.auth_sessions
 FOR EACH ROW EXECUTE FUNCTION reject_password_change_fixture()`},
		{"commit", `CREATE CONSTRAINT TRIGGER reject_change AFTER UPDATE ON public.auth_sessions
 DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_password_change_fixture()`},
	} {
		t.Run(failure.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "change-rollback@example.invalid", MustChangePassword: true})
			token, _ := addPasswordChangeSession(t, f, a.UserID, "rollback-change", sessionTestTime.Add(time.Hour), nil)
			second, _ := addPasswordChangeSession(t, f, a.UserID, "second-rollback-change", sessionTestTime.Add(time.Hour), nil)
			_, err := f.DB.Conn.Exec(t.Context(), `CREATE FUNCTION reject_password_change_fixture() RETURNS trigger
 LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic password change failure'; END $$`)
			authdb.Require(t, "tạo hàm gây lỗi DB case", err)
			_, err = f.DB.Conn.Exec(t.Context(), failure.trigger)
			authdb.Require(t, "tạo trigger gây lỗi DB case", err)
			repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(func() time.Time { return sessionTestTime }))
			cmd := auth.ChangeInitialPasswordCommand{
				TokenHash: token, UserID: a.UserID, VerifiedPasswordHash: f.PasswordHash,
				NewPasswordHash: replacementHash(t, "Synthetic rollback password"),
			}
			if err := repo.ChangeInitialPassword(t.Context(), cmd); !errors.Is(err, auth.ErrPersistence) {
				t.Fatal("lỗi SQL phải trả lỗi lưu trữ an toàn")
			}
			assertPasswordChangeState(t, f, a.UserID, f.PasswordHash, true, 2)
			for _, hash := range []auth.SessionHash{token, second} {
				if sessionSnapshot(t, repo, hash).RevokedAt != nil {
					t.Fatal("lỗi SQL phải rollback thu hồi phiên")
				}
			}
		})
	}
}

func TestChangeInitialPasswordUnknownCommitOutcome(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "change-commit@example.invalid", MustChangePassword: true})
	token, _ := addPasswordChangeSession(t, f, a.UserID, "change-commit", sessionTestTime.Add(time.Hour), nil)
	clock := clockFunc(func() time.Time { return sessionTestTime })
	repo := postgres.NewAuthRepository(lostCommitReplyDB{f.DB.Conn}, clock)
	cmd := auth.ChangeInitialPasswordCommand{
		TokenHash: token, UserID: a.UserID, VerifiedPasswordHash: f.PasswordHash,
		NewPasswordHash: replacementHash(t, "Synthetic committed password"),
	}
	if err := repo.ChangeInitialPassword(t.Context(), cmd); !errors.Is(err, auth.ErrUnavailable) {
		t.Fatal("mất phản hồi commit phải trả lỗi kết nối")
	}
	assertPasswordChangeState(t, f, a.UserID, cmd.NewPasswordHash, false, 1)
	reader := postgres.NewAuthRepository(f.DB.Conn, clock)
	if !sameOptionalTime(sessionSnapshot(t, reader, token).RevokedAt, &sessionTestTime) {
		t.Fatal("commit đã xảy ra phải thu hồi phiên cùng với đổi mật khẩu")
	}
}
