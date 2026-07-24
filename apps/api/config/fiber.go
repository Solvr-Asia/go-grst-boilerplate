package config

import (
	"time"

	"veemon/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"go.uber.org/zap"
)

func NewFiber(cfg *Config, log *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               cfg.ServiceName,
		DisableStartupMessage: true,
		ErrorHandler:          NewErrorHandler(log),
		Prefork:               cfg.Prefork,
		// Bound slow/idle connections so a stuck client cannot hold a worker
		// indefinitely. Values are conservative defaults; tune per workload.
		ReadTimeout:  time.Duration(cfg.HTTPReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTPWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.HTTPIdleTimeout) * time.Second,
	})

	// Security response headers (X-Frame-Options, X-Content-Type-Options, etc.)
	app.Use(helmet.New())

	// CORS
	corsConfig := cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Trace-ID",
	}
	// AllowCredentials cannot be used with wildcard origins
	if cfg.CORSOrigins != "*" {
		corsConfig.AllowCredentials = true
	}
	app.Use(cors.New(corsConfig))

	// Middleware (order matters — outermost first). Logger is placed OUTSIDE
	// Recovery so that a panic (recovered below into a 500) is still logged and
	// traced; Recovery wraps the handler so it can turn panics into responses.
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.TracingMiddleware(cfg.ServiceName))
	app.Use(middleware.LoggerMiddleware(log))
	// Global per-IP rate limit as a coarse abuse guard. Stricter, endpoint-
	// specific limits are applied on auth routes during route registration.
	app.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimitConfig()))
	app.Use(middleware.RecoveryMiddleware(log))
	app.Use(middleware.TimeoutMiddleware(time.Duration(cfg.RequestTimeout) * time.Second))

	return app
}

func NewErrorHandler(log *zap.Logger) fiber.ErrorHandler {
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
