package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeServiceUnavail ErrorCode = "SERVICE_UNAVAILABLE"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WriteJSON writes the error as JSON response
func (e *AppError) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": e,
	})
}

// Error constructors

func NewValidationError(message string, details string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

func NewInternalError(message string, err error) *AppError {
	details := ""
	if err != nil {
		details = err.Error()
	}
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewServiceUnavailableError(service string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeServiceUnavail,
		Message:    fmt.Sprintf("%s is unavailable", service),
		Details:    err.Error(),
		HTTPStatus: http.StatusServiceUnavailable,
	}
}
