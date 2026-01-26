package docs

import (
	"github.com/gofiber/fiber/v2"
	scalar "github.com/yokeTH/gofiber-scalar"
)

// @title Go-GRST-Boilerplate API
// @version 1.0
// @description A production-ready Go monolithic application boilerplate using Go Fiber for REST API, gRPC for service-to-service communication, with Domain-Driven Design (DDD) and Clean Architecture.

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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
			"description": "A production-ready Go monolithic application boilerplate using Go Fiber for REST API, gRPC for service-to-service communication, with Domain-Driven Design (DDD) and Clean Architecture.",
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
			{"url": "http://localhost:3000", "description": "Development server"},
		},
		"tags": []map[string]string{
			{"name": "Auth", "description": "Authentication endpoints"},
			{"name": "User", "description": "User management endpoints"},
			{"name": "Admin", "description": "Admin endpoints"},
			{"name": "Employee", "description": "Employee endpoints"},
			{"name": "Health", "description": "Health check endpoints"},
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Health"},
					"summary":     "Health check",
					"description": "Returns the health status of the service",
					"operationId": "healthCheck",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
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
					"summary":     "Readiness check",
					"description": "Returns the readiness status including dependency health",
					"operationId": "readinessCheck",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is ready",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ReadinessResponse",
									},
								},
							},
						},
						"503": map[string]interface{}{
							"description": "Service is not ready",
						},
					},
				},
			},
			"/api/v1/auth/register": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Auth"},
					"summary":     "Register a new user",
					"description": "Creates a new user account",
					"operationId": "register",
					"requestBody": map[string]interface{}{
						"required": true,
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
							"description": "User created successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/RegisterResponse",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Validation error",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
						"409": map[string]interface{}{
							"description": "Email already exists",
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
					"summary":     "Login user",
					"description": "Authenticates a user and returns a JWT token",
					"operationId": "login",
					"requestBody": map[string]interface{}{
						"required": true,
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
							"description": "Login successful",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/LoginResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Invalid credentials",
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
			"/api/v1/user/profile": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"User"},
					"summary":     "Get user profile",
					"description": "Returns the profile of the authenticated user",
					"operationId": "getProfile",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User profile",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/UserProfileResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
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
			"/api/v1/admin/users": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "List all users",
					"description": "Returns a paginated list of all users (admin only)",
					"operationId": "listUsers",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number",
							"schema":      map[string]interface{}{"type": "integer", "default": 1},
						},
						{
							"name":        "size",
							"in":          "query",
							"description": "Page size",
							"schema":      map[string]interface{}{"type": "integer", "default": 10, "maximum": 100},
						},
						{
							"name":        "search",
							"in":          "query",
							"description": "Search term",
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "sortBy",
							"in":          "query",
							"description": "Sort field",
							"schema":      map[string]interface{}{"type": "string", "enum": []string{"created_at", "name", "email"}},
						},
						{
							"name":        "sortOrder",
							"in":          "query",
							"description": "Sort order",
							"schema":      map[string]interface{}{"type": "string", "enum": []string{"asc", "desc"}},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of users",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ListUsersResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
				},
			},
			"/api/v1/employee/payslip": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Employee"},
					"summary":     "Get payslip",
					"description": "Returns the payslip for a specific month and year",
					"operationId": "getPayslip",
					"security":    []map[string][]string{{"BearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "year",
							"in":          "query",
							"required":    true,
							"description": "Year",
							"schema":      map[string]interface{}{"type": "integer", "minimum": 2000, "maximum": 2100},
						},
						{
							"name":        "month",
							"in":          "query",
							"required":    true,
							"description": "Month",
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 12},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Payslip data",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/PayslipResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Employee role required",
						},
						"404": map[string]interface{}{
							"description": "Payslip not found",
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
					"bearerFormat": "JWT",
					"description":  "Enter your JWT token",
				},
			},
			"schemas": map[string]interface{}{
				"HealthResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{"type": "string", "example": "ok"},
					},
				},
				"ReadinessResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{"type": "string", "example": "ready"},
						"checks": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"database": map[string]interface{}{"type": "string", "example": "healthy"},
								"redis":    map[string]interface{}{"type": "string", "example": "healthy"},
								"rabbitmq": map[string]interface{}{"type": "string", "example": "healthy"},
							},
						},
					},
				},
				"RegisterRequest": map[string]interface{}{
					"type":     "object",
					"required": []string{"email", "password", "name"},
					"properties": map[string]interface{}{
						"email":    map[string]interface{}{"type": "string", "format": "email", "example": "user@example.com"},
						"password": map[string]interface{}{"type": "string", "minLength": 8, "example": "password123"},
						"name":     map[string]interface{}{"type": "string", "minLength": 2, "example": "John Doe"},
						"phone":    map[string]interface{}{"type": "string", "example": "081234567890"},
					},
				},
				"RegisterResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":    map[string]interface{}{"type": "string", "format": "uuid"},
								"email": map[string]interface{}{"type": "string"},
								"name":  map[string]interface{}{"type": "string"},
							},
						},
					},
				},
				"LoginRequest": map[string]interface{}{
					"type":     "object",
					"required": []string{"email", "password"},
					"properties": map[string]interface{}{
						"email":    map[string]interface{}{"type": "string", "format": "email", "example": "user@example.com"},
						"password": map[string]interface{}{"type": "string", "example": "password123"},
					},
				},
				"LoginResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"token":      map[string]interface{}{"type": "string"},
								"expires_at": map[string]interface{}{"type": "string", "format": "date-time"},
							},
						},
					},
				},
				"UserProfileResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"$ref": "#/components/schemas/UserProfile",
						},
					},
				},
				"UserProfile": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":        map[string]interface{}{"type": "string", "format": "uuid"},
						"email":     map[string]interface{}{"type": "string", "format": "email"},
						"name":      map[string]interface{}{"type": "string"},
						"phone":     map[string]interface{}{"type": "string"},
						"status":    map[string]interface{}{"type": "string", "enum": []string{"active", "inactive", "pending"}},
						"roles":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"createdAt": map[string]interface{}{"type": "string", "format": "date-time"},
					},
				},
				"ListUsersResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"$ref": "#/components/schemas/UserProfile"},
						},
						"meta": map[string]interface{}{
							"$ref": "#/components/schemas/Pagination",
						},
					},
				},
				"Pagination": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"page":       map[string]interface{}{"type": "integer"},
						"size":       map[string]interface{}{"type": "integer"},
						"total":      map[string]interface{}{"type": "integer"},
						"totalPages": map[string]interface{}{"type": "integer"},
					},
				},
				"PayslipResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"$ref": "#/components/schemas/Payslip",
						},
					},
				},
				"Payslip": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":          map[string]interface{}{"type": "string", "format": "uuid"},
						"employeeId":  map[string]interface{}{"type": "string", "format": "uuid"},
						"year":        map[string]interface{}{"type": "integer"},
						"month":       map[string]interface{}{"type": "integer"},
						"basicSalary": map[string]interface{}{"type": "number"},
						"allowances":  map[string]interface{}{"type": "number"},
						"deductions":  map[string]interface{}{"type": "number"},
						"grossSalary": map[string]interface{}{"type": "number"},
						"netSalary":   map[string]interface{}{"type": "number"},
						"status":      map[string]interface{}{"type": "string"},
						"paidAt":      map[string]interface{}{"type": "string", "format": "date-time"},
					},
				},
				"ErrorResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": false},
						"error": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"code":    map[string]interface{}{"type": "integer"},
								"message": map[string]interface{}{"type": "string"},
							},
						},
					},
				},
			},
		},
	}
}
