package handler

import (
	"context"
	"time"

	"go-grst-boilerplate/app/usecase/user"
	pb "go-grst-boilerplate/handler/grpc/user"
	"go-grst-boilerplate/pkg/errors"
	"go-grst-boilerplate/pkg/middleware"
	"go-grst-boilerplate/pkg/token"

	"google.golang.org/protobuf/types/known/emptypb"
)

type userHandler struct {
	pb.UnimplementedUserApiServer
	userUC       user.UseCase
	tokenService *token.TokenService
}

func NewUserHandler(userUC user.UseCase, tokenService *token.TokenService) pb.UserApiServer {
	return &userHandler{
		userUC:       userUC,
		tokenService: tokenService,
	}
}

// Register creates a new user account.
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

// Login authenticates a user and returns a PASETO token.
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

	// Generate PASETO token
	accessToken, err := h.tokenService.GenerateToken(
		userEntity.ID,
		userEntity.Email,
		userEntity.Roles,
		userEntity.CompanyCode,
	)
	if err != nil {
		return nil, errors.Internal(50003, "failed to generate token")
	}

	return &pb.LoginRes{
		Token: accessToken,
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

// RefreshToken re-issues a new PASETO token from a valid existing token.
func (h *userHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenRes, error) {
	authCtx := getAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.Unauthorized("authentication required")
	}

	// Re-issue a new token with the same claims
	newToken, err := h.tokenService.GenerateToken(
		authCtx.UserID,
		authCtx.Email,
		authCtx.Roles,
		authCtx.CompanyCode,
	)
	if err != nil {
		return nil, errors.Internal(50010, "failed to refresh token")
	}

	return &pb.RefreshTokenRes{
		Token: newToken,
	}, nil
}

// GetMe returns the profile of the currently authenticated user.
func (h *userHandler) GetMe(ctx context.Context, req *emptypb.Empty) (*pb.UserProfile, error) {
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

// Logout invalidates the current session (stateless — returns success).
func (h *userHandler) Logout(ctx context.Context, req *emptypb.Empty) (*pb.LogoutRes, error) {
	authCtx := getAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.Unauthorized("authentication required")
	}

	// Stateless logout — token is not blacklisted.
	// Client should discard the token on their side.
	return &pb.LogoutRes{
		Message: "successfully logged out",
	}, nil
}

// ListUsers returns a paginated list of all users (admin only).
func (h *userHandler) ListUsers(ctx context.Context, req *pb.ListUsersReq) (*pb.ListUsersRes, error) {
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
			TotalPages: int32(totalPages), // #nosec G115 -- totalPages is bounded by pagination
		},
	}, nil
}

// GetUser returns a single user by ID (admin only).
func (h *userHandler) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.UserProfile, error) {
	userEntity, err := h.userUC.GetUser(ctx, req.Id)
	if err != nil {
		if err == user.ErrNotFound {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal(50006, "failed to get user")
	}

	return &pb.UserProfile{
		Id:        userEntity.ID,
		Email:     userEntity.Email,
		Name:      userEntity.Name,
		Phone:     userEntity.Phone,
		Status:    string(userEntity.Status),
		CreatedAt: userEntity.CreatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateUser updates a user by ID (admin only).
func (h *userHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserReq) (*pb.UserProfile, error) {
	if err := pb.ValidateRequest(req); err != nil {
		return nil, err
	}

	userEntity, err := h.userUC.UpdateUser(ctx, req.Id, user.UpdateInput{
		Name:   req.Name,
		Phone:  req.Phone,
		Status: req.Status,
	})
	if err != nil {
		if err == user.ErrNotFound {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal(50007, "failed to update user")
	}

	return &pb.UserProfile{
		Id:        userEntity.ID,
		Email:     userEntity.Email,
		Name:      userEntity.Name,
		Phone:     userEntity.Phone,
		Status:    string(userEntity.Status),
		CreatedAt: userEntity.CreatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteUser soft-deletes a user by ID (admin only).
func (h *userHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserReq) (*pb.DeleteUserRes, error) {
	err := h.userUC.DeleteUser(ctx, req.Id)
	if err != nil {
		if err == user.ErrNotFound {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal(50008, "failed to delete user")
	}

	return &pb.DeleteUserRes{
		Message: "user deleted successfully",
	}, nil
}

func getAuthFromContext(ctx context.Context) *middleware.AuthContext {
	if auth, ok := ctx.Value("auth").(*middleware.AuthContext); ok {
		return auth
	}
	return nil
}
