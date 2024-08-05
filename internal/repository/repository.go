package repository

import (
	"context"

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
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.User, error)
	Close() error
}
