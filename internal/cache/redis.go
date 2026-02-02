package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ Cache = (*RedisCache)(nil)

// RedisCache implementa um [Cache] usando o Redis como banco de dados.
type RedisCache struct {
	rdb *redis.Client
}

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{rdb: rdb}
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, ErrCacheMiss
		default:
			return nil, fmt.Errorf("failed to get from redis: %w", err)
		}
	}

	return b, nil
}

func (r *RedisCache) Put(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	err := r.rdb.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key/value to redis: %w", err)
	}
	return nil
}

func (r *RedisCache) Del(ctx context.Context, key string) error {
	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to del from redis: %w", err)
	}
	return nil
}

func (r *RedisCache) Remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error) {
	data, err := r.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	if !errors.Is(err, ErrCacheMiss) {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	data, err = fn()
	if err != nil {
		return nil, err
	}

	err = r.Put(ctx, key, data, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to set to cache: %w", err)
	}

	return data, nil
}
