# Email Verification Implementation

This document describes the email verification functionality implemented using Ory Kratos's native verification flow API.

## Overview

Email verification allows authenticated users to verify their email addresses using either:
- **Link Method**: User receives an email with a verification link
- **Code Method**: User receives an email with a verification code to enter manually

This implementation uses Kratos's native verification flow, which is designed for API clients (mobile apps, SPAs, etc.) and does not rely on cookies.

**Important**: All verification endpoints require authentication via `X-Session-Token` header. Users must be logged in to verify their email.

---

## Implementation Summary

### Files Modified/Created

| File | Status | Description |
|------|--------|-------------|
| `internal/auth/interfaces.go` | Modified | Added `VerificationFlowManager` interface |
| `internal/auth/kratos.go` | Modified | Implemented verification methods and helper functions |
| `internal/validation/validation.go` | Modified | Added validation for verification inputs |
| `internal/validation/validation_test.go` | Modified | Added 17 test cases for verification validation |
| `internal/handlers/auth.go` | Modified | Added 3 verification handler functions |
| `internal/handlers/auth_test.go` | Modified | Added 18 test cases for verification handlers |
| `cmd/server/main.go` | Modified | Added 3 verification routes |
| `README.md` | Modified | Updated with verification endpoints and interfaces |
| `CLAUDE.md` | Modified | Updated with verification information |

---

## Architecture Changes

### 1. Interfaces (`internal/auth/interfaces.go`)

Added new `VerificationFlowManager` interface:

```go
type VerificationFlowManager interface {
    CreateVerificationFlow(ctx context.Context) (*ory.VerificationFlow, error)
    UpdateVerificationFlow(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error)
}
```

Updated `KratosService` to include verification:

```go
type KratosService interface {
    SessionValidator
    LoginFlowManager
    RegistrationFlowManager
    LogoutFlowManager
    VerificationFlowManager  // NEW
}
```

### 2. Kratos Client Implementation (`internal/auth/kratos.go`)

**Verification Flow Methods:**

```go
// CreateVerificationFlow creates a new native verification flow
func (k *KratosClient) CreateVerificationFlow(ctx context.Context) (*ory.VerificationFlow, error)

// UpdateVerificationFlow submits verification data (email or code)
func (k *KratosClient) UpdateVerificationFlow(ctx context.Context, flowID string, body ory.UpdateVerificationFlowBody) (*ory.VerificationFlow, error)
```

**Helper Functions:**

```go
// BuildCodeVerificationBody creates a verification body for code method
func BuildCodeVerificationBody(email, code string) ory.UpdateVerificationFlowBody

// BuildLinkVerificationBody creates a verification body for link method
func BuildLinkVerificationBody(email string) ory.UpdateVerificationFlowBody
```

### 3. Validation (`internal/validation/validation.go`)

**New Input Types:**

```go
type VerificationEmailInput struct {
    Email string
}

type VerificationCodeInput struct {
    Email string
    Code  string
}
```

**Validation Functions:**

```go
// ValidateVerificationEmailInput validates verification email request
func ValidateVerificationEmailInput(body io.Reader) (*VerificationEmailInput, *apperrors.AppError)

// ValidateVerificationCodeInput validates verification code submission
func ValidateVerificationCodeInput(body io.Reader) (*VerificationCodeInput, *apperrors.AppError)
```

**Validation Rules:**
- Email format validation using `net/mail` package
- Email required and non-empty after trimming
- Code required and minimum 6 characters
- Invalid JSON handling
- Whitespace trimming on all inputs

### 4. HTTP Handlers (`internal/handlers/auth.go`)

**CreateVerificationFlow** - `GET /api/v1/users/verification` (Protected)
```go
func (h *AuthHandler) CreateVerificationFlow(w http.ResponseWriter, r *http.Request)
```
- **Requires**: `X-Session-Token` header (enforced by AuthMiddleware)
- Creates a new verification flow via Kratos
- Returns flow object with flow ID
- Returns 503 if Kratos is unavailable
- Returns 401 if session token is invalid/missing

**RequestVerificationEmail** - `POST /api/v1/users/verification/flow?flow={id}` (Protected)
```go
func (h *AuthHandler) RequestVerificationEmail(w http.ResponseWriter, r *http.Request)
```
- **Requires**: `X-Session-Token` header (enforced by AuthMiddleware)
- Validates flow ID from query parameter
- Validates email input from request body
- Sends verification email/link via Kratos
- Returns 400 for validation errors
- Returns 401 if session token is invalid/missing
- Returns 500 for Kratos errors

**SubmitVerificationCode** - `POST /api/v1/users/verification/code?flow={id}` (Protected)
```go
func (h *AuthHandler) SubmitVerificationCode(w http.ResponseWriter, r *http.Request)
```
- **Requires**: `X-Session-Token` header (enforced by AuthMiddleware)
- Validates flow ID from query parameter
- Validates email and code from request body
- Submits code for verification via Kratos
- Returns 400 for invalid/expired codes or validation errors
- Returns 401 if session token is invalid/missing

### 5. Routes (`cmd/server/main.go`)

Added to the `/api/v1/users/verification` route group (protected, requires authentication):

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Route("/users", func(r chi.Router) {
        // Protected verification routes
        r.Route("/verification", func(r chi.Router) {
            r.Use(middleware.AuthMiddleware(kratosClient))  // Requires X-Session-Token
            r.Get("/", authHandler.CreateVerificationFlow)
            r.Post("/flow", authHandler.RequestVerificationEmail)
            r.Post("/code", authHandler.SubmitVerificationCode)
        })
    })
})
```

**Notes**:
- Verification endpoints are under `/api/v1/users/verification` to organize user-related operations
- All endpoints are protected because users must be authenticated to verify their email address
- API versioning with `/api/v1` prefix allows for future API changes

---

## API Usage Flow

### Flow 1: Link-Based Verification

**Prerequisite**: User must be authenticated (have a valid session token from login)

**Step 1: Create Verification Flow**
```bash
GET /api/v1/users/verification
X-Session-Token: <your-session-token>
```

Response:
```json
{
  "id": "verification-flow-id",
  "type": "api",
  "ui": { ... },
  "state": "choose_method"
}
```

**Step 2: Request Verification Email with Link**
```bash
POST /api/v1/users/verification/flow?flow=verification-flow-id
X-Session-Token: <your-session-token>
Content-Type: application/json

{
  "email": "user@example.com"
}
```

Response:
```json
{
  "id": "verification-flow-id",
  "state": "sent_email",
  "ui": { ... }
}
```

User receives email with verification link and clicks it to verify.

### Flow 2: Code-Based Verification

**Prerequisite**: User must be authenticated (have a valid session token from login)

**Step 1: Create Verification Flow**
```bash
GET /api/v1/users/verification
X-Session-Token: <your-session-token>
```

**Step 2: Request Verification Code**
```bash
POST /api/v1/users/verification/flow?flow=verification-flow-id
X-Session-Token: <your-session-token>
Content-Type: application/json

{
  "email": "user@example.com"
}
```

User receives email with verification code (e.g., "123456").

**Step 3: Submit Verification Code**
```bash
POST /api/v1/users/verification/code?flow=verification-flow-id
X-Session-Token: <your-session-token>
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

Response on success:
```json
{
  "id": "verification-flow-id",
  "state": "passed_challenge",
  "ui": { ... }
}
```

---

## Error Handling

All verification endpoints follow the existing error handling pattern:

**Unauthorized (401 Unauthorized):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid or expired session",
    "details": ""
  }
}
```
**Note**: This error is returned if the `X-Session-Token` header is missing or invalid (enforced by AuthMiddleware).

**Validation Errors (400 Bad Request):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required",
    "details": ""
  }
}
```

**Invalid/Expired Code (400 Bad Request):**
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid or expired verification code",
    "details": ""
  }
}
```

**Kratos Unavailable (503 Service Unavailable):**
```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Kratos is unavailable",
    "details": "..."
  }
}
```

---

## Testing

### Validation Tests (17 test cases)

File: `internal/validation/validation_test.go`

**TestValidateVerificationEmailInput:**
- Valid email
- Missing email
- Empty email
- Whitespace email
- Invalid email format
- Invalid JSON
- Email with whitespace trimming

**TestValidateVerificationCodeInput:**
- Valid input
- Missing email
- Empty email
- Invalid email format
- Missing code
- Empty code
- Code too short (< 6 characters)
- Whitespace code
- Invalid JSON
- Valid with whitespace trimming

### Handler Tests (18 test cases)

File: `internal/handlers/auth_test.go`

**TestAuthHandler_CreateVerificationFlow:**
- Success case
- Kratos error (503)

**TestAuthHandler_RequestVerificationEmail:**
- Success case
- Missing flow ID
- Invalid JSON
- Missing email
- Invalid email format
- Kratos error

**TestAuthHandler_SubmitVerificationCode:**
- Success case
- Missing flow ID
- Invalid JSON
- Missing email
- Missing code
- Code too short
- Invalid email format
- Kratos error - invalid code

All tests use mock implementations of `KratosService` to avoid dependencies on external services.

---

## Configuration

No additional configuration is required. The verification functionality uses the existing Kratos configuration:

- `KRATOS_PUBLIC_URL`: Used for all verification API calls
- Kratos must be configured with verification enabled (configured in Kratos itself, not this application)

---

## Key Design Patterns

### 1. Interface-Based Design
- `VerificationFlowManager` interface allows for easy mocking in tests
- Handlers depend on interface, not concrete implementation

### 2. Validation Layer
- All input validation centralized in `internal/validation`
- Consistent error responses
- No type assertions without checks

### 3. Helper Functions
- `BuildCodeVerificationBody()` and `BuildLinkVerificationBody()` encapsulate body construction
- Prevents duplication and ensures correct structure

### 4. Handler Pattern
All verification handlers follow the established pattern:
1. Validate URL parameters (flow ID)
2. Validate request body
3. Call service layer
4. Return structured response or error

### 5. Table-Driven Tests
- Each test function uses table-driven approach
- Tests cover success paths, validation errors, and service errors

---

## Integration with Existing Architecture

The verification feature integrates seamlessly with existing patterns:

- **Error Handling**: Uses `pkg/errors/errors.go` for structured errors
- **Validation**: Uses `internal/validation/validation.go` pattern
- **Response Formatting**: Uses `internal/response/response.go`
- **Mock Testing**: Extends existing mock pattern in test files
- **Route Organization**: Follows `/api/v1` versioned structure with domain-based organization (user operations under `/users/*`)
- **Authentication**: Uses `middleware.AuthMiddleware` to enforce session-based authentication
- **API Versioning**: Part of the `/api/v1` prefix for forward compatibility

---

## References

- [Ory Kratos GitHub](https://github.com/ory/kratos)
- [Ory Email Verification Documentation](https://www.ory.com/docs/kratos/self-service/flows/verify-email-account-activation)
- [Ory Kratos Self-Service Flows](https://www.ory.sh/docs/kratos/self-service)

---

## Future Enhancements

Potential improvements to consider:

1. **Resend Verification Email**: Add endpoint to resend verification email if code/link expires
2. **Verification Status Check**: Add endpoint to check if email is already verified
3. **Rate Limiting**: Implement rate limiting on verification requests to prevent abuse
4. **Metrics**: Add metrics/logging for verification attempts and success rates
5. **Custom Templates**: Configure custom email templates in Kratos for verification emails
