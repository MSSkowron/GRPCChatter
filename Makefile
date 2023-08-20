BINARY_NAME=grpcchatter

build:
	@go build -o bin/$(BINARY_NAME) ./cmd/grpcchatter/main.go

run: build
	@./bin/$(BINARY_NAME)

test: 
	@go test -race -vet=off ./...

lint:
	staticcheck ./...
	golint ./...

clean:
	@rm -rf bin

format:
	@go fmt ./...

proto:
	protoc --go_out=./proto/gen --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:./proto/gen --go-grpc_opt=paths=source_relative ./proto/grpcchatter.proto

.PHONY: build run test lint clean format
