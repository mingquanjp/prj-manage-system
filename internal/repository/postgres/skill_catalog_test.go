package postgres

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	sc "ai-staffing/internal/service/skillcatalog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestSkillCatalogCommitClassification(t *testing.T) {
	for _, tc := range []struct {
		name        string
		input, want error
		unknown     bool
	}{
		{"rollback xác nhận", pgx.ErrTxCommitRollback, sc.ErrPersistence, false},
		{"constraint từ chối", &pgconn.PgError{Code: "23514", Severity: "ERROR"}, sc.ErrPersistence, false},
		{"deadlock từ chối", &pgconn.PgError{Code: "40P01", Severity: "ERROR"}, sc.ErrUnavailable, false},
		{"mất phản hồi", io.ErrUnexpectedEOF, sc.ErrUnavailable, true},
		{"kết nối FATAL", &pgconn.PgError{Code: "08006", Severity: "FATAL"}, sc.ErrUnavailable, true},
		{"SQL outcome unknown", &pgconn.PgError{Code: "40003", Severity: "ERROR"}, sc.ErrPersistence, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := skillCatalogCommitError(tc.input)
			if !errors.Is(err, tc.want) || errors.Is(err, sc.ErrCommitUnknown) != tc.unknown {
				t.Fatal("phân loại commit không khớp nguồn lỗi")
			}
		})
	}
}

func TestSkillCatalogConstraintMappingAndTimeout(t *testing.T) {
	for _, name := range []string{"uq_msk_category_name", "uq_msk_skill_name", "skill_categories_category_name_key", "skills_skill_name_key"} {
		if !errors.Is(skillCatalogError(&pgconn.PgError{Code: "23505", ConstraintName: name}), sc.ErrDuplicateName) {
			t.Fatal("unique danh mục không được map duplicate")
		}
	}
	if !errors.Is(skillCatalogError(&pgconn.PgError{Code: "23505", ConstraintName: "another_unique"}), sc.ErrPersistence) {
		t.Fatal("không được gọi constraint ngoài phạm vi là tên trùng")
	}
	for _, name := range []string{"skills_category_id_fkey", "employee_skills_skill_id_fkey", "employee_skill_change_requests_skill_id_fkey", "task_requirements_skill_id_fkey", "cv_extracted_skills_skill_id_fkey"} {
		if !errors.Is(skillCatalogError(&pgconn.PgError{Code: "23503", ConstraintName: name}), sc.ErrCatalogInUse) {
			t.Fatal("FK danh mục không được map in-use")
		}
	}
	if _, err := NewSkillCatalogRepository(nil, time.Second); !errors.Is(err, sc.ErrValidation) {
		t.Fatal("constructor nhận DB thiếu")
	}
	repo, err := NewSkillCatalogRepository(catalogDeadlineDB{}, time.Millisecond)
	if err != nil {
		t.Fatal("constructor timeout hợp lệ bị từ chối")
	}
	_, err = repo.List(t.Context(), sc.ListQuery{ActorID: 1, Kind: sc.Skills})
	if !errors.Is(err, sc.ErrUnavailable) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("deadline nội bộ mất nguồn unavailable")
	}
}
