package postgres

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type catalogDeadlineDB struct{}

func (catalogDeadlineDB) BeginTx(ctx context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
func TestCatalogRepositoryTimeout(t *testing.T) {
	r, err := NewCatalogRepository(catalogDeadlineDB{}, time.Millisecond)
	if err != nil {
		t.Fatal("constructor")
	}
	for _, write := range []bool{false, true} {
		var err error
		if write {
			_, err = r.CreateCatalog(t.Context(), org.CatalogCommand{ActorID: 1, Kind: org.Departments, Name: "name"})
		} else {
			_, err = r.ListCatalog(t.Context(), org.ListCatalogQuery{ActorID: 1, Kind: org.Departments, Page: 1, PageSize: 20})
		}
		if !errors.Is(err, org.ErrUnavailable) || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("timeout nội bộ phải là lỗi tạm thời")
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = r.CreateCatalog(ctx, org.CatalogCommand{ActorID: 1, Kind: org.Departments, Name: "name"})
	if !errors.Is(err, context.Canceled) || errors.Is(err, org.ErrUnavailable) {
		t.Fatal("caller cancel phải giữ nguyên")
	}
	if _, err = NewCatalogRepository(catalogDeadlineDB{}, 0); !errors.Is(err, org.ErrValidation) {
		t.Fatal("timeout phải hữu hạn và dương")
	}
}
func TestCatalogPersistenceClassification(t *testing.T) {
	spec, _ := catalogSpecification(org.Departments)
	for _, tc := range []struct{ err, want error }{
		{&pgconn.PgError{Code: "23505", ConstraintName: spec.unique}, org.ErrDuplicateName},
		{&pgconn.PgError{Code: "23505", ConstraintName: "other"}, org.ErrPersistence},
		{&pgconn.PgError{Code: "23503", ConstraintName: spec.fk}, org.ErrCatalogInUse},
		{&pgconn.PgError{Code: "23503", ConstraintName: "audit_events_actor_id_fkey"}, org.ErrPersistence},
		{&pgconn.PgError{Code: "55P03"}, org.ErrUnavailable},
		{&pgconn.PgError{Code: "40P01"}, org.ErrUnavailable},
		{&pgconn.PgError{Code: "57014"}, org.ErrUnavailable},
		{io.ErrUnexpectedEOF, org.ErrUnavailable},
	} {
		if !errors.Is(catalogError(tc.err, spec), tc.want) {
			t.Fatal("SQLSTATE/constraint map sai")
		}
	}
}
