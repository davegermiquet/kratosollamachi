# Go Boilerplate with Chi, LangChainGo, and Kratos

A production-ready Go boilerplate featuring clean architecture, comprehensive testing, and best practices.

## Features

- **Chi** - Lightweight, idiomatic HTTP router
- **LangChainGo** - LLM integration supporting Ollama, OpenAI, and other providers
- **Ory Kratos** - Cloud-native authentication and user management
- **Clean Architecture** - Separation of concerns with interfaces for testability
- **Comprehensive Tests** - Unit tests with mocks for all components
- **Structured Errors** - Consistent error handling across the application

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go                    # Application entry point
├── config/
│   ├── config.go                      # Configuration management
│   └── config_test.go                 # Config tests
├── internal/
│   ├── auth/
│   │   ├── interfaces.go              # Auth service interfaces
│   │   └── kratos.go                  # Kratos client implementation
│   ├── handlers/
│   │   ├── auth.go                    # Authentication handlers
│   │   ├── auth_test.go               # Auth handler tests
│   │   ├── llm.go                     # LLM handlers
│   │   └── llm_test.go                # LLM handler tests
│   ├── langchain/
│   │   ├── interfaces.go              # LLM service interfaces
│   │   ├── client.go                  # LangChain client wrapper
│   │   └── client_test.go             # LangChain tests
│   ├── middleware/
│   │   ├── auth.go                    # Authentication middleware
│   │   └── auth_test.go               # Middleware tests
│   ├── response/
│   │   └── response.go                # Response helpers
│   └── validation/
│       ├── validation.go              # Input validation
│       └── validation_test.go         # Validation tests
├── pkg/
│   └── errors/
│       ├── errors.go                  # Structured error handling
│       └── errors_test.go             # Error tests
├── .env.example                       # Environment variables template
├── go.mod                             # Go module definition
├── Makefile                           # Build automation
└── README.md                          # This file
```

## Architecture

### Design Principles

1. **Interface-Based Design**: All external dependencies are abstracted behind interfaces, enabling easy testing and swapping implementations.

2. **Separation of Concerns**:
   - `handlers/` - HTTP request/response handling
   - `validation/` - Input validation logic
   - `response/` - Standardized response formatting
   - `auth/` - Authentication logic
   - `langchain/` - LLM integration

3. **Structured Errors**: All errors are typed with codes, messages, and appropriate HTTP status codes.

### Key Interfaces

```go
// SessionValidator - validates user sessions
type SessionValidator interface {
    ValidateSession(ctx context.Context, sessionToken string) (*ory.Session, error)
}

// LLMService - LLM operations
type LLMService interface {
    GenerateContent(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error)
    Chat(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error)
}
```

## Quick Start

### Prerequisites

- Go 1.22 or higher
- Ory Kratos running (or accessible endpoint)
- LLM provider (Ollama, OpenAI, etc.)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd <project-directory>

# Install dependencies
make install

# Copy environment template
cp .env.example .env

# Edit .env with your configuration
```

### Running

```bash
# Development
make run

# With hot reload (requires air)
make dev

# Build and run
make build
./server
```

### Testing

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run with race detector
make test-race

# Generate coverage report
make coverage
```

## API Endpoints

### Health Check

```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### Authentication

#### Create Login Flow
```bash
GET /auth/login
```

#### Submit Login
```bash
POST /auth/login/flow?flow=<flow_id>
Content-Type: application/json

{
  "email": "user@example.com",
  "pass": "password123"
}
```

#### Create Registration Flow
```bash
GET /auth/registration
```

#### Submit Registration
```bash
POST /auth/registration/flow?flow=<flow_id>
Content-Type: application/json

{
  "email": "user@example.com",
  "pass": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Get Current Session
```bash
GET /auth/whoami
X-Session-Token: <session_token>
```

#### Logout
```bash
POST /auth/logout
Cookie: ory_kratos_session=<session_cookie>
```

### LLM Endpoints (Protected)

All LLM endpoints require authentication via `X-Session-Token` header.

#### Chat
```bash
POST /llm/chat
X-Session-Token: <session_token>
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
{
  "content": "Hello! How can I help you today?"
}
```

#### Generate
```bash
POST /llm/generate
X-Session-Token: <session_token>
Content-Type: application/json

{
  "prompt": "Write a haiku about coding"
}
```

Response:
```json
{
  "content": "Lines of code cascade\nBugs emerge then disappear\nShip it, call it done"
}
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required",
    "details": ""
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Input validation failed |
| `BAD_REQUEST` | 400 | Malformed request |
| `UNAUTHORIZED` | 401 | Authentication required or failed |
| `NOT_FOUND` | 404 | Resource not found |
| `INTERNAL_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | External service unavailable |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `ENVIRONMENT` | development | Environment (development/production) |
| `KRATOS_PUBLIC_URL` | http://localhost:4433 | Kratos public API |
| `KRATOS_ADMIN_URL` | http://localhost:4434 | Kratos admin API |
| `LLM_PROVIDER` | ollama | LLM provider (ollama/openai) |
| `LLM_MODEL` | llama2 | Model name |
| `LLM_BASE_URL` | http://localhost:11434 | LLM API base URL |
| `LLM_API_KEY` | | API key (for OpenAI) |
| `ALLOWED_ORIGINS` | http://localhost:3000 | CORS allowed origins |

## Testing Strategy

### Unit Tests

Each package has corresponding `_test.go` files with:

- **Table-driven tests** for comprehensive coverage
- **Mock implementations** for external dependencies
- **Edge case handling** validation

### Running Specific Tests

```bash
# Test a specific package
go test -v ./internal/handlers/...

# Test a specific function
go test -v -run TestAuthHandler_CreateLoginFlow ./internal/handlers/

# Test with coverage for specific package
go test -coverprofile=coverage.out ./internal/validation/
```

## Extending the Application

### Adding a New Handler

1. Define the handler struct and constructor:
```go
type MyHandler struct {
    service MyService
}

func NewMyHandler(service MyService) *MyHandler {
    return &MyHandler{service: service}
}
```

2. Implement handler methods:
```go
func (h *MyHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Validate input
    input, err := validation.ValidateMyInput(r.Body)
    if err != nil {
        err.WriteJSON(w)
        return
    }
    
    // Call service
    result, svcErr := h.service.DoSomething(r.Context(), input)
    if svcErr != nil {
        apperrors.NewInternalError("operation failed", svcErr).WriteJSON(w)
        return
    }
    
    response.Success(w, result)
}
```

3. Add routes in `main.go`:
```go
r.Route("/my-resource", func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(kratosClient))
    r.Post("/", myHandler.HandleRequest)
})
```

4. Write tests with mocks.

## License

MIT
