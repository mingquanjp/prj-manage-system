package gmail

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCallbackRejectsInvalidStateAndDoesNotEchoCode(t *testing.T) {
	const base = "http://127.0.0.1:43210/"
	codes := make(chan string, 1)
	handler := callbackHandler("127.0.0.1:43210", "expected", codes)
	for _, query := range []string{"?state=wrong&code=private", "?state=expected&state=wrong&code=private"} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest("GET", base+query, nil))
		if rec.Code != 400 || len(codes) != 0 || strings.Contains(rec.Body.String(), "private") {
			t.Fatal("Callback chấp nhận state sai hoặc lộ code")
		}
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", base+"?state=expected&code=private", nil))
	if rec.Code != 200 || len(codes) != 1 || <-codes != "private" || strings.Contains(rec.Body.String(), "private") {
		t.Fatal("Callback hợp lệ không chuyển code an toàn")
	}
}

func TestCredentialFilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private", "gmail.json")
	value := credentials{"client", "secret", "refresh", "sender@example.invalid"}
	if err := writeCredentials(path, value); err != nil {
		t.Fatal(err)
	}
	s, err := NewFromFile(path)
	if err != nil || s == nil {
		t.Fatal("Không đọc được file hợp lệ")
	}
	st, err := os.Stat(path)
	if err != nil || st.Mode().Perm() != 0600 {
		t.Fatal("Quyền credential file không đúng")
	}
	link := filepath.Join(filepath.Dir(path), "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFromFile(link); err != ErrConfig {
		t.Fatal("Đã đọc credential thông qua symlink")
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFromFile(path); err != ErrConfig {
		t.Fatal("Đã đọc credential có quyền quá rộng")
	}
}
