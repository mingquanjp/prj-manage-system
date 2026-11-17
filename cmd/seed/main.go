// Command seed nạp dữ liệu mẫu phục vụ phát triển và trình diễn.
//
//	go run ./cmd/seed
//
// Chạy lại nhiều lần vẫn an toàn: mọi thao tác ghi đều idempotent.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"ai-staffing/db/seed"
	"ai-staffing/internal/config"
	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Mật khẩu dùng chung cho toàn bộ tài khoản mẫu. Chỉ phục vụ môi trường phát triển.
const demoPassword = "Password@123"

type person struct {
	Email      string
	LegacyCode string
	Fullname   string
	Phone      string
	SystemRole string // Quyền hiện hành chỉ từ user_roles; trống là nhân viên nền.
	Department string
	Position   string
	Seniority  string
}

var people = []person{
	{"admin@company.vn", "EMP001", "Trần Quốc Bảo", "0901000001", "ADMIN", "Phòng Hành chính - Nhân sự", "Quản trị hệ thống", ""},
	{"pm.hoa@company.vn", "EMP002", "Lê Thị Thu Hoa", "0901000002", "", "Phòng Sản phẩm", "Project Manager", "Senior"},
	{"pm.nam@company.vn", "EMP003", "Phạm Hoài Nam", "0901000003", "", "Phòng Phát triển phần mềm", "Project Manager", "Senior"},

	{"dev.an@company.vn", "EMP004", "Nguyễn Văn An", "0901000004", "", "Phòng Phát triển phần mềm", "Backend Developer", "Senior"},
	{"dev.binh@company.vn", "EMP005", "Trần Thanh Bình", "0901000005", "", "Phòng Phát triển phần mềm", "Backend Developer", "Middle"},
	{"dev.cuong@company.vn", "EMP006", "Lê Quốc Cường", "0901000006", "", "Phòng Phát triển phần mềm", "Backend Developer", "Junior"},
	{"dev.dung@company.vn", "EMP007", "Phạm Thùy Dung", "0901000007", "", "Phòng Phát triển phần mềm", "Frontend Developer", "Senior"},
	{"dev.giang@company.vn", "EMP008", "Vũ Hương Giang", "0901000008", "", "Phòng Phát triển phần mềm", "Frontend Developer", "Middle"},
	{"dev.hai@company.vn", "EMP009", "Đỗ Trung Hải", "0901000009", "", "Phòng Phát triển phần mềm", "Frontend Developer", "Junior"},
	{"dev.khanh@company.vn", "EMP010", "Ngô Gia Khánh", "0901000010", "", "Phòng Phát triển phần mềm", "Fullstack Developer", "Senior"},
	{"dev.linh@company.vn", "EMP011", "Bùi Mỹ Linh", "0901000011", "", "Phòng Phát triển phần mềm", "Fullstack Developer", "Middle"},
	{"dev.minh@company.vn", "EMP012", "Hoàng Nhật Minh", "0901000012", "", "Phòng Phát triển phần mềm", "Mobile Developer", "Middle"},
	{"dev.ngan@company.vn", "EMP013", "Đặng Kim Ngân", "0901000013", "", "Phòng Phát triển phần mềm", "Mobile Developer", "Junior"},

	{"qa.phuc@company.vn", "EMP014", "Lý Hồng Phúc", "0901000014", "", "Phòng Kiểm thử", "QA Engineer", "Senior"},
	{"qa.quyen@company.vn", "EMP015", "Trịnh Bảo Quyên", "0901000015", "", "Phòng Kiểm thử", "QA Engineer", "Middle"},

	{"ops.son@company.vn", "EMP016", "Mai Trường Sơn", "0901000016", "", "Phòng Hạ tầng", "DevOps Engineer", "Senior"},
	{"ops.tuan@company.vn", "EMP017", "Chu Anh Tuấn", "0901000017", "", "Phòng Hạ tầng", "DevOps Engineer", "Junior"},

	{"ba.uyen@company.vn", "EMP018", "Dương Tố Uyên", "0901000018", "", "Phòng Sản phẩm", "Business Analyst", "Senior"},
	{"ba.vy@company.vn", "EMP019", "Tạ Thảo Vy", "0901000019", "", "Phòng Sản phẩm", "Business Analyst", "Junior"},

	{"lead.xuan@company.vn", "EMP020", "Đinh Thanh Xuân", "0901000020", "", "Phòng Phát triển phần mềm", "Technical Lead", "Senior"},
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("seed: %v", err)
	}
}

func run() error {
	if err := config.RequireDev("seed dữ liệu mẫu"); err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Luôn dùng simple protocol: các file dữ liệu mẫu chứa nhiều câu lệnh trong
	// một file, extended protocol không chạy được.
	pool, err := postgres.NewPool(ctx, cfg.Database.RuntimeURL, true)
	if err != nil {
		return err
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("không băm được mật khẩu seed")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return errors.New("không bắt đầu được transaction seed")
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
	fmt.Println("✓ hồ sơ năng lực (đã đánh dấu Quản trị viên xác nhận)")

	if err := tx.Commit(ctx); err != nil {
		return errors.New("không xác nhận được commit seed; cần đối soát, không tự chạy lại")
	}

	fmt.Println("Hoàn tất dữ liệu mẫu; đăng nhập bằng email tổng hợp đã cấu hình trong seed.")

	return nil
}

func execSQLFile(ctx context.Context, tx pgx.Tx, name string) error {
	content, err := seed.FS.ReadFile(name)
	if err != nil {
		return errors.New("không đọc được nguồn SQL seed")
	}

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return errors.New("không thực thi được SQL seed")
	}

	return nil
}

// seedPeople nhận diện bộ mẫu theo email tường minh. Khi chạy lại, không sửa
// credential/quyền/hồ sơ hiện có; mã EMP legacy chỉ dùng đối soát, không cấp lại.
func seedPeople(ctx context.Context, tx pgx.Tx, passwordHash string) error {
	// Chỉ seed dev được tuần tự hóa theo bảng để tránh hai lượt tạo user mồ côi.
	if _, err := tx.Exec(ctx, `LOCK TABLE public.users, public.employees IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return errors.New("không khóa được dữ liệu seed")
	}
	for _, p := range people {
		var userID int
		var existingEmail, code string
		err := tx.QueryRow(ctx, `SELECT user_id,email,employee_code FROM employees
   WHERE lower(btrim(email))=$1 OR employee_code=$2`, p.Email, p.LegacyCode).Scan(&userID, &existingEmail, &code)
		if err == nil {
			// QueryRow không được che hai row khớp hai khóa khác nhau.
			var matches int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM employees WHERE lower(btrim(email))=$1 OR employee_code=$2`, p.Email, p.LegacyCode).Scan(&matches); err != nil || matches != 1 || auth.NormalizeLoginEmail(existingEmail) != p.Email {
				return errors.New("seed: email hoặc mã legacy xung đột; cần đối soát")
			}
			var linked bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)`, userID).Scan(&linked); err != nil || !linked {
				return errors.New("seed: liên kết tài khoản không hợp lệ")
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return errors.New("không đọc được tài khoản seed")
		}
		code, err = postgres.AllocateEmployeeCode(ctx, tx)
		if err != nil {
			return errors.New("không cấp được mã seed")
		}
		if err := tx.QueryRow(ctx, `INSERT INTO users(password_hash,system_role,status,must_change_password)
    VALUES($1,'EMPLOYEE','ACTIVE',false) RETURNING user_id`, passwordHash).Scan(&userID); err != nil {
			return errors.New("không tạo được tài khoản seed")
		}
		if _, err := tx.Exec(ctx, `INSERT INTO employees(user_id,employee_code,fullname,email,phone,department_id,position_id,seniority_level)
    VALUES($1,$2,$3,$4,$5,
      (SELECT department_id FROM departments WHERE department_name=$6),
      (SELECT position_id FROM positions WHERE position_name=$7),NULLIF($8,''))`,
			userID, code, p.Fullname, p.Email, p.Phone, p.Department, p.Position, p.Seniority); err != nil {
			return errors.New("không tạo được hồ sơ seed")
		}
		if p.SystemRole != "" {
			if _, err := tx.Exec(ctx, `INSERT INTO user_roles(user_id,role_code) VALUES($1,$2)`, userID, p.SystemRole); err != nil {
				return errors.New("không tạo được quyền seed")
			}
		}
	}
	return nil
}
