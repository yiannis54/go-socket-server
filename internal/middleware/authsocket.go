package middleware

import (
	"context"
	"net/http"

	"github.com/yiannis54/go-socket-server/internal/config"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

func AuthMiddleware(cfg *config.EnvConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get(cfg.TokenKey)
		if !validateToken(r, token) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Optional: set userID when token validation provides it.
		// r = r.WithContext(context.WithValue(r.Context(), UserIDContextKey, userID))
		next.ServeHTTP(w, r)
	})
}

func validateToken(r *http.Request, _ string) bool {
	// Implement token authorization based on strategy used.
	//
	// if err := tokenService.ValidateAccessToken(token); err != nil {
	// 	return false
	// }
	return true
}

// UserIDFromRequest returns the userID stored in the request context, if any.
func UserIDFromRequest(ctx context.Context) (string, bool) {
	v := ctx.Value(UserIDContextKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
