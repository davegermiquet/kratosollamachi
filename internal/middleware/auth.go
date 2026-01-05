package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/davegermiquet/kratos-chi-ollama/internal/auth"
)

type contextKey string

const SessionContextKey contextKey = "session"

// AuthMiddleware validates the session using Kratos
func AuthMiddleware(kratosClient *auth.KratosClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get session from cookie first
			cookie := r.Header.Get("Cookie")

			// Try to get session token from Authorization header
			token := ""
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}

			// Validate session with Kratos
			session, err := kratosClient.ToSession(r.Context(), cookie, token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add session to context
			ctx := context.WithValue(r.Context(), SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware validates the session but doesn't reject unauthenticated requests
func OptionalAuthMiddleware(kratosClient *auth.KratosClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie := r.Header.Get("Cookie")
			token := ""
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}

			session, err := kratosClient.ToSession(r.Context(), cookie, token)
			if err == nil {
				ctx := context.WithValue(r.Context(), SessionContextKey, session)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Continue without session
			next.ServeHTTP(w, r)
		})
	}
}
