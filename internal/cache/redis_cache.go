package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (float64, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("key not found")
		}
		return 0, fmt.Errorf("failed to get from cache: %w", err)
	}

	rate, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse cached value: %w", err)
	}

	return rate, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value float64, expiration time.Duration) error {
	err := c.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set in cache: %w", err)
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}
