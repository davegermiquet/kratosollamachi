# Kratos-Chi-Ollama API

A Go REST API boilerplate with Ory Kratos authentication and LangChain LLM integration.

---

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [How It Works](#how-it-works)
4. [Setup & Installation](#setup--installation)
5. [Configuration](#configuration)
6. [API Endpoints](#api-endpoints)
7. [Code Architecture](#code-architecture)
8. [Testing](#testing)
9. [Adding New Features](#adding-new-features)

---

## Overview

This application provides:

- **Authentication** via Ory Kratos (login, registration, session management, email verification, password recovery)
- **LLM Integration** via LangChain (chat and text generation with Ollama/OpenAI)
- **Clean Architecture** with interfaces for easy testing and swapping implementations

---

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   ├── config.go                # Configuration loading & validation
│   └── config_test.go
├── pkg/
│   └── errors/
│       ├── errors.go            # Structured error types
│       └── errors_test.go
├── internal/
│   ├── auth/
│   │   ├── interfaces.go        # Auth service interfaces
│   │   └── kratos.go            # Kratos client implementation
│   ├── handlers/
│   │   ├── auth.go              # Auth HTTP handlers
│   │   ├── auth_test.go
│   │   ├── llm.go               # LLM HTTP handlers
│   │   └── llm_test.go
│   ├── langchain/
│   │   ├── interfaces.go        # LLM service interfaces
│   │   ├── client.go            # LangChain client wrapper
│   │   └── client_test.go
│   ├── middleware/
│   │   ├── auth.go              # Authentication middleware
│   │   └── auth_test.go
│   ├── response/
│   │   └── response.go          # JSON response helpers
│   └── validation/
│       ├── validation.go        # Input validation functions
│       └── validation_test.go
├── go.mod
├── Makefile
└── .env.example
```

---

## How It Works

### Request Flow

```
HTTP Request
     │
     ▼
┌─────────────────┐
│   Chi Router    │  ← Routes requests to handlers
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Middleware    │  ← AuthMiddleware validates session token
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Handler      │  ← Validates input, calls services
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Validation    │  ← Parses & validates request body
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Service      │  ← Kratos or LangChain client
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Response      │  ← Formats JSON response
└─────────────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| `cmd/server/main.go` | Initializes dependencies, sets up routes, starts server |
| `config/` | Loads environment variables, validates configuration |
| `pkg/errors/` | Provides structured error types with HTTP status codes |
| `internal/auth/` | Interfaces + Kratos client for authentication, email verification, and password recovery |
| `internal/langchain/` | Interfaces + LangChain client for LLM operations |
| `internal/handlers/` | HTTP handlers that process requests |
| `internal/middleware/` | Extracts session tokens, validates authentication |
| `internal/validation/` | Validates and parses request bodies |
| `internal/response/` | Helper functions for JSON responses |

---

## Setup & Installation

### Prerequisites

- Go 1.22+
- Ory Kratos instance running
- Ollama or OpenAI API access

### Install

```bash
# Clone the repo
git clone <your-repo-url>
cd kratos-chi-ollama

# Install dependencies
make install

# Copy environment file
cp .env.example .env

# Edit .env with your settings
nano .env

# Run the server
make run
```

### Build

```bash
make build
./server
```

---

## Configuration

Edit `.env` file:

```env
# Server
PORT=8080
ENVIRONMENT=development

# Ory Kratos
KRATOS_PUBLIC_URL=http://localhost:4433
KRATOS_ADMIN_URL=http://localhost:4434

# LLM Provider (ollama or openai)
LLM_PROVIDER=ollama
LLM_MODEL=llama2
LLM_BASE_URL=http://localhost:11434
LLM_API_KEY=

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

---

## API Endpoints

### Health Check

```
GET /health
```

Response:
```json
{"status": "healthy", "version": "1.0.0"}
```

---

### Authentication Endpoints

#### Create Login Flow

```
GET /api/v1/users/auth/login
```

Returns a Kratos login flow with form fields.

---

#### Submit Login

```
POST /api/v1/users/auth/login/flow?flow=<flow_id>
Content-Type: application/json

{
  "email": "user@example.com",
  "pass": "yourpassword"
}
```

Returns session token on success.

---

#### Create Registration Flow

```
GET /api/v1/users/auth/registration
```

Returns:
```json
{
  "flow_id": "abc123",
  "csrf_token": "xyz",
  "action": "https://...",
  "method": "POST",
  "fields": [...]
}
```

---

#### Submit Registration

```
POST /api/v1/users/auth/registration/flow?flow=<flow_id>
Content-Type: application/json

{
  "email": "newuser@example.com",
  "pass": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

---

### Password Recovery Endpoints (Public)

#### Create Recovery Flow

```
GET /api/v1/users/recovery
```

Returns a Kratos recovery flow.

---

#### Request Recovery Code

```
POST /api/v1/users/recovery/flow?flow=<flow_id>
Content-Type: application/json

{
  "email": "user@example.com"
}
```

Sends a recovery code to the user's email address.

---

#### Submit Recovery Code and New Password

```
POST /api/v1/users/recovery/code?flow=<flow_id>
Content-Type: application/json

{
  "code": "123456",
  "password": "NewSecurePassword123!"
}
```

Verifies the recovery code and resets the password to the new password. The email is not required in this step as the flow already knows which account is being recovered.

**Implementation Details:**

The password recovery uses Kratos's `continue_transitions` feature, which works as follows:

1. When you submit the recovery code with password, Kratos validates the code but doesn't directly update the password
2. Instead, Kratos returns a `continue_with` response containing:
   - A session token (via `ContinueWithSetOrySessionToken`) for authentication
   - A settings flow (via `ContinueWithSettingsUi`) where the password update actually happens
3. The API automatically processes these `continue_with` actions:
   - Extracts the session token
   - Submits the password to the settings flow using the session token for authentication
   - Completes the password update

This approach ensures proper security and allows Kratos to handle the full recovery lifecycle natively.

**Challenges Solved:**

- **Browser Redirects**: Initial implementations triggered browser-based flows. Fixed by properly configuring Kratos with `continue_transitions` and handling native flows correctly.
- **Password Not Updating**: Initially sent password in recovery flow's `transient_payload`, but with `continue_transitions`, the password must be submitted to the subsequent settings flow.
- **401 Authentication**: Settings flows require authentication. Fixed by extracting the session token from `continue_with` actions and passing it when updating the settings flow.
- **CSRF Tokens**: Learned that `continue_with` flows don't require CSRF tokens as they're authenticated via the session token.

---

### Session Management Endpoints (Protected - Require Authentication)

All session management endpoints require `X-Session-Token` header.

#### Get Current Session (Who Am I)

```
GET /api/v1/app/misc/whoami
X-Session-Token: <your-session-token>
```

Returns current user session info.

---

#### Logout

```
GET /api/v1/app/misc/logout
X-Session-Token: <your-session-token>
```

Logs out the current session by disabling it using Ory's native logout API.

Response:
```json
{"message": "Successfully logged out"}
```

---

### LLM Endpoints (Protected - Require Authentication)

All LLM endpoints require `X-Session-Token` header.

#### Chat

```
POST /api/v1/app/llm/chat
X-Session-Token: <your-session-token>
Content-Type: application/json

{
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ]
}
```

Response:
```json
{"content": "Hello! How can I help you today?"}
```

---

#### Generate Text

```
POST /api/v1/app/llm/generate
X-Session-Token: <your-session-token>
Content-Type: application/json

{
  "prompt": "Write a poem about coding"
}
```

Response:
```json
{"content": "Lines of code..."}
```

---

### Email Verification Endpoints (Protected - Require Authentication)

All verification endpoints require `X-Session-Token` header.

#### Create Verification Flow

```
GET /api/v1/users/verification
X-Session-Token: <your-session-token>
```

Returns a Kratos verification flow.

---

#### Request Verification Email

```
POST /api/v1/users/verification/flow?flow=<flow_id>
X-Session-Token: <your-session-token>
Content-Type: application/json

{
  "email": "user@example.com"
}
```

Sends a verification email/link to the logged-in user.

---

#### Submit Verification Code

```
POST /api/v1/users/verification/code?flow=<flow_id>
X-Session-Token: <your-session-token>
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

Verifies the email using the provided code.

---

## Code Architecture

### Interfaces (for testability)

**Auth interfaces** (`internal/auth/interfaces.go`):

```go
type SessionValidator interface {
    ValidateSession(ctx context.Context, token string) (*ory.Session, error)
}

type VerificationFlowManager interface {
    CreateVerificationFlow(ctx context.Context) (*ory.VerificationFlow, error)
    UpdateVerificationFlow(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error)
}

type KratosService interface {
    SessionValidator
    LoginFlowManager
    RegistrationFlowManager
    LogoutFlowManager
    VerificationFlowManager
}
```

**LLM interface** (`internal/langchain/interfaces.go`):

```go
type LLMService interface {
    GenerateContent(ctx context.Context, prompt string, ...) (string, error)
    Chat(ctx context.Context, messages []llms.MessageContent, ...) (string, error)
}
```

### Error Handling

All errors use structured types (`pkg/errors/errors.go`):

```go
// Error response format
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required",
    "details": ""
  }
}
```

Error codes:
- `VALIDATION_ERROR` (400)
- `BAD_REQUEST` (400)
- `UNAUTHORIZED` (401)
- `NOT_FOUND` (404)
- `INTERNAL_ERROR` (500)
- `SERVICE_UNAVAILABLE` (503)

### Input Validation

All input validation in `internal/validation/validation.go`:

```go
// Example usage in handler
input, err := validation.ValidateLoginInput(r.Body)
if err != nil {
    err.WriteJSON(w)  // Returns structured error
    return
}
```

Validates:
- Email format
- Password length (6+ for login, 8+ for registration)
- Required fields
- JSON parsing

---

## Testing

### Run All Tests

```bash
make test
```

### Run with Verbose Output

```bash
make test-verbose
```

### Run with Coverage

```bash
make coverage
# Opens coverage.html report
```

### Run Specific Package

```bash
go test -v ./internal/handlers/...
go test -v ./internal/validation/...
```

### Run Integration Tests

Integration tests require a running Ory Kratos instance. They test the complete authentication flow end-to-end.

```bash
# Set environment variables
export INTEGRATION_TEST=true

# Run integration tests
go test -v -tags=integration -run TestUserAuthenticationFlow

# Or with timeout
go test -v -tags=integration -timeout 30s -run TestUserAuthenticationFlow
```

**Integration Test Flow:**
1. Creates a registration flow
2. Registers a new test user
3. Creates a login flow
4. Logs in with test credentials
5. Calls whoami to verify session
6. Logs out the session
7. Verifies session is invalid after logout

### Test Structure

Each package has mock implementations:

```go
// Example mock for testing
type MockKratosService struct {
    ValidateSessionFunc func(ctx context.Context, token string) (*ory.Session, error)
    // ...
}

func (m *MockKratosService) ValidateSession(ctx context.Context, token string) (*ory.Session, error) {
    if m.ValidateSessionFunc != nil {
        return m.ValidateSessionFunc(ctx, token)
    }
    return nil, errors.New("not implemented")
}
```

---

## Adding New Features

### 1. Add a New Handler

```go
// internal/handlers/myhandler.go
package handlers

type MyHandler struct {
    service MyService
}

func NewMyHandler(service MyService) *MyHandler {
    return &MyHandler{service: service}
}

func (h *MyHandler) DoSomething(w http.ResponseWriter, r *http.Request) {
    // 1. Validate input
    input, err := validation.ValidateMyInput(r.Body)
    if err != nil {
        err.WriteJSON(w)
        return
    }
    
    // 2. Call service
    result, svcErr := h.service.DoSomething(r.Context(), input)
    if svcErr != nil {
        apperrors.NewInternalError("failed", svcErr).WriteJSON(w)
        return
    }
    
    // 3. Return response
    response.Success(w, result)
}
```

### 2. Add Route in main.go

```go
// cmd/server/main.go
myHandler := handlers.NewMyHandler(myService)

r.Route("/my-resource", func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(kratosClient))  // Protected
    r.Post("/", myHandler.DoSomething)
})
```

### 3. Add Validation

```go
// internal/validation/validation.go
func ValidateMyInput(body io.Reader) (*MyInput, *apperrors.AppError) {
    var req struct {
        Field string `json:"field"`
    }
    
    if err := json.NewDecoder(body).Decode(&req); err != nil {
        return nil, apperrors.NewValidationError("Invalid JSON", err.Error())
    }
    
    if req.Field == "" {
        return nil, apperrors.NewValidationError("field is required", "")
    }
    
    return &MyInput{Field: req.Field}, nil
}
```

### 4. Add Tests

```go
// internal/handlers/myhandler_test.go
func TestMyHandler_DoSomething(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        wantStatus int
    }{
        {
            name:       "success",
            body:       `{"field": "value"}`,
            wantStatus: http.StatusOK,
        },
        {
            name:       "missing field",
            body:       `{}`,
            wantStatus: http.StatusBadRequest,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation
        })
    }
}
```

---

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the application |
| `make build` | Build binary |
| `make test` | Run unit tests |
| `make test-verbose` | Run tests with output |
| `make test-integration` | Run integration tests (requires Kratos) |
| `make coverage` | Generate coverage report |
| `make clean` | Remove build artifacts |
| `make install` | Install dependencies |

---

## License

MIT
