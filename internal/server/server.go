package server

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

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
)

// GRPCChatterServer represents a GRPCChatter server.
type GRPCChatterServer struct {
	proto.UnimplementedGRPCChatterServer

	address             string
	port                int
	maxMessageQueueSize int

	mu    sync.Mutex
	rooms map[shortCode]*room
}

type shortCode string

type room struct {
	shortCode shortCode
	name      string
	password  string
	clients   []*client
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
func NewGRPCChatterServer(opts ...Opt) *GRPCChatterServer {
	server := &GRPCChatterServer{
		address:             DefaultAddress,
		port:                DefaultPort,
		maxMessageQueueSize: DefaultMaxMessageQueueSize,
		rooms:               make(map[shortCode]*room),
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

// TODO: Implement
func (s *GRPCChatterServer) CreateChatRoom(ctx context.Context, req *proto.CreateChatRoomRequest) (*proto.CreateChatRoomResponse, error) {
	return nil, nil
}

// TODO: Implement
func (s *GRPCChatterServer) JoinChatRoom(ctx context.Context, req *proto.JoinChatRoomRequest) (*proto.JoinChatRoomResponse, error) {
	return nil, nil
}

// Chat is a server-side streaming RPC handler that receives messages from clients and broadcasts them to all other clients.
func (s *GRPCChatterServer) Chat(chs proto.GRPCChatter_ChatServer) error {
	md, ok := metadata.FromIncomingContext(chs.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "Missing headers")
	}

	shortCodes := md.Get("shortCode")
	if len(shortCodes) == 0 {
		return status.Errorf(codes.Unauthenticated, "Missing short code")
	}
	roomShortCode := shortCodes[0]

	// TODO: Validate shortCode

	tokens := md.Get("token")
	if len(tokens) == 0 {
		return status.Errorf(codes.Unauthenticated, "Missing token")
	}
	userToken := tokens[0]

	// TODO: Validate token

	userNames := md.Get("userName")
	if len(userNames) == 0 {
		return status.Errorf(codes.Unauthenticated, "Missing user name")
	}
	userName := userNames[0]

	// TODO: Validate userName

	// TODO: Check userName permssions to the room based on the shortCode and token

	room := s.rooms[shortCode(roomShortCode)]

	c := &client{
		name:         userName,
		messageQueue: make(chan message, s.maxMessageQueueSize),
	}

	s.addClientToRoom(c, room)

	logger.Info(fmt.Sprintf("Client [ID: %s] joined the chat room with code [%s] using token [%s]", c.name, roomShortCode, userToken))

	wg := &sync.WaitGroup{}
	wg.Add(2)

	receiveCh := make(chan struct{}, 1)
	sendCh := make(chan struct{}, 1)

	go s.receive(chs, c, room, sendCh, receiveCh, wg)
	go s.send(chs, c, receiveCh, sendCh, wg)

	wg.Wait()

	close(receiveCh)
	close(sendCh)

	s.removeClientFromRoom(c, room)

	return nil
}

func (s *GRPCChatterServer) receive(chs proto.GRPCChatter_ChatServer, c *client, r *room, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		default:
			mssg, err := chs.Recv()
			if err != nil {
				if status.Code(err) == codes.Canceled {
					logger.Info(fmt.Sprintf("Client [UserName: %s] left the chat", c.name))
				} else {
					logger.Error(fmt.Sprintf("Failed to receive message from client [UserName: %s]: %s", c.name, status.Convert(err).Message()))
				}

				sendStopCh <- struct{}{}

				return
			}

			msg := message{
				sender: c.name,
				body:   mssg.Body,
			}

			logger.Info(fmt.Sprintf("Received message: {Sender: %s; Body: %s} from client [UserName: %s]", msg.sender, msg.body, c.name))

			s.mu.Lock()
			for _, client := range r.clients {
				if client.name != c.name {
					client.messageQueue <- msg
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *GRPCChatterServer) send(chs proto.GRPCChatter_ChatServer, c *client, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		case msg := <-c.messageQueue:
			if err := chs.Send(&proto.ServerMessage{
				Name: msg.sender,
				Body: msg.body,
			}); err != nil {
				logger.Error(fmt.Sprintf("Failed to send message to client [UserName: %s]: %s", c.name, status.Convert(err).Message()))

				sendStopCh <- struct{}{}

				return
			}

			logger.Info(fmt.Sprintf("Sent message: {Sender: %s; Body: %s} to client [UserName: %s]", msg.sender, msg.body, c.name))
		}
	}
}

func (s *GRPCChatterServer) addClientToRoom(c *client, r *room) {
	r.clients = append(r.clients, c)

	logger.Debug(fmt.Sprintf("Added client [UserName: %s] to chat room client's list with code [%s]", c.name, r.shortCode))
}

func (s *GRPCChatterServer) removeClientFromRoom(c *client, r *room) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, client := range r.clients {
		if client.name == c.name {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)

			logger.Debug(fmt.Sprintf("Removed client [UserName: %s] from chat room client's list with code [%s]", c.name, r.shortCode))

			break
		}
	}

	logger.Debug(fmt.Sprintf("Client [UserName: %s] was not found in the chat room client's list with code [%s] and was not removed", c.name, r.shortCode))
}
