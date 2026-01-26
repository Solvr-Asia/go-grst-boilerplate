# Go-GRST-Boilerplate

A production-ready Go monolithic application boilerplate using **Go Fiber** for REST API, **gRPC** for service-to-service communication, with **Domain-Driven Design (DDD)** and **Clean Architecture**.

## Features

- **Dual Protocol**: REST (Fiber) + gRPC in single application
- **Clean Architecture**: Clear separation of concerns with DDD
- **Database**: PostgreSQL with GORM ORM + OpenTelemetry tracing
- **Caching**: Redis with Redigo connection pooling
- **Message Queue**: RabbitMQ for async processing
- **Logging**: Zap structured logging with trace context
- **Tracing**: OpenTelemetry with Jaeger exporter
- **Configuration**: Viper with .env file support
- **Validation**: go-playground/validator with custom validators
- **Authentication**: JWT with role-based access control (RBAC)
- **Testing**: Testify with mocking support
- **Migrations**: golang-migrate with SQL files support
- **Resilience**: Circuit breaker, retry, timeout with failsafe-go
- **Rate Limiting**: Request throttling with configurable limits
- **Metrics**: Prometheus metrics with `/metrics` endpoint
- **API Documentation**: OpenAPI/Swagger with Scalar UI
- **CI/CD**: GitHub Actions for lint, test, build, and docker

## Tech Stack

| Component | Technology |
|-----------|------------|
| Web Framework | [Go Fiber v2](https://gofiber.io/) |
| gRPC | [google.golang.org/grpc](https://grpc.io/) |
| ORM | [GORM](https://gorm.io/) |
| Database | PostgreSQL |
| Cache | Redis ([Redigo](https://github.com/gomodule/redigo)) |
| Message Queue | RabbitMQ ([amqp091-go](https://github.com/rabbitmq/amqp091-go)) |
| Logger | [Zap](https://github.com/uber-go/zap) |
| Tracing | [OpenTelemetry](https://opentelemetry.io/) |
| Config | [Viper](https://github.com/spf13/viper) |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| Testing | [Testify](https://github.com/stretchr/testify) |
| Auth | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Resilience | [failsafe-go](https://failsafe-go.dev/) |
| Metrics | [Prometheus](https://prometheus.io/) |
| API Docs | [Scalar](https://github.com/yokeTH/gofiber-scalar) |

## Project Structure

```
go-grst-boilerplate/
├── main.go                     # Application entry point
├── cmd/
│   └── migrate/                # Migration CLI tool
│       └── main.go
├── config/
│   └── config.go               # Viper configuration
├── migrations/                 # SQL migration files
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   └── ...
├── database/
│   ├── migrate/                # Migration helper
│   │   └── migrate.go
│   └── seeds/                  # Database seeders
│       └── seeder.go
├── entity/                     # Domain entities
│   ├── user.go
│   └── payslip.go
├── repository/                 # Data access layer
│   └── user_repository/
│       └── repository.go
├── app/                        # Business logic layer
│   └── usecase/
│       └── user/
│           ├── usecase.go
│           └── usecase_test.go
├── handler/                    # Presentation layer
│   ├── grpc/
│   │   └── user/
│   │       ├── user.pb.go      # Generated protobuf
│   │       ├── user_grpc.pb.go # Generated gRPC service
│   │       └── user_protokit.go
│   └── http/
│       └── user/
│           ├── handler.go
│           └── routes.go
├── pkg/                        # Shared utilities
│   ├── database/               # GORM setup
│   ├── redis/                  # Redigo client
│   ├── rabbitmq/               # RabbitMQ client
│   ├── logger/                 # Zap logger
│   ├── telemetry/              # OpenTelemetry setup
│   ├── jwt/                    # JWT service
│   ├── validation/             # Validator with custom rules
│   ├── middleware/             # HTTP/gRPC middleware + rate limiting
│   ├── response/               # Standard responses
│   ├── errors/                 # Error types
│   ├── resilience/             # Circuit breaker, retry, timeout
│   └── metrics/                # Prometheus metrics
├── docs/                       # API documentation
│   └── swagger.go              # OpenAPI spec + Scalar UI
├── .github/workflows/          # CI/CD pipelines
│   ├── ci.yml                  # Lint, test, build
│   └── release.yml             # Release automation
├── docker-compose.yml          # Infrastructure services
├── Dockerfile                  # Multi-stage build
├── Makefile                    # Build commands
├── .golangci.yml               # Linter configuration
├── .goreleaser.yml             # Release configuration
├── LICENSE                     # MIT License
├── .env.example                # Environment template
├── claude.md                   # AI coding guidelines
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make

### 1. Clone and Setup

```bash
git clone https://github.com/yourusername/go-grst-boilerplate.git
cd go-grst-boilerplate

# Copy environment file
cp .env.example .env
```

### 2. Start Infrastructure

```bash
# Start PostgreSQL, Redis, RabbitMQ, and Jaeger
make compose-up
```

### 3. Run Application

```bash
# Run directly
make run

# Or build and run
make build
./bin/server
```

### 4. Test Endpoints

```bash
# Health check
curl http://localhost:3000/health

# Register user
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'

# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Get profile (with token)
curl http://localhost:3000/api/v1/user/profile \
  -H "Authorization: Bearer <your-token>"
```

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
# Application
APP_NAME=go-grst-boilerplate
APP_ENV=development
HTTP_PORT=3000
GRPC_PORT=50051

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=grst_db
DB_TIMEZONE=Asia/Jakarta

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRATION=24

# Telemetry
OTEL_ENABLED=true
OTEL_EXPORTER=jaeger
OTEL_ENDPOINT=localhost:4317

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

## API Endpoints

### Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/register` | No | Register new user |
| POST | `/api/v1/auth/login` | No | Login user |

### User

| Method | Endpoint | Auth | Roles | Description |
|--------|----------|------|-------|-------------|
| GET | `/api/v1/user/profile` | Yes | Any | Get user profile |
| GET | `/api/v1/admin/users` | Yes | admin, superadmin | List all users |
| GET | `/api/v1/employee/payslip` | Yes | employee | Get payslip |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |

## gRPC Services

```protobuf
service UserApi {
    rpc Register(RegisterReq) returns (RegisterRes);
    rpc Login(LoginReq) returns (LoginRes);
    rpc GetProfile(google.protobuf.Empty) returns (UserProfile);
    rpc ListAllUsers(ListUsersReq) returns (ListUsersRes);
    rpc GetMyPayslip(GetPayslipReq) returns (Payslip);
}
```

Connect via gRPC at `localhost:50051`.

## Response Format

### Success Response

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "email": "test@example.com",
        "name": "Test User"
    }
}
```

### Success with Pagination

```json
{
    "success": true,
    "data": [...],
    "meta": {
        "page": 1,
        "size": 10,
        "total": 100,
        "totalPages": 10
    }
}
```

### Error Response

```json
{
    "success": false,
    "error": {
        "code": 40001,
        "message": "validation failed: email is required"
    }
}
```

## Validation Rules

Built-in validators:

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must not be empty | `validate:"required"` |
| `email` | Valid email format | `validate:"email"` |
| `min=N` | Minimum length/value | `validate:"min=8"` |
| `max=N` | Maximum length/value | `validate:"max=100"` |
| `oneof=a b` | Value must be one of | `validate:"oneof=asc desc"` |
| `gte=N` | Greater than or equal | `validate:"gte=1"` |
| `lte=N` | Less than or equal | `validate:"lte=100"` |

Custom validators:

| Tag | Description |
|-----|-------------|
| `phone` | Valid phone number |
| `password` | Min 8 chars, upper, lower, digit |
| `nik` | Indonesian NIK (16 digits) |

## Database Migrations

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations with SQL files.

### Migration Commands

```bash
# Run all pending migrations
make migrate

# Rollback all migrations
make migrate-down

# Rollback last migration only
make migrate-rollback

# Show current migration version
make migrate-status

# Create new migration (creates up and down files)
make migrate-create name=create_orders_table

# Run database seeders
make seed

# Drop all tables and re-run migrations
make fresh

# Drop all, migrate, and seed
make fresh-seed

# Rollback all and re-run migrations
make refresh

# Rollback all, migrate, and seed
make refresh-seed

# Reset database (rollback all)
make reset
```

### Migration File Format

Migration files are stored in `migrations/` directory:

```
migrations/
├── 000001_create_users_table.up.sql    # Creates users table
├── 000001_create_users_table.down.sql  # Drops users table
├── 000002_create_payslips_table.up.sql
├── 000002_create_payslips_table.down.sql
└── ...
```

### Creating New Migrations

```bash
# Create a new migration
make migrate-create name=add_orders_table

# This creates:
# - migrations/000004_add_orders_table.up.sql
# - migrations/000004_add_orders_table.down.sql
```

### Seeding Data

The seeder creates sample data for development:

```bash
# Run all seeders
make seed
```

Default seed data:
- **superadmin@example.com** (password: `SuperAdmin123!`) - Roles: superadmin, admin
- **admin@example.com** (password: `Admin123!`) - Roles: admin
- **employee1@example.com** (password: `Employee123!`) - Roles: employee
- **employee2@example.com** (password: `Employee123!`) - Roles: employee
- **user@example.com** (password: `User123!`) - Roles: user

## Testing

```bash
# Run all tests
make test

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Resilience Patterns

This boilerplate uses [failsafe-go](https://failsafe-go.dev/) for resilience patterns:

### Circuit Breaker

Prevents cascading failures by stopping requests to failing services:

```go
import "go-grst-boilerplate/pkg/resilience"

// Create executor with circuit breaker
executor := resilience.New[*http.Response](
    "payment-service",
    resilience.DefaultConfig(),
    resilience.WithLogger[*http.Response](logger),
)

// Execute with resilience
response, err := executor.Execute(ctx, func(ctx context.Context) (*http.Response, error) {
    return httpClient.Do(request)
})
```

### Resilient HTTP Client

Pre-built HTTP client with circuit breaker, retry, and timeout:

```go
client := resilience.NewHTTPClient("external-api", resilience.DefaultHTTPClientConfig(), logger)

// Automatic retry, circuit breaker, and timeout
resp, err := client.Get(ctx, "https://api.example.com/data")
```

### Configuration

```go
cfg := resilience.Config{
    CBFailureThreshold: 5,           // Failures before opening circuit
    CBSuccessThreshold: 3,           // Successes before closing circuit
    CBDelay:            30 * time.Second, // Wait time in open state
    RetryMaxAttempts:   3,           // Max retry attempts
    RetryDelay:         100 * time.Millisecond,
    RetryMaxDelay:      2 * time.Second,
    Timeout:            10 * time.Second,
}
```

## Rate Limiting

Built-in rate limiting middleware to prevent abuse:

```go
import "go-grst-boilerplate/pkg/middleware"

// Default: 100 requests per minute per IP
app.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimitConfig()))

// Custom configuration
app.Use(middleware.RateLimitMiddleware(middleware.RateLimitConfig{
    Max:      50,
    Duration: 1 * time.Minute,
    KeyGenerator: func(c *fiber.Ctx) string {
        return c.IP()
    },
}))

// User-based rate limiting (after auth)
app.Use(middleware.UserRateLimiter(config))

// API key-based rate limiting
app.Use(middleware.APIKeyRateLimiter(config, "X-API-Key"))
```

## Security (OWASP Top 10)

This boilerplate implements security best practices following the [OWASP Top 10](https://owasp.org/www-project-top-ten/):

| OWASP Risk | Implementation |
|------------|----------------|
| **A01: Broken Access Control** | Role-based access control (RBAC) middleware, resource ownership verification |
| **A02: Cryptographic Failures** | bcrypt password hashing, secure JWT secrets, TLS enforcement |
| **A03: Injection** | Parameterized queries with GORM, input validation with go-playground/validator |
| **A04: Insecure Design** | Rate limiting, account lockout, secure defaults |
| **A05: Security Misconfiguration** | Environment-based config, secure CORS, security headers |
| **A06: Vulnerable Components** | CI/CD with gosec and Trivy scanning, dependabot |
| **A07: Auth Failures** | Strong password policy, JWT expiration, failed login tracking |
| **A08: Data Integrity** | Request validation, file checksum verification |
| **A09: Logging Failures** | Structured security logging with Zap, audit trails |
| **A10: SSRF** | URL whitelist validation, internal IP blocking |

### Security Features

```go
// Role-based access control
app.Get("/admin/users",
    middleware.AuthMiddleware(jwtService),
    middleware.RequireRoles("admin", "superadmin"),
    handler.ListUsers,
)

// Input validation
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,password"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

// Secure password hashing
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Parameterized queries (SQL injection prevention)
db.Where("email = ?", email).First(&user)
```

### Security Scanning

The CI/CD pipeline includes:

```yaml
# .github/workflows/ci.yml
- gosec: Static security analysis
- trivy: Container vulnerability scanning
- golangci-lint: Code quality with security linters
```

Run locally:

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...
```

See [CLAUDE.md](CLAUDE.md) for detailed security guidelines and code examples.

## Prometheus Metrics

Automatic HTTP metrics collection with `/metrics` endpoint:

```go
import "go-grst-boilerplate/pkg/metrics"

// Initialize metrics
m := metrics.Init("myapp")

// Add metrics middleware
app.Use(m.Middleware())

// Expose metrics endpoint
app.Get("/metrics", m.Handler())

// Record custom metrics
m.RecordUserRegistered()
m.RecordDBQuery("select", "users", duration)
m.RecordCacheHit("redis")
m.SetCircuitBreakerState("payment-service", 0) // 0=closed, 1=half-open, 2=open
```

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | Request latency |
| `http_requests_in_flight` | Gauge | Current active requests |
| `db_queries_total` | Counter | Database queries |
| `cache_hits_total` | Counter | Cache hits |
| `circuit_breaker_state` | Gauge | Circuit breaker state |

## API Documentation

Interactive API documentation using Scalar UI:

- **Scalar UI**: http://localhost:3000/docs
- **OpenAPI JSON**: http://localhost:3000/docs/openapi.json

Setup in your application:

```go
import "go-grst-boilerplate/docs"

// Add Swagger documentation
docs.SetupSwagger(app)
```

## Docker

### Pull from GitHub Container Registry

```bash
# Pull latest version
docker pull ghcr.io/yourusername/go-grst-boilerplate:latest

# Pull specific version (semantic versioning)
docker pull ghcr.io/yourusername/go-grst-boilerplate:1.0.0
docker pull ghcr.io/yourusername/go-grst-boilerplate:1.0
docker pull ghcr.io/yourusername/go-grst-boilerplate:1
```

### Build Image Locally

```bash
make docker
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop all services
docker-compose down
```

## Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/) (SemVer):

- **MAJOR** version (X.0.0): Breaking changes
- **MINOR** version (0.X.0): New features (backward compatible)
- **PATCH** version (0.0.X): Bug fixes (backward compatible)

### Creating a Release

```bash
# Create a new release tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Or use make command
make release VERSION=1.0.0
```

### Conventional Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) for automatic changelog generation:

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat:` | New feature | Minor |
| `fix:` | Bug fix | Patch |
| `perf:` | Performance improvement | Patch |
| `refactor:` | Code refactoring | None |
| `docs:` | Documentation | None |
| `test:` | Tests | None |
| `chore:` | Maintenance | None |
| `BREAKING CHANGE:` | Breaking change | Major |

Examples:
```bash
git commit -m "feat: add user authentication"
git commit -m "fix: resolve login timeout issue"
git commit -m "feat!: redesign API response format"  # Breaking change
```

## Observability

### Tracing (Jaeger)

Access Jaeger UI at: http://localhost:16686

### Logging

Logs are structured JSON with trace context:

```json
{
    "level": "info",
    "ts": "2024-01-15T10:30:00.000Z",
    "caller": "handler/user_handler.go:45",
    "msg": "user registered",
    "trace_id": "abc123",
    "span_id": "def456",
    "user_id": "user-123"
}
```

## Make Commands

```bash
# Application
make run              # Run the application
make build            # Build binary
make test             # Run tests
make lint             # Run linter
make clean            # Clean build artifacts

# Database Migrations
make migrate          # Run all pending migrations
make migrate-down     # Rollback all migrations
make migrate-rollback # Rollback last migration
make migrate-status   # Show current version
make migrate-create name=<name>  # Create new migration
make seed             # Run database seeders
make fresh            # Drop all and re-migrate
make fresh-seed       # Drop all, migrate, and seed
make refresh          # Rollback all and re-migrate
make reset            # Rollback all migrations

# Docker
make docker           # Build Docker image
make compose-up       # Start infrastructure
make compose-down     # Stop infrastructure

# Other
make proto            # Generate protobuf (if applicable)
make deps             # Download dependencies
make install-tools    # Install dev tools
```

## Architecture Decisions

### Why Fiber?

- High performance (up to 10x faster than net/http)
- Express.js-like syntax
- Zero memory allocation in hot paths
- Built-in middleware ecosystem

### Why gRPC alongside REST?

- REST for external clients (web, mobile)
- gRPC for internal service-to-service communication
- Shared business logic via Clean Architecture

### Why Redigo over go-redis?

- Lower memory footprint
- Simpler API
- Better connection pool control
- Well-tested in production environments

### Why Zap over other loggers?

- Blazing fast (zero allocation in hot paths)
- Structured logging out of the box
- Easy integration with OpenTelemetry

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests with race detection (`go test -race ./...`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Go Fiber](https://gofiber.io/)
- [GORM](https://gorm.io/)
- [OpenTelemetry](https://opentelemetry.io/)
- [Goroutine Problems Reference](https://github.com/superbolang/golang-goroutines_problem)
