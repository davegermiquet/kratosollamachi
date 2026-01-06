//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/davegermiquet/kratos-chi-ollama/config"
	"github.com/davegermiquet/kratos-chi-ollama/internal/auth"
	"github.com/davegermiquet/kratos-chi-ollama/internal/handlers"
	"github.com/davegermiquet/kratos-chi-ollama/internal/langchain"
	"github.com/davegermiquet/kratos-chi-ollama/internal/middleware"
	"github.com/davegermiquet/kratos-chi-ollama/internal/response"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// TestUserAuthenticationFlow tests the complete authentication flow:
// 1. Register a new user
// 2. Login as that user
// 3. Call whoami to verify session
// 4. Logout
// 5. Verify session is invalid after logout
func TestUserAuthenticationFlow(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize dependencies
	kratosClient := auth.NewKratosClient(cfg.Kratos.PublicURL, cfg.Kratos.AdminURL)

	llmClient, err := langchain.NewClient(langchain.Config{
		Provider: langchain.Provider(cfg.LLM.Provider),
		Model:    cfg.LLM.Model,
		BaseURL:  cfg.LLM.BaseURL,
		APIKey:   cfg.LLM.APIKey,
	})
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(kratosClient)
	llmHandler := handlers.NewLLMHandler(llmClient)

	// Create test server
	r := setupRouter(cfg, authHandler, llmHandler, kratosClient)
	server := httptest.NewServer(r)
	defer server.Close()

	// Generate unique test user credentials
	timestamp := time.Now().Unix()
	testEmail := fmt.Sprintf("testuser%d@example.com", timestamp)
	testPassword := "T3st!P@ssw0rd#2024$Secure%"
	testFirstName := "Test"
	testLastName := "User"

	t.Logf("Starting integration test with user: %s", testEmail)

	// Step 1: Create Registration Flow
	t.Run("CreateRegistrationFlow", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/users/auth/registration")
		if err != nil {
			t.Fatalf("Failed to create registration flow: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var flowResp response.RegistrationFlowResponse
		if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
			t.Fatalf("Failed to decode registration flow response: %v", err)
		}

		if flowResp.FlowID == "" {
			t.Fatal("Registration flow ID is empty")
		}

		t.Logf("✓ Registration flow created: %s", flowResp.FlowID)
	})

	// Step 2: Submit Registration
	var registrationFlowID string
	t.Run("SubmitRegistration", func(t *testing.T) {
		// First get a fresh flow
		resp, err := http.Get(server.URL + "/api/v1/users/auth/registration")
		if err != nil {
			t.Fatalf("Failed to create registration flow: %v", err)
		}
		defer resp.Body.Close()

		var flowResp response.RegistrationFlowResponse
		if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
			t.Fatalf("Failed to decode registration flow response: %v", err)
		}
		registrationFlowID = flowResp.FlowID

		// Submit registration
		regData := map[string]interface{}{
			"email":      testEmail,
			"pass":       testPassword,
			"first_name": testFirstName,
			"last_name":  testLastName,
		}
		reqBody, _ := json.Marshal(regData)

		url := fmt.Sprintf("%s/api/v1/users/auth/registration/flow?flow=%s", server.URL, registrationFlowID)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to submit registration: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			t.Fatalf("Expected status 201 or 200, got %d. Response: %+v", resp.StatusCode, errorResp)
		}

		t.Logf("✓ User registered successfully: %s", testEmail)
	})

	// Step 3: Create Login Flow
	var loginFlowID string
	t.Run("CreateLoginFlow", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/users/auth/login")
		if err != nil {
			t.Fatalf("Failed to create login flow: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var flowResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
			t.Fatalf("Failed to decode login flow response: %v", err)
		}

		id, ok := flowResp["id"].(string)
		if !ok || id == "" {
			t.Fatal("Login flow ID is empty")
		}
		loginFlowID = id

		t.Logf("✓ Login flow created: %s", loginFlowID)
	})

	// Step 4: Submit Login and Get Session Token
	var sessionToken string
	t.Run("SubmitLogin", func(t *testing.T) {
		loginData := map[string]interface{}{
			"email": testEmail,
			"pass":  testPassword,
		}
		reqBody, _ := json.Marshal(loginData)

		url := fmt.Sprintf("%s/api/v1/users/auth/login/flow?flow=%s", server.URL, loginFlowID)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to submit login: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			t.Fatalf("Expected status 200, got %d. Response: %+v", resp.StatusCode, errorResp)
		}

		var loginResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		// Extract session token
		if token, ok := loginResp["session_token"].(string); ok && token != "" {
			sessionToken = token
		} else {
			// Try nested structure
			if session, ok := loginResp["session"].(map[string]interface{}); ok {
				if token, ok := session["token"].(string); ok {
					sessionToken = token
				}
			}
		}

		if sessionToken == "" {
			t.Fatalf("Session token is empty. Response: %+v", loginResp)
		}

		t.Logf("✓ Login successful, session token received: %s...", sessionToken[:20])
	})

	// Step 5: Call WhoAmI to Verify Session
	t.Run("WhoAmI", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/api/v1/app/misc/whoami", nil)
		if err != nil {
			t.Fatalf("Failed to create whoami request: %v", err)
		}
		req.Header.Set("X-Session-Token", sessionToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to call whoami: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			t.Fatalf("Expected status 200, got %d. Response: %+v", resp.StatusCode, errorResp)
		}

		var whoamiResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&whoamiResp); err != nil {
			t.Fatalf("Failed to decode whoami response: %v", err)
		}

		// Verify the session contains the user's email
		if identity, ok := whoamiResp["identity"].(map[string]interface{}); ok {
			if traits, ok := identity["traits"].(map[string]interface{}); ok {
				if email, ok := traits["email"].(string); ok {
					if email != testEmail {
						t.Fatalf("Expected email %s, got %s", testEmail, email)
					}
					t.Logf("✓ WhoAmI verified session for user: %s", email)
				}
			}
		}
	})

	// Step 6: Logout
	t.Run("Logout", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/api/v1/app/misc/logout", nil)
		if err != nil {
			t.Fatalf("Failed to create logout request: %v", err)
		}
		req.Header.Set("X-Session-Token", sessionToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to call logout: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			t.Fatalf("Expected status 200, got %d. Response: %+v", resp.StatusCode, errorResp)
		}

		var logoutResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&logoutResp); err != nil {
			t.Fatalf("Failed to decode logout response: %v", err)
		}

		if msg, ok := logoutResp["message"].(string); ok {
			if msg != "Successfully logged out" {
				t.Fatalf("Expected logout message 'Successfully logged out', got '%s'", msg)
			}
			t.Logf("✓ Logout successful: %s", msg)
		} else {
			t.Fatalf("Logout response missing message field: %+v", logoutResp)
		}
	})

	// Step 7: Verify Session is Invalid After Logout
	t.Run("VerifySessionInvalidAfterLogout", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/api/v1/app/misc/whoami", nil)
		if err != nil {
			t.Fatalf("Failed to create whoami request: %v", err)
		}
		req.Header.Set("X-Session-Token", sessionToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to call whoami: %v", err)
		}
		defer resp.Body.Close()

		// After logout, whoami should return 401 Unauthorized
		if resp.StatusCode != http.StatusUnauthorized {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			t.Fatalf("Expected status 401 (session should be invalid), got %d. Response: %+v", resp.StatusCode, errorResp)
		}

		t.Logf("✓ Session correctly invalidated after logout")
	})

	t.Log("✅ Integration test completed successfully!")
}

// setupRouter creates and configures the Chi router for testing
func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, llmHandler *handlers.LLMHandler, kratosClient *auth.KratosClient) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Session-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, response.HealthResponse{
			Status:  "healthy",
			Version: "1.0.0",
		})
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Users routes
		r.Route("/users", func(r chi.Router) {
			// Public auth routes
			r.Route("/auth", func(r chi.Router) {
				r.Get("/login", authHandler.CreateLoginFlow)
				r.Post("/login/flow", authHandler.SubmitLogin)
				r.Get("/registration", authHandler.CreateRegistrationFlow)
				r.Post("/registration/flow", authHandler.SubmitRegistration)
			})

			// Protected verification routes
			r.Route("/verification", func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(kratosClient))
				r.Get("/", authHandler.CreateVerificationFlow)
				r.Post("/flow", authHandler.RequestVerificationEmail)
				r.Post("/code", authHandler.SubmitVerificationCode)
			})
		})

		// App routes
		r.Route("/app", func(r chi.Router) {
			// Protected LLM routes
			r.Route("/llm", func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(kratosClient))
				r.Post("/chat", llmHandler.Chat)
				r.Post("/generate", llmHandler.Generate)
			})

			// Protected misc routes (session management, etc)
			r.Route("/misc", func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(kratosClient))
				r.Get("/whoami", authHandler.WhoAmI)
				r.Get("/logout", authHandler.Logout)
			})
		})
	})

	return r
}
