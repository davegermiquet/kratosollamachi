package validation

import (
	"strings"
	"testing"
)

func TestValidateLoginInput(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantEmail   string
		wantPass    string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid input",
			body:      `{"email": "test@example.com", "pass": "password123"}`,
			wantEmail: "test@example.com",
			wantPass:  "password123",
			wantErr:   false,
		},
		{
			name:        "missing email",
			body:        `{"pass": "password123"}`,
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name:        "empty email",
			body:        `{"email": "", "pass": "password123"}`,
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name:        "invalid email format",
			body:        `{"email": "notanemail", "pass": "password123"}`,
			wantErr:     true,
			errContains: "invalid email format",
		},
		{
			name:        "missing password",
			body:        `{"email": "test@example.com"}`,
			wantErr:     true,
			errContains: "password is required",
		},
		{
			name:        "password too short",
			body:        `{"email": "test@example.com", "pass": "12345"}`,
			wantErr:     true,
			errContains: "at least 6 characters",
		},
		{
			name:        "invalid JSON",
			body:        `{invalid json}`,
			wantErr:     true,
			errContains: "Invalid JSON",
		},
		{
			name:        "whitespace email",
			body:        `{"email": "   ", "pass": "password123"}`,
			wantErr:     true,
			errContains: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.body)
			result, err := ValidateLoginInput(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateLoginInput() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Message, tt.errContains) {
					t.Errorf("ValidateLoginInput() error = %q, want error containing %q", err.Message, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateLoginInput() unexpected error: %v", err)
				return
			}

			if result.Email != tt.wantEmail {
				t.Errorf("ValidateLoginInput() Email = %q, want %q", result.Email, tt.wantEmail)
			}

			if result.Password != tt.wantPass {
				t.Errorf("ValidateLoginInput() Password = %q, want %q", result.Password, tt.wantPass)
			}
		})
	}
}

func TestValidateRegistrationInput(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid input",
			body:    `{"email": "test@example.com", "pass": "password123", "first_name": "John", "last_name": "Doe"}`,
			wantErr: false,
		},
		{
			name:        "missing first_name",
			body:        `{"email": "test@example.com", "pass": "password123", "last_name": "Doe"}`,
			wantErr:     true,
			errContains: "first_name is required",
		},
		{
			name:        "missing last_name",
			body:        `{"email": "test@example.com", "pass": "password123", "first_name": "John"}`,
			wantErr:     true,
			errContains: "last_name is required",
		},
		{
			name:        "password too short for registration",
			body:        `{"email": "test@example.com", "pass": "1234567", "first_name": "John", "last_name": "Doe"}`,
			wantErr:     true,
			errContains: "at least 8 characters",
		},
		{
			name:        "invalid email",
			body:        `{"email": "invalid", "pass": "password123", "first_name": "John", "last_name": "Doe"}`,
			wantErr:     true,
			errContains: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.body)
			result, err := ValidateRegistrationInput(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateRegistrationInput() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Message, tt.errContains) {
					t.Errorf("ValidateRegistrationInput() error = %q, want error containing %q", err.Message, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateRegistrationInput() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("ValidateRegistrationInput() returned nil result")
			}
		})
	}
}

func TestValidateChatInput(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantErr     bool
		errContains string
		wantMsgCnt  int
	}{
		{
			name:       "valid single message",
			body:       `{"messages": [{"role": "user", "content": "Hello!"}]}`,
			wantErr:    false,
			wantMsgCnt: 1,
		},
		{
			name:       "valid multiple messages",
			body:       `{"messages": [{"role": "system", "content": "You are helpful"}, {"role": "user", "content": "Hello!"}]}`,
			wantErr:    false,
			wantMsgCnt: 2,
		},
		{
			name:        "empty messages array",
			body:        `{"messages": []}`,
			wantErr:     true,
			errContains: "messages array cannot be empty",
		},
		{
			name:        "message with empty content",
			body:        `{"messages": [{"role": "user", "content": ""}]}`,
			wantErr:     true,
			errContains: "empty content",
		},
		{
			name:        "message with whitespace content",
			body:        `{"messages": [{"role": "user", "content": "   "}]}`,
			wantErr:     true,
			errContains: "empty content",
		},
		{
			name:       "unknown role defaults to user",
			body:       `{"messages": [{"role": "unknown", "content": "Hello!"}]}`,
			wantErr:    false,
			wantMsgCnt: 1,
		},
		{
			name:        "invalid JSON",
			body:        `{not valid}`,
			wantErr:     true,
			errContains: "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.body)
			result, err := ValidateChatInput(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateChatInput() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Message, tt.errContains) {
					t.Errorf("ValidateChatInput() error = %q, want error containing %q", err.Message, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateChatInput() unexpected error: %v", err)
				return
			}

			if len(result.Messages) != tt.wantMsgCnt {
				t.Errorf("ValidateChatInput() message count = %d, want %d", len(result.Messages), tt.wantMsgCnt)
			}
		})
	}
}

func TestValidateGenerateInput(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantPrompt  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid prompt",
			body:       `{"prompt": "Write a story"}`,
			wantPrompt: "Write a story",
			wantErr:    false,
		},
		{
			name:        "empty prompt",
			body:        `{"prompt": ""}`,
			wantErr:     true,
			errContains: "prompt cannot be empty",
		},
		{
			name:        "whitespace prompt",
			body:        `{"prompt": "   "}`,
			wantErr:     true,
			errContains: "prompt cannot be empty",
		},
		{
			name:        "missing prompt",
			body:        `{}`,
			wantErr:     true,
			errContains: "prompt cannot be empty",
		},
		{
			name:        "invalid JSON",
			body:        `{bad json}`,
			wantErr:     true,
			errContains: "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.body)
			result, err := ValidateGenerateInput(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateGenerateInput() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Message, tt.errContains) {
					t.Errorf("ValidateGenerateInput() error = %q, want error containing %q", err.Message, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateGenerateInput() unexpected error: %v", err)
				return
			}

			if result.Prompt != tt.wantPrompt {
				t.Errorf("ValidateGenerateInput() Prompt = %q, want %q", result.Prompt, tt.wantPrompt)
			}
		})
	}
}

func TestValidateFlowID(t *testing.T) {
	tests := []struct {
		name    string
		flowID  string
		wantErr bool
	}{
		{
			name:    "valid flow ID",
			flowID:  "abc123",
			wantErr: false,
		},
		{
			name:    "empty flow ID",
			flowID:  "",
			wantErr: true,
		},
		{
			name:    "whitespace flow ID",
			flowID:  "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFlowID(tt.flowID)

			if tt.wantErr && err == nil {
				t.Error("ValidateFlowID() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ValidateFlowID() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateVerificationEmailInput(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantEmail   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid email",
			body:      `{"email": "test@example.com"}`,
			wantEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name:        "missing email",
			body:        `{}`,
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name:        "empty email",
			body:        `{"email": ""}`,
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name:        "whitespace email",
			body:        `{"email": "   "}`,
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name:        "invalid email format",
			body:        `{"email": "notanemail"}`,
			wantErr:     true,
			errContains: "invalid email format",
		},
		{
			name:        "invalid JSON",
			body:        `{invalid}`,
			wantErr:     true,
			errContains: "Invalid JSON",
		},
		{
			name:      "email with whitespace trimmed",
			body:      `{"email": "  test@example.com  "}`,
			wantEmail: "test@example.com",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.body)
			result, err := ValidateVerificationEmailInput(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateVerificationEmailInput() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Message, tt.errContains) {
					t.Errorf("ValidateVerificationEmailInput() error = %q, want error containing %q", err.Message, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateVerificationEmailInput() unexpected error: %v", err)
				return
			}

			if result.Email != tt.wantEmail {
				t.Errorf("ValidateVerificationEmailInput() Email = %q, want %q", result.Email, tt.wantEmail)
			}
		})
	}
}

func TestValidateVerificationCodeInput(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantEmail   string
		wantCode    string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid input",
			body:      `{"email": "test@example.com", "code": "123456"}`,
			wantEmail: "test@example.com",
			wantCode:  "123456",
			wantErr:   false,
		},
		{
			name:        "missing email",
			body:        `{"code": "123456"}`,
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name:        "empty email",
			body:        `{"email": "", "code": "123456"}`,
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name:        "invalid email format",
			body:        `{"email": "notanemail", "code": "123456"}`,
			wantErr:     true,
			errContains: "invalid email format",
		},
		{
			name:        "missing code",
			body:        `{"email": "test@example.com"}`,
			wantErr:     true,
			errContains: "code is required",
		},
		{
			name:        "empty code",
			body:        `{"email": "test@example.com", "code": ""}`,
			wantErr:     true,
			errContains: "code is required",
		},
		{
			name:        "code too short",
			body:        `{"email": "test@example.com", "code": "12345"}`,
			wantErr:     true,
			errContains: "at least 6 characters",
		},
		{
			name:        "whitespace code",
			body:        `{"email": "test@example.com", "code": "   "}`,
			wantErr:     true,
			errContains: "code is required",
		},
		{
			name:        "invalid JSON",
			body:        `{invalid}`,
			wantErr:     true,
			errContains: "Invalid JSON",
		},
		{
			name:      "valid with trimming",
			body:      `{"email": "  test@example.com  ", "code": "  123456  "}`,
			wantEmail: "test@example.com",
			wantCode:  "123456",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.body)
			result, err := ValidateVerificationCodeInput(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateVerificationCodeInput() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Message, tt.errContains) {
					t.Errorf("ValidateVerificationCodeInput() error = %q, want error containing %q", err.Message, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateVerificationCodeInput() unexpected error: %v", err)
				return
			}

			if result.Email != tt.wantEmail {
				t.Errorf("ValidateVerificationCodeInput() Email = %q, want %q", result.Email, tt.wantEmail)
			}

			if result.Code != tt.wantCode {
				t.Errorf("ValidateVerificationCodeInput() Code = %q, want %q", result.Code, tt.wantCode)
			}
		})
	}
}
