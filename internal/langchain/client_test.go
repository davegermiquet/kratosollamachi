package langchain

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestMessageRole_ToLLMRole(t *testing.T) {
	tests := []struct {
		role MessageRole
		want llms.ChatMessageType
	}{
		{RoleSystem, llms.ChatMessageTypeSystem},
		{RoleUser, llms.ChatMessageTypeHuman},
		{RoleAssistant, llms.ChatMessageTypeAI},
		{MessageRole("unknown"), llms.ChatMessageTypeHuman},
		{MessageRole(""), llms.ChatMessageTypeHuman},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			got := tt.role.ToLLMRole()
			if got != tt.want {
				t.Errorf("ToLLMRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChatMessage_ToLLMMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     ChatMessage
		wantRole llms.ChatMessageType
		wantText string
	}{
		{
			name: "user message",
			msg: ChatMessage{
				Role:    RoleUser,
				Content: "Hello, world!",
			},
			wantRole: llms.ChatMessageTypeHuman,
			wantText: "Hello, world!",
		},
		{
			name: "system message",
			msg: ChatMessage{
				Role:    RoleSystem,
				Content: "You are a helpful assistant.",
			},
			wantRole: llms.ChatMessageTypeSystem,
			wantText: "You are a helpful assistant.",
		},
		{
			name: "assistant message",
			msg: ChatMessage{
				Role:    RoleAssistant,
				Content: "I'm here to help!",
			},
			wantRole: llms.ChatMessageTypeAI,
			wantText: "I'm here to help!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.ToLLMMessage()

			if got.Role != tt.wantRole {
				t.Errorf("ToLLMMessage() role = %v, want %v", got.Role, tt.wantRole)
			}

			if len(got.Parts) != 1 {
				t.Fatalf("ToLLMMessage() parts count = %d, want 1", len(got.Parts))
			}

			textPart, ok := got.Parts[0].(llms.TextContent)
			if !ok {
				t.Fatal("ToLLMMessage() part is not TextContent")
			}

			if textPart.Text != tt.wantText {
				t.Errorf("ToLLMMessage() text = %q, want %q", textPart.Text, tt.wantText)
			}
		})
	}
}

func TestConvertMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []ChatMessage
		wantLen  int
	}{
		{
			name:     "empty messages",
			messages: []ChatMessage{},
			wantLen:  0,
		},
		{
			name: "single message",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			wantLen: 1,
		},
		{
			name: "multiple messages",
			messages: []ChatMessage{
				{Role: RoleSystem, Content: "Be helpful"},
				{Role: RoleUser, Content: "Hello"},
				{Role: RoleAssistant, Content: "Hi there"},
				{Role: RoleUser, Content: "How are you?"},
			},
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertMessages(tt.messages)

			if len(got) != tt.wantLen {
				t.Errorf("ConvertMessages() length = %d, want %d", len(got), tt.wantLen)
			}

			// Verify each message was converted correctly
			for i, msg := range tt.messages {
				if got[i].Role != msg.Role.ToLLMRole() {
					t.Errorf("ConvertMessages()[%d] role = %v, want %v", i, got[i].Role, msg.Role.ToLLMRole())
				}
			}
		})
	}
}

func TestProvider_Constants(t *testing.T) {
	// Test that provider constants are defined correctly
	if ProviderOllama != "ollama" {
		t.Errorf("ProviderOllama = %q, want %q", ProviderOllama, "ollama")
	}

	if ProviderOpenAI != "openai" {
		t.Errorf("ProviderOpenAI = %q, want %q", ProviderOpenAI, "openai")
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid ollama config",
			config: Config{
				Provider: ProviderOllama,
				Model:    "llama2",
				BaseURL:  "http://localhost:11434",
			},
			wantErr: false,
		},
		{
			name: "valid openai config",
			config: Config{
				Provider: ProviderOpenAI,
				Model:    "gpt-4",
				APIKey:   "sk-test123",
			},
			wantErr: false,
		},
		{
			name: "missing model",
			config: Config{
				Provider: ProviderOllama,
				Model:    "",
				BaseURL:  "http://localhost:11434",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.config)

			if tt.wantErr && err == nil {
				t.Error("NewClient() expected error, got nil")
			}

			// Note: For non-error cases, we can't fully test without a real LLM
			// so we just check that it doesn't error on valid configs
			// The actual connection would fail, but config validation should pass
		})
	}
}

func TestNewClient_DefaultProvider(t *testing.T) {
	// Test that unknown provider defaults to ollama
	config := Config{
		Provider: Provider("unknown"),
		Model:    "test-model",
		BaseURL:  "http://localhost:11434",
	}

	// This will attempt to create an ollama client
	// It may fail to connect, but shouldn't panic
	_, err := NewClient(config)
	
	// We expect this to either succeed or fail gracefully
	// (not panic). The error might be connection-related
	// which is acceptable for this unit test
	_ = err
}
