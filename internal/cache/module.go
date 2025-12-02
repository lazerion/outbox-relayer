package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/lazerion/outbox-relayer/internal/config"
	"github.com/lazerion/outbox-relayer/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	if cfg.Redis.Host == "" {
		return nil, fmt.Errorf("redis host is empty")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return rdb, nil
}

func NewMessageCacheProvider(redis *redis.Client, cfg *config.Config) MessageCache {
	return NewRedisMessageCache(redis, cfg.Redis.TTL)
}

func StartCacheConsumer(lc fx.Lifecycle, cacheCh chan service.SentMessageEvent, cache MessageCache) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting Redis message cache consumer...")
			cache.StartConsumer(ctx, cacheCh)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			close(cacheCh)
			return nil
		},
	})
}

var Module = fx.Module(
	"redis",
	fx.Provide(
		NewRedisClient,
		NewMessageCacheProvider,
	),
	fx.Invoke(StartCacheConsumer),
)
