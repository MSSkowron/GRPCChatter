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

.PHONY: build run test lint clean
