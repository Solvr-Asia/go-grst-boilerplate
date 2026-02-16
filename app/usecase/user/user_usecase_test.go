package user

import (
	"context"
	"testing"

	"go-grst-boilerplate/entity"
	"go-grst-boilerplate/repository/user_repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository is a mock implementation of user_repository.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindAll(ctx context.Context, params user_repository.ListParams) ([]entity.User, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	input := RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Phone:    "081234567890",
	}

	// Mock FindByEmail returns not found
	mockRepo.On("FindByEmail", ctx, input.Email).Return(nil, gorm.ErrRecordNotFound)

	// Mock Create succeeds
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil)

	result, err := uc.Register(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Email, result.Email)
	assert.Equal(t, input.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestRegister_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	input := RegisterInput{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	existingUser := &entity.User{
		ID:    "existing-id",
		Email: input.Email,
	}

	// Mock FindByEmail returns existing user
	mockRepo.On("FindByEmail", ctx, input.Email).Return(existingUser, nil)

	result, err := uc.Register(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	userID := "user-123"
	expectedUser := &entity.User{
		ID:     userID,
		Email:  "test@example.com",
		Name:   "Test User",
		Status: entity.UserStatusActive,
	}

	mockRepo.On("FindByID", ctx, userID).Return(expectedUser, nil)

	result, err := uc.GetProfile(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	userID := "non-existent"

	mockRepo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := uc.GetProfile(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestListAll_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	input := ListInput{
		Page:      1,
		Size:      10,
		Search:    "",
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	expectedUsers := []entity.User{
		{ID: "user-1", Email: "user1@example.com", Name: "User 1"},
		{ID: "user-2", Email: "user2@example.com", Name: "User 2"},
	}
	expectedTotal := int64(2)

	mockRepo.On("FindAll", ctx, mock.AnythingOfType("user_repository.ListParams")).Return(expectedUsers, expectedTotal, nil)

	users, total, err := uc.ListAll(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, users, 2)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	userID := "user-123"
	existingUser := &entity.User{
		ID:     userID,
		Email:  "test@example.com",
		Name:   "Old Name",
		Phone:  "081234567890",
		Status: entity.UserStatusActive,
	}

	mockRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil)

	result, err := uc.UpdateUser(ctx, userID, UpdateInput{
		Name:  "New Name",
		Phone: "089876543210",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, "089876543210", result.Phone)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	userID := "user-123"
	existingUser := &entity.User{ID: userID}

	mockRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Delete", ctx, userID).Return(nil)

	err := uc.DeleteUser(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	userID := "non-existent"

	mockRepo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}
