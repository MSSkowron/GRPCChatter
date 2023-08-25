package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/MSSkowron/GRPCChatter/internal/services"
	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// DefaultPort is the default port the server listens on.
	DefaultPort = 5000
	// DefaultAddress is the default address the server listens on.
	DefaultAddress = ""
	// DefaultMaxMessageQueueSize is the default max size of the message queue that is used to store messages to be sent to clients.
	DefaultMaxMessageQueueSize = 255
	grpcHeaderTokenKey         = "token"
)

// GRPCChatterServer represents a GRPCChatter server.
type GRPCChatterServer struct {
	proto.UnimplementedGRPCChatterServer

	tokenService     services.TokenService
	shortCodeService services.ShortCodeService

	address             string
	port                int
	maxMessageQueueSize int

	mu    sync.RWMutex
	rooms map[string]*room
}

type room struct {
	shortCode string
	name      string
	password  string

	clients []*client
}

type client struct {
	name         string
	messageQueue chan message
}

type message struct {
	sender string
	body   string
}

// NewGRPCChatterServer creates a new GRPCChatter server.
func NewGRPCChatterServer(tokenService services.TokenService, shortCodeService services.ShortCodeService, opts ...Opt) *GRPCChatterServer {
	server := &GRPCChatterServer{
		tokenService:        tokenService,
		shortCodeService:    shortCodeService,
		address:             DefaultAddress,
		port:                DefaultPort,
		maxMessageQueueSize: DefaultMaxMessageQueueSize,
		rooms:               make(map[string]*room),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

// Opt represents an option that can be passed to NewGRPCChatterServer.
type Opt func(*GRPCChatterServer)

// WithAddress sets the address the server listens on.
func WithAddress(address string) Opt {
	return func(s *GRPCChatterServer) {
		s.address = address
	}
}

// WithPort sets the port the server listens on.
func WithPort(port int) Opt {
	return func(s *GRPCChatterServer) {
		s.port = port
	}
}

// WithMaxMessageQueueSize sets the max size of the message queue that is used to send messages to clients.
func WithMaxMessageQueueSize(size int) Opt {
	return func(s *GRPCChatterServer) {
		s.maxMessageQueueSize = size
	}
}

// ListenAndServe starts the server and listens for incoming connections.
func (s *GRPCChatterServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.address+":"+strconv.Itoa(s.port))
	if err != nil {
		return fmt.Errorf("failed to create tcp listener on %s:%d: %w", s.address, s.port, err)
	}

	logger.Info(fmt.Sprintf("Server started listening on %s:%d", s.address, s.port))

	grpcServer := grpc.NewServer()

	proto.RegisterGRPCChatterServer(grpcServer, s)

	if err := grpcServer.Serve(ln); err != nil {
		return fmt.Errorf("failed to run grpc server on %s:%d: %w", s.address, s.port, err)
	}

	return nil
}

// CreateChatRoom is an RPC handler that creates a new chat room.
func (s *GRPCChatterServer) CreateChatRoom(ctx context.Context, req *proto.CreateChatRoomRequest) (*proto.CreateChatRoomResponse, error) {
	var (
		roomName     = req.GetRoomName()
		roomPassword = req.GetRoomPassword()
	)

	logger.Info(fmt.Sprintf("Received RPC CreateChatRoom request [{RoomName: %s, RoomPassword: %s}]", roomName, roomPassword))

	s.mu.Lock()
	defer s.mu.Unlock()

	roomShortCode := s.shortCodeService.GenerateShortCode(roomName)

	s.addRoom(roomShortCode, roomName, roomPassword)

	return &proto.CreateChatRoomResponse{
		ShortCode: string(roomShortCode),
	}, nil
}

// JoinChatRoom is an RPC handler that allows a user to join an existing chat room.
func (s *GRPCChatterServer) JoinChatRoom(ctx context.Context, req *proto.JoinChatRoomRequest) (*proto.JoinChatRoomResponse, error) {
	var (
		userName      = req.GetUserName()
		roomShortCode = req.GetShortCode()
		roomPassword  = req.GetRoomPassword()
	)

	logger.Info(fmt.Sprintf("Received RPC JoinChatRoom request [{UserName: %s, ShortCode: %s, RoomPassword: %s}]", userName, roomShortCode, roomPassword))

	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomShortCode]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
	}

	if roomPassword != room.password {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid room password for chat room with short code [%s]. Please make sure you have the correct password.", roomShortCode)
	}

	s.addClientToRoom(&client{
		name:         userName,
		messageQueue: make(chan message, s.maxMessageQueueSize),
	}, room)

	token, err := s.tokenService.GenerateToken(userName, roomShortCode)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error while generating token.")
	}

	return &proto.JoinChatRoomResponse{
		Token: token,
	}, nil
}

// Chat is a server-side streaming RPC handler that receives messages from clients and broadcasts them to all other clients.
func (s *GRPCChatterServer) Chat(chs proto.GRPCChatter_ChatServer) error {
	md, ok := metadata.FromIncomingContext(chs.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "Missing gRPC headers: %s. Please include your authentication token in the '%s' gRPC header.", grpcHeaderTokenKey, grpcHeaderTokenKey)
	}

	tokens := md.Get(grpcHeaderTokenKey)
	if len(tokens) == 0 {
		return status.Errorf(codes.Unauthenticated, "Authentication token missing in gRPC headers. Please include your token in the '%s' gRPC header.", grpcHeaderTokenKey)
	}
	userToken := tokens[0]

	if err := s.tokenService.ValidateToken(userToken); err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return status.Error(codes.Internal, "Internal server error while validating token.")
	}

	roomShortCode, err := s.tokenService.GetShortCodeFromToken(userToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return status.Error(codes.Internal, "Internal server error while retrieving short code from token.")
	}

	userName, err := s.tokenService.GetUserNameFromToken(userToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return status.Error(codes.Internal, "Internal server error while retrieving user name from token.")
	}

	logger.Info(fmt.Sprintf("Client [UserName: %s] established message stream with the chat room with short code [%s] using token [%s]", userName, roomShortCode, userToken))

	s.mu.RLock()
	room := s.rooms[roomShortCode]
	if room == nil {
		return status.Error(codes.NotFound, "Room not found. The requested chat room does not exist.")
	}

	var c *client
	for _, client := range room.clients {
		if client.name == userName {
			c = client
			break
		}
	}
	if c == nil {
		return status.Error(codes.PermissionDenied, "No permission to access this room. You do not have permission to participate in this chat room.")
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	receiveCh := make(chan struct{}, 1)
	sendCh := make(chan struct{}, 1)

	go s.receive(chs, c, room, sendCh, receiveCh, wg)
	go s.send(chs, c, room, receiveCh, sendCh, wg)
	s.mu.RUnlock()

	wg.Wait()

	close(receiveCh)
	close(sendCh)

	s.mu.Lock()
	s.removeClientFromRoom(c, room)
	s.mu.Unlock()

	return nil
}

func (s *GRPCChatterServer) receive(chs proto.GRPCChatter_ChatServer, c *client, r *room, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	roomShortCode := r.shortCode
	for {
		select {
		case <-receiveStopCh:
			return
		default:
			mssg, err := chs.Recv()
			if err != nil {
				if status.Code(err) == codes.Canceled {
					logger.Info(fmt.Sprintf("Client [UserName: %s] left the chat room with short code [%s]", c.name, roomShortCode))
				} else {
					logger.Error(fmt.Sprintf("Failed to receive message from client [UserName: %s] in chat room with short code [%s]: %s", c.name, roomShortCode, status.Convert(err).Message()))
				}

				sendStopCh <- struct{}{}

				return
			}

			msg := message{
				sender: c.name,
				body:   mssg.Body,
			}

			logger.Info(fmt.Sprintf("Received message [{Sender: %s, Body: %s}] from client [UserName: %s] in chat room with short code [%s]", msg.sender, msg.body, c.name, roomShortCode))

			s.mu.RLock()
			for _, client := range r.clients {
				if client.name != c.name {
					client.messageQueue <- msg
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *GRPCChatterServer) send(chs proto.GRPCChatter_ChatServer, c *client, r *room, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	roomShortCode := r.shortCode
	for {
		select {
		case <-receiveStopCh:
			return
		case msg := <-c.messageQueue:
			if err := chs.Send(&proto.ServerMessage{
				UserName: msg.sender,
				Body:     msg.body,
			}); err != nil {
				logger.Error(fmt.Sprintf("Failed to send message to client [UserName: %s] in chat room with short code [%s]: %s", c.name, roomShortCode, status.Convert(err).Message()))

				sendStopCh <- struct{}{}

				return
			}

			logger.Info(fmt.Sprintf("Sent message [{Sender: %s, Body: %s}] to client [UserName: %s] in chat room with short code [%s]", msg.sender, msg.body, c.name, roomShortCode))
		}
	}
}

// It should be called with the s.mu read-write mutex locked.
func (s *GRPCChatterServer) addRoom(shortCode string, name, password string) {
	s.rooms[shortCode] = &room{
		shortCode: shortCode,
		name:      name,
		password:  password,
		clients:   make([]*client, 0),
	}

	logger.Info(fmt.Sprintf("Created room [%s] with short code [%s] and password [%s]", name, shortCode, password))
}

// It should be called with the s.mu read-write mutex locked.
func (s *GRPCChatterServer) addClientToRoom(c *client, r *room) {
	r.clients = append(r.clients, c)

	logger.Info(fmt.Sprintf("Added client [UserName: %s] to chat room client's list with short code [%s]", c.name, r.shortCode))
}

// It should be called with the s.mu read-write mutex locked.
func (s *GRPCChatterServer) removeClientFromRoom(c *client, r *room) {
	for i, client := range r.clients {
		if client.name == c.name {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)

			logger.Info(fmt.Sprintf("Removed client [UserName: %s] from chat room client's list with short code [%s]", c.name, r.shortCode))

			return
		}
	}

	logger.Info(fmt.Sprintf("Client [UserName: %s] was not found in the chat room client's list with short code [%s] and was not removed", c.name, r.shortCode))
}
