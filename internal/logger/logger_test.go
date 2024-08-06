package logger_test

import (
	"context"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/logger"
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
func (m *MockLogRepository) CreatePartition(ctx context.Context, time time.Time) error {
	args := m.Called(ctx, time)
	return args.Error(0)
}
func (m *MockLogRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestLogger_Info(t *testing.T) {
	mockRepo := new(MockLogRepository)
	logger.InitLogger(mockRepo)

	mockRepo.On("SaveLog", mock.Anything, mock.AnythingOfType("model.Log")).Return(nil)

	logger.Info("Test info message")

	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertCalled(t, "SaveLog", mock.Anything, mock.MatchedBy(func(log model.Log) bool {
		return log.Level == model.LogLevelInfo && log.Message == "Test info message"
	}))
}

func TestLogger_Error(t *testing.T) {
	mockRepo := new(MockLogRepository)
	logger.InitLogger(mockRepo)

	mockRepo.On("SaveLog", mock.Anything, mock.AnythingOfType("model.Log")).Return(nil)

	logger.Error("Test error message")

	time.Sleep(100 * time.Millisecond)
	mockRepo.AssertCalled(t, "SaveLog", mock.Anything, mock.MatchedBy(func(log model.Log) bool {
		return log.Level == model.LogLevelError && log.Message == "Test error message"
	}))
}

func TestLogger_Shutdown(t *testing.T) {
	mockRepo := new(MockLogRepository)
	logger.InitLogger(mockRepo)

	mockRepo.On("SaveLog", mock.Anything, mock.AnythingOfType("model.Log")).Return(nil)
	mockRepo.On("Close").Return(nil)

	logger.Info("Test shutdown message")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := logger.Shutdown(ctx)
	assert.NoError(t, err)

	mockRepo.AssertCalled(t, "Close")
}
