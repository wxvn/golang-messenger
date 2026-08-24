package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	redisclient "github.com/wxvn/golang-messenger/internal/redis"
)

const userCacheTTL = 5 * time.Minute

var ErrCacheMiss = errors.New("cache miss")

type UserCache interface {
	Get(ctx context.Context, userID uuid.UUID) (User, error)
	Set(ctx context.Context, user User) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type RedisUserCache struct {
	client *redisclient.Client
}

func NewRedisUserCache(client *redisclient.Client) *RedisUserCache {
	return &RedisUserCache{
		client: client,
	}
}

func userCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s", userID)
}

func (c *RedisUserCache) Get(ctx context.Context, userID uuid.UUID) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.client.OpTimeout())
	defer cancel()

	key := userCacheKey(userID)

	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return User{}, ErrCacheMiss
		}

		return User{}, fmt.Errorf("get user from redis: %w", err)
	}

	var user User

	if err := json.Unmarshal([]byte(value), &user); err != nil {
		return User{}, fmt.Errorf("unmarshal user from redis: %w", err)
	}

	return user, nil
}

func (c *RedisUserCache) Set(ctx context.Context, user User) error {
	ctx, cancel := context.WithTimeout(ctx, c.client.OpTimeout())
	defer cancel()

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}

	key := userCacheKey(user.ID)

	if err := c.client.Set(
		ctx,
		key,
		data,
		userCacheTTL,
	).Err(); err != nil {
		return fmt.Errorf("set user to redis: %w", err)
	}

	return nil
}

func (c *RedisUserCache) Delete(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.client.OpTimeout())
	defer cancel()

	key := userCacheKey(userID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete user from redis: %w", err)
	}

	return nil
}
