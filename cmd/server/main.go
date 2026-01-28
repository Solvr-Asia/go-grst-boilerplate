package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-grst-boilerplate/config"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	log, err := config.NewLogger(cfg)
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
	otel, err := config.NewTelemetry(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to initialize telemetry", zap.Error(err))
	}
	defer otel.Shutdown(ctx)

	log.Info("OpenTelemetry initialized",
		zap.Bool("enabled", cfg.OTelEnabled),
		zap.String("exporter", cfg.OTelExporterType),
	)

	// Initialize database
	db, err := config.NewDatabase(cfg, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize Redis
	redisClient, err := config.NewRedis(cfg)
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
	rabbitClient, err := config.NewRabbitMQ(cfg, log.Logger)
	if err != nil {
		log.Warn("Failed to connect to RabbitMQ, messaging disabled", zap.Error(err))
	} else {
		defer rabbitClient.Close()
		log.Info("RabbitMQ connection established",
			zap.String("host", cfg.RabbitMQHost),
			zap.Int("port", cfg.RabbitMQPort),
		)
	}

	// Create Fiber app
	app := config.NewFiber(cfg, log.Logger)

	// Bootstrap application (wire layers, routes, health checks)
	result := config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log.Logger,
		Cfg:      cfg,
		Redis:    redisClient,
		RabbitMQ: rabbitClient,
	})

	// Start servers
	errChan := make(chan error, 2)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			errChan <- fmt.Errorf("gRPC listen error: %w", err)
			return
		}
		log.Info("gRPC server listening", zap.Int("port", cfg.GRPCPort))
		if err := result.GRPCServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("gRPC serve error: %w", err)
		}
	}()

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

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result.GRPCServer.GracefulStop()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Error("Fiber shutdown error", zap.Error(err))
		}

		log.Info("Servers stopped gracefully")
	case err := <-errChan:
		log.Fatal("Server error", zap.Error(err))
	}
}
