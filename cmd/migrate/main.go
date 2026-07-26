// Command migrate chạy database migration.
//
//	go run ./cmd/migrate up        — áp dụng toàn bộ migration còn thiếu
//	go run ./cmd/migrate down      — lùi lại một bước (chỉ khi APP_ENV=dev)
//	go run ./cmd/migrate down-all  — lùi toàn bộ, xóa sạch schema (chỉ khi APP_ENV=dev)
//	go run ./cmd/migrate version
//
// Cố ý không có lệnh "drop". Drop của golang-migrate chỉ xóa bảng, không xóa
// kiểu ENUM, nên lần "up" kế tiếp sẽ hỏng vì CREATE TYPE gặp kiểu đã tồn tại.
// Dùng down-all để rollback qua chính file .down.sql, nơi đã xóa đủ cả 15 kiểu.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"ai-staffing/internal/config"
	"ai-staffing/migrations"

	"github.com/golang-migrate/migrate/v4"
	pgxdriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run() error {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// PgBouncer ở chế độ transaction pooling không giữ được advisory lock mà
	// golang-migrate dùng để chống chạy song song, nên luôn đi endpoint trực tiếp.
	if cfg.IsPooledEndpoint() {
		fmt.Println("DATABASE_URL trỏ tới endpoint -pooler; tự chuyển sang endpoint trực tiếp để chạy migration")
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("không đọc được thư mục migrations: %w", err)
	}

	db, err := sql.Open("pgx", cfg.DirectURL())
	if err != nil {
		return fmt.Errorf("không mở được kết nối: %w", err)
	}
	defer db.Close()

	driver, err := pgxdriver.WithInstance(db, &pgxdriver.Config{})
	if err != nil {
		return fmt.Errorf("không khởi tạo được driver migrate: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("không khởi tạo được migrate: %w", err)
	}

	switch command {
	case "up":
		err = m.Up()
	case "down":
		if err := requireDev(cfg, command); err != nil {
			return err
		}
		err = m.Steps(-1)
	case "down-all":
		if err := requireDev(cfg, command); err != nil {
			return err
		}
		err = m.Down()
	case "version":
		version, dirty, verr := m.Version()
		if verr != nil {
			return verr
		}
		fmt.Printf("phiên bản schema: %d (dirty=%t)\n", version, dirty)
		return nil
	default:
		return fmt.Errorf("lệnh không hợp lệ: %q (dùng up | down | down-all | version)", command)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("schema đã ở phiên bản mới nhất, không có gì để áp dụng")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Printf("migrate %s: hoàn tất\n", command)
	return nil
}

// requireDev chặn các lệnh phá hủy dữ liệu khi không chạy ở môi trường phát triển.
func requireDev(cfg *config.Config, command string) error {
	if cfg.IsDev() {
		return nil
	}

	return fmt.Errorf(
		"lệnh %q xóa dữ liệu nên chỉ chạy được khi APP_ENV=dev (hiện tại: %q)",
		command, cfg.AppEnv,
	)
}
