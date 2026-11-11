package http

import (
	"context"
	"net/http"

	"ai-staffing/internal/service/auth"
)

type sessionAuthenticator interface {
	Authenticate(context.Context, auth.SessionSecret) (auth.Principal, error)
}

type requestSession struct {
	principal auth.Principal
	secret    auth.SessionSecret
}

type sessionContextKey struct{}

func authenticateSession(authenticator sessionAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret, ok := readSessionSecret(r)
			if !ok {
				writeAuthSessionError(w, r, auth.ErrUnauthenticated)
				return
			}
			principal, err := authenticator.Authenticate(r.Context(), secret)
			if err != nil {
				writeAuthSessionError(w, r, err)
				return
			}
			session := requestSession{principal: principal, secret: secret}
			ctx := context.WithValue(r.Context(), sessionContextKey{}, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sessionFromRequest(r *http.Request) (requestSession, bool) {
	session, ok := r.Context().Value(sessionContextKey{}).(requestSession)
	return session, ok
}
