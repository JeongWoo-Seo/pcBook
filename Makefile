proto:
	rm ./pb/*.go || true
	protoc --proto_path=./proto \
       --go_out=paths=source_relative:./pb \
       --go-grpc_out=paths=source_relative:./pb \
       ./proto/*.proto

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

test:
	go test -cover -race ./...

.PHONY: proto run test