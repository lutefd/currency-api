package worker

import (
	"context"

	"github.com/Lutefd/challenge-bravo/internal/model"
)

type ExternalAPIClient interface {
	FetchRates(ctx context.Context) (*model.ExchangeRates, error)
}
