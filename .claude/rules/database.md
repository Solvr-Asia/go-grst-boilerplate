# Database Best Practices

## 1. GORM Usage

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

See also [codebase-conventions](codebase-conventions.md): golang-migrate is
authoritative, and updates are column-scoped (`Updates`, not `Save`).

## 2. Connection Pool

```go
// Configure connection pool
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```
