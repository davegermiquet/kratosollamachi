package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server  ServerConfig
	Kratos  KratosConfig
	LLM     LLMConfig
	CORS    CORSConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port        int
	Environment string
}

// KratosConfig holds Kratos-specific configuration
type KratosConfig struct {
	PublicURL string
	AdminURL  string
}

// LLMConfig holds LLM-specific configuration
type LLMConfig struct {
	Provider string
	Model    string
	BaseURL  string
	APIKey   string
}

// CORSConfig holds CORS-specific configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:        port,
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Kratos: KratosConfig{
			PublicURL: getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
			AdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		},
		LLM: LLMConfig{
			Provider: getEnv("LLM_PROVIDER", "ollama"),
			Model:    getEnv("LLM_MODEL", "llama2"),
			BaseURL:  getEnv("LLM_BASE_URL", "http://localhost:11434"),
			APIKey:   getEnv("LLM_API_KEY", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseOrigins(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Server.Port)
	}

	if c.Kratos.PublicURL == "" {
		return fmt.Errorf("KRATOS_PUBLIC_URL is required")
	}

	if c.LLM.Model == "" {
		return fmt.Errorf("LLM_MODEL is required")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseOrigins(origins string) []string {
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
