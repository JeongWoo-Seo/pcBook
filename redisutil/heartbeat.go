package redisutil

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func UpdateLaptopHeartbeat(ctx context.Context, rm *RedisManager, laptopID string) error {
	if err := rm.AllowRequest(); err != nil {
		return err
	}

	now := time.Now().Unix()

	err := rm.Client.ZAdd(ctx, "laptop:alive", redis.Z{
		Score:  float64(now),
		Member: laptopID,
	}).Err()

	if err != nil {
		rm.connectionFailure(err)
		return err
	}

	rm.connectionSuccess()
	return nil
}

func StartCleanup(ctx context.Context, rm *RedisManager, tick time.Duration, expire int64) {
	ticker := time.NewTicker(tick)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := rm.AllowRequest(); err != nil {
					continue
				}

				now := time.Now().Unix()

				err := rm.Client.ZRemRangeByScore( //min,max 값은 문자열로 입력하고 redis 내부에서는 float 형태로 처리됨
					ctx,
					"laptop:alive",
					"-inf",
					strconv.FormatInt(now-expire, 10), //10진법 값을 문자열로 변환
				).Err()

				if err != nil {
					rm.connectionFailure(err)
					log.Println("cleanup error:", err)
					continue
				}

				rm.connectionSuccess()

			case <-ctx.Done():
				return
			}
		}
	}()
}
