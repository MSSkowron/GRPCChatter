FROM golang:1.21-alpine as builder

RUN adduser -D -g '' grpcchatter
WORKDIR /home/grpcchatter

COPY . .

RUN CGO_ENABLED=0 go build -a -tags netgo,usergo -ldflags "-extldflags '-static' -s -w" -o grpcchatter ./cmd/grpcchatter

FROM busybox:musl

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /home/grpcchatter/configs/default_config.env /home/configs/default_config.env
COPY --from=builder /home/grpcchatter/grpcchatter /home/grpcchatter

USER grpcchatter
WORKDIR /home

ENTRYPOINT [ "/home/grpcchatter" ]
