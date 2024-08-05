package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func (m *MockUserRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := NewUserHandler(mockRepo)

	t.Run("Successful registration", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.UserDB")).Return(nil).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"testpass"}`)
		req, _ := http.NewRequest("POST", "/register", body)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response model.UserDB
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ID)
		assert.Equal(t, "testuser", response.Username)
		assert.Empty(t, response.Password) // Password should not be returned
		assert.Equal(t, model.RoleUser, response.Role)
		assert.NotEmpty(t, response.APIKey)
	})

	t.Run("Invalid request payload", func(t *testing.T) {
		body := bytes.NewBufferString(`{"invalid":"json"}`)
		req, _ := http.NewRequest("POST", "/register", body)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Empty username or password", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"","password":""}`)
		req, _ := http.NewRequest("POST", "/register", body)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.UserDB")).Return(assert.AnError).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"testpass"}`)
		req, _ := http.NewRequest("POST", "/register", body)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := NewUserHandler(mockRepo)

	t.Run("Successful login", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
		mockUser := &model.UserDB{
			ID:       uuid.New(),
			Username: "testuser",
			Password: string(hashedPassword),
			Role:     model.RoleUser,
			APIKey:   "test-api-key",
		}
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(mockUser, nil).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"testpass"}`)
		req, _ := http.NewRequest("POST", "/login", body)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response model.User
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, mockUser.ID, response.ID)
		assert.Equal(t, "testuser", response.Username)
		assert.Equal(t, model.RoleUser, response.Role)
		assert.Equal(t, "test-api-key", response.APIKey)
	})

	t.Run("Invalid request payload", func(t *testing.T) {
		body := bytes.NewBufferString(`{"invalid":"json"}`)
		req, _ := http.NewRequest("POST", "/login", body)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Empty username or password", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"","password":""}`)
		req, _ := http.NewRequest("POST", "/login", body)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("User not found", func(t *testing.T) {
		mockRepo.On("GetByUsername", mock.Anything, "nonexistent").Return((*model.UserDB)(nil), assert.AnError).Once()

		body := bytes.NewBufferString(`{"username":"nonexistent","password":"testpass"}`)
		req, _ := http.NewRequest("POST", "/login", body)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Incorrect password", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
		mockUser := &model.UserDB{
			ID:       uuid.New(),
			Username: "testuser",
			Password: string(hashedPassword),
			Role:     model.RoleUser,
			APIKey:   "test-api-key",
		}
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(mockUser, nil).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"wrongpass"}`)
		req, _ := http.NewRequest("POST", "/login", body)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
