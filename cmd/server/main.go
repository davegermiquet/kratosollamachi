package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/davegermiquet/kratos-chi-ollama/config"
	"github.com/davegermiquet/kratos-chi-ollama/internal/auth"
	"github.com/davegermiquet/kratos-chi-ollama/internal/handlers"
	"github.com/davegermiquet/kratos-chi-ollama/internal/langchain"
	mw "github.com/davegermiquet/kratos-chi-ollama/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize clients
	kratosClient := auth.NewKratosClient(cfg.KratosPublicURL, cfg.KratosAdminURL)

	llmClient, err := langchain.NewClient(langchain.Config{
		Provider: cfg.LLMProvider,
		Model:    cfg.LLMModel,
		BaseURL:  cfg.LLMBaseURL,
		APIKey:   cfg.LLMAPIKey,
	})
	if err != nil {
		log.Fatalf("Failed to initialize LLM client: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(kratosClient)
	llmHandler := handlers.NewLLMHandler(llmClient)

	// Setup Chi router
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Public auth routes (no authentication required)
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", authHandler.CreateLoginFlow)
		r.Post("/login/flow", authHandler.GetLoginFlow)
		r.Get("/registration", authHandler.CreateRegistrationFlow)
		r.Post("/registration/flow", authHandler.GetRegistrationFlow)
	})

	// Protected auth routes (authentication required)
	r.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware(kratosClient))
		r.Post("/auth/logout", authHandler.Logout)
		r.Get("/auth/whoami", authHandler.WhoAmI)
	})

	// Protected LLM routes (authentication required)
	r.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware(kratosClient))

		r.Route("/llm", func(r chi.Router) {
			r.Post("/chat", llmHandler.Chat)
			r.Post("/generate", llmHandler.Generate)
		})
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Kratos Public URL: %s", cfg.KratosPublicURL)
	log.Printf("LLM Provider: %s", cfg.LLMProvider)
	log.Printf("LLM Model: %s", cfg.LLMModel)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
