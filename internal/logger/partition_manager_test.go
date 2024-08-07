package logger

import (
	"context"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) SaveLog(ctx context.Context, log model.Log) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockLogRepository) CreatePartition(ctx context.Context, month time.Time) error {
	args := m.Called(ctx, month)
	return args.Error(0)
}

func (m *MockLogRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewPartitionManager(t *testing.T) {
	mockRepo := new(MockLogRepository)
	pm := NewPartitionManager(mockRepo)

	assert.NotNil(t, pm)
	assert.Equal(t, mockRepo, pm.repo)
	assert.NotNil(t, pm.cron)
}

func timeEqual(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day() &&
		t1.Hour() == t2.Hour() &&
		t1.Minute() == t2.Minute() &&
		t1.Second() == t2.Second()
}

func TestPartitionManager_Start(t *testing.T) {
	mockRepo := new(MockLogRepository)
	pm := NewPartitionManager(mockRepo)

	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now())
	})).Return(nil)
	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now().AddDate(0, 1, 0))
	})).Return(nil)
	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now().AddDate(0, 2, 0))
	})).Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pm.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(LoggerSleepDuration)

	mockRepo.AssertExpectations(t)

	cancel()
}

func TestPartitionManager_createInitialPartitions(t *testing.T) {
	mockRepo := new(MockLogRepository)
	pm := NewPartitionManager(mockRepo)

	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now())
	})).Return(nil)
	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now().AddDate(0, 1, 0))
	})).Return(nil)
	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now().AddDate(0, 2, 0))
	})).Return(nil)

	err := pm.createInitialPartitions(context.Background())
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestPartitionManager_createNextMonthPartition(t *testing.T) {
	mockRepo := new(MockLogRepository)
	pm := NewPartitionManager(mockRepo)

	expectedMonth := time.Now().AddDate(0, 3, 0)
	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return t.Year() == expectedMonth.Year() && t.Month() == expectedMonth.Month()
	})).Return(nil)

	err := pm.createNextMonthPartition(context.Background())
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
func TestPartitionManager_Start_FailedInitialPartition(t *testing.T) {
	mockRepo := new(MockLogRepository)
	pm := NewPartitionManager(mockRepo)

	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now())
	})).Return(nil)
	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return timeEqual(t, time.Now().AddDate(0, 1, 0))
	})).Return(assert.AnError)

	err := pm.Start(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create initial partitions")

	mockRepo.AssertExpectations(t)
}

func TestPartitionManager_cronJob(t *testing.T) {
	mockRepo := new(MockLogRepository)
	pm := NewPartitionManager(mockRepo)

	nextMonth := time.Now().AddDate(0, 3, 0)
	mockRepo.On("CreatePartition", mock.Anything, mock.MatchedBy(func(month time.Time) bool {
		return month.Year() == nextMonth.Year() && month.Month() == nextMonth.Month()
	})).Return(nil)

	pm.cron.Start()
	entries := pm.cron.Entries()
	assert.Len(t, entries, 1)
	entries[0].Job.Run()

	time.Sleep(LoggerSleepDuration)

	mockRepo.AssertExpectations(t)
}
