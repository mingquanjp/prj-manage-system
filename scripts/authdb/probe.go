// Probe chỉ đọc metadata qua TCP host; được manage.py gọi sau guard sở hữu Docker.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ai-staffing/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	if err := probe(); err != nil {
		fmt.Fprintln(os.Stderr, "AUTH_TEST_DB_PROBE_FAILED: không xác minh được kết nối test")
		os.Exit(1)
	}
}

func probe() error {
	raw := os.Getenv("TEST_DATABASE_URL")
	// Không dùng loader dotenv, không fallback DATABASE_URL.
	endpoint, err := config.ParseDatabaseURL(raw)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, endpoint.RuntimeURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	var result []byte
	err = conn.QueryRow(ctx, `SELECT json_build_object(
 'major', current_setting('server_version_num')::int / 10000,
 'database', current_database(), 'user', current_user,
 'owner', pg_get_userbyid(datdba) = current_user,
 'public_tables', (SELECT count(*) FROM pg_tables WHERE schemaname='public'),
 'system_id', (SELECT system_identifier::text FROM pg_control_system()))
 FROM pg_database WHERE datname=current_database()`).Scan(&result)
	if err != nil {
		return err
	}
	var decoded map[string]any
	if err = json.Unmarshal(result, &decoded); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(decoded)
}
