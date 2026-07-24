# Testing Guidelines

## 1. Unit Tests

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

## 2. Mock Dependencies

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

## 3. Run Tests with Race Detection

```bash
# Always run with race detection
go test -race -v ./...

# With coverage
go test -race -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
