package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lazerion/outbox-relayer/internal/service"
	"github.com/redis/go-redis/v9"
)

type MessageCache interface {
	CacheMessage(ctx context.Context, messageID string, sentAt time.Time) error
	StartConsumer(ctx context.Context, cacheCh <-chan service.SentMessageEvent)
}

type RedisMessageCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisMessageCache(client *redis.Client, ttl time.Duration) MessageCache {
	return &RedisMessageCache{client: client, ttl: ttl}
}

func (r *RedisMessageCache) StartConsumer(ctx context.Context, cacheCh <-chan service.SentMessageEvent) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-cacheCh:
				if !ok {
					return
				}
				if err := r.CacheMessage(ctx, evt.MessageID, evt.SentAt); err != nil {
					log.Printf("failed to cache message %s: %v", evt.MessageID, err)
				}
			}
		}
	}()
}

func (r *RedisMessageCache) CacheMessage(ctx context.Context, messageID string, sentAt time.Time) error {
	key := fmt.Sprintf("message:%s", messageID)
	return r.client.Set(ctx, key, sentAt.Format(time.RFC3339), r.ttl).Err()
}
