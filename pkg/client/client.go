package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/MSSkowron/GRPCChatter/internal/dto"
	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	// ErrConnectionNotExist is returned when attempting an operation on a non-existent server connection.
	ErrConnectionNotExist = errors.New("server connection does not exist")
	// ErrStreamNotExist is returned when attempting an operation on a non-existent server stream.
	ErrStreamNotExist = errors.New("server stream does not exist")
	// ErrAlreadyJoinedChatRoom is returned when a client attempts to join a chat room when they have already joined one.
	ErrAlreadyJoinedChatRoom = errors.New("client is already in a chat room")
	// ErrConnectionClosed is returned when the connection with the server has been closed.
	ErrConnectionClosed = errors.New("server connection has been closed")
	// ErrNotJoinedChatRoom is returned when a client has not joined any chat room.
	ErrNotJoinedChatRoom = errors.New("client has not joined any chat room")
	// ErrNotLoggedIn is returned when a client is not logged in.
	ErrNotLoggedIn = errors.New("client is not logged in")
)

// Client represents a chat client.
type Client struct {
	restServerAddress string
	grpcServerAddress string

	mu         sync.RWMutex
	conn       *grpc.ClientConn
	grpcClient proto.GRPCChatterClient
	stream     proto.GRPCChatter_ChatClient
	chatToken  string
	authToken  string

	receiveQueue chan Message
	sendQueue    chan string

	closeCh chan struct{}
	wg      sync.WaitGroup
}

// Message represents an incoming chat message.
type Message struct {
	Sender string // Sender is the name of the user who sent the message.
	Body   string // Body contains the content of the chat message.
}

// NewClient creates a new chat client with the given name and server address.
func NewClient(restServerAddress, grpcServerAddres string) *Client {
	return &Client{
		restServerAddress: restServerAddress,
		grpcServerAddress: grpcServerAddres,
	}
}

// Register registers the user with the server.
func (c *Client) Register(username, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := dto.UserRegisterDTO{
		Username: username,
		Password: password,
	}

	resp, err := c.postJSON(fmt.Sprintf("%s/register", c.restServerAddress), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// Login logs in the user with the server.
func (c *Client) Login(username, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := dto.UserLoginDTO{
		Username: username,
		Password: password,
	}

	resp, err := c.postJSON(fmt.Sprintf("%s/login", c.restServerAddress), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	respBody := dto.TokenDTO{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.authToken = respBody.Token

	return nil
}

// CreateChatRoom creates a new chat room with the provided name and password.
// Upon successful creation, it returns the shortcode of the newly created chat room.
// The Login() method must be called before the first usage while it requires authorization token.
func (c *Client) CreateChatRoom(roomName, roomPassword string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.authToken == "" {
		return "", ErrNotLoggedIn
	}

	if c.conn == nil {
		if err := c.connect(); err != nil {
			return "", err
		}
	}

	md := metadata.New(map[string]string{
		"token": c.authToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := c.grpcClient.CreateChatRoom(ctx, &proto.CreateChatRoomRequest{
		RoomName:     roomName,
		RoomPassword: roomPassword,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat room: %w", err)
	}

	return resp.GetShortCode(), nil
}

// JoinChatRoom connects the client to a specific chat room.
// Returns ErrAlreadyJoined if the client is already connected to a chat room. To leave the current chat room, use the Disconnect method.
// The Login() method must be called before the first usage.
func (c *Client) JoinChatRoom(shortCode string, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stream != nil {
		return ErrAlreadyJoinedChatRoom
	}

	if c.authToken == "" {
		return ErrNotLoggedIn
	}

	if c.conn == nil {
		if err := c.connect(); err != nil {
			return err
		}
	}

	md := metadata.New(map[string]string{
		"token": c.authToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := c.grpcClient.JoinChatRoom(ctx, &proto.JoinChatRoomRequest{
		ShortCode:    shortCode,
		RoomPassword: password,
	})
	if err != nil {
		return fmt.Errorf("failed to join the chat room: %w", err)
	}

	c.chatToken = resp.GetToken()

	md = metadata.New(map[string]string{
		"token": c.chatToken,
	})
	ctx = metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.grpcClient.Chat(ctx)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to establish a chat stream: %w", err)
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

// ListChatRoomUsers retrieves the list of users in the currently joined chat room.
// The JoinChatRoom() method must be called before the first usage.
func (c *Client) ListChatRoomUsers() ([]string, error) {
	c.mu.RLock()
	if c.chatToken == "" {
		c.mu.RUnlock()
		return nil, ErrNotJoinedChatRoom
	}

	if c.stream == nil {
		c.mu.RUnlock()
		return nil, ErrStreamNotExist
	}

	if c.conn == nil {
		c.mu.RUnlock()
		return nil, ErrConnectionNotExist
	}
	c.mu.RUnlock()

	md := metadata.New(map[string]string{
		"token": c.chatToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := c.grpcClient.ListChatRoomUsers(ctx, &proto.ListChatRoomUsersRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list the chat room users: %w", err)
	}

	users := make([]string, len(resp.Users))
	for _, user := range resp.GetUsers() {
		users = append(users, user.GetUserName())
	}

	return users, nil
}

// Send sends a message to the server.
// It blocks until the message is sent or returns immediately when an error occured.
// The JoinChatRoom() method must be called before the first usage.
func (c *Client) Send(message string) error {
	c.mu.RLock()
	if c.chatToken == "" {
		c.mu.RUnlock()
		return ErrNotJoinedChatRoom
	}

	if c.stream == nil {
		c.mu.RUnlock()
		return ErrStreamNotExist
	}

	if c.conn == nil {
		c.mu.RUnlock()
		return ErrConnectionNotExist
	}
	c.mu.RUnlock()

	select {
	case c.sendQueue <- message:
	case <-c.closeCh:
		return ErrConnectionClosed
	}

	return nil
}

// Receive receives a message from the server.
// It blocks until a message arrives or returns immediately when an error occured.
// The JoinChatRoom() method must be called before the first usage.
func (c *Client) Receive() (Message, error) {
	c.mu.RLock()
	if c.chatToken == "" {
		c.mu.RUnlock()
		return Message{}, ErrNotJoinedChatRoom
	}

	if c.stream == nil {
		c.mu.RUnlock()
		return Message{}, ErrStreamNotExist
	}

	if c.conn == nil {
		c.mu.RUnlock()
		return Message{}, ErrConnectionNotExist
	}
	c.mu.RUnlock()

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
			c.mu.RLock()
			if err := c.stream.Send(&proto.ClientMessage{Body: msg}); err != nil {
				c.mu.RUnlock()
				c.close()
				return
			}
			c.mu.RUnlock()
		case <-c.closeCh:
			return
		}
	}
}

func (c *Client) receive() {
	defer c.wg.Done()
	for {
		c.mu.RLock()
		msg, err := c.stream.Recv()
		if err != nil {
			c.mu.RUnlock()
			c.close()
			return
		}
		c.mu.RUnlock()

		select {
		case c.receiveQueue <- Message{
			Sender: msg.UserName,
			Body:   msg.Body,
		}:
		case <-c.closeCh:
			return
		}
	}
}

// It should be called with the c.mu read-write mutex locked.
func (c *Client) connect() error {
	conn, err := grpc.Dial(c.grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server at %s: %w", c.grpcServerAddress, err)
	}

	c.conn = conn
	c.grpcClient = proto.NewGRPCChatterClient(conn)

	return nil
}

// Disconnect disconnects the client from the server, closing the connection with the server.
func (c *Client) Disconnect() {
	c.close()
	c.wg.Wait()
}

// It should be called without the c.mu read-write mutex locked.
func (c *Client) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		c.clear()
		return
	}

	select {
	case <-c.closeCh:
	default:
		close(c.closeCh)
	}

	c.conn.Close()

	c.clear()
}

func (c *Client) clear() {
	c.conn, c.grpcClient, c.stream, c.authToken, c.chatToken = nil, nil, nil, "", ""
}

func (c *Client) postJSON(url string, data any) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

func (c *Client) handleErrorResponse(resp *http.Response) error {
	respErr := dto.ErrorDTO{}
	if err := json.NewDecoder(resp.Body).Decode(&respErr); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return fmt.Errorf("status: %s error: %s", resp.Status, respErr.Error)
}
