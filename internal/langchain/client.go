package langchain

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// Provider represents supported LLM providers
type Provider string

const (
	ProviderOllama Provider = "ollama"
	ProviderOpenAI Provider = "openai"
)

// Config holds the configuration for LLM client
type Config struct {
	Provider Provider
	Model    string
	BaseURL  string
	APIKey   string
}

// Client wraps langchaingo LLM functionality
type Client struct {
	llm    llms.Model
	config Config
}

// Ensure Client implements LLMService
var _ LLMService = (*Client)(nil)

// NewClient creates a new langchain client based on the provider
func NewClient(cfg Config) (*Client, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("model name is required")
	}

	var llm llms.Model
	var err error

	switch cfg.Provider {
	case ProviderOllama:
		llm, err = createOllamaClient(cfg)
	case ProviderOpenAI:
		llm, err = createOpenAIClient(cfg)
	default:
		// Default to Ollama
		llm, err = createOllamaClient(cfg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &Client{llm: llm, config: cfg}, nil
}

func createOllamaClient(cfg Config) (llms.Model, error) {
	opts := []ollama.Option{
		ollama.WithModel(cfg.Model),
	}

	if cfg.BaseURL != "" {
		opts = append(opts, ollama.WithServerURL(cfg.BaseURL))
	}

	return ollama.New(opts...)
}

func createOpenAIClient(cfg Config) (llms.Model, error) {
	opts := []openai.Option{
		openai.WithModel(cfg.Model),
	}

	if cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(cfg.APIKey))
	}

	if cfg.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
	}

	return openai.New(opts...)
}

// GenerateContent generates text from a prompt
func (c *Client) GenerateContent(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	response, err := llms.GenerateFromSinglePrompt(ctx, c.llm, prompt, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return response, nil
}

// Chat sends messages and gets a response
func (c *Client) Chat(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("messages cannot be empty")
	}

	response, err := c.llm.GenerateContent(ctx, messages, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to generate chat response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", nil
	}

	return response.Choices[0].Content, nil
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() Config {
	return c.config
}
