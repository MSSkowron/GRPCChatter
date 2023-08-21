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
	// ErrConnectionNotExists is returned when attempting an operation on a non-existent server connection.
	ErrConnectionNotExists = errors.New("connection with the server does not exist")
	// ErrStreamNotExists is returned when attempting an operation on a non-existent server stream.
	ErrStreamNotExists = errors.New("stream with the server does not exist")
	// ErrAlreadyJoined is returned when a client attempts to join the chat server more than once.
	ErrAlreadyJoined = errors.New("already joined")
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
	mu            sync.Mutex
}

// Message represents a chat message.
type Message struct {
	Sender string // The sender's name.
	Body   string // The message content.
}

// NewClient creates a new chat client.
func NewClient(name string, serverAddress string) *Client {
	return &Client{
		name:          name,
		serverAddress: serverAddress,
	}
}

// Join connects the client to the server, initializes message channels, and starts receiving and sending messages.
// Returns ErrAlreadyJoined when a connection with the server has already been established.
func (c *Client) Join() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil || c.stream != nil {
		return ErrAlreadyJoined
	}

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
// It blocks until the message is sent or returns immediately when the stream is closed, and the message is discarded.
// The Join() method must be called before the first usage.
func (c *Client) Send(message string) error {
	c.mu.Lock()
	if c.conn == nil {
		return ErrConnectionNotExists
	}

	if c.stream == nil {
		return ErrStreamNotExists
	}
	c.mu.Unlock()

	select {
	case c.sendQueue <- message:
	case <-c.closeCh:
	}

	return nil
}

// Receive receives a message from the server.
// It blocks until a message arrives or returns immediately when the stream is closed, returning an empty message.
// The Join() method must be called before the first usage.
func (c *Client) Receive() (Message, error) {
	c.mu.Lock()
	if c.conn == nil {
		return Message{}, ErrConnectionNotExists
	}

	if c.stream == nil {
		return Message{}, ErrStreamNotExists
	}
	c.mu.Unlock()

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
				c.close()
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
			c.close()
			return
		}
		c.receiveQueue <- Message{
			Sender: msg.Name,
			Body:   msg.Body,
		}
	}
}

func (c *Client) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.stream == nil {
		return
	}

	select {
	case <-c.closeCh:
	default:
		close(c.closeCh)
	}

	c.conn.Close()

	c.conn = nil
	c.stream = nil
}

// Close gracefully terminates the client, closing the connection with the server and cleaning up associated resources.
func (c *Client) Close() {
	c.close()
	c.wg.Wait()
}
