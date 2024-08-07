package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresLogRepository(t *testing.T) {
	t.Run("With real connection", func(t *testing.T) {
		t.Skip("Skipping integration test")

		repo, err := NewPostgresLogRepository("your_real_connection_string", nil)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
		repo.Close()
	})

	t.Run("With mock database", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		repo, err := NewPostgresLogRepository("", db)
		assert.NoError(t, err)
		assert.NotNil(t, repo)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPostgresLogRepository_SaveLog(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresLogRepository{db: db}

	t.Run("Successful log save", func(t *testing.T) {
		log := model.Log{
			ID:        uuid.New(),
			Level:     model.LogLevelInfo,
			Message:   "Test log message",
			Timestamp: time.Now(),
			Source:    "test",
		}

		mock.ExpectExec("INSERT INTO logs").
			WithArgs(log.ID, log.Level, log.Message, log.Timestamp, log.Source).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SaveLog(context.Background(), log)
		assert.NoError(t, err)
	})

	t.Run("Failed log save", func(t *testing.T) {
		log := model.Log{
			ID:        uuid.New(),
			Level:     model.LogLevelError,
			Message:   "Test error message",
			Timestamp: time.Now(),
			Source:    "test",
		}

		mock.ExpectExec("INSERT INTO logs").
			WithArgs(log.ID, log.Level, log.Message, log.Timestamp, log.Source).
			WillReturnError(fmt.Errorf("database error"))

		err := repo.SaveLog(context.Background(), log)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save log")
	})
}

func TestPostgresLogRepository_CreatePartition(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresLogRepository{db: db}

	t.Run("Successful partition creation", func(t *testing.T) {
		month := time.Date(2023, time.May, 1, 0, 0, 0, 0, time.UTC)

		mock.ExpectExec("CREATE TABLE IF NOT EXISTS logs_y2023m05 PARTITION OF logs").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.CreatePartition(context.Background(), month)
		assert.NoError(t, err)
	})

	t.Run("Failed partition creation", func(t *testing.T) {
		month := time.Date(2023, time.June, 1, 0, 0, 0, 0, time.UTC)

		mock.ExpectExec("CREATE TABLE IF NOT EXISTS logs_y2023m06 PARTITION OF logs").
			WillReturnError(fmt.Errorf("database error"))

		err := repo.CreatePartition(context.Background(), month)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create partition logs_y2023m06")
	})
}

func TestPostgresLogRepository_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &PostgresLogRepository{db: db}

	mock.ExpectClose()

	err = repo.Close()
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
