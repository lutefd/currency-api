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

func TestPostgresUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresUserRepository{db: db}

	t.Run("Successful creation", func(t *testing.T) {
		user := &model.UserDB{
			ID:        uuid.New(),
			Username:  "testuser",
			Password:  "hashedpassword",
			Role:      model.RoleUser,
			APIKey:    uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Username, user.Password, user.Role, user.APIKey, user.CreatedAt, user.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
	})
}

func TestPostgresUserRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresUserRepository{db: db}

	t.Run("Successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "password", "role", "api_key", "created_at", "updated_at"}).
			AddRow(uuid.New(), "testuser", "hashedpassword", model.RoleUser, uuid.New().String(), time.Now(), time.Now())

		mock.ExpectQuery("SELECT .+ FROM users WHERE username = \\$1").
			WithArgs("testuser").
			WillReturnRows(rows)

		user, err := repo.GetByUsername(context.Background(), "testuser")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT .+ FROM users WHERE username = \\$1").
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByUsername(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestPostgresUserRepository_GetByAPIKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresUserRepository{db: db}

	t.Run("Successful retrieval", func(t *testing.T) {
		apiKey := uuid.New().String()
		rows := sqlmock.NewRows([]string{"id", "username", "password", "role", "api_key", "created_at", "updated_at"}).
			AddRow(uuid.New(), "testuser", "hashedpassword", model.RoleUser, apiKey, time.Now(), time.Now())

		mock.ExpectQuery("SELECT .+ FROM users WHERE api_key = \\$1").
			WithArgs(apiKey).
			WillReturnRows(rows)

		user, err := repo.GetByAPIKey(context.Background(), apiKey)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, apiKey, user.APIKey)
	})
}

func TestPostgresUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresUserRepository{db: db}

	t.Run("Successful update", func(t *testing.T) {
		user := &model.UserDB{
			Username:  "testuser",
			Password:  "newhashpassword",
			Role:      model.RoleAdmin,
			APIKey:    uuid.New().String(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec("UPDATE users SET").
			WithArgs(user.Password, user.Role, user.APIKey, user.UpdatedAt, user.Username).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), user)
		assert.NoError(t, err)
	})

	t.Run("User not found", func(t *testing.T) {
		user := &model.UserDB{
			Username:  "nonexistent",
			Password:  "newhashpassword",
			Role:      model.RoleAdmin,
			APIKey:    uuid.New().String(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec("UPDATE users SET").
			WithArgs(user.Password, user.Role, user.APIKey, user.UpdatedAt, user.Username).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresUserRepository{db: db}

	t.Run("Successful deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM users WHERE username = \\$1").
			WithArgs("testuser").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), "testuser")
		assert.NoError(t, err)
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM users WHERE username = \\$1").
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestPostgresUserRepository_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &PostgresUserRepository{db: db}

	mock.ExpectClose()

	err = repo.Close()
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
