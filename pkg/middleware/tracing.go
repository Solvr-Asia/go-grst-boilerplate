package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("fiber-middleware")

func TracingMiddleware(serviceName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract trace context from incoming request headers
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.UserContext(), propagation.HeaderCarrier(c.GetReqHeaders()))

		// Start a new span
		spanName := c.Method() + " " + c.Path()
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Method()),
				semconv.HTTPURLKey.String(c.OriginalURL()),
				semconv.HTTPRouteKey.String(c.Route().Path),
				semconv.NetHostNameKey.String(c.Hostname()),
				semconv.UserAgentOriginalKey.String(c.Get("User-Agent")),
				attribute.String("http.client_ip", c.IP()),
			),
		)
		defer span.End()

		// Store trace context in Fiber context
		c.SetUserContext(ctx)

		// Add trace ID to response headers
		if span.SpanContext().IsValid() {
			c.Set("X-Trace-ID", span.SpanContext().TraceID().String())
		}

		// Process request
		err := c.Next()

		// Set status code attribute
		statusCode := c.Response().StatusCode()
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(statusCode))

		// Mark span as error if status >= 400
		if statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}
