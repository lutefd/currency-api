package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/cache"
	"github.com/Lutefd/challenge-bravo/internal/repository"
)

// Mock implementations (same as before, but with some additions)

type mockCurrencyRepository struct {
	repository.CurrencyRepository
	closeCalled bool
	closeErr    error
}

func (m *mockCurrencyRepository) Close() error {
	m.closeCalled = true
	return m.closeErr
}

type mockCache struct {
	cache.Cache
	closeCalled bool
	closeErr    error
}

func (m *mockCache) Close() error {
	m.closeCalled = true
	return m.closeErr
}

type mockLogRepository struct {
	repository.LogRepository
	closeCalled bool
	closeErr    error
}

func (m *mockLogRepository) Close() error {
	m.closeCalled = true
	return m.closeErr
}

type mockRateUpdater struct {
	startCalled bool
}

func (m *mockRateUpdater) Start(ctx context.Context) {
	m.startCalled = true
}

type mockPartitionManager struct {
	startErr error
}

func (m *mockPartitionManager) Start(ctx context.Context) error {
	return m.startErr
}

func TestRunWorker(t *testing.T) {
	tests := []struct {
		name                   string
		deps                   *dependencies
		expectedErrMsg         string
		timeout                time.Duration
		expectRateUpdaterStart bool
		setupContext           func() (context.Context, context.CancelFunc)
	}{
		{
			name: "Success case",
			deps: &dependencies{
				currencyRepo: &mockCurrencyRepository{},
				cache:        &mockCache{},
				logRepo:      &mockLogRepository{},
				rateUpdater:  &mockRateUpdater{},
				partitionMgr: &mockPartitionManager{},
			},
			expectedErrMsg:         "context deadline exceeded",
			timeout:                100 * time.Millisecond,
			expectRateUpdaterStart: true,
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 100*time.Millisecond)
			},
		},
		{
			name: "Partition manager start error",
			deps: &dependencies{
				currencyRepo: &mockCurrencyRepository{},
				cache:        &mockCache{},
				logRepo:      &mockLogRepository{},
				rateUpdater:  &mockRateUpdater{},
				partitionMgr: &mockPartitionManager{startErr: errors.New("partition manager error")},
			},
			expectedErrMsg:         "failed to start partition manager: partition manager error",
			timeout:                100 * time.Millisecond,
			expectRateUpdaterStart: false,
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 100*time.Millisecond)
			},
		},
		{
			name: "Context cancelled immediately",
			deps: &dependencies{
				currencyRepo: &mockCurrencyRepository{},
				cache:        &mockCache{},
				logRepo:      &mockLogRepository{},
				rateUpdater:  &mockRateUpdater{},
				partitionMgr: &mockPartitionManager{},
			},
			expectedErrMsg:         "context canceled",
			expectRateUpdaterStart: false,
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupContext()
			defer cancel()

			err := runWorker(ctx, tt.deps)

			if tt.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("Expected error with message '%s', but got no error", tt.expectedErrMsg)
				} else if err.Error() != tt.expectedErrMsg {
					t.Errorf("Expected error with message '%s', but got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			mockRateUpdater, ok := tt.deps.rateUpdater.(*mockRateUpdater)
			if !ok {
				t.Fatal("rateUpdater is not a mockRateUpdater")
			}
			if tt.expectRateUpdaterStart && !mockRateUpdater.startCalled {
				t.Error("Expected RateUpdater.Start to be called, but it wasn't")
			} else if !tt.expectRateUpdaterStart && mockRateUpdater.startCalled {
				t.Error("Expected RateUpdater.Start not to be called, but it was")
			}

			mockCurrencyRepo, ok := tt.deps.currencyRepo.(*mockCurrencyRepository)
			if !ok {
				t.Fatal("currencyRepo is not a mockCurrencyRepository")
			}
			mockCache, ok := tt.deps.cache.(*mockCache)
			if !ok {
				t.Fatal("cache is not a mockCache")
			}
			mockLogRepo, ok := tt.deps.logRepo.(*mockLogRepository)
			if !ok {
				t.Fatal("logRepo is not a mockLogRepository")
			}

			if !mockCurrencyRepo.closeCalled {
				t.Error("Expected currency repository Close to be called")
			}
			if !mockCache.closeCalled {
				t.Error("Expected cache Close to be called")
			}
			if !mockLogRepo.closeCalled {
				t.Error("Expected log repository Close to be called")
			}
		})
	}
}
