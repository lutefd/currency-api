package service

import (
	"context"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/google/uuid"
)

type CurrencyServiceInterface interface {
	Convert(ctx context.Context, from, to string, amount float64) (float64, error)
	AddCurrency(ctx context.Context, currency *model.Currency) error
	UpdateCurrency(ctx context.Context, code string, rate float64, updatedBy uuid.UUID) error
	RemoveCurrency(ctx context.Context, code string) error
}

type UserServiceInterface interface {
	GetByUsername(ctx context.Context, username string) (model.User, error)
	GetByAPIKey(ctx context.Context, apiKey string) (model.User, error)
	Authenticate(ctx context.Context, username, password string) (model.User, error)
	Create(ctx context.Context, username, password string) (model.User, error)
	Update(ctx context.Context, username, password string) error
	Delete(ctx context.Context, username string) error
}
