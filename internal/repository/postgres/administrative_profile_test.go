package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestAdministrativeProfileRepositoryTimeoutAndMapping(t *testing.T) {
	if _, e := NewAdministrativeProfileRepository(catalogDeadlineDB{}, 0); !errors.Is(e, org.ErrValidation) {
		t.Fatal("timeout không hợp lệ")
	}
	r, e := NewAdministrativeProfileRepository(catalogDeadlineDB{}, time.Millisecond)
	if e != nil {
		t.Fatal(e)
	}
	for _, call := range []func() error{
		func() error {
			_, e := r.ReadAdministrativeProfile(t.Context(), org.AdministrativeProfileQuery{ActorID: 1, EmployeeID: 2})
			return e
		},
		func() error {
			_, e := r.ListAdministrativeEmployees(t.Context(), org.ListAdministrativeEmployeesQuery{ActorID: 1, Page: 1, PageSize: 20})
			return e
		},
		func() error {
			value := "x"
			_, e := r.EditAdministrativeProfile(t.Context(), org.EditAdministrativeProfileCommand{ActorID: 1, EmployeeID: 2, ExpectedBeforeValues: org.AdministrativeValues{Phone: org.AdministrativeField[string]{Present: true}}, Changes: org.AdministrativeValues{Phone: org.AdministrativeField[string]{Present: true, Value: &value}}})
			return e
		},
	} {
		if e := call(); !errors.Is(e, org.ErrUnavailable) || !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("deadline nội bộ sai nguồn")
		}
	}
	for _, tc := range []struct{ input, errorKind error }{
		{&pgconn.PgError{Code: "23503", ConstraintName: "employees_position_id_fkey"}, org.ErrCatalogNotFound},
		{&pgconn.PgError{Code: "23503", ConstraintName: "audit_events_actor_id_fkey"}, org.ErrPersistence},
		{&pgconn.PgError{Code: "23505", ConstraintName: "uq_employees_normalized_email"}, org.ErrEmailConflict},
		{&pgconn.PgError{Code: "40P01"}, org.ErrUnavailable},
	} {
		if !errors.Is(administrativeError(tc.input), tc.errorKind) {
			t.Fatal("mapping lỗi DB sai")
		}
	}
}
