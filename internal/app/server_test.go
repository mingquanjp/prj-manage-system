package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func startBodyDeadlineServer(t *testing.T, handler http.Handler) (*http.Server, string) {
	t.Helper()
	server := NewHTTPServer("127.0.0.1:0", handler)
	if server.ReadTimeout <= 0 || server.ReadTimeout >= server.WriteTimeout {
		t.Fatal("hạn đọc phải hữu hạn và chừa thời gian xử lý/phản hồi")
	}
	// Rút ngắn cùng cấu hình production để kiểm TCP mà không chờ đủ 15 giây mỗi ca.
	server.ReadTimeout /= 30
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		_ = server.Close()
		select {
		case err := <-done:
			if !errors.Is(err, http.ErrServerClosed) {
				t.Error("server test không dừng đúng cách")
			}
		case <-time.After(3 * time.Second):
			t.Error("server test chưa giải phóng listener")
		}
	})
	return server, listener.Addr().String()
}

func TestHTTPServerStopsIncompleteBody(t *testing.T) {
	for _, tc := range []struct{ name, framing string }{
		{"content length", "Content-Length: 2\r\n\r\n{"},
		{"chunked", "Transfer-Encoding: chunked\r\n\r\n2\r\n{"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			readResult := make(chan error, 1)
			_, address := startBodyDeadlineServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := io.Copy(io.Discard, r.Body)
				readResult <- err
				w.WriteHeader(http.StatusBadRequest)
			}))
			conn, err := net.DialTimeout("tcp", address, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if _, err = fmt.Fprintf(conn, "POST / HTTP/1.1\r\nHost: %s\r\n%s", address, tc.framing); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-readResult:
				var timeout net.Error
				if !errors.As(err, &timeout) || !timeout.Timeout() {
					t.Fatal("body dang dở phải kết thúc do hạn đọc")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("handler vẫn chờ body dang dở")
			}
			if err = conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			response, err := http.ReadResponse(bufio.NewReader(conn), nil)
			if err != nil {
				t.Fatal("handler phải kết thúc phản hồi sau lỗi đọc")
			}
			defer response.Body.Close()
			if _, err = io.Copy(io.Discard, response.Body); err != nil {
				t.Fatal("phản hồi sau hạn đọc chưa kết thúc")
			}
		})
	}
}

func TestHTTPServerAllowsProcessingAfterBodyRead(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	handlerContext := make(chan context.Context, 1)
	server, address := startBodyDeadlineServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handlerContext <- r.Context()
		close(started)
		select {
		case <-release:
			w.WriteHeader(http.StatusNoContent)
		case <-r.Context().Done():
		}
	}))
	result := make(chan error, 1)
	go func() {
		client := http.Client{Timeout: 3 * time.Second}
		response, err := client.Post("http://"+address, "application/json", strings.NewReader("{}"))
		if err == nil {
			defer response.Body.Close()
			if response.StatusCode != http.StatusNoContent {
				err = errors.New("body đầy đủ không được bị coi là hết hạn đọc")
			}
		}
		result <- err
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("body đầy đủ chưa tới bước xử lý")
	}
	ctx := <-handlerContext
	select {
	case <-time.After(2 * server.ReadTimeout):
	case <-ctx.Done():
		t.Fatal("hạn đọc không được hủy xử lý sau khi nhận đủ body")
	}
	close(release)
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}
