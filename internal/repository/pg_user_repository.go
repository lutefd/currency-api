package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lutefd/challenge-bravo/internal/model"
	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(connURL string) (*PostgresUserRepository, error) {
	db, err := sql.Open("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresUserRepository{db: db}, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, username, password, role, api_key, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.Password, user.Role, user.APIKey, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, username, password, role, api_key, created_at, updated_at FROM users WHERE username = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role, &user.APIKey, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.User, error) {
	query := `SELECT id, username, password, role, api_key, created_at, updated_at FROM users WHERE api_key = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, apiKey).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role, &user.APIKey, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *PostgresUserRepository) Close() error {
	return r.db.Close()
}
