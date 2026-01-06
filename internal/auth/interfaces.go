package auth

import (
	"context"

	ory "github.com/ory/client-go"
)

// SessionValidator validates sessions
type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionToken string) (*ory.Session, error)
}

// LoginFlowManager manages login flows
type LoginFlowManager interface {
	CreateLoginFlow(ctx context.Context) (*ory.LoginFlow, error)
	UpdateLoginFlow(ctx context.Context, flowID string, body ory.UpdateLoginFlowBody) (*ory.SuccessfulNativeLogin, error)
}

// RegistrationFlowManager manages registration flows
type RegistrationFlowManager interface {
	CreateRegistrationFlow(ctx context.Context) (*ory.RegistrationFlow, error)
	UpdateRegistrationFlow(ctx context.Context, flowID string, body ory.UpdateRegistrationFlowBody) (*ory.SuccessfulNativeRegistration, error)
}

// LogoutFlowManager manages logout flows
type LogoutFlowManager interface {
	CreateLogoutFlow(ctx context.Context, cookie string) (*ory.LogoutFlow, error)
	PerformNativeLogout(ctx context.Context, sessionToken ory.PerformNativeLogoutBody) error
}

// VerificationFlowManager manages email verification flows
type VerificationFlowManager interface {
	CreateVerificationFlow(ctx context.Context) (*ory.VerificationFlow, error)
	UpdateVerificationFlow(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error)
}

// KratosService combines all auth operations
type KratosService interface {
	SessionValidator
	LoginFlowManager
	RegistrationFlowManager
	LogoutFlowManager
	VerificationFlowManager
}
