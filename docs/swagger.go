package docs

import (
	"github.com/gofiber/fiber/v2"
	scalar "github.com/yokeTH/gofiber-scalar"
)

// SetupSwagger configures Swagger documentation with Scalar UI
func SetupSwagger(app *fiber.App) {
	// Serve OpenAPI spec
	app.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
		return c.JSON(GetOpenAPISpec())
	})

	// Serve Scalar UI
	app.Get("/docs/*", scalar.New(scalar.Config{
		Title:      "Go-GRST-Boilerplate API",
		RawSpecUrl: "/docs/openapi.json",
	}))
}

// GetOpenAPISpec returns the OpenAPI specification
func GetOpenAPISpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Go-GRST-Boilerplate API",
			"description": "A production-ready Go monolithic application boilerplate built with **Go Fiber** for REST API and **gRPC** for service-to-service communication, following **Domain-Driven Design (DDD)** and **Clean Architecture** principles.\n\n### Features\n- 🔐 **PASETO v4** authentication (symmetric encryption)\n- 📊 **Paginated** user listing with search and sort\n- 🗂️ **API versioning** (`/api/v1`)\n- 🔍 **OpenTelemetry** distributed tracing\n- 🐰 **RabbitMQ** message queue integration\n- 💾 **PostgreSQL** with GORM ORM\n\n### Authentication\nAll protected endpoints require a valid PASETO token in the `Authorization` header:\n```\nAuthorization: Bearer v4.local.xxxxx...\n```\nObtain a token via the **Login** endpoint, and refresh it via the **Refresh** endpoint before it expires.",
			"version":     "1.0.0",
			"contact": map[string]string{
				"name":  "API Support",
				"email": "support@example.com",
			},
			"license": map[string]string{
				"name": "MIT",
				"url":  "https://opensource.org/licenses/MIT",
			},
		},
		"servers": []map[string]string{
			{"url": "http://localhost:3000", "description": "Local development server"},
		},
		"tags": []map[string]interface{}{
			{"name": "Health", "description": "Service health and readiness probes for load balancers and orchestrators (e.g., Kubernetes liveness/readiness probes)."},
			{"name": "Auth", "description": "Authentication endpoints for user registration, login, token refresh, profile retrieval, and logout. Uses PASETO v4 symmetric encryption for secure, stateless token management."},
			{"name": "Users", "description": "User management resource endpoints (admin only). Provides full CRUD operations for managing user accounts, including listing with pagination/search/sort, viewing individual profiles, updating user details, and soft-deleting accounts."},
		},
		"paths": map[string]interface{}{
			// --- Health ---
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Health"},
					"summary":     "Liveness probe",
					"description": "Returns the liveness status of the service. Use this endpoint for Kubernetes liveness probes or basic uptime monitoring. A `200 OK` response indicates the service process is running and accepting connections. This does **not** verify downstream dependencies — use `/ready` for that.",
					"operationId": "healthCheck",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is alive and accepting connections",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/HealthResponse",
									},
								},
							},
						},
					},
				},
			},
			"/ready": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Health"},
					"summary":     "Readiness probe",
					"description": "Returns the readiness status of the service including the health of all downstream dependencies (PostgreSQL, Redis, RabbitMQ). Use this for Kubernetes readiness probes. A `503 Service Unavailable` response means one or more dependencies are unhealthy and the service should be temporarily removed from the load balancer rotation.",
					"operationId": "readinessCheck",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "All dependencies are healthy — service is ready to accept traffic",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ReadinessResponse",
									},
								},
							},
						},
						"503": map[string]interface{}{
							"description": "One or more dependencies are unhealthy — service should not receive traffic",
						},
					},
				},
			},

			// --- Auth ---
			"/api/v1/auth/register": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Auth"},
					"summary":     "Register a new user account",
					"description": "Creates a new user account with the provided email, password, and name. The email must be unique across all accounts. After successful registration, the user receives a confirmation with their generated UUID. The account starts in `pending` status and the user should proceed to the **Login** endpoint to obtain an access token.\n\n**Password requirements**: minimum 8 characters, maximum 128 characters.\n\n**Duplicate email**: returns `409 Conflict` if the email is already registered.",
					"operationId": "register",
					"requestBody": map[string]interface{}{
						"required":    true,
						"description": "User registration payload with email, password, display name, and optional phone number",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/RegisterRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Account created successfully — returns the new user's ID, email, and name",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/RegisterResponse",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Validation error — missing required fields, invalid email format, or password too short",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
						"409": map[string]interface{}{
							"description": "Conflict — a user with this email address already exists",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Auth"},
					"summary":     "Authenticate and obtain access token",
					"description": "Authenticates a user with email and password credentials. On success, returns a PASETO v4 access token (symmetric encryption) along with the user's profile information. The token should be included in subsequent requests via the `Authorization: Bearer <token>` header.\n\n**Token format**: `v4.local.xxxxx...` (PASETO v4 local/symmetric)\n\n**Token expiration**: configurable via `JWT_EXPIRATION` environment variable (default: 24 hours)\n\n**Invalid credentials**: returns `401 Unauthorized` with a generic error message (does not reveal whether the email exists).",
					"operationId": "login",
					"requestBody": map[string]interface{}{
						"required":    true,
						"description": "Login credentials — email address and password",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/LoginRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Authentication successful — returns PASETO access token and user profile",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/LoginResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Authentication failed — invalid email or password",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/auth/refresh": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Auth"},
					"summary":     "Refresh access token",
					"description": "Issues a new PASETO access token using the current valid token. Use this endpoint to extend the user's session without requiring re-authentication. The old token remains valid until its original expiration time (stateless — no token rotation).\n\n**When to use**: call this before the current token expires to maintain an active session.\n\n**Requires**: valid, non-expired PASETO token in the `Authorization` header.",
					"operationId": "refreshToken",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "New access token issued successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/RefreshTokenResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Token is invalid, expired, or missing",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/auth/me": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Auth"},
					"summary":     "Get current user profile",
					"description": "Returns the full profile of the currently authenticated user, including their ID, email, name, phone, status, and account creation timestamp. This endpoint extracts the user identity from the PASETO token and fetches the latest profile data from the database.\n\n**Use case**: display the logged-in user's profile in the UI, verify token claims against the database, or retrieve the latest user status.",
					"operationId": "getMe",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Current user's profile data retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/UserProfileResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Not authenticated — token is invalid, expired, or missing",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/auth/logout": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Auth"},
					"summary":     "Logout current session",
					"description": "Terminates the current user session. Since PASETO tokens are stateless, this endpoint returns a success response to signal the client to discard the token. The token itself remains technically valid until its expiration time.\n\n**Client responsibility**: remove the stored token from local storage, cookies, or memory upon receiving the success response.\n\n**Note**: for server-side token revocation, consider implementing a Redis-backed token blacklist.",
					"operationId": "logout",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Logout acknowledged — client should discard the token",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/LogoutResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Not authenticated — token is invalid or missing",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},

			// --- Users Resource ---
			"/api/v1/users": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Users"},
					"summary":     "List all users (paginated)",
					"description": "Returns a paginated list of all user accounts. Supports full-text search across name and email fields, configurable sorting, and adjustable page size.\n\n**Access**: requires `admin` or `superadmin` role.\n\n**Default behavior**: returns page 1 with 10 results per page, sorted by `created_at` descending (newest first).\n\n**Search**: case-insensitive partial match on `name` and `email` fields using `ILIKE`.",
					"operationId": "listUsers",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number for pagination (1-indexed). Defaults to 1.",
							"schema":      map[string]interface{}{"type": "integer", "default": 1, "minimum": 1, "example": 1},
						},
						{
							"name":        "size",
							"in":          "query",
							"description": "Number of records per page. Must be between 1 and 100. Defaults to 10.",
							"schema":      map[string]interface{}{"type": "integer", "default": 10, "minimum": 1, "maximum": 100, "example": 10},
						},
						{
							"name":        "search",
							"in":          "query",
							"description": "Full-text search term. Searches across `name` and `email` fields using case-insensitive partial matching (SQL `ILIKE`).",
							"schema":      map[string]interface{}{"type": "string", "maxLength": 100, "example": "john"},
						},
						{
							"name":        "sortBy",
							"in":          "query",
							"description": "Field to sort results by. Allowed values: `created_at`, `name`, `email`. Defaults to `created_at`.",
							"schema":      map[string]interface{}{"type": "string", "enum": []string{"created_at", "name", "email"}, "default": "created_at"},
						},
						{
							"name":        "sortOrder",
							"in":          "query",
							"description": "Sort direction. `asc` for ascending (A→Z, oldest first), `desc` for descending (Z→A, newest first). Defaults to `desc`.",
							"schema":      map[string]interface{}{"type": "string", "enum": []string{"asc", "desc"}, "default": "desc"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Paginated list of users with pagination metadata (page, size, total, totalPages)",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ListUsersResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Not authenticated — token is invalid, expired, or missing",
						},
						"403": map[string]interface{}{
							"description": "Forbidden — requires `admin` or `superadmin` role",
						},
					},
				},
			},
			"/api/v1/users/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Users"},
					"summary":     "Get user by ID",
					"description": "Retrieves the full profile of a specific user by their UUID. Returns all user fields including status, creation date, and contact information.\n\n**Access**: requires `admin` or `superadmin` role.\n\n**Not found**: returns `404` if the user does not exist or has been soft-deleted.",
					"operationId": "getUser",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "Unique user identifier (UUID v4 format)",
							"schema":      map[string]interface{}{"type": "string", "format": "uuid", "example": "550e8400-e29b-41d4-a716-446655440000"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User profile retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/UserProfileResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Not authenticated",
						},
						"403": map[string]interface{}{
							"description": "Forbidden — requires `admin` or `superadmin` role",
						},
						"404": map[string]interface{}{
							"description": "User not found or has been deleted",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
				"put": map[string]interface{}{
					"tags":        []string{"Users"},
					"summary":     "Update user by ID",
					"description": "Updates the profile of a specific user. Only the fields provided in the request body will be updated — omitted fields are left unchanged (partial update / PATCH semantics).\n\n**Updatable fields**: `name`, `phone`, `status`.\n\n**Status values**: `active`, `inactive`, `pending`.\n\n**Access**: requires `admin` or `superadmin` role.",
					"operationId": "updateUser",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "Unique user identifier (UUID v4 format)",
							"schema":      map[string]interface{}{"type": "string", "format": "uuid"},
						},
					},
					"requestBody": map[string]interface{}{
						"required":    true,
						"description": "Fields to update — all fields are optional, only provided fields will be changed",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/UpdateUserRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User updated successfully — returns the complete updated profile",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/UserProfileResponse",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Validation error — invalid status value or field format",
						},
						"401": map[string]interface{}{
							"description": "Not authenticated",
						},
						"403": map[string]interface{}{
							"description": "Forbidden — requires `admin` or `superadmin` role",
						},
						"404": map[string]interface{}{
							"description": "User not found",
						},
					},
				},
				"delete": map[string]interface{}{
					"tags":        []string{"Users"},
					"summary":     "Delete user by ID (soft delete)",
					"description": "Soft-deletes a user account by setting the `deleted_at` timestamp. The user record is retained in the database but excluded from all queries. This operation is **irreversible** through the API — data recovery requires direct database access.\n\n**Behavior**: the user's token will continue to work until expiration, but their profile will return `404` on subsequent lookups.\n\n**Access**: requires `admin` or `superadmin` role.",
					"operationId": "deleteUser",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "Unique user identifier (UUID v4 format)",
							"schema":      map[string]interface{}{"type": "string", "format": "uuid"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User deleted successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/DeleteResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Not authenticated",
						},
						"403": map[string]interface{}{
							"description": "Forbidden — requires `admin` or `superadmin` role",
						},
						"404": map[string]interface{}{
							"description": "User not found or already deleted",
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "PASETO",
					"description":  "PASETO v4 symmetric token. Obtain via the Login endpoint. Format: `v4.local.xxxxx...`",
				},
			},
			"schemas": map[string]interface{}{
				"HealthResponse": map[string]interface{}{
					"type":        "object",
					"description": "Liveness probe response indicating the service process is running",
					"properties": map[string]interface{}{
						"status":  map[string]interface{}{"type": "string", "example": "ok", "description": "Service status — always `ok` if the endpoint responds"},
						"service": map[string]interface{}{"type": "string", "example": "go-grst-boilerplate", "description": "Service name from configuration"},
					},
				},
				"ReadinessResponse": map[string]interface{}{
					"type":        "object",
					"description": "Readiness probe response with individual dependency health checks",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "object",
							"description": "Health status of each downstream dependency",
							"properties": map[string]interface{}{
								"database": map[string]interface{}{"type": "string", "enum": []string{"healthy", "unhealthy"}, "description": "PostgreSQL connection status"},
								"redis":    map[string]interface{}{"type": "string", "enum": []string{"healthy", "unhealthy", "disabled"}, "description": "Redis connection status"},
								"rabbitmq": map[string]interface{}{"type": "string", "enum": []string{"healthy", "unhealthy", "disabled"}, "description": "RabbitMQ connection status"},
							},
						},
					},
				},
				"RegisterRequest": map[string]interface{}{
					"type":        "object",
					"description": "Payload for creating a new user account",
					"required":    []string{"email", "password", "name"},
					"properties": map[string]interface{}{
						"email":    map[string]interface{}{"type": "string", "format": "email", "description": "Unique email address — used for login", "example": "john.doe@example.com"},
						"password": map[string]interface{}{"type": "string", "minLength": 8, "maxLength": 128, "description": "Account password (hashed with bcrypt before storage)", "example": "SecureP@ss123"},
						"name":     map[string]interface{}{"type": "string", "minLength": 2, "maxLength": 100, "description": "Display name shown in the user profile", "example": "John Doe"},
						"phone":    map[string]interface{}{"type": "string", "description": "Optional phone number for contact purposes", "example": "+62812345678"},
					},
				},
				"RegisterResponse": map[string]interface{}{
					"type":        "object",
					"description": "Successful registration response containing the new user's identifiers",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":    map[string]interface{}{"type": "string", "format": "uuid", "description": "Auto-generated UUID v4 identifier", "example": "550e8400-e29b-41d4-a716-446655440000"},
								"email": map[string]interface{}{"type": "string", "format": "email", "description": "Registered email address"},
								"name":  map[string]interface{}{"type": "string", "description": "Display name"},
							},
						},
					},
				},
				"LoginRequest": map[string]interface{}{
					"type":        "object",
					"description": "Login credentials for authentication",
					"required":    []string{"email", "password"},
					"properties": map[string]interface{}{
						"email":    map[string]interface{}{"type": "string", "format": "email", "description": "Registered email address", "example": "john.doe@example.com"},
						"password": map[string]interface{}{"type": "string", "description": "Account password", "example": "SecureP@ss123"},
					},
				},
				"LoginResponse": map[string]interface{}{
					"type":        "object",
					"description": "Successful authentication response with PASETO access token and user profile",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"token": map[string]interface{}{"type": "string", "description": "PASETO v4 access token — include in `Authorization: Bearer <token>` header", "example": "v4.local.xxxxxxxxxxxxxxxxxxxxx"},
								"user":  map[string]interface{}{"$ref": "#/components/schemas/UserProfile"},
							},
						},
					},
				},
				"RefreshTokenResponse": map[string]interface{}{
					"type":        "object",
					"description": "New access token issued from a valid existing token",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"token": map[string]interface{}{"type": "string", "description": "New PASETO v4 access token with refreshed expiration", "example": "v4.local.yyyyyyyyyyyyyyyyyyyyy"},
							},
						},
					},
				},
				"LogoutResponse": map[string]interface{}{
					"type":        "object",
					"description": "Logout acknowledgment — client should discard the stored token",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"message": map[string]interface{}{"type": "string", "example": "successfully logged out", "description": "Confirmation message"},
							},
						},
					},
				},
				"UserProfileResponse": map[string]interface{}{
					"type":        "object",
					"description": "Standard response wrapper containing a user profile object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"$ref": "#/components/schemas/UserProfile",
						},
					},
				},
				"UserProfile": map[string]interface{}{
					"type":        "object",
					"description": "Complete user profile with all public fields",
					"properties": map[string]interface{}{
						"id":        map[string]interface{}{"type": "string", "format": "uuid", "description": "Unique user identifier (UUID v4)", "example": "550e8400-e29b-41d4-a716-446655440000"},
						"email":     map[string]interface{}{"type": "string", "format": "email", "description": "User's email address (unique)", "example": "john.doe@example.com"},
						"name":      map[string]interface{}{"type": "string", "description": "User's display name", "example": "John Doe"},
						"phone":     map[string]interface{}{"type": "string", "description": "User's phone number", "example": "+62812345678"},
						"status":    map[string]interface{}{"type": "string", "enum": []string{"active", "inactive", "pending"}, "description": "Account status: `active` (fully verified), `inactive` (disabled by admin), `pending` (awaiting verification)", "example": "active"},
						"createdAt": map[string]interface{}{"type": "string", "format": "date-time", "description": "Account creation timestamp in RFC 3339 format", "example": "2026-01-15T10:30:00Z"},
					},
				},
				"ListUsersResponse": map[string]interface{}{
					"type":        "object",
					"description": "Paginated list of user profiles with metadata for building pagination UI",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type":        "array",
							"description": "Array of user profiles for the current page",
							"items":       map[string]interface{}{"$ref": "#/components/schemas/UserProfile"},
						},
						"meta": map[string]interface{}{
							"$ref": "#/components/schemas/Pagination",
						},
					},
				},
				"Pagination": map[string]interface{}{
					"type":        "object",
					"description": "Pagination metadata for building navigation controls",
					"properties": map[string]interface{}{
						"page":       map[string]interface{}{"type": "integer", "description": "Current page number (1-indexed)", "example": 1},
						"size":       map[string]interface{}{"type": "integer", "description": "Number of records per page", "example": 10},
						"total":      map[string]interface{}{"type": "integer", "description": "Total number of records matching the query across all pages", "example": 42},
						"totalPages": map[string]interface{}{"type": "integer", "description": "Total number of pages (calculated as ⌈total ÷ size⌉)", "example": 5},
					},
				},
				"UpdateUserRequest": map[string]interface{}{
					"type":        "object",
					"description": "Partial update payload — only include the fields you want to change. Omitted fields will not be modified.",
					"properties": map[string]interface{}{
						"name":   map[string]interface{}{"type": "string", "minLength": 2, "maxLength": 100, "description": "Updated display name", "example": "Jane Doe"},
						"phone":  map[string]interface{}{"type": "string", "description": "Updated phone number", "example": "+62898765432"},
						"status": map[string]interface{}{"type": "string", "enum": []string{"active", "inactive", "pending"}, "description": "Updated account status", "example": "active"},
					},
				},
				"DeleteResponse": map[string]interface{}{
					"type":        "object",
					"description": "Confirmation that a resource was deleted successfully",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"message": map[string]interface{}{"type": "string", "example": "user deleted successfully", "description": "Confirmation message"},
							},
						},
					},
				},
				"ErrorResponse": map[string]interface{}{
					"type":        "object",
					"description": "Standard error response with error code and human-readable message",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": false},
						"error": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"code":    map[string]interface{}{"type": "integer", "description": "Application-specific error code for programmatic handling", "example": 40901},
								"message": map[string]interface{}{"type": "string", "description": "Human-readable error description", "example": "email already registered"},
							},
						},
					},
				},
			},
		},
	}
}
