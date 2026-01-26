package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
		}

		if requestID, ok := c.Locals("request_id").(string); ok {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Add trace context if available
		if span := trace.SpanFromContext(c.UserContext()); span.SpanContext().IsValid() {
			fields = append(fields,
				zap.String("trace_id", span.SpanContext().TraceID().String()),
				zap.String("span_id", span.SpanContext().SpanID().String()),
			)
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Warn("Request failed", fields...)
		} else if c.Response().StatusCode() >= 500 {
			logger.Error("Request completed with server error", fields...)
		} else if c.Response().StatusCode() >= 400 {
			logger.Warn("Request completed with client error", fields...)
		} else {
			logger.Info("Request completed", fields...)
		}

		return err
	}
}
