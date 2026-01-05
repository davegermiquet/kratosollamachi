# Go Boilerplate with Chi, LangChainGo, and Kratos

A production-ready Go boilerplate featuring:
- **Chi** - Lightweight, idiomatic HTTP router
- **LangChainGo** - LLM integration supporting Ollama, OpenAI, and other providers
- **Ory Kratos** - Cloud-native authentication and user management

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration management
├── internal/
│   ├── auth/
│   │   └── kratos.go            # Kratos client integration
│   ├── handlers/
│   │   ├── auth.go              # Authentication handlers
│   │   └── llm.go               # LLM handlers
│   ├── middleware/
│   │   └── auth.go              # Authentication middleware
│   └── langchain/
│       └── client.go            # LangChain client wrapper
├── .env.example                 # Environment variables template
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## Prerequisites

- Go 1.22 or higher
- Ory Kratos running in Kubernetes (or accessible endpoint)
- LLM provider (Ollama, OpenAI, etc.)

## Setup

### 1. Clone and Install Dependencies

```bash
# Install dependencies
go mod download
go mod tidy
```

### 2. Configure Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your configuration
# Update KRATOS_PUBLIC_URL and KRATOS_ADMIN_URL to point to your Kubernetes endpoints
# Example for Kubernetes:
# KRATOS_PUBLIC_URL=http://kratos-public.default.svc.cluster.local:4433
# KRATOS_ADMIN_URL=http://kratos-admin.default.svc.cluster.local:4434

# Configure your LLM provider:
# For Ollama:
# LLM_PROVIDER=ollama
# LLM_MODEL=llama2
# LLM_BASE_URL=http://localhost:11434
# LLM_API_KEY=

# For OpenAI:
# LLM_PROVIDER=openai
# LLM_MODEL=gpt-4
# LLM_BASE_URL=
# LLM_API_KEY=your-openai-api-key
```

### 3. Run the Application

```bash
# Development
go run cmd/server/main.go

# Build and run
go build -o server cmd/server/main.go
./server
```

The server will start on `http://localhost:8080` by default.

## API Endpoints

### Health Check

```bash
GET /health
```

### Authentication Endpoints

#### Create Login Flow
```bash
GET /auth/login
```

#### Get Login Flow
```bash
GET /auth/login/flow?flow=<flow_id>
```

#### Create Registration Flow
```bash
GET /auth/registration
```

#### Get Registration Flow
```bash
GET /auth/registration/flow?flow=<flow_id>
```

#### Get Current Session (Protected)
```bash
GET /auth/whoami
Authorization: Bearer <session_token>
```

#### Logout (Protected)
```bash
POST /auth/logout
Cookie: ory_kratos_session=<session_cookie>
```

### LLM Endpoints (All Protected)

#### Chat Completion
```bash
POST /llm/chat
Authorization: Bearer <session_token>
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

#### Text Generation
```bash
POST /llm/generate
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "prompt": "Write a story about..."
}
```

Response:
```json
{
  "content": "Once upon a time..."
}
```

## Authentication Flow

### Login Flow Example

1. Create a login flow:
```bash
curl http://localhost:8080/auth/login
```

2. This returns a flow with an ID and UI fields. Submit the form to Kratos directly (typically done by frontend)

3. After successful login, Kratos sets a session cookie

4. Use the session cookie or token for protected endpoints:
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/auth/whoami
```

### Using with Kubernetes Kratos

When running in Kubernetes alongside Kratos:

```yaml
# Example environment variables
KRATOS_PUBLIC_URL=http://kratos-public.default.svc.cluster.local:4433
KRATOS_ADMIN_URL=http://kratos-admin.default.svc.cluster.local:4434
```

If you're accessing from outside the cluster, use port-forwarding or ingress:

```bash
# Port forward Kratos public
kubectl port-forward svc/kratos-public 4433:4433

# Then use in .env
KRATOS_PUBLIC_URL=http://localhost:4433
```

## LLM Configuration

### Using Ollama

```env
LLM_PROVIDER=ollama
LLM_MODEL=llama2
LLM_BASE_URL=http://localhost:11434
LLM_API_KEY=
```

### Using OpenAI

```env
LLM_PROVIDER=openai
LLM_MODEL=gpt-4
LLM_BASE_URL=
LLM_API_KEY=sk-your-api-key-here
```

### Supported Providers

LangChainGo supports multiple LLM providers:
- **Ollama** - Local LLM inference
- **OpenAI** - GPT models
- **Anthropic** - Claude models (configure with custom base URL)
- And more providers supported by langchaingo

## Configuration

All configuration is done via environment variables (see `.env.example`):

- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment name (development/production)
- `KRATOS_PUBLIC_URL` - Kratos public API endpoint
- `KRATOS_ADMIN_URL` - Kratos admin API endpoint
- `LLM_PROVIDER` - LLM provider (ollama, openai, etc.)
- `LLM_MODEL` - Model name to use
- `LLM_BASE_URL` - Base URL for the LLM provider
- `LLM_API_KEY` - API key (if required by provider)
- `ALLOWED_ORIGINS` - Comma-separated CORS origins

## Middleware

### AuthMiddleware

Validates session and rejects unauthenticated requests:

```go
r.Group(func(r chi.Router) {
    r.Use(mw.AuthMiddleware(kratosClient))
    // Protected routes here
})
```

### OptionalAuthMiddleware

Validates session but allows unauthenticated requests:

```go
r.Group(func(r chi.Router) {
    r.Use(mw.OptionalAuthMiddleware(kratosClient))
    // Routes that work with or without auth
})
```

## Development

### Adding New Routes

Edit `cmd/server/main.go` and add routes to the Chi router:

```go
r.Get("/my-route", myHandler)
```

### Adding New Handlers

Create handlers in `internal/handlers/`:

```go
func (h *MyHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Your logic here
}
```

## Production Deployment

### Kubernetes Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      containers:
      - name: api-server
        image: your-registry/api-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: KRATOS_PUBLIC_URL
          value: "http://kratos-public.default.svc.cluster.local:4433"
        - name: KRATOS_ADMIN_URL
          value: "http://kratos-admin.default.svc.cluster.local:4434"
        - name: LLM_PROVIDER
          value: "ollama"
        - name: LLM_MODEL
          value: "llama2"
        - name: LLM_BASE_URL
          value: "http://ollama.default.svc.cluster.local:11434"
```

## License

MIT
