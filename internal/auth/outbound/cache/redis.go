package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
)

type Redis struct {
	client *redis.Client
	config config.Config
}

var ErrEmptyResendKey = errors.New("empty resend key")

func NewRedis(client *redis.Client, config config.Config) *Redis {
	return &Redis{client: client, config: config}
}

func (c *Redis) RegisterResendAllow(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, ErrEmptyResendKey
	}
	if ttl <= 0 {
		return true, nil
	}

	added, err := c.client.SetNX(ctx, key, struct{}{}, ttl).Result()
	if err != nil {
		return false, err
	}

	return added, nil
}
