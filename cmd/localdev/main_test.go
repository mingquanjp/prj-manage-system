package main

import "testing"

func TestRootConfigOnlyAcceptsOwnedLocalEndpoint(t *testing.T) {
	good := "postgresql://ai_staffing_test:synthetic@127.0.0.1:55438/ai_staffing_auth_test?sslmode=disable"
	if _, err := rootConfig(good, "synthetic-cluster"); err != nil {
		t.Fatal("đích local hợp lệ bị từ chối")
	}
	for _, tc := range []struct{ name, raw, system string }{
		{"thiếu URL", "", "cluster"},
		{"thiếu ID", good, ""},
		{"remote", "postgresql://user@remote.test/db?sslmode=disable", "cluster"},
		{"khác DB", "postgresql://ai_staffing_test@127.0.0.1:55438/development?sslmode=disable", "cluster"},
		{"khác port", "postgresql://ai_staffing_test@127.0.0.1:5432/ai_staffing_auth_test?sslmode=disable", "cluster"},
		{"khác user", "postgresql://postgres@127.0.0.1:55438/ai_staffing_auth_test?sslmode=disable", "cluster"},
		{"thêm option", good + "&options=x", "cluster"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := rootConfig(tc.raw, tc.system); err == nil {
				t.Fatal("phải từ chối đích không đúng guard")
			}
		})
	}
}
