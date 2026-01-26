# Go Monolith Boilerplate with Protokit Framework

A monolithic Go application using **Go Fiber** for REST, **gRPC** for service-to-service communication, Domain-Driven Design (DDD), Clean Architecture, and **Protokit** framework with Protocol Buffers as the source of truth.

---

## Protokit Framework

**Protokit** (`github.com/solvr-asia/protokit`) is a Contract-Driven Development framework that:
- Uses Proto files as the **single source of truth** for API contracts
- Generates **Go Fiber handlers** for REST and **gRPC handlers** for service-to-service
- Provides built-in **middleware** for auth, logging, recovery, and request tracking
- Auto-generates **validation**, **auth configs**, and **client code** from proto annotations

### Protokit Repository Structure

```
github.com/solvr-asia/protokit/
├── cmd/
│   ├── protokit/                    # CLI tool
│   │   └── main.go
│   └── protoc-gen-protokit/         # Protoc plugin
│       └── main.go
├── fiber/                           # Fiber server & middleware
│   ├── server.go                    # Fiber + gRPC server
│   ├── handler.go                   # Handler wrapper
│   ├── response.go                  # Response helpers
│   ├── middleware/
│   │   ├── auth.go                  # Auth middleware
│   │   ├── logger.go                # Logger middleware
│   │   ├── recovery.go              # Recovery middleware
│   │   └── requestid.go             # Request ID middleware
│   └── errors/
│       └── errors.go                # Error handling
├── grpc/                            # gRPC components
│   ├── server.go                    # gRPC server
│   ├── interceptor/
│   │   ├── auth.go                  # Auth interceptor
│   │   ├── logger.go                # Logger interceptor
│   │   ├── recovery.go              # Recovery interceptor
│   │   └── requestid.go             # Request ID interceptor
│   └── client/
│       └── client.go                # Client helpers
├── validation/
│   ├── validator.go                 # Validation engine
│   └── rules.go                     # Built-in rules
├── proto/
│   └── protokit/
│       └── api/
│           └── extensions.proto     # Proto extensions
├── generator/                       # Code generator
│   ├── generator.go
│   └── templates/
│       ├── fiber_handler.go.tmpl    # Fiber handler template
│       ├── grpc_handler.go.tmpl     # gRPC handler template
│       ├── validation.go.tmpl
│       └── client.go.tmpl
├── examples/
│   └── simple-api/
├── go.mod
└── README.md
```

---

## Protokit Proto Extensions

### extensions.proto

```protobuf
// proto/protokit/api/extensions.proto
syntax = "proto3";
package protokit.api;

option go_package = "github.com/solvr-asia/protokit/proto/protokit/api";

import "google/protobuf/descriptor.proto";

// Method-level authentication options
message AuthOption {
    bool need_auth = 1;              // Requires authentication
    repeated string roles = 2;        // Allowed roles (empty = any authenticated user)
}

// Extend method options for auth
extend google.protobuf.MethodOptions {
    AuthOption auth = 50001;
}

// Extend field options for validation and defaults
extend google.protobuf.FieldOptions {
    string validate = 50002;          // (protokit.api.validate) = "required|email"
    string default = 50003;           // (protokit.api.default) = "10"
}
```

### Using Extensions in Proto Files

```protobuf
// contract/user.proto
syntax = "proto3";
package user;
option go_package = ".;user";

import "google/protobuf/empty.proto";
import "protokit/api/extensions.proto";

service UserApi {
    // Public endpoint - no auth required
    rpc Register(RegisterReq) returns (RegisterRes) {
        // No protokit.api.auth = public
    }

    // Protected endpoint - requires authentication
    rpc GetProfile(google.protobuf.Empty) returns (UserProfile) {
        option (protokit.api.auth) = { need_auth: true };
    }

    // Role-based access - admin only
    rpc ListAllUsers(ListUsersReq) returns (ListUsersRes) {
        option (protokit.api.auth) = { need_auth: true, roles: ["admin", "superadmin"] };
    }

    // Employee-only endpoint
    rpc GetMyPayslip(GetPayslipReq) returns (Payslip) {
        option (protokit.api.auth) = { need_auth: true, roles: ["employee"] };
    }
}

message RegisterReq {
    string email = 1 [json_name = "email", (protokit.api.validate) = "required|email"];
    string password = 2 [json_name = "password", (protokit.api.validate) = "required|min=8|max=128"];
    string name = 3 [json_name = "name", (protokit.api.validate) = "required|min=2|max=100"];
    string phone = 4 [json_name = "phone", (protokit.api.validate) = "phone"];
}

message RegisterRes {
    string id = 1 [json_name = "id"];
    string email = 2 [json_name = "email"];
    string name = 3 [json_name = "name"];
}

message UserProfile {
    string id = 1 [json_name = "id"];
    string email = 2 [json_name = "email"];
    string name = 3 [json_name = "name"];
    string phone = 4 [json_name = "phone"];
    string status = 5 [json_name = "status"];
    string created_at = 6 [json_name = "createdAt"];
}

message ListUsersReq {
    int32 page = 1 [json_name = "page", (protokit.api.default) = "1"];
    int32 size = 2 [json_name = "size", (protokit.api.default) = "10"];
    string search = 3 [json_name = "search"];
    string sort_by = 4 [json_name = "sortBy", (protokit.api.default) = "created_at"];
    string sort_order = 5 [json_name = "sortOrder", (protokit.api.default) = "desc"];
}

message ListUsersRes {
    repeated UserProfile users = 1 [json_name = "users"];
    Pagination pagination = 2 [json_name = "pagination"];
}

message Pagination {
    int32 page = 1 [json_name = "page"];
    int32 size = 2 [json_name = "size"];
    int64 total = 3 [json_name = "total"];
    int32 total_pages = 4 [json_name = "totalPages"];
}

message GetPayslipReq {
    int32 year = 1 [json_name = "year", (protokit.api.validate) = "required|min=2000|max=2100"];
    int32 month = 2 [json_name = "month", (protokit.api.validate) = "required|min=1|max=12"];
}

message Payslip {
    string id = 1 [json_name = "id"];
    string employee_id = 2 [json_name = "employeeId"];
    int32 year = 3 [json_name = "year"];
    int32 month = 4 [json_name = "month"];
    double gross_salary = 5 [json_name = "grossSalary"];
    double net_salary = 6 [json_name = "netSalary"];
}
```

---

## Protokit Validation Rules

| Rule | Description | Example |
|------|-------------|---------|
| `required` | Field must not be empty | `required` |
| `email` | Valid email format | `email` |
| `min=N` | Minimum length/value | `min=8` |
| `max=N` | Maximum length/value | `max=100` |
| `len=N` | Exact length | `len=6` |
| `regex=PATTERN` | Regex match | `regex=^[A-Z]+$` |
| `phone` | Valid phone number | `phone` |
| `url` | Valid URL | `url` |
| `uuid` | Valid UUID | `uuid` |
| `date` | Date format (YYYY-MM-DD) | `date` |
| `datetime` | DateTime format (RFC3339) | `datetime` |
| `time` | Time format (HH:MM:SS) | `time` |
| `in=a,b,c` | Value in list | `in=active,inactive` |
| `notin=a,b,c` | Value not in list | `notin=deleted` |
| `numeric` | Numeric string | `numeric` |
| `alpha` | Alphabetic only | `alpha` |
| `alphanum` | Alphanumeric only | `alphanum` |

Multiple rules: `required|email|max=100`

---

## Protokit Fiber Server

### Server Implementation

```go
// github.com/solvr-asia/protokit/fiber/server.go
package fiber

import (
    "context"
    "fmt"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "google.golang.org/grpc"
)

// Server holds both Fiber and gRPC servers
type Server struct {
    fiber        *fiber.App
    grpc         *grpc.Server
    httpPort     int
    grpcPort     int
    interceptors []grpc.UnaryServerInterceptor
}

// Config for server
type Config struct {
    AppName     string
    HTTPPort    int
    GRPCPort    int
    EnableCORS  bool
    CORSOrigins string
}

// NewServer creates a new Protokit server with Fiber + gRPC
func NewServer(cfg Config, opts ...ServerOption) *Server {
    s := &Server{
        httpPort: cfg.HTTPPort,
        grpcPort: cfg.GRPCPort,
    }

    // Apply options
    for _, opt := range opts {
        opt(s)
    }

    // Create Fiber app
    s.fiber = fiber.New(fiber.Config{
        AppName:               cfg.AppName,
        DisableStartupMessage: false,
        ErrorHandler:          defaultErrorHandler,
    })

    // Enable CORS if configured
    if cfg.EnableCORS {
        s.fiber.Use(cors.New(cors.Config{
            AllowOrigins:     cfg.CORSOrigins,
            AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
            AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
            AllowCredentials: true,
        }))
    }

    // Create gRPC server with interceptors
    s.grpc = grpc.NewServer(
        grpc.ChainUnaryInterceptor(s.interceptors...),
    )

    return s
}

// ServerOption configures the server
type ServerOption func(*Server)

// WithGRPCInterceptor adds a gRPC interceptor
func WithGRPCInterceptor(interceptor grpc.UnaryServerInterceptor) ServerOption {
    return func(s *Server) {
        s.interceptors = append(s.interceptors, interceptor)
    }
}

// Fiber returns the Fiber app for adding routes and middleware
func (s *Server) Fiber() *fiber.App {
    return s.fiber
}

// GRPC returns the gRPC server for registering services
func (s *Server) GRPC() *grpc.Server {
    return s.grpc
}

// Listen starts both servers
func (s *Server) Listen() error {
    errChan := make(chan error, 2)

    // Start gRPC server
    go func() {
        lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
        if err != nil {
            errChan <- fmt.Errorf("gRPC listen error: %w", err)
            return
        }
        if err := s.grpc.Serve(lis); err != nil {
            errChan <- fmt.Errorf("gRPC serve error: %w", err)
        }
    }()

    // Start Fiber server
    go func() {
        if err := s.fiber.Listen(fmt.Sprintf(":%d", s.httpPort)); err != nil {
            errChan <- fmt.Errorf("Fiber listen error: %w", err)
        }
    }()

    return <-errChan
}

// ListenGraceful starts servers with graceful shutdown
func (s *Server) ListenGraceful(shutdownTimeout time.Duration) error {
    errChan := make(chan error, 2)

    // Start gRPC server
    go func() {
        lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
        if err != nil {
            errChan <- err
            return
        }
        if err := s.grpc.Serve(lis); err != nil {
            errChan <- err
        }
    }()

    // Start Fiber server
    go func() {
        if err := s.fiber.Listen(fmt.Sprintf(":%d", s.httpPort)); err != nil {
            errChan <- err
        }
    }()

    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case <-quit:
        ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
        defer cancel()

        // Graceful shutdown
        s.grpc.GracefulStop()
        return s.fiber.ShutdownWithContext(ctx)
    case err := <-errChan:
        return err
    }
}

// Default error handler for Fiber
func defaultErrorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    message := "Internal Server Error"

    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
        message = e.Message
    }

    return c.Status(code).JSON(fiber.Map{
        "success": false,
        "error": fiber.Map{
            "code":    code,
            "message": message,
        },
    })
}
```

### Response Helpers

```go
// github.com/solvr-asia/protokit/fiber/response.go
package fiber

import (
    "github.com/gofiber/fiber/v2"
)

// Response is the standard API response
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Meta    interface{} `json:"meta,omitempty"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
    Success bool `json:"success"`
    Error   struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
    } `json:"error"`
}

// Success sends a success response
func Success(c *fiber.Ctx, data interface{}) error {
    return c.JSON(Response{
        Success: true,
        Data:    data,
    })
}

// SuccessWithMeta sends a success response with metadata
func SuccessWithMeta(c *fiber.Ctx, data, meta interface{}) error {
    return c.JSON(Response{
        Success: true,
        Data:    data,
        Meta:    meta,
    })
}

// Created sends a 201 created response
func Created(c *fiber.Ctx, data interface{}) error {
    return c.Status(fiber.StatusCreated).JSON(Response{
        Success: true,
        Data:    data,
    })
}

// NoContent sends a 204 no content response
func NoContent(c *fiber.Ctx) error {
    return c.SendStatus(fiber.StatusNoContent)
}

// Error sends an error response
func Error(c *fiber.Ctx, status int, code int, message string) error {
    return c.Status(status).JSON(ErrorResponse{
        Success: false,
        Error: struct {
            Code    int    `json:"code"`
            Message string `json:"message"`
        }{
            Code:    code,
            Message: message,
        },
    })
}

// BadRequest sends a 400 error
func BadRequest(c *fiber.Ctx, code int, message string) error {
    return Error(c, fiber.StatusBadRequest, code, message)
}

// Unauthorized sends a 401 error
func Unauthorized(c *fiber.Ctx, message string) error {
    return Error(c, fiber.StatusUnauthorized, 401, message)
}

// Forbidden sends a 403 error
func Forbidden(c *fiber.Ctx, message string) error {
    return Error(c, fiber.StatusForbidden, 403, message)
}

// NotFound sends a 404 error
func NotFound(c *fiber.Ctx, message string) error {
    return Error(c, fiber.StatusNotFound, 404, message)
}

// Conflict sends a 409 error
func Conflict(c *fiber.Ctx, code int, message string) error {
    return Error(c, fiber.StatusConflict, code, message)
}

// InternalError sends a 500 error
func InternalError(c *fiber.Ctx, code int, message string) error {
    return Error(c, fiber.StatusInternalServerError, code, message)
}
```

### Error Types

```go
// github.com/solvr-asia/protokit/fiber/errors/errors.go
package errors

import (
    "fmt"
    "net/http"

    "github.com/gofiber/fiber/v2"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// AppError represents an application error
type AppError struct {
    HTTPStatus int
    GRPCCode   codes.Code
    Code       int
    Message    string
    Err        error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
    return e.Err
}

// GRPCStatus returns gRPC status for the error
func (e *AppError) GRPCStatus() *status.Status {
    return status.New(e.GRPCCode, e.Message)
}

// FiberError converts to Fiber error response
func (e *AppError) FiberError(c *fiber.Ctx) error {
    return c.Status(e.HTTPStatus).JSON(fiber.Map{
        "success": false,
        "error": fiber.Map{
            "code":    e.Code,
            "message": e.Message,
        },
    })
}

// Error constructors
func New(httpStatus int, grpcCode codes.Code, code int, message string) *AppError {
    return &AppError{
        HTTPStatus: httpStatus,
        GRPCCode:   grpcCode,
        Code:       code,
        Message:    message,
    }
}

func Wrap(err error, httpStatus int, grpcCode codes.Code, code int, message string) *AppError {
    return &AppError{
        HTTPStatus: httpStatus,
        GRPCCode:   grpcCode,
        Code:       code,
        Message:    message,
        Err:        err,
    }
}

// Common error constructors
func BadRequest(code int, message string) *AppError {
    return New(http.StatusBadRequest, codes.InvalidArgument, code, message)
}

func Unauthorized(message string) *AppError {
    return New(http.StatusUnauthorized, codes.Unauthenticated, 401, message)
}

func Forbidden(message string) *AppError {
    return New(http.StatusForbidden, codes.PermissionDenied, 403, message)
}

func NotFound(message string) *AppError {
    return New(http.StatusNotFound, codes.NotFound, 404, message)
}

func Conflict(code int, message string) *AppError {
    return New(http.StatusConflict, codes.AlreadyExists, code, message)
}

func Internal(code int, message string) *AppError {
    return New(http.StatusInternalServerError, codes.Internal, code, message)
}

// ValidationError creates a validation error
func ValidationError(message string) *AppError {
    return New(http.StatusBadRequest, codes.InvalidArgument, 400, message)
}
```

---

## Protokit Fiber Middleware

### Auth Middleware

```go
// github.com/solvr-asia/protokit/fiber/middleware/auth.go
package middleware

import (
    "strings"

    "github.com/gofiber/fiber/v2"
    "github.com/solvr-asia/protokit/fiber/errors"
)

// AuthContext holds authenticated user information
type AuthContext struct {
    UserID      string
    Email       string
    Roles       []string
    CompanyCode string
    Token       string
}

// AuthConfig defines auth requirements for routes
type AuthConfig struct {
    NeedAuth     bool
    AllowedRoles []string
}

// TokenValidator validates tokens and returns auth context
type TokenValidator func(token string) (*AuthContext, error)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(validator TokenValidator, config AuthConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Skip if no auth required
        if !config.NeedAuth {
            return c.Next()
        }

        // Extract token from Authorization header
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return errors.Unauthorized("missing authorization header").FiberError(c)
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            return errors.Unauthorized("invalid authorization format").FiberError(c)
        }

        token := parts[1]

        // Validate token
        authCtx, err := validator(token)
        if err != nil {
            return errors.Unauthorized("invalid token").FiberError(c)
        }

        // Check roles if specified
        if len(config.AllowedRoles) > 0 {
            if !hasAnyRole(authCtx.Roles, config.AllowedRoles) {
                return errors.Forbidden("insufficient permissions").FiberError(c)
            }
        }

        // Store auth context in locals
        c.Locals("auth", authCtx)

        return c.Next()
    }
}

// GetAuthContext retrieves auth context from Fiber context
func GetAuthContext(c *fiber.Ctx) (*AuthContext, bool) {
    auth, ok := c.Locals("auth").(*AuthContext)
    return auth, ok
}

// MustGetAuthContext retrieves auth context or panics
func MustGetAuthContext(c *fiber.Ctx) *AuthContext {
    auth, ok := GetAuthContext(c)
    if !ok {
        panic("auth context not found")
    }
    return auth
}

func hasAnyRole(userRoles, allowedRoles []string) bool {
    roleSet := make(map[string]bool)
    for _, role := range userRoles {
        roleSet[role] = true
    }
    for _, allowed := range allowedRoles {
        if roleSet[allowed] {
            return true
        }
    }
    return false
}
```

### Logger Middleware

```go
// github.com/solvr-asia/protokit/fiber/middleware/logger.go
package middleware

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/sirupsen/logrus"
)

// LoggerMiddleware creates a logging middleware
func LoggerMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        // Process request
        err := c.Next()

        // Log after request
        duration := time.Since(start)
        entry := logrus.WithFields(logrus.Fields{
            "method":     c.Method(),
            "path":       c.Path(),
            "status":     c.Response().StatusCode(),
            "duration":   duration.String(),
            "ip":         c.IP(),
            "request_id": c.Locals("request_id"),
        })

        if err != nil {
            entry.WithError(err).Warn("Request failed")
        } else if c.Response().StatusCode() >= 400 {
            entry.Warn("Request completed with error status")
        } else {
            entry.Info("Request completed")
        }

        return err
    }
}
```

### Recovery Middleware

```go
// github.com/solvr-asia/protokit/fiber/middleware/recovery.go
package middleware

import (
    "runtime/debug"

    "github.com/gofiber/fiber/v2"
    "github.com/sirupsen/logrus"
)

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        defer func() {
            if r := recover(); r != nil {
                logrus.WithFields(logrus.Fields{
                    "panic": r,
                    "stack": string(debug.Stack()),
                    "path":  c.Path(),
                }).Error("Panic recovered")

                c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "success": false,
                    "error": fiber.Map{
                        "code":    500,
                        "message": "Internal server error",
                    },
                })
            }
        }()

        return c.Next()
    }
}
```

### Request ID Middleware

```go
// github.com/solvr-asia/protokit/fiber/middleware/requestid.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get or generate request ID
        requestID := c.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Store in locals and set response header
        c.Locals("request_id", requestID)
        c.Set("X-Request-ID", requestID)

        return c.Next()
    }
}

// GetRequestID retrieves request ID from context
func GetRequestID(c *fiber.Ctx) string {
    if id, ok := c.Locals("request_id").(string); ok {
        return id
    }
    return ""
}
```

---

## Protokit gRPC Interceptors

### Auth Interceptor (for gRPC)

```go
// github.com/solvr-asia/protokit/grpc/interceptor/auth.go
package interceptor

import (
    "context"
    "strings"

    "github.com/solvr-asia/protokit/fiber/errors"
    "github.com/solvr-asia/protokit/fiber/middleware"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

type contextKey string

const authContextKey contextKey = "auth_context"

// GRPCAuthInterceptor creates gRPC auth interceptor
func GRPCAuthInterceptor(validator middleware.TokenValidator, authConfig map[string]middleware.AuthConfig) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Get auth config for method
        config, exists := authConfig[info.FullMethod]
        if !exists || !config.NeedAuth {
            return handler(ctx, req)
        }

        // Extract token
        token, err := extractTokenFromContext(ctx)
        if err != nil {
            return nil, errors.Unauthorized("missing authorization").GRPCStatus().Err()
        }

        // Validate token
        authCtx, err := validator(token)
        if err != nil {
            return nil, errors.Unauthorized("invalid token").GRPCStatus().Err()
        }

        // Check roles
        if len(config.AllowedRoles) > 0 {
            if !hasAnyRole(authCtx.Roles, config.AllowedRoles) {
                return nil, errors.Forbidden("insufficient permissions").GRPCStatus().Err()
            }
        }

        // Add to context
        ctx = context.WithValue(ctx, authContextKey, authCtx)

        return handler(ctx, req)
    }
}

// GetGRPCAuthContext retrieves auth context from gRPC context
func GetGRPCAuthContext(ctx context.Context) (*middleware.AuthContext, error) {
    authCtx, ok := ctx.Value(authContextKey).(*middleware.AuthContext)
    if !ok {
        return nil, errors.Unauthorized("not authenticated")
    }
    return authCtx, nil
}

func extractTokenFromContext(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", errors.Unauthorized("no metadata")
    }

    authHeader := md.Get("authorization")
    if len(authHeader) == 0 {
        return "", errors.Unauthorized("no authorization header")
    }

    parts := strings.SplitN(authHeader[0], " ", 2)
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return "", errors.Unauthorized("invalid authorization format")
    }

    return parts[1], nil
}

func hasAnyRole(userRoles, allowedRoles []string) bool {
    roleSet := make(map[string]bool)
    for _, role := range userRoles {
        roleSet[role] = true
    }
    for _, allowed := range allowedRoles {
        if roleSet[allowed] {
            return true
        }
    }
    return false
}
```

---

## Protokit Validation

```go
// github.com/solvr-asia/protokit/validation/validator.go
package validation

import (
    "fmt"
    "net/mail"
    "regexp"
    "strconv"
    "strings"
    "unicode"

    "github.com/solvr-asia/protokit/fiber/errors"
)

// Validator validates struct fields
type Validator struct {
    customRules map[string]RuleFunc
}

// RuleFunc is a custom validation rule
type RuleFunc func(fieldName string, value interface{}, param string) error

// NewValidator creates a new validator
func NewValidator() *Validator {
    return &Validator{
        customRules: make(map[string]RuleFunc),
    }
}

// RegisterRule adds a custom validation rule
func (v *Validator) RegisterRule(name string, fn RuleFunc) {
    v.customRules[name] = fn
}

// ValidateField validates a field value against rules
func (v *Validator) ValidateField(fieldName string, value interface{}, rules string) error {
    if rules == "" {
        return nil
    }

    for _, rule := range strings.Split(rules, "|") {
        rule = strings.TrimSpace(rule)
        if rule == "" {
            continue
        }

        ruleName, param := parseRule(rule)
        if err := v.applyRule(fieldName, value, ruleName, param); err != nil {
            return err
        }
    }

    return nil
}

func parseRule(rule string) (string, string) {
    parts := strings.SplitN(rule, "=", 2)
    if len(parts) > 1 {
        return parts[0], parts[1]
    }
    return parts[0], ""
}

func (v *Validator) applyRule(fieldName string, value interface{}, ruleName, param string) error {
    // Check custom rules first
    if fn, ok := v.customRules[ruleName]; ok {
        return fn(fieldName, value, param)
    }

    // Built-in rules
    switch ruleName {
    case "required":
        return validateRequired(fieldName, value)
    case "email":
        return validateEmail(fieldName, value)
    case "min":
        return validateMin(fieldName, value, param)
    case "max":
        return validateMax(fieldName, value, param)
    case "len":
        return validateLen(fieldName, value, param)
    case "regex":
        return validateRegex(fieldName, value, param)
    case "in":
        return validateIn(fieldName, value, param)
    case "notin":
        return validateNotIn(fieldName, value, param)
    case "phone":
        return validatePhone(fieldName, value)
    case "url":
        return validateURL(fieldName, value)
    case "uuid":
        return validateUUID(fieldName, value)
    case "date":
        return validateDate(fieldName, value)
    case "datetime":
        return validateDateTime(fieldName, value)
    case "time":
        return validateTime(fieldName, value)
    case "numeric":
        return validateNumeric(fieldName, value)
    case "alpha":
        return validateAlpha(fieldName, value)
    case "alphanum":
        return validateAlphaNum(fieldName, value)
    }

    return nil
}

// Validation functions
func validateRequired(field string, value interface{}) error {
    if isEmpty(value) {
        return errors.ValidationError(fmt.Sprintf("%s is required", field))
    }
    return nil
}

func validateEmail(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if _, err := mail.ParseAddress(str); err != nil {
        return errors.ValidationError(fmt.Sprintf("%s must be a valid email", field))
    }
    return nil
}

func validateMin(field string, value interface{}, param string) error {
    min, _ := strconv.Atoi(param)
    switch v := value.(type) {
    case string:
        if len(v) > 0 && len(v) < min {
            return errors.ValidationError(fmt.Sprintf("%s must be at least %d characters", field, min))
        }
    case int32:
        if v != 0 && int(v) < min {
            return errors.ValidationError(fmt.Sprintf("%s must be at least %d", field, min))
        }
    case int64:
        if v != 0 && int(v) < min {
            return errors.ValidationError(fmt.Sprintf("%s must be at least %d", field, min))
        }
    }
    return nil
}

func validateMax(field string, value interface{}, param string) error {
    max, _ := strconv.Atoi(param)
    switch v := value.(type) {
    case string:
        if len(v) > max {
            return errors.ValidationError(fmt.Sprintf("%s must be at most %d characters", field, max))
        }
    case int32:
        if int(v) > max {
            return errors.ValidationError(fmt.Sprintf("%s must be at most %d", field, max))
        }
    case int64:
        if int(v) > max {
            return errors.ValidationError(fmt.Sprintf("%s must be at most %d", field, max))
        }
    }
    return nil
}

func validateLen(field string, value interface{}, param string) error {
    length, _ := strconv.Atoi(param)
    if str, ok := value.(string); ok && str != "" && len(str) != length {
        return errors.ValidationError(fmt.Sprintf("%s must be exactly %d characters", field, length))
    }
    return nil
}

func validateRegex(field string, value interface{}, pattern string) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if matched, _ := regexp.MatchString(pattern, str); !matched {
        return errors.ValidationError(fmt.Sprintf("%s format is invalid", field))
    }
    return nil
}

func validateIn(field string, value interface{}, param string) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    for _, allowed := range strings.Split(param, ",") {
        if str == strings.TrimSpace(allowed) {
            return nil
        }
    }
    return errors.ValidationError(fmt.Sprintf("%s must be one of: %s", field, param))
}

func validateNotIn(field string, value interface{}, param string) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    for _, disallowed := range strings.Split(param, ",") {
        if str == strings.TrimSpace(disallowed) {
            return errors.ValidationError(fmt.Sprintf("%s must not be: %s", field, disallowed))
        }
    }
    return nil
}

func validatePhone(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if matched, _ := regexp.MatchString(`^[\+]?[0-9\s\-\(\)]{8,20}$`, str); !matched {
        return errors.ValidationError(fmt.Sprintf("%s must be a valid phone number", field))
    }
    return nil
}

func validateURL(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if matched, _ := regexp.MatchString(`^https?://[^\s]+$`, str); !matched {
        return errors.ValidationError(fmt.Sprintf("%s must be a valid URL", field))
    }
    return nil
}

func validateUUID(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if matched, _ := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, str); !matched {
        return errors.ValidationError(fmt.Sprintf("%s must be a valid UUID", field))
    }
    return nil
}

func validateDate(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, str); !matched {
        return errors.ValidationError(fmt.Sprintf("%s must be a valid date (YYYY-MM-DD)", field))
    }
    return nil
}

func validateDateTime(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, str); !matched {
        return errors.ValidationError(fmt.Sprintf("%s must be a valid datetime", field))
    }
    return nil
}

func validateTime(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    if matched, _ := regexp.MatchString(`^\d{2}:\d{2}:\d{2}$`, str); !matched {
        return errors.ValidationError(fmt.Sprintf("%s must be a valid time (HH:MM:SS)", field))
    }
    return nil
}

func validateNumeric(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    for _, r := range str {
        if !unicode.IsDigit(r) {
            return errors.ValidationError(fmt.Sprintf("%s must contain only numbers", field))
        }
    }
    return nil
}

func validateAlpha(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    for _, r := range str {
        if !unicode.IsLetter(r) {
            return errors.ValidationError(fmt.Sprintf("%s must contain only letters", field))
        }
    }
    return nil
}

func validateAlphaNum(field string, value interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil
    }
    for _, r := range str {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
            return errors.ValidationError(fmt.Sprintf("%s must contain only letters and numbers", field))
        }
    }
    return nil
}

func isEmpty(value interface{}) bool {
    switch v := value.(type) {
    case string:
        return v == ""
    case int, int32, int64:
        return v == 0
    case float32, float64:
        return v == 0
    case nil:
        return true
    }
    return false
}
```

---

## Generated Code Structure

After running `protoc` with `protoc-gen-protokit`:

```
handler/
├── grpc/                           # Generated gRPC handlers
│   └── user/
│       ├── user.pb.go              # Proto messages
│       ├── user_grpc.pb.go         # gRPC service interface
│       └── user.protokit.go        # Auth config, validation
├── http/                           # Generated Fiber handlers
│   └── user/
│       └── user_handler.go         # Fiber route handlers
└── user_handler.go                 # Your handler implementation
```

### Generated Protokit Code

```go
// handler/grpc/user/user.protokit.go (GENERATED)
package user

import (
    "github.com/solvr-asia/protokit/fiber/middleware"
    "github.com/solvr-asia/protokit/validation"
)

// AuthConfigMethods contains auth config for each gRPC method
var AuthConfigMethods = map[string]middleware.AuthConfig{
    "/user.UserApi/Register":     {NeedAuth: false, AllowedRoles: nil},
    "/user.UserApi/GetProfile":   {NeedAuth: true, AllowedRoles: nil},
    "/user.UserApi/ListAllUsers": {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
    "/user.UserApi/GetMyPayslip": {NeedAuth: true, AllowedRoles: []string{"employee"}},
}

// RouteAuthConfig contains auth config for each REST route
var RouteAuthConfig = map[string]middleware.AuthConfig{
    "POST /api/v1/user/register":    {NeedAuth: false, AllowedRoles: nil},
    "GET /api/v1/user/profile":      {NeedAuth: true, AllowedRoles: nil},
    "GET /api/v1/admin/users":       {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
    "GET /api/v1/employee/payslip":  {NeedAuth: true, AllowedRoles: []string{"employee"}},
}

// ValidateRequest validates proto request messages
func ValidateRequest(req interface{}) error {
    v := validation.NewValidator()

    switch r := req.(type) {
    case *RegisterReq:
        if err := v.ValidateField("email", r.Email, "required|email"); err != nil {
            return err
        }
        if err := v.ValidateField("password", r.Password, "required|min=8|max=128"); err != nil {
            return err
        }
        if err := v.ValidateField("name", r.Name, "required|min=2|max=100"); err != nil {
            return err
        }
        if err := v.ValidateField("phone", r.Phone, "phone"); err != nil {
            return err
        }

    case *ListUsersReq:
        // Apply defaults
        if r.Page == 0 {
            r.Page = 1
        }
        if r.Size == 0 {
            r.Size = 10
        }
        if r.SortBy == "" {
            r.SortBy = "created_at"
        }
        if r.SortOrder == "" {
            r.SortOrder = "desc"
        }

    case *GetPayslipReq:
        if err := v.ValidateField("year", r.Year, "required|min=2000|max=2100"); err != nil {
            return err
        }
        if err := v.ValidateField("month", r.Month, "required|min=1|max=12"); err != nil {
            return err
        }
    }

    return nil
}
```

### Generated Fiber Handler

```go
// handler/http/user/user_handler.go (GENERATED)
package user

import (
    "github.com/gofiber/fiber/v2"
    "github.com/solvr-asia/protokit/fiber/middleware"
    pb "myapp/handler/grpc/user"
)

// UserRoutes holds the user service handler
type UserRoutes struct {
    handler pb.UserApiServer
}

// NewUserRoutes creates new user routes
func NewUserRoutes(handler pb.UserApiServer) *UserRoutes {
    return &UserRoutes{handler: handler}
}

// RegisterRoutes registers all user routes
func (r *UserRoutes) RegisterRoutes(app *fiber.App, validator middleware.TokenValidator) {
    api := app.Group("/api/v1")

    // Public routes
    api.Post("/user/register", r.Register)

    // Protected routes
    api.Get("/user/profile",
        middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/user/profile"]),
        r.GetProfile,
    )

    api.Get("/admin/users",
        middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/admin/users"]),
        r.ListAllUsers,
    )

    api.Get("/employee/payslip",
        middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/employee/payslip"]),
        r.GetMyPayslip,
    )
}

// Register handles POST /api/v1/user/register
func (r *UserRoutes) Register(c *fiber.Ctx) error {
    var req pb.RegisterReq
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
    }

    if err := pb.ValidateRequest(&req); err != nil {
        return err.(*errors.AppError).FiberError(c)
    }

    // Convert Fiber context to gRPC context
    ctx := c.UserContext()

    resp, err := r.handler.Register(ctx, &req)
    if err != nil {
        return handleError(c, err)
    }

    return response.Created(c, resp)
}

// GetProfile handles GET /api/v1/user/profile
func (r *UserRoutes) GetProfile(c *fiber.Ctx) error {
    ctx := c.UserContext()

    // Pass auth context to gRPC handler
    if authCtx, ok := middleware.GetAuthContext(c); ok {
        ctx = context.WithValue(ctx, "auth", authCtx)
    }

    resp, err := r.handler.GetProfile(ctx, &emptypb.Empty{})
    if err != nil {
        return handleError(c, err)
    }

    return response.Success(c, resp)
}

// ListAllUsers handles GET /api/v1/admin/users
func (r *UserRoutes) ListAllUsers(c *fiber.Ctx) error {
    req := &pb.ListUsersReq{
        Page:      int32(c.QueryInt("page", 1)),
        Size:      int32(c.QueryInt("size", 10)),
        Search:    c.Query("search"),
        SortBy:    c.Query("sortBy", "created_at"),
        SortOrder: c.Query("sortOrder", "desc"),
    }

    if err := pb.ValidateRequest(req); err != nil {
        return err.(*errors.AppError).FiberError(c)
    }

    ctx := c.UserContext()
    if authCtx, ok := middleware.GetAuthContext(c); ok {
        ctx = context.WithValue(ctx, "auth", authCtx)
    }

    resp, err := r.handler.ListAllUsers(ctx, req)
    if err != nil {
        return handleError(c, err)
    }

    return response.SuccessWithMeta(c, resp.Users, resp.Pagination)
}

// GetMyPayslip handles GET /api/v1/employee/payslip
func (r *UserRoutes) GetMyPayslip(c *fiber.Ctx) error {
    req := &pb.GetPayslipReq{
        Year:  int32(c.QueryInt("year")),
        Month: int32(c.QueryInt("month")),
    }

    if err := pb.ValidateRequest(req); err != nil {
        return err.(*errors.AppError).FiberError(c)
    }

    ctx := c.UserContext()
    if authCtx, ok := middleware.GetAuthContext(c); ok {
        ctx = context.WithValue(ctx, "auth", authCtx)
    }

    resp, err := r.handler.GetMyPayslip(ctx, req)
    if err != nil {
        return handleError(c, err)
    }

    return response.Success(c, resp)
}

func handleError(c *fiber.Ctx, err error) error {
    if appErr, ok := err.(*errors.AppError); ok {
        return appErr.FiberError(c)
    }
    return response.InternalError(c, 500, "internal server error")
}
```

---

## Project Structure (Application)

```
myapp/
├── main.go
├── service.yaml
├── go.mod
├── Dockerfile
├── config/
│   └── config.go
├── contract/                       # Proto files (Source of Truth)
│   ├── user.proto
│   └── order.proto
├── handler/
│   ├── grpc/                       # Generated gRPC code
│   │   ├── user/
│   │   │   ├── user.pb.go
│   │   │   ├── user_grpc.pb.go
│   │   │   └── user.protokit.go
│   │   └── order/
│   ├── http/                       # Generated Fiber routes
│   │   ├── user/
│   │   │   └── user_handler.go
│   │   └── order/
│   ├── user_handler.go             # Your implementation
│   └── order_handler.go
├── app/
│   └── usecase/
│       ├── user/
│       └── order/
├── entity/
├── repository/
├── pkg/
├── migrations/
└── clients/
    └── grpc/
```

---

## Main Application

```go
// main.go
package main

import (
    "context"
    "fmt"
    "time"

    "myapp/app/usecase/user"
    "myapp/config"
    "myapp/handler"
    pb_user "myapp/handler/grpc/user"
    http_user "myapp/handler/http/user"
    "myapp/repository/user_repository"

    "github.com/solvr-asia/protokit/fiber"
    "github.com/solvr-asia/protokit/fiber/middleware"
    "github.com/solvr-asia/protokit/grpc/interceptor"

    "github.com/sirupsen/logrus"
    "google.golang.org/grpc/reflection"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    cfg := config.New()

    // Initialize database
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBTimezone)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        logrus.Fatalf("Failed to connect to database: %v", err)
    }

    // Initialize layers
    userRepo := user_repository.New(db)
    userUC := user.NewUseCase(userRepo)
    userHandler := handler.NewUserHandler(userUC)

    // Create token validator
    tokenValidator := createTokenValidator(cfg)

    // Create Protokit server
    server := fiber.NewServer(fiber.Config{
        AppName:     cfg.ServiceName,
        HTTPPort:    cfg.HTTPPort,
        GRPCPort:    cfg.GRPCPort,
        EnableCORS:  true,
        CORSOrigins: "*",
    },
        // gRPC interceptors
        fiber.WithGRPCInterceptor(interceptor.GRPCAuthInterceptor(tokenValidator, pb_user.AuthConfigMethods)),
    )

    // Setup Fiber middleware
    app := server.Fiber()
    app.Use(middleware.RequestIDMiddleware())
    app.Use(middleware.RecoveryMiddleware())
    app.Use(middleware.LoggerMiddleware())

    // Health check
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // Register HTTP routes (Fiber)
    userRoutes := http_user.NewUserRoutes(userHandler)
    userRoutes.RegisterRoutes(app, tokenValidator)

    // Register gRPC services
    pb_user.RegisterUserApiServer(server.GRPC(), userHandler)
    reflection.Register(server.GRPC())

    logrus.Infof("Starting server - HTTP:%d gRPC:%d", cfg.HTTPPort, cfg.GRPCPort)

    // Start with graceful shutdown
    if err := server.ListenGraceful(30 * time.Second); err != nil {
        logrus.Fatalf("Server error: %v", err)
    }
}

func createTokenValidator(cfg *config.Config) middleware.TokenValidator {
    return func(token string) (*middleware.AuthContext, error) {
        // Implement your token validation logic
        // - Validate JWT
        // - Call auth service
        // - etc.
        return &middleware.AuthContext{
            UserID:      "user-123",
            Email:       "user@example.com",
            Roles:       []string{"employee"},
            CompanyCode: "COMPANY-001",
            Token:       token,
        }, nil
    }
}
```

---

## Handler Implementation

```go
// handler/user_handler.go
package handler

import (
    "context"
    "time"

    "myapp/app/usecase/user"
    "myapp/entity"
    pb "myapp/handler/grpc/user"

    "github.com/solvr-asia/protokit/fiber/errors"
    "github.com/solvr-asia/protokit/fiber/middleware"
    "google.golang.org/protobuf/types/known/emptypb"
)

type userHandler struct {
    pb.UnimplementedUserApiServer
    userUC user.UseCase
}

func NewUserHandler(userUC user.UseCase) pb.UserApiServer {
    return &userHandler{userUC: userUC}
}

func (h *userHandler) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error) {
    // Validation already done by generated code

    result, err := h.userUC.Register(ctx, user.RegisterInput{
        Email:    req.Email,
        Password: req.Password,
        Name:     req.Name,
        Phone:    req.Phone,
    })
    if err != nil {
        if err == user.ErrEmailExists {
            return nil, errors.Conflict(40901, "email already registered")
        }
        return nil, errors.Internal(50001, "failed to register user")
    }

    return &pb.RegisterRes{
        Id:    result.ID,
        Email: result.Email,
        Name:  result.Name,
    }, nil
}

func (h *userHandler) GetProfile(ctx context.Context, req *emptypb.Empty) (*pb.UserProfile, error) {
    // Get auth context
    authCtx := getAuthFromContext(ctx)
    if authCtx == nil {
        return nil, errors.Unauthorized("authentication required")
    }

    profile, err := h.userUC.GetProfile(ctx, authCtx.UserID)
    if err != nil {
        if err == user.ErrNotFound {
            return nil, errors.NotFound("user not found")
        }
        return nil, errors.Internal(50002, "failed to get profile")
    }

    return &pb.UserProfile{
        Id:        profile.ID,
        Email:     profile.Email,
        Name:      profile.Name,
        Phone:     profile.Phone,
        Status:    string(profile.Status),
        CreatedAt: profile.CreatedAt.Format(time.RFC3339),
    }, nil
}

func (h *userHandler) ListAllUsers(ctx context.Context, req *pb.ListUsersReq) (*pb.ListUsersRes, error) {
    users, total, err := h.userUC.ListAll(ctx, user.ListInput{
        Page:      int(req.Page),
        Size:      int(req.Size),
        Search:    req.Search,
        SortBy:    req.SortBy,
        SortOrder: req.SortOrder,
    })
    if err != nil {
        return nil, errors.Internal(50003, "failed to list users")
    }

    pbUsers := make([]*pb.UserProfile, len(users))
    for i, u := range users {
        pbUsers[i] = &pb.UserProfile{
            Id:        u.ID,
            Email:     u.Email,
            Name:      u.Name,
            Phone:     u.Phone,
            Status:    string(u.Status),
            CreatedAt: u.CreatedAt.Format(time.RFC3339),
        }
    }

    totalPages := (total + int64(req.Size) - 1) / int64(req.Size)

    return &pb.ListUsersRes{
        Users: pbUsers,
        Pagination: &pb.Pagination{
            Page:       req.Page,
            Size:       req.Size,
            Total:      total,
            TotalPages: int32(totalPages),
        },
    }, nil
}

func (h *userHandler) GetMyPayslip(ctx context.Context, req *pb.GetPayslipReq) (*pb.Payslip, error) {
    authCtx := getAuthFromContext(ctx)
    if authCtx == nil {
        return nil, errors.Unauthorized("authentication required")
    }

    payslip, err := h.userUC.GetPayslip(ctx, authCtx.UserID, int(req.Year), int(req.Month))
    if err != nil {
        if err == user.ErrNotFound {
            return nil, errors.NotFound("payslip not found")
        }
        return nil, errors.Internal(50004, "failed to get payslip")
    }

    return &pb.Payslip{
        Id:          payslip.ID,
        EmployeeId:  payslip.EmployeeID,
        Year:        int32(payslip.Year),
        Month:       int32(payslip.Month),
        GrossSalary: payslip.GrossSalary,
        NetSalary:   payslip.NetSalary,
    }, nil
}

func getAuthFromContext(ctx context.Context) *middleware.AuthContext {
    if auth, ok := ctx.Value("auth").(*middleware.AuthContext); ok {
        return auth
    }
    return nil
}
```

---

## Makefile

```makefile
.PHONY: proto build run test docker

# Generate code from proto files
proto:
	protokit generate

# Manual protoc
proto-manual:
	protoc \
		--proto_path=contract \
		--proto_path=$(GOPATH)/src \
		--go_out=handler/grpc \
		--go_opt=paths=source_relative \
		--go-grpc_out=handler/grpc \
		--go-grpc_opt=paths=source_relative \
		--protokit_out=handler \
		--protokit_opt=paths=source_relative \
		contract/*.proto

build:
	go build -o bin/server main.go

run:
	go run main.go

test:
	go test -v -cover ./...

docker:
	docker build -t myapp:latest .

install-tools:
	go install github.com/solvr-asia/protokit/cmd/protokit@latest
	go install github.com/solvr-asia/protokit/cmd/protoc-gen-protokit@latest
```

---

## Key Dependencies

```go
// go.mod
require (
    github.com/solvr-asia/protokit v0.1.0
    github.com/gofiber/fiber/v2 v2.52.0
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.32.0
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    github.com/sirupsen/logrus v1.9.3
    github.com/google/uuid v1.5.0
    github.com/joho/godotenv v1.5.1
    github.com/kelseyhightower/envconfig v1.4.0
)
```

---

## Summary

| Component | Description |
|-----------|-------------|
| **protokit/fiber** | Fiber server, response helpers, middleware |
| **protokit/fiber/middleware** | Auth, Logger, Recovery, RequestID middleware |
| **protokit/fiber/errors** | Error types with Fiber & gRPC support |
| **protokit/grpc** | gRPC server, interceptors |
| **protokit/validation** | Validation engine |
| **protoc-gen-protokit** | Generates Fiber handlers + gRPC code |

**Data Flow:**
```
REST Request (Fiber)
    → Middleware (RequestID → Recovery → Logger → Auth)
    → Generated Fiber Handler
    → ValidateRequest()
    → Your Handler Implementation
    → UseCase → Repository → Database

gRPC Request
    → Interceptors (Auth → Logger → Recovery)
    → Your Handler Implementation
    → UseCase → Repository → Database
```

**Dual Protocol:**
- **REST via Fiber**: High-performance HTTP API for clients
- **gRPC**: Efficient service-to-service communication
- **Single Handler**: One implementation serves both protocols
