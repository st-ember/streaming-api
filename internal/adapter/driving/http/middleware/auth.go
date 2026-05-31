package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/ports/token"
)

type contextKey string

const userClaimsKey contextKey = "user_claims"

func Auth(t token.Token, logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenStr string

			// Primary for http requests
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// Expected format: "Bearer <token>"
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					logger.Errorf(r.Context(), log.CategoryAuth, "", "invalid auth header format")
					http.Error(w, "invalid token format", http.StatusUnauthorized)
					return
				}

				tokenStr = parts[1]
			}

			// Fallback to url for websockets
			if tokenStr == "" {
				tokenStr = r.URL.Query().Get("token")
			}

			// Final validation
			if tokenStr == "" {
				logger.Errorf(r.Context(), log.CategoryAuth, "", "no token provided")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Parse claims
			claims, err := t.ParseAccess(tokenStr)
			if err != nil {
				logger.Errorf(r.Context(), log.CategoryAuth, "", "parse token: %v", err)
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Inject into context
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*token.AccessClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*token.AccessClaims)
	if !ok {
		return nil, false
	}

	return claims, true
}
