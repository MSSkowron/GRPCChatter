package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn         *grpc.ClientConn
	name         string
	stream       proto.GRPCChatter_ChatClient
	receiveQueue chan Message
	sendQueue    chan string
	closeOnce    sync.Once
	closeCh      chan struct{}
	wg           sync.WaitGroup
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
		name:         name,
		conn:         conn,
		stream:       stream,
		receiveQueue: make(chan Message),
		sendQueue:    make(chan string),
		closeCh:      make(chan struct{}),
	}

	return client, nil
}

func (c *Client) Join() {
	c.wg.Add(2)
	go c.send()
	go c.receive()
}

func (c *Client) Send(message string) {
	select {
	case c.sendQueue <- message:
	case <-c.closeCh:
	}
}

func (c *Client) Receive() Message {
	select {
	case msg := <-c.receiveQueue:
		return msg
	case <-c.closeCh:
		return Message{}
	}
}

func (c *Client) send() {
	defer c.wg.Done()
	for {
		select {
		case msg := <-c.sendQueue:
			if err := c.stream.Send(&proto.ClientMessage{
				Name: c.name,
				Body: msg,
			}); err != nil {
				return
			}
		case <-c.closeCh:
			return
		}
	}
}

func (c *Client) receive() {
	defer c.wg.Done()
	for {
		msg, err := c.stream.Recv()
		if err != nil {
			c.Close()
			return
		}
		c.receiveQueue <- Message{
			Sender: msg.Name,
			Body:   msg.Body,
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closeCh)
		c.conn.Close()
		c.wg.Wait()
	})
}
