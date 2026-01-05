package handlers

import (
	"net/http"

	"github.com/davegermiquet/kratos-chi-ollama/internal/langchain"
	"github.com/davegermiquet/kratos-chi-ollama/internal/response"
	"github.com/davegermiquet/kratos-chi-ollama/internal/validation"
	apperrors "github.com/davegermiquet/kratos-chi-ollama/pkg/errors"
)

// LLMHandler handles LLM-related requests
type LLMHandler struct {
	llm langchain.LLMService
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(llm langchain.LLMService) *LLMHandler {
	return &LLMHandler{llm: llm}
}

// Chat handles POST /llm/chat
func (h *LLMHandler) Chat(w http.ResponseWriter, r *http.Request) {
	input, err := validation.ValidateChatInput(r.Body)
	if err != nil {
		err.WriteJSON(w)
		return
	}

	// Convert to langchain message format
	messages := make([]langchain.ChatMessage, 0, len(input.Messages))
	for _, msg := range input.Messages {
		messages = append(messages, langchain.ChatMessage{
			Role:    langchain.MessageRole(msg.Role),
			Content: msg.Content,
		})
	}

	llmMessages := langchain.ConvertMessages(messages)

	content, chatErr := h.llm.Chat(r.Context(), llmMessages)
	if chatErr != nil {
		apperrors.NewServiceUnavailableError("LLM", chatErr).WriteJSON(w)
		return
	}

	response.Success(w, response.ChatResponse{Content: content})
}

// Generate handles POST /llm/generate
func (h *LLMHandler) Generate(w http.ResponseWriter, r *http.Request) {
	input, err := validation.ValidateGenerateInput(r.Body)
	if err != nil {
		err.WriteJSON(w)
		return
	}

	content, genErr := h.llm.GenerateContent(r.Context(), input.Prompt)
	if genErr != nil {
		apperrors.NewServiceUnavailableError("LLM", genErr).WriteJSON(w)
		return
	}

	response.Success(w, response.GenerateResponse{Content: content})
}
