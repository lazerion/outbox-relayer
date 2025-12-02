package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/lazerion/outbox-relayer/internal/cache"
	"github.com/lazerion/outbox-relayer/internal/service"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   0,
	})

	return s, rdb
}

func TestCacheMessage(t *testing.T) {
	s, rdb := newTestRedis(t)
	defer s.Close()

	mc := cache.NewRedisMessageCache(rdb, 5*time.Minute)

	ctx := context.Background()
	sentAt := time.Now().UTC()
	err := mc.CacheMessage(ctx, "abc123", sentAt)
	if err != nil {
		t.Fatalf("CacheMessage failed: %v", err)
	}

	v, err := rdb.Get(ctx, "message:abc123").Result()
	if err != nil {
		t.Fatalf("redis GET failed: %v", err)
	}

	if v != sentAt.Format(time.RFC3339) {
		t.Fatalf("unexpected redis value: got %s want %s", v, sentAt.Format(time.RFC3339))
	}
}

func TestStartConsumer_ConsumesEvents(t *testing.T) {
	s, rdb := newTestRedis(t)
	defer s.Close()

	mc := cache.NewRedisMessageCache(rdb, 5*time.Minute)

	cacheCh := make(chan service.SentMessageEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mc.StartConsumer(ctx, cacheCh)

	sentAt := time.Now().UTC()
	cacheCh <- service.SentMessageEvent{
		MessageID: "xyz789",
		SentAt:    sentAt,
	}

	time.Sleep(50 * time.Millisecond)

	v, err := rdb.Get(ctx, "message:xyz789").Result()
	if err != nil {
		t.Fatalf("redis GET failed: %v", err)
	}

	if v != sentAt.Format(time.RFC3339) {
		t.Fatalf("unexpected redis value: got %s want %s", v, sentAt.Format(time.RFC3339))
	}
}

func TestStartConsumer_StopsOnContextCancel(t *testing.T) {
	s, rdb := newTestRedis(t)
	defer s.Close()

	mc := cache.NewRedisMessageCache(rdb, 5*time.Minute)

	cacheCh := make(chan service.SentMessageEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())

	mc.StartConsumer(ctx, cacheCh)

	cancel()

	cacheCh <- service.SentMessageEvent{
		MessageID: "should_not_write",
		SentAt:    time.Now().UTC(),
	}

	time.Sleep(50 * time.Millisecond)

	_, err := rdb.Get(context.Background(), "message:should_not_write").Result()
	if err == nil {
		t.Fatalf("expected no entry to be written after context cancel")
	}
}

func TestStartConsumer_StopsOnChannelClose(t *testing.T) {
	s, rdb := newTestRedis(t)
	defer s.Close()

	mc := cache.NewRedisMessageCache(rdb, 5*time.Minute)

	cacheCh := make(chan service.SentMessageEvent)
	ctx := context.Background()

	mc.StartConsumer(ctx, cacheCh)

	close(cacheCh)

	time.Sleep(50 * time.Millisecond)

	_, err := rdb.Get(ctx, "anything").Result()
	if err == nil {
		t.Fatalf("expected no writes after channel close")
	}
}
