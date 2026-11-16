package postgres

import (
	"context"
	"errors"
	"time"

	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
)

type ProfileRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ org.ProfileRepository = (*ProfileRepository)(nil)

func NewProfileRepository(db catalogDB, timeout time.Duration) (*ProfileRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, org.ErrValidation
	}
	return &ProfileRepository{db, timeout}, nil
}
func (r *ProfileRepository) ReadSelf(ctx context.Context, q org.SelfProfileQuery) (result org.SelfProfile, resultErr error) {
	if err := org.ValidateSelfProfileQuery(q); err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer func() {
		if errors.Is(resultErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), org.ErrUnavailable) {
			resultErr = org.ErrUnavailable
		}
	}()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, catalogError(err, catalogSpec{})
	}
	defer rollbackAuthTransaction(ctx, tx)
	// Đọc actor riêng để không gộp mất quyền/khóa vào lỗi 404 của ID yêu cầu.
	var active, limited bool
	var employeeActive *bool
	err = tx.QueryRow(ctx, `SELECT u.status='ACTIVE',u.must_change_password,e.employment_status='ACTIVE'
 FROM public.users u LEFT JOIN public.employees e ON e.user_id=u.user_id WHERE u.user_id=$1`, q.ActorID).Scan(&active, &limited, &employeeActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, org.ErrUnauthenticated
	}
	if err != nil {
		return result, catalogError(err, catalogSpec{})
	}
	if !active || employeeActive == nil || !*employeeActive {
		return result, org.ErrUnauthenticated
	}
	if limited {
		return result, org.ErrPasswordChangeRequired
	}
	var positionID, departmentID, levelID *int
	var positionName, departmentName, levelName *string
	p := org.SelfProfile{}
	err = tx.QueryRow(ctx, `SELECT e.user_id,e.employee_id,e.employee_code,e.fullname,e.email,e.phone,e.address,
 p.position_id,p.position_name,d.department_id,d.department_name,j.job_level_id,j.job_level_name,e.employment_status
 FROM public.employees e LEFT JOIN public.positions p ON p.position_id=e.position_id
 LEFT JOIN public.departments d ON d.department_id=e.department_id
 LEFT JOIN public.job_levels j ON j.job_level_id=e.job_level_id
 WHERE e.user_id=$1 AND e.employee_id=$2`, q.ActorID, q.EmployeeID).Scan(&p.UserID, &p.EmployeeID, &p.EmployeeCode, &p.Fullname, &p.CompanyEmail, &p.Phone, &p.Address, &positionID, &positionName, &departmentID, &departmentName, &levelID, &levelName, &p.EmploymentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, org.ErrNotFound
	}
	if err != nil {
		return result, catalogError(err, catalogSpec{})
	}
	// Position bắt buộc; dữ liệu nền hỏng không giả thành hồ sơ không tồn tại.
	if positionID == nil || positionName == nil {
		return result, org.ErrPersistence
	}
	p.Position = org.CatalogItem{ID: *positionID, Name: *positionName}
	if departmentID != nil && departmentName != nil {
		p.Department = &org.CatalogItem{ID: *departmentID, Name: *departmentName}
	}
	if levelID != nil && levelName != nil {
		p.JobLevel = &org.CatalogItem{ID: *levelID, Name: *levelName}
	}
	if err = tx.Commit(ctx); err != nil {
		return result, catalogError(err, catalogSpec{})
	}
	return p, nil
}
