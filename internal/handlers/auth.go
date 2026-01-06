package handlers

import (
	"net/http"
	"github.com/davegermiquet/kratos-chi-ollama/internal/auth"
	"github.com/davegermiquet/kratos-chi-ollama/internal/middleware"
	"github.com/davegermiquet/kratos-chi-ollama/internal/response"
	"github.com/davegermiquet/kratos-chi-ollama/internal/validation"
	apperrors "github.com/davegermiquet/kratos-chi-ollama/pkg/errors"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	kratos auth.KratosService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(kratos auth.KratosService) *AuthHandler {
	return &AuthHandler{kratos: kratos}
}

// CreateLoginFlow handles GET /auth/login
func (h *AuthHandler) CreateLoginFlow(w http.ResponseWriter, r *http.Request) {
	flow, err := h.kratos.CreateLoginFlow(r.Context())
	if err != nil {
		apperrors.NewServiceUnavailableError("Kratos", err).WriteJSON(w)
		return
	}

	response.Success(w, flow)
}

// SubmitLogin handles POST /auth/login/flow
func (h *AuthHandler) SubmitLogin(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if err := validation.ValidateFlowID(flowID); err != nil {
		err.WriteJSON(w)
		return
	}

	input, validationErr := validation.ValidateLoginInput(r.Body)
	if validationErr != nil {
		validationErr.WriteJSON(w)
		return
	}

	loginBody := auth.BuildPasswordLoginBody(input.Email, input.Password)

	result, err := h.kratos.UpdateLoginFlow(r.Context(), flowID, loginBody)
	if err != nil {
		apperrors.NewUnauthorizedError("invalid credentials").WriteJSON(w)
		return
	}

	response.Success(w, result)
}

// CreateRegistrationFlow handles GET /auth/registration
func (h *AuthHandler) CreateRegistrationFlow(w http.ResponseWriter, r *http.Request) {
	flow, err := h.kratos.CreateRegistrationFlow(r.Context())
	if err != nil {
		apperrors.NewServiceUnavailableError("Kratos", err).WriteJSON(w)
		return
	}

	// Build clean response
	resp := response.RegistrationFlowResponse{
		FlowID:    flow.Id,
		CSRFToken: auth.ExtractCSRFToken(flow),
		Action:    flow.Ui.Action,
		Method:    flow.Ui.Method,
		Fields:    auth.ExtractFormFields(flow),
	}

	if !flow.ExpiresAt.IsZero() {
		resp.ExpiresAt = flow.ExpiresAt.String()
	}

	response.Success(w, resp)
}

// SubmitRegistration handles POST /auth/registration/flow
func (h *AuthHandler) SubmitRegistration(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if err := validation.ValidateFlowID(flowID); err != nil {
		err.WriteJSON(w)
		return
	}

	input, validationErr := validation.ValidateRegistrationInput(r.Body)
	if validationErr != nil {
		validationErr.WriteJSON(w)
		return
	}

	regBody := auth.BuildPasswordRegistrationBody(
		input.Email,
		input.Password,
		input.FirstName,
		input.LastName,
	)

	result, err := h.kratos.UpdateRegistrationFlow(r.Context(), flowID, regBody)
	if err != nil {
		apperrors.NewInternalError("registration failed", err).WriteJSON(w)
		return
	}

	response.Created(w, result)
}

// CreateVerificationFlow handles GET /auth/verification
func (h *AuthHandler) CreateVerificationFlow(w http.ResponseWriter, r *http.Request) {
	flow, err := h.kratos.CreateVerificationFlow(r.Context())
	if err != nil {
		apperrors.NewServiceUnavailableError("Kratos", err).WriteJSON(w)
		return
	}

	response.Success(w, flow)
}

// RequestVerificationEmail handles POST /auth/verification/flow - sends verification email
func (h *AuthHandler) RequestVerificationEmail(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if err := validation.ValidateFlowID(flowID); err != nil {
		err.WriteJSON(w)
		return
	}

	input, validationErr := validation.ValidateVerificationEmailInput(r.Body)
	if validationErr != nil {
		validationErr.WriteJSON(w)
		return
	}

	verificationBody := auth.BuildCodeVerificationBody(input.Email)

	flow, err := h.kratos.UpdateVerificationFlow(r.Context(), flowID, verificationBody)
	if err != nil {
		apperrors.NewInternalError("failed to send verification email", err).WriteJSON(w)
		return
	}

	response.Success(w, flow)
}

// SubmitVerificationCode handles POST /auth/verification/code - verifies with code
func (h *AuthHandler) SubmitVerificationCode(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if err := validation.ValidateFlowID(flowID); err != nil {
		err.WriteJSON(w)
		return
	}

	input, validationErr := validation.ValidateVerificationCodeInput(r.Body)
	if validationErr != nil {
		validationErr.WriteJSON(w)
		return
	}

	verificationBody := auth.BuildCodeVerificationBodySubmit(input.Email,input.Code)

	flow, err := h.kratos.UpdateVerificationFlow(r.Context(), flowID, verificationBody)
	if err != nil {
		apperrors.NewBadRequestError("invalid or expired verification code").WriteJSON(w)
		return
	}

	response.Success(w, flow)
}

// Logout handles POST /api/v1/app/misc/logout - performs native logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := middleware.ExtractSessionToken(r)
	if sessionToken == "" {
		apperrors.NewUnauthorizedError("no session token provided").WriteJSON(w)
		return
	}
	sessionBody := auth.BuildNewPerformNativeLogoutBody	(sessionToken)

	err := h.kratos.PerformNativeLogout(r.Context(), *sessionBody)
	if err != nil {
		apperrors.NewInternalError("failed to logout", err).WriteJSON(w)
		return
	}

	response.Success(w, map[string]string{"message": "Successfully logged out"})
}

// WhoAmI handles GET /auth/whoami - returns current session
func (h *AuthHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	sessionToken := middleware.ExtractSessionToken(r)
	if sessionToken == "" {
		apperrors.NewUnauthorizedError("no session token provided").WriteJSON(w)
		return
	}

	session, err := h.kratos.ValidateSession(r.Context(), sessionToken)
	if err != nil {
		apperrors.NewUnauthorizedError("invalid or expired session").WriteJSON(w)
		return
	}

	response.Success(w, session)
}
