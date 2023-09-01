FROM golang:1.21-alpine 

USER root 

WORKDIR /app 

COPY go.mod ./
COPY go.sum ./

COPY . ./ 

RUN go build -o /bin/grpcchatter cmd/grpcchatter/main.go

CMD ["/bin/grpcchatter"]
