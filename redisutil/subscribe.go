package redisutil

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 테스트용 모든 오리진 허용
}

func handleWebSockets(rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// ⭐️ 3. Redis Subscribe 설정
	pubsub := rdb.Subscribe(context.Background(), "laptop-updates")
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Println("Web client connected and subscribing to Redis...")

	// ⭐️ 4. Redis 메시지를 기다렸다가 웹소켓으로 전달하는 루프
	for msg := range ch {
		// Redis에서 받은 메시지를 그대로 웹소켓으로 전송
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		if err != nil {
			log.Printf("WS write error: %v", err)
			break
		}
	}
}

func StartRedisSubTest(ctx context.Context, rdb *redis.Client) {
	sub := rdb.Subscribe(ctx, "laptop-updates")

	go func() {
		ch := sub.Channel()

		log.Println("[Redis-TEST] subscribed to laptop-updates")

		for msg := range ch {
			log.Printf("[Redis-TEST] received message: %s\n", msg.Payload)
		}
	}()
}
