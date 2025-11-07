package redisx

import (
	"context"
	"fmt"
	"time"

	redisLib "github.com/redis/go-redis/v9"
)

// func NewRedis(ctx context.Context, config *redis.Options) (*redis.Client, error) {
// 	client := redis.NewClient(config)

// 	if err := client.Ping(ctx).Err(); err != nil {
// 		return nil, err
// 	}

// 	return client, nil
// }

type Redis interface {
	TTL(ctx context.Context, key string) (time.Duration, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (bool, error)
	Client() *redisLib.Client
}

type redisx struct {
	client *redisLib.Client
}

func NewRedis(ctx context.Context, config *redisLib.Options) (Redis, error) {
	client := redisLib.NewClient(config)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("[redisx] : %w", err)
	}

	return &redisx{
		client: client,
	}, nil
}

func (r *redisx) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return time.Duration(0), fmt.Errorf("[redisx] : %w", err)
	}

	return ttl, nil
}

func (r *redisx) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("[redisx] : %w", err)
	}

	return nil
}

func (r *redisx) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("[redisx] : %w", err)
	}

	return value, nil
}

func (r *redisx) Del(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("[redisx] : %w", err)
	}

	return nil
}

func (r *redisx) Exists(ctx context.Context, keys ...string) (bool, error) {
	exists, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, fmt.Errorf("[redisx] : %w", err)
	}

	return exists > 0, nil
}

func (r *redisx) Client() *redisLib.Client {
	return r.client
}
