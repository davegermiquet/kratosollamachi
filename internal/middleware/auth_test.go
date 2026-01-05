package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	ory "github.com/ory/client-go"
)

// MockSessionValidator is a mock implementation of SessionValidator
type MockSessionValidator struct {
	ValidateFunc func(ctx context.Context, token string) (*ory.Session, error)
}

func (m *MockSessionValidator) ValidateSession(ctx context.Context, token string) (*ory.Session, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func TestExtractSessionToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   map[string]string
		wantToken string
	}{
		{
			name:      "X-Session-Token header",
			headers:   map[string]string{"X-Session-Token": "token123"},
			wantToken: "token123",
		},
		{
			name:      "Authorization Bearer header",
			headers:   map[string]string{"Authorization": "Bearer bearertoken456"},
			wantToken: "bearertoken456",
		},
		{
			name:      "X-Session-Token takes precedence",
			headers:   map[string]string{"X-Session-Token": "token123", "Authorization": "Bearer bearertoken456"},
			wantToken: "token123",
		},
		{
			name:      "no token",
			headers:   map[string]string{},
			wantToken: "",
		},
		{
			name:      "invalid Authorization header",
			headers:   map[string]string{"Authorization": "Basic abc"},
			wantToken: "",
		},
		{
			name:      "empty X-Session-Token",
			headers:   map[string]string{"X-Session-Token": ""},
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			got := ExtractSessionToken(req)
			if got != tt.wantToken {
				t.Errorf("ExtractSessionToken() = %q, want %q", got, tt.wantToken)
			}
		})
	}
}

func TestGetSessionFromContext(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		wantOK     bool
		wantNil    bool
	}{
		{
			name: "session in context",
			ctx: context.WithValue(context.Background(), SessionContextKey, &ory.Session{
				Id: "session123",
			}),
			wantOK:  true,
			wantNil: false,
		},
		{
			name:    "no session in context",
			ctx:     context.Background(),
			wantOK:  false,
			wantNil: true,
		},
		{
			name:    "wrong type in context",
			ctx:     context.WithValue(context.Background(), SessionContextKey, "not a session"),
			wantOK:  false,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, ok := GetSessionFromContext(tt.ctx)

			if ok != tt.wantOK {
				t.Errorf("GetSessionFromContext() ok = %v, want %v", ok, tt.wantOK)
			}

			if tt.wantNil && session != nil {
				t.Error("GetSessionFromContext() expected nil session")
			}

			if !tt.wantNil && session == nil {
				t.Error("GetSessionFromContext() expected non-nil session")
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		validateResult *ory.Session
		validateErr    error
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "valid session",
			token:          "valid-token",
			validateResult: &ory.Session{Id: "session123"},
			validateErr:    nil,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "missing token",
			token:          "",
			validateResult: nil,
			validateErr:    nil,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			validateResult: nil,
			validateErr:    errors.New("invalid session"),
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			validator := &MockSessionValidator{
				ValidateFunc: func(ctx context.Context, token string) (*ory.Session, error) {
					return tt.validateResult, tt.validateErr
				},
			}

			middleware := AuthMiddleware(validator)
			handler := middleware(next)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.token != "" {
				req.Header.Set("X-Session-Token", tt.token)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if nextCalled != tt.wantNextCalled {
				t.Errorf("AuthMiddleware() next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		token             string
		validateResult    *ory.Session
		validateErr       error
		wantNextCalled    bool
		wantSessionInCtx  bool
	}{
		{
			name:             "valid session",
			token:            "valid-token",
			validateResult:   &ory.Session{Id: "session123"},
			validateErr:      nil,
			wantNextCalled:   true,
			wantSessionInCtx: true,
		},
		{
			name:             "missing token - still calls next",
			token:            "",
			validateResult:   nil,
			validateErr:      nil,
			wantNextCalled:   true,
			wantSessionInCtx: false,
		},
		{
			name:             "invalid token - still calls next",
			token:            "invalid-token",
			validateResult:   nil,
			validateErr:      errors.New("invalid session"),
			wantNextCalled:   true,
			wantSessionInCtx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			var sessionInCtx bool

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				_, sessionInCtx = GetSessionFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			validator := &MockSessionValidator{
				ValidateFunc: func(ctx context.Context, token string) (*ory.Session, error) {
					return tt.validateResult, tt.validateErr
				},
			}

			middleware := OptionalAuthMiddleware(validator)
			handler := middleware(next)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.token != "" {
				req.Header.Set("X-Session-Token", tt.token)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if nextCalled != tt.wantNextCalled {
				t.Errorf("OptionalAuthMiddleware() next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}

			if sessionInCtx != tt.wantSessionInCtx {
				t.Errorf("OptionalAuthMiddleware() session in context = %v, want %v", sessionInCtx, tt.wantSessionInCtx)
			}
		})
	}
}

func TestRecoverer(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
	}{
		{
			name: "no panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "panic recovered",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Recoverer(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Recoverer() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
