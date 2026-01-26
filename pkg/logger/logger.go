package logger

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

type Logger struct {
	*zap.Logger
}

type Config struct {
	Level       string
	Format      string // "json" or "console"
	Environment string
	ServiceName string
}

func New(cfg Config) (*Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.String("service", cfg.ServiceName),
			zap.String("environment", cfg.Environment),
		),
	)

	globalLogger = logger

	return &Logger{Logger: logger}, nil
}

func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return l.Logger.With(
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return l.Logger
}

func (l *Logger) WithRequestID(requestID string) *zap.Logger {
	return l.Logger.With(zap.String("request_id", requestID))
}

func (l *Logger) WithFields(fields ...zap.Field) *zap.Logger {
	return l.Logger.With(fields...)
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Global logger functions
func L() *zap.Logger {
	if globalLogger == nil {
		var err error
		globalLogger, err = zap.NewProduction()
		if err != nil {
			// Fallback to no-op logger to prevent nil pointer panic
			globalLogger = zap.NewNop()
		}
	}
	return globalLogger
}

func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}

func WithContext(ctx context.Context) *zap.Logger {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return L().With(
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return L()
}
