# Coding Standards

## 1. File Organization

- One file per struct/interface when the file exceeds 200 lines
- Group related functions together
- Keep files under 500 lines when possible
- Use meaningful file names that reflect their content

## 2. Naming Conventions

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

## 3. Error Handling

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

## 4. Context Usage

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

## 5. Interface Design

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
