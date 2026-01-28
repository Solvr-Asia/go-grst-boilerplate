package database

import (
	"context"
	"fmt"
	"time"

	"go-grst-boilerplate/entity"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

type zapGormLogger struct {
	logger        *zap.Logger
	slowThreshold time.Duration
}

func newZapGormLogger(zapLogger *zap.Logger) *zapGormLogger {
	return &zapGormLogger{
		logger:        zapLogger,
		slowThreshold: 200 * time.Millisecond,
	}
}

func (l *zapGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *zapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Sugar().Infof(msg, data...)
}

func (l *zapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Sugar().Warnf(msg, data...)
}

func (l *zapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Sugar().Errorf(msg, data...)
}

func (l *zapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	if err != nil {
		l.logger.Error("gorm query error", append(fields, zap.Error(err))...)
		return
	}

	if elapsed > l.slowThreshold {
		l.logger.Warn("gorm slow query", fields...)
		return
	}

	l.logger.Debug("gorm query", fields...)
}

// Config holds database connection parameters.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

func New(cfg Config, zapLogger *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
		cfg.Timezone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newZapGormLogger(zapLogger),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Add OpenTelemetry tracing plugin
	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("failed to add tracing plugin: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	zapLogger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
	)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Payslip{},
	)
}

// WithContext returns a new DB with context for tracing
func WithContext(db *gorm.DB, ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}
