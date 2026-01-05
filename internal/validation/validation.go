package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/mail"
	"strings"

	apperrors "github.com/davegermiquet/kratos-chi-ollama/pkg/errors"
)

// LoginInput represents validated login input
type LoginInput struct {
	Email    string
	Password string
}

// RegistrationInput represents validated registration input
type RegistrationInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// ChatInput represents validated chat input
type ChatInput struct {
	Messages []MessageInput
}

// MessageInput represents a single chat message
type MessageInput struct {
	Role    string
	Content string
}

// GenerateInput represents validated generation input
type GenerateInput struct {
	Prompt string
}

// ValidateLoginInput validates and parses login request body
func ValidateLoginInput(body io.Reader) (*LoginInput, *apperrors.AppError) {
	var data map[string]interface{}
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return nil, apperrors.NewValidationError("Invalid JSON body", err.Error())
	}

	email, ok := data["email"].(string)
	if !ok || strings.TrimSpace(email) == "" {
		return nil, apperrors.NewValidationError("email is required", "")
	}

	if !isValidEmail(email) {
		return nil, apperrors.NewValidationError("invalid email format", "")
	}

	password, ok := data["pass"].(string)
	if !ok || password == "" {
		return nil, apperrors.NewValidationError("password is required", "")
	}

	if len(password) < 6 {
		return nil, apperrors.NewValidationError("password must be at least 6 characters", "")
	}

	return &LoginInput{
		Email:    email,
		Password: password,
	}, nil
}

// ValidateRegistrationInput validates and parses registration request body
func ValidateRegistrationInput(body io.Reader) (*RegistrationInput, *apperrors.AppError) {
	var data map[string]interface{}
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return nil, apperrors.NewValidationError("Invalid JSON body", err.Error())
	}

	email, ok := data["email"].(string)
	if !ok || strings.TrimSpace(email) == "" {
		return nil, apperrors.NewValidationError("email is required", "")
	}

	if !isValidEmail(email) {
		return nil, apperrors.NewValidationError("invalid email format", "")
	}

	password, ok := data["pass"].(string)
	if !ok || password == "" {
		return nil, apperrors.NewValidationError("password is required", "")
	}

	if len(password) < 8 {
		return nil, apperrors.NewValidationError("password must be at least 8 characters", "")
	}

	firstName, ok := data["first_name"].(string)
	if !ok || strings.TrimSpace(firstName) == "" {
		return nil, apperrors.NewValidationError("first_name is required", "")
	}

	lastName, ok := data["last_name"].(string)
	if !ok || strings.TrimSpace(lastName) == "" {
		return nil, apperrors.NewValidationError("last_name is required", "")
	}

	return &RegistrationInput{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

// ValidateChatInput validates chat request
func ValidateChatInput(body io.Reader) (*ChatInput, *apperrors.AppError) {
	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return nil, apperrors.NewValidationError("Invalid JSON body", err.Error())
	}

	if len(req.Messages) == 0 {
		return nil, apperrors.NewValidationError("messages array cannot be empty", "")
	}

	validRoles := map[string]bool{"system": true, "user": true, "assistant": true}
	messages := make([]MessageInput, 0, len(req.Messages))

	for i, msg := range req.Messages {
		if strings.TrimSpace(msg.Content) == "" {
			return nil, apperrors.NewValidationError(
				fmt.Sprintf("message at index %d has empty content", i), "")
		}

		role := strings.ToLower(msg.Role)
		if !validRoles[role] {
			role = "user" // default to user role
		}

		messages = append(messages, MessageInput{
			Role:    role,
			Content: msg.Content,
		})
	}

	return &ChatInput{Messages: messages}, nil
}

// ValidateGenerateInput validates generation request
func ValidateGenerateInput(body io.Reader) (*GenerateInput, *apperrors.AppError) {
	var req struct {
		Prompt string `json:"prompt"`
	}

	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return nil, apperrors.NewValidationError("Invalid JSON body", err.Error())
	}

	if strings.TrimSpace(req.Prompt) == "" {
		return nil, apperrors.NewValidationError("prompt cannot be empty", "")
	}

	return &GenerateInput{Prompt: req.Prompt}, nil
}

// ValidateFlowID validates a flow ID parameter
func ValidateFlowID(flowID string) *apperrors.AppError {
	if strings.TrimSpace(flowID) == "" {
		return apperrors.NewValidationError("flow parameter is required", "")
	}
	return nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
