package infra

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// NewRedis cria um novo client do Redis para a url especificada.
func NewRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
