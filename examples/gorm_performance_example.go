package main

// This file demonstrates GORM performance optimization usage
// Run with: go run examples/gorm_performance_example.go

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-grst-boilerplate/config"
	"go-grst-boilerplate/entity"
	"go-grst-boilerplate/pkg/database"

	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize logger
	logger, err := config.NewLogger(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database
	db, err := config.NewDatabase(cfg, logger.Logger)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== GORM Performance Examples ===\n")

	// Example 1: Prepared Statements
	example1PreparedStatements(db)

	// Example 2: Field Selection
	example2FieldSelection(db)

	// Example 3: Batch Processing
	example3BatchProcessing(db)

	// Example 4: Pagination
	example4Pagination(db)

	// Example 5: Manual Transactions
	example5ManualTransactions(db)
}

// Example 1: Using Prepared Statements
func example1PreparedStatements(db *gorm.DB) {
	fmt.Println("▶ Example 1: Prepared Statements")
	fmt.Println("PrepareStmt caches SQL statements for faster repeated queries\n")

	// Method 1: Global (already enabled via DB_PREPARE_STMT=true)
	var user1 entity.User
	db.First(&user1, 1) // First call: prepares statement
	db.First(&user1, 2) // Subsequent: uses cached statement
	db.First(&user1, 3) // Fast!

	// Method 2: Session mode (when global is disabled)
	tx := database.WithPreparedStmt(db)
	tx.First(&user1, 1)
	tx.Find(&[]entity.User{})

	fmt.Println("✓ Queries executed with prepared statements\n")
}

// Example 2: Field Selection (Avoid SELECT *)
func example2FieldSelection(db *gorm.DB) {
	fmt.Println("▶ Example 2: Field Selection")
	fmt.Println("Select only needed fields to reduce data transfer\n")

	var users []entity.User

	// Bad: SELECT * (slow for large tables)
	// db.Find(&users)

	// Good: SELECT specific fields (fast)
	database.SelectFields(db, "id", "name", "email").Find(&users)
	fmt.Printf("✓ Selected 3 fields from %d users\n\n", len(users))

	// Alternative: Use smaller struct (auto-selection)
	type UserSummary struct {
		ID    uint
		Name  string
		Email string
	}

	var summaries []UserSummary
	db.Model(&entity.User{}).Find(&summaries)
	fmt.Printf("✓ Auto-selected fields using UserSummary struct\n\n")
}

// Example 3: Batch Processing
func example3BatchProcessing(db *gorm.DB) {
	fmt.Println("▶ Example 3: Batch Processing")
	fmt.Println("Process large datasets without loading everything into memory\n")

	var users []entity.User

	// Process users in batches of 100
	err := database.FindInBatches(db, &users, 100, func(tx *gorm.DB, batch int) error {
		fmt.Printf("  Processing batch %d (%d records)\n", batch, len(users))

		for _, user := range users {
			// Process each user (send email, update cache, etc.)
			_ = user // Do something with user
		}

		return nil // Continue to next batch
	})

	if err != nil {
		fmt.Printf("✗ Error: %v\n", err)
	} else {
		fmt.Println("✓ All batches processed successfully\n")
	}
}

// Example 4: Pagination
func example4Pagination(db *gorm.DB) {
	fmt.Println("▶ Example 4: Pagination")
	fmt.Println("Efficiently paginate large datasets\n")

	page := 1
	pageSize := 10

	var users []entity.User

	// Get page 1 (first 10 users)
	database.Paginate(db, page, pageSize).Find(&users)
	fmt.Printf("✓ Page %d: %d users\n", page, len(users))

	// Get page 2 (next 10 users)
	page = 2
	database.Paginate(db, page, pageSize).Find(&users)
	fmt.Printf("✓ Page %d: %d users\n\n", page, len(users))

	// Combined with field selection
	database.Paginate(
		database.SelectFields(db, "id", "name"),
		1,
		pageSize,
	).Find(&users)
	fmt.Println("✓ Pagination with field selection\n")
}

// Example 5: Manual Transactions
func example5ManualTransactions(db *gorm.DB) {
	fmt.Println("▶ Example 5: Manual Transactions")
	fmt.Println("Use when DB_SKIP_DEFAULT_TRANSACTION=true\n")

	// Wrap operations in transaction for ACID guarantees
	err := database.WithTransaction(db, func(tx *gorm.DB) error {
		// All operations in this function run in a transaction

		user := entity.User{
			Name:        "Transaction Test",
			Email:       fmt.Sprintf("test_%d@example.com", time.Now().Unix()),
			Phone:       "+1234567890",
			CompanyCode: "TEST",
		}

		// Create user
		if err := tx.Create(&user).Error; err != nil {
			return err // Automatically rolls back
		}

		// Simulate error condition
		// return errors.New("simulated error") // Would rollback user creation

		// If we reach here, transaction commits
		return nil
	})

	if err != nil {
		fmt.Printf("✗ Transaction failed: %v\n", err)
	} else {
		fmt.Println("✓ Transaction committed successfully\n")
	}
}

// Bonus: Full-Featured API Example
func bonusFullAPIExample(db *gorm.DB, page, pageSize int) ([]entity.User, int64, error) {
	ctx := context.Background()

	var users []entity.User
	var total int64

	// Get total count (for pagination metadata)
	dbCtx := database.WithContext(db, ctx)
	dbCtx.Model(&entity.User{}).Count(&total)

	// Get paginated results with field selection
	err := database.Paginate(
		database.SelectFields(
			database.WithContext(db, ctx),
			"id", "name", "email", "created_at",
		),
		page,
		pageSize,
	).
		Where("deleted_at IS NULL"). // Soft delete filter
		Order("created_at DESC").    // Sort newest first
		Find(&users).Error

	return users, total, err
}
