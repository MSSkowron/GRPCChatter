package client

import (
	"context"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: Synchronize goroutines
type Client struct {
	conn   *grpc.ClientConn
	name   string
	stream proto.GRPCChatter_ChatClient

	receiveQueue chan Message
	sendQueue    chan string
}

type Message struct {
	Sender string
	Body   string
}

func NewClient(name string, serverAddress string) (*Client, error) {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server %w", err)
	}

	stream, err := proto.NewGRPCChatterClient(conn).Chat(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create a stream with server: %w", err)
	}

	client := &Client{
		name:   name,
		conn:   conn,
		stream: stream,

		receiveQueue: make(chan Message),
		sendQueue:    make(chan string),
	}

	return client, nil
}

func (c *Client) Join() {
	go c.send()
	go c.receive()
}

func (c *Client) Send(message string) {
	c.sendQueue <- message
}

func (c *Client) Receive() Message {
	return <-c.receiveQueue
}

func (c *Client) send() {
	for {
		msg := <-c.sendQueue
		c.stream.Send(&proto.ClientMessage{
			Name: c.name,
			Body: msg,
		})
	}
}

func (c *Client) receive() {
	for {
		msg, err := c.stream.Recv()
		if err != nil {
			return
		}

		c.receiveQueue <- Message{
			Sender: msg.Name,
			Body:   msg.Body,
		}
	}

}

func (c *Client) Close() {
	c.conn.Close()
}
