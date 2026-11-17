package postgres

import (
	"math"
	"testing"
)

func TestFormatEmployeeCode(t *testing.T) {
	for _, tc := range []struct {
		n    int64
		want string
	}{
		{1, "EMP000001"}, {999999, "EMP999999"}, {1000000, "EMP1000000"}, {math.MaxInt64, "EMP9223372036854775807"},
	} {
		got, err := FormatEmployeeCode(tc.n)
		if err != nil || got != tc.want {
			t.Fatalf("format %d sai", tc.n)
		}
	}
	for _, n := range []int64{0, -1, math.MinInt64} {
		if _, err := FormatEmployeeCode(n); err == nil {
			t.Fatal("phải từ chối số không dương")
		}
	}
}
