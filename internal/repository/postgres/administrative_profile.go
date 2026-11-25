package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AdministrativeProfileRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ org.AdministrativeProfileRepository = (*AdministrativeProfileRepository)(nil)

func NewAdministrativeProfileRepository(db catalogDB, timeout time.Duration) (*AdministrativeProfileRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, org.ErrValidation
	}
	return &AdministrativeProfileRepository{db, timeout}, nil
}
func administrativeDeadline(ctx context.Context, resultErr *error) {
	if errors.Is(*resultErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), org.ErrUnavailable) {
		*resultErr = errors.Join(org.ErrUnavailable, *resultErr)
	}
}
func administrativeError(err error) error {
	var p *pgconn.PgError
	if errors.As(err, &p) && p.Code == "23503" {
		switch p.ConstraintName {
		case "employees_position_id_fkey", "employees_department_id_fkey", "employees_job_level_id_fkey":
			return org.ErrCatalogNotFound
		}
	}
	return accountError(err)
}

const administrativeSelect = `SELECT e.user_id,e.employee_id,e.employee_code,e.fullname,e.email,e.phone,e.address,
 p.position_id,p.position_name,d.department_id,d.department_name,j.job_level_id,j.job_level_name,e.employment_status
 FROM public.employees e LEFT JOIN public.positions p ON p.position_id=e.position_id
 LEFT JOIN public.departments d ON d.department_id=e.department_id
 LEFT JOIN public.job_levels j ON j.job_level_id=e.job_level_id`

func scanAdministrative(row pgx.Row) (org.AdministrativeProfile, error) {
	var p org.AdministrativeProfile
	var pi, di, ji *int
	var pn, dn, jn *string
	err := row.Scan(&p.UserID, &p.EmployeeID, &p.EmployeeCode, &p.Fullname, &p.CompanyEmail, &p.Phone, &p.Address, &pi, &pn, &di, &dn, &ji, &jn, &p.EmploymentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, org.ErrNotFound
	}
	if err != nil {
		return p, administrativeError(err)
	}
	if pi == nil || pn == nil {
		return p, org.ErrPersistence
	}
	p.Position = org.CatalogItem{ID: *pi, Name: *pn}
	if di != nil && dn != nil {
		p.Department = &org.CatalogItem{ID: *di, Name: *dn}
	}
	if ji != nil && jn != nil {
		p.JobLevel = &org.CatalogItem{ID: *ji, Name: *jn}
	}
	p.CompanyEmail = auth.NormalizeLoginEmail(p.CompanyEmail)
	return p, nil
}
func (r *AdministrativeProfileRepository) ReadAdministrativeProfile(ctx context.Context, q org.AdministrativeProfileQuery) (result org.AdministrativeProfile, resultErr error) {
	if err := org.ValidateAdministrativeProfileQuery(q); err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer administrativeDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, administrativeError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = checkCatalogActor(ctx, tx, q.ActorID, false); err != nil {
		return result, err
	}
	if err = administrativeHR(ctx, tx, q.ActorID); err != nil {
		return result, err
	}
	p, err := scanAdministrative(tx.QueryRow(ctx, administrativeSelect+` WHERE e.employee_id=$1`, q.EmployeeID))
	if err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, administrativeError(err)
	}
	return p, nil
}
func administrativeHR(ctx context.Context, tx pgx.Tx, id int) error {
	var hr bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.user_roles WHERE user_id=$1 AND role_code='HR')`, id).Scan(&hr); err != nil {
		return administrativeError(err)
	}
	if !hr {
		return org.ErrForbidden
	}
	return nil
}
func (r *AdministrativeProfileRepository) ListAdministrativeEmployees(ctx context.Context, q org.ListAdministrativeEmployeesQuery) (result org.AdministrativeEmployeePage, resultErr error) {
	q, err := org.ValidateListAdministrativeEmployees(q)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer administrativeDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, administrativeError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = checkCatalogActor(ctx, tx, q.ActorID, false); err != nil {
		return result, err
	}
	if err = administrativeHR(ctx, tx, q.ActorID); err != nil {
		return result, err
	}
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(q.Query) + "%"
	filter := ` WHERE (e.employee_code ILIKE $1 ESCAPE '\' OR e.fullname ILIKE $1 ESCAPE '\' OR e.email ILIKE $1 ESCAPE '\')`
	page := org.AdministrativeEmployeePage{Items: []org.EmployeeSummary{}, Page: q.Page, PageSize: q.PageSize}
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM public.employees e`+filter, pattern).Scan(&page.Total); err != nil {
		return result, administrativeError(err)
	}
	// Projection danh sách không đọc phone/address; count và items dùng cùng snapshot.
	rows, err := tx.Query(ctx, `SELECT e.employee_id,e.employee_code,e.fullname,e.email,p.position_id,p.position_name,d.department_id,d.department_name,j.job_level_id,j.job_level_name,e.employment_status
 FROM public.employees e LEFT JOIN public.positions p ON p.position_id=e.position_id LEFT JOIN public.departments d ON d.department_id=e.department_id LEFT JOIN public.job_levels j ON j.job_level_id=e.job_level_id`+filter+` ORDER BY e.employee_id LIMIT $2 OFFSET $3`, pattern, q.PageSize, q.Offset())
	if err != nil {
		return result, administrativeError(err)
	}
	for rows.Next() {
		var p org.EmployeeSummary
		var pi, di, ji *int
		var pn, dn, jn *string
		if err = rows.Scan(&p.EmployeeID, &p.EmployeeCode, &p.Fullname, &p.CompanyEmail, &pi, &pn, &di, &dn, &ji, &jn, &p.EmploymentStatus); err != nil {
			rows.Close()
			return result, administrativeError(err)
		}
		if pi == nil || pn == nil {
			rows.Close()
			return result, org.ErrPersistence
		}
		p.Position = org.CatalogItem{ID: *pi, Name: *pn}
		if di != nil && dn != nil {
			p.Department = &org.CatalogItem{ID: *di, Name: *dn}
		}
		if ji != nil && jn != nil {
			p.JobLevel = &org.CatalogItem{ID: *ji, Name: *jn}
		}
		p.CompanyEmail = auth.NormalizeLoginEmail(p.CompanyEmail)
		page.Items = append(page.Items, p)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return result, administrativeError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return result, administrativeError(err)
	}
	return page, nil
}

func administrativeValues(v org.AdministrativeValues) map[string]any {
	m := map[string]any{}
	for k, f := range map[string]org.AdministrativeField[string]{"fullname": v.Fullname, "company_email": v.CompanyEmail, "phone": v.Phone, "address": v.Address} {
		if f.Present {
			if f.Value == nil {
				m[k] = nil
			} else {
				m[k] = *f.Value
			}
		}
	}
	for k, f := range map[string]org.AdministrativeField[int]{"position_id": v.PositionID, "department_id": v.DepartmentID, "job_level_id": v.JobLevelID} {
		if f.Present {
			if f.Value == nil {
				m[k] = nil
			} else {
				m[k] = *f.Value
			}
		}
	}
	return m
}
func administrativeCurrent(p org.AdministrativeProfile) map[string]any {
	m := map[string]any{"fullname": p.Fullname, "company_email": p.CompanyEmail, "phone": nil, "address": nil, "position_id": p.Position.ID, "department_id": nil, "job_level_id": nil}
	if p.Phone != nil {
		m["phone"] = *p.Phone
	}
	if p.Address != nil {
		m["address"] = *p.Address
	}
	if p.Department != nil {
		m["department_id"] = p.Department.ID
	}
	if p.JobLevel != nil {
		m["job_level_id"] = p.JobLevel.ID
	}
	return m
}
func lockAdministrativeParticipants(ctx context.Context, tx pgx.Tx, q org.AdministrativeProfileQuery) error {
	var actorEmployee, targetUser int
	err := tx.QueryRow(ctx, `SELECT employee_id FROM public.employees WHERE user_id=$1`, q.ActorID).Scan(&actorEmployee)
	if errors.Is(err, pgx.ErrNoRows) {
		return org.ErrUnauthenticated
	}
	if err != nil {
		return administrativeError(err)
	}
	err = tx.QueryRow(ctx, `SELECT user_id FROM public.employees WHERE employee_id=$1`, q.EmployeeID).Scan(&targetUser)
	missing := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !missing {
		return administrativeError(err)
	}
	users, employees := []int{q.ActorID}, []int{actorEmployee}
	if !missing {
		users = append(users, targetUser)
		employees = append(employees, q.EmployeeID)
	}
	sort.Ints(users)
	sort.Ints(employees)
	for _, g := range []struct {
		ids []int
		sql string
	}{{users, `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`}, {employees, `SELECT employee_id FROM public.employees WHERE employee_id=$1 FOR UPDATE`}} {
		for i, id := range g.ids {
			if i > 0 && id == g.ids[i-1] {
				continue
			}
			var locked int
			if err = tx.QueryRow(ctx, g.sql, id).Scan(&locked); err != nil {
				return administrativeError(err)
			}
		}
	}
	if err = checkCatalogActor(ctx, tx, q.ActorID, false); err != nil {
		return err
	}
	if err = administrativeHR(ctx, tx, q.ActorID); err != nil {
		return err
	}
	if missing {
		return org.ErrNotFound
	}
	var currentActor, currentTarget int
	if err = tx.QueryRow(ctx, `SELECT employee_id FROM public.employees WHERE user_id=$1`, q.ActorID).Scan(&currentActor); err != nil {
		return administrativeError(err)
	}
	if err = tx.QueryRow(ctx, `SELECT user_id FROM public.employees WHERE employee_id=$1`, q.EmployeeID).Scan(&currentTarget); err != nil {
		return administrativeError(err)
	}
	if currentActor != actorEmployee || currentTarget != targetUser {
		return org.ErrPersistence
	}
	return nil
}
func (r *AdministrativeProfileRepository) EditAdministrativeProfile(ctx context.Context, c org.EditAdministrativeProfileCommand) (result org.AdministrativeProfile, resultErr error) {
	c, err := org.ValidateEditAdministrativeProfile(c)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer administrativeDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return result, administrativeError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = lockAdministrativeParticipants(ctx, tx, org.AdministrativeProfileQuery{ActorID: c.ActorID, EmployeeID: c.EmployeeID}); err != nil {
		return result, err
	}
	for _, cat := range []struct {
		field org.AdministrativeField[int]
		sql   string
	}{{c.Changes.DepartmentID, `SELECT department_id FROM public.departments WHERE department_id=$1 FOR KEY SHARE`}, {c.Changes.PositionID, `SELECT position_id FROM public.positions WHERE position_id=$1 FOR KEY SHARE`}, {c.Changes.JobLevelID, `SELECT job_level_id FROM public.job_levels WHERE job_level_id=$1 FOR KEY SHARE`}} {
		if cat.field.Present && cat.field.Value != nil {
			var id int
			err = tx.QueryRow(ctx, cat.sql, *cat.field.Value).Scan(&id)
			if errors.Is(err, pgx.ErrNoRows) {
				return result, org.ErrCatalogNotFound
			}
			if err != nil {
				return result, administrativeError(err)
			}
		}
	}
	p, err := scanAdministrative(tx.QueryRow(ctx, administrativeSelect+` WHERE e.employee_id=$1`, c.EmployeeID))
	if err != nil {
		return result, err
	}
	current, before, changes := administrativeCurrent(p), administrativeValues(c.ExpectedBeforeValues), administrativeValues(c.Changes)
	changedBefore, changedAfter := map[string]any{}, map[string]any{}
	for k, v := range before {
		if !reflect.DeepEqual(v, current[k]) {
			return result, org.ErrStale
		}
		if !reflect.DeepEqual(changes[k], current[k]) {
			changedBefore[k] = current[k]
			changedAfter[k] = changes[k]
		}
	}
	if len(changedAfter) > 0 {
		if email, changed := changedAfter["company_email"]; changed {
			var canonical string
			if err = tx.QueryRow(ctx, `SELECT lower(btrim($1::text))`, email).Scan(&canonical); err != nil {
				return result, administrativeError(err)
			}
			if canonical != email {
				return result, org.ErrPersistence
			}
		}
		// Tên cột chỉ lấy từ allowlist cố định, mọi giá trị đều parameter.
		columns := map[string]string{"fullname": "fullname", "company_email": "email", "phone": "phone", "address": "address", "position_id": "position_id", "department_id": "department_id", "job_level_id": "job_level_id"}
		keys := []string{}
		for k := range changedAfter {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		args := []any{c.EmployeeID}
		sets := []string{}
		for _, k := range keys {
			args = append(args, changedAfter[k])
			sets = append(sets, columns[k]+"=$"+strconv.Itoa(len(args)))
		}
		tag, updateErr := tx.Exec(ctx, `UPDATE public.employees SET `+strings.Join(sets, ",")+` WHERE employee_id=$1`, args...)
		if updateErr != nil {
			return result, administrativeError(updateErr)
		}
		if tag.RowsAffected() != 1 {
			return result, org.ErrPersistence
		}
		b, _ := json.Marshal(changedBefore)
		a, _ := json.Marshal(changedAfter)
		if _, err = tx.Exec(ctx, `INSERT INTO public.audit_events(actor_id,entity_type,entity_id,operation,before_values,after_values) VALUES($1,'employee',$2,'UPDATE_ADMINISTRATIVE_PROFILE',$3::jsonb,$4::jsonb)`, c.ActorID, c.EmployeeID, b, a); err != nil {
			return result, administrativeError(err)
		}
		if _, emailChanged := changedAfter["company_email"]; emailChanged {
			rows, e := tx.Query(ctx, `SELECT session_id FROM public.auth_sessions WHERE user_id=$1 ORDER BY session_id FOR UPDATE`, p.UserID)
			if e != nil {
				return result, administrativeError(e)
			}
			for rows.Next() {
				var id int64
				if e = rows.Scan(&id); e != nil {
					rows.Close()
					return result, administrativeError(e)
				}
			}
			rows.Close()
			if e = rows.Err(); e != nil {
				return result, administrativeError(e)
			}
			if _, err = tx.Exec(ctx, `UPDATE public.auth_sessions SET revoked_at=clock_timestamp() WHERE user_id=$1 AND revoked_at IS NULL`, p.UserID); err != nil {
				return result, administrativeError(err)
			}
		}
		p, err = scanAdministrative(tx.QueryRow(ctx, administrativeSelect+` WHERE e.employee_id=$1`, c.EmployeeID))
		if err != nil {
			return result, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return result, accountCommitError(err)
	}
	return p, nil
}
