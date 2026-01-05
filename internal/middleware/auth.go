package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/davegermiquet/kratos-chi-ollama/internal/auth"
	apperrors "github.com/davegermiquet/kratos-chi-ollama/pkg/errors"
	ory "github.com/ory/client-go"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// SessionContextKey is the key for storing session in context
	SessionContextKey ContextKey = "session"
)

// ExtractSessionToken extracts session token from request
func ExtractSessionToken(r *http.Request) string {
	// 1. Try X-Session-Token header (recommended for APIs)
	if token := r.Header.Get("X-Session-Token"); token != "" {
		return token
	}

	// 2. Try Authorization header with Bearer scheme
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// GetSessionFromContext retrieves the session from context
func GetSessionFromContext(ctx context.Context) (*ory.Session, bool) {
	session, ok := ctx.Value(SessionContextKey).(*ory.Session)
	return session, ok
}

// AuthMiddleware validates the session and rejects unauthenticated requests
func AuthMiddleware(validator auth.SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionToken := ExtractSessionToken(r)

			if sessionToken == "" {
				apperrors.NewUnauthorizedError("missing session token").WriteJSON(w)
				return
			}

			session, err := validator.ValidateSession(r.Context(), sessionToken)
			if err != nil {
				apperrors.NewUnauthorizedError("invalid or expired session").WriteJSON(w)
				return
			}

			// Add session to context
			ctx := context.WithValue(r.Context(), SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware validates session but allows unauthenticated requests
func OptionalAuthMiddleware(validator auth.SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionToken := ExtractSessionToken(r)

			if sessionToken != "" {
				session, err := validator.ValidateSession(r.Context(), sessionToken)
				if err == nil && session != nil {
					ctx := context.WithValue(r.Context(), SessionContextKey, session)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs incoming requests
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In production, use a proper logger
		next.ServeHTTP(w, r)
	})
}

// Recoverer recovers from panics
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				apperrors.NewInternalError("internal server error", nil).WriteJSON(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
