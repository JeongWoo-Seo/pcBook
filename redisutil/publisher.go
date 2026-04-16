package redisutil

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/redis/go-redis/v9"
)

var ErrRedisOpenCircuit = errors.New("redis circuit open")

type RedisManager struct {
	Client  *redis.Client
	healthy atomic.Bool

	mu sync.Mutex

	failCnt   int //현재 실패 횟수
	threshold int //circuit break 발생 실패 횟수

	openUntilTime time.Time
	openDuration  time.Duration

	halfOpen bool //openDuration 이후 redis 연결 test
}

func NewRedisManager() *RedisManager {
	rdb := redis.NewClient(&redis.Options{
		Addr:            "localhost:6379",
		DB:              0,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 1 * time.Second,
	})

	rm := &RedisManager{
		Client:       rdb,
		threshold:    3,
		openDuration: 10 * time.Second,
	}

	rm.healthy.Store(true)

	return rm
}

// redis 연결이 정상일 때는 ping 테스트 하지 않음
func (rm *RedisManager) StartRedisMonitor(ctx context.Context, interval time.Duration) {
	go func() {
		tricker := time.NewTicker(interval)
		defer tricker.Stop()

		for {
			select {
			case <-tricker.C:
				if rm.IsCircuitOpen() {
					continue
				}

				err := rm.Client.Ping(ctx).Err()
				if err != nil {
					rm.connectionFailure(err)
				} else {
					rm.connectionSuccess()
				}

			case <-ctx.Done():
				return
			}
		}
	}()

}

func (rm *RedisManager) AllowRequest() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	if now.Before(rm.openUntilTime) { //open 상태
		return ErrRedisOpenCircuit
	}

	if !rm.openUntilTime.IsZero() && !rm.halfOpen {
		rm.halfOpen = true
	}

	return nil
}

func (rm *RedisManager) IsCircuitOpen() bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	return time.Now().Before(rm.openUntilTime)
}

func (rm *RedisManager) connectionSuccess() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.halfOpen || rm.failCnt > 0 || !rm.healthy.Load() {
		log.Println("redis recovered")
	}

	rm.failCnt = 0
	rm.halfOpen = false
	rm.openUntilTime = time.Time{}
	rm.healthy.Store(true)
}

func (rm *RedisManager) connectionFailure(err error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.failCnt++
	rm.healthy.Store(false)

	if rm.failCnt >= rm.threshold {
		rm.openUntilTime = time.Now().Add(rm.openDuration)
		rm.halfOpen = false

		log.Println("redis circuit open:", err)
		return
	}
	log.Println("reids fail:", err)
}

func PublishToRedis(ctx context.Context, rm *RedisManager, laptop *pb.LaptopInfo) error {
	if err := rm.AllowRequest(); err != nil {
		return err
	}

	data, err := json.Marshal(laptop)
	if err != nil {
		return err
	}

	channel := "laptop:" + laptop.GetId() + ":metrics"

	err = rm.Client.Publish(ctx, channel, data).Err()
	if err != nil {
		rm.connectionFailure(err)
		return err
	}

	rm.connectionSuccess()
	return nil

}
