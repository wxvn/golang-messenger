package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
	opTimeout time.Duration
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{
		Client:    client,
		opTimeout: config.Timeout,
	}, nil
}

func (c *Client) OpTimeout() time.Duration {
	return c.opTimeout
}
