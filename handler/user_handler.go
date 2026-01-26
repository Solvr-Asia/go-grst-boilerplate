package handler

import (
	"context"
	"time"

	"go-grst-boilerplate/app/usecase/user"
	pb "go-grst-boilerplate/handler/grpc/user"
	"go-grst-boilerplate/pkg/errors"
	"go-grst-boilerplate/pkg/jwt"
	"go-grst-boilerplate/pkg/middleware"

	"google.golang.org/protobuf/types/known/emptypb"
)

type userHandler struct {
	pb.UnimplementedUserApiServer
	userUC       user.UseCase
	tokenService *jwt.TokenService
}

func NewUserHandler(userUC user.UseCase, tokenService *jwt.TokenService) pb.UserApiServer {
	return &userHandler{
		userUC:       userUC,
		tokenService: tokenService,
	}
}

func (h *userHandler) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error) {
	if err := pb.ValidateRequest(req); err != nil {
		return nil, err
	}

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

func (h *userHandler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	if err := pb.ValidateRequest(req); err != nil {
		return nil, err
	}

	userEntity, err := h.userUC.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == user.ErrInvalidCreds {
			return nil, errors.Unauthorized("invalid email or password")
		}
		return nil, errors.Internal(50002, "failed to login")
	}

	// Generate JWT token
	token, err := h.tokenService.GenerateToken(
		userEntity.ID,
		userEntity.Email,
		[]string{"employee"}, // Default role
		"",                   // Company code
	)
	if err != nil {
		return nil, errors.Internal(50003, "failed to generate token")
	}

	return &pb.LoginRes{
		Token: token,
		User: &pb.UserProfile{
			Id:        userEntity.ID,
			Email:     userEntity.Email,
			Name:      userEntity.Name,
			Phone:     userEntity.Phone,
			Status:    string(userEntity.Status),
			CreatedAt: userEntity.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *userHandler) GetProfile(ctx context.Context, req *emptypb.Empty) (*pb.UserProfile, error) {
	authCtx := getAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.Unauthorized("authentication required")
	}

	profile, err := h.userUC.GetProfile(ctx, authCtx.UserID)
	if err != nil {
		if err == user.ErrNotFound {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal(50004, "failed to get profile")
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
	if err := pb.ValidateRequest(req); err != nil {
		return nil, err
	}

	users, total, err := h.userUC.ListAll(ctx, user.ListInput{
		Page:      int(req.Page),
		Size:      int(req.Size),
		Search:    req.Search,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, errors.Internal(50005, "failed to list users")
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
	if err := pb.ValidateRequest(req); err != nil {
		return nil, err
	}

	authCtx := getAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.Unauthorized("authentication required")
	}

	payslip, err := h.userUC.GetPayslip(ctx, authCtx.UserID, int(req.Year), int(req.Month))
	if err != nil {
		if err == user.ErrPayslipNotFound {
			return nil, errors.NotFound("payslip not found")
		}
		return nil, errors.Internal(50006, "failed to get payslip")
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
