// Command migrate chạy database migration.
//
//	go run ./cmd/migrate up        — áp dụng toàn bộ migration còn thiếu
//	go run ./cmd/migrate down      — lùi lại một bước (APP_ENV=dev tường minh)
//	go run ./cmd/migrate down-all  — lùi toàn bộ (APP_ENV=dev tường minh)
//	go run ./cmd/migrate version
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"ai-staffing/internal/config"
	"ai-staffing/migrations"

	"github.com/golang-migrate/migrate/v4"
	pgxdriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run() error {
	command, importActorID, err := migrationArguments(os.Args[1:])
	if err != nil {
		return err
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
	if cfg.Database.Pooled {
		fmt.Println("DATABASE_URL trỏ tới endpoint -pooler; tự chuyển sang endpoint trực tiếp để chạy migration")
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("không đọc được thư mục migrations: %w", err)
	}

	connection, err := pgx.ParseConfig(cfg.Database.DirectURL)
	if err != nil {
		return errors.New("DB_POOL_INIT_FAILED: không đọc được cấu hình migration")
	}
	connection.RuntimeParams["app.msk_import_actor_id"] = strconv.Itoa(importActorID)
	db := stdlib.OpenDB(*connection)
	defer db.Close()

	driver, err := pgxdriver.WithInstance(db, &pgxdriver.Config{})
	if err != nil {
		return errors.New("DB_CONNECT_FAILED: không khởi tạo được driver migration")
	}

	guarded := migrations.WithM1Preflight(driver, db, nil)
	guarded = migrations.WithM2Preflight(guarded, db)
	guarded = migrations.WithPCAPreflight(guarded, db)
	guarded = migrations.WithMSKPreflight(guarded, db, importActorID)
	m, err := migrate.NewWithInstance("iofs", source, "postgres", guarded)
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

// migrationArguments kiểm đầu vào trước khi cấu hình có thể mở kết nối DB.
func migrationArguments(args []string) (string, int, error) {
	command := "up"
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	actor := flags.Int("msk-import-actor", 0, "user_id người thực hiện import đã xác minh")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return "", 0, errors.New("tham số migration không hợp lệ")
	}
	if *actor < 0 || int64(*actor) > 2147483647 || (*actor != 0 && command != "up") {
		return "", 0, errors.New("actor import chỉ nhận user_id dương cho lệnh up")
	}
	return command, *actor, nil
}
