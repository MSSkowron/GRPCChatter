package server

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

	mu      sync.Mutex
	clients []*client
}

type client struct {
	id           int
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

// Chat is a server-side streaming RPC handler that receives messages from clients and broadcasts them to all other clients.
func (s *GRPCChatterServer) Chat(chs proto.GRPCChatter_ChatServer) error {
	c := &client{
		id:           rand.Intn(1e6),
		messageQueue: make(chan message, s.maxMessageQueueSize),
	}

	logger.Info(fmt.Sprintf("Client [ID: %d] joined the chat", c.id))

	s.addClient(c)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	receiveCh := make(chan struct{}, 1)
	sendCh := make(chan struct{}, 1)

	go s.receive(chs, c, sendCh, receiveCh, wg)
	go s.send(chs, c, receiveCh, sendCh, wg)

	wg.Wait()

	close(receiveCh)
	close(sendCh)

	s.removeClient(c.id)

	return nil
}

func (s *GRPCChatterServer) receive(chs proto.GRPCChatter_ChatServer, c *client, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		default:
			mssg, err := chs.Recv()
			if err != nil {
				if status.Code(err) == codes.Canceled {
					logger.Info(fmt.Sprintf("Client [ID: %d] left the chat", c.id))
				} else {
					logger.Error(fmt.Sprintf("Failed to receive message from client [ID: %d]: %s", c.id, status.Convert(err).Message()))
				}

				sendStopCh <- struct{}{}

				return
			}

			msg := message{
				sender: mssg.Name,
				body:   mssg.Body,
			}

			logger.Info(fmt.Sprintf("Received message: {Sender: %s; Body: %s} from client [ID: %d]", msg.sender, msg.body, c.id))

			s.mu.Lock()
			for _, client := range s.clients {
				if client.id != c.id {
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
				logger.Error(fmt.Sprintf("Failed to send message to client [ID: %d]: %s", c.id, status.Convert(err).Message()))

				sendStopCh <- struct{}{}

				return
			}

			logger.Info(fmt.Sprintf("Sent message: {Sender: %s; Body: %s} to client [ID: %d]", msg.sender, msg.body, c.id))
		}
	}
}

func (s *GRPCChatterServer) addClient(c *client) {
	s.mu.Lock()
	s.clients = append(s.clients, c)
	s.mu.Unlock()

	logger.Debug(fmt.Sprintf("Added client [ID: %d] to server client's list", c.id))
}

func (s *GRPCChatterServer) removeClient(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, c := range s.clients {
		if c.id == id {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)

			logger.Debug(fmt.Sprintf("Removed client [ID: %d] from server client's list", id))

			break
		}
	}

	logger.Debug(fmt.Sprintf("Client [ID: %d] was not found in the server's client list and was not removed", id))
}
