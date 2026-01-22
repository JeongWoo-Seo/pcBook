package redisutil

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/redis/go-redis/v9"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to connect redis: %v", err)
	}

	return rdb
}

func TestRedisPublishSubscribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := newTestRedisClient(t)
	defer rdb.Close()

	channel := "laptop-updates"
	sub := rdb.Subscribe(ctx, channel)
	defer sub.Close()

	// 구독 준비 대기
	_, err := sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	msgCh := sub.Channel()

	laptop := &pb.LaptopInfo{
		Id: "laptop",
	}

	go func() {
		err := PublishToRedis(ctx, rdb, laptop)
		if err != nil {
			t.Errorf("publish failed: %v", err)
		}
	}()

	select {
	case msg := <-msgCh:
		var received pb.LaptopInfo
		err := json.Unmarshal([]byte(msg.Payload), &received)
		if err != nil {
			t.Fatalf("json unmarshal failed: %v", err)
		}

		if received.Id != laptop.Id {
			t.Fatalf("unexpected laptop id: %s", received.Id)
		}

	case <-ctx.Done():
		t.Fatal("timeout waiting redis message")
	}
}
