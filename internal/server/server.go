package server

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DefaultPort                = 5000
	DefaultAddress             = ""
	DefaultMaxMessageQueueSize = 255
)

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

func NewGRPCChatterServer(opts ...ServerOpt) *GRPCChatterServer {
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

type ServerOpt func(*GRPCChatterServer)

func WithAddress(address string) ServerOpt {
	return func(s *GRPCChatterServer) {
		s.address = address
	}
}

func WithPort(port int) ServerOpt {
	return func(s *GRPCChatterServer) {
		s.port = port
	}
}

func WithMaxMessageQueueSize(size int) ServerOpt {
	return func(s *GRPCChatterServer) {
		s.maxMessageQueueSize = size
	}
}

func (s *GRPCChatterServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.address+":"+strconv.Itoa(s.port))
	if err != nil {
		return fmt.Errorf("failed to create tcp listener on %s:%d: %w", s.address, s.port, err)
	}

	log.Printf("GRPChatter Server is listening on %s:%d\n", s.address, s.port)

	grpcServer := grpc.NewServer()

	proto.RegisterGRPCChatterServer(grpcServer, s)

	if err := grpcServer.Serve(ln); err != nil {
		return fmt.Errorf("failed to run grpc server on %s:%d: %w", s.address, s.port, err)
	}

	return nil
}

func (s *GRPCChatterServer) Chat(chs proto.GRPCChatter_ChatServer) error {
	c := &client{
		id:           rand.Intn(1e6),
		messageQueue: make(chan message, s.maxMessageQueueSize),
	}

	log.Printf("Client [%d] joined the chat\n", c.id)

	s.addClient(c)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	errCh := make(chan error)
	go s.receive(chs, c, errCh, wg)
	go s.send(chs, c, errCh, wg)

	wg.Wait()

	s.removeClient(c.id)

	return nil
}

func (s *GRPCChatterServer) receive(chs proto.GRPCChatter_ChatServer, c *client, errCh chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		mssg, err := chs.Recv()
		if err != nil {
			if status.Code(err) == codes.Canceled {
				log.Printf("Client [%d] left the chat\n", c.id)
			} else {
				log.Printf("Failed to receive message from client %d: %s\n", c.id, status.Convert(err).Message())
			}

			errCh <- err
			return
		}

		msg := message{
			sender: mssg.Name,
			body:   mssg.Body,
		}

		log.Printf("Received a new message: {Sender: %s; Body: %s} from client [%d]\n", msg.sender, msg.body, c.id)

		s.mu.Lock()
		for _, client := range s.clients {
			if client.id != c.id {
				select {
				case client.messageQueue <- msg:
				case <-errCh:
					s.mu.Unlock()
					return
				}
			}
		}
		s.mu.Unlock()
	}
}

func (s *GRPCChatterServer) send(chs proto.GRPCChatter_ChatServer, c *client, errCh chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case msg := <-c.messageQueue:
			if err := chs.Send(&proto.ServerMessage{
				Name: msg.sender,
				Body: msg.body,
			}); err != nil {
				log.Printf("Failed to send message to client [%d]: %s", c.id, status.Convert(err).Message())
				errCh <- err
				return
			}
		case <-errCh:
			return
		}
	}
}

func (s *GRPCChatterServer) addClient(c *client) {
	s.mu.Lock()
	s.clients = append(s.clients, c)
	s.mu.Unlock()
}

func (s *GRPCChatterServer) removeClient(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, c := range s.clients {
		if c.id == id {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
}
