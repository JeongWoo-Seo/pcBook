package redisutil

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:            "localhost:6379",
		DB:              0,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 1 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := waitRedisReady(ctx, rdb); err != nil {
		log.Fatalf("Redis connection fail: %v", err)
	}

	log.Println("Redis connected")
	return rdb
}

func PublishToRedis(ctx context.Context, rdb *redis.Client, laptop *pb.LaptopInfo) error {
	data, err := json.Marshal(laptop)
	if err != nil {
		return err
	}

	channel := "laptop:" + laptop.GetId() + ":metrics"
	retryTime := time.Second

	for {
		err := rdb.Publish(ctx, channel, data).Err()
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		log.Println("redis publish failed:", err)

		time.Sleep(withJitter(retryTime))

		retryTime *= 2
		if retryTime >= 10*time.Second {
			retryTime = 10 * time.Second
		}
	}
}

func waitRedisReady(ctx context.Context, rdb *redis.Client) error {
	retryTime := time.Second

	for {
		err := rdb.Ping(ctx).Err()
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		log.Println("wait redis ready:", err)

		time.Sleep(withJitter(retryTime))

		retryTime *= 2
		if retryTime >= 10*time.Second {
			retryTime = 10 * time.Second
		}
	}
}

func withJitter(d time.Duration) time.Duration {
	n := rand.Int63n(int64(d) / 2)
	return d + time.Duration(n)
}
