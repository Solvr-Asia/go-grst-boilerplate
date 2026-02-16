package user

import (
	"context"

	"go-grst-boilerplate/pkg/errors"
	"go-grst-boilerplate/pkg/middleware"
	"go-grst-boilerplate/pkg/response"

	pb "go-grst-boilerplate/handler/grpc/user"

	"github.com/gofiber/fiber/v2"
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

	// --- Auth routes ---
	auth := v1.Group("/auth")
	auth.Post("/register", r.Register)
	auth.Post("/login", r.Login)
	auth.Post("/refresh",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["POST /api/v1/auth/refresh"]),
		r.RefreshToken,
	)
	auth.Get("/me",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/auth/me"]),
		r.GetMe,
	)
	auth.Post("/logout",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["POST /api/v1/auth/logout"]),
		r.Logout,
	)

	// --- Users resource routes (admin only) ---
	users := v1.Group("/users")
	users.Get("/",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/users"]),
		r.ListUsers,
	)
	users.Get("/:id",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/users/:id"]),
		r.GetUser,
	)
	users.Put("/:id",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["PUT /api/v1/users/:id"]),
		r.UpdateUser,
	)
	users.Delete("/:id",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["DELETE /api/v1/users/:id"]),
		r.DeleteUser,
	)
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

	return response.Created(c, resp)
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

	return response.Success(c, resp)
}

func (r *UserRoutes) RefreshToken(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.RefreshToken(ctx, &pb.RefreshTokenReq{})
	if err != nil {
		return handleError(c, err)
	}

	return response.Success(c, resp)
}

func (r *UserRoutes) GetMe(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.GetMe(ctx, &emptypb.Empty{})
	if err != nil {
		return handleError(c, err)
	}

	return response.Success(c, resp)
}

func (r *UserRoutes) Logout(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.Logout(ctx, &emptypb.Empty{})
	if err != nil {
		return handleError(c, err)
	}

	return response.Success(c, resp)
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
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.ListUsers(ctx, req)
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessWithMeta(c, resp.Users, resp.Pagination)
}

func (r *UserRoutes) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, 400, "user id is required")
	}

	ctx := c.UserContext()
	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.GetUser(ctx, &pb.GetUserReq{Id: id})
	if err != nil {
		return handleError(c, err)
	}

	return response.Success(c, resp)
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
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.UpdateUser(ctx, &req)
	if err != nil {
		return handleError(c, err)
	}

	return response.Success(c, resp)
}

func (r *UserRoutes) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, 400, "user id is required")
	}

	ctx := c.UserContext()
	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.DeleteUser(ctx, &pb.DeleteUserReq{Id: id})
	if err != nil {
		return handleError(c, err)
	}

	return response.Success(c, resp)
}

// --- Error Helper ---

func handleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.FiberError(c)
	}
	return response.InternalError(c, 500, "internal server error")
}
