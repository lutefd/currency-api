package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (float64, error)
	Set(ctx context.Context, key string, value float64, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}
