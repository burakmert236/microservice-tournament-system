package cache

import (
	"context"
	"time"

	"github.com/burakmert236/goodswipe-common/config"
	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg config.RedisConfig) (*RedisClient, *apperrors.AppError) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		apperrors.Wrap(err, apperrors.CodeInternalServer, "redis client ping error")
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() *apperrors.AppError {
	if err := r.client.Close(); err != nil {
		apperrors.Wrap(err, apperrors.CodeInternalServer, "redis client close error")
	}

	return nil
}

func (r *RedisClient) Ping(ctx context.Context) *apperrors.AppError {
	if err := r.client.Ping(ctx).Err(); err != nil {
		apperrors.Wrap(err, apperrors.CodeInternalServer, "redis client ping error")
	}

	return nil
}

func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
