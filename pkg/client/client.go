package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	// ErrConnectionNotExists is returned when attempting an operation on a non-existent server connection.
	ErrConnectionNotExists = errors.New("connection with the server does not exist")
	// ErrStreamNotExists is returned when attempting an operation on a non-existent server stream.
	ErrStreamNotExists = errors.New("stream with the server does not exist")
	// ErrAlreadyJoined is returned when a client attempts to join the chat server more than once.
	ErrAlreadyJoined = errors.New("already joined")
	// ErrConnectionClosed is returned when a connection with the server has been closed.
	ErrConnectionClosed = errors.New("connection closed")
)

// Client represents a chat client.
type Client struct {
	name          string
	serverAddress string

	conn       *grpc.ClientConn
	grpcClient proto.GRPCChatterClient
	stream     proto.GRPCChatter_ChatClient

	receiveQueue chan Message
	sendQueue    chan string

	closeCh chan struct{}
	wg      sync.WaitGroup
}

// Message represents an incoming chat message.
type Message struct {
	Sender string
	Body   string
}

// NewClient creates a new chat client.
func NewClient(name string, serverAddress string) *Client {
	return &Client{
		name:          name,
		serverAddress: serverAddress,
	}
}

func (c *Client) CreateChatRoom(roomName, roomPassword string) (string, error) {
	if c.conn == nil {
		if err := c.connect(); err != nil {
			return "", err
		}
	}

	resp, err := c.grpcClient.CreateChatRoom(context.Background(), &proto.CreateChatRoomRequest{
		RoomName:     roomName,
		RoomPassword: roomPassword,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat room: %w", err)
	}

	return resp.GetShortCode(), nil
}

// Join connects the client to the server, initializes message channels, and starts receiving and sending messages.
// Returns ErrAlreadyJoined when a connection with the server has already been established.
func (c *Client) JoinChatRoom(shortCode string, password string) error {
	if c.stream != nil {
		return ErrAlreadyJoined
	}

	if c.conn == nil {
		if err := c.connect(); err != nil {
			return err
		}
	}

	resp, err := c.grpcClient.JoinChatRoom(context.Background(), &proto.JoinChatRoomRequest{
		ShortCode:    shortCode,
		RoomPassword: password,
	})
	if err != nil {
		return fmt.Errorf("failed to join the chat room: %w", err)
	}

	md := metadata.New(map[string]string{
		"userName":  c.name,
		"shortCode": shortCode,
		"token":     resp.GetToken(),
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.grpcClient.Chat(ctx)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to create a stream with server at %s: %w", c.serverAddress, err)
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
	if c.conn == nil {
		return ErrConnectionNotExists
	}

	if c.stream == nil {
		return ErrStreamNotExists
	}

	select {
	case c.sendQueue <- message:
	case <-c.closeCh:
		return ErrConnectionClosed
	}

	return nil
}

// Receive receives a message from the server.
// It blocks until a message arrives or returns immediately when the stream is closed, returning an empty message.
// The Join() method must be called before the first usage.
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
		return Message{}, ErrConnectionClosed
	}
}

func (c *Client) send() {
	defer c.wg.Done()
	for {
		select {
		case msg := <-c.sendQueue:
			if err := c.stream.Send(&proto.ClientMessage{Body: msg}); err != nil {
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

		select {
		case c.receiveQueue <- Message{
			Sender: msg.Name,
			Body:   msg.Body,
		}:
		case <-c.closeCh:
			return
		}
	}
}

func (c *Client) connect() error {
	conn, err := grpc.Dial(c.serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server at %s: %w", c.serverAddress, err)
	}

	c.conn = conn
	c.grpcClient = proto.NewGRPCChatterClient(conn)

	return nil
}

func (c *Client) close() {
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
	c.grpcClient = nil
	c.stream = nil
}

// Disconnect gracefully disconnects the client from the server, closing the connection with the server and cleaning up associated resources.
func (c *Client) Disconnect() {
	c.close()
	c.wg.Wait()
}
