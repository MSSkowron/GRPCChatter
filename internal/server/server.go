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
	DefaultAddress     = ""
	grpcHeaderTokenKey = "token"
)

// GRPCChatterServer represents a GRPCChatter server.
type GRPCChatterServer struct {
	proto.UnimplementedGRPCChatterServer

	tokenService     services.TokenService
	shortCodeService services.ShortCodeService
	roomService      services.RoomService

	address string
	port    int
}

// NewGRPCChatterServer creates a new GRPCChatter server.
func NewGRPCChatterServer(tokenService services.TokenService, shortCodeService services.ShortCodeService, roomService services.RoomService, opts ...Opt) *GRPCChatterServer {
	server := &GRPCChatterServer{
		tokenService:     tokenService,
		shortCodeService: shortCodeService,
		roomService:      roomService,
		address:          DefaultAddress,
		port:             DefaultPort,
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

// ListenAndServe starts the server and listens for incoming connections.
func (s *GRPCChatterServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.address+":"+strconv.Itoa(s.port))
	if err != nil {
		return fmt.Errorf("failed to create tcp listener on %s:%d: %w", s.address, s.port, err)
	}

	logger.Info(fmt.Sprintf("Server listening on %s:%d", s.address, s.port))

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

	roomShortCode := s.shortCodeService.GenerateShortCode(roomName)

	if err := s.roomService.CreateRoom(roomShortCode, roomName, roomPassword); err != nil {
		return nil, status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	logger.Info(fmt.Sprintf("Created room [%s] with short code [%s]", roomName, roomShortCode))

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

	if !s.roomService.RoomExists(roomShortCode) {
		return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
	}

	if err := s.roomService.CheckPassword(roomShortCode, roomPassword); err != nil {
		if errors.Is(err, services.ErrRoomDoesNotExist) {
			return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
		}
		if errors.Is(err, services.ErrInvalidPassword) {
			return nil, status.Errorf(codes.PermissionDenied, "Invalid room password for chat room with short code [%s]. Please make sure you have the correct password.", roomShortCode)
		}

		return nil, status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	if err := s.roomService.AddUserToRoom(roomShortCode, userName); err != nil {
		if errors.Is(err, services.ErrRoomDoesNotExist) {
			return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
		}
		if errors.Is(err, services.ErrUserAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "User with username [%s] already exists in the chat room with short code [%s].", userName, roomPassword)
		}

		return nil, status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	logger.Info(fmt.Sprintf("Added user [UserName: %s] to chat room user's list with short code [%s]", userName, roomShortCode))

	token, err := s.tokenService.GenerateToken(userName, roomShortCode)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error while generating token.")
	}

	logger.Info(fmt.Sprintf("Generated token [%s] for user [UserName: %s] to chat room with short code [%s]", token, userName, roomShortCode))

	return &proto.JoinChatRoomResponse{
		Token: token,
	}, nil
}

// Chat is a server-side streaming RPC handler that receives messages from users and broadcasts them to all other users.
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

	logger.Info(fmt.Sprintf("User [UserName: %s] established message stream with the chat room with short code [%s] using token [%s]", userName, roomShortCode, userToken))

	if !s.roomService.RoomExists(roomShortCode) {
		return status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
	}

	is, err := s.roomService.IsUserInRoom(roomShortCode, userName)
	if !is {
		return status.Error(codes.PermissionDenied, "No permission to access this room. You do not have permission to participate in this chat room.")
	}
	if err != nil {
		if errors.Is(err, services.ErrRoomDoesNotExist) {
			return status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
		}

		return status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	receiveCh := make(chan struct{}, 1)
	sendCh := make(chan struct{}, 1)

	go s.receive(chs, userName, roomShortCode, sendCh, receiveCh, wg)
	go s.send(chs, userName, roomShortCode, receiveCh, sendCh, wg)

	wg.Wait()

	close(receiveCh)
	close(sendCh)

	logger.Info(fmt.Sprintf("Closed message stream with user [UserName: %s] and the chat room with short code [%s]", userName, roomShortCode))

	return nil
}

func (s *GRPCChatterServer) receive(chs proto.GRPCChatter_ChatServer, userName, roomShortCode string, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		default:
			mssg, err := chs.Recv()
			if err != nil {
				if status.Code(err) == codes.Canceled {
					logger.Info(fmt.Sprintf("User [UserName: %s] left the chat room with short code [%s]", userName, roomShortCode))

					_ = s.roomService.RemoveUserFromRoom(roomShortCode, userName)

					logger.Info(fmt.Sprintf("Removed user [UserName: %s] from chat room user's list with short code [%s]", userName, roomShortCode))

				} else {
					logger.Error(fmt.Sprintf("Failed to receive message from user [UserName: %s] in chat room with short code [%s]: %s", userName, roomShortCode, status.Convert(err).Message()))
				}

				sendStopCh <- struct{}{}

				return
			}

			body := mssg.GetBody()

			logger.Info(fmt.Sprintf("Received message [Body: %s] from user [UserName: %s] in chat room with short code [%s]", body, userName, roomShortCode))

			if err := s.roomService.BroadcastMessageToRoom(roomShortCode, &services.Message{
				Sender: userName,
				Body:   body,
			}); err != nil {
				logger.Error(fmt.Sprintf("Failed to broadcast message from user [UserName: %s] in chat room with short code [%s]: %s", userName, roomShortCode, status.Convert(err).Message()))

				sendStopCh <- struct{}{}

				return
			}
		}
	}
}

func (s *GRPCChatterServer) send(chs proto.GRPCChatter_ChatServer, userName, roomShortCode string, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		default:
			msg, err := s.roomService.GetUserMessage(roomShortCode, userName)
			if err != nil {
				sendStopCh <- struct{}{}

				return
			}

			if err := chs.Send(&proto.ServerMessage{
				UserName: msg.Sender,
				Body:     msg.Body,
			}); err != nil {
				logger.Error(fmt.Sprintf("Failed to send message to user [UserName: %s] in chat room with short code [%s]: %s", userName, roomShortCode, status.Convert(err).Message()))

				sendStopCh <- struct{}{}

				return
			}

			logger.Info(fmt.Sprintf("Sent message [{Sender: %s, Body: %s}] to user [UserName: %s] in chat room with short code [%s]", msg.Sender, msg.Body, userName, roomShortCode))
		}
	}
}
