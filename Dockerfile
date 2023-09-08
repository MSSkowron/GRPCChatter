FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/grpcchatter cmd/grpcchatter/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/configs/default_config.env /app/configs/default_config.env
COPY --from=builder /bin/grpcchatter /bin/grpcchatter

CMD ["/bin/grpcchatter"]
