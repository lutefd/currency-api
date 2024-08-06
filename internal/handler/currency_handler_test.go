package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lutefd/challenge-bravo/internal/handler"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCurrencyService struct {
	mock.Mock
}

func (m *MockCurrencyService) Convert(ctx context.Context, from, to string, amount float64) (float64, error) {
	args := m.Called(ctx, from, to, amount)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCurrencyService) AddCurrency(ctx context.Context, curr *model.Currency) error {
	args := m.Called(ctx, curr)
	return args.Error(0)
}
func (m *MockCurrencyService) UpdateCurrency(ctx context.Context, code string, rate float64, updatedBy uuid.UUID) error {
	args := m.Called(ctx, code, rate, updatedBy)
	return args.Error(1)
}

func (m *MockCurrencyService) RemoveCurrency(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func TestConvertCurrency(t *testing.T) {
	mockService := new(MockCurrencyService)
	h := handler.NewCurrencyHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/convert?from=USD&to=EUR&amount=100", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		mockService.On("Convert", mock.Anything, "USD", "EUR", 100.0).Return(90.0, nil)

		handler := http.HandlerFunc(h.ConvertCurrency)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, `{"from":"USD","to":"EUR","amount":100,"result":90}`, rr.Body.String())
		mockService.AssertExpectations(t)
	})

	t.Run("Missing Parameters", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/convert", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.ConvertCurrency)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.JSONEq(t, `{"error":"Missing required parameters"}`, rr.Body.String())
	})

}

func TestAddCurrency(t *testing.T) {
	mockService := new(MockCurrencyService)
	h := handler.NewCurrencyHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		body := strings.NewReader(`{"code":"USD","rate":1.0}`)
		req, err := http.NewRequest("POST", "/currency", body)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		userID := uuid.New()
		user := &model.User{ID: userID, Username: "testuser", Role: model.RoleAdmin}
		ctx := context.WithValue(req.Context(), "user", user)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		mockService.On("AddCurrency", mock.Anything, mock.AnythingOfType("*model.Currency")).Return(nil).Once()

		handler := http.HandlerFunc(h.AddCurrency)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.JSONEq(t, `{"message":"currency added successfully"}`, rr.Body.String())
		mockService.AssertExpectations(t)

		mockService.AssertCalled(t, "AddCurrency", mock.Anything, mock.MatchedBy(func(c *model.Currency) bool {
			return c.Code == "USD" && c.Rate == 1.0 && c.CreatedBy == userID && c.UpdatedBy == userID
		}))
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		body := strings.NewReader(`{"code":"USD"}`)
		req, err := http.NewRequest("POST", "/currency", body)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.AddCurrency)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.JSONEq(t, `{"error":"invalid currency code or rate"}`, rr.Body.String())
	})

	t.Run("No User in Context", func(t *testing.T) {
		body := strings.NewReader(`{"code":"USD","rate":1.0}`)
		req, err := http.NewRequest("POST", "/currency", body)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(h.AddCurrency)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.JSONEq(t, `{"error":"user information not available"}`, rr.Body.String())
	})
}

func TestRemoveCurrency(t *testing.T) {
	mockService := new(MockCurrencyService)
	h := handler.NewCurrencyHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/currency/USD", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		mockService.On("RemoveCurrency", mock.Anything, "USD").Return(nil)

		router := chi.NewRouter()
		router.Delete("/currency/{code}", h.RemoveCurrency)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, `{"message":"currency removed successfully"}`, rr.Body.String())
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Code - Empty", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/currency/", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		router := chi.NewRouter()
		router.Delete("/currency/{code}", h.RemoveCurrency)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Invalid Code - Length", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/currency/RR", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		router := chi.NewRouter()
		router.Delete("/currency/{code}", h.RemoveCurrency)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.JSONEq(t, `{"error":"invalid currency code, must be 3 characters long following ISO 4217"}`, rr.Body.String())

	})
}
