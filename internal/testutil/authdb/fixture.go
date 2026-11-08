//go:build integration

package authdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
	"ai-staffing/migrations"
	"golang.org/x/crypto/bcrypt"
)

const FixturePassword = "Only synthetic auth repository fixture"

type Fixture struct {
	DB           *Case
	PasswordHash string
	positionID   int
	nextAccount  int
}

type Account struct {
	UserID       int
	EmployeeID   int
	EmployeeCode string
	Fullname     string
}

type AccountOptions struct {
	Email              string
	LegacyRole         string
	UserStatus         string
	EmploymentStatus   string
	Roles              []string
	MustChangePassword bool
	WithoutEmployee    bool
}

func NewFixture(t *testing.T) *Fixture {
	t.Helper()
	db := NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	Require(t, "nâng fixture đến migration auth", m.Migrate(3))
	m.Close(t)

	hash, err := bcrypt.GenerateFromPassword([]byte(FixturePassword), bcrypt.DefaultCost)
	Require(t, "tạo hash fixture", err)
	f := &Fixture{DB: db, PasswordHash: string(hash)}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.Conn.QueryRow(ctx, `INSERT INTO public.positions(position_name)
 VALUES ('Fixture position') RETURNING position_id`).Scan(&f.positionID)
	Require(t, "tạo vị trí fixture", err)
	return f
}

func (f *Fixture) AddAccount(t *testing.T, opts AccountOptions) Account {
	t.Helper()
	if opts.LegacyRole == "" {
		opts.LegacyRole = "EMPLOYEE"
	}
	if opts.UserStatus == "" {
		opts.UserStatus = "ACTIVE"
	}
	if opts.EmploymentStatus == "" {
		opts.EmploymentStatus = "ACTIVE"
	}

	f.nextAccount++
	n := f.nextAccount
	account := Account{UserID: 100 + n}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := f.DB.Conn.Begin(ctx)
	Require(t, "bắt đầu giao dịch fixture", err)
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO public.users
 (user_id, username, password_hash, system_role, status, must_change_password)
 OVERRIDING SYSTEM VALUE VALUES ($1,$2,$3,$4,$5,$6)`,
		account.UserID, fmt.Sprintf("fixture-%d", n), f.PasswordHash,
		opts.LegacyRole, opts.UserStatus, opts.MustChangePassword)
	Require(t, "tạo user fixture", err)

	if !opts.WithoutEmployee {
		account.EmployeeID = 1000 + n
		account.EmployeeCode = fmt.Sprintf("FIX-%d", n)
		account.Fullname = fmt.Sprintf("Fixture %d", n)
		_, err = tx.Exec(ctx, `INSERT INTO public.employees
 (employee_id, user_id, employee_code, fullname, email, position_id, employment_status)
 OVERRIDING SYSTEM VALUE VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			account.EmployeeID, account.UserID, account.EmployeeCode, account.Fullname,
			opts.Email, f.positionID, opts.EmploymentStatus)
		Require(t, "tạo employee fixture", err)
	}

	for _, role := range opts.Roles {
		_, err = tx.Exec(ctx, `INSERT INTO public.user_roles(user_id,role_code) VALUES ($1,$2)`, account.UserID, role)
		Require(t, "tạo role fixture", err)
	}
	Require(t, "commit tài khoản fixture", tx.Commit(ctx))
	return account
}

func (f *Fixture) AddSession(t *testing.T, userID int, hash auth.SessionHash, created, expires time.Time, revoked *time.Time) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var id int
	err := f.DB.Conn.QueryRow(ctx, `INSERT INTO public.auth_sessions
 (user_id, token_hash, created_at, expires_at, revoked_at)
 VALUES ($1,$2,$3,$4,$5) RETURNING session_id`, userID, hash[:], created, expires, revoked).Scan(&id)
	Require(t, "tạo phiên fixture", err)
	return id
}
