package redisutil

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const laptopTTL = 10 * time.Second

func UpdateLaptopHeartbeat(
	ctx context.Context,
	rdb *redis.Client,
	laptopID string,
) error {

	key := "laptop:alive:" + laptopID

	return rdb.Set(ctx, key, "alive", laptopTTL).Err()
}
