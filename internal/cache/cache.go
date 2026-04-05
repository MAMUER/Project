package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type Client struct {
	rdb *redis.Client
}

// Добавляем метод Close с проверкой ошибок
func (c *Client) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

func NewClient(addr, password string, db int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}
