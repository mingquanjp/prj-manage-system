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

func TestFindLoginIdentity(t *testing.T) {
	f := authdb.NewFixture(t)
	repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(time.Now))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const email = "  Fixture.Login@EXAMPLE.invalid  "
	active := f.AddAccount(t, authdb.AccountOptions{Email: email})
	locked := f.AddAccount(t, authdb.AccountOptions{Email: "locked@example.invalid", UserStatus: "LOCKED"})
	disabled := f.AddAccount(t, authdb.AccountOptions{Email: "disabled@example.invalid", UserStatus: "DISABLED"})
	inactive := f.AddAccount(t, authdb.AccountOptions{Email: "inactive@example.invalid", EmploymentStatus: "INACTIVE"})
	resigned := f.AddAccount(t, authdb.AccountOptions{Email: "resigned@example.invalid", EmploymentStatus: "RESIGNED"})

	for _, tc := range []struct {
		name, input, stored        string
		account                    authdb.Account
		userActive, employeeActive bool
	}{
		{"chuẩn hóa", "\tFIXTURE.Login@example.INVALID\n", email, active, true, true},
		{"tài khoản khóa", "locked@example.invalid", "locked@example.invalid", locked, false, true},
		{"tài khoản vô hiệu", "disabled@example.invalid", "disabled@example.invalid", disabled, false, true},
		{"hồ sơ ngừng hoạt động", "inactive@example.invalid", "inactive@example.invalid", inactive, true, false},
		{"hồ sơ đã nghỉ", "resigned@example.invalid", "resigned@example.invalid", resigned, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, found, err := repo.FindLoginIdentity(ctx, tc.input)
			authdb.Require(t, "đọc danh tính", err)
			want := auth.LoginIdentity{
				UserID: tc.account.UserID, EmployeeID: tc.account.EmployeeID,
				Email: tc.stored, PasswordHash: f.PasswordHash,
				UserActive: tc.userActive, EmployeeActive: tc.employeeActive,
			}
			if !found || got != want || got.UserID == got.EmployeeID {
				t.Fatal("projection đăng nhập sai ID, email, hash hoặc trạng thái")
			}
		})
	}

	for _, input := range []string{"missing@example.invalid", "' OR true --"} {
		got, found, err := repo.FindLoginIdentity(ctx, input)
		if found || err != nil || got != (auth.LoginIdentity{}) {
			t.Fatal("email không khớp phải trả found=false, không trả dữ liệu")
		}
	}

	f.AddAccount(t, authdb.AccountOptions{Email: "fixture.login@example.invalid", UserStatus: "LOCKED"})
	got, found, err := repo.FindLoginIdentity(ctx, "fixture.login@example.invalid")
	if found || !errors.Is(err, auth.ErrIdentityConflict) || got != (auth.LoginIdentity{}) {
		t.Fatal("phải từ chối email trùng kể cả khi chỉ một tài khoản hoạt động")
	}
}

func TestReadSession(t *testing.T) {
	f := authdb.NewFixture(t)
	repo := postgres.NewAuthRepository(f.DB.Conn, clockFunc(time.Now))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	a := f.AddAccount(t, authdb.AccountOptions{
		Email: "restricted@example.invalid", LegacyRole: "PM",
		Roles: []string{"HR", "ADMIN", "DIRECTOR"}, MustChangePassword: true,
	})
	b := f.AddAccount(t, authdb.AccountOptions{
		Email: "ordinary@example.invalid", LegacyRole: "ADMIN", EmploymentStatus: "RESIGNED",
	})
	withoutEmployee := f.AddAccount(t, authdb.AccountOptions{WithoutEmployee: true})
	created := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	revoked := created.Add(time.Hour)
	expires := created.Add(8 * time.Hour)
	earlierExpiry := created.Add(2 * time.Hour)
	activeHash := sha256.Sum256([]byte("authdb-first-session"))
	revokedHash := sha256.Sum256([]byte("authdb-second-session"))
	otherHash := sha256.Sum256([]byte("authdb-other-user-session"))
	unlinkedHash := sha256.Sum256([]byte("authdb-no-employee-session"))
	f.AddSession(t, a.UserID, activeHash, created, expires, nil)
	f.AddSession(t, a.UserID, revokedHash, created, earlierExpiry, &revoked)
	f.AddSession(t, b.UserID, otherHash, created, expires, nil)
	f.AddSession(t, withoutEmployee.UserID, unlinkedHash, created, expires, nil)

	read := func(hash auth.SessionHash) auth.SessionSnapshot {
		t.Helper()
		got, found, err := repo.ReadSession(ctx, hash)
		authdb.Require(t, "đọc phiên", err)
		if !found {
			t.Fatal("không tìm thấy projection phiên đã tạo")
		}
		return got
	}
	first := read(activeHash)
	p := first.Principal
	if p.UserID != a.UserID || p.EmployeeID != a.EmployeeID || p.EmployeeCode != a.EmployeeCode ||
		p.Fullname != a.Fullname || !p.MustChangePassword || !first.UserActive || !first.EmployeeActive ||
		!slices.Equal(p.SystemRoles, []string{"ADMIN", "DIRECTOR", "HR"}) ||
		!first.CreatedAt.Equal(created) || !first.ExpiresAt.Equal(expires) || first.RevokedAt != nil {
		t.Fatal("projection phiên hạn chế sai danh tính, quyền thực, trạng thái hoặc thời điểm")
	}

	second := read(revokedHash)
	if second.Principal.UserID != a.UserID || !second.ExpiresAt.Equal(earlierExpiry) ||
		second.RevokedAt == nil || !second.RevokedAt.Equal(revoked) {
		t.Fatal("phải đọc đúng phiên đã thu hồi theo hash, gồm thời điểm hết hạn riêng")
	}

	other := read(otherHash)
	if other.Principal.UserID != b.UserID || other.Principal.EmployeeID != b.EmployeeID ||
		other.Principal.MustChangePassword || other.Principal.SystemRoles == nil ||
		len(other.Principal.SystemRoles) != 0 || !other.UserActive || other.EmployeeActive {
		t.Fatal("role rỗng phải là tập rỗng; legacy ADMIN không tự cấp quyền")
	}

	for _, hash := range []auth.SessionHash{{}, unlinkedHash} {
		got, found, err := repo.ReadSession(ctx, hash)
		if found || err != nil || !reflect.DeepEqual(got, auth.SessionSnapshot{}) {
			t.Fatal("hash không có hoặc thiếu employee phải trả found=false và projection rỗng")
		}
	}

	_, err := f.DB.Conn.Exec(ctx, `DELETE FROM public.user_roles WHERE user_id=$1 AND role_code='ADMIN'`, a.UserID)
	authdb.Require(t, "đổi role fixture", err)
	_, err = f.DB.Conn.Exec(ctx, `UPDATE public.users SET status='LOCKED', must_change_password=false WHERE user_id=$1`, a.UserID)
	authdb.Require(t, "đổi trạng thái user fixture", err)
	_, err = f.DB.Conn.Exec(ctx, `UPDATE public.employees SET employment_status='INACTIVE' WHERE employee_id=$1`, a.EmployeeID)
	authdb.Require(t, "đổi trạng thái employee fixture", err)
	now := read(activeHash)
	if now.UserActive || now.EmployeeActive || now.Principal.MustChangePassword ||
		!slices.Equal(now.Principal.SystemRoles, []string{"DIRECTOR", "HR"}) ||
		!now.ExpiresAt.Equal(expires) || now.RevokedAt != nil {
		t.Fatal("lần đọc sau phải thấy dữ liệu mới và giữ nguyên thông tin phiên")
	}
}

func TestAuthReadDatabaseFailure(t *testing.T) {
	// DB case chưa migrate tạo lỗi SQL thật mà không sửa schema dùng cho ca khác.
	db := authdb.NewCase(t)
	repo := postgres.NewAuthRepository(db.Conn, clockFunc(time.Now))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	login, found, err := repo.FindLoginIdentity(ctx, "fixture@example.invalid")
	if found || err != auth.ErrPersistence || login != (auth.LoginIdentity{}) {
		t.Fatal("lỗi SQL khi đọc login phải được phân loại, không trả projection hoặc lỗi driver")
	}
	session, found, err := repo.ReadSession(ctx, auth.SessionHash{})
	if found || err != auth.ErrPersistence || !reflect.DeepEqual(session, auth.SessionSnapshot{}) {
		t.Fatal("lỗi SQL khi đọc phiên phải được phân loại, không trả projection hoặc lỗi driver")
	}
}

func TestAuthRepositoryRootRemainsEmpty(t *testing.T) {
	authdb.AssertRootEmpty(t)
}
