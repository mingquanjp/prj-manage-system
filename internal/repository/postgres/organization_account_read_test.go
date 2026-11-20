package postgres

import (
	"context"
	"errors"
	"testing"

	org "ai-staffing/internal/service/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type previewErrorReader struct{ err error }

func (r previewErrorReader) QueryRow(context.Context, string, ...any) pgx.Row {
	return previewErrorRow{r.err}
}

type previewErrorRow struct{ err error }

func (r previewErrorRow) Scan(...any) error { return r.err }

func TestPreviewPreservesQueryFailureClassification(t *testing.T) {
	for _, tc := range []struct{ err, want error }{
		{&pgconn.PgError{Code: "08006", Message: "private error"}, org.ErrUnavailable},
		{&pgconn.PgError{Code: "42501", Message: "private error"}, org.ErrPersistence},
		{context.DeadlineExceeded, context.DeadlineExceeded},
	} {
		reader := &previewSequenceReader{tx: previewErrorReader{tc.err}}
		_, err := PreviewEmployeeCode(t.Context(), reader)
		if err == nil || !errors.Is(accountError(reader.err), tc.want) {
			t.Fatal("preview làm mất phân loại lỗi nguồn")
		}
	}
}
