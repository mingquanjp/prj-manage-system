package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type catalogDB interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}
type CatalogRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ org.CatalogRepository = (*CatalogRepository)(nil)

// Timeout do composition root cấp, giới hạn cả chờ khóa và transaction.
func NewCatalogRepository(db catalogDB, timeout time.Duration) (*CatalogRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, org.ErrValidation
	}
	return &CatalogRepository{db: db, timeout: timeout}, nil
}

type catalogSpec struct{ table, id, name, entity, unique, legacyUnique, fk string }

func catalogSpecification(kind org.CatalogKind) (catalogSpec, error) {
	switch kind {
	case org.Departments:
		return catalogSpec{"public.departments", "department_id", "department_name", "department", "uq_departments_normalized_name", "departments_department_name_key", "employees_department_id_fkey"}, nil
	case org.Positions:
		return catalogSpec{"public.positions", "position_id", "position_name", "position", "uq_positions_normalized_name", "positions_position_name_key", "employees_position_id_fkey"}, nil
	case org.JobLevels:
		return catalogSpec{"public.job_levels", "job_level_id", "job_level_name", "job_level", "uq_job_levels_normalized_name", "", "employees_job_level_id_fkey"}, nil
	default:
		return catalogSpec{}, org.ErrValidation
	}
}
func (r *CatalogRepository) ListCatalog(ctx context.Context, q org.ListCatalogQuery) (result org.CatalogPage, resultErr error) {
	q, err := org.ValidateListCatalog(q)
	if err != nil {
		return org.CatalogPage{}, err
	}
	spec, _ := catalogSpecification(q.Kind)
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer func() {
		// Phân biệt deadline nội bộ với context do caller hủy/hết hạn.
		if errors.Is(resultErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), org.ErrUnavailable) {
			resultErr = errors.Join(org.ErrUnavailable, resultErr)
		}
	}()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return org.CatalogPage{}, catalogError(err, spec)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = checkCatalogActor(ctx, tx, q.ActorID, false); err != nil {
		return org.CatalogPage{}, err
	}
	// Escape cả ký tự escape trước %, _; giá trị luôn là parameter.
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(q.Query) + "%"
	filter := " FROM " + spec.table + " WHERE " + spec.name + ` ILIKE $1 ESCAPE '\'`
	page := org.CatalogPage{Items: []org.CatalogItem{}, Page: q.Page, PageSize: q.PageSize}
	if err = tx.QueryRow(ctx, "SELECT count(*)"+filter, pattern).Scan(&page.Total); err != nil {
		return org.CatalogPage{}, catalogError(err, spec)
	}
	rows, err := tx.Query(ctx, "SELECT "+spec.id+","+spec.name+filter+" ORDER BY "+spec.id+" LIMIT $2 OFFSET $3", pattern, q.PageSize, q.Offset())
	if err != nil {
		return org.CatalogPage{}, catalogError(err, spec)
	}
	for rows.Next() {
		var item org.CatalogItem
		if err = rows.Scan(&item.ID, &item.Name); err != nil {
			rows.Close()
			return org.CatalogPage{}, catalogError(err, spec)
		}
		page.Items = append(page.Items, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return org.CatalogPage{}, catalogError(err, spec)
	}
	if err = tx.Commit(ctx); err != nil {
		return org.CatalogPage{}, catalogError(err, spec)
	}
	return page, nil
}

func (r *CatalogRepository) CreateCatalog(ctx context.Context, c org.CatalogCommand) (org.CatalogItem, error) {
	return r.mutateCatalog(ctx, c, "CREATE")
}
func (r *CatalogRepository) RenameCatalog(ctx context.Context, c org.CatalogCommand) (org.CatalogItem, error) {
	return r.mutateCatalog(ctx, c, "RENAME")
}
func (r *CatalogRepository) DeleteCatalog(ctx context.Context, c org.CatalogCommand) error {
	_, err := r.mutateCatalog(ctx, c, "DELETE")
	return err
}

func (r *CatalogRepository) mutateCatalog(ctx context.Context, c org.CatalogCommand, operation string) (result org.CatalogItem, resultErr error) {
	c, err := org.ValidateCatalogCommand(c, operation != "CREATE", operation != "DELETE")
	if err != nil {
		return org.CatalogItem{}, err
	}
	spec, _ := catalogSpecification(c.Kind)
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, org.ErrUnavailable)
	defer cancel()
	defer func() {
		// Phân biệt deadline nội bộ với context do caller hủy/hết hạn.
		if errors.Is(resultErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), org.ErrUnavailable) {
			resultErr = errors.Join(org.ErrUnavailable, resultErr)
		}
	}()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return org.CatalogItem{}, catalogError(err, spec)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = checkCatalogActor(ctx, tx, c.ActorID, true); err != nil {
		return org.CatalogItem{}, err
	}
	var before, after *org.CatalogItem
	if operation != "CREATE" {
		var current org.CatalogItem
		err = tx.QueryRow(ctx, "SELECT "+spec.id+","+spec.name+" FROM "+spec.table+" WHERE "+spec.id+"=$1 FOR UPDATE", c.ID).Scan(&current.ID, &current.Name)
		if errors.Is(err, pgx.ErrNoRows) {
			return org.CatalogItem{}, org.ErrNotFound
		}
		if err != nil {
			return org.CatalogItem{}, catalogError(err, spec)
		}
		if current.Name != c.ExpectedBeforeName {
			return org.CatalogItem{}, org.ErrStale
		}
		before = &current
		if operation == "RENAME" && c.Name == current.Name {
			if err = tx.Commit(ctx); err != nil {
				return org.CatalogItem{}, catalogCommitError(err, spec)
			}
			return current, nil
		}
	}
	item := org.CatalogItem{ID: c.ID, Name: c.Name}
	switch operation {
	case "CREATE":
		err = tx.QueryRow(ctx, "INSERT INTO "+spec.table+" ("+spec.name+") VALUES ($1) RETURNING "+spec.id+","+spec.name, c.Name).Scan(&item.ID, &item.Name)
		after = &item
	case "RENAME":
		err = tx.QueryRow(ctx, "UPDATE "+spec.table+" SET "+spec.name+"=$1,updated_at=clock_timestamp() WHERE "+spec.id+"=$2 RETURNING "+spec.id+","+spec.name, c.Name, c.ID).Scan(&item.ID, &item.Name)
		after = &item
	case "DELETE":
		// Không khóa employee sau catalog: Admin tạo tài khoản có thể đang giữ employee.
		var used bool
		err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM public.employees WHERE "+spec.id+"=$1)", c.ID).Scan(&used)
		if err == nil && used {
			return org.CatalogItem{}, org.ErrCatalogInUse
		}
		if err == nil {
			_, err = tx.Exec(ctx, "DELETE FROM "+spec.table+" WHERE "+spec.id+"=$1", c.ID)
		}
	}
	if err != nil {
		return org.CatalogItem{}, catalogError(err, spec)
	}
	if err = writeCatalogAudit(ctx, tx, c.ActorID, spec.entity, item.ID, operation, before, after); err != nil {
		return org.CatalogItem{}, catalogError(err, spec)
	}
	if err = tx.Commit(ctx); err != nil {
		return org.CatalogItem{}, catalogCommitError(err, spec)
	}
	return item, nil
}

func checkCatalogActor(ctx context.Context, tx pgx.Tx, userID int, write bool) error {
	lock := ""
	if write {
		lock = " FOR UPDATE"
	}
	var active, limited, employeeActive bool
	err := tx.QueryRow(ctx, `SELECT status='ACTIVE',must_change_password FROM public.users WHERE user_id=$1`+lock, userID).Scan(&active, &limited)
	if errors.Is(err, pgx.ErrNoRows) {
		return org.ErrUnauthenticated
	}
	if err != nil {
		return catalogError(err, catalogSpec{})
	}
	err = tx.QueryRow(ctx, `SELECT employment_status='ACTIVE' FROM public.employees WHERE user_id=$1`+lock, userID).Scan(&employeeActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return org.ErrUnauthenticated
	}
	if err != nil {
		return catalogError(err, catalogSpec{})
	}
	if !active || !employeeActive {
		return org.ErrUnauthenticated
	}
	if limited {
		return org.ErrPasswordChangeRequired
	}
	var allowed bool
	query := `SELECT EXISTS(SELECT 1 FROM public.user_roles WHERE user_id=$1 AND role_code='HR')`
	if !write {
		query = `SELECT EXISTS(SELECT 1 FROM public.user_roles WHERE user_id=$1 AND role_code IN ('HR','ADMIN'))`
	}
	if err = tx.QueryRow(ctx, query, userID).Scan(&allowed); err != nil {
		return catalogError(err, catalogSpec{})
	}
	if !allowed {
		return org.ErrForbidden
	}
	return nil
}
func writeCatalogAudit(ctx context.Context, tx pgx.Tx, actor int, entity string, id int, op string, before, after *org.CatalogItem) error {
	payload := func(item *org.CatalogItem) []byte {
		if item == nil {
			return nil
		}
		// Cấu trúc đóng, không serialize actor/principal hay entity DB thô.
		b, _ := json.Marshal(struct {
			Version int    `json:"schema_version"`
			ID      int    `json:"id"`
			Name    string `json:"name"`
		}{1, item.ID, item.Name})
		return b
	}
	_, err := tx.Exec(ctx, `INSERT INTO public.audit_events(actor_id,entity_type,entity_id,operation,before_values,after_values) VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb)`, actor, entity, id, op, payload(before), payload(after))
	return err
}
func catalogCommitError(err error, spec catalogSpec) error {
	return errors.Join(org.ErrCommitUnknown, catalogError(err, spec))
}
func catalogError(err error, spec catalogSpec) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == spec.unique || (spec.legacyUnique != "" && pgErr.ConstraintName == spec.legacyUnique) {
				return org.ErrDuplicateName
			}
		case "23503":
			if spec.fk != "" && pgErr.ConstraintName == spec.fk {
				return org.ErrCatalogInUse
			}
		case "55P03", "40P01", "57014":
			return org.ErrUnavailable
		}
	}
	switch e := authPersistenceError(err); e {
	case context.Canceled, context.DeadlineExceeded:
		return e
	case auth.ErrUnavailable:
		return org.ErrUnavailable
	default:
		return org.ErrPersistence
	}
}
