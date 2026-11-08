//go:build integration

// Package authdb dựng database case trong cluster PostgreSQL test đã xác minh.
package authdb

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	driver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
)

const baseDB = "ai_staffing_auth_test"
const dbUser = "ai_staffing_test"

// Không in err: pgx/migrate có thể kèm SQL, hash hoặc thông tin kết nối.
func Require(t *testing.T, stage string, err error) {
	t.Helper()
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			t.Fatalf("%s: SQLSTATE %s (không xuất lỗi thô)", stage, pgErr.Code)
		}
		t.Fatalf("%s: không thành công (không xuất lỗi thô)", stage)
	}
}

func rootConfig(t *testing.T) *pgx.ConnConfig {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	u, err := url.Parse(raw)
	Require(t, "URL test", err)
	if raw == "" || u.Scheme != "postgresql" || u.Hostname() != "127.0.0.1" || u.Port() != "55438" ||
		u.User == nil || u.User.Username() != dbUser || u.Path != "/"+baseDB || u.RawQuery != "sslmode=disable" || u.Fragment != "" {
		t.Fatal("thiếu hoặc sai TEST_DATABASE_URL; không fallback DB khác")
	}
	cfg, err := pgx.ParseConfig(raw)
	Require(t, "parse config test", err)
	cfg.ConnectTimeout = 5 * time.Second
	return cfg
}

func verifyCluster(t *testing.T, conn *pgx.Conn, name string) {
	t.Helper()
	expected := os.Getenv("AUTH_TEST_DB_SYSTEM_ID")
	if expected == "" {
		t.Fatal("thiếu cluster ID được runner xác minh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var major int
	var database, user, system string
	var owner bool
	err := conn.QueryRow(ctx, `SELECT current_setting('server_version_num')::int/10000,
 current_database(),current_user,pg_get_userbyid(datdba)=current_user,
 (SELECT system_identifier::text FROM pg_control_system())
 FROM pg_database WHERE datname=current_database()`).Scan(&major, &database, &user, &owner, &system)
	Require(t, "metadata cluster", err)
	if major != 18 || database != name || user != dbUser || !owner || system != expected {
		t.Fatal("đích/owner/cluster test không khớp; từ chối DDL/dọn")
	}
}

type Case struct {
	cfg  *pgx.ConnConfig
	Conn *pgx.Conn
}

func NewCase(t *testing.T) *Case {
	t.Helper()
	cfg := rootConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	root, err := pgx.ConnectConfig(ctx, cfg)
	Require(t, "kết nối DB test gốc", err)
	defer root.Close(context.Background())
	verifyCluster(t, root, baseDB)
	var random [8]byte
	_, err = rand.Read(random[:])
	Require(t, "sinh tên DB test", err)
	name := "auth_case_" + hex.EncodeToString(random[:])
	_, err = root.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()+" TEMPLATE template0")
	Require(t, "tạo DB case test", err)
	// Chỉ đăng ký tên vừa CREATE thành công; không nhận DB có sẵn để dọn.
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		admin, err := pgx.ConnectConfig(cleanup, cfg)
		if err != nil {
			t.Error("không nối được để dọn DB case")
			return
		}
		defer admin.Close(cleanup)
		verifyCluster(t, admin, baseDB)
		var owned bool
		err = admin.QueryRow(cleanup, `SELECT pg_get_userbyid(datdba)=current_user FROM pg_database WHERE datname=$1`, name).Scan(&owned)
		if err != nil || !owned {
			t.Error("DB case không đúng owner; không xóa")
			return
		}
		_, err = admin.Exec(cleanup, "DROP DATABASE "+pgx.Identifier{name}.Sanitize())
		if err != nil {
			t.Error("không dọn được DB case; giữ để kiểm tra, không FORCE/prune")
		}
	})
	child := cfg.Copy()
	child.Database = name
	conn, err := pgx.ConnectConfig(ctx, child)
	Require(t, "kết nối DB case", err)
	t.Cleanup(func() { conn.Close(context.Background()) })
	verifyCluster(t, conn, name)
	return &Case{cfg: child, Conn: conn}
}

type migrationEngine = migrate.Migrate

type Migrator struct {
	*migrationEngine
	closeOnce sync.Once
}

func (r *Migrator) Close(t *testing.T) {
	t.Helper()
	r.closeOnce.Do(func() {
		a, b := r.migrationEngine.Close()
		if a != nil || b != nil {
			t.Error("đóng migration runner thất bại")
		}
	})
}

func (db *Case) OpenMigrator(t *testing.T, source fs.FS) *Migrator {
	t.Helper()
	// Dùng đúng driver hiện hành; không loader/CLI dotenv, không SQL version thủ công.
	sqlDB := stdlibDB(t, db.cfg)
	d, err := driver.WithInstance(sqlDB, &driver.Config{MultiStatementEnabled: false, StatementTimeout: 35 * time.Second})
	if err != nil {
		sqlDB.Close()
	}
	Require(t, "driver migration", err)
	src, err := iofs.New(source, ".")
	Require(t, "nguồn migration", err)
	m, err := migrate.NewWithInstance("iofs", src, "pgx", d)
	Require(t, "runner migration", err)
	r := &Migrator{migrationEngine: m}
	t.Cleanup(func() { r.Close(t) })
	return r
}

func stdlibDB(t *testing.T, cfg *pgx.ConnConfig) *sql.DB {
	t.Helper()
	// ConnString() giữ URL gốc, không phản ánh cfg.Database đã đổi cho từng case.
	// OpenDB dùng trực tiếp config đã xác minh để không migrate nhầm DB gốc test.
	db := stdlib.OpenDB(*cfg)
	var name string
	err := db.QueryRow("SELECT current_database()").Scan(&name)
	if err != nil || name != cfg.Database {
		db.Close()
		t.Fatal("driver migration không đúng database case")
	}
	return db
}

func AssertRootEmpty(t *testing.T) {
	t.Helper()
	cfg := rootConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.ConnectConfig(ctx, cfg)
	Require(t, "probe DB gốc", err)
	defer conn.Close(ctx)
	verifyCluster(t, conn, baseDB)
	var tables, cases int
	Require(t, "DB gốc không bị migrate", conn.QueryRow(ctx, `SELECT count(*) FROM pg_tables WHERE schemaname='public'`).Scan(&tables))
	Require(t, "DB case đã dọn", conn.QueryRow(ctx, `SELECT count(*) FROM pg_database WHERE datname LIKE 'auth_case_%'`).Scan(&cases))
	if tables != 0 || cases != 0 {
		t.Fatal("DB gốc phải trống và không còn DB case; không tự dọn tài nguyên không thuộc lần chạy")
	}
}
