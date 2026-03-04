package redisutil

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func UpdateLaptopHeartbeat(ctx context.Context, rdb *redis.Client, laptopID string) error {
	now := time.Now().Unix()

	err := rdb.ZAdd(ctx, "laptop:alive", redis.Z{
		Score:  float64(now),
		Member: laptopID,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

func StartCleanup(ctx context.Context, rdb *redis.Client, tick time.Duration, expire int64) {
	ticker := time.NewTicker(tick)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now().Unix()

				err := rdb.ZRemRangeByScore( //min,max 값은 문자열로 입력하고 redis 내부에서는 float 형태로 처리됨
					ctx,
					"laptop:alive",
					"-inf",
					strconv.FormatInt(now-expire, 10), //10진법 값을 문자열로 변환
				).Err()

				if err != nil {
					log.Println("cleanup error:", err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
