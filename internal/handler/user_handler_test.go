package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lutefd/challenge-bravo/internal/handler"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (model.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserService) GetByAPIKey(ctx context.Context, apiKey string) (model.User, error) {
	args := m.Called(ctx, apiKey)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserService) Create(ctx context.Context, username, password string) (model.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, username, password string) error {
	args := m.Called(ctx, username, password)
	return args.Error(0)
}

func (m *MockUserService) Delete(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

func (m *MockUserService) Authenticate(ctx context.Context, username, password string) (model.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(model.User), args.Error(1)
}

func TestUserHandler_Register(t *testing.T) {
	mockService := new(MockUserService)
	handler := handler.NewUserHandler(mockService)

	t.Run("Successful registration", func(t *testing.T) {
		newUser := model.User{
			ID:       uuid.New(),
			Username: "newuser",
			Role:     model.RoleUser,
			APIKey:   "test-api-key",
		}
		mockService.On("Create", mock.Anything, "newuser", "password123").Return(newUser, nil).Once()

		body := bytes.NewBufferString(`{"username":"newuser","password":"password123"}`)
		req, _ := http.NewRequest("POST", "/register", body)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response model.User
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, newUser, response)

		mockService.AssertExpectations(t)
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

	t.Run("Service error", func(t *testing.T) {
		mockService.On("Create", mock.Anything, "newuser", "password123").Return(model.User{}, errors.New("service error")).Once()

		body := bytes.NewBufferString(`{"username":"newuser","password":"password123"}`)
		req, _ := http.NewRequest("POST", "/register", body)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_Login(t *testing.T) {
	mockService := new(MockUserService)
	handler := handler.NewUserHandler(mockService)

	t.Run("Successful login", func(t *testing.T) {
		authenticatedUser := model.User{
			ID:       uuid.New(),
			Username: "testuser",
			Role:     model.RoleUser,
			APIKey:   "test-api-key",
		}
		mockService.On("Authenticate", mock.Anything, "testuser", "password123").Return(authenticatedUser, nil).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"password123"}`)
		req, _ := http.NewRequest("POST", "/login", body)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response model.User
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, authenticatedUser, response)

		mockService.AssertExpectations(t)
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

	t.Run("Authentication failure", func(t *testing.T) {
		mockService.On("Authenticate", mock.Anything, "testuser", "wrongpassword").Return(model.User{}, errors.New("invalid credentials")).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"wrongpassword"}`)
		req, _ := http.NewRequest("POST", "/login", body)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		mockService.AssertExpectations(t)
	})
}
