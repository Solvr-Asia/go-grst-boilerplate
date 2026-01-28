package config

import (
	"go-grst-boilerplate/app/usecase/user"
	"go-grst-boilerplate/handler"
	pb_user "go-grst-boilerplate/handler/grpc/user"
	http_user "go-grst-boilerplate/handler/http/user"
	"go-grst-boilerplate/pkg/jwt"
	"go-grst-boilerplate/pkg/middleware"
	"go-grst-boilerplate/pkg/rabbitmq"
	"go-grst-boilerplate/pkg/redis"
	"go-grst-boilerplate/repository/user_repository"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

// BootstrapConfig holds all dependencies for application wiring.
type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *zap.Logger
	Cfg      *Config
	Redis    *redis.Client
	RabbitMQ *rabbitmq.Client
}

// BootstrapResult holds the wired components ready to be started.
type BootstrapResult struct {
	GRPCServer *grpc.Server
}

// Bootstrap wires repositories, usecases, handlers, and routes.
func Bootstrap(b *BootstrapConfig) *BootstrapResult {
	// Layers
	userRepo := user_repository.New(b.DB)
	userUC := user.NewUseCase(userRepo)
	tokenService := jwt.NewTokenService(b.Cfg.JWTSecret, b.Cfg.JWTExpiration)
	userHandler := handler.NewUserHandler(userUC, tokenService)

	// Token validator
	tokenValidator := createTokenValidator(tokenService)

	// Health check
	registerHealthChecks(b)

	// HTTP routes
	userRoutes := http_user.NewUserRoutes(userHandler)
	userRoutes.RegisterRoutes(b.App, tokenValidator)

	// gRPC server
	grpcServer := grpc.NewServer()
	pb_user.RegisterUserApiServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	return &BootstrapResult{
		GRPCServer: grpcServer,
	}
}

func createTokenValidator(tokenService *jwt.TokenService) middleware.TokenValidator {
	return func(token string) (*middleware.AuthContext, error) {
		claims, err := tokenService.ValidateToken(token)
		if err != nil {
			return nil, err
		}

		return &middleware.AuthContext{
			UserID:      claims.UserID,
			Email:       claims.Email,
			Roles:       claims.Roles,
			CompanyCode: claims.CompanyCode,
			Token:       token,
		}, nil
	}
}

func registerHealthChecks(b *BootstrapConfig) {
	b.App.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": b.Cfg.ServiceName,
		})
	})

	b.App.Get("/ready", func(c *fiber.Ctx) error {
		checks := make(map[string]string)

		// Check database
		sqlDB, err := b.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			checks["database"] = "unhealthy"
		} else {
			checks["database"] = "healthy"
		}

		// Check Redis
		if b.Redis != nil {
			conn := b.Redis.Conn()
			_, err := conn.Do("PING")
			conn.Close()
			if err != nil {
				checks["redis"] = "unhealthy"
			} else {
				checks["redis"] = "healthy"
			}
		} else {
			checks["redis"] = "disabled"
		}

		// Check RabbitMQ
		if b.RabbitMQ != nil {
			checks["rabbitmq"] = "healthy"
		} else {
			checks["rabbitmq"] = "disabled"
		}

		allHealthy := true
		for _, v := range checks {
			if v == "unhealthy" {
				allHealthy = false
				break
			}
		}

		status := fiber.StatusOK
		if !allHealthy {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(fiber.Map{
			"status": checks,
		})
	})
}
