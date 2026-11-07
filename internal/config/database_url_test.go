package config

import (
	"errors"
	"strings"
	"testing"
)

func TestParseDatabaseURL(t *testing.T) {
	const pooled = "ep-demo-123-pooler.ap-southeast-1.aws.neon.tech"
	const direct = "ep-demo-123.ap-southeast-1.aws.neon.tech"
	for _, tt := range []struct {
		name, raw, want string
		pooled          bool
	}{
		{"postgres local", "postgres://user:fake@localhost:5432/demo", "", false},
		{"postgresql IPv4", "postgresql://127.0.0.1/demo", "", false},
		{"IPv6", "postgres://[::1]:5432/demo", "", false},
		{"direct", "postgres://user:fake@" + direct + "/demo?sslmode=require", "", false},
		{"pooled", "postgres://user:fake@" + pooled + "/demo?sslmode=require", "postgres://user:fake@" + direct + "/demo?sslmode=require", true},
		{"port and encoded credentials", "postgresql://us%65r:p%40ss%2f%2F@" + pooled + ":5432/db%2Fname?x=a%2fb&x=c&sslmode=require", "postgresql://us%65r:p%40ss%2f%2F@" + direct + ":5432/db%2Fname?x=a%2fb&x=c&sslmode=require", true},
		{"markers everywhere", "postgres://us-pooler.er:pass-pooler.word@" + pooled + "/db-pooler.name?application_name=test-pooler.marker", "postgres://us-pooler.er:pass-pooler.word@" + direct + "/db-pooler.name?application_name=test-pooler.marker", true},
		{"password only", "postgres://user:p-pooler.word@localhost/demo", "", false},
		{"path only", "postgres://localhost/db-pooler.name", "", false},
		{"query only", "postgres://localhost/demo?application_name=x-pooler.name", "", false},
		{"custom pooler", "postgres://custom-pooler.example.test/demo", "", false},
		{"suffix spoof", "postgres://" + pooled + ".attacker.example/demo", "", false},
		{"missing domain boundary", "postgres://ep-demo-pooler.region.aws.evilneon.tech/demo", "", false},
		{"not endpoint label", "postgres://other-pooler.region.aws.neon.tech/demo", "", false},
		{"empty endpoint", "postgres://ep--pooler.region.aws.neon.tech/demo", "", false},
		{"upper case host", "postgres://EP-DEMO-POOLER.REGION.AWS.NEON.TECH:5432/demo", "postgres://EP-DEMO.REGION.AWS.NEON.TECH:5432/demo", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDatabaseURL(tt.raw)
			if err != nil {
				t.Fatal(err)
			}
			want := tt.want
			if want == "" {
				want = tt.raw
			}
			if got.RuntimeURL != tt.raw || got.DirectURL != want || got.Pooled != tt.pooled {
				t.Fatal("endpoint hoặc phần không liên quan bị thay đổi")
			}
		})
	}
}

func TestParseDatabaseURLRejectsUnsupported(t *testing.T) {
	for _, raw := range []string{
		"", " ", "host=localhost user=SECRET_MARKER", "mysql://localhost/db", "postgres:opaque", "postgres:///db",
		"postgres://SECRET_MARKER%zz@localhost/db", "postgres://localhost/db#SECRET_MARKER", "postgres://localhost/db#",
		"postgres://a,b/db", "postgres://a:5432,b:5432/db", "postgres://[not-ip]/db", "postgres://::1/db", "postgres://[fe80::1%25en0]/db",
		"postgres://localhost:/db", "postgres://localhost:0/db", "postgres://localhost:65536/db", "postgres://localhost:abc/db",
		"postgres://bad_host/db", "postgres://localhost/db?host=/tmp", "postgres://localhost/db?%68ost=evil.test",
		"postgres://localhost/db?hostaddr=127.0.0.1", "postgres://localhost/db?port=5433", "postgres://localhost/db?user=other",
		"postgres://localhost/db?password=SECRET_MARKER", "postgres://localhost/db?dbname=other", "postgres://localhost/db?service=x",
		"postgres://localhost/db?servicefile=/tmp/SECRET_MARKER", "postgres://localhost/db?sslmode=%zz",
	} {
		t.Run(raw, func(t *testing.T) {
			got, err := ParseDatabaseURL(raw)
			if !errors.Is(err, ErrDatabaseURLInvalid) || got != (DatabaseEndpoint{}) {
				t.Fatal("phải từ chối và không trả endpoint một phần")
			}
			if strings.Contains(err.Error(), "SECRET_MARKER") {
				t.Fatal("lộ giá trị đầu vào")
			}
		})
	}
}
