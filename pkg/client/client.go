package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrConnectionNotExists = errors.New("connection with the server does not exist")
	ErrStreamNotExists     = errors.New("stream with the server does not exist")
)

// Client represents a chat client.
type Client struct {
	name          string
	serverAddress string
	conn          *grpc.ClientConn
	stream        proto.GRPCChatter_ChatClient
	receiveQueue  chan Message
	sendQueue     chan string
	closeCh       chan struct{}
	wg            sync.WaitGroup
}

// Message represents a chat message.
type Message struct {
	Sender string
	Body   string
}

// NewClient creates a new chat client.
func NewClient(name string, serverAddress string) (*Client, error) {
	client := &Client{
		name:          name,
		serverAddress: serverAddress,
	}

	return client, nil
}

// Join connects to the server, starts receiving messages, and enables sending messages.
func (c *Client) Join() error {
	conn, err := grpc.Dial(c.serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	c.conn = conn

	stream, err := proto.NewGRPCChatterClient(c.conn).Chat(context.Background())
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to create a stream with server: %w", err)
	}
	c.stream = stream

	c.receiveQueue = make(chan Message)
	c.sendQueue = make(chan string)
	c.closeCh = make(chan struct{})

	c.wg.Add(2)
	go c.send()
	go c.receive()

	return nil
}

// Send sends a message to the server.
// It blocks until the message is sent or returns immediately when the stream is closed and returns an empty message.
// Join() should be called before the first usage.
func (c *Client) Send(message string) error {
	if c.conn == nil {
		return ErrConnectionNotExists
	}

	if c.stream == nil {
		return ErrStreamNotExists
	}

	select {
	case c.sendQueue <- message:
	case <-c.closeCh:
	}

	return nil
}

// Receive receives a message from the server.
// It blocks until a message comes in and returns the incoming message or returns immediately when the stream is closed and returns an empty message.
// Join() should be called before the first usage.
func (c *Client) Receive() (Message, error) {
	if c.conn == nil {
		return Message{}, ErrConnectionNotExists
	}

	if c.stream == nil {
		return Message{}, ErrStreamNotExists
	}

	select {
	case msg := <-c.receiveQueue:
		return msg, nil
	case <-c.closeCh:
		return Message{}, nil
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
				c.Close()
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

// Close gracefully terminates the client, closing the connection with the server and cleaning up associated resources.
func (c *Client) Close() {
	if _, ok := <-c.closeCh; ok {
		close(c.closeCh)
		c.wg.Wait()
	}

	if c.conn != nil {
		c.conn.Close()
	}

	c.conn = nil
	c.stream = nil
}
