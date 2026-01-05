package handlers

import (
	"io"
	"fmt"
	"encoding/json"
	"net/http"
	"github.com/davegermiquet/kratos-chi-ollama/internal/auth"
	"github.com/davegermiquet/kratos-chi-ollama/internal/middleware"
	ory "github.com/ory/client-go"
)

type AuthHandler struct {
	kratos *auth.KratosClient
}

func NewAuthHandler(kratos *auth.KratosClient) *AuthHandler {
	return &AuthHandler{kratos: kratos}
}

// CreateLoginFlow handles GET /auth/login
func (h *AuthHandler) CreateLoginFlow(w http.ResponseWriter, r *http.Request) {
	refresh := r.URL.Query().Get("refresh") == "true"


	flow, err := h.kratos.CreateLoginFlow(r.Context(), refresh)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, flow)
}

// GetLoginFlow handles GET /auth/login/flow
func (h *AuthHandler) GetLoginFlow(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		http.Error(w, "flow parameter is required", http.StatusBadRequest)
		return
	}

	flow, err := h.kratos.GetLoginFlow(r.Context(), flowID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, flow)
}

func extractCSRFToken(flow *ory.RegistrationFlow) string {
	if flow == nil || flow.Ui.Nodes == nil {
		return ""
	}

	for _, node := range flow.Ui.Nodes {
		// Check if this is an input node
		if node.Attributes.UiNodeInputAttributes != nil {
			attrs := node.Attributes.UiNodeInputAttributes
			
			// Look for the csrf_token field
			if attrs.Name == "csrf_token" {
				// The value could be string or interface{}, handle both
				switch v := attrs.Value.(type) {
				case string:
					return v
				default:
					// If it's another type, try to convert
					if str, ok := v.(string); ok {
						return str
					}
				}
			}
		}
	}
	
	return ""
}
// CreateRegistrationFlow handles GET /auth/registration
func (h *AuthHandler) CreateRegistrationFlow(w http.ResponseWriter, r *http.Request) {

	flow, _ := h.kratos.CreateRegistrationFlow(r.Context())
	fmt.Println(flow)
	csrfToken := extractCSRFToken(flow)

	// Build a cleaner response
	response := map[string]interface{}{
		"flow_id":    flow.Id,
		"csrf_token": csrfToken,
		"expires_at": flow.ExpiresAt,
		"action":     flow.Ui.Action,
		"method":     flow.Ui.Method,
		"fields":     extractFormFields(flow),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper to extract form fields for easier frontend consumption
func extractFormFields(flow *ory.RegistrationFlow) []map[string]interface{} {
	fields := []map[string]interface{}{}
	
	for _, node := range flow.Ui.Nodes {
		if node.Attributes.UiNodeInputAttributes != nil {
			attrs := node.Attributes.UiNodeInputAttributes
			
			field := map[string]interface{}{
				"name":     attrs.Name,
				"type":     attrs.Type,
				"required": attrs.GetRequired(),
				"value":    attrs.Value,
			}
			
			// Add label if available
			if node.Meta.Label != nil {
				field["label"] = node.Meta.Label.Text
			}
			
			fields = append(fields, field)
		}
	}
	
	return fields
}

// GetRegistrationFlow handles GET /auth/registration/flow
func (h *AuthHandler) GetRegistrationFlow(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		http.Error(w, "flow parameter is required", http.StatusBadRequest)
		return
	}
	fmt.Println(flowID)
	bodyBytes, err := io.ReadAll(r.Body)
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	var data map[string]interface{}
    err = json.Unmarshal([]byte(bodyString), &data)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    traits := map[string]interface{}{
    "email": data["email"],
    "name": map[string]string{
        "first": data["first_name"].(string),
        "last":  data["last_name"].(string),
    },
	}
	flowBody := ory.UpdateRegistrationFlowBody{}
	flowBody.UpdateRegistrationFlowWithPasswordMethod = &ory.UpdateRegistrationFlowWithPasswordMethod{
    Method:   "password",
    Password: data["pass"].(string),
    Traits:   traits,
	}

	flow, err := h.kratos.GetRegistrationFlow(r.Context(), flowID,flowBody)
	fmt.Println(flow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, flow)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := r.Header.Get("Cookie")

	flow, err := h.kratos.CreateLogoutFlow(r.Context(), cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, flow)
}

// WhoAmI handles GET /auth/whoami - returns current session
func (h *AuthHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(middleware.SessionContextKey).(*ory.Session)
	respondJSON(w, http.StatusOK, session)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
