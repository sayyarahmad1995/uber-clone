package identity

import (
	"context"
	"net/http"
	"strings"
)

type contextKey struct{}

func Middleware(provider Provider, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			http.Error(w, ErrUnauthenticated.Error(), http.StatusUnauthorized)
			return
		}

		principal, err := provider.AuthenticateVerified(r.Context(), token)
		if err != nil {
			http.Error(w, ErrUnauthenticated.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey{}, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(contextKey{}).(Principal)
	return principal, ok
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
