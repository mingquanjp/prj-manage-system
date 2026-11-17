//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"math"
	"testing"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/testutil/authdb"
)

func TestEmployeeCodeAllocatorPreviewAndRollback(t *testing.T) {
	f := authdb.NewFixture(t)
	ctx := t.Context()
	for range 2 {
		code, err := postgres.PreviewEmployeeCode(ctx, f.DB.Conn)
		authdb.Require(t, "preview không cấp số", err)
		if code != "EMP000001" {
			t.Fatal("preview phải ổn định khi chưa cấp số")
		}
	}
	tx, err := f.DB.Conn.Begin(ctx)
	authdb.Require(t, "begin cấp mã", err)
	defer tx.Rollback(context.Background())
	code, err := postgres.AllocateEmployeeCode(ctx, tx)
	authdb.Require(t, "cấp mã trong transaction", err)
	if code != "EMP000001" {
		t.Fatal("mã đầu sai")
	}
	authdb.Require(t, "rollback cấp mã", tx.Rollback(ctx))
	code, err = postgres.PreviewEmployeeCode(ctx, f.DB.Conn)
	authdb.Require(t, "preview sau rollback", err)
	if code != "EMP000002" {
		t.Fatal("preview không được tái dùng số rollback")
	}
	// DML đặc quyền ngoài allocator có thể gây collision: UNIQUE là chốt cuối.
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "code@example.invalid"})
	_, err = f.DB.Conn.Exec(ctx, `UPDATE employees SET employee_code='EMP000002' WHERE employee_id=$1`, actor.EmployeeID)
	authdb.Require(t, "fixture collision ngoài app", err)
	tx, err = f.DB.Conn.Begin(ctx)
	authdb.Require(t, "begin collision", err)
	code, err = postgres.AllocateEmployeeCode(ctx, tx)
	authdb.Require(t, "cấp số collision", err)
	if code != "EMP000002" {
		t.Fatal("allocator không được tự skip mã bị chèn ngoài app")
	}
	// Chèn một employee mới cùng mã để thực thi UNIQUE thay vì so chuỗi giả.
	var user int
	authdb.Require(t, "user test collision", tx.QueryRow(ctx, `INSERT INTO users(password_hash,system_role) VALUES('synthetic','EMPLOYEE') RETURNING user_id`).Scan(&user))
	_, err = tx.Exec(ctx, `INSERT INTO employees(user_id,employee_code,fullname,email,position_id)
 SELECT $1,$2,'Synthetic code collision','collision@example.invalid',position_id FROM positions LIMIT 1`, user, code)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatal("UNIQUE phải chặn collision")
	}
	authdb.Require(t, "rollback collision", tx.Rollback(ctx))
	code, err = postgres.PreviewEmployeeCode(ctx, f.DB.Conn)
	authdb.Require(t, "preview sau collision", err)
	if code != "EMP000003" {
		t.Fatal("collision cũng mất số, không retry ngầm")
	}
}

func TestEmployeeCodeExhaustionAndLargeFormat(t *testing.T) {
	f := authdb.NewFixture(t)
	ctx := t.Context()
	// Chỉ sửa sequence trong case hy sinh để kiểm biên; startup/seed không làm việc này.
	_, err := f.DB.Conn.Exec(ctx, `SELECT setval('public.employee_code_seq',1000000,false)`)
	authdb.Require(t, "fixture số dài", err)
	code, err := postgres.PreviewEmployeeCode(ctx, f.DB.Conn)
	authdb.Require(t, "preview số dài", err)
	if code != "EMP1000000" {
		t.Fatal("không được cắt số dài")
	}
	_, err = f.DB.Conn.Exec(ctx, `SELECT setval('public.employee_code_seq',$1,false)`, int64(math.MaxInt64))
	authdb.Require(t, "fixture cuối dải", err)
	code, err = postgres.PreviewEmployeeCode(ctx, f.DB.Conn)
	authdb.Require(t, "preview số cuối", err)
	if code != "EMP9223372036854775807" {
		t.Fatal("preview cuối dải sai")
	}
	tx, err := f.DB.Conn.Begin(ctx)
	authdb.Require(t, "begin số cuối", err)
	_, err = postgres.AllocateEmployeeCode(ctx, tx)
	authdb.Require(t, "cấp số cuối", err)
	authdb.Require(t, "rollback vẫn giữ cuối dải", tx.Rollback(ctx))
	if _, err = postgres.PreviewEmployeeCode(ctx, f.DB.Conn); err == nil {
		t.Fatal("preview hết dải phải lỗi")
	}
	tx, err = f.DB.Conn.Begin(ctx)
	authdb.Require(t, "begin sau hết dải", err)
	defer tx.Rollback(context.Background())
	if _, err = postgres.AllocateEmployeeCode(ctx, tx); err == nil {
		t.Fatal("allocator không được quay vòng")
	}
}
