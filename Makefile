BINARY_NAME = grpcchatter
BUILD_DIR = bin
PROTO_DIR = proto
DOCKER_IMAGE = grpcchatter:latest
DOCKER_USERNAME = mateuszskowron21

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/grpcchatter/main.go

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -race -vet=off ./...

lint:
	staticcheck ./...
	golint ./...

clean:
	rm -rf $(BUILD_DIR)

format:
	go fmt ./...

proto:
	protoc \
		--go_out=./$(PROTO_DIR)/gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=require_unimplemented_servers=false:./$(PROTO_DIR)/gen \
		--go-grpc_opt=paths=source_relative \
		./$(PROTO_DIR)/grpcchatter.proto

docker-publish:
	docker build -t $(DOCKER_IMAGE) . -f build/Dockerfile
	docker tag $(DOCKER_IMAGE) $(DOCKER_USERNAME)/$(DOCKER_IMAGE)
	docker push $(DOCKER_USERNAME)/$(DOCKER_IMAGE)

.PHONY: build run test lint clean format proto docker-publish
