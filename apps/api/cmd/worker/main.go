// Command worker consumes messages from RabbitMQ.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"veemon/config"
	"veemon/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	// Queue names - adjust these to match your queues
	DefaultQueue = "default_queue"

	// Exchange configuration
	DefaultExchange     = "default_exchange"
	DefaultExchangeType = "topic"

	// Routing keys
	DefaultRoutingKey = "default.#"

	// Consumer configuration
	ConsumerTag       = "worker-consumer"
	PrefetchCount     = 10
	ConcurrentWorkers = 5
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
	defer func() { _ = log.Sync() }()

	log.Info("Starting worker",
		zap.String("service", cfg.ServiceName+"-worker"),
		zap.String("environment", cfg.Environment),
	)

	// Initialize OpenTelemetry
	ctx := context.Background()
	otel, err := config.NewTelemetry(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to initialize telemetry", zap.Error(err))
	}
	defer func() { _ = otel.Shutdown(ctx) }()

	log.Info("OpenTelemetry initialized",
		zap.Bool("enabled", cfg.OTelEnabled),
		zap.String("exporter", cfg.OTelExporterType),
	)

	// Initialize database (optional - only if worker needs database access)
	db, err := config.NewDatabase(cfg, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	log.Info("Database connection established")

	// Initialize Redis (optional - only if worker needs caching)
	redisClient, err := config.NewRedis(cfg)
	if err != nil {
		log.Warn("Failed to connect to Redis, caching disabled", zap.Error(err))
	} else {
		defer func() { _ = redisClient.Close() }()
		log.Info("Redis connection established",
			zap.String("host", cfg.RedisHost),
			zap.Int("port", cfg.RedisPort),
		)
	}

	// Initialize RabbitMQ
	rabbitClient, err := config.NewRabbitMQ(cfg, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer func() { _ = rabbitClient.Close() }()
	log.Info("RabbitMQ connection established",
		zap.String("host", cfg.RabbitMQHost),
		zap.Int("port", cfg.RabbitMQPort),
	)

	// Set up RabbitMQ topology
	if err := setupTopology(rabbitClient, log.Logger); err != nil {
		log.Fatal("Failed to setup RabbitMQ topology", zap.Error(err))
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumers. Each consumer runs on its own channel, sets its own QoS,
	// and self-heals across connection/channel drops.
	for i := 0; i < ConcurrentWorkers; i++ {
		workerID := i + 1
		consumerTag := fmt.Sprintf("%s-%d", ConsumerTag, workerID)
		log.Info("Starting consumer",
			zap.Int("worker_id", workerID),
			zap.String("consumer_tag", consumerTag),
			zap.String("queue", DefaultQueue),
		)

		if err := rabbitClient.ConsumeWithHandler(ctx, rabbitmq.ConsumeOptions{
			Queue:         DefaultQueue,
			ConsumerTag:   consumerTag,
			AutoAck:       false,
			PrefetchCount: PrefetchCount,
		}, func(handlerCtx context.Context, msg amqp.Delivery) error {
			return handleMessage(handlerCtx, msg, log.Logger, db, redisClient)
		}); err != nil {
			log.Fatal("Failed to start consumer", zap.Int("worker_id", workerID), zap.Error(err))
		}
	}

	log.Info("All consumers started",
		zap.Int("concurrent_workers", ConcurrentWorkers),
		zap.Int("prefetch_count", PrefetchCount),
	)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down worker...")

	// Stop consumers accepting new work, then wait (bounded) for in-flight
	// messages to finish before closing the connection.
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	rabbitClient.WaitConsumers(shutdownCtx)

	if shutdownCtx.Err() != nil {
		log.Warn("Shutdown timeout reached before all consumers drained")
	}
	log.Info("Worker stopped gracefully")
}

// setupTopology declares exchanges, queues, and bindings
func setupTopology(client *rabbitmq.Client, log *zap.Logger) error {
	// Declare exchange
	if err := client.DeclareExchange(
		DefaultExchange,
		DefaultExchangeType,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // args
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Info("Exchange declared",
		zap.String("exchange", DefaultExchange),
		zap.String("type", DefaultExchangeType),
	)

	// Declare queue
	queue, err := client.DeclareQueue(
		DefaultQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Info("Queue declared",
		zap.String("queue", queue.Name),
		zap.Int("messages", queue.Messages),
		zap.Int("consumers", queue.Consumers),
	)

	// Bind queue to exchange
	if err := client.BindQueue(
		DefaultQueue,
		DefaultRoutingKey,
		DefaultExchange,
		false, // no-wait
		nil,   // args
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Info("Queue bound to exchange",
		zap.String("queue", DefaultQueue),
		zap.String("exchange", DefaultExchange),
		zap.String("routing_key", DefaultRoutingKey),
	)

	return nil
}

// handleMessage processes a single message
func handleMessage(ctx context.Context, msg amqp.Delivery, log *zap.Logger, db interface{}, redis interface{}) error {
	log.Info("Processing message",
		zap.String("routing_key", msg.RoutingKey),
		zap.String("content_type", msg.ContentType),
		zap.Int("body_size", len(msg.Body)),
	)

	// Parse message body. Do NOT log the raw body or decoded payload — messages
	// may contain PII or secrets. Log only non-sensitive metadata.
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		log.Error("Failed to unmarshal message",
			zap.Error(err),
			zap.String("routing_key", msg.RoutingKey),
			zap.Int("body_size", len(msg.Body)),
		)
		return fmt.Errorf("invalid message format: %w", err)
	}

	// TODO: Implement your business logic here
	// Example:
	// - Process the message based on routing key
	// - Update database records
	// - Call external APIs
	// - Send notifications
	// - Cache results in Redis

	log.Info("Message processed successfully",
		zap.String("routing_key", msg.RoutingKey),
		zap.Int("fields", len(payload)),
	)

	// Simulate processing time (remove this in production)
	time.Sleep(100 * time.Millisecond)

	return nil
}
