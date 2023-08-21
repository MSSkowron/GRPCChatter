package server

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultPort    = "5000"
	defaultAddress = ""
)

type GRPCChatterServer struct {
	proto.UnimplementedGRPCChatterServer

	address string
	port    string

	clients []*client
}

type client struct {
	id           int
	messageQueue chan message
}

type message struct {
	clientName string
	body       string
}

func NewGRPCChatterServer(opts ...ServerOpt) *GRPCChatterServer {
	server := &GRPCChatterServer{
		address: defaultAddress,
		port:    defaultPort,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func (s *GRPCChatterServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.address+":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to create tcp listener on %s:%s: %w", s.address, s.port, err)
	}

	log.Printf("GRPChatter Server is listening on %s:%s\n", s.address, s.port)

	grpcServer := grpc.NewServer()

	proto.RegisterGRPCChatterServer(grpcServer, s)

	if err := grpcServer.Serve(ln); err != nil {
		return fmt.Errorf("failed to run grpc server on %s:%s: %w", s.address, s.port, err)
	}

	return nil
}

func (s *GRPCChatterServer) Chat(chs proto.GRPCChatter_ChatServer) error {
	c := &client{
		id:           rand.Intn(1e6),
		messageQueue: make(chan message, 255),
	}

	log.Printf("Client [%d] joined the chat\n", c.id)

	s.clients = append(s.clients, c)

	errCh := make(chan error)

	go s.receive(chs, c, errCh)
	go s.send(chs, c, errCh)

	<-errCh

	s.removeClient(c.id)

	return nil
}

func (s *GRPCChatterServer) receive(chs proto.GRPCChatter_ChatServer, c *client, errCh chan<- error) {
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
			clientName: mssg.Name,
			body:       mssg.Body,
		}

		log.Printf("Received a new message: {Sender: %s; Body: %s} from client [%d]\n", msg.clientName, msg.body, c.id)

		for _, client := range s.clients {
			if client.id != c.id {
				client.messageQueue <- msg
			}
		}
	}
}

func (s *GRPCChatterServer) send(chs proto.GRPCChatter_ChatServer, c *client, errCh chan<- error) {
	for {
		msg := <-c.messageQueue
		if err := chs.Send(&proto.ServerMessage{
			Name: msg.clientName,
			Body: msg.body,
		}); err != nil {
			log.Printf("Failed to send message to client [%d]: %s", c.id, status.Convert(err).Message())
			errCh <- err
			return
		}
	}
}

func (s *GRPCChatterServer) removeClient(id int) {
	index := -1
	for i, c := range s.clients {
		if c.id == id {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	s.clients = append(s.clients[:index], s.clients[index+1:]...)
}

type ServerOpt func(*GRPCChatterServer)

func WithAddress(address string) ServerOpt {
	return func(s *GRPCChatterServer) {
		s.address = address
	}
}

func WithPort(port string) ServerOpt {
	return func(s *GRPCChatterServer) {
		s.port = port
	}
}
