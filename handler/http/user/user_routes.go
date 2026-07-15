// Package user registers the HTTP routes for the user service.
package user

import (
	"time"

	"go-grst-boilerplate/pkg/errors"
	"go-grst-boilerplate/pkg/middleware"
	"go-grst-boilerplate/pkg/response"

	pb "go-grst-boilerplate/handler/grpc/user"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserRoutes struct {
	handler pb.UserApiServer
}

func NewUserRoutes(handler pb.UserApiServer) *UserRoutes {
	return &UserRoutes{handler: handler}
}

func (r *UserRoutes) RegisterRoutes(app *fiber.App, validator middleware.TokenValidator) {
	v1 := app.Group("/api/v1")

	// Stricter per-IP limit for unauthenticated credential endpoints, on top of
	// the global limiter, to blunt brute-force and enumeration attempts.
	authStrict := middleware.DefaultRateLimitConfig()
	authStrict.Max = 10
	authStrict.Duration = time.Minute
	authLimiter := middleware.RateLimitMiddleware(authStrict)

	// --- Auth routes ---
	auth := v1.Group("/auth")
	auth.Post("/register", authLimiter, r.Register)
	auth.Post("/login", authLimiter, r.Login)
	auth.Post("/refresh",
		middleware.AuthMiddleware(validator, mustAuthConfig("POST /api/v1/auth/refresh")),
		r.RefreshToken,
	)
	auth.Get("/me",
		middleware.AuthMiddleware(validator, mustAuthConfig("GET /api/v1/auth/me")),
		r.GetMe,
	)
	auth.Post("/logout",
		middleware.AuthMiddleware(validator, mustAuthConfig("POST /api/v1/auth/logout")),
		r.Logout,
	)

	// --- Users resource routes (admin only) ---
	users := v1.Group("/users")
	users.Get("/",
		middleware.AuthMiddleware(validator, mustAuthConfig("GET /api/v1/users")),
		r.ListUsers,
	)
	users.Get("/:id",
		middleware.AuthMiddleware(validator, mustAuthConfig("GET /api/v1/users/:id")),
		r.GetUser,
	)
	users.Put("/:id",
		middleware.AuthMiddleware(validator, mustAuthConfig("PUT /api/v1/users/:id")),
		r.UpdateUser,
	)
	users.Delete("/:id",
		middleware.AuthMiddleware(validator, mustAuthConfig("DELETE /api/v1/users/:id")),
		r.DeleteUser,
	)
}

// mustAuthConfig returns the auth policy for a route, panicking at startup if
// none is configured. This makes auth wiring fail-closed: a route can never be
// registered with an implicit zero-value (unauthenticated) policy because a
// missing entry crashes the process during route registration.
func mustAuthConfig(route string) middleware.AuthConfig {
	cfg, ok := pb.RouteAuthConfig[route]
	if !ok {
		panic("no auth policy configured for route: " + route)
	}
	return cfg
}

// --- Auth Handlers ---

func (r *UserRoutes) Register(c *fiber.Ctx) error {
	var req pb.RegisterReq
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, 400, "invalid request body")
	}

	ctx := c.UserContext()
	resp, err := r.handler.Register(ctx, &req)
	if err != nil {
		return handleError(c, err)
	}

	return response.CreatedProto(c, resp)
}

func (r *UserRoutes) Login(c *fiber.Ctx) error {
	var req pb.LoginReq
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, 400, "invalid request body")
	}

	ctx := c.UserContext()
	resp, err := r.handler.Login(ctx, &req)
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessProto(c, resp)
}

func (r *UserRoutes) RefreshToken(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = middleware.WithAuthContext(ctx, authCtx)
	}

	resp, err := r.handler.RefreshToken(ctx, &pb.RefreshTokenReq{})
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessProto(c, resp)
}

func (r *UserRoutes) GetMe(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = middleware.WithAuthContext(ctx, authCtx)
	}

	resp, err := r.handler.GetMe(ctx, &emptypb.Empty{})
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessProto(c, resp)
}

func (r *UserRoutes) Logout(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = middleware.WithAuthContext(ctx, authCtx)
	}

	resp, err := r.handler.Logout(ctx, &emptypb.Empty{})
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessProto(c, resp)
}

// --- Users Resource Handlers ---

func (r *UserRoutes) ListUsers(c *fiber.Ctx) error {
	req := &pb.ListUsersReq{
		Page:      int32(c.QueryInt("page", 1)),  // #nosec G115 -- pagination page is bounded
		Size:      int32(c.QueryInt("size", 10)), // #nosec G115 -- pagination size is bounded
		Search:    c.Query("search"),
		SortBy:    c.Query("sortBy", "created_at"),
		SortOrder: c.Query("sortOrder", "desc"),
	}

	ctx := c.UserContext()
	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = middleware.WithAuthContext(ctx, authCtx)
	}

	resp, err := r.handler.ListUsers(ctx, req)
	if err != nil {
		return handleError(c, err)
	}

	users := make([]proto.Message, len(resp.Users))
	for i, u := range resp.Users {
		users[i] = u
	}
	return response.SuccessProtoList(c, users, resp.Pagination)
}

func (r *UserRoutes) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, 400, "user id is required")
	}

	ctx := c.UserContext()
	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = middleware.WithAuthContext(ctx, authCtx)
	}

	resp, err := r.handler.GetUser(ctx, &pb.GetUserReq{Id: id})
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessProto(c, resp)
}

func (r *UserRoutes) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, 400, "user id is required")
	}

	var req pb.UpdateUserReq
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, 400, "invalid request body")
	}
	req.Id = id

	ctx := c.UserContext()
	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = middleware.WithAuthContext(ctx, authCtx)
	}

	resp, err := r.handler.UpdateUser(ctx, &req)
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessProto(c, resp)
}

func (r *UserRoutes) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, 400, "user id is required")
	}

	ctx := c.UserContext()
	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = middleware.WithAuthContext(ctx, authCtx)
	}

	resp, err := r.handler.DeleteUser(ctx, &pb.DeleteUserReq{Id: id})
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessProto(c, resp)
}

// --- Error Helper ---

func handleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.FiberError(c)
	}
	return response.InternalError(c, 500, "internal server error")
}
