package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-grst-boilerplate/app/usecase/user"
	"go-grst-boilerplate/config"
	"go-grst-boilerplate/handler"
	pb_user "go-grst-boilerplate/handler/grpc/user"
	http_user "go-grst-boilerplate/handler/http/user"
	"go-grst-boilerplate/pkg/database"
	"go-grst-boilerplate/pkg/jwt"
	"go-grst-boilerplate/pkg/logger"
	"go-grst-boilerplate/pkg/middleware"
	"go-grst-boilerplate/pkg/rabbitmq"
	"go-grst-boilerplate/pkg/redis"
	"go-grst-boilerplate/pkg/telemetry"
	"go-grst-boilerplate/repository/user_repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration using Viper
	cfg, err := config.New()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger (Zap)
	log, err := logger.New(logger.Config{
		Level:       cfg.LogLevel,
		Format:      cfg.LogFormat,
		Environment: cfg.Environment,
		ServiceName: cfg.ServiceName,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting application",
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
		zap.Int("http_port", cfg.HTTPPort),
		zap.Int("grpc_port", cfg.GRPCPort),
	)

	// Initialize OpenTelemetry
	ctx := context.Background()
	otel, err := telemetry.New(ctx, telemetry.Config{
		ServiceName:  cfg.OTelServiceName,
		Environment:  cfg.Environment,
		Endpoint:     cfg.OTelEndpoint,
		ExporterType: cfg.OTelExporterType,
		Enabled:      cfg.OTelEnabled,
	})
	if err != nil {
		log.Fatal("Failed to initialize telemetry", zap.Error(err))
	}
	defer otel.Shutdown(ctx)

	log.Info("OpenTelemetry initialized",
		zap.Bool("enabled", cfg.OTelEnabled),
		zap.String("exporter", cfg.OTelExporterType),
	)

	// Initialize database
	db, err := database.New(cfg, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Auto migrate
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize Redis (using redigo)
	redisClient, err := redis.New(redis.Config{
		Host:        cfg.RedisHost,
		Port:        cfg.RedisPort,
		Password:    cfg.RedisPassword,
		DB:          cfg.RedisDB,
		MaxIdle:     cfg.RedisMaxIdle,
		MaxActive:   cfg.RedisMaxActive,
		IdleTimeout: cfg.RedisIdleTimeout,
	})
	if err != nil {
		log.Warn("Failed to connect to Redis, caching disabled", zap.Error(err))
	} else {
		defer redisClient.Close()
		log.Info("Redis connection established",
			zap.String("host", cfg.RedisHost),
			zap.Int("port", cfg.RedisPort),
		)
	}

	// Initialize RabbitMQ
	rabbitClient, err := rabbitmq.New(rabbitmq.Config{
		Host:     cfg.RabbitMQHost,
		Port:     cfg.RabbitMQPort,
		User:     cfg.RabbitMQUser,
		Password: cfg.RabbitMQPassword,
		VHost:    cfg.RabbitMQVHost,
	}, log.Logger)
	if err != nil {
		log.Warn("Failed to connect to RabbitMQ, messaging disabled", zap.Error(err))
	} else {
		defer rabbitClient.Close()
		log.Info("RabbitMQ connection established",
			zap.String("host", cfg.RabbitMQHost),
			zap.Int("port", cfg.RabbitMQPort),
		)
	}

	// Initialize layers
	userRepo := user_repository.New(db)
	userUC := user.NewUseCase(userRepo)
	tokenService := jwt.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)
	userHandler := handler.NewUserHandler(userUC, tokenService)

	// Create token validator
	tokenValidator := createTokenValidator(tokenService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               cfg.ServiceName,
		DisableStartupMessage: true,
		ErrorHandler:          createErrorHandler(log.Logger),
	})

	// Setup CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Trace-ID",
		AllowCredentials: true,
	}))

	// Setup middleware (order matters!)
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.TracingMiddleware(cfg.ServiceName))
	app.Use(middleware.RecoveryMiddleware(log.Logger))
	app.Use(middleware.LoggerMiddleware(log.Logger))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": cfg.ServiceName,
		})
	})

	// Readiness check (includes dependencies)
	app.Get("/ready", func(c *fiber.Ctx) error {
		checks := make(map[string]string)

		// Check database
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			checks["database"] = "unhealthy"
		} else {
			checks["database"] = "healthy"
		}

		// Check Redis
		if redisClient != nil {
			conn := redisClient.Conn()
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
		if rabbitClient != nil {
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

	// Register HTTP routes
	userRoutes := http_user.NewUserRoutes(userHandler)
	userRoutes.RegisterRoutes(app, tokenValidator)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb_user.RegisterUserApiServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	// Start servers
	errChan := make(chan error, 2)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			errChan <- fmt.Errorf("gRPC listen error: %w", err)
			return
		}
		log.Info("gRPC server listening", zap.Int("port", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("gRPC serve error: %w", err)
		}
	}()

	// Start Fiber server
	go func() {
		log.Info("HTTP server listening", zap.Int("port", cfg.HTTPPort))
		if err := app.Listen(fmt.Sprintf(":%d", cfg.HTTPPort)); err != nil {
			errChan <- fmt.Errorf("Fiber listen error: %w", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Shutting down servers...")

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		grpcServer.GracefulStop()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Error("Fiber shutdown error", zap.Error(err))
		}

		log.Info("Servers stopped gracefully")
	case err := <-errChan:
		log.Fatal("Server error", zap.Error(err))
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

func createErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		log.Error("Request error",
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
			zap.String("path", c.Path()),
		)

		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    code,
				"message": message,
			},
		})
	}
}
