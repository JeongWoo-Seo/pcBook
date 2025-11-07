proto:
	rm ./pb/*.go || true
	protoc --proto_path=./proto \
       --go_out=paths=source_relative:./pb \
       --go-grpc_out=paths=source_relative:./pb \
       ./proto/*.proto

server:
	go run cmd/server/main.go -port 8080

server1:
	go run cmd/server/main.go -port 50051

server2:
	go run cmd/server/main.go -port 50052

server1-tls:
	go run cmd/server/main.go -port 50051 -tls

server2-tls:
	go run cmd/server/main.go -port 50052 -tls

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

client-tls:
	go run cmd/client/main.go -address 0.0.0.0:8080 -tls

test:
	go test -cover -race ./...

evans:
	evans -r repl -p 8080

cert:
	cd cert; ./gen.sh; cd ..

.PHONY: proto run test server client evans cert