package database

import (
	"context"
	"fmt"

	"blog/internal/config"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       0,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{rdb}, nil
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
