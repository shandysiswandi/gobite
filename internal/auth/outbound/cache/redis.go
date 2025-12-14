package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
)

type Redis struct {
	client *redis.Client
	config pkgconfig.Config
}

func NewRedis(client *redis.Client, config pkgconfig.Config) *Redis {
	return &Redis{client: client, config: config}
}

func (c *Redis) ttlForTokens() (time.Duration, time.Duration) {
	acTTL := time.Duration(c.config.GetInt("jwt.access.ttl")) * time.Minute      // 15 minutes
	refTTL := time.Duration(c.config.GetInt("jwt.refresh.ttl")) * 24 * time.Hour // 7 days

	return acTTL, refTTL
}

func (c *Redis) SaveTokensID(ctx context.Context, acID, refID string) error {
	acTTL, refTTL := c.ttlForTokens()

	_, err := c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, acID, true, acTTL)
		pipe.Set(ctx, refID, true, refTTL)
		return nil
	})

	return err
}

func (c *Redis) DeleteTokensID(ctx context.Context, acID, refID string) error {
	_, err := c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if acID != "" {
			pipe.Del(ctx, acID)
		}
		if refID != "" {
			pipe.Del(ctx, refID)
		}
		return nil
	})

	return err
}

func (c *Redis) IsTokenIDExist(ctx context.Context, id string) (bool, error) {
	count, err := c.client.Exists(ctx, id).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (c *Redis) RotateTokensID(ctx context.Context, oldRefID, newAcID, newRefID string) error {
	acTTL, refTTL := c.ttlForTokens()

	_, err := c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if oldRefID != "" {
			pipe.Del(ctx, oldRefID)
		}

		pipe.Set(ctx, newAcID, true, acTTL)
		pipe.Set(ctx, newRefID, true, refTTL)

		return nil
	})

	return err
}
