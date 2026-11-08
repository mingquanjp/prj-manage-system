//go:build integration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
)

type clockFunc func() time.Time

func (c clockFunc) Now() time.Time { return c() }

var sessionTestTime = time.Date(2026, 11, 9, 8, 0, 0, 123456000, time.UTC)

func loginCommand(t *testing.T, repo *postgres.AuthRepository, email, token string) auth.FinalizeLoginCommand {
	t.Helper()
	identity, found, err := repo.FindLoginIdentity(context.Background(), email)
	authdb.Require(t, "đọc nền login fixture", err)
	if !found {
		t.Fatal("thiếu nền login fixture")
	}
	return auth.FinalizeLoginCommand{VerifiedIdentity: identity, NewTokenHash: sha256.Sum256([]byte(token)), TTL: 8 * time.Hour}
}

func sessionSnapshot(t *testing.T, repo *postgres.AuthRepository, hash auth.SessionHash) auth.SessionSnapshot {
	t.Helper()
	s, found, err := repo.ReadSession(context.Background(), hash)
	authdb.Require(t, "đọc lại phiên", err)
	if !found {
		t.Fatal("thiếu phiên đã tạo")
	}
	return s
}

func assertNoSession(t *testing.T, db *authdb.Case, repo *postgres.AuthRepository, hash auth.SessionHash) {
	t.Helper()
	var count int
	authdb.Require(t, "kiểm trực tiếp phiên vắng mặt", db.Conn.QueryRow(context.Background(),
		`SELECT count(*) FROM public.auth_sessions WHERE token_hash=$1`, hash[:]).Scan(&count))
	if count != 0 {
		t.Fatal("còn phiên trong DB dù projection có thể không thấy employee")
	}
	s, found, err := repo.ReadSession(context.Background(), hash)
	if found || err != nil || !reflect.DeepEqual(s, auth.SessionSnapshot{}) {
		t.Fatal("không được để lại phiên mới hoặc projection lỗi")
	}
}

func TestFinalizeLogin(t *testing.T) {
	for _, previous := range []string{"không có", "cùng user", "user khác", "hết hạn", "đã thu hồi", "hash không có"} {
		t.Run(previous, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "login@example.invalid", Roles: []string{"HR"}})
			b := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid"})
			repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(func() time.Time { return sessionTestTime }))
			cmd := loginCommand(t, repo, "login@example.invalid", "new session")
			old := auth.SessionHash(sha256.Sum256([]byte("old session")))
			independent := sha256.Sum256([]byte("another device"))
			created := sessionTestTime.Add(-time.Hour)
			expires := sessionTestTime.Add(time.Hour)
			owner := a.UserID
			var revoked *time.Time
			switch previous {
			case "user khác":
				owner = b.UserID
			case "hết hạn":
				expires = sessionTestTime
			case "đã thu hồi":
				value := sessionTestTime.Add(-time.Minute)
				revoked = &value
			}
			if previous != "không có" {
				cmd.PreviousTokenHash = &old
			}
			if previous != "hash không có" {
				f.AddSession(t, owner, old, created, expires, revoked)
			}
			f.AddSession(t, a.UserID, independent, created, sessionTestTime.Add(time.Hour), nil)
			// Cờ và role phải lấy hiện hành, không lấy từ thời điểm lookup.
			_, err := f.DB.Conn.Exec(context.Background(), `UPDATE public.users SET must_change_password=true WHERE user_id=$1`, a.UserID)
			authdb.Require(t, "đổi cờ fixture", err)
			_, err = f.DB.Conn.Exec(context.Background(), `INSERT INTO public.user_roles VALUES ($1,'DIRECTOR')`, a.UserID)
			authdb.Require(t, "bổ sung role fixture", err)

			got, err := repo.FinalizeLogin(context.Background(), cmd)
			authdb.Require(t, "hoàn tất login", err)
			stored := sessionSnapshot(t, repo, cmd.NewTokenHash)
			if !reflect.DeepEqual(got, stored) || got.Principal.EmployeeID != a.EmployeeID ||
				!got.Principal.MustChangePassword || !slices.Equal(got.Principal.SystemRoles, []string{"DIRECTOR", "HR"}) ||
				!got.CreatedAt.Equal(sessionTestTime) || !got.ExpiresAt.Equal(sessionTestTime.Add(8*time.Hour)) {
				t.Fatal("snapshot cấp phiên phải trùng dữ liệu commit và thời điểm hiện hành")
			}
			if previous != "hash không có" {
				wantRevoked := revoked
				if previous == "cùng user" {
					wantRevoked = &sessionTestTime
				}
				before := sessionSnapshot(t, repo, old)
				if !sameOptionalTime(before.RevokedAt, wantRevoked) {
					t.Fatal("thu hồi sai cookie cũ")
				}
			}
			if sessionSnapshot(t, repo, independent).RevokedAt != nil {
				t.Fatal("login không được thu hồi thiết bị khác")
			}
		})
	}
}

func TestFinalizeLoginRejectsChangedBasis(t *testing.T) {
	for _, change := range []string{"hash", "email", "user khóa", "employee vô hiệu", "employee khác", "employee thiếu", "user không có", "TTL sai"} {
		t.Run(change, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "basis@example.invalid"})
			repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(func() time.Time { return sessionTestTime }))
			cmd := loginCommand(t, repo, "basis@example.invalid", "rejected session")
			var err error
			switch change {
			case "hash":
				_, err = f.DB.Conn.Exec(context.Background(), `UPDATE public.users SET password_hash='changed fixture' WHERE user_id=$1`, a.UserID)
			case "email":
				_, err = f.DB.Conn.Exec(context.Background(), `UPDATE public.employees SET email='changed@example.invalid' WHERE user_id=$1`, a.UserID)
			case "user khóa":
				_, err = f.DB.Conn.Exec(context.Background(), `UPDATE public.users SET status='LOCKED' WHERE user_id=$1`, a.UserID)
			case "employee vô hiệu":
				_, err = f.DB.Conn.Exec(context.Background(), `UPDATE public.employees SET employment_status='INACTIVE' WHERE user_id=$1`, a.UserID)
			case "employee khác":
				cmd.VerifiedIdentity.EmployeeID++
			case "employee thiếu":
				_, err = f.DB.Conn.Exec(context.Background(), `DELETE FROM public.employees WHERE user_id=$1`, a.UserID)
			case "user không có":
				cmd.VerifiedIdentity.UserID++
			case "TTL sai":
				cmd.TTL = time.Hour
			}
			authdb.Require(t, "đổi nền fixture", err)
			got, err := repo.FinalizeLogin(context.Background(), cmd)
			want := auth.ErrInvalidCredentials
			if change == "TTL sai" {
				want = auth.ErrPersistence
			}
			if !errors.Is(err, want) || !reflect.DeepEqual(got, auth.SessionSnapshot{}) {
				t.Fatal("nền không hợp lệ phải trả lỗi và snapshot rỗng")
			}
			assertNoSession(t, f.DB, repo, cmd.NewTokenHash)
		})
	}
}

func TestRevokeCurrentSession(t *testing.T) {
	for _, state := range []string{"hợp lệ", "hết hạn", "đã thu hồi", "user khóa", "employee vô hiệu", "hash không có"} {
		t.Run(state, func(t *testing.T) {
			f := authdb.NewFixture(t)
			opts := authdb.AccountOptions{Email: "logout@example.invalid", MustChangePassword: true}
			if state == "user khóa" {
				opts.UserStatus = "LOCKED"
			}
			if state == "employee vô hiệu" {
				opts.EmploymentStatus = "INACTIVE"
			}
			a := f.AddAccount(t, opts)
			repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(func() time.Time { return sessionTestTime }))
			hash := sha256.Sum256([]byte("logout session"))
			other := sha256.Sum256([]byte("other logout session"))
			expires := sessionTestTime.Add(time.Hour)
			var revoked *time.Time
			if state == "hết hạn" {
				expires = sessionTestTime
			}
			if state == "đã thu hồi" {
				x := sessionTestTime.Add(-time.Minute)
				revoked = &x
			}
			if state != "hash không có" {
				f.AddSession(t, a.UserID, hash, sessionTestTime.Add(-time.Hour), expires, revoked)
			}
			f.AddSession(t, a.UserID, other, sessionTestTime.Add(-time.Hour), sessionTestTime.Add(time.Hour), nil)
			err := repo.RevokeCurrentSession(context.Background(), hash)
			if state == "hợp lệ" {
				authdb.Require(t, "logout", err)
				revoked = &sessionTestTime
				if err := repo.RevokeCurrentSession(context.Background(), hash); !errors.Is(err, auth.ErrUnauthenticated) {
					t.Fatal("không chấp thuận lại phiên đã thu hồi")
				}
			} else if !errors.Is(err, auth.ErrUnauthenticated) {
				t.Fatal("phiên mất hiệu lực phải bị từ chối")
			}
			if state != "hash không có" && !sameOptionalTime(sessionSnapshot(t, repo, hash).RevokedAt, revoked) {
				t.Fatal("thời điểm thu hồi sai hoặc thay đổi phiên không hợp lệ")
			}
			if sessionSnapshot(t, repo, other).RevokedAt != nil {
				t.Fatal("logout ảnh hưởng phiên khác")
			}
		})
	}
}

func TestAuthSessionRollback(t *testing.T) {
	for _, failure := range []string{"insert trùng hash", "commit login", "write logout", "commit logout"} {
		t.Run(failure, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "rollback@example.invalid"})
			repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(func() time.Time { return sessionTestTime }))
			cmd := loginCommand(t, repo, "rollback@example.invalid", "rollback new")
			old := auth.SessionHash(sha256.Sum256([]byte("rollback old")))
			cmd.PreviousTokenHash = &old
			f.AddSession(t, a.UserID, old, sessionTestTime.Add(-time.Hour), sessionTestTime.Add(time.Hour), nil)
			if failure == "insert trùng hash" {
				// UPDATE phiên cũ thành công trước khi INSERT mới vi phạm unique.
				cmd.NewTokenHash = old
			} else {
				_, err := f.DB.Conn.Exec(context.Background(), `CREATE FUNCTION reject_auth_fixture() RETURNS trigger
 LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic auth failure'; END $$`)
				authdb.Require(t, "hàm gây lỗi DB case", err)
				statement := `CREATE CONSTRAINT TRIGGER reject_auth_commit AFTER INSERT OR UPDATE ON public.auth_sessions
 DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_auth_fixture()`
				if failure == "write logout" {
					statement = `CREATE TRIGGER reject_auth_write BEFORE UPDATE ON public.auth_sessions
 FOR EACH ROW EXECUTE FUNCTION reject_auth_fixture()`
				}
				_, err = f.DB.Conn.Exec(context.Background(), statement)
				authdb.Require(t, "cài lỗi DB case", err)
			}
			var err error
			if failure == "write logout" || failure == "commit logout" {
				err = repo.RevokeCurrentSession(context.Background(), old)
			} else {
				var got auth.SessionSnapshot
				got, err = repo.FinalizeLogin(context.Background(), cmd)
				if !reflect.DeepEqual(got, auth.SessionSnapshot{}) {
					t.Fatal("lỗi ghi/commit không được trả snapshot")
				}
			}
			if !errors.Is(err, auth.ErrPersistence) {
				t.Fatal("lỗi DB thật phải được phân loại an toàn")
			}
			if sessionSnapshot(t, repo, old).RevokedAt != nil {
				t.Fatal("ghi thu hồi phải rollback")
			}
			if cmd.NewTokenHash != old {
				assertNoSession(t, f.DB, repo, cmd.NewTokenHash)
			}
		})
	}
}

// PostgreSQL có thể trả cùng thời điểm với biểu diễn múi giờ khác clock test.
func sameOptionalTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}
