package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	Environment       string
	KratosPublicURL   string
	KratosAdminURL    string
	LLMProvider       string
	LLMModel          string
	LLMBaseURL        string
	LLMAPIKey         string
	AllowedOrigins    []string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		KratosPublicURL: getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
		KratosAdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		LLMProvider:     getEnv("LLM_PROVIDER", "ollama"),
		LLMModel:        getEnv("LLM_MODEL", "llama2"),
		LLMBaseURL:      getEnv("LLM_BASE_URL", "http://localhost:11434"),
		LLMAPIKey:       getEnv("LLM_API_KEY", ""),
		AllowedOrigins:  strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
