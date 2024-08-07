package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
)

type PostgresLogRepository struct {
	db *sql.DB
}

func NewPostgresLogRepository(connURL string, db *sql.DB) (*PostgresLogRepository, error) {
	if db == nil {
		var err error
		db, err = sql.Open("postgres", connURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		err = db.Ping()
		if err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
	}

	return &PostgresLogRepository{db: db}, nil
}

func (r *PostgresLogRepository) SaveLog(ctx context.Context, log model.Log) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO logs (id, level, message, timestamp, source)
		VALUES ($1, $2, $3, $4, $5)
	`, log.ID, log.Level, log.Message, log.Timestamp, log.Source)
	if err != nil {
		return fmt.Errorf("failed to save log: %w", err)
	}
	return nil
}

func (r *PostgresLogRepository) CreatePartition(ctx context.Context, month time.Time) error {
	partitionName := fmt.Sprintf("logs_y%04dm%02d", month.Year(), month.Month())
	startDate := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF logs
		FOR VALUES FROM ('%s') TO ('%s')
	`, partitionName, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create partition %s: %w", partitionName, err)
	}

	return nil
}
func (r *PostgresLogRepository) Close() error {
	return r.db.Close()
}
