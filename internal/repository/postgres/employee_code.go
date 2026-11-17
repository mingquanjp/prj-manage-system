package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
)

var ErrEmployeeCode = errors.New("không cấp hoặc đọc được mã nhân viên")

// FormatEmployeeCode đệm tối thiểu sáu chữ số, không cắt số dài hoặc quay vòng.
func FormatEmployeeCode(value int64) (string, error) {
	if value < 1 {
		return "", ErrEmployeeCode
	}
	code := fmt.Sprintf("EMP%06d", value)
	if len(code) > 30 {
		return "", ErrEmployeeCode
	}
	return code, nil
}

// AllocateEmployeeCode chỉ nhận transaction của caller. Rollback không trả lại số.
// Caller vẫn phải INSERT với UNIQUE(employee_code), không tự retry mutation.
func AllocateEmployeeCode(ctx context.Context, tx pgx.Tx) (string, error) {
	var value int64
	if err := tx.QueryRow(ctx, `SELECT nextval('public.employee_code_seq')`).Scan(&value); err != nil {
		return "", ErrEmployeeCode
	}
	return FormatEmployeeCode(value)
}

type sequenceReader interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// PreviewEmployeeCode chỉ đọc ứng viên, không tiêu thụ/giữ chỗ hoặc bảo đảm mã khi lưu.
func PreviewEmployeeCode(ctx context.Context, db sequenceReader) (string, error) {
	var value int64
	var called bool
	if err := db.QueryRow(ctx, `SELECT last_value,is_called FROM public.employee_code_seq`).Scan(&value, &called); err != nil {
		return "", ErrEmployeeCode
	}
	if called {
		if value == math.MaxInt64 {
			return "", ErrEmployeeCode
		}
		value++
	}
	return FormatEmployeeCode(value)
}
