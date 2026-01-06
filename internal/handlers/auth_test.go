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
	ValidateSessionFunc        func(ctx context.Context, token string) (*ory.Session, error)
	CreateLoginFlowFunc        func(ctx context.Context) (*ory.LoginFlow, error)
	UpdateLoginFlowFunc        func(ctx context.Context, flowID string, body ory.UpdateLoginFlowBody) (*ory.SuccessfulNativeLogin, error)
	CreateRegistrationFlowFunc func(ctx context.Context) (*ory.RegistrationFlow, error)
	UpdateRegistrationFlowFunc func(ctx context.Context, flowID string, body ory.UpdateRegistrationFlowBody) (*ory.SuccessfulNativeRegistration, error)
	CreateLogoutFlowFunc       func(ctx context.Context, cookie string) (*ory.LogoutFlow, error)
	PerformNativeLogoutFunc    func(ctx context.Context, sessionToken string) error
	CreateVerificationFlowFunc func(ctx context.Context) (*ory.VerificationFlow, error)
	UpdateVerificationFlowFunc func(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error)
	CreateRecoveryFlowFunc     func(ctx context.Context) (*ory.RecoveryFlow, error)
	UpdateRecoveryFlowFunc     func(ctx context.Context, flowID string, body ory.UpdateRecoveryFlowBody) (*ory.RecoveryFlow, error)
	CreateSettingsFlowFunc     func(ctx context.Context) (*ory.SettingsFlow, error)
	UpdateSettingsFlowFunc     func(ctx context.Context, flowID string, body ory.UpdateSettingsFlowBody, sessionToken string) (*ory.SettingsFlow, error)
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

func (m *MockKratosService) PerformNativeLogout(ctx context.Context, sessionToken string) error {
	if m.PerformNativeLogoutFunc != nil {
		return m.PerformNativeLogoutFunc(ctx, sessionToken)
	}
	return errors.New("not implemented")
}

func (m *MockKratosService) CreateVerificationFlow(ctx context.Context) (*ory.VerificationFlow, error) {
	if m.CreateVerificationFlowFunc != nil {
		return m.CreateVerificationFlowFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) UpdateVerificationFlow(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error) {
	if m.UpdateVerificationFlowFunc != nil {
		return m.UpdateVerificationFlowFunc(ctx, flowID, body)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) CreateRecoveryFlow(ctx context.Context) (*ory.RecoveryFlow, error) {
	if m.CreateRecoveryFlowFunc != nil {
		return m.CreateRecoveryFlowFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) UpdateRecoveryFlow(ctx context.Context, flowID string, body ory.UpdateRecoveryFlowBody) (*ory.RecoveryFlow, error) {
	if m.UpdateRecoveryFlowFunc != nil {
		return m.UpdateRecoveryFlowFunc(ctx, flowID, body)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) CreateSettingsFlow(ctx context.Context) (*ory.SettingsFlow, error) {
	if m.CreateSettingsFlowFunc != nil {
		return m.CreateSettingsFlowFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockKratosService) UpdateSettingsFlow(ctx context.Context, flowID string, body ory.UpdateSettingsFlowBody, sessionToken string) (*ory.SettingsFlow, error) {
	if m.UpdateSettingsFlowFunc != nil {
		return m.UpdateSettingsFlowFunc(ctx, flowID, body, sessionToken)
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
		name         string
		sessionToken string
		mockErr      error
		wantStatus   int
	}{
		{
			name:         "success",
			sessionToken: "valid-token-123",
			mockErr:      nil,
			wantStatus:   http.StatusOK,
		},
		{
			name:         "missing session token",
			sessionToken: "",
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:         "kratos error",
			sessionToken: "valid-token-123",
			mockErr:      errors.New("kratos error"),
			wantStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				PerformNativeLogoutFunc: func(ctx context.Context, sessionToken string) error {
					return tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/app/misc/logout", nil)
			if tt.sessionToken != "" {
				req.Header.Set("X-Session-Token", tt.sessionToken)
			}
			w := httptest.NewRecorder()

			handler.Logout(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Logout() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_CreateVerificationFlow(t *testing.T) {
	tests := []struct {
		name       string
		mockFlow   *ory.VerificationFlow
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			mockFlow: &ory.VerificationFlow{
				Id: "verification123",
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
				CreateVerificationFlowFunc: func(ctx context.Context) (*ory.VerificationFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/auth/verification", nil)
			w := httptest.NewRecorder()

			handler.CreateVerificationFlow(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateVerificationFlow() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_RequestVerificationEmail(t *testing.T) {
	tests := []struct {
		name       string
		flowID     string
		body       string
		mockFlow   *ory.VerificationFlow
		mockErr    error
		wantStatus int
	}{
		{
			name:   "success",
			flowID: "flow123",
			body:   `{"email": "test@example.com"}`,
			mockFlow: &ory.VerificationFlow{
				Id: "verification123",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing flow ID",
			flowID:     "",
			body:       `{"email": "test@example.com"}`,
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
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email format",
			flowID:     "flow123",
			body:       `{"email": "notanemail"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "kratos error",
			flowID:     "flow123",
			body:       `{"email": "test@example.com"}`,
			mockFlow:   nil,
			mockErr:    errors.New("kratos error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				UpdateVerificationFlowFunc: func(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			url := "/auth/verification/flow"
			if tt.flowID != "" {
				url += "?flow=" + tt.flowID
			}
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RequestVerificationEmail(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("RequestVerificationEmail() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_SubmitVerificationCode(t *testing.T) {
	tests := []struct {
		name       string
		flowID     string
		body       string
		mockFlow   *ory.VerificationFlow
		mockErr    error
		wantStatus int
	}{
		{
			name:   "success",
			flowID: "flow123",
			body:   `{"email": "test@example.com", "code": "123456"}`,
			mockFlow: &ory.VerificationFlow{
				Id: "verification123",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing flow ID",
			flowID:     "",
			body:       `{"email": "test@example.com", "code": "123456"}`,
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
			body:       `{"code": "123456"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing code",
			flowID:     "flow123",
			body:       `{"email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "code too short",
			flowID:     "flow123",
			body:       `{"email": "test@example.com", "code": "12345"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email format",
			flowID:     "flow123",
			body:       `{"email": "notanemail", "code": "123456"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "kratos error - invalid code",
			flowID:     "flow123",
			body:       `{"email": "test@example.com", "code": "123456"}`,
			mockFlow:   nil,
			mockErr:    errors.New("invalid code"),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockKratosService{
				UpdateVerificationFlowFunc: func(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mock)
			url := "/auth/verification/code"
			if tt.flowID != "" {
				url += "?flow=" + tt.flowID
			}
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SubmitVerificationCode(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SubmitVerificationCode() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_CreateRecoveryFlow(t *testing.T) {
	tests := []struct {
		name       string
		mockFlow   *ory.RecoveryFlow
		mockErr    error
		wantStatus int
	}{
		{
			name: "success",
			mockFlow: &ory.RecoveryFlow{
				Id:        "recovery-flow-123",
				ExpiresAt: ptrTime(time.Now().Add(10 * time.Minute)),
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
			mockKratos := &MockKratosService{
				CreateRecoveryFlowFunc: func(ctx context.Context) (*ory.RecoveryFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mockKratos)
			req := httptest.NewRequest(http.MethodGet, "/auth/recovery", nil)
			w := httptest.NewRecorder()

			handler.CreateRecoveryFlow(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateRecoveryFlow() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_RequestRecoveryCode(t *testing.T) {
	tests := []struct {
		name       string
		flowID     string
		body       string
		mockFlow   *ory.RecoveryFlow
		mockErr    error
		wantStatus int
	}{
		{
			name:   "success",
			flowID: "flow-123",
			body:   `{"email": "user@example.com"}`,
			mockFlow: &ory.RecoveryFlow{
				Id: "flow-123",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing flow ID",
			flowID:     "",
			body:       `{"email": "user@example.com"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			flowID:     "flow-123",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			flowID:     "flow-123",
			body:       `{"email": "invalid-email"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			flowID:     "flow-123",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "kratos error",
			flowID: "flow-123",
			body:   `{"email": "user@example.com"}`,
			mockFlow: nil,
			mockErr:    errors.New("kratos error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKratos := &MockKratosService{
				UpdateRecoveryFlowFunc: func(ctx context.Context, flowID string, body ory.UpdateRecoveryFlowBody) (*ory.RecoveryFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mockKratos)
			url := "/auth/recovery/flow?flow=" + tt.flowID
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RequestRecoveryCode(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("RequestRecoveryCode() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_SubmitRecoveryCode(t *testing.T) {
	tests := []struct {
		name       string
		flowID     string
		body       string
		mockFlow   *ory.RecoveryFlow
		mockErr    error
		wantStatus int
	}{
		{
			name:   "success",
			flowID: "flow-123",
			body:   `{"code": "123456", "password": "NewSecurePassword123!"}`,
			mockFlow: &ory.RecoveryFlow{
				Id: "flow-123",
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing flow ID",
			flowID:     "",
			body:       `{"code": "123456", "password": "NewPass123!"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing code",
			flowID:     "flow-123",
			body:       `{"password": "NewPass123!"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			flowID:     "flow-123",
			body:       `{"code": "123456"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "password too short",
			flowID:     "flow-123",
			body:       `{"code": "123456", "password": "short"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "code too short",
			flowID:     "flow-123",
			body:       `{"code": "123", "password": "NewPass123!"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			flowID:     "flow-123",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "kratos error - invalid code",
			flowID:     "flow-123",
			body:       `{"code": "123456", "password": "NewPass123!"}`,
			mockFlow:   nil,
			mockErr:    errors.New("invalid or expired code"),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKratos := &MockKratosService{
				UpdateRecoveryFlowFunc: func(ctx context.Context, flowID string, body ory.UpdateRecoveryFlowBody) (*ory.RecoveryFlow, error) {
					return tt.mockFlow, tt.mockErr
				},
			}

			handler := NewAuthHandler(mockKratos)
			url := "/auth/recovery/code?flow=" + tt.flowID
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SubmitRecoveryCode(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SubmitRecoveryCode() status = %d, want %d", w.Code, tt.wantStatus)
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
