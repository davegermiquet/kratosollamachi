package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save current env and restore after test
	originalEnv := map[string]string{
		"PORT":              os.Getenv("PORT"),
		"ENVIRONMENT":       os.Getenv("ENVIRONMENT"),
		"KRATOS_PUBLIC_URL": os.Getenv("KRATOS_PUBLIC_URL"),
		"KRATOS_ADMIN_URL":  os.Getenv("KRATOS_ADMIN_URL"),
		"LLM_PROVIDER":      os.Getenv("LLM_PROVIDER"),
		"LLM_MODEL":         os.Getenv("LLM_MODEL"),
		"LLM_BASE_URL":      os.Getenv("LLM_BASE_URL"),
		"LLM_API_KEY":       os.Getenv("LLM_API_KEY"),
		"ALLOWED_ORIGINS":   os.Getenv("ALLOWED_ORIGINS"),
	}

	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		check   func(*Config) bool
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"LLM_MODEL": "llama2", // Required field
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.Server.Port == 8080 &&
					c.Server.Environment == "development" &&
					c.LLM.Model == "llama2"
			},
		},
		{
			name: "custom port",
			envVars: map[string]string{
				"PORT":      "3000",
				"LLM_MODEL": "llama2",
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.Server.Port == 3000
			},
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"PORT":      "not-a-number",
				"LLM_MODEL": "llama2",
			},
			wantErr: true,
		},
		{
			name: "production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"LLM_MODEL":   "llama2",
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.IsProduction() && !c.IsDevelopment()
			},
		},
		{
			name: "custom kratos URLs",
			envVars: map[string]string{
				"KRATOS_PUBLIC_URL": "http://kratos:4433",
				"KRATOS_ADMIN_URL":  "http://kratos:4434",
				"LLM_MODEL":         "llama2",
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.Kratos.PublicURL == "http://kratos:4433" &&
					c.Kratos.AdminURL == "http://kratos:4434"
			},
		},
		{
			name: "multiple CORS origins",
			envVars: map[string]string{
				"ALLOWED_ORIGINS": "http://localhost:3000,http://localhost:8080,https://example.com",
				"LLM_MODEL":       "llama2",
			},
			wantErr: false,
			check: func(c *Config) bool {
				return len(c.CORS.AllowedOrigins) == 3
			},
		},
		{
			name: "LLM configuration",
			envVars: map[string]string{
				"LLM_PROVIDER": "openai",
				"LLM_MODEL":    "gpt-4",
				"LLM_BASE_URL": "https://api.openai.com",
				"LLM_API_KEY":  "sk-test123",
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.LLM.Provider == "openai" &&
					c.LLM.Model == "gpt-4" &&
					c.LLM.BaseURL == "https://api.openai.com" &&
					c.LLM.APIKey == "sk-test123"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars
			for key := range originalEnv {
				os.Unsetenv(key)
			}

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Error("Load() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}

			if tt.check != nil && !tt.check(cfg) {
				t.Error("Load() config check failed")
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{Port: 8080, Environment: "development"},
				Kratos: KratosConfig{PublicURL: "http://localhost:4433"},
				LLM:    LLMConfig{Model: "llama2"},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Server: ServerConfig{Port: 0, Environment: "development"},
				Kratos: KratosConfig{PublicURL: "http://localhost:4433"},
				LLM:    LLMConfig{Model: "llama2"},
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{Port: 70000, Environment: "development"},
				Kratos: KratosConfig{PublicURL: "http://localhost:4433"},
				LLM:    LLMConfig{Model: "llama2"},
			},
			wantErr: true,
		},
		{
			name: "empty kratos public URL",
			config: &Config{
				Server: ServerConfig{Port: 8080, Environment: "development"},
				Kratos: KratosConfig{PublicURL: ""},
				LLM:    LLMConfig{Model: "llama2"},
			},
			wantErr: true,
		},
		{
			name: "empty LLM model",
			config: &Config{
				Server: ServerConfig{Port: 8080, Environment: "development"},
				Kratos: KratosConfig{PublicURL: "http://localhost:4433"},
				LLM:    LLMConfig{Model: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"development", true},
		{"production", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Server: ServerConfig{Environment: tt.env}}
			if got := cfg.IsDevelopment(); got != tt.want {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Server: ServerConfig{Environment: tt.env}}
			if got := cfg.IsProduction(); got != tt.want {
				t.Errorf("IsProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		input string
		want  int // number of origins
	}{
		{"http://localhost:3000", 1},
		{"http://localhost:3000,http://localhost:8080", 2},
		{"http://localhost:3000, http://localhost:8080, https://example.com", 3},
		{"", 0},
		{"  ,  ,  ", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseOrigins(tt.input)
			if len(got) != tt.want {
				t.Errorf("parseOrigins(%q) = %d origins, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}
