package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lutefd/challenge-bravo/internal/commons"
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
	return args.Error(0)
}

func (m *MockCurrencyService) RemoveCurrency(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func TestConvertCurrency(t *testing.T) {
	mockService := new(MockCurrencyService)
	h := handler.NewCurrencyHandler(mockService)

	tests := []struct {
		name           string
		from           string
		to             string
		amount         string
		expectedStatus int
		expectedBody   string
		mockBehavior   func()
	}{
		{
			name:           "Valid conversion",
			from:           "USD",
			to:             "EUR",
			amount:         "100.00",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"amount":100,"from":"USD","result":85,"to":"EUR"}`,
			mockBehavior: func() {
				mockService.On("Convert", mock.Anything, "USD", "EUR", 100.0).Return(85.0, nil).Once()
			},
		},
		{
			name:           "Valid conversion with comma",
			from:           "USD",
			to:             "EUR",
			amount:         "100,00",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"amount":100,"from":"USD","result":85,"to":"EUR"}`,
			mockBehavior: func() {
				mockService.On("Convert", mock.Anything, "USD", "EUR", 100.0).Return(85.0, nil).Once()
			},
		},
		{
			name:           "Negative amount",
			from:           "USD",
			to:             "EUR",
			amount:         "-100.00",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"amount must be non-negative"}`,
			mockBehavior:   func() {},
		},
		{
			name:           "Invalid amount",
			from:           "USD",
			to:             "EUR",
			amount:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid amount"}`,
			mockBehavior:   func() {},
		},
		{
			name:           "From currency not found",
			from:           "XYZ",
			to:             "EUR",
			amount:         "100.00",
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"currency not found: XYZ"}`,
			mockBehavior: func() {
				mockService.On("Convert", mock.Anything, "XYZ", "EUR", 100.0).Return(0.0, fmt.Errorf("%w: XYZ", model.ErrCurrencyNotFound)).Once()
			},
		},
		{
			name:           "To currency not found",
			from:           "USD",
			to:             "XYZ",
			amount:         "100.00",
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"currency not found: XYZ"}`,
			mockBehavior: func() {
				mockService.On("Convert", mock.Anything, "USD", "XYZ", 100.0).Return(0.0, fmt.Errorf("%w: XYZ", model.ErrCurrencyNotFound)).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			req, _ := http.NewRequest("GET", "/convert?from="+tt.from+"&to="+tt.to+"&amount="+tt.amount, nil)
			rr := httptest.NewRecorder()

			h.ConvertCurrency(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}

func TestAddCurrency(t *testing.T) {
	mockService := new(MockCurrencyService)
	h := handler.NewCurrencyHandler(mockService)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedBody   string
		mockBehavior   func()
	}{
		{
			name: "Valid currency with period",
			payload: map[string]interface{}{
				"code":        "USD",
				"rate_to_usd": 1.0,
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"currency added successfully"}`,
			mockBehavior: func() {
				mockService.On("AddCurrency", mock.Anything, mock.MatchedBy(func(c *model.Currency) bool {
					return c.Code == "USD" && c.Rate == 1.0
				})).Return(nil).Once()
			},
		},
		{
			name: "Valid currency with comma",
			payload: map[string]interface{}{
				"code":        "EUR",
				"rate_to_usd": "0,85",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"currency added successfully"}`,
			mockBehavior: func() {
				mockService.On("AddCurrency", mock.Anything, mock.MatchedBy(func(c *model.Currency) bool {
					return c.Code == "EUR" && c.Rate == 0.85
				})).Return(nil).Once()
			},
		},
		{
			name: "Negative rate",
			payload: map[string]interface{}{
				"code":        "USD",
				"rate_to_usd": -1.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"rate must be positive"}`,
			mockBehavior:   func() {},
		},
		{
			name: "Invalid rate type",
			payload: map[string]interface{}{
				"code":        "USD",
				"rate_to_usd": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid rate: strconv.ParseFloat: parsing \"invalid\": invalid syntax"}`,
			mockBehavior:   func() {},
		},
		{
			name: "Missing code",
			payload: map[string]interface{}{
				"rate_to_usd": 1.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid currency code"}`,
			mockBehavior:   func() {},
		},
		{
			name: "Invalid code length - Above Maximum Allowed",
			payload: map[string]interface{}{
				"code":        "USDDDD",
				"rate_to_usd": 1.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   fmt.Sprintf(`{"error":"invalid currency code, must be up to %d characters"}`, commons.AllowedCurrencyLength),
			mockBehavior:   func() {},
		},
		{
			name: "Invalid code length - Below Minimum Allowed",
			payload: map[string]interface{}{
				"code":        "US",
				"rate_to_usd": 1.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   fmt.Sprintf(`{"error":"invalid currency code, must be at least %d characters"}`, commons.MinimumCurrencyLength),
			mockBehavior:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/currency", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			userID := uuid.New()
			user := model.User{ID: userID, Username: "testuser", Role: model.RoleAdmin}
			ctx := context.WithValue(req.Context(), commons.UserContextKey, user)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			h.AddCurrency(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
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

	t.Run("Invalid Code - Length Under Minimum", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/currency/R", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		router := chi.NewRouter()
		router.Delete("/currency/{code}", h.RemoveCurrency)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.JSONEq(t, fmt.Sprintf(`{"error":"invalid currency code, must be at least %d characters"}`, commons.MinimumCurrencyLength), rr.Body.String())

	})
	t.Run("Invalid Code - Length", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/currency/RRRRRR", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		router := chi.NewRouter()
		router.Delete("/currency/{code}", h.RemoveCurrency)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.JSONEq(t, fmt.Sprintf(`{"error":"invalid currency code, must be up to %d characters"}`, commons.AllowedCurrencyLength), rr.Body.String())

	})
}

func TestUpdateCurrency(t *testing.T) {
	mockService := new(MockCurrencyService)
	h := handler.NewCurrencyHandler(mockService)

	tests := []struct {
		name           string
		code           string
		payload        map[string]interface{}
		expectedStatus int
		expectedBody   string
		mockBehavior   func()
	}{
		{
			name: "Valid update with period",
			code: "USD",
			payload: map[string]interface{}{
				"rate_to_usd": 1.5,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"currency updated successfully"}`,
			mockBehavior: func() {
				mockService.On("UpdateCurrency", mock.Anything, "USD", 1.5, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
			},
		},
		{
			name: "Valid update with comma",
			code: "EUR",
			payload: map[string]interface{}{
				"rate_to_usd": "0,95",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"currency updated successfully"}`,
			mockBehavior: func() {
				mockService.On("UpdateCurrency", mock.Anything, "EUR", 0.95, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
			},
		},
		{
			name: "Negative rate",
			code: "GBP",
			payload: map[string]interface{}{
				"rate_to_usd": -1.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"rate must be positive"}`,
			mockBehavior:   func() {},
		},
		{
			name: "Invalid rate type",
			code: "JPY",
			payload: map[string]interface{}{
				"rate_to_usd": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid rate: strconv.ParseFloat: parsing \"invalid\": invalid syntax"}`,
			mockBehavior:   func() {},
		},
		{
			name:           "Missing rate",
			code:           "CAD",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid rate: unsupported rate type"}`,
			mockBehavior:   func() {},
		},
		{
			name: "Currency not found",
			code: "XYZ",
			payload: map[string]interface{}{
				"rate_to_usd": 1.0,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"currency not found"}`,
			mockBehavior: func() {
				mockService.On("UpdateCurrency", mock.Anything, "XYZ", 1.0, mock.AnythingOfType("uuid.UUID")).Return(model.ErrCurrencyNotFound).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("PUT", "/currency/"+tt.code, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			userID := uuid.New()
			user := model.User{ID: userID, Username: "testuser", Role: model.RoleAdmin}
			ctx := context.WithValue(req.Context(), commons.UserContextKey, user)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", tt.code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			h.UpdateCurrency(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
