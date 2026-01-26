package user

import (
	"go-grst-boilerplate/pkg/middleware"
	"go-grst-boilerplate/pkg/validation"
)

// AuthConfigMethods contains auth config for each gRPC method
var AuthConfigMethods = map[string]middleware.AuthConfig{
	"/user.UserApi/Register":     {NeedAuth: false, AllowedRoles: nil},
	"/user.UserApi/Login":        {NeedAuth: false, AllowedRoles: nil},
	"/user.UserApi/GetProfile":   {NeedAuth: true, AllowedRoles: nil},
	"/user.UserApi/ListAllUsers": {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"/user.UserApi/GetMyPayslip": {NeedAuth: true, AllowedRoles: []string{"employee"}},
}

// RouteAuthConfig contains auth config for each REST route
var RouteAuthConfig = map[string]middleware.AuthConfig{
	"POST /api/v1/auth/register":   {NeedAuth: false, AllowedRoles: nil},
	"POST /api/v1/auth/login":      {NeedAuth: false, AllowedRoles: nil},
	"GET /api/v1/user/profile":     {NeedAuth: true, AllowedRoles: nil},
	"GET /api/v1/admin/users":      {NeedAuth: true, AllowedRoles: []string{"admin", "superadmin"}},
	"GET /api/v1/employee/payslip": {NeedAuth: true, AllowedRoles: []string{"employee"}},
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

type GetPayslipRequest struct {
	Year  int32 `json:"year" validate:"required,gte=2000,lte=2100"`
	Month int32 `json:"month" validate:"required,gte=1,lte=12"`
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

	case *GetPayslipReq:
		validateReq := GetPayslipRequest{
			Year:  r.Year,
			Month: r.Month,
		}
		return validation.Validate(validateReq)
	}

	return nil
}
