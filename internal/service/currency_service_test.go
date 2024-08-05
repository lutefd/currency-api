package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/service"
)

type mockRepository struct {
	currencies map[string]*model.Currency
}

func (m *mockRepository) GetByCode(ctx context.Context, code string) (*model.Currency, error) {
	if currency, ok := m.currencies[code]; ok {
		return currency, nil
	}
	return nil, errors.New("currency not found")
}

func (m *mockRepository) Create(ctx context.Context, currency *model.Currency) error {
	m.currencies[currency.Code] = currency
	return nil
}

func (m *mockRepository) Update(ctx context.Context, currency *model.Currency) error {
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

type mockExternalAPI struct {
	rates map[string]float64
}

func (m *mockExternalAPI) FetchRates(ctx context.Context) (*model.ExchangeRates, error) {
	return &model.ExchangeRates{
		Rates: m.rates,
	}, nil
}

func TestCurrencyService_Convert(t *testing.T) {
	repo := &mockRepository{
		currencies: map[string]*model.Currency{
			"USD": {Code: "USD", Rate: 1.0},
			"EUR": {Code: "EUR", Rate: 0.85},
		},
	}
	cache := &mockCache{
		data: map[string]float64{},
	}
	externalAPI := &mockExternalAPI{
		rates: map[string]float64{
			"USD": 1.0,
			"EUR": 0.85,
			"GBP": 0.75,
		},
	}

	currencyService := service.NewCurrencyService(repo, cache, externalAPI)

	tests := []struct {
		name          string
		from          string
		to            string
		amount        float64
		expected      float64
		expectedError bool
	}{
		{"USD to EUR", "USD", "EUR", 100, 85, false},
		{"EUR to USD", "EUR", "USD", 85, 100, false},
		{"USD to GBP", "USD", "GBP", 100, 75, false},
		{"Invalid currency", "USD", "XYZ", 100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := currencyService.Convert(context.Background(), tt.from, tt.to, tt.amount)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, but got %f", tt.expected, result)
				}
			}
		})
	}
}

func TestCurrencyService_AddCurrency(t *testing.T) {
	repo := &mockRepository{
		currencies: map[string]*model.Currency{},
	}
	cache := &mockCache{
		data: map[string]float64{},
	}
	externalAPI := &mockExternalAPI{
		rates: map[string]float64{},
	}

	currencyService := service.NewCurrencyService(repo, cache, externalAPI)

	tests := []struct {
		name          string
		code          string
		rate          float64
		expectedError bool
	}{
		{"Add new currency", "JPY", 110.0, false},
		{"Add existing currency", "JPY", 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := currencyService.AddCurrency(context.Background(), tt.code, tt.rate)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				currency, err := repo.GetByCode(context.Background(), tt.code)
				if err != nil {
					t.Errorf("failed to get added currency: %v", err)
				}
				if currency.Rate != tt.rate {
					t.Errorf("expected rate %f, but got %f", tt.rate, currency.Rate)
				}
			}
		})
	}
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
	externalAPI := &mockExternalAPI{
		rates: map[string]float64{},
	}

	currencyService := service.NewCurrencyService(repo, cache, externalAPI)

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
