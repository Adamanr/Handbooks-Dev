package database

import (
	"context"
	"handbooks/internal/config"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

func NewRedisConnection(ctx context.Context, cfg config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		slog.ErrorContext(ctx, "cant ping redis", slog.String("error", err.Error()))
		return nil, err
	}

	return rdb, err
}
