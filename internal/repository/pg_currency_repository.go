package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lutefd/challenge-bravo/internal/model"
	_ "github.com/lib/pq"
)

type PostgresCurrencyRepository struct {
	db *sql.DB
}

func NewPostgresCurrencyRepository(connURL string) (*PostgresCurrencyRepository, error) {
	db, err := sql.Open("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresCurrencyRepository{db: db}, nil
}

func (r *PostgresCurrencyRepository) GetByCode(ctx context.Context, code string) (*model.Currency, error) {
	query := `SELECT code, rate, updated_at, created_by, updated_by, created_at FROM currencies WHERE code = $1`
	var currency model.Currency
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&currency.Code, &currency.Rate, &currency.UpdatedAt,
		&currency.CreatedBy, &currency.UpdatedBy, &currency.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrCurrencyNotFound
		}
		return nil, fmt.Errorf("failed to get currency: %w", err)
	}
	return &currency, nil
}

func (r *PostgresCurrencyRepository) Create(ctx context.Context, currency *model.Currency) error {
	query := `INSERT INTO currencies (code, rate, updated_at, created_by, updated_by, created_at)
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		currency.Code, currency.Rate, currency.UpdatedAt,
		currency.CreatedBy, currency.UpdatedBy, currency.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create currency: %w", err)
	}
	return nil
}

func (r *PostgresCurrencyRepository) Update(ctx context.Context, currency *model.Currency) error {
	query := `UPDATE currencies SET rate = $2, updated_at = $3, updated_by = $4 WHERE code = $1`
	_, err := r.db.ExecContext(ctx, query,
		currency.Code, currency.Rate, currency.UpdatedAt, currency.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to update currency: %w", err)
	}
	return nil
}

func (r *PostgresCurrencyRepository) Delete(ctx context.Context, code string) error {
	query := `DELETE FROM currencies WHERE code = $1`
	_, err := r.db.ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to delete currency: %w", err)
	}
	return nil
}

func (r *PostgresCurrencyRepository) Close() error {
	return r.db.Close()
}
