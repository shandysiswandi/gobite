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

func (c *Redis) SaveTokensID(ctx context.Context, acID, refID string) error {
	acTTL := time.Duration(c.config.GetInt("jwt.access.ttl")) * time.Minute          // 15 minutes
	refTTL := time.Duration(c.config.GetInt("jwt.refresh.ttl")) * 7 * 24 * time.Hour // 7 days

	_, err := c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, acID, true, acTTL)
		pipe.Set(ctx, refID, true, refTTL)
		return nil
	})

	return err
}
