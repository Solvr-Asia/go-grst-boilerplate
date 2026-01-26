# Claude Code Guidelines for Go-GRST-Boilerplate

This document contains coding standards, best practices, and rules for AI assistants working on this codebase.

## Project Overview

This is a Go monolithic application using:
- **Go Fiber** for REST API
- **gRPC** for service-to-service communication
- **Domain-Driven Design (DDD)** with Clean Architecture
- **GORM** for database ORM (PostgreSQL)
- **Redigo** for Redis caching
- **RabbitMQ** for message queuing
- **Zap** for structured logging
- **OpenTelemetry** for distributed tracing
- **Viper** for configuration management
- **go-playground/validator** for validation
- **failsafe-go** for resilience (circuit breaker, retry, timeout)
- **Prometheus** for metrics and monitoring
- **Scalar** for API documentation (OpenAPI/Swagger)

---

## Architecture Layers

```
handler/        → Presentation layer (HTTP/gRPC handlers)
app/usecase/    → Business logic layer
repository/     → Data access layer
entity/         → Domain entities
pkg/            → Shared utilities and infrastructure
config/         → Configuration management
```

**Data Flow:**
```
Request → Handler → UseCase → Repository → Database
                ↓
            Response
```

---

## Coding Standards

### 1. File Organization

- One file per struct/interface when the file exceeds 200 lines
- Group related functions together
- Keep files under 500 lines when possible
- Use meaningful file names that reflect their content

### 2. Naming Conventions

```go
// Packages: lowercase, no underscores
package userrepository  // ✗ Bad
package user_repository // ✓ Good (exception for clarity)
package userrepo        // ✓ Good

// Interfaces: describe behavior, often end with -er
type Reader interface{}
type UserRepository interface{}

// Structs: noun, PascalCase
type User struct{}
type UserService struct{}

// Functions: verb + noun, camelCase for private, PascalCase for public
func GetUser() {}      // ✓ Public
func getUserByID() {}  // ✓ Private

// Constants: PascalCase for exported, camelCase for unexported
const MaxRetries = 3
const defaultTimeout = 30

// Variables: camelCase
var userCount int
var ErrNotFound = errors.New("not found") // Exported errors start with Err
```

### 3. Error Handling

```go
// ✓ Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ✓ Good: Define domain errors
var (
    ErrNotFound    = errors.New("resource not found")
    ErrEmailExists = errors.New("email already exists")
)

// ✓ Good: Check specific errors
if errors.Is(err, gorm.ErrRecordNotFound) {
    return ErrNotFound
}

// ✗ Bad: Ignoring errors
result, _ := doSomething()

// ✗ Bad: Generic error messages
if err != nil {
    return errors.New("error occurred")
}
```

### 4. Context Usage

```go
// ✓ Good: Always pass context as first parameter
func (uc *UseCase) GetUser(ctx context.Context, id string) (*User, error)

// ✓ Good: Use context for cancellation and timeouts
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

// ✓ Good: Check context cancellation in long operations
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // continue processing
}
```

### 5. Interface Design

```go
// ✓ Good: Small, focused interfaces
type UserReader interface {
    FindByID(ctx context.Context, id string) (*User, error)
}

type UserWriter interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
}

// ✓ Good: Compose interfaces
type UserRepository interface {
    UserReader
    UserWriter
}

// ✗ Bad: Large, monolithic interfaces
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindAll(ctx context.Context) ([]User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    // ... 20 more methods
}
```

---

## Goroutine Best Practices

Reference: https://github.com/superbolang/golang-goroutines_problem

### 1. Goroutine Leaks (CRITICAL)

Goroutines that never terminate consume resources indefinitely.

```go
// ✗ Bad: Goroutine leak - no way to stop
func startWorker() {
    go func() {
        for {
            doWork()
        }
    }()
}

// ✓ Good: Use context for cancellation
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // Clean exit
            default:
                doWork()
            }
        }
    }()
}

// ✓ Good: Use done channel
func startWorker(done <-chan struct{}) {
    go func() {
        for {
            select {
            case <-done:
                return
            default:
                doWork()
            }
        }
    }()
}
```

### 2. Race Conditions (CRITICAL)

Multiple goroutines accessing shared memory without synchronization.

```go
// ✗ Bad: Race condition
var counter int

func increment() {
    go func() { counter++ }()
    go func() { counter++ }()
}

// ✓ Good: Use mutex
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}

// ✓ Good: Use atomic operations
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

// ✓ Good: Use channels for communication
func increment(counterCh chan<- int) {
    counterCh <- 1
}
```

**Always run tests with race detection:**
```bash
go test -race ./...
```

### 3. Deadlocks

Goroutines blocked indefinitely, waiting on resources held by others.

```go
// ✗ Bad: Potential deadlock with unbuffered channel
func process() {
    ch := make(chan int)
    ch <- 1  // Blocks forever - no receiver
    <-ch
}

// ✓ Good: Use buffered channel or goroutine
func process() {
    ch := make(chan int, 1)  // Buffered
    ch <- 1
    <-ch
}

// ✓ Good: Use select with timeout
func process() {
    ch := make(chan int)
    select {
    case ch <- 1:
        // sent
    case <-time.After(5 * time.Second):
        // timeout
    }
}
```

### 4. Resource Exhaustion

Creating excessive goroutines that exhaust system resources.

```go
// ✗ Bad: Unbounded goroutine creation
func processItems(items []Item) {
    for _, item := range items {
        go process(item)  // Could create millions of goroutines
    }
}

// ✓ Good: Worker pool pattern
func processItems(ctx context.Context, items []Item) {
    const numWorkers = 10
    jobs := make(chan Item, len(items))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case item, ok := <-jobs:
                    if !ok {
                        return
                    }
                    process(item)
                }
            }
        }()
    }

    // Send jobs
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    wg.Wait()
}

// ✓ Good: Use semaphore pattern
func processItems(items []Item) {
    sem := make(chan struct{}, 10)  // Limit to 10 concurrent
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        sem <- struct{}{}  // Acquire
        go func(item Item) {
            defer wg.Done()
            defer func() { <-sem }()  // Release
            process(item)
        }(item)
    }

    wg.Wait()
}
```

### 5. Channel Blocking

Operations on channels causing unexpected blocking.

```go
// ✗ Bad: Blocking forever
func fetch(url string) string {
    ch := make(chan string)
    go func() {
        // If this panics, main goroutine blocks forever
        ch <- httpGet(url)
    }()
    return <-ch
}

// ✓ Good: Select with timeout
func fetch(ctx context.Context, url string) (string, error) {
    ch := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := httpGet(url)
        if err != nil {
            errCh <- err
            return
        }
        ch <- result
    }()

    select {
    case result := <-ch:
        return result, nil
    case err := <-errCh:
        return "", err
    case <-ctx.Done():
        return "", ctx.Err()
    case <-time.After(30 * time.Second):
        return "", errors.New("request timeout")
    }
}
```

### 6. Unhandled Panics

Panic in a goroutine crashes only that goroutine, leaving application in inconsistent state.

```go
// ✗ Bad: Panic crashes silently
func processAsync(data string) {
    go func() {
        // If this panics, main app continues but this goroutine dies
        process(data)
    }()
}

// ✓ Good: Recover from panics
func processAsync(data string) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Recovered from panic: %v\nStack: %s", r, debug.Stack())
                // Report to monitoring system
            }
        }()
        process(data)
    }()
}

// ✓ Good: Use error channel for goroutine errors
func processAsync(ctx context.Context, data string) <-chan error {
    errCh := make(chan error, 1)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                errCh <- fmt.Errorf("panic: %v", r)
            }
        }()
        if err := process(data); err != nil {
            errCh <- err
            return
        }
        errCh <- nil
    }()
    return errCh
}
```

### 7. Goroutine Monitoring

```go
// Monitor goroutine count in production
import "runtime"

func monitorGoroutines() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        count := runtime.NumGoroutine()
        log.Printf("Active goroutines: %d", count)
        if count > 10000 {
            log.Warn("High goroutine count detected!")
        }
    }
}
```

---

## Database Best Practices

### 1. GORM Usage

```go
// ✓ Good: Use context with timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

var user User
if err := db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
    return err
}

// ✓ Good: Use transactions for multiple operations
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    return nil
})

// ✓ Good: Select only needed fields
db.Select("id", "name", "email").Find(&users)

// ✗ Bad: N+1 query problem
for _, user := range users {
    db.Find(&user.Orders)  // Executes N queries
}

// ✓ Good: Preload associations
db.Preload("Orders").Find(&users)
```

### 2. Connection Pool

```go
// Configure connection pool
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

---

## Redis Best Practices

### 1. Connection Pool with Redigo

```go
// ✓ Good: Use connection pool
pool := &redis.Pool{
    MaxIdle:     10,
    MaxActive:   100,
    IdleTimeout: 240 * time.Second,
    Wait:        true,
    Dial: func() (redis.Conn, error) {
        return redis.Dial("tcp", addr)
    },
}

// ✓ Good: Always close connection
conn := pool.Get()
defer conn.Close()

// ✓ Good: Check connection errors
if err := conn.Err(); err != nil {
    return err
}
```

---

## Testing Guidelines

### 1. Unit Tests

```go
// Use table-driven tests
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"invalid email", "not-an-email", true},
        {"empty email", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. Mock Dependencies

```go
// Use interfaces for dependencies
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
}

// Create mock for testing
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}
```

### 3. Run Tests with Race Detection

```bash
# Always run with race detection
go test -race -v ./...

# With coverage
go test -race -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Logging Best Practices

### 1. Structured Logging with Zap

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

---

## Security Guidelines

### 1. Input Validation

```go
// ✓ Good: Always validate input
type RegisterRequest struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8,max=128"`
    Name     string `validate:"required,min=2,max=100"`
}

if err := validation.Validate(req); err != nil {
    return err
}
```

### 2. SQL Injection Prevention

```go
// ✓ Good: Use parameterized queries
db.Where("email = ?", email).First(&user)

// ✗ Bad: String concatenation
db.Where("email = '" + email + "'").First(&user)
```

### 3. Password Handling

```go
// ✓ Good: Use bcrypt
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// ✓ Good: Compare securely
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
```

---

## API Response Standards

### Success Response

```json
{
    "success": true,
    "data": { ... },
    "meta": {
        "page": 1,
        "size": 10,
        "total": 100
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

---

## Common Commands

```bash
# Run application
make run

# Run tests
make test

# Run tests with race detection
go test -race ./...

# Build application
make build

# Docker compose
make compose-up
make compose-down

# Generate proto (if using protobuf)
make proto
```

---

## Code Review Checklist

- [ ] All errors are handled properly
- [ ] Context is passed to all I/O operations
- [ ] No goroutine leaks (all goroutines can terminate)
- [ ] Race conditions checked (`go test -race`)
- [ ] Sensitive data is not logged
- [ ] Input validation is present
- [ ] Tests are included for new functionality
- [ ] No hardcoded credentials or secrets
- [ ] Database queries use parameterized inputs
- [ ] Resources are properly closed (defer)
