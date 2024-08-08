package service

import (
	"context"
	"testing"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.UserDB) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.UserDB, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*model.UserDB), args.Error(1)
}

func (m *MockUserRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.UserDB, error) {
	args := m.Called(ctx, apiKey)
	return args.Get(0).(*model.UserDB), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.UserDB) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

func (m *MockUserRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestUserService_GetByUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	ctx := context.Background()
	username := "testuser"
	userDB := &model.UserDB{
		ID:       uuid.New(),
		Username: username,
		Role:     model.RoleUser,
		APIKey:   "test-api-key",
	}

	mockRepo.On("GetByUsername", ctx, username).Return(userDB, nil)

	user, err := service.GetByUsername(ctx, username)

	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, userDB.ID, user.ID)
	assert.Equal(t, userDB.Role, user.Role)
	assert.Equal(t, userDB.APIKey, user.APIKey)

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByAPIKey(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	ctx := context.Background()
	api_key := "test-api-key"
	userDB := &model.UserDB{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     model.RoleUser,
		APIKey:   "test-api-key",
	}

	mockRepo.On("GetByAPIKey", ctx, api_key).Return(userDB, nil)

	user, err := service.GetByAPIKey(ctx, api_key)

	assert.NoError(t, err)
	assert.Equal(t, userDB.Username, user.Username)
	assert.Equal(t, userDB.ID, user.ID)
	assert.Equal(t, userDB.Role, user.Role)
	assert.Equal(t, api_key, user.APIKey)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Create(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	ctx := context.Background()
	username := "newuser"
	password := "password123"

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.UserDB")).Return(nil)

	user, err := service.Create(ctx, username, password)

	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, model.RoleUser, user.Role)
	assert.NotEmpty(t, user.APIKey)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Authenticate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	ctx := context.Background()
	username := "testuser"
	password := "password123"

	hashedPassword, _ := generateHashedPassword(password)
	userDB := &model.UserDB{
		ID:       uuid.New(),
		Username: username,
		Password: hashedPassword,
		Role:     model.RoleUser,
		APIKey:   "test-api-key",
	}

	mockRepo.On("GetByUsername", ctx, username).Return(userDB, nil)

	user, err := service.Authenticate(ctx, username, password)

	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, userDB.ID, user.ID)
	assert.Equal(t, userDB.Role, user.Role)
	assert.Equal(t, userDB.APIKey, user.APIKey)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Authenticate_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	ctx := context.Background()
	username := "testuser"
	password := "wrongpassword"

	hashedPassword, _ := generateHashedPassword("correctpassword")
	userDB := &model.UserDB{
		ID:       uuid.New(),
		Username: username,
		Password: hashedPassword,
		Role:     model.RoleUser,
		APIKey:   "test-api-key",
	}

	mockRepo.On("GetByUsername", ctx, username).Return(userDB, nil)

	_, err := service.Authenticate(ctx, username, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())

	mockRepo.AssertExpectations(t)
}

func generateHashedPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
