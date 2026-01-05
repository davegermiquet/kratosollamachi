package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/davegermiquet/kratos-chi-ollama/internal/langchain"
	"github.com/tmc/langchaingo/llms"
)

type LLMHandler struct {
	client *langchain.Client
}

func NewLLMHandler(client *langchain.Client) *LLMHandler {
	return &LLMHandler{client: client}
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Messages []Message `json:"messages"`
}

// ChatResponse represents the chat response
type ChatResponse struct {
	Content string `json:"content"`
}

// GenerateRequest represents the incoming generation request
type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

// GenerateResponse represents the generation response
type GenerateResponse struct {
	Content string `json:"content"`
}

// Chat handles POST /llm/chat
func (h *LLMHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Messages array cannot be empty", http.StatusBadRequest)
		return
	}

	// Convert messages to langchaingo format
	var messages []llms.MessageContent
	for _, msg := range req.Messages {
		var role llms.ChatMessageType
		switch msg.Role {
		case "system":
			role = llms.ChatMessageTypeSystem
		case "assistant":
			role = llms.ChatMessageTypeAI
		case "user":
			role = llms.ChatMessageTypeHuman
		default:
			role = llms.ChatMessageTypeHuman
		}

		messages = append(messages, llms.MessageContent{
			Role: role,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: msg.Content},
			},
		})
	}

	content, err := h.client.Chat(r.Context(), messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, ChatResponse{Content: content})
}

// Generate handles POST /llm/generate
func (h *LLMHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
		return
	}

	content, err := h.client.GenerateContent(r.Context(), req.Prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, GenerateResponse{Content: content})
}
