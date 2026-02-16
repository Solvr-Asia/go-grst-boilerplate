package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-grst-boilerplate/config"
	"go-grst-boilerplate/pkg/rabbitmq"

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
	ConsumerTag      = "worker-consumer"
	PrefetchCount    = 10
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
	defer log.Sync()

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
	defer otel.Shutdown(ctx)

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
		defer redisClient.Close()
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
	defer rabbitClient.Close()
	log.Info("RabbitMQ connection established",
		zap.String("host", cfg.RabbitMQHost),
		zap.Int("port", cfg.RabbitMQPort),
	)

	// Set up RabbitMQ topology
	if err := setupTopology(rabbitClient, log.Logger); err != nil {
		log.Fatal("Failed to setup RabbitMQ topology", zap.Error(err))
	}

	// Set QoS
	if err := rabbitClient.SetQoS(PrefetchCount, 0, false); err != nil {
		log.Fatal("Failed to set QoS", zap.Error(err))
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumers
	errChan := make(chan error, ConcurrentWorkers)

	for i := 0; i < ConcurrentWorkers; i++ {
		workerID := i + 1
		go func(id int) {
			consumerTag := fmt.Sprintf("%s-%d", ConsumerTag, id)
			log.Info("Starting consumer",
				zap.Int("worker_id", id),
				zap.String("consumer_tag", consumerTag),
				zap.String("queue", DefaultQueue),
			)

			err := rabbitClient.ConsumeWithHandler(ctx, rabbitmq.ConsumeOptions{
				Queue:       DefaultQueue,
				ConsumerTag: consumerTag,
				AutoAck:     false,
				Exclusive:   false,
				NoLocal:     false,
				NoWait:      false,
			}, func(handlerCtx context.Context, msg amqp.Delivery) error {
				return handleMessage(handlerCtx, msg, log.Logger, db, redisClient)
			})

			if err != nil {
				errChan <- fmt.Errorf("consumer %d error: %w", id, err)
			}
		}(workerID)
	}

	log.Info("All consumers started",
		zap.Int("concurrent_workers", ConcurrentWorkers),
		zap.Int("prefetch_count", PrefetchCount),
	)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Shutting down worker...")
		cancel()

		// Give consumers time to finish processing current messages
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		<-shutdownCtx.Done()
		log.Info("Worker stopped gracefully")
	case err := <-errChan:
		log.Fatal("Consumer error", zap.Error(err))
	}
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

	// Parse message body
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		log.Error("Failed to unmarshal message",
			zap.Error(err),
			zap.String("body", string(msg.Body)),
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
		zap.Any("payload", payload),
	)

	// Simulate processing time (remove this in production)
	time.Sleep(100 * time.Millisecond)

	return nil
}
