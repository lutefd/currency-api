package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/cache"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) (*cache.RedisCache, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	redisCache, err := cache.NewRedisCache(mr.Addr(), "")
	if err != nil {
		t.Fatalf("failed to create Redis cache: %v", err)
	}

	return redisCache, mr
}

func TestNewRedisCache(t *testing.T) {
	redisCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	assert.NotNil(t, redisCache)
}

func TestGet(t *testing.T) {
	redisCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	_, err := redisCache.Get(ctx, "non_existent_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")

	err = redisCache.Set(ctx, "test_key", 123.45, time.Minute)
	assert.NoError(t, err)

	value, err := redisCache.Get(ctx, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, 123.45, value)

	mr.Set("invalid_key", "not_a_float")
	_, err = redisCache.Get(ctx, "invalid_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse cached value")
}

func TestSet(t *testing.T) {
	redisCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	err := redisCache.Set(ctx, "test_key", 123.45, time.Minute)
	assert.NoError(t, err)

	value, err := redisCache.Get(ctx, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, 123.45, value)

	err = redisCache.Set(ctx, "no_expiration_key", 678.90, 0)
	assert.NoError(t, err)

	value, err = redisCache.Get(ctx, "no_expiration_key")
	assert.NoError(t, err)
	assert.Equal(t, 678.90, value)
}

func TestDelete(t *testing.T) {
	redisCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	err := redisCache.Set(ctx, "test_key", 123.45, time.Minute)
	assert.NoError(t, err)

	err = redisCache.Delete(ctx, "test_key")
	assert.NoError(t, err)

	_, err = redisCache.Get(ctx, "test_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")

	err = redisCache.Delete(ctx, "non_existent_key")
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	redisCache, mr := setupTestRedis(t)
	defer mr.Close()

	err := redisCache.Close()
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = redisCache.Get(ctx, "test_key")
	assert.Error(t, err)
}
