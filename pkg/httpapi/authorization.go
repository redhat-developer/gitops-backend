package httpapi

import (
	"context"
	"net/http"
	"strings"
)

const (
	authHeader      = "Authorization"
	authTokenCtxKey = "auth-token"
)

// AuthToken gets the auth token from the context.
func AuthToken(ctx context.Context) (string, bool) {
	hook, ok := ctx.Value(authTokenCtxKey).(string)
	return hook, ok
}

// WithAuthToken sets the auth token into the context.
func WithAuthToken(ctx context.Context, t string) context.Context {
	return context.WithValue(ctx, authTokenCtxKey, t)
}

// AuthenticationMiddleware wraps an http.Handler and checks for the presence of
// an 'Authorization' header with a bearer token.
//
// This token is placed into the context, and is accessible via the AuthToken
// function.
//
// No attempt to validate the actual token is made.
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerValue := extractToken(r.Header.Get(authHeader))
		if headerValue == "" {
			http.Error(w, "Authentication required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.Clone(WithAuthToken(r.Context(), headerValue)))
	})
}

func extractToken(s string) string {
	parts := strings.Split(s, " ")
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
