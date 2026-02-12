package redisutil

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
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
	return rdb.Publish(ctx, channel, data).Err()
}
