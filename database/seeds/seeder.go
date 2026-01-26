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
		s.SeedPayslips,
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

// SeedPayslips seeds the payslips table
func (s *Seeder) SeedPayslips(ctx context.Context) error {
	fmt.Println("Seeding payslips...")

	// Get employees
	var employees []entity.User
	if err := s.db.WithContext(ctx).Where("? = ANY(roles)", "employee").Find(&employees).Error; err != nil {
		return fmt.Errorf("failed to fetch employees: %w", err)
	}

	if len(employees) == 0 {
		fmt.Println("  No employees found, skipping payslips...")
		return nil
	}

	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	for _, emp := range employees {
		// Create payslips for the last 3 months
		for i := 0; i < 3; i++ {
			month := currentMonth - i
			year := currentYear
			if month <= 0 {
				month += 12
				year--
			}

			// Check if payslip already exists
			var existingPayslip entity.Payslip
			result := s.db.WithContext(ctx).
				Where("employee_id = ? AND year = ? AND month = ?", emp.ID, year, month).
				First(&existingPayslip)
			if result.Error == nil {
				fmt.Printf("  Payslip for %s (%d-%02d) already exists, skipping...\n", emp.Name, year, month)
				continue
			}

			basicSalary := 5000000.0 + float64(i*500000) // Vary salary a bit
			allowances := 1500000.0
			deductions := 500000.0
			grossSalary := basicSalary + allowances
			netSalary := grossSalary - deductions

			payslip := entity.Payslip{
				ID:          uuid.New().String(),
				EmployeeID:  emp.ID,
				Year:        year,
				Month:       month,
				BasicSalary: basicSalary,
				Allowances:  allowances,
				Deductions:  deductions,
				GrossSalary: grossSalary,
				NetSalary:   netSalary,
				Status:      "paid",
				PaidAt:      timePtr(time.Date(year, time.Month(month), 25, 0, 0, 0, 0, time.Local)),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := s.db.WithContext(ctx).Create(&payslip).Error; err != nil {
				return fmt.Errorf("failed to seed payslip for %s: %w", emp.Name, err)
			}

			fmt.Printf("  Created payslip for %s (%d-%02d)\n", emp.Name, year, month)
		}
	}

	fmt.Println("Payslips seeded successfully!")
	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
