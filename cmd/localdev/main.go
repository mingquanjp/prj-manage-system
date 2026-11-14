// Command localdev chạy auth local với dữ liệu giả riêng, không dùng loader .env hoặc DATABASE_URL.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-staffing/internal/app"
	"ai-staffing/migrations"
	"github.com/golang-migrate/migrate/v4"
	driver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

const baseDB = "ai_staffing_auth_test"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		// Lỗi driver có thể chứa credential hoặc SQL; chỉ in thông báo giai đoạn đã làm an toàn.
		fmt.Fprintln(os.Stderr, "DỪNG local dev:", err)
		os.Exit(1)
	}
}

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

func rootConfig(raw, systemID string) (*pgx.ConnConfig, error) {
	u, err := url.Parse(raw)
	if err != nil || raw == "" || systemID == "" || u.Scheme != "postgresql" ||
		u.Hostname() != "127.0.0.1" || u.Port() != "55438" || u.User == nil ||
		u.User.Username() != "ai_staffing_test" || u.Path != "/"+baseDB ||
		u.RawQuery != "sslmode=disable" || u.Fragment != "" {
		return nil, errors.New("chỉ nhận cluster local sở hữu từ runner")
	}
	cfg, err := pgx.ParseConfig(raw)
	if err != nil {
		return nil, errors.New("cấu hình local không hợp lệ")
	}
	cfg.ConnectTimeout = 5 * time.Second
	return cfg, nil
}

func verify(ctx context.Context, conn *pgx.Conn, name, systemID string) error {
	var major int
	var database, user, system string
	var owner bool
	err := conn.QueryRow(ctx, `SELECT current_setting('server_version_num')::int/10000,
 current_database(),current_user,pg_get_userbyid(datdba)=current_user,
 (SELECT system_identifier::text FROM pg_control_system())
 FROM pg_database WHERE datname=current_database()`).Scan(&major, &database, &user, &owner, &system)
	if err != nil || major != 18 || database != name || user != "ai_staffing_test" || !owner || system != systemID {
		return errors.New("DB/owner/cluster local không khớp; từ chối thao tác")
	}
	return nil
}

func run(ctx context.Context) (result error) {
	systemID := os.Getenv("AUTH_LOCAL_SYSTEM_ID")
	cfg, err := rootConfig(os.Getenv("AUTH_LOCAL_DATABASE_URL"), systemID)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return errors.New("cổng backend 8080 đã bị dùng hoặc không mở được")
	}
	defer listener.Close()
	setup, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	root, err := pgx.ConnectConfig(setup, cfg)
	if err != nil {
		return errors.New("không kết nối được PostgreSQL local")
	}
	defer root.Close(context.Background())
	if err = verify(setup, root, baseDB, systemID); err != nil {
		return err
	}
	var random [8]byte
	if _, err = rand.Read(random[:]); err != nil {
		return errors.New("không tạo được định danh DB local")
	}
	name := "auth_dev_" + hex.EncodeToString(random[:])
	if _, err = root.Exec(setup, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()+" TEMPLATE template0"); err != nil {
		return errors.New("không tạo được DB local riêng")
	}
	// Chỉ dọn tên vừa CREATE thành công; không nhận hoặc reset DB có sẵn.
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 15*time.Second)
		defer stop()
		admin, err := pgx.ConnectConfig(cleanup, cfg)
		if err != nil {
			result = errors.New("chưa dọn được DB local; giữ để đối soát")
			return
		}
		defer admin.Close(context.Background())
		if err = verify(cleanup, admin, baseDB, systemID); err != nil {
			result = err
			return
		}
		var owned bool
		if err = admin.QueryRow(cleanup, `SELECT pg_get_userbyid(datdba)=current_user FROM pg_database WHERE datname=$1`, name).Scan(&owned); err != nil || !owned {
			result = errors.New("owner DB local đã đổi; không xóa")
			return
		}
		if _, err = admin.Exec(cleanup, "DROP DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
			result = errors.New("chưa dọn được DB local; không FORCE/prune")
		}
	}()
	child := cfg.Copy()
	child.Database = name
	if err = prepare(setup, child, systemID); err != nil {
		return err
	}
	poolConfig, err := pgxpool.ParseConfig("postgresql://127.0.0.1/unused?sslmode=disable")
	if err != nil {
		return errors.New("không tạo được cấu hình pool local")
	}
	poolConfig.ConnConfig = child.Copy()
	pool, err := pgxpool.NewWithConfig(setup, poolConfig)
	if err != nil {
		return errors.New("không mở được pool local")
	}
	defer pool.Close()
	handler, err := app.LocalAuthHandler(pool, clock{})
	if err != nil {
		return errors.New("không lắp được auth local")
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	done := make(chan error, 1)
	go func() { done <- srv.Serve(listener) }()
	fmt.Println("Backend local sẵn sàng: 127.0.0.1:8080; mở FE tại http://127.0.0.1:5174. Swagger: http://127.0.0.1:8080/docs/; Ctrl+C để dừng.")
	select {
	case <-ctx.Done():
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return errors.New("HTTP local dừng ngoài dự kiến")
		}
	}
	shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	if err = srv.Shutdown(shutdown); err != nil {
		_ = srv.Close()
		return errors.New("HTTP local chưa tắt hoàn toàn")
	}
	return nil
}

func prepare(ctx context.Context, cfg *pgx.ConnConfig, systemID string) error {
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return errors.New("không kết nối được DB local vừa tạo")
	}
	defer conn.Close(context.Background())
	if err = verify(ctx, conn, cfg.Database, systemID); err != nil {
		return err
	}
	sqlDB := stdlib.OpenDB(*cfg)
	defer sqlDB.Close()
	var database string
	if err = sqlDB.QueryRowContext(ctx, "SELECT current_database()").Scan(&database); err != nil || database != cfg.Database {
		return errors.New("driver migration không đúng DB local")
	}
	d, err := driver.WithInstance(sqlDB, &driver.Config{MultiStatementEnabled: false, StatementTimeout: 35 * time.Second})
	if err != nil {
		return errors.New("không mở được driver migration local")
	}
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		_ = d.Close()
		return errors.New("không mở được nguồn migration")
	}
	m, err := migrate.NewWithInstance("iofs", source, "pgx", d)
	if err != nil {
		_ = source.Close()
		_ = d.Close()
		return errors.New("không tạo được migration local")
	}
	err = m.Migrate(3)
	a, b := m.Close()
	if err != nil || a != nil || b != nil {
		return errors.New("migration DB local không hoàn tất")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("không tạo được hash dữ liệu local")
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.New("không tạo được dữ liệu local")
	}
	defer tx.Rollback(context.Background())
	var position int
	if err = tx.QueryRow(ctx, `INSERT INTO positions(position_name) VALUES ('Vị trí local demo') RETURNING position_id`).Scan(&position); err != nil {
		return errors.New("không tạo được vị trí local")
	}
	for i, account := range []struct {
		email, name, role string
		limited           bool
	}{
		{"quan1652005@gmail.com", "Nhân viên local", "HR", false},
		{"quan1652005+limited@gmail.com", "Tài khoản local buộc đổi", "ADMIN", true},
	} {
		var user int
		if err = tx.QueryRow(ctx, `INSERT INTO users(username,password_hash,system_role,status,must_change_password)
 VALUES ($1,$2,'EMPLOYEE','ACTIVE',$3) RETURNING user_id`, fmt.Sprintf("local-%d", i+1), string(hash), account.limited).Scan(&user); err != nil {
			return errors.New("không tạo được tài khoản local")
		}
		if _, err = tx.Exec(ctx, `INSERT INTO employees(user_id,employee_code,fullname,email,position_id,employment_status)
 VALUES ($1,$2,$3,$4,$5,'ACTIVE')`, user, fmt.Sprintf("LOCAL-%03d", i+1), account.name, account.email, position); err != nil {
			return errors.New("không tạo được hồ sơ local")
		}
		if _, err = tx.Exec(ctx, `INSERT INTO user_roles(user_id,role_code) VALUES ($1,$2)`, user, account.role); err != nil {
			return errors.New("không tạo được quyền local")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return errors.New("không commit được dữ liệu local")
	}
	return nil
}
