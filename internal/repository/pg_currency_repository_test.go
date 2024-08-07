package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresCurrencyRepository(t *testing.T) {
	t.Run("With real connection", func(t *testing.T) {
		t.Skip("Skipping integration test")

		repo, err := NewPostgresCurrencyRepository("your_real_connection_string", nil)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
		repo.Close()
	})

	t.Run("With mock database", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()

		repo, err := NewPostgresCurrencyRepository("", db)
		assert.NoError(t, err)
		assert.NotNil(t, repo)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
func TestPostgresCurrencyRepository_GetByCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresCurrencyRepository{db: db}

	t.Run("Successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"code", "rate", "updated_at", "created_by", "updated_by", "created_at"}).
			AddRow("USD", 1.0, time.Now(), uuid.New(), uuid.New(), time.Now())

		mock.ExpectQuery("SELECT code, rate, updated_at, created_by, updated_by, created_at FROM currencies WHERE code = \\$1").
			WithArgs("USD").
			WillReturnRows(rows)

		currency, err := repo.GetByCode(context.Background(), "USD")
		assert.NoError(t, err)
		assert.NotNil(t, currency)
		assert.Equal(t, "USD", currency.Code)
	})

	t.Run("Currency not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT code, rate, updated_at, created_by, updated_by, created_at FROM currencies WHERE code = \\$1").
			WithArgs("EUR").
			WillReturnError(sql.ErrNoRows)

		currency, err := repo.GetByCode(context.Background(), "EUR")
		assert.Error(t, err)
		assert.Nil(t, currency)
		assert.Equal(t, model.ErrCurrencyNotFound, err)
	})
}

func TestPostgresCurrencyRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresCurrencyRepository{db: db}

	t.Run("Successful creation", func(t *testing.T) {
		currency := &model.Currency{
			Code:      "EUR",
			Rate:      0.85,
			UpdatedAt: time.Now(),
			CreatedBy: uuid.New(),
			UpdatedBy: uuid.New(),
			CreatedAt: time.Now(),
		}

		mock.ExpectExec("INSERT INTO currencies").
			WithArgs(currency.Code, currency.Rate, currency.UpdatedAt, currency.CreatedBy, currency.UpdatedBy, currency.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), currency)
		assert.NoError(t, err)
	})
}

func TestPostgresCurrencyRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresCurrencyRepository{db: db}

	t.Run("Successful update", func(t *testing.T) {
		currency := &model.Currency{
			Code:      "USD",
			Rate:      1.1,
			UpdatedAt: time.Now(),
			UpdatedBy: uuid.New(),
		}

		mock.ExpectExec("UPDATE currencies SET").
			WithArgs(currency.Code, currency.Rate, currency.UpdatedAt, currency.UpdatedBy).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), currency)
		assert.NoError(t, err)
	})

	t.Run("Currency not found", func(t *testing.T) {
		currency := &model.Currency{
			Code:      "XYZ",
			Rate:      1.0,
			UpdatedAt: time.Now(),
			UpdatedBy: uuid.New(),
		}

		mock.ExpectExec("UPDATE currencies SET").
			WithArgs(currency.Code, currency.Rate, currency.UpdatedAt, currency.UpdatedBy).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), currency)
		assert.Error(t, err)
		assert.Equal(t, model.ErrCurrencyNotFound, err)
	})
}

func TestPostgresCurrencyRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresCurrencyRepository{db: db}

	t.Run("Successful deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM currencies WHERE code = \\$1").
			WithArgs("USD").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), "USD")
		assert.NoError(t, err)
	})
}

func TestPostgresCurrencyRepository_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &PostgresCurrencyRepository{db: db}

	mock.ExpectClose()

	err = repo.Close()
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
