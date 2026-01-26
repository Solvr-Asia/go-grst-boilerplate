package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("pkg/rabbitmq")

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger
	config  Config
	mu      sync.RWMutex
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

type PublishOptions struct {
	Exchange    string
	RoutingKey  string
	Mandatory   bool
	Immediate   bool
	ContentType string
	Headers     map[string]interface{}
}

type ConsumeOptions struct {
	Queue       string
	ConsumerTag string
	AutoAck     bool
	Exclusive   bool
	NoLocal     bool
	NoWait      bool
	Args        amqp.Table
}

type Message struct {
	Body    []byte
	Headers map[string]interface{}
}

func New(cfg Config, logger *zap.Logger) (*Client, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &Client{
		conn:    conn,
		channel: channel,
		logger:  logger,
		config:  cfg,
	}

	return client, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Channel() *amqp.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channel
}

// DeclareExchange declares an exchange
func (c *Client) DeclareExchange(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return c.channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

// DeclareQueue declares a queue
func (c *Client) DeclareQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return c.channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

// BindQueue binds a queue to an exchange
func (c *Client) BindQueue(queueName, routingKey, exchangeName string, noWait bool, args amqp.Table) error {
	return c.channel.QueueBind(queueName, routingKey, exchangeName, noWait, args)
}

// Publish publishes a message with tracing
func (c *Client) Publish(ctx context.Context, opts PublishOptions, message interface{}) error {
	ctx, span := tracer.Start(ctx, "rabbitmq.Publish",
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination", opts.Exchange),
			attribute.String("messaging.rabbitmq.routing_key", opts.RoutingKey),
		))
	defer span.End()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	headers := make(amqp.Table)
	for k, v := range opts.Headers {
		headers[k] = v
	}

	// Inject trace context into headers
	propagator := otel.GetTextMapPropagator()
	carrier := make(propagation.MapCarrier)
	propagator.Inject(ctx, carrier)
	for k, v := range carrier {
		headers[k] = v
	}

	contentType := opts.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	publishing := amqp.Publishing{
		ContentType:  contentType,
		Body:         body,
		Headers:      headers,
		Timestamp:    time.Now(),
		DeliveryMode: amqp.Persistent,
	}

	return c.channel.PublishWithContext(ctx, opts.Exchange, opts.RoutingKey, opts.Mandatory, opts.Immediate, publishing)
}

// Consume starts consuming messages from a queue
func (c *Client) Consume(opts ConsumeOptions) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		opts.Queue,
		opts.ConsumerTag,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
}

// ConsumeWithHandler consumes messages and processes them with a handler
func (c *Client) ConsumeWithHandler(ctx context.Context, opts ConsumeOptions, handler func(ctx context.Context, msg amqp.Delivery) error) error {
	msgs, err := c.Consume(opts)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				// Extract trace context from headers
				carrier := make(propagation.MapCarrier)
				for k, v := range msg.Headers {
					if s, ok := v.(string); ok {
						carrier[k] = s
					}
				}
				propagator := otel.GetTextMapPropagator()
				msgCtx := propagator.Extract(ctx, carrier)

				msgCtx, span := tracer.Start(msgCtx, "rabbitmq.Consume",
					trace.WithAttributes(
						attribute.String("messaging.system", "rabbitmq"),
						attribute.String("messaging.destination", opts.Queue),
					))

				if err := handler(msgCtx, msg); err != nil {
					c.logger.Error("Failed to process message",
						zap.Error(err),
						zap.String("queue", opts.Queue),
					)
					span.RecordError(err)
					if !opts.AutoAck {
						if nackErr := msg.Nack(false, true); nackErr != nil {
							c.logger.Error("Failed to nack message",
								zap.Error(nackErr),
								zap.String("queue", opts.Queue),
							)
						}
					}
				} else {
					if !opts.AutoAck {
						if ackErr := msg.Ack(false); ackErr != nil {
							c.logger.Error("Failed to ack message",
								zap.Error(ackErr),
								zap.String("queue", opts.Queue),
							)
						}
					}
				}
				span.End()
			}
		}
	}()

	return nil
}

// PublishJSON is a convenience method for publishing JSON messages
func (c *Client) PublishJSON(ctx context.Context, exchange, routingKey string, message interface{}) error {
	return c.Publish(ctx, PublishOptions{
		Exchange:    exchange,
		RoutingKey:  routingKey,
		ContentType: "application/json",
	}, message)
}

// SetQoS sets quality of service
func (c *Client) SetQoS(prefetchCount, prefetchSize int, global bool) error {
	return c.channel.Qos(prefetchCount, prefetchSize, global)
}
