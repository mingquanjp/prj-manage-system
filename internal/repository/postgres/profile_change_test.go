package postgres

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	pc "ai-staffing/internal/service/profilechange"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestProfileChangeRepositoryClassificationsAndTimeout(t *testing.T) {
	for _, tc := range []struct {
		input, want error
		unknown     bool
	}{
		{io.ErrUnexpectedEOF, pc.ErrUnavailable, true},
		{&pgconn.PgError{Code: "23514", Severity: "ERROR"}, pc.ErrPersistence, false},
		{&pgconn.PgError{Code: "08006", Severity: "FATAL"}, pc.ErrUnavailable, true},
		{&pgconn.PgError{Code: "40003", Severity: "ERROR"}, pc.ErrPersistence, true},
		{&pgconn.PgError{Code: "40P01", Severity: "ERROR"}, pc.ErrUnavailable, false},
		{pgx.ErrTxCommitRollback, pc.ErrPersistence, false},
	} {
		err := profileChangeCommitError(tc.input)
		if !errors.Is(err, tc.want) || errors.Is(err, pc.ErrCommitUnknown) != tc.unknown {
			t.Fatal("commit outcome map sai")
		}
	}
	if _, e := NewProfileChangeRepository(catalogDeadlineDB{}, 0); !errors.Is(e, pc.ErrValidation) {
		t.Fatal("timeout không hợp lệ")
	}
	repo, e := NewProfileChangeRepository(catalogDeadlineDB{}, time.Millisecond)
	if e != nil {
		t.Fatal("constructor")
	}
	for _, read := range []bool{false, true} {
		var err error
		if read {
			_, err = repo.Read(t.Context(), pc.DetailQuery{ActorID: 1, RequestID: 1})
		} else {
			_, err = repo.List(t.Context(), pc.ListQuery{ActorID: 1})
		}
		if !errors.Is(err, pc.ErrUnavailable) || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("deadline nội bộ phải map unavailable")
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := repo.List(ctx, pc.ListQuery{ActorID: 1})
	if !errors.Is(err, context.Canceled) || errors.Is(err, pc.ErrUnavailable) {
		t.Fatal("caller cancel bị đổi nguồn")
	}
	for _, b := range []string{`{}`, `null`, `{"schema_version":1,"request_id":1,"result_code":"OTHER"}`, `{"schema_version":1,"request_id":1,"result_code":"CREATED","extra":true}`} {
		if _, err := decodeProfileReference([]byte(b)); !errors.Is(err, pc.ErrPersistence) {
			t.Fatal("reference hỏng được chấp nhận")
		}
	}
}
