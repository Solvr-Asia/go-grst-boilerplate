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
	api := app.Group("/api/v1")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/register", r.Register)
	auth.Post("/login", r.Login)

	// Protected routes
	user := api.Group("/user")
	user.Get("/profile",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/user/profile"]),
		r.GetProfile,
	)

	// Admin routes
	admin := api.Group("/admin")
	admin.Get("/users",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/admin/users"]),
		r.ListAllUsers,
	)

	// Employee routes
	employee := api.Group("/employee")
	employee.Get("/payslip",
		middleware.AuthMiddleware(validator, pb.RouteAuthConfig["GET /api/v1/employee/payslip"]),
		r.GetMyPayslip,
	)
}

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

func (r *UserRoutes) GetProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()

	if authCtx, ok := middleware.GetAuthContext(c); ok {
		ctx = context.WithValue(ctx, "auth", authCtx)
	}

	resp, err := r.handler.GetProfile(ctx, &emptypb.Empty{})
	if err != nil {
		return handleError(c, err)
	}

	return response.Success(c, resp)
}

func (r *UserRoutes) ListAllUsers(c *fiber.Ctx) error {
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

	resp, err := r.handler.ListAllUsers(ctx, req)
	if err != nil {
		return handleError(c, err)
	}

	return response.SuccessWithMeta(c, resp.Users, resp.Pagination)
}

func (r *UserRoutes) GetMyPayslip(c *fiber.Ctx) error {
	req := &pb.GetPayslipReq{
		Year:  int32(c.QueryInt("year")),  // #nosec G115 -- year is bounded (reasonable calendar year)
		Month: int32(c.QueryInt("month")), // #nosec G115 -- month is 1-12
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
