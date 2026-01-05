package langchain

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// Client wraps langchaingo LLM functionality
type Client struct {
	llm llms.Model
}

// Config holds the configuration for LLM client
type Config struct {
	Provider string // "ollama" or "openai"
	Model    string
	BaseURL  string // For Ollama or custom OpenAI endpoint
	APIKey   string // For OpenAI
}

// NewClient creates a new langchain client based on the provider
func NewClient(cfg Config) (*Client, error) {
	var llm llms.Model
	var err error

	switch cfg.Provider {
	case "ollama":
		llm, err = ollama.New(
			ollama.WithModel(cfg.Model),
			ollama.WithServerURL(cfg.BaseURL),
		)
	case "openai":
		opts := []openai.Option{
			openai.WithModel(cfg.Model),
		}
		if cfg.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.APIKey))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
		}
		llm, err = openai.New(opts...)
	default:
		// Default to Ollama
		llm, err = ollama.New(
			ollama.WithModel(cfg.Model),
			ollama.WithServerURL(cfg.BaseURL),
		)
	}

	if err != nil {
		return nil, err
	}

	return &Client{llm: llm}, nil
}

// GenerateContent generates text from a prompt
func (c *Client) GenerateContent(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	response, err := llms.GenerateFromSinglePrompt(ctx, c.llm, prompt, opts...)
	if err != nil {
		return "", err
	}
	return response, nil
}

// Chat sends messages and gets a response
func (c *Client) Chat(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error) {
	response, err := c.llm.GenerateContent(ctx, messages, opts...)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", nil
	}

	return response.Choices[0].Content, nil
}

// StreamGenerateContent generates text with streaming
func (c *Client) StreamGenerateContent(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	// For streaming, you can implement custom logic here
	// This is a simple implementation that collects all chunks
	return llms.GenerateFromSinglePrompt(ctx, c.llm, prompt, opts...)
}
