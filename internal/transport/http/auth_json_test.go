package http

import "testing"

func TestDecodeAuthStringPairUnicode(t *testing.T) {
	for _, tc := range []struct {
		name, encoded, want string
		valid               bool
	}{
		{"cặp hợp lệ", `"\uD83D\uDE00"`, "😀", true},
		{"high đơn", `"\uD83D"`, "", false},
		{"low đơn", `"\uDE00"`, "", false},
		{"high nối ký tự thường", `"\uD83D\u0061"`, "", false},
		{"high nối high", `"\uD83D\uD83D"`, "", false},
		{"U+FFFD hợp lệ", `"�\uFFFD"`, "��", true},
		{"backslash nguyên văn", `"\\uD83D"`, "\\uD83D", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"email":"an@example.test","password":` + tc.encoded + "}")
			fields, valid := decodeAuthStringPair(body, "email", "password")
			if valid != tc.valid || fields["password"] != tc.want {
				t.Fatalf("decode = %q, %v; muốn %q, %v", fields["password"], valid, tc.want, tc.valid)
			}
		})
	}
}
