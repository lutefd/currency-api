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

type MockCurrencyRepository struct {
	mock.Mock
}

func (m *MockCurrencyRepository) GetByCode(ctx context.Context, code string) (*model.Currency, error) {
	args := m.Called(ctx, code)
	if args.Get(0) != nil {
		return args.Get(0).(*model.Currency), args.Error(1)
	}
	return nil, args.Error(1)
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
	if args.Get(0) != nil {
		return args.Get(0).(*model.ExchangeRates), args.Error(1)
	}
	return nil, args.Error(1)
}

func newTestRateUpdater() (*RateUpdater, *MockCurrencyRepository, *MockCache, *MockExternalAPIClient) {
	repo := &MockCurrencyRepository{}
	cache := &MockCache{}
	externalAPI := &MockExternalAPIClient{}
	updater := NewRateUpdater(repo, cache, externalAPI, 10*time.Millisecond)
	return updater, repo, cache, externalAPI
}

func TestNewRateUpdater(t *testing.T) {
	repo := &MockCurrencyRepository{}
	cache := &MockCache{}
	externalAPI := &MockExternalAPIClient{}
	interval := time.Hour

	updater := NewRateUpdater(repo, cache, externalAPI, interval)

	assert.Equal(t, repo, updater.repo, "expected repo to be correctly assigned")
	assert.Equal(t, cache, updater.cache, "expected cache to be correctly assigned")
	assert.Equal(t, externalAPI, updater.externalAPI, "expected externalAPI to be correctly assigned")
	assert.Equal(t, interval, updater.interval, "expected interval to be correctly assigned")
}

func TestRateUpdater_updateRates(t *testing.T) {
	updater, repo, cache, externalAPI := newTestRateUpdater()

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

	assert.NoError(t, err, "expected no error during updateRates")
	externalAPI.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestRateUpdater_updateRates_Error(t *testing.T) {
	updater, _, _, externalAPI := newTestRateUpdater()

	ctx := context.Background()

	externalAPI.On("FetchRates", ctx).Return((*model.ExchangeRates)(nil), errors.New("API error"))

	err := updater.updateRates(ctx)

	assert.Error(t, err, "expected an error due to failed API fetch")
	assert.Contains(t, err.Error(), "failed to fetch rates", "expected error to contain specific message")
	externalAPI.AssertExpectations(t)
}

func TestRateUpdater_populateRates(t *testing.T) {
	updater, repo, cache, externalAPI := newTestRateUpdater()

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

	assert.NoError(t, err, "expected no error during populateRates")
	externalAPI.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestRateUpdater_Start(t *testing.T) {
	updater, repo, cache, externalAPI := newTestRateUpdater()
	updater.interval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
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

	doneChan := make(chan struct{})

	go func() {
		updater.Start(ctx)
		close(doneChan)
	}()

	time.Sleep(35 * time.Millisecond)
	cancel()

	<-doneChan

	externalAPI.AssertNumberOfCalls(t, "FetchRates", 4)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
