package seeds

import (
	"context"
	"fmt"
	"time"

	"go-grst-boilerplate/entity"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seeder handles database seeding
type Seeder struct {
	db *gorm.DB
}

// New creates a new Seeder instance
func New(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

// SeedAll runs all seeders
func (s *Seeder) SeedAll(ctx context.Context) error {
	seeders := []func(context.Context) error{
		s.SeedUsers,
	}

	for _, seeder := range seeders {
		if err := seeder(ctx); err != nil {
			return err
		}
	}

	return nil
}

// SeedUsers seeds the users table
func (s *Seeder) SeedUsers(ctx context.Context) error {
	fmt.Println("Seeding users...")

	users := []struct {
		Email       string
		Password    string
		Name        string
		Phone       string
		Roles       []string
		CompanyCode string
	}{
		{
			Email:       "superadmin@example.com",
			Password:    "SuperAdmin123!",
			Name:        "Super Admin",
			Phone:       "081234567890",
			Roles:       []string{"superadmin", "admin"},
			CompanyCode: "COMPANY-001",
		},
		{
			Email:       "admin@example.com",
			Password:    "Admin123!",
			Name:        "Admin User",
			Phone:       "081234567891",
			Roles:       []string{"admin"},
			CompanyCode: "COMPANY-001",
		},
		{
			Email:       "employee1@example.com",
			Password:    "Employee123!",
			Name:        "John Doe",
			Phone:       "081234567892",
			Roles:       []string{"employee"},
			CompanyCode: "COMPANY-001",
		},
		{
			Email:       "employee2@example.com",
			Password:    "Employee123!",
			Name:        "Jane Smith",
			Phone:       "081234567893",
			Roles:       []string{"employee"},
			CompanyCode: "COMPANY-001",
		},
		{
			Email:       "user@example.com",
			Password:    "User123!",
			Name:        "Regular User",
			Phone:       "081234567894",
			Roles:       []string{"user"},
			CompanyCode: "COMPANY-001",
		},
	}

	for _, u := range users {
		// Check if user already exists
		var existingUser entity.User
		result := s.db.WithContext(ctx).Where("email = ?", u.Email).First(&existingUser)
		if result.Error == nil {
			fmt.Printf("  User %s already exists, skipping...\n", u.Email)
			continue
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", u.Email, err)
		}

		user := entity.User{
			ID:          uuid.New().String(),
			Email:       u.Email,
			Password:    string(hashedPassword),
			Name:        u.Name,
			Phone:       u.Phone,
			Status:      entity.UserStatusActive,
			Roles:       u.Roles,
			CompanyCode: u.CompanyCode,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return fmt.Errorf("failed to seed user %s: %w", u.Email, err)
		}

		fmt.Printf("  Created user: %s (%s)\n", u.Name, u.Email)
	}

	fmt.Println("Users seeded successfully!")
	return nil
}
