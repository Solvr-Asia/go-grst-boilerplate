# Logging Best Practices

## 1. Structured Logging with Zap

```go
// ✓ Good: Use structured fields
logger.Info("user registered",
    zap.String("user_id", user.ID),
    zap.String("email", user.Email),
)

// ✓ Good: Include trace context
logger.WithContext(ctx).Info("processing request")

// ✗ Bad: String interpolation
logger.Info(fmt.Sprintf("user %s registered", user.ID))

// ✗ Bad: Logging sensitive data
logger.Info("user login", zap.String("password", password))  // NEVER DO THIS
```

5xx causes are logged server-side and never leaked to clients (see
[codebase-conventions](codebase-conventions.md) and [security](security.md) A09).
