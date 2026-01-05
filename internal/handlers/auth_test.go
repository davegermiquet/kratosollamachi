package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ory "github.com/ory/client-go"
)

// MockKratosService implements auth.KratosService for testing
type MockKratosService struct {
	ValidateSessionFunc       func(ctx context.Context, token string) (*ory.Session, error)
	CreateLoginFlowFunc       func(ctx context.Context) (*ory.LoginFlow, error)
	UpdateLoginFlowFunc       func(ctx context.Context, flowID string, body ory.UpdateLoginFlowBody) (*ory.SuccessfulNativeLogin, error)
	CreateRegistrationFlowFunc func(ctx context.Context) (*ory.RegistrationFlow, error)
	UpdateRegistrationFlowFunc func(ctx context.Context, flowID string, body ory.UpdateRegistrationFlowBody) (*ory.SuccessfulNativeRegistration, error)
	CreateLogoutFlowFunc      func(ctx context.Context, cookie string) (*ory.LogoutFlow, error)
}

func (m *MockKratosService) ValidateSession(ctx context.Context, token string) (*ory.Session, error) {
	if m.ValidateSessionFunc != nil {
		return m.ValidateSessionFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) CreateLoginFlow(ctx context.Context) (*ory.LoginFlow, error) {
	if m.CreateLoginFlowFunc != nil {
		return m.CreateLoginFlowFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) UpdateLoginFlow(ctx context.Context, flowID string, body ory.UpdateLoginFlowBody) (*ory.SuccessfulNativeLogin, error) {
	if m.UpdateLoginFlowFunc != nil {
		return m.UpdateLoginFlowFunc(ctx, flowID, body)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) CreateRegistrationFlow(ctx context.Context) (*ory.RegistrationFlow, error) {
	if m.CreateRegistrationFlowFunc != nil {
		return m.CreateRegistrationFlowFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) UpdateRegistrationFlow(ctx context.Context, flowID string, body ory.UpdateRegistrationFlowBody) (*ory.SuccessfulNativeRegistration, error) {
	if m.UpdateRegistrationFlowFunc != nil {
		return m.UpdateRegistrationFlowFunc(ctx, flowID, body)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) CreateLogoutFlow(ctx context.Context, cookie string) (*ory.LogoutFlow, error) {
	if m.CreateLogoutFlowFunc != nil {
		return m.CreateLogoutFlowFunc(ctx, cookie)
	}
	return nil, errors.New("not implemented")
}

func TestAuthHandler_CreateLoginFlow(t *testing.T) {
	tests := []struct {
		name       string
		mockFlow   *ory.LoginFlow
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			mockFlow: &ory.LoginFlow{
				Id: "flow123",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "kratos error",
			mockFlow:   nil,
			mockErr:    errors.New("kratos unavailable"),
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				CreateLoginFlowFunc: func(ctx context.Context) (*ory.LoginFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
			w := httptest.NewRecorder()

			handler.CreateLoginFlow(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateLoginFlow() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_SubmitLogin(t *testing.T) {
	tests := []struct {
		name       string
		flowID     string
		body       string
		mockResult *ory.SuccessfulNativeLogin
		mockErr    error
		wantStatus int
	}{
		{
			name:   "success",
			flowID: "flow123",
			body:   `{"email": "test@example.com", "pass": "password123"}`,
			mockResult: &ory.SuccessfulNativeLogin{
				SessionToken: ptrString("token123"),
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing flow ID",
			flowID:     "",
			body:       `{"email": "test@example.com", "pass": "password123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			flowID:     "flow123",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			flowID:     "flow123",
			body:       `{"pass": "password123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "kratos auth error",
			flowID:     "flow123",
			body:       `{"email": "test@example.com", "pass": "wrongpassword"}`,
			mockResult: nil,
			mockErr:    errors.New("invalid credentials"),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				UpdateLoginFlowFunc: func(ctx context.Context, flowID string, body ory.UpdateLoginFlowBody) (*ory.SuccessfulNativeLogin, error) {
					return tt.mockResult, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			
			url := "/auth/login/flow"
			if tt.flowID != "" {
				url += "?flow=" + tt.flowID
			}
			
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SubmitLogin(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SubmitLogin() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_CreateRegistrationFlow(t *testing.T) {
	tests := []struct {
		name       string
		mockFlow   *ory.RegistrationFlow
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			mockFlow: &ory.RegistrationFlow{
				Id: "flow123",
				Ui: ory.UiContainer{
					Action: "https://example.com/submit",
					Method: "POST",
					Nodes:  []ory.UiNode{},
				},
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "kratos error",
			mockFlow:   nil,
			mockErr:    errors.New("kratos unavailable"),
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				CreateRegistrationFlowFunc: func(ctx context.Context) (*ory.RegistrationFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/auth/registration", nil)
			w := httptest.NewRecorder()

			handler.CreateRegistrationFlow(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateRegistrationFlow() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_WhoAmI(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		mockSession *ory.Session
		mockErr     error
		wantStatus  int
	}{
		{
			name:  "valid session",
			token: "valid-token",
			mockSession: &ory.Session{
				Id: "session123",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			mockSession: nil,
			mockErr:     errors.New("invalid session"),
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				ValidateSessionFunc: func(ctx context.Context, token string) (*ory.Session, error) {
					return tt.mockSession, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/auth/whoami", nil)
			if tt.token != "" {
				req.Header.Set("X-Session-Token", tt.token)
			}
			w := httptest.NewRecorder()

			handler.WhoAmI(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("WhoAmI() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name       string
		cookie     string
		mockFlow   *ory.LogoutFlow
		mockErr    error
		wantStatus int
	}{
		{
			name:   "success",
			cookie: "session=abc123",
			mockFlow: &ory.LogoutFlow{
				LogoutUrl: "https://example.com/logout",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing cookie",
			cookie:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "kratos error",
			cookie:     "session=abc123",
			mockFlow:   nil,
			mockErr:    errors.New("kratos error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				CreateLogoutFlowFunc: func(ctx context.Context, cookie string) (*ory.LogoutFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			if tt.cookie != "" {
				req.Header.Set("Cookie", tt.cookie)
			}
			w := httptest.NewRecorder()

			handler.Logout(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Logout() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// Helper function to create string pointer
func ptrString(s string) *string {
	return &s
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
