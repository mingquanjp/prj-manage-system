// Command migrate chạy database migration.
//
//	go run ./cmd/migrate up        — áp dụng toàn bộ migration còn thiếu
//	go run ./cmd/migrate down      — lùi lại một bước (APP_ENV=dev tường minh)
//	go run ./cmd/migrate down-all  — lùi toàn bộ (APP_ENV=dev tường minh)
//	go run ./cmd/migrate version
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
	if err := validateCommand(command); err != nil {
		return err
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
		err = m.Steps(-1)
	case "down-all":
		// Rollback qua file down để dọn cả ENUM, không chỉ bảng như Drop().
		err = m.Down()
	case "version":
		version, dirty, verr := m.Version()
		if verr != nil {
			return verr
		}
		fmt.Printf("phiên bản schema: %d (dirty=%t)\n", version, dirty)
		return nil
	default:
		return fmt.Errorf("lệnh không hợp lệ: %q", command)
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

// validateCommand chặn lệnh không hợp lệ/nguy hiểm trước khi đọc cấu hình DB.
func validateCommand(command string) error {
	switch command {
	case "up", "version":
		return nil
	case "down", "down-all":
		return config.RequireDev("migrate " + command)
	default:
		return fmt.Errorf("lệnh không hợp lệ: %q (dùng up | down | down-all | version)", command)
	}
}
