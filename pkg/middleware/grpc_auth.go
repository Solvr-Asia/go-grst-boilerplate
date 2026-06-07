package middleware

import (
	"context"
	"strings"

	"go-grst-boilerplate/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const grpcAuthContextKey = "auth"

func GRPCAuthInterceptor(validator TokenValidator, authConfig map[string]AuthConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		config, ok := authConfig[info.FullMethod]
		if !ok || !config.NeedAuth {
			return handler(ctx, req)
		}

		token, err := extractBearerToken(ctx)
		if err != nil {
			return nil, errors.Unauthorized("missing authorization").GRPCStatus().Err()
		}

		authCtx, err := validator(token)
		if err != nil {
			return nil, errors.Unauthorized("invalid token").GRPCStatus().Err()
		}

		if len(config.AllowedRoles) > 0 && !hasAnyRole(authCtx.Roles, config.AllowedRoles) {
			return nil, errors.Forbidden("insufficient permissions").GRPCStatus().Err()
		}

		authCtx.Token = token
		return handler(context.WithValue(ctx, grpcAuthContextKey, authCtx), req)
	}
}

func GetGRPCAuthContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(grpcAuthContextKey).(*AuthContext)
	return authCtx, ok
}

func extractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.Unauthorized("missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.Unauthorized("missing authorization header")
	}

	parts := strings.SplitN(authHeaders[0], " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] == "" {
		return "", errors.Unauthorized("invalid authorization format")
	}

	return parts[1], nil
}
