package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/service/organization"
	sc "ai-staffing/internal/service/skillcatalog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SkillCatalogRepository đọc danh mục và ghi thay đổi cùng audit trong PostgreSQL.
type SkillCatalogRepository struct {
	db      catalogDB
	timeout time.Duration
}

var _ sc.Repository = (*SkillCatalogRepository)(nil)

// NewSkillCatalogRepository giới hạn cả thời gian chờ khóa trong timeout được cấp.
func NewSkillCatalogRepository(db catalogDB, timeout time.Duration) (*SkillCatalogRepository, error) {
	if db == nil || timeout <= 0 {
		return nil, sc.ErrValidation
	}
	return &SkillCatalogRepository{db, timeout}, nil
}
func skillCatalogError(err error) error {
	if err == nil {
		return nil
	}
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		switch pe.Code {
		case "23505":
			switch pe.ConstraintName {
			case "uq_msk_category_name", "uq_msk_skill_name", "skill_categories_category_name_key", "skills_skill_name_key":
				return sc.ErrDuplicateName
			}
		case "23503":
			switch pe.ConstraintName {
			case "skills_category_id_fkey", "employee_skills_skill_id_fkey", "employee_skill_change_requests_skill_id_fkey", "task_requirements_skill_id_fkey", "cv_extracted_skills_skill_id_fkey":
				return sc.ErrCatalogInUse
			}
		case "55P03", "40P01", "57014":
			return sc.ErrUnavailable
		}
	}
	switch e := authPersistenceError(err); e {
	case context.Canceled, context.DeadlineExceeded:
		return e
	case auth.ErrUnavailable:
		return sc.ErrUnavailable
	default:
		return sc.ErrPersistence
	}
}

func skillCatalogCommitError(err error) error {
	mapped := skillCatalogError(err)
	if errors.Is(accountCommitError(err), organization.ErrCommitUnknown) {
		return errors.Join(sc.ErrCommitUnknown, mapped)
	}
	return mapped
}
func skillCatalogDeadline(ctx context.Context, err *error) {
	if errors.Is(*err, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), sc.ErrUnavailable) {
		*err = errors.Join(sc.ErrUnavailable, *err)
	}
}
func skillCatalogActor(ctx context.Context, tx pgx.Tx, id int, write bool) error {
	lock := ""
	if write {
		lock = " FOR UPDATE"
	}
	var active, limited, employeeActive bool
	err := tx.QueryRow(ctx, `SELECT status='ACTIVE',must_change_password FROM public.users WHERE user_id=$1`+lock, id).Scan(&active, &limited)
	if errors.Is(err, pgx.ErrNoRows) {
		return sc.ErrUnauthenticated
	}
	if err != nil {
		return skillCatalogError(err)
	}
	err = tx.QueryRow(ctx, `SELECT employment_status='ACTIVE' FROM public.employees WHERE user_id=$1`+lock, id).Scan(&employeeActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return sc.ErrUnauthenticated
	}
	if err != nil {
		return skillCatalogError(err)
	}
	if !active || !employeeActive {
		return sc.ErrUnauthenticated
	}
	if limited {
		return sc.ErrPasswordChangeRequired
	}
	if write {
		var director bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.user_roles WHERE user_id=$1 AND role_code='DIRECTOR')`, id).Scan(&director)
		if err != nil {
			return skillCatalogError(err)
		}
		if !director {
			return sc.ErrForbidden
		}
	}
	return nil
}

// List đọc actor, tổng và danh sách trong một transaction REPEATABLE READ.
func (r *SkillCatalogRepository) List(ctx context.Context, q sc.ListQuery) (result sc.Page, resultErr error) {
	q, err := sc.ValidateList(q)
	if err != nil {
		return sc.Page{}, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, sc.ErrUnavailable)
	defer cancel()
	defer skillCatalogDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return sc.Page{}, skillCatalogError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = skillCatalogActor(ctx, tx, q.ActorID, false); err != nil {
		return sc.Page{}, err
	}
	p := sc.Page{Items: []sc.Item{}, Page: q.Page, PageSize: q.PageSize}
	if q.Kind == sc.Categories {
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM public.skill_categories WHERE strpos(lower(category_name),lower($1))>0`, q.Query).Scan(&p.Total); err != nil {
			return sc.Page{}, skillCatalogError(err)
		}
		rows, e := tx.Query(ctx, `SELECT category_id,category_name FROM public.skill_categories WHERE strpos(lower(category_name),lower($1))>0 ORDER BY category_name,category_id LIMIT $2 OFFSET $3`, q.Query, q.PageSize, q.Offset())
		if e != nil {
			return sc.Page{}, skillCatalogError(e)
		}
		for rows.Next() {
			var item sc.Item
			if e = rows.Scan(&item.ID, &item.Name); e != nil {
				rows.Close()
				return sc.Page{}, skillCatalogError(e)
			}
			p.Items = append(p.Items, item)
		}
		rows.Close()
		if e = rows.Err(); e != nil {
			return sc.Page{}, skillCatalogError(e)
		}
	} else {
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM public.skills s WHERE strpos(lower(s.skill_name),lower($1))>0 AND ($2=0 OR s.category_id=$2)`, q.Query, q.CategoryID).Scan(&p.Total); err != nil {
			return sc.Page{}, skillCatalogError(err)
		}
		rows, e := tx.Query(ctx, `SELECT s.skill_id,s.skill_name,c.category_id,c.category_name FROM public.skills s JOIN public.skill_categories c ON c.category_id=s.category_id WHERE strpos(lower(s.skill_name),lower($1))>0 AND ($2=0 OR s.category_id=$2) ORDER BY s.skill_name,s.skill_id LIMIT $3 OFFSET $4`, q.Query, q.CategoryID, q.PageSize, q.Offset())
		if e != nil {
			return sc.Page{}, skillCatalogError(e)
		}
		for rows.Next() {
			var item sc.Item
			if e = rows.Scan(&item.ID, &item.Name, &item.Category.ID, &item.Category.Name); e != nil {
				rows.Close()
				return sc.Page{}, skillCatalogError(e)
			}
			p.Items = append(p.Items, item)
		}
		rows.Close()
		if e = rows.Err(); e != nil {
			return sc.Page{}, skillCatalogError(e)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return sc.Page{}, skillCatalogError(err)
	}
	return p, nil
}

// Create ghi danh mục và audit cùng transaction sau kiểm quyền Giám đốc.
func (r *SkillCatalogRepository) Create(ctx context.Context, c sc.Command) (sc.Item, error) {
	return r.mutate(ctx, c, "CREATE")
}

// Update khóa nhóm trước kỹ năng và từ chối before cũ thay vì ghi đè.
func (r *SkillCatalogRepository) Update(ctx context.Context, c sc.Command) (sc.Item, error) {
	return r.mutate(ctx, c, "UPDATE")
}

// Delete giữ các tham chiếu hiện có; FK xung đột không bị chuyển thành cascade.
func (r *SkillCatalogRepository) Delete(ctx context.Context, c sc.Command) error {
	_, err := r.mutate(ctx, c, "DELETE")
	return err
}
func (r *SkillCatalogRepository) mutate(ctx context.Context, c sc.Command, operation string) (result sc.Item, resultErr error) {
	c, err := sc.ValidateCommand(c, operation != "CREATE", operation != "DELETE")
	if err != nil {
		return sc.Item{}, err
	}
	ctx, cancel := context.WithTimeoutCause(ctx, r.timeout, sc.ErrUnavailable)
	defer cancel()
	defer skillCatalogDeadline(ctx, &resultErr)
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return sc.Item{}, skillCatalogError(err)
	}
	defer rollbackAuthTransaction(ctx, tx)
	if err = skillCatalogActor(ctx, tx, c.ActorID, true); err != nil {
		return sc.Item{}, err
	}
	var before, after *sc.Item
	item := sc.Item{ID: c.ID, Name: c.Name}
	if c.Kind == sc.Categories {
		if operation != "CREATE" {
			var current sc.Item
			err = tx.QueryRow(ctx, `SELECT category_id,category_name FROM public.skill_categories WHERE category_id=$1 FOR UPDATE`, c.ID).Scan(&current.ID, &current.Name)
			if errors.Is(err, pgx.ErrNoRows) {
				return sc.Item{}, sc.ErrNotFound
			}
			if err != nil {
				return sc.Item{}, skillCatalogError(err)
			}
			if current.Name != c.ExpectedBefore.Name {
				return sc.Item{}, sc.ErrStale
			}
			before = &current
		}
	} else {
		oldCategory := 0
		if operation != "CREATE" {
			err = tx.QueryRow(ctx, `SELECT category_id FROM public.skills WHERE skill_id=$1`, c.ID).Scan(&oldCategory)
			if errors.Is(err, pgx.ErrNoRows) {
				return sc.Item{}, sc.ErrNotFound
			}
			if err != nil {
				return sc.Item{}, skillCatalogError(err)
			}
		}
		ids := []int{}
		if oldCategory != 0 {
			ids = append(ids, oldCategory)
		}
		if operation != "DELETE" && c.CategoryID != oldCategory {
			ids = append(ids, c.CategoryID)
		}
		sort.Ints(ids)
		categories := map[int]string{}
		// Nhóm cũ/mới được khóa trước skill; nếu liên kết đổi khi chờ thì trả stale, không đảo thứ tự khóa.
		for _, id := range ids {
			var name string
			err = tx.QueryRow(ctx, `SELECT category_name FROM public.skill_categories WHERE category_id=$1 FOR KEY SHARE`, id).Scan(&name)
			if errors.Is(err, pgx.ErrNoRows) {
				if id == oldCategory {
					var currentCategory int
					checkErr := tx.QueryRow(ctx, `SELECT category_id FROM public.skills WHERE skill_id=$1`, c.ID).Scan(&currentCategory)
					if errors.Is(checkErr, pgx.ErrNoRows) {
						return sc.Item{}, sc.ErrNotFound
					}
					if checkErr != nil {
						return sc.Item{}, skillCatalogError(checkErr)
					}
					return sc.Item{}, sc.ErrStale
				}
				return sc.Item{}, sc.ErrNotFound
			}
			if err != nil {
				return sc.Item{}, skillCatalogError(err)
			}
			categories[id] = name
		}
		item.Category = sc.Category{ID: c.CategoryID, Name: categories[c.CategoryID]}
		if operation != "CREATE" {
			var current sc.Item
			err = tx.QueryRow(ctx, `SELECT skill_id,skill_name,category_id FROM public.skills WHERE skill_id=$1 FOR UPDATE`, c.ID).Scan(&current.ID, &current.Name, &current.Category.ID)
			if errors.Is(err, pgx.ErrNoRows) {
				return sc.Item{}, sc.ErrNotFound
			}
			if err != nil {
				return sc.Item{}, skillCatalogError(err)
			}
			if current.Category.ID != oldCategory || current.Name != c.ExpectedBefore.Name || current.Category.ID != c.ExpectedBefore.CategoryID {
				return sc.Item{}, sc.ErrStale
			}
			current.Category.Name = categories[current.Category.ID]
			before = &current
		}
	}
	if operation == "UPDATE" && before.Name == item.Name && before.Category.ID == item.Category.ID {
		if err = tx.Commit(ctx); err != nil {
			return sc.Item{}, skillCatalogCommitError(err)
		}
		return *before, nil
	}
	if c.Kind == sc.Categories {
		switch operation {
		case "CREATE":
			err = tx.QueryRow(ctx, `INSERT INTO public.skill_categories(category_name) VALUES ($1) RETURNING category_id,category_name`, c.Name).Scan(&item.ID, &item.Name)
		case "UPDATE":
			err = tx.QueryRow(ctx, `UPDATE public.skill_categories SET category_name=$1 WHERE category_id=$2 RETURNING category_id,category_name`, c.Name, c.ID).Scan(&item.ID, &item.Name)
		case "DELETE":
			_, err = tx.Exec(ctx, `DELETE FROM public.skill_categories WHERE category_id=$1`, c.ID)
		}
	} else {
		switch operation {
		case "CREATE":
			err = tx.QueryRow(ctx, `INSERT INTO public.skills(skill_name,category_id) VALUES ($1,$2) RETURNING skill_id,skill_name`, c.Name, c.CategoryID).Scan(&item.ID, &item.Name)
		case "UPDATE":
			err = tx.QueryRow(ctx, `UPDATE public.skills SET skill_name=$1,category_id=$2 WHERE skill_id=$3 RETURNING skill_id,skill_name`, c.Name, c.CategoryID, c.ID).Scan(&item.ID, &item.Name)
		case "DELETE":
			_, err = tx.Exec(ctx, `DELETE FROM public.skills WHERE skill_id=$1`, c.ID)
		}
	}
	if err != nil {
		return sc.Item{}, skillCatalogError(err)
	}
	if operation != "DELETE" {
		after = &item
	}
	if err = writeSkillCatalogAudit(ctx, tx, c.ActorID, c.Kind, item.ID, operation, before, after); err != nil {
		return sc.Item{}, skillCatalogError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return sc.Item{}, skillCatalogCommitError(err)
	}
	return item, nil
}
func writeSkillCatalogAudit(ctx context.Context, tx pgx.Tx, actor int, kind sc.Kind, id int, operation string, before, after *sc.Item) error {
	payload := func(v *sc.Item) []byte {
		if v == nil {
			return nil
		}
		if kind == sc.Categories {
			b, _ := json.Marshal(struct {
				Version int    `json:"schema_version"`
				ID      int    `json:"id"`
				Name    string `json:"name"`
			}{1, v.ID, v.Name})
			return b
		}
		b, _ := json.Marshal(struct {
			Version    int    `json:"schema_version"`
			ID         int    `json:"id"`
			Name       string `json:"name"`
			CategoryID int    `json:"category_id"`
		}{1, v.ID, v.Name, v.Category.ID})
		return b
	}
	entity := "SKILL_CATEGORY"
	if kind == sc.Skills {
		entity = "SKILL"
	}
	tag, err := tx.Exec(ctx, `INSERT INTO public.audit_events(actor_id,entity_type,entity_id,operation,before_values,after_values) VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb)`, actor, entity, id, operation, payload(before), payload(after))
	if err == nil && tag.RowsAffected() != 1 {
		return sc.ErrPersistence
	}
	return err
}
