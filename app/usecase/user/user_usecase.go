package user

import (
	"context"
	"errors"

	"go-grst-boilerplate/entity"
	"go-grst-boilerplate/repository/user_repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailExists    = errors.New("email already registered")
	ErrNotFound       = errors.New("user not found")
	ErrInvalidCreds   = errors.New("invalid credentials")
	ErrPayslipNotFound = errors.New("payslip not found")
)

type UseCase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	GetProfile(ctx context.Context, userID string) (*entity.User, error)
	ListAll(ctx context.Context, input ListInput) ([]entity.User, int64, error)
	GetPayslip(ctx context.Context, employeeID string, year, month int) (*entity.Payslip, error)
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Phone    string
}

type RegisterOutput struct {
	ID    string
	Email string
	Name  string
}

type ListInput struct {
	Page      int
	Size      int
	Search    string
	SortBy    string
	SortOrder string
}

type useCase struct {
	userRepo user_repository.Repository
}

func NewUseCase(userRepo user_repository.Repository) UseCase {
	return &useCase{userRepo: userRepo}
}

func (uc *useCase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	// Check if email exists
	existing, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Email:    input.Email,
		Password: string(hashedPassword),
		Name:     input.Name,
		Phone:    input.Phone,
		Status:   entity.UserStatusPending,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

func (uc *useCase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCreds
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return user, nil
}

func (uc *useCase) GetProfile(ctx context.Context, userID string) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (uc *useCase) ListAll(ctx context.Context, input ListInput) ([]entity.User, int64, error) {
	return uc.userRepo.FindAll(ctx, user_repository.ListParams{
		Page:      input.Page,
		Size:      input.Size,
		Search:    input.Search,
		SortBy:    input.SortBy,
		SortOrder: input.SortOrder,
	})
}

func (uc *useCase) GetPayslip(ctx context.Context, employeeID string, year, month int) (*entity.Payslip, error) {
	payslip, err := uc.userRepo.FindPayslip(ctx, employeeID, year, month)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPayslipNotFound
		}
		return nil, err
	}
	return payslip, nil
}
