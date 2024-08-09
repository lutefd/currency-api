package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	currencies map[string]*model.Currency
}

func (m *mockRepository) GetByCode(ctx context.Context, code string) (*model.Currency, error) {
	currency, ok := m.currencies[code]
	if !ok {
		return nil, errors.New("currency not found")
	}
	return currency, nil
}

func (m *mockRepository) Create(ctx context.Context, currency *model.Currency) error {
	m.currencies[currency.Code] = currency
	return nil
}

func (m *mockRepository) Update(ctx context.Context, currency *model.Currency) error {
	if _, ok := m.currencies[currency.Code]; !ok {
		return errors.New("currency not found")
	}
	m.currencies[currency.Code] = currency
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, code string) error {
	delete(m.currencies, code)
	return nil
}

func (m *mockRepository) Close() error {
	return nil
}

type mockCache struct {
	data map[string]float64
}

func (m *mockCache) Get(ctx context.Context, key string) (float64, error) {
	if rate, ok := m.data[key]; ok {
		return rate, nil
	}
	return 0, errors.New("key not found")
}

func (m *mockCache) Set(ctx context.Context, key string, value float64, expiration time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCache) Close() error {
	return nil
}
func TestCurrencyService_Convert(t *testing.T) {
	repo := &mockRepository{
		currencies: map[string]*model.Currency{
			"USD": {Code: "USD", Rate: 1.0},
			"EUR": {Code: "EUR", Rate: 0.85},
			"GBP": {Code: "GBP", Rate: 0.75},
		},
	}
	cache := &mockCache{
		data: map[string]float64{
			"USD": 1.0,
			"EUR": 0.85,
			"GBP": 0.75,
		},
	}

	currencyService := service.NewCurrencyService(repo, cache)

	tests := []struct {
		name          string
		from          string
		to            string
		amount        float64
		expected      float64
		expectedError bool
	}{
		{"USD to EUR", "USD", "EUR", 100, 85, false},
		{"EUR to GBP", "EUR", "GBP", 100, 88.24, false},
		{"GBP to USD", "GBP", "USD", 75, 100, false},
		{"Invalid from currency", "INVALID", "USD", 100, 0, true},
		{"Invalid to currency", "USD", "INVALID", 100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := currencyService.Convert(context.Background(), tt.from, tt.to, tt.amount)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.01)
			}
		})
	}
}

func TestCurrencyService_AddCurrency(t *testing.T) {
	repo := &mockRepository{
		currencies: make(map[string]*model.Currency),
	}
	cache := &mockCache{
		data: make(map[string]float64),
	}

	currencyService := service.NewCurrencyService(repo, cache)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("Add new currency", func(t *testing.T) {
		newCurrency := &model.Currency{
			Code:      "JPY",
			Rate:      110.0,
			CreatedBy: userID,
			UpdatedBy: userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := currencyService.AddCurrency(ctx, newCurrency)

		assert.NoError(t, err)
		assert.Equal(t, newCurrency, repo.currencies["JPY"])
		assert.Equal(t, 110.0, cache.data["JPY"])
	})

	t.Run("Add existing currency", func(t *testing.T) {
		existingCurrency := &model.Currency{
			Code:      "USD",
			Rate:      1.0,
			CreatedBy: userID,
			UpdatedBy: userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.currencies["USD"] = existingCurrency

		newCurrency := &model.Currency{
			Code:      "USD",
			Rate:      1.1,
			CreatedBy: userID,
			UpdatedBy: userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := currencyService.AddCurrency(ctx, newCurrency)

		assert.Error(t, err)
		assert.Equal(t, existingCurrency, repo.currencies["USD"])
	})
}

func TestCurrencyService_UpdateCurrency(t *testing.T) {
	repo := &mockRepository{
		currencies: make(map[string]*model.Currency),
	}
	cache := &mockCache{
		data: make(map[string]float64),
	}

	currencyService := service.NewCurrencyService(repo, cache)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("Update existing currency", func(t *testing.T) {
		existingCurrency := &model.Currency{
			Code:      "EUR",
			Rate:      0.85,
			CreatedBy: uuid.New(),
			UpdatedBy: uuid.New(),
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		}
		repo.currencies["EUR"] = existingCurrency

		originalUpdatedAt := existingCurrency.UpdatedAt

		err := currencyService.UpdateCurrency(ctx, "EUR", 0.82, userID)

		assert.NoError(t, err)
		updatedCurrency := repo.currencies["EUR"]
		assert.Equal(t, 0.82, updatedCurrency.Rate)
		assert.Equal(t, userID, updatedCurrency.UpdatedBy)
		assert.True(t, updatedCurrency.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be later than the original time")
		assert.Equal(t, 0.82, cache.data["EUR"])
	})

	t.Run("Update non-existing currency", func(t *testing.T) {
		err := currencyService.UpdateCurrency(ctx, "GBP", 0.75, userID)

		assert.Error(t, err)
		assert.NotContains(t, repo.currencies, "GBP")
		assert.NotContains(t, cache.data, "GBP")
	})
}
func TestCurrencyService_RemoveCurrency(t *testing.T) {
	repo := &mockRepository{
		currencies: map[string]*model.Currency{
			"USD": {Code: "USD", Rate: 1.0},
			"EUR": {Code: "EUR", Rate: 0.85},
		},
	}
	cache := &mockCache{
		data: map[string]float64{
			"USD": 1.0,
			"EUR": 0.85,
		},
	}

	currencyService := service.NewCurrencyService(repo, cache)

	tests := []struct {
		name          string
		code          string
		expectedError bool
	}{
		{"remove existing currency", "USD", false},
		{"remove non-existing currency", "JPY", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := currencyService.RemoveCurrency(context.Background(), tt.code)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				_, err := repo.GetByCode(context.Background(), tt.code)
				if err == nil {
					t.Errorf("currency should have been removed, but it still exists")
				}
				_, err = cache.Get(context.Background(), tt.code)
				if err == nil {
					t.Errorf("currency should have been removed from cache, but it still exists")
				}
			}
		})
	}
}
