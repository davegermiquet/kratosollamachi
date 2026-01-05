package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	ory "github.com/ory/client-go"
)

type KratosClient struct {
	frontend *ory.APIClient
	admin    *ory.APIClient
}

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

// ToSession validates a session cookie or token and returns the session
func (k *KratosClient) ToSession(ctx context.Context, cookie string, token string) (*ory.Session, error) {
	req := k.frontend.FrontendAPI.ToSession(ctx)

	if cookie != "" {
		req = req.Cookie(cookie)
	}
	if token != "" {
		req = req.XSessionToken(token)
	}

	session, resp, err := req.Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("unauthorized: %w", err)
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return session, nil
}

// CreateLoginFlow creates a new login flow
func (k *KratosClient) CreateLoginFlow(ctx context.Context, refresh bool) (*ory.LoginFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateBrowserLoginFlow(ctx).
		Refresh(refresh).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to create login flow: %w (status: %d)", err, resp.StatusCode)
	}

	return flow, nil
}

// CreateRegistrationFlow creates a new registration flow
func (k *KratosClient) CreateRegistrationFlow(ctx context.Context) (*ory.RegistrationFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to create registration flow: %w (status: %d)", err, resp.StatusCode)
	}

	return flow, nil
}

// GetLoginFlow retrieves an existing login flow
func (k *KratosClient) GetLoginFlow(ctx context.Context, flowID string) (*ory.LoginFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.GetLoginFlow(ctx).
		Id(flowID).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get login flow: %w (status: %d)", err, resp.StatusCode)
	}

	return flow, nil
}

// GetRegistrationFlow retrieves an existing registration flow
func (k *KratosClient) GetRegistrationFlow(ctx context.Context, flowID string,updateRegistrationFlowBody ory.UpdateRegistrationFlowBody) (*ory.SuccessfulNativeRegistration, error) {
	fmt.Println(flowID)
	resp, r , err:=	k.frontend.FrontendAPI.UpdateRegistrationFlow(context.Background()).Flow(flowID).UpdateRegistrationFlowBody(updateRegistrationFlowBody).Execute()
		if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FrontendAPI.UpdateRegistrationFlow``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateRegistrationFlow`: SuccessfulNativeRegistration
	fmt.Fprintf(os.Stdout, "Response from `FrontendAPI.UpdateRegistrationFlow`: %v\n", resp)
	return resp, nil
	//return flow, nil
}

// CreateLogoutFlow creates a logout flow
func (k *KratosClient) CreateLogoutFlow(ctx context.Context, cookie string) (*ory.LogoutFlow, error) {
	flow, resp, err := k.frontend.FrontendAPI.CreateBrowserLogoutFlow(ctx).
		Cookie(cookie).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to create logout flow: %w (status: %d)", err, resp.StatusCode)
	}

	return flow, nil
}
