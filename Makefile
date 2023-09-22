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

docker-publish:
	docker build -t grpcchatter:latest . -f build/Dockerfile
	docker tag grpcchatter:latest mateuszskowron21/grpcchatter:latest
	docker push mateuszskowron21/grpcchatter:latest

.PHONY: build run test lint clean format proto docker-publish
