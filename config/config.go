package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName string `mapstructure:"SERVICE_NAME"`
	Environment string `mapstructure:"ENVIRONMENT"`

	// HTTP Server
	HTTPPort int `mapstructure:"HTTP_PORT"`

	// gRPC Server
	GRPCPort int `mapstructure:"GRPC_PORT"`

	// Database
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     int    `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBTimezone string `mapstructure:"DB_TIMEZONE"`
	DBSSLMode  string `mapstructure:"DB_SSL_MODE"`

	// Redis
	RedisHost        string `mapstructure:"REDIS_HOST"`
	RedisPort        int    `mapstructure:"REDIS_PORT"`
	RedisPassword    string `mapstructure:"REDIS_PASSWORD"`
	RedisDB          int    `mapstructure:"REDIS_DB"`
	RedisMaxIdle     int    `mapstructure:"REDIS_MAX_IDLE"`
	RedisMaxActive   int    `mapstructure:"REDIS_MAX_ACTIVE"`
	RedisIdleTimeout int    `mapstructure:"REDIS_IDLE_TIMEOUT"`

	// RabbitMQ
	RabbitMQHost     string `mapstructure:"RABBITMQ_HOST"`
	RabbitMQPort     int    `mapstructure:"RABBITMQ_PORT"`
	RabbitMQUser     string `mapstructure:"RABBITMQ_USER"`
	RabbitMQPassword string `mapstructure:"RABBITMQ_PASSWORD"`
	RabbitMQVHost    string `mapstructure:"RABBITMQ_VHOST"`

	// JWT
	JWTSecret     string `mapstructure:"JWT_SECRET"`
	JWTExpiration int    `mapstructure:"JWT_EXPIRATION"`

	// CORS
	CORSOrigins string `mapstructure:"CORS_ORIGINS"`

	// OpenTelemetry
	OTelEnabled      bool   `mapstructure:"OTEL_ENABLED"`
	OTelEndpoint     string `mapstructure:"OTEL_ENDPOINT"`
	OTelServiceName  string `mapstructure:"OTEL_SERVICE_NAME"`
	OTelExporterType string `mapstructure:"OTEL_EXPORTER_TYPE"`

	// Logger
	LogLevel  string `mapstructure:"LOG_LEVEL"`
	LogFormat string `mapstructure:"LOG_FORMAT"`
}

func New() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from .env file
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	// Read config file (ignore error if not found)
	_ = v.ReadInConfig() //nolint:errcheck

	// Read from environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Service
	v.SetDefault("SERVICE_NAME", "go-grst-boilerplate")
	v.SetDefault("ENVIRONMENT", "development")

	// HTTP/gRPC
	v.SetDefault("HTTP_PORT", 3000)
	v.SetDefault("GRPC_PORT", 50051)

	// Database
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "go_grst_db")
	v.SetDefault("DB_TIMEZONE", "Asia/Jakarta")
	v.SetDefault("DB_SSL_MODE", "disable")

	// Redis
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_MAX_IDLE", 10)
	v.SetDefault("REDIS_MAX_ACTIVE", 100)
	v.SetDefault("REDIS_IDLE_TIMEOUT", 240)

	// RabbitMQ
	v.SetDefault("RABBITMQ_HOST", "localhost")
	v.SetDefault("RABBITMQ_PORT", 5672)
	v.SetDefault("RABBITMQ_USER", "guest")
	v.SetDefault("RABBITMQ_PASSWORD", "guest")
	v.SetDefault("RABBITMQ_VHOST", "/")

	// JWT
	v.SetDefault("JWT_SECRET", "your-secret-key-change-in-production")
	v.SetDefault("JWT_EXPIRATION", 24)

	// CORS
	v.SetDefault("CORS_ORIGINS", "*")

	// OpenTelemetry
	v.SetDefault("OTEL_ENABLED", true)
	v.SetDefault("OTEL_ENDPOINT", "localhost:4317")
	v.SetDefault("OTEL_SERVICE_NAME", "go-grst-boilerplate")
	v.SetDefault("OTEL_EXPORTER_TYPE", "noop")

	// Logger
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
}

// MustNew returns config or panics
func MustNew() *Config {
	cfg, err := New()
	if err != nil {
		panic(err)
	}
	return cfg
}
