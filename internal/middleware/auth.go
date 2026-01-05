package middleware

import (

	"net/http"

 "github.com/davegermiquet/kratos-chi-ollama/internal/auth"
)


type contextKey string

func getSessionToken(r *http.Request) string  {
	// 1. Try X-Session-Token header (recommended for APIs)
	if token := r.Header.Get("X-Session-Token"); token != "" {
		return token
	}

	return ""
}


// AuthMiddleware validates the session using Kratos
func AuthMiddleware(kratosClient *auth.KratosClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get session from cookie first
			sessionToken := getSessionToken(r)

			// Validate session with Kratos
			_, err := kratosClient.Validate_Session(r.Context(),sessionToken)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

