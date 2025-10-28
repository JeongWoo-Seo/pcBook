proto:
	rm ./pb/*.go || true
	protoc --proto_path=./proto \
       --go_out=paths=source_relative:./pb \
       --go-grpc_out=paths=source_relative:./pb \
       ./proto/*.proto

run:
	go run main.go

test:
	go test -cover -race ./...

.PHONY: proto run test