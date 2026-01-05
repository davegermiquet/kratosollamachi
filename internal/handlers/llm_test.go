package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// MockLLMService implements langchain.LLMService for testing
type MockLLMService struct {
	GenerateFunc func(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error)
	ChatFunc     func(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error)
}

func (m *MockLLMService) GenerateContent(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, prompt, opts...)
	}
	return "", errors.New("not implemented")
}

func (m *MockLLMService) Chat(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error) {
	if m.ChatFunc != nil {
		return m.ChatFunc(ctx, messages, opts...)
	}
	return "", errors.New("not implemented")
}

func TestLLMHandler_Chat(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		mockResponse string
		mockErr      error
		wantStatus   int
		wantContains string
	}{
		{
			name:         "success single message",
			body:         `{"messages": [{"role": "user", "content": "Hello!"}]}`,
			mockResponse: "Hello! How can I help you?",
			mockErr:      nil,
			wantStatus:   http.StatusOK,
			wantContains: "Hello! How can I help you?",
		},
		{
			name:         "success multiple messages",
			body:         `{"messages": [{"role": "system", "content": "You are helpful"}, {"role": "user", "content": "Hi"}]}`,
			mockResponse: "Hi there!",
			mockErr:      nil,
			wantStatus:   http.StatusOK,
			wantContains: "Hi there!",
		},
		{
			name:       "empty messages",
			body:       `{"messages": []}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			body:       `{not valid json}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "message with empty content",
			body:       `{"messages": [{"role": "user", "content": ""}]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:         "LLM service error",
			body:         `{"messages": [{"role": "user", "content": "Hello!"}]}`,
			mockResponse: "",
			mockErr:      errors.New("LLM unavailable"),
			wantStatus:   http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLLMService{
				ChatFunc: func(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error) {
					return tt.mockResponse, tt.mockErr
				},
			}

			handler := NewLLMHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/llm/chat", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Chat(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Chat() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("Chat() body = %q, want to contain %q", w.Body.String(), tt.wantContains)
			}
		})
	}
}

func TestLLMHandler_Generate(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		mockResponse string
		mockErr      error
		wantStatus   int
		wantContains string
	}{
		{
			name:         "success",
			body:         `{"prompt": "Write a story about a cat"}`,
			mockResponse: "Once upon a time, there was a cat...",
			mockErr:      nil,
			wantStatus:   http.StatusOK,
			wantContains: "Once upon a time",
		},
		{
			name:       "empty prompt",
			body:       `{"prompt": ""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "whitespace prompt",
			body:       `{"prompt": "   "}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing prompt",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:         "LLM service error",
			body:         `{"prompt": "Write something"}`,
			mockResponse: "",
			mockErr:      errors.New("LLM unavailable"),
			wantStatus:   http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLLMService{
				GenerateFunc: func(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
					return tt.mockResponse, tt.mockErr
				},
			}

			handler := NewLLMHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/llm/generate", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Generate(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Generate() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("Generate() body = %q, want to contain %q", w.Body.String(), tt.wantContains)
			}
		})
	}
}

func TestLLMHandler_Chat_MessageRoleConversion(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		wantRole llms.ChatMessageType
	}{
		{
			name:     "system role",
			role:     "system",
			wantRole: llms.ChatMessageTypeSystem,
		},
		{
			name:     "user role",
			role:     "user",
			wantRole: llms.ChatMessageTypeHuman,
		},
		{
			name:     "assistant role",
			role:     "assistant",
			wantRole: llms.ChatMessageTypeAI,
		},
		{
			name:     "unknown role defaults to user",
			role:     "unknown",
			wantRole: llms.ChatMessageTypeHuman,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedMessages []llms.MessageContent

			mock := &MockLLMService{
				ChatFunc: func(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error) {
					capturedMessages = messages
					return "response", nil
				},
			}

			handler := NewLLMHandler(mock)
			body := `{"messages": [{"role": "` + tt.role + `", "content": "test"}]}`
			req := httptest.NewRequest(http.MethodPost, "/llm/chat", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Chat(w, req)

			if len(capturedMessages) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(capturedMessages))
			}

			if capturedMessages[0].Role != tt.wantRole {
				t.Errorf("Message role = %v, want %v", capturedMessages[0].Role, tt.wantRole)
			}
		})
	}
}
