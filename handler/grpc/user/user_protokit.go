package user

import (
	"go-grst-boilerplate/pkg/middleware"
	"go-grst-boilerplate/pkg/validation"
)

// AuthConfigMethods contains auth config for each gRPC method
var AuthConfigMethods = map[string]middleware.AuthConfig{
	"/user.UserApi/Register":     {NeedAuth: false, AllowedRoles: nil},
	"/user.UserApi/Login":        {NeedAuth: false, AllowedRoles: nil},
	"/user.UserApi/RefreshToken": {NeedAuth: true, AllowedRoles: nil},
	"/user.UserApi/GetMe":        {NeedAuth: true, AllowedRoles: nil},
	"/user.UserApi/Logout":       {NeedAuth: true, AllowedRoles: nil},
	"/user.UserApi/ListUsers":    {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"/user.UserApi/GetUser":      {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"/user.UserApi/UpdateUser":   {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"/user.UserApi/DeleteUser":   {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
}

// RouteAuthConfig contains auth config for each REST route
var RouteAuthConfig = map[string]middleware.AuthConfig{
	"POST /api/v1/auth/register": {NeedAuth: false, AllowedRoles: nil},
	"POST /api/v1/auth/login":    {NeedAuth: false, AllowedRoles: nil},
	"POST /api/v1/auth/refresh":  {NeedAuth: true, AllowedRoles: nil},
	"GET /api/v1/auth/me":        {NeedAuth: true, AllowedRoles: nil},
	"POST /api/v1/auth/logout":   {NeedAuth: true, AllowedRoles: nil},
	"GET /api/v1/users":          {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"GET /api/v1/users/:id":      {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"PUT /api/v1/users/:id":      {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"DELETE /api/v1/users/:id":   {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
}

// Request DTOs with validation tags

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Phone    string `json:"phone" validate:"omitempty,phone"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ListUsersRequest struct {
	Page      int32  `json:"page" validate:"omitempty,gte=1"`
	Size      int32  `json:"size" validate:"omitempty,gte=1,lte=100"`
	Search    string `json:"search" validate:"omitempty,max=100"`
	SortBy    string `json:"sortBy" validate:"omitempty,oneof=created_at name email"`
	SortOrder string `json:"sortOrder" validate:"omitempty,oneof=asc desc"`
}

type UpdateUserRequest struct {
	Name   string `json:"name" validate:"omitempty,min=2,max=100"`
	Phone  string `json:"phone" validate:"omitempty"`
	Status string `json:"status" validate:"omitempty,oneof=active inactive pending"`
}

// ValidateRequest validates proto request messages using go-playground/validator
func ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *RegisterReq:
		validateReq := RegisterRequest{
			Email:    r.Email,
			Password: r.Password,
			Name:     r.Name,
			Phone:    r.Phone,
		}
		return validation.Validate(validateReq)

	case *LoginReq:
		validateReq := LoginRequest{
			Email:    r.Email,
			Password: r.Password,
		}
		return validation.Validate(validateReq)

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

		validateReq := ListUsersRequest{
			Page:      r.Page,
			Size:      r.Size,
			Search:    r.Search,
			SortBy:    r.SortBy,
			SortOrder: r.SortOrder,
		}
		return validation.Validate(validateReq)

	case *UpdateUserReq:
		validateReq := UpdateUserRequest{
			Name:   r.Name,
			Phone:  r.Phone,
			Status: r.Status,
		}
		return validation.Validate(validateReq)
	}

	return nil
}
