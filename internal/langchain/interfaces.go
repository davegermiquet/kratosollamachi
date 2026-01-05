package langchain

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// LLMService defines the LLM operations
type LLMService interface {
	GenerateContent(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error)
	Chat(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error)
}

// MessageRole represents chat message roles
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// ToLLMRole converts MessageRole to langchaingo ChatMessageType
func (r MessageRole) ToLLMRole() llms.ChatMessageType {
	switch r {
	case RoleSystem:
		return llms.ChatMessageTypeSystem
	case RoleAssistant:
		return llms.ChatMessageTypeAI
	case RoleUser:
		return llms.ChatMessageTypeHuman
	default:
		return llms.ChatMessageTypeHuman
	}
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    MessageRole
	Content string
}

// ToLLMMessage converts ChatMessage to langchaingo MessageContent
func (m ChatMessage) ToLLMMessage() llms.MessageContent {
	return llms.MessageContent{
		Role: m.Role.ToLLMRole(),
		Parts: []llms.ContentPart{
			llms.TextContent{Text: m.Content},
		},
	}
}

// ConvertMessages converts a slice of ChatMessage to langchaingo format
func ConvertMessages(messages []ChatMessage) []llms.MessageContent {
	result := make([]llms.MessageContent, 0, len(messages))
	for _, msg := range messages {
		result = append(result, msg.ToLLMMessage())
	}
	return result
}
