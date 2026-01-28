package config

import (
	"go-grst-boilerplate/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
)

func NewFiber(cfg *Config, log *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               cfg.ServiceName,
		DisableStartupMessage: true,
		ErrorHandler:          NewErrorHandler(log),
	})

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

	// Middleware (order matters!)
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.TracingMiddleware(cfg.ServiceName))
	app.Use(middleware.RecoveryMiddleware(log))
	app.Use(middleware.LoggerMiddleware(log))

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
