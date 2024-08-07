package api_middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/commons"
	api_middleware "github.com/Lutefd/challenge-bravo/internal/middleware"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/time/rate"
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

func createTestRequest(apiKey string) *http.Request {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", apiKey)
	return req
}

func TestAuthMiddleware_Authenticate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authMiddleware := api_middleware.NewAuthMiddleware(mockRepo)

	tests := []struct {
		name           string
		apiKey         string
		setupMock      func()
		expectedStatus int
		checkUser      func(*testing.T, *http.Request)
	}{
		{
			name:   "Valid API Key",
			apiKey: "valid-api-key",
			setupMock: func() {
				mockRepo.On("GetByAPIKey", mock.Anything, "valid-api-key").Return(
					&model.UserDB{
						ID:       uuid.New(),
						Username: "testuser",
						Role:     model.RoleUser,
						APIKey:   "valid-api-key",
					}, nil,
				)
			},
			expectedStatus: http.StatusOK,
			checkUser: func(t *testing.T, r *http.Request) {
				user, ok := r.Context().Value(commons.UserContextKey).(model.User)
				assert.True(t, ok)
				assert.Equal(t, "testuser", user.Username)
			},
		},
		{
			name:           "Missing API Key",
			apiKey:         "",
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			checkUser:      func(t *testing.T, r *http.Request) {},
		},
		{
			name:   "Invalid API Key",
			apiKey: "invalid-api-key",
			setupMock: func() {
				mockRepo.On("GetByAPIKey", mock.Anything, "invalid-api-key").Return((*model.UserDB)(nil), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			checkUser:      func(t *testing.T, r *http.Request) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := createTestRequest(tt.apiKey)
			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				tt.checkUser(t, r)
			})

			handler := authMiddleware.Authenticate(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		user           model.User
		requiredRole   model.Role
		expectedStatus int
	}{
		{
			name:           "User has required role",
			user:           model.User{Username: "admin", Role: model.RoleAdmin},
			requiredRole:   model.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User doesn't have required role",
			user:           model.User{Username: "user", Role: model.RoleUser},
			requiredRole:   model.RoleAdmin,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "No user in context",
			user:           model.User{},
			requiredRole:   model.RoleAdmin,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()

			if tt.user != (model.User{}) {
				ctx := context.WithValue(req.Context(), commons.UserContextKey, tt.user)
				req = req.WithContext(ctx)
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := api_middleware.RequireRole(tt.requiredRole)(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	originalClients := api_middleware.Clients
	api_middleware.Clients = make(map[string]*rate.Limiter)
	defer func() {
		api_middleware.Clients = originalClients
	}()

	tests := []struct {
		name           string
		ip             string
		setupLimiter   func()
		expectedStatus int
	}{
		{
			name: "Under rate limit",
			ip:   "192.168.0.1",
			setupLimiter: func() {
				limiter := rate.NewLimiter(rate.Inf, commons.AllowedRPS)
				api_middleware.Clients["192.168.0.1"] = limiter
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Exceeds rate limit",
			ip:   "192.168.0.1",
			setupLimiter: func() {
				limiter := rate.NewLimiter(rate.Every(time.Second), 1)
				api_middleware.Clients["192.168.0.1"] = limiter
				_ = api_middleware.Clients["192.168.0.1"].Allow()
			},
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupLimiter()

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.ip
			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := api_middleware.RateLimitMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
