//go:build integration

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"ai-staffing/internal/testutil/authdb"
	"ai-staffing/migrations"
)

func TestLocaldevPrepareOnM1(t *testing.T) {
	db := authdb.NewCase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	authdb.Require(t, "chuẩn bị schema và fixture localdev", prepare(ctx, db.Conn.Config(), os.Getenv("AUTH_TEST_DB_SYSTEM_ID")))

	var version int
	var dirty bool
	var accounts int
	authdb.Require(t, "kiểm localdev sau prepare", db.Conn.QueryRow(ctx, `SELECT version, dirty,
		(SELECT count(*) FROM public.employees WHERE address IS NULL AND job_level_id IS NULL)
		FROM public.schema_migrations`).Scan(&version, &dirty, &accounts))
	if version != migrations.M1Version || dirty || accounts != 2 {
		t.Fatal("localdev phải có schema M1 sạch và đủ hồ sơ fixture")
	}
}
