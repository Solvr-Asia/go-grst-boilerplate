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
cmd/server/     → Application entry point
config/         → Configuration, bootstrap, and infrastructure init
handler/        → Presentation layer (HTTP/gRPC handlers)
app/usecase/    → Business logic layer
repository/     → Data access layer
entity/         → Domain entities
pkg/            → Shared utilities and infrastructure
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

## Security Guidelines (OWASP Top 10)

This project follows OWASP Top 10 security guidelines. Reference: https://owasp.org/www-project-top-ten/

### A01:2021 - Broken Access Control

```go
// ✓ Good: Role-based access control middleware
func RequireRoles(roles ...string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authCtx, ok := middleware.GetAuthContext(c)
        if !ok {
            return errors.Unauthorized("authentication required")
        }
        for _, role := range roles {
            if authCtx.HasRole(role) {
                return c.Next()
            }
        }
        return errors.Forbidden("insufficient permissions")
    }
}

// ✓ Good: Verify resource ownership
func (uc *UserUseCase) GetPayslip(ctx context.Context, userID string, year, month int) (*entity.Payslip, error) {
    // User can only access their own payslip
    authUser := ctx.Value("auth").(*middleware.AuthContext)
    if authUser.UserID != userID && !authUser.HasRole("admin") {
        return nil, errors.Forbidden("cannot access other user's payslip")
    }
    return uc.repo.FindPayslip(ctx, userID, year, month)
}
```

### A02:2021 - Cryptographic Failures

```go
// ✓ Good: Use bcrypt for passwords
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// ✓ Good: Secure password comparison (constant-time)
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))

// ✓ Good: Use strong JWT secret (min 256 bits)
// In .env: JWT_SECRET=<random-32-byte-hex-string>

// ✗ Bad: Weak or hardcoded secrets
const jwtSecret = "secret123"  // NEVER DO THIS
```

### A03:2021 - Injection

```go
// ✓ Good: Use parameterized queries (GORM)
db.Where("email = ?", email).First(&user)
db.Where("id IN ?", ids).Find(&users)

// ✗ Bad: String concatenation (SQL injection)
db.Where("email = '" + email + "'").First(&user)
db.Raw("SELECT * FROM users WHERE email = '" + email + "'")

// ✓ Good: Command injection prevention
// Avoid os/exec with user input. If necessary, whitelist allowed values
allowedCommands := map[string]bool{"status": true, "version": true}
if !allowedCommands[userInput] {
    return errors.BadRequest("invalid command")
}
```

### A04:2021 - Insecure Design

```go
// ✓ Good: Rate limiting to prevent abuse
app.Use(middleware.RateLimitMiddleware(middleware.RateLimitConfig{
    Max:      100,
    Duration: 1 * time.Minute,
}))

// ✓ Good: Implement account lockout
type LoginAttempt struct {
    FailedAttempts int
    LockedUntil    time.Time
}

// ✓ Good: Use secure defaults
cfg := Config{
    JWTExpiration: 24 * time.Hour,  // Not indefinite
    MaxLoginAttempts: 5,
    LockoutDuration: 15 * time.Minute,
}
```

### A05:2021 - Security Misconfiguration

```go
// ✓ Good: Disable debug in production
if cfg.Environment == "production" {
    gin.SetMode(gin.ReleaseMode)
    app.Settings.DisableStartupMessage = true
}

// ✓ Good: Secure CORS configuration
app.Use(cors.New(cors.Config{
    AllowOrigins:     cfg.CORSOrigins,  // Not "*" in production
    AllowMethods:     "GET,POST,PUT,DELETE",
    AllowHeaders:     "Origin,Content-Type,Authorization",
    AllowCredentials: true,
}))

// ✓ Good: Security headers
app.Use(helmet.New())  // Sets X-Content-Type-Options, X-Frame-Options, etc.

// ✓ Good: File permissions (0600 for sensitive files)
os.WriteFile(path, data, 0600)  // Not 0644 or 0777
```

### A06:2021 - Vulnerable and Outdated Components

```bash
# ✓ Good: Regularly update dependencies
go get -u ./...
go mod tidy

# ✓ Good: Check for vulnerabilities
govulncheck ./...

# ✓ Good: Use dependabot or renovate for automated updates
```

### A07:2021 - Identification and Authentication Failures

```go
// ✓ Good: Strong password validation
type RegisterRequest struct {
    Password string `validate:"required,min=8,password"`  // Custom validator
}

// Custom password validator: requires upper, lower, digit
func passwordValidator(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    var hasUpper, hasLower, hasDigit bool
    for _, r := range password {
        switch {
        case r >= 'A' && r <= 'Z': hasUpper = true
        case r >= 'a' && r <= 'z': hasLower = true
        case r >= '0' && r <= '9': hasDigit = true
        }
    }
    return hasUpper && hasLower && hasDigit
}

// ✓ Good: JWT with proper expiration
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": user.ID,
    "exp":     time.Now().Add(24 * time.Hour).Unix(),  // Not indefinite
    "iat":     time.Now().Unix(),
})
```

### A08:2021 - Software and Data Integrity Failures

```go
// ✓ Good: Validate input data integrity
func (uc *UseCase) ProcessPayment(ctx context.Context, req PaymentRequest) error {
    // Validate request signature if from external source
    if !verifySignature(req.Data, req.Signature, publicKey) {
        return errors.BadRequest("invalid signature")
    }
    // Process payment...
}

// ✓ Good: Use checksums for file uploads
func validateUpload(file multipart.File, expectedHash string) error {
    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        return err
    }
    if hex.EncodeToString(hash.Sum(nil)) != expectedHash {
        return errors.BadRequest("file integrity check failed")
    }
    return nil
}
```

### A09:2021 - Security Logging and Monitoring Failures

```go
// ✓ Good: Log security events
logger.Warn("failed login attempt",
    zap.String("email", email),
    zap.String("ip", c.IP()),
    zap.String("user_agent", c.Get("User-Agent")),
)

logger.Info("user logged in",
    zap.String("user_id", user.ID),
    zap.String("ip", c.IP()),
)

// ✓ Good: Log access to sensitive data
logger.Info("payslip accessed",
    zap.String("accessor_id", authCtx.UserID),
    zap.String("target_user_id", targetUserID),
    zap.String("action", "view_payslip"),
)

// ✗ Bad: Logging sensitive data
logger.Info("user login", zap.String("password", password))  // NEVER DO THIS
```

### A10:2021 - Server-Side Request Forgery (SSRF)

```go
// ✓ Good: Whitelist allowed URLs/hosts
var allowedHosts = map[string]bool{
    "api.example.com": true,
    "cdn.example.com": true,
}

func fetchURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return errors.BadRequest("invalid URL")
    }
    if !allowedHosts[u.Host] {
        return errors.Forbidden("host not allowed")
    }
    // Proceed with request...
}

// ✓ Good: Block internal IPs
func isInternalIP(ip net.IP) bool {
    privateBlocks := []string{
        "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8",
    }
    for _, block := range privateBlocks {
        _, cidr, _ := net.ParseCIDR(block)
        if cidr.Contains(ip) {
            return true
        }
    }
    return false
}
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

# Build application
make build

# Run with hot reload
make dev

# Docker compose
make compose-up
make compose-down

# Generate proto (if using protobuf)
make proto

# Database migrations
make migrate
make migrate-create name=create_orders_table
make fresh-seed
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
