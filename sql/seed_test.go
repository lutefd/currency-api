package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/server"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockDeps := dependencies{
		loadConfig: func() (server.Config, error) {
			return server.Config{PostgresConn: "mock"}, nil
		},
		openDB: func(driverName, dataSourceName string) (*sql.DB, error) {
			return db, nil
		},
		newUUID: func() uuid.UUID {
			return uuid.MustParse("00000000-0000-0000-0000-000000000000")
		},
		timeNow: func() time.Time {
			return time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		},
		loadEnv: func(...string) error {
			return nil
		},
	}

	mock.ExpectPing()

	mock.ExpectExec("INSERT INTO users").WithArgs(
		uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		"admin",
		sqlmock.AnyArg(),
		model.RoleAdmin,
		"00000000-0000-0000-0000-000000000000",
		mockDeps.timeNow(),
		mockDeps.timeNow(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = run(context.Background(), mockDeps)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAdminUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockDeps := dependencies{
		newUUID: func() uuid.UUID {
			return uuid.MustParse("00000000-0000-0000-0000-000000000000")
		},
		timeNow: func() time.Time {
			return time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	mock.ExpectExec("INSERT INTO users").WithArgs(
		uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		"admin",
		sqlmock.AnyArg(),
		model.RoleAdmin,
		"00000000-0000-0000-0000-000000000000",
		mockDeps.timeNow(),
		mockDeps.timeNow(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = createAdminUser(context.Background(), db, mockDeps)
	assert.NoError(t, err)
	fmt.Println("success creating an admin user")
	assert.NoError(t, mock.ExpectationsWereMet())
}
