package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		contains []string
	}{
		{
			name: "error with details",
			appErr: &AppError{
				Code:    ErrCodeValidation,
				Message: "invalid input",
				Details: "email is required",
			},
			contains: []string{"VALIDATION_ERROR", "invalid input", "email is required"},
		},
		{
			name: "error without details",
			appErr: &AppError{
				Code:    ErrCodeUnauthorized,
				Message: "access denied",
			},
			contains: []string{"UNAUTHORIZED", "access denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.appErr.Error()
			for _, want := range tt.contains {
				if !strings.Contains(errStr, want) {
					t.Errorf("AppError.Error() = %q, should contain %q", errStr, want)
				}
			}
		})
	}
}

func TestAppError_WriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		appErr     *AppError
		wantStatus int
		wantCode   ErrorCode
	}{
		{
			name:       "validation error",
			appErr:     NewValidationError("bad input", ""),
			wantStatus: http.StatusBadRequest,
			wantCode:   ErrCodeValidation,
		},
		{
			name:       "unauthorized error",
			appErr:     NewUnauthorizedError("not allowed"),
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name:       "not found error",
			appErr:     NewNotFoundError("user"),
			wantStatus: http.StatusNotFound,
			wantCode:   ErrCodeNotFound,
		},
		{
			name:       "internal error",
			appErr:     NewInternalError("something broke", errors.New("db connection failed")),
			wantStatus: http.StatusInternalServerError,
			wantCode:   ErrCodeInternal,
		},
		{
			name:       "bad request error",
			appErr:     NewBadRequestError("invalid request"),
			wantStatus: http.StatusBadRequest,
			wantCode:   ErrCodeBadRequest,
		},
		{
			name:       "service unavailable error",
			appErr:     NewServiceUnavailableError("LLM", errors.New("timeout")),
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   ErrCodeServiceUnavail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.appErr.WriteJSON(w)

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("WriteJSON() status = %d, want %d", w.Code, tt.wantStatus)
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("WriteJSON() Content-Type = %q, want %q", contentType, "application/json")
			}

			// Check response body
			var response map[string]map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			errorObj, ok := response["error"]
			if !ok {
				t.Fatal("Response missing 'error' field")
			}

			code, ok := errorObj["code"].(string)
			if !ok || ErrorCode(code) != tt.wantCode {
				t.Errorf("WriteJSON() code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test message", "test details")

	if err.Code != ErrCodeValidation {
		t.Errorf("NewValidationError() code = %v, want %v", err.Code, ErrCodeValidation)
	}
	if err.Message != "test message" {
		t.Errorf("NewValidationError() message = %q, want %q", err.Message, "test message")
	}
	if err.Details != "test details" {
		t.Errorf("NewValidationError() details = %q, want %q", err.Details, "test details")
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("NewValidationError() status = %d, want %d", err.HTTPStatus, http.StatusBadRequest)
	}
}

func TestNewInternalError_NilError(t *testing.T) {
	err := NewInternalError("something failed", nil)

	if err.Details != "" {
		t.Errorf("NewInternalError() with nil error should have empty details, got %q", err.Details)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("user")

	if !strings.Contains(err.Message, "user") {
		t.Errorf("NewNotFoundError() message should contain resource name, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "not found") {
		t.Errorf("NewNotFoundError() message should contain 'not found', got %q", err.Message)
	}
}
