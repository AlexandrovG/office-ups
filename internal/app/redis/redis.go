package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "ups_service."

type Client struct {
	client *redis.Client
}

func New(ctx context.Context, host, password string, port int) (*Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       0,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return &Client{client: redisClient}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}