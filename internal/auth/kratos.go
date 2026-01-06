package auth

import (
	"context"
	"fmt"
	"net/http"

	ory "github.com/ory/client-go"
)

// KratosClient implements KratosService interface
type KratosClient struct {
	frontend *ory.APIClient
	admin    *ory.APIClient
}

// Ensure KratosClient implements KratosService
var _ KratosService = (*KratosClient)(nil)

// NewKratosClient creates a new Kratos client
func NewKratosClient(publicURL, adminURL string) *KratosClient {
	frontendConfig := ory.NewConfiguration()
	frontendConfig.Servers = ory.ServerConfigurations{
		{URL: publicURL},
	}

	adminConfig := ory.NewConfiguration()
	adminConfig.Servers = ory.ServerConfigurations{
		{URL: adminURL},
	}

	return &KratosClient{
		frontend: ory.NewAPIClient(frontendConfig),
		admin:    ory.NewAPIClient(adminConfig),
	}
}

// ValidateSession validates a session token and returns the session
func (k *KratosClient) ValidateSession(ctx context.Context, sessionToken string) (*ory.Session, error) {
	if sessionToken == "" {
		return nil, fmt.Errorf("session token is required")
	}

	session, resp, err := k.frontend.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).
		Execute()

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("unauthorized: invalid or expired session")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return session, nil
}

// CreateLoginFlow creates a new native login flow
func (k *KratosClient) CreateLoginFlow(ctx context.Context) (*ory.LoginFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create login flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// UpdateLoginFlow submits login credentials
func (k *KratosClient) UpdateLoginFlow(ctx context.Context, flowID string, body ory.UpdateLoginFlowBody) (*ory.SuccessfulNativeLogin, error) {
	result, resp, err := k.frontend.FrontendAPI.UpdateLoginFlow(ctx).
		Flow(flowID).
		UpdateLoginFlowBody(body).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to update login flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return result, nil
}

// CreateRegistrationFlow creates a new native registration flow
func (k *KratosClient) CreateRegistrationFlow(ctx context.Context) (*ory.RegistrationFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create registration flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// UpdateRegistrationFlow submits registration data
func (k *KratosClient) UpdateRegistrationFlow(ctx context.Context, flowID string, body ory.UpdateRegistrationFlowBody) (*ory.SuccessfulNativeRegistration, error) {
	result, resp, err := k.frontend.FrontendAPI.UpdateRegistrationFlow(ctx).
		Flow(flowID).
		UpdateRegistrationFlowBody(body).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to update registration flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return result, nil
}

// CreateLogoutFlow creates a browser logout flow
func (k *KratosClient) CreateLogoutFlow(ctx context.Context, cookie string) (*ory.LogoutFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateBrowserLogoutFlow(ctx).
		Cookie(cookie).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to create logout flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// PerformNativeLogout performs a native logout by disabling the session
func (k *KratosClient) PerformNativeLogout(ctx context.Context, sessionToken ory.PerformNativeLogoutBody) error {
	
	resp, err := k.frontend.FrontendAPI.PerformNativeLogout(ctx).PerformNativeLogoutBody(sessionToken).Execute()

	if err != nil {
		return fmt.Errorf("failed to logout: %w (status: %d)", err, getStatusCode(resp))
	}

	return nil
}

// CreateVerificationFlow creates a new native verification flow
func (k *KratosClient) CreateVerificationFlow(ctx context.Context) (*ory.VerificationFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateNativeVerificationFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create verification flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// UpdateVerificationFlow submits verification data (email or code)
func (k *KratosClient) UpdateVerificationFlow(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.UpdateVerificationFlow(ctx).
		Flow(flowID).
		UpdateVerificationFlowBody(body).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to update verification flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// Helper functions

func getStatusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

// ExtractCSRFToken extracts CSRF token from a registration flow
func ExtractCSRFToken(flow *ory.RegistrationFlow) string {
	if flow == nil || flow.Ui.Nodes == nil {
		return ""
	}

	for _, node := range flow.Ui.Nodes {
		if node.Attributes.UiNodeInputAttributes != nil {
			attrs := node.Attributes.UiNodeInputAttributes
			if attrs.Name == "csrf_token" {
				if str, ok := attrs.Value.(string); ok {
					return str
				}
			}
		}
	}
	return ""
}

// ExtractFormFields extracts form fields from a registration flow
func ExtractFormFields(flow *ory.RegistrationFlow) []map[string]interface{} {
	fields := make([]map[string]interface{}, 0)

	for _, node := range flow.Ui.Nodes {
		if node.Attributes.UiNodeInputAttributes != nil {
			attrs := node.Attributes.UiNodeInputAttributes

			field := map[string]interface{}{
				"name":     attrs.Name,
				"type":     attrs.Type,
				"required": attrs.GetRequired(),
				"value":    attrs.Value,
			}

			if node.Meta.Label != nil {
				field["label"] = node.Meta.Label.Text
			}

			fields = append(fields, field)
		}
	}

	return fields
}

func BuildNewPerformNativeLogoutBody(sessionToken string) *ory.PerformNativeLogoutBody{
	return ory.NewPerformNativeLogoutBody(sessionToken)
}
// BuildPasswordLoginBody creates a login body for password authentication
func BuildPasswordLoginBody(email, password string) ory.UpdateLoginFlowBody {
	return ory.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &ory.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: email,
			Password:   password,
		},
	}
}

// BuildPasswordRegistrationBody creates a registration body for password authentication
func BuildPasswordRegistrationBody(email, password, firstName, lastName string) ory.UpdateRegistrationFlowBody {
	traits := map[string]interface{}{
		"email": email,
		"name": map[string]string{
			"first": firstName,
			"last":  lastName,
		},
	}

	return ory.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithPasswordMethod: &ory.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: password,
			Traits:   traits,
		},
	}
}

// BuildCodeVerificationBody creates a verification body for code method
func BuildCodeVerificationBodySubmit(email, code string) ory.UpdateVerificationFlowBody {
	return ory.UpdateVerificationFlowBody{
		UpdateVerificationFlowWithCodeMethod: &ory.UpdateVerificationFlowWithCodeMethod{
			Method: "code",
			Email:  &email,
			Code:   &code,
		},
	}
}

func BuildCodeVerificationBody(email string) ory.UpdateVerificationFlowBody {
	return ory.UpdateVerificationFlowBody{
		UpdateVerificationFlowWithCodeMethod: &ory.UpdateVerificationFlowWithCodeMethod{
			Method: "code",
			Email:  &email,
		},
	}
}

// CreateRecoveryFlow creates a new native recovery flow
func (k *KratosClient) CreateRecoveryFlow(ctx context.Context) (*ory.RecoveryFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateNativeRecoveryFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create recovery flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// UpdateRecoveryFlow submits recovery data (email or code with new password)
func (k *KratosClient) UpdateRecoveryFlow(ctx context.Context, flowID string, body ory.UpdateRecoveryFlowBody) (*ory.RecoveryFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.UpdateRecoveryFlow(ctx).
		Flow(flowID).
		UpdateRecoveryFlowBody(body).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to update recovery flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// BuildCodeRecoveryBody creates a recovery body for requesting recovery code
func BuildCodeRecoveryBody(email string) ory.UpdateRecoveryFlowBody {
	return ory.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithCodeMethod: &ory.UpdateRecoveryFlowWithCodeMethod{
			Method: "code",
			RecoveryAddress:  &email,
		},
	}
}

// BuildCodeRecoveryBodySubmit creates a recovery body for submitting code
// Note: With continue_transitions, password is NOT sent here
// It's sent in the subsequent settings flow
func BuildCodeRecoveryBodySubmit(code, password string) ory.UpdateRecoveryFlowBody {
	return ory.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithCodeMethod: &ory.UpdateRecoveryFlowWithCodeMethod{
			Method: "code",
			Code:   &code,
		},
	}
}

// CreateSettingsFlow creates a new native settings flow
func (k *KratosClient) CreateSettingsFlow(ctx context.Context) (*ory.SettingsFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateNativeSettingsFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create settings flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// UpdateSettingsFlow submits settings data with optional session token for authentication
func (k *KratosClient) UpdateSettingsFlow(ctx context.Context, flowID string, body ory.UpdateSettingsFlowBody, sessionToken string) (*ory.SettingsFlow, error) {
	req := k.frontend.FrontendAPI.UpdateSettingsFlow(ctx).
		Flow(flowID).
		UpdateSettingsFlowBody(body)

	// Add session token if provided (needed for privileged flows like password recovery)
	if sessionToken != "" {
		req = req.XSessionToken(sessionToken)
	}

	flow, resp, err := req.Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to update settings flow: %w (status: %d)", err, getStatusCode(resp))
	}
	return flow, nil
}

// BuildPasswordSettingsBody creates a settings body for password update
func BuildPasswordSettingsBody(password string) ory.UpdateSettingsFlowBody {
	return ory.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithPasswordMethod: &ory.UpdateSettingsFlowWithPasswordMethod{
			Method:   "password",
			Password: password,
		},
	}
}

// BuildPasswordSettingsBodyWithCSRF creates a settings body for password update with CSRF token
func BuildPasswordSettingsBodyWithCSRF(password, csrfToken string) ory.UpdateSettingsFlowBody {
	return ory.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithPasswordMethod: &ory.UpdateSettingsFlowWithPasswordMethod{
			Method:    "password",
			Password:  password,
			CsrfToken: &csrfToken,
		},
	}
}

// ExtractCSRFTokenFromSettings extracts CSRF token from a settings flow
func ExtractCSRFTokenFromSettings(flow *ory.SettingsFlow) string {
	if flow == nil || flow.Ui.Nodes == nil {
		return ""
	}

	for _, node := range flow.Ui.Nodes {
		if node.Attributes.UiNodeInputAttributes != nil {
			attrs := node.Attributes.UiNodeInputAttributes
			if attrs.Name == "csrf_token" {
				if str, ok := attrs.Value.(string); ok {
					return str
				}
			}
		}
	}
	return ""
}

// ExtractCSRFTokenFromContinueFlow extracts CSRF token from a continue settings flow
// Note: ContinueWithSettingsUiFlow doesn't expose UI nodes, so CSRF token may not be needed
// for flows returned via continue_with as they're already authenticated
func ExtractCSRFTokenFromContinueFlow(flow ory.ContinueWithSettingsUiFlow) string {
	// ContinueWithSettingsUiFlow doesn't have UI nodes exposed
	// CSRF protection is handled differently for continue_with flows
	// They're already authenticated via the session token
	return ""
}
