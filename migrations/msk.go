package migrations

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/jackc/pgx/v5/pgconn"
)

const MSKVersion = 7

type mskDriver struct {
	database.Driver
	db            *sql.DB
	importActorID int
}

// WithMSKPreflight kiểm nguồn dưới advisory lock trước khi ghi dirty. Caller
// truyền actor import đã xác minh; DDL kiểm lại guard dưới khóa bảng.
func WithMSKPreflight(driver database.Driver, db *sql.DB, importActorID int) database.Driver {
	return &mskDriver{Driver: driver, db: db, importActorID: importActorID}
}

func (d *mskDriver) SetVersion(version int, dirty bool) error {
	if version == MSKVersion && dirty {
		previous, wasDirty, err := d.Driver.Version()
		if err != nil {
			return errors.New("MSK: không đọc được version nguồn")
		}
		if previous == MSKVersion-1 && !wasDirty {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := CheckMSK(ctx, d.db, d.importActorID); err != nil {
				return err
			}
		}
	}
	return d.Driver.SetVersion(version, dirty)
}

// CheckMSK chỉ đọc schema và dữ liệu nguồn ở mốc PCA; không xuất dữ liệu legacy.
// importActorID=0 chỉ hợp lệ khi không có official cần lưu provenance.
func CheckMSK(ctx context.Context, db *sql.DB, importActorID int) error {
	if importActorID < 0 || int64(importActorID) > 2147483647 {
		return errors.New("MSK: actor import không hợp lệ")
	}
	source, err := FS.ReadFile("000007_manual_skills.up.sql")
	if err != nil {
		return errors.New("MSK: không đọc được nguồn guard")
	}
	_, remaining, foundStart := strings.Cut(string(source), "-- MSK-GUARD-BEGIN")
	guard, _, foundEnd := strings.Cut(remaining, "-- MSK-GUARD-END")
	if !foundStart || !foundEnd || strings.TrimSpace(guard) == "" {
		return errors.New("MSK: nguồn guard không hợp lệ")
	}

	// Cùng một guard cho preflight và DDL để không lệch điều kiện bảo toàn dữ liệu.
	// READ ONLY giữ preflight không thể ghi; cấu hình actor chỉ tồn tại trong tx này.
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return errors.New("MSK: không mở được snapshot preflight")
	}
	defer tx.Rollback()
	var sourceReady bool
	err = tx.QueryRowContext(ctx, `SELECT count(*)=1 FROM public.schema_migrations WHERE version=6 AND NOT dirty`).Scan(&sourceReady)
	if err != nil || !sourceReady {
		return errors.New("MSK: schema nguồn không ở mốc 000006 sạch")
	}
	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.msk_import_actor_id',$1,true)`, strconv.Itoa(importActorID)); err != nil {
		return errors.New("MSK: không truyền được actor import cho guard")
	}
	if _, err = tx.ExecContext(ctx, guard); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Chỉ trả thông báo cố định của guard; lỗi provider khác có thể chứa dữ liệu.
			switch pgErr.Message {
			case "MSK: schema nguồn lệch mốc 000006",
				"MSK: ràng buộc hoặc phụ thuộc nguồn cần review",
				"MSK: tên danh mục hoặc trạng thái legacy cần mapping được review",
				"MSK: nguồn hoặc liên kết approval legacy không hợp lệ",
				"MSK: thiếu actor import đã xác minh",
				"MSK: audit import cần review schema hoặc dependency":
				return errors.New(pgErr.Message)
			}
		}
		return errors.New("MSK: không hoàn tất kiểm schema và dữ liệu preflight")
	}
	if err = tx.Commit(); err != nil {
		return errors.New("MSK: snapshot preflight chưa hoàn tất")
	}
	return nil
}
