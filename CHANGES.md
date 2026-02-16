# Prefork and Worker Implementation Summary

## Changes Made

### 1. Server with Prefork Support

#### Configuration Changes
- **File**: `config/config.go`
  - Added `Prefork bool` field to the Config struct
  - Added default value `PREFORK=false` in setDefaults()

- **File**: `config/fiber.go`
  - Added `Prefork: cfg.Prefork` to Fiber configuration
  - Enables Fiber's prefork mode when PREFORK=true is set

- **File**: `cmd/server/main.go`
  - Added prefork status logging on startup

#### What is Prefork?

Prefork is a feature in Fiber that spawns multiple child processes to handle requests. This approach:
- Utilizes multiple CPU cores more efficiently
- Improves performance under high load
- Each child process handles a subset of incoming connections
- Similar to Nginx's worker processes model

#### How to Enable Prefork

Add to your `.env` file:
```env
PREFORK=true
```

Or set environment variable:
```bash
export PREFORK=true
```

⚠️ **Important Notes**:
- Prefork is recommended for production environments only
- Do NOT use with debugging/hot-reload (air) in development
- Each child process will have its own in-memory state
- Use Redis/Database for shared state across processes
- gRPC server runs normally (not affected by prefork)

### 2. Worker Service for RabbitMQ Consumers

#### New Files Created

- **File**: `cmd/worker/main.go`
  - Standalone worker application for consuming RabbitMQ messages
  - Features:
    - Configurable concurrent workers (default: 5)
    - Quality of Service (QoS) with prefetch count (default: 10)
    - Graceful shutdown handling
    - OpenTelemetry integration for distributed tracing
    - Manual message acknowledgment (ACK/NACK)
    - Database and Redis support (optional)
    - Automatic topology setup (exchanges, queues, bindings)

- **File**: `cmd/worker/README.md`
  - Comprehensive documentation for the worker service
  - Configuration guide
  - Usage examples
  - Message handler implementation patterns
  - Monitoring and troubleshooting tips

#### Makefile Updates

- **File**: `Makefile`
  - Added `run-worker`: Run worker in development mode
  - Added `build-worker`: Build worker binary
  - Updated `.PHONY` and help section

## Usage

### Running the Server

```bash
# Without prefork (development)
make run

# With prefork (production)
PREFORK=true make run

# Or build and run
make build
PREFORK=true ./bin/go-grst-boilerplate
```

### Running the Worker

```bash
# Development mode
make run-worker

# Build and run
make build-worker
./bin/go-grst-boilerplate-worker
```

### Running Both Together

```bash
# Terminal 1: Run server
make run

# Terminal 2: Run worker
make run-worker
```

## Configuration

### Environment Variables

```env
# Server Configuration
SERVICE_NAME=go-grst-boilerplate
ENVIRONMENT=production
HTTP_PORT=3000
GRPC_PORT=50051
PREFORK=true  # Enable prefork mode

# RabbitMQ (required for worker)
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/

# Database (optional for worker)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_grst_db

# Redis (optional for worker)
REDIS_HOST=localhost
REDIS_PORT=6379
```

## Worker Customization

To customize the worker for your use case, edit `cmd/worker/main.go`:

### 1. Configure Queue Settings

```go
const (
    DefaultQueue        = "your_queue_name"
    DefaultExchange     = "your_exchange"
    DefaultExchangeType = "topic"  // or "direct", "fanout", "headers"
    DefaultRoutingKey   = "your.routing.key.#"
    ConcurrentWorkers   = 10  // Adjust based on workload
    PrefetchCount       = 20  // Messages per worker
)
```

### 2. Implement Message Handler

```go
func handleMessage(ctx context.Context, msg amqp.Delivery, log *zap.Logger, db interface{}, redis interface{}) error {
    // Parse message
    var payload YourPayloadType
    if err := json.Unmarshal(msg.Body, &payload); err != nil {
        return fmt.Errorf("invalid message: %w", err)
    }

    // Process message
    // - Call business logic
    // - Update database
    // - Send notifications
    // - Cache results
    
    log.Info("Message processed", zap.Any("data", payload))
    return nil
}
```

### 3. Add Multiple Queues

To consume from multiple queues, start multiple consumers:

```go
// Consumer for emails
go rabbitClient.ConsumeWithHandler(ctx, rabbitmq.ConsumeOptions{
    Queue: "email_queue",
    ConsumerTag: "email-consumer",
    AutoAck: false,
}, handleEmailMessage)

// Consumer for notifications
go rabbitClient.ConsumeWithHandler(ctx, rabbitmq.ConsumeOptions{
    Queue: "notification_queue",
    ConsumerTag: "notification-consumer",
    AutoAck: false,
}, handleNotificationMessage)
```

## Publishing Messages from Server

To send messages from your server application for the worker to consume:

```go
import (
    "context"
    "go-grst-boilerplate/pkg/rabbitmq"
)

// In your service layer
func (s *MyService) CreateTask(ctx context.Context, data TaskData) error {
    // Publish to RabbitMQ
    err := s.rabbitMQ.Publish(ctx, rabbitmq.PublishOptions{
        Exchange:   "default_exchange",
        RoutingKey: "default.task.created",
    }, data)
    
    if err != nil {
        return fmt.Errorf("failed to queue task: %w", err)
    }
    
    return nil
}
```

## Production Deployment

### Docker / Kubernetes

You would typically deploy server and worker as separate services:

```yaml
# Server deployment
- name: api-server
  image: your-app:latest
  command: ["/app/server"]
  env:
    - name: PREFORK
      value: "true"

# Worker deployment  
- name: worker
  image: your-app:latest
  command: ["/app/worker"]
  replicas: 3  # Scale workers independently
```

### Systemd Services

```ini
# /etc/systemd/system/go-grst-server.service
[Service]
Environment="PREFORK=true"
ExecStart=/usr/local/bin/go-grst-boilerplate

# /etc/systemd/system/go-grst-worker.service
[Service]
ExecStart=/usr/local/bin/go-grst-boilerplate-worker
```

## Monitoring

Both server and worker include:
- Structured logging with zap
- OpenTelemetry distributed tracing
- Automatic request/message correlation
- Error tracking and metrics

Monitor worker health by checking:
- Log output for processing success/errors
- RabbitMQ queue depth (should stay low)
- Message processing latency (via OpenTelemetry)
- Worker process health (CPU, memory usage)

## Next Steps

1. ✅ Customize queue names and routing keys in `cmd/worker/main.go`
2. ✅ Implement your message handler logic
3. ✅ Add message publishing to your server handlers/services
4. ✅ Set up RabbitMQ exchanges and queues (done automatically by worker)
5. ✅ Test message flow end-to-end
6. ✅ Configure monitoring and alerting
7. ✅ Deploy to production with prefork enabled for server
