package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Success writes a success response with data
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created writes a 201 created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 no content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ChatResponse represents a chat API response
type ChatResponse struct {
	Content string `json:"content"`
}

// GenerateResponse represents a generation API response
type GenerateResponse struct {
	Content string `json:"content"`
}

// RegistrationFlowResponse represents a clean registration flow response
type RegistrationFlowResponse struct {
	FlowID    string                   `json:"flow_id"`
	CSRFToken string                   `json:"csrf_token"`
	ExpiresAt string                   `json:"expires_at,omitempty"`
	Action    string                   `json:"action"`
	Method    string                   `json:"method"`
	Fields    []map[string]interface{} `json:"fields"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}
