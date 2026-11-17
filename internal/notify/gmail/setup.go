package gmail

import (
	"context"
	"crypto/subtle"
	"errors"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

var ErrAuthorization = errors.New("không hoàn tất cấp quyền Gmail")

// Authorize mở callback local trong một lượt cấp quyền cho chủ hộp thư.
func Authorize(ctx context.Context, clientPath, from string, showURL func(string)) error {
	var downloaded struct {
		Installed struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"installed"`
	}
	path := credentialPath()
	if !filepath.IsAbs(path) || !validAddress(from) || readPrivateJSON(clientPath, &downloaded) != nil ||
		!nonblank(downloaded.Installed.ClientID) || !nonblank(downloaded.Installed.ClientSecret) {
		return ErrConfig
	}

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return ErrAuthorization
	}
	defer listener.Close()
	config := oauth2.Config{
		ClientID: downloaded.Installed.ClientID, ClientSecret: downloaded.Installed.ClientSecret,
		Endpoint: oauthEndpoint, Scopes: []string{scope},
		RedirectURL: "http://" + listener.Addr().String(),
	}
	state, verifier := oauth2.GenerateVerifier(), oauth2.GenerateVerifier()
	codes := make(chan string, 1)
	server := &http.Server{
		Handler:           callbackHandler(listener.Addr().String(), state, codes),
		ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second, MaxHeaderBytes: 8 << 10,
	}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	defer server.Close()
	waitCtx, stop := context.WithTimeout(ctx, 5*time.Minute)
	defer stop()
	showURL(config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce,
		oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("login_hint", from)))

	var code string
	select {
	case code = <-codes:
	case <-serverErrors:
		return ErrAuthorization
	case <-waitCtx.Done():
		return ErrAuthorization
	}
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), time.Second)
	_ = server.Shutdown(shutdownCtx)
	cancelShutdown()
	if code == "" {
		return ErrAuthorization
	}

	exchangeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	exchangeCtx = context.WithValue(exchangeCtx, oauth2.HTTPClient, providerClient())
	token, err := config.Exchange(exchangeCtx, code, oauth2.VerifierOption(verifier))
	if err != nil || token.RefreshToken == "" || token.Extra("scope") != scope {
		return ErrAuthorization
	}
	return writeCredentials(path, credentials{config.ClientID, config.ClientSecret, token.RefreshToken, from})
}

func callbackHandler(host, state string, codes chan<- string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		q := r.URL.Query()
		if r.Method != http.MethodGet || r.Host != host || r.URL.Path != "/" ||
			len(q["state"]) != 1 || subtle.ConstantTimeCompare([]byte(q.Get("state")), []byte(state)) != 1 {
			http.Error(w, "Callback không hợp lệ.", http.StatusBadRequest)
			return
		}
		code := q.Get("code")
		if q.Get("error") != "" || len(q["code"]) != 1 {
			code = ""
		}
		// Chỉ nhận một phản hồi có state hợp lệ.
		select {
		case codes <- code:
		default:
		}
		_, _ = w.Write([]byte("Đã nhận phản hồi Google. Xem kết quả trong terminal và đóng tab này."))
		_ = http.NewResponseController(w).Flush()
	})
}
