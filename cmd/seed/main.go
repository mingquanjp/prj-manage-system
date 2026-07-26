// Command seed nạp dữ liệu mẫu phục vụ phát triển và trình diễn.
//
//	go run ./cmd/seed
//
// Chạy lại nhiều lần vẫn an toàn: mọi thao tác ghi đều idempotent.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ai-staffing/db/seed"
	"ai-staffing/internal/config"
	"ai-staffing/internal/repository/postgres"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Mật khẩu dùng chung cho toàn bộ tài khoản mẫu. Chỉ phục vụ môi trường phát triển.
const demoPassword = "Password@123"

type person struct {
	Username   string
	Code       string
	Fullname   string
	Phone      string
	Role       string // ADMIN | PM | EMPLOYEE
	Department string
	Position   string
	Seniority  string
}

var people = []person{
	{"admin", "EMP001", "Trần Quốc Bảo", "0901000001", "ADMIN", "Phòng Hành chính - Nhân sự", "Quản trị hệ thống", ""},
	{"pm.hoa", "EMP002", "Lê Thị Thu Hoa", "0901000002", "PM", "Phòng Sản phẩm", "Project Manager", "Senior"},
	{"pm.nam", "EMP003", "Phạm Hoài Nam", "0901000003", "PM", "Phòng Phát triển phần mềm", "Project Manager", "Senior"},

	{"dev.an", "EMP004", "Nguyễn Văn An", "0901000004", "EMPLOYEE", "Phòng Phát triển phần mềm", "Backend Developer", "Senior"},
	{"dev.binh", "EMP005", "Trần Thanh Bình", "0901000005", "EMPLOYEE", "Phòng Phát triển phần mềm", "Backend Developer", "Middle"},
	{"dev.cuong", "EMP006", "Lê Quốc Cường", "0901000006", "EMPLOYEE", "Phòng Phát triển phần mềm", "Backend Developer", "Junior"},
	{"dev.dung", "EMP007", "Phạm Thùy Dung", "0901000007", "EMPLOYEE", "Phòng Phát triển phần mềm", "Frontend Developer", "Senior"},
	{"dev.giang", "EMP008", "Vũ Hương Giang", "0901000008", "EMPLOYEE", "Phòng Phát triển phần mềm", "Frontend Developer", "Middle"},
	{"dev.hai", "EMP009", "Đỗ Trung Hải", "0901000009", "EMPLOYEE", "Phòng Phát triển phần mềm", "Frontend Developer", "Junior"},
	{"dev.khanh", "EMP010", "Ngô Gia Khánh", "0901000010", "EMPLOYEE", "Phòng Phát triển phần mềm", "Fullstack Developer", "Senior"},
	{"dev.linh", "EMP011", "Bùi Mỹ Linh", "0901000011", "EMPLOYEE", "Phòng Phát triển phần mềm", "Fullstack Developer", "Middle"},
	{"dev.minh", "EMP012", "Hoàng Nhật Minh", "0901000012", "EMPLOYEE", "Phòng Phát triển phần mềm", "Mobile Developer", "Middle"},
	{"dev.ngan", "EMP013", "Đặng Kim Ngân", "0901000013", "EMPLOYEE", "Phòng Phát triển phần mềm", "Mobile Developer", "Junior"},

	{"qa.phuc", "EMP014", "Lý Hồng Phúc", "0901000014", "EMPLOYEE", "Phòng Kiểm thử", "QA Engineer", "Senior"},
	{"qa.quyen", "EMP015", "Trịnh Bảo Quyên", "0901000015", "EMPLOYEE", "Phòng Kiểm thử", "QA Engineer", "Middle"},

	{"ops.son", "EMP016", "Mai Trường Sơn", "0901000016", "EMPLOYEE", "Phòng Hạ tầng", "DevOps Engineer", "Senior"},
	{"ops.tuan", "EMP017", "Chu Anh Tuấn", "0901000017", "EMPLOYEE", "Phòng Hạ tầng", "DevOps Engineer", "Junior"},

	{"ba.uyen", "EMP018", "Dương Tố Uyên", "0901000018", "EMPLOYEE", "Phòng Sản phẩm", "Business Analyst", "Senior"},
	{"ba.vy", "EMP019", "Tạ Thảo Vy", "0901000019", "EMPLOYEE", "Phòng Sản phẩm", "Business Analyst", "Junior"},

	{"lead.xuan", "EMP020", "Đinh Thanh Xuân", "0901000020", "EMPLOYEE", "Phòng Phát triển phần mềm", "Technical Lead", "Senior"},
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("seed: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Toàn bộ tài khoản mẫu dùng chung một mật khẩu nằm công khai trong repo, trong
	// đó có tài khoản quản trị. Chạy nhầm vào cơ sở dữ liệu thật là mở toang hệ thống.
	if !cfg.IsDev() {
		return fmt.Errorf(
			"seed tạo tài khoản mẫu với mật khẩu công khai nên chỉ chạy được khi APP_ENV=dev (hiện tại: %q)",
			cfg.AppEnv,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Luôn dùng simple protocol: các file dữ liệu mẫu chứa nhiều câu lệnh trong
	// một file, extended protocol không chạy được.
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL, true)
	if err != nil {
		return err
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("không băm được mật khẩu: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := execSQLFile(ctx, tx, "01_reference.sql"); err != nil {
		return err
	}
	fmt.Println("✓ danh mục nền (phòng ban, chức danh, kỹ năng)")

	if err := seedPeople(ctx, tx, string(hash)); err != nil {
		return err
	}
	fmt.Printf("✓ %d tài khoản và hồ sơ nhân sự\n", len(people))

	if err := execSQLFile(ctx, tx, "02_employee_skills.sql"); err != nil {
		return err
	}
	fmt.Println("✓ hồ sơ năng lực: bản ghi đã duyệt và bản ghi chờ Quản trị viên duyệt")

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("không commit được: %w", err)
	}

	fmt.Printf("\nHoàn tất. Tài khoản mẫu — mật khẩu chung: %s\n", demoPassword)
	fmt.Println("  admin    → Quản trị viên")
	fmt.Println("  pm.hoa   → Quản lý dự án")
	fmt.Println("  dev.an   → Nhân viên")

	return nil
}

func execSQLFile(ctx context.Context, tx pgx.Tx, name string) error {
	content, err := seed.FS.ReadFile(name)
	if err != nil {
		return fmt.Errorf("không đọc được %s: %w", name, err)
	}

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("lỗi khi chạy %s: %w", name, err)
	}

	return nil
}

// seedPeople tạo tài khoản trước, rồi tạo hồ sơ nhân sự liên kết 1-1 với tài khoản đó.
//
// ON CONFLICT ... DO UPDATE ghi đè đúng các cột dữ liệu chứ không chỉ chạm vào
// updated_at: bảng people dưới đây là trạng thái mong muốn, nên sửa vai trò hay
// phòng ban ở đó rồi chạy lại seed phải thấy thay đổi thật trong cơ sở dữ liệu.
func seedPeople(ctx context.Context, tx pgx.Tx, passwordHash string) error {
	const insertUser = `
		INSERT INTO users (username, password_hash, system_role)
		VALUES ($1, $2, $3::user_system_role)
		ON CONFLICT (username) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			system_role   = EXCLUDED.system_role
		RETURNING user_id`

	const insertEmployee = `
		INSERT INTO employees (
			user_id, employee_code, fullname, email, phone,
			department_id, position_id, seniority_level
		)
		VALUES (
			$1, $2, $3, $4, $5,
			(SELECT department_id FROM departments WHERE department_name = $6),
			(SELECT position_id   FROM positions   WHERE position_name   = $7),
			NULLIF($8, '')
		)
		ON CONFLICT (employee_code) DO UPDATE SET
			fullname        = EXCLUDED.fullname,
			email           = EXCLUDED.email,
			phone           = EXCLUDED.phone,
			department_id   = EXCLUDED.department_id,
			position_id     = EXCLUDED.position_id,
			seniority_level = EXCLUDED.seniority_level`

	for _, p := range people {
		var userID int
		if err := tx.QueryRow(ctx, insertUser, p.Username, passwordHash, p.Role).Scan(&userID); err != nil {
			return fmt.Errorf("tạo tài khoản %s: %w", p.Username, err)
		}

		email := p.Username + "@company.vn"
		_, err := tx.Exec(ctx, insertEmployee,
			userID, p.Code, p.Fullname, email, p.Phone,
			p.Department, p.Position, p.Seniority,
		)
		if err != nil {
			return fmt.Errorf("tạo hồ sơ %s: %w", p.Code, err)
		}
	}

	return nil
}
