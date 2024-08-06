package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockCurrencyRepository struct {
	mock.Mock
}

func (m *MockCurrencyRepository) GetByCode(ctx context.Context, code string) (*model.Currency, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*model.Currency), args.Error(1)
}

func (m *MockCurrencyRepository) Create(ctx context.Context, currency *model.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *MockCurrencyRepository) Update(ctx context.Context, currency *model.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *MockCurrencyRepository) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockCurrencyRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (float64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value float64, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockExternalAPIClient struct {
	mock.Mock
}

func (m *MockExternalAPIClient) FetchRates(ctx context.Context) (*model.ExchangeRates, error) {
	args := m.Called(ctx)
	return args.Get(0).(*model.ExchangeRates), args.Error(1)
}

func TestNewRateUpdater(t *testing.T) {
	repo := &MockCurrencyRepository{}
	cache := &MockCache{}
	externalAPI := &MockExternalAPIClient{}
	interval := time.Hour

	updater := NewRateUpdater(repo, cache, externalAPI, interval)

	assert.Equal(t, repo, updater.repo)
	assert.Equal(t, cache, updater.cache)
	assert.Equal(t, externalAPI, updater.externalAPI)
	assert.Equal(t, interval, updater.interval)
}

func TestRateUpdater_updateRates(t *testing.T) {
	repo := &MockCurrencyRepository{}
	cache := &MockCache{}
	externalAPI := &MockExternalAPIClient{}
	updater := NewRateUpdater(repo, cache, externalAPI, time.Hour)

	ctx := context.Background()
	mockRates := &model.ExchangeRates{
		Timestamp: time.Now().Unix(),
		Rates: map[string]float64{
			"USD": 1.0,
			"EUR": 0.85,
		},
	}

	externalAPI.On("FetchRates", ctx).Return(mockRates, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*model.Currency")).Return(nil)
	cache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("float64"), 1*time.Hour).Return(nil)

	err := updater.updateRates(ctx)

	assert.NoError(t, err)
	externalAPI.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestRateUpdater_updateRates_Error(t *testing.T) {
	repo := &MockCurrencyRepository{}
	cache := &MockCache{}
	externalAPI := &MockExternalAPIClient{}
	updater := NewRateUpdater(repo, cache, externalAPI, time.Hour)

	ctx := context.Background()

	externalAPI.On("FetchRates", ctx).Return((*model.ExchangeRates)(nil), errors.New("API error"))

	err := updater.updateRates(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch rates")
	externalAPI.AssertExpectations(t)
}

func TestRateUpdater_populateRates(t *testing.T) {
	repo := &MockCurrencyRepository{}
	cache := &MockCache{}
	externalAPI := &MockExternalAPIClient{}
	updater := NewRateUpdater(repo, cache, externalAPI, time.Hour)

	ctx := context.Background()
	mockRates := &model.ExchangeRates{
		Timestamp: time.Now().Unix(),
		Rates: map[string]float64{
			"USD": 1.0,
			"EUR": 0.85,
		},
	}

	externalAPI.On("FetchRates", ctx).Return(mockRates, nil)
	repo.On("GetByCode", ctx, mock.AnythingOfType("string")).Return((*model.Currency)(nil), errors.New("currency not found"))
	repo.On("Create", ctx, mock.AnythingOfType("*model.Currency")).Return(nil)
	cache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("float64"), 1*time.Hour).Return(nil)

	err := updater.populateRates(ctx)

	assert.NoError(t, err)
	externalAPI.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestRateUpdater_Start(t *testing.T) {
	repo := &MockCurrencyRepository{}
	cache := &MockCache{}
	externalAPI := &MockExternalAPIClient{}
	updater := NewRateUpdater(repo, cache, externalAPI, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	mockRates := &model.ExchangeRates{
		Timestamp: time.Now().Unix(),
		Rates: map[string]float64{
			"USD": 1.0,
			"EUR": 0.85,
		},
	}

	externalAPI.On("FetchRates", mock.Anything).Return(mockRates, nil)
	repo.On("GetByCode", mock.Anything, mock.AnythingOfType("string")).Return((*model.Currency)(nil), errors.New("currency not found"))
	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.Currency")).Return(nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Currency")).Return(nil)
	cache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("float64"), 1*time.Hour).Return(nil)

	go updater.Start(ctx)

	<-ctx.Done()

	// The updater should have run at least twice (initial populate + at least one update)
	externalAPI.AssertNumberOfCalls(t, "FetchRates", 3)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
