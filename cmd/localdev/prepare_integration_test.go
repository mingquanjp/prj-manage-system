//go:build integration

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"ai-staffing/internal/testutil/authdb"
	"ai-staffing/migrations"
	"golang.org/x/crypto/bcrypt"
)

func TestPrepareHeadLocaldev(t *testing.T) {
	// Case do guard tạo và sở hữu, không mở/reset database auth_dev đang chạy.
	db := authdb.NewCase(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	cfg := db.Conn.Config()
	if err := prepare(ctx, cfg, "wrong cluster identifier"); err == nil {
		t.Fatal("prepare phải chặn cluster không khớp trước DDL")
	}
	var tables int
	authdb.Require(t, "kiểm chưa DDL sau guard từ chối", db.Conn.QueryRow(ctx, `SELECT count(*) FROM pg_tables WHERE schemaname='public'`).Scan(&tables))
	if tables != 0 {
		t.Fatal("guard từ chối đã tạo bảng")
	}
	authdb.Require(t, "prepare localdev bằng migration và seed thật", prepare(ctx, cfg, os.Getenv("AUTH_TEST_DB_SYSTEM_ID")))
	var version int
	var dirty bool
	authdb.Require(t, "head localdev", db.Conn.QueryRow(ctx, `SELECT version,dirty FROM schema_migrations`).Scan(&version, &dirty))
	if version != migrations.PCAVersion || dirty {
		t.Fatal("localdev chưa ở head PCA sạch")
	}
	rows, err := db.Conn.Query(ctx, `SELECT u.password_hash,u.system_role,u.must_change_password,e.email,r.role_code
 FROM users u JOIN employees e USING(user_id) JOIN user_roles r USING(user_id) ORDER BY u.user_id`)
	authdb.Require(t, "đọc fixture localdev", err)
	defer rows.Close()
	var count int
	for rows.Next() {
		var hash, legacy, email, role string
		var limited bool
		authdb.Require(t, "scan fixture localdev", rows.Scan(&hash, &legacy, &limited, &email, &role))
		if legacy != "EMPLOYEE" || bcrypt.CompareHashAndPassword([]byte(hash), []byte("12345678")) != nil {
			t.Fatal("credential hoặc legacy role localdev không hợp lệ")
		}
		switch email {
		case "quan1652005@gmail.com":
			if limited || role != "HR" {
				t.Fatal("fixture HR sai")
			}
		case "quan1652005+limited@gmail.com":
			if !limited || role != "ADMIN" {
				t.Fatal("fixture limited sai")
			}
		case "quan1652005+admin@gmail.com":
			if limited || role != "ADMIN" {
				t.Fatal("fixture Admin sai")
			}
		default:
			t.Fatal("localdev có fixture ngoài tập tổng hợp")
		}
		count++
	}
	authdb.Require(t, "đọc hết fixture localdev", rows.Err())
	if count != 3 {
		t.Fatal("localdev không có đúng ba tài khoản")
	}
	rows.Close()
	var nullable bool
	authdb.Require(t, "địa chỉ và cấp bậc localdev giữ NULL", db.Conn.QueryRow(ctx, `SELECT bool_and(address IS NULL AND job_level_id IS NULL) FROM employees`).Scan(&nullable))
	if !nullable {
		t.Fatal("localdev tự điền địa chỉ hoặc cấp bậc")
	}

	var usernameAbsent bool
	authdb.Require(t, "không username ở localdev", db.Conn.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='username')`).Scan(&usernameAbsent))
	if !usernameAbsent {
		t.Fatal("schema đích còn username")
	}
}
