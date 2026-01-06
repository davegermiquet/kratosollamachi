package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/davegermiquet/kratos-chi-ollama/config"
	"github.com/davegermiquet/kratos-chi-ollama/internal/auth"
	"github.com/davegermiquet/kratos-chi-ollama/internal/handlers"
	"github.com/davegermiquet/kratos-chi-ollama/internal/langchain"
	"github.com/davegermiquet/kratos-chi-ollama/internal/middleware"
	"github.com/davegermiquet/kratos-chi-ollama/internal/response"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
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
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(kratosClient)
	llmHandler := handlers.NewLLMHandler(llmClient)

	// Create router
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
				r.Post("/logout", authHandler.Logout)
			})
		})
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting server on %s in %s mode", addr, cfg.Server.Environment)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
