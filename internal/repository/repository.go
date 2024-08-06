package repository

import (
	"context"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
)

type CurrencyRepository interface {
	GetByCode(ctx context.Context, code string) (*model.Currency, error)
	Create(ctx context.Context, currency *model.Currency) error
	Update(ctx context.Context, currency *model.Currency) error
	Delete(ctx context.Context, code string) error
	Close() error
}

type UserRepository interface {
	Create(ctx context.Context, user *model.UserDB) error
	GetByUsername(ctx context.Context, username string) (*model.UserDB, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.UserDB, error)
	Update(ctx context.Context, user *model.UserDB) error
	Delete(ctx context.Context, username string) error
	Close() error
}

type LogRepository interface {
	SaveLog(ctx context.Context, log model.Log) error
	CreatePartition(ctx context.Context, month time.Time) error
	Close() error
}
