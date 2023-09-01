package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/MSSkowron/GRPCChatter/internal/service"
	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/MSSkowron/GRPCChatter/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type contextKey string

const (
	// DefaultPort is the default port the server listens on.
	DefaultPort = 5000
	// DefaultAddress is the default address the server listens on.
	DefaultAddress     = ""
	grpcHeaderTokenKey = "token"

	contextKeyRPCID     = contextKey("rpcID")
	contextKeyShortCode = contextKey("shortCode")
	contextKeyUserID    = contextKey("userID")
	contextKeyUserName  = contextKey("userName")
)

// Server represents a gRPC server.
type Server struct {
	proto.UnimplementedGRPCChatterServer

	chatTokenService service.ChatTokenService
	userTokenService service.UserTokenService
	shortCodeService service.ShortCodeService
	roomService      service.RoomService

	address string
	port    int

	authorizedUserTokenUnaryMethods  map[string]struct{}
	authorizedChatTokenUnaryMethods  map[string]struct{}
	authorizedUserTokenStreamMethods map[string]struct{}
	authorizedChatTokenStreamMethods map[string]struct{}
}

// NewServer creates a new GRPCChatter server.
func NewServer(chatTokenService service.ChatTokenService, userTokenService service.UserTokenService, shortCodeService service.ShortCodeService, roomService service.RoomService, opts ...Opt) *Server {
	server := &Server{
		chatTokenService: chatTokenService,
		userTokenService: userTokenService,
		shortCodeService: shortCodeService,
		roomService:      roomService,
		address:          DefaultAddress,
		port:             DefaultPort,
		authorizedUserTokenUnaryMethods: map[string]struct{}{
			"/proto.GRPCChatter/CreateChatRoom": {},
			"/proto.GRPCChatter/JoinChatRoom":   {},
		},
		authorizedChatTokenUnaryMethods: map[string]struct{}{
			"/proto.GRPCChatter/ListChatRoomUsers": {},
		},
		authorizedUserTokenStreamMethods: map[string]struct{}{},
		authorizedChatTokenStreamMethods: map[string]struct{}{
			"/proto.GRPCChatter/Chat": {},
		},
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

// Opt represents an option that can be passed to NewServer.
type Opt func(*Server)

// WithAddress sets the address the server listens on.
func WithAddress(address string) Opt {
	return func(s *Server) {
		s.address = address
	}
}

// WithPort sets the port the server listens on.
func WithPort(port int) Opt {
	return func(s *Server) {
		s.port = port
	}
}

// ListenAndServe starts the server and listens for incoming connections.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.address+":"+strconv.Itoa(s.port))
	if err != nil {
		return fmt.Errorf("failed to create tcp listener on %s:%d: %w", s.address, s.port, err)
	}

	logger.Info(fmt.Sprintf("Server listening on %s:%d", s.address, s.port))

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(s.unaryLogInterceptor, s.unaryAuthorizationInterceptor),
		grpc.ChainStreamInterceptor(s.streamLogInterceptor, s.streamAuthorizationInterceptor),
	)
	proto.RegisterGRPCChatterServer(grpcServer, s)

	if err := grpcServer.Serve(ln); err != nil {
		return fmt.Errorf("failed to run grpc server on %s:%d: %w", s.address, s.port, err)
	}

	return nil
}

// CreateChatRoom is an RPC handler that creates a new chat room.
func (s *Server) CreateChatRoom(ctx context.Context, req *proto.CreateChatRoomRequest) (*proto.CreateChatRoomResponse, error) {
	rpcID, userName := ctx.Value(contextKeyRPCID).(string), ctx.Value(contextKeyUserName).(string)

	roomName := req.GetRoomName()
	roomPassword := req.GetRoomPassword()

	roomShortCode := s.shortCodeService.GenerateShortCode(roomName)

	if err := s.roomService.CreateRoom(roomShortCode, roomName, roomPassword); err != nil {
		return nil, status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	logger.Info(fmt.Sprintf("[ID: %s]: User [%s] created room [%s] with short code [%s]", rpcID, userName, roomName, roomShortCode))

	return &proto.CreateChatRoomResponse{
		ShortCode: string(roomShortCode),
	}, nil
}

// JoinChatRoom is an RPC handler that allows a user to join an existing chat room.
func (s *Server) JoinChatRoom(ctx context.Context, req *proto.JoinChatRoomRequest) (*proto.JoinChatRoomResponse, error) {
	rpcID, userName := ctx.Value(contextKeyRPCID).(string), ctx.Value(contextKeyUserName).(string)

	roomShortCode := req.GetShortCode()
	roomPassword := req.GetRoomPassword()

	if !s.roomService.RoomExists(roomShortCode) {
		return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
	}

	if err := s.roomService.CheckPassword(roomShortCode, roomPassword); err != nil {
		if errors.Is(err, service.ErrRoomDoesNotExist) {
			return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
		}
		if errors.Is(err, service.ErrInvalidRoomPassword) {
			return nil, status.Errorf(codes.PermissionDenied, "Invalid room password for chat room with short code [%s]. Please make sure you have the correct password.", roomShortCode)
		}

		return nil, status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	if err := s.roomService.AddUserToRoom(roomShortCode, userName); err != nil {
		if errors.Is(err, service.ErrRoomDoesNotExist) {
			return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", roomShortCode)
		}
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "User with username [%s] already exists in the chat room with short code [%s].", userName, roomPassword)
		}

		return nil, status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	logger.Info(fmt.Sprintf("[ID: %s]: Added user [%s] to chat room user's list with short code [%s]", rpcID, userName, roomShortCode))

	token, err := s.chatTokenService.GenerateToken(userName, roomShortCode)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error while generating token.")
	}

	logger.Info(fmt.Sprintf("[ID: %s]: Generated token [%s] for user [%s] to chat room with short code [%s]", rpcID, token, userName, roomShortCode))

	return &proto.JoinChatRoomResponse{
		Token: token,
	}, nil
}

// ListChatRoomUsers is an RPC handler that lists the users in a chat room.
func (s *Server) ListChatRoomUsers(ctx context.Context, req *emptypb.Empty) (*proto.ListChatRoomUsersResponse, error) {
	rpcID, shortCode, userName := ctx.Value(contextKeyRPCID).(string), ctx.Value(contextKeyShortCode).(string), ctx.Value(contextKeyUserName).(string)

	users, err := s.roomService.GetRoomUsers(shortCode)
	if err != nil {
		if errors.Is(err, service.ErrRoomDoesNotExist) {
			return nil, status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", shortCode)
		}

		return nil, status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	var resUsers []*proto.User
	for _, user := range users {
		if user != userName {
			resUsers = append(resUsers, &proto.User{
				UserName: user,
			})
		}
	}

	logger.Info(fmt.Sprintf("[ID: %s]: Listed chat room with short code [%s] users [%v]", rpcID, shortCode, users))

	return &proto.ListChatRoomUsersResponse{
		Users: resUsers,
	}, nil
}

// Chat is a server-side streaming RPC handler that receives messages from users and broadcasts them to all other users.
func (s *Server) Chat(chs proto.GRPCChatter_ChatServer) error {
	rpcID, shortCode, userName := chs.Context().Value(contextKeyRPCID).(string), chs.Context().Value(contextKeyShortCode).(string), chs.Context().Value(contextKeyUserName).(string)

	logger.Info(fmt.Sprintf("[ID: %s]: User [%s] established message stream with the chat room with short code [%s]", rpcID, userName, shortCode))

	wg := &sync.WaitGroup{}
	wg.Add(2)

	receiveCh := make(chan struct{}, 1)
	sendCh := make(chan struct{}, 1)

	go s.receive(rpcID, chs, userName, shortCode, sendCh, receiveCh, wg)
	go s.send(rpcID, chs, userName, shortCode, receiveCh, sendCh, wg)

	wg.Wait()

	close(receiveCh)
	close(sendCh)

	logger.Info(fmt.Sprintf("[ID: %s]: Closed message stream with user [%s] and the chat room with short code [%s]", rpcID, userName, shortCode))

	return nil
}

func (s *Server) receive(id string, chs proto.GRPCChatter_ChatServer, userName, roomShortCode string, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		default:
			mssg, err := chs.Recv()
			if err != nil {
				if status.Code(err) == codes.Canceled {
					logger.Info(fmt.Sprintf("[ID: %s]: User [%s] left the chat room with short code [%s]", id, userName, roomShortCode))

					_ = s.roomService.RemoveUserFromRoom(roomShortCode, userName)

					logger.Info(fmt.Sprintf("[ID: %s]: Removed user [%s] from chat room user's list with short code [%s]", id, userName, roomShortCode))

				} else {
					logger.Error(fmt.Sprintf("[ID: %s]: Failed to receive message from user [%s] in chat room with short code [%s]: %s", id, userName, roomShortCode, status.Convert(err).Message()))
				}

				sendStopCh <- struct{}{}

				return
			}

			body := mssg.GetBody()

			logger.Info(fmt.Sprintf("[ID: %s]: Received message [{Body: %s}] from user [%s] in chat room with short code [%s]", id, body, userName, roomShortCode))

			if err := s.roomService.BroadcastMessageToRoom(roomShortCode, &service.Message{
				Sender: userName,
				Body:   body,
			}); err != nil {
				logger.Error(fmt.Sprintf("[ID: %s]: Failed to broadcast message from user [%s] in chat room with short code [%s]: %s", id, userName, roomShortCode, status.Convert(err).Message()))

				sendStopCh <- struct{}{}

				return
			}
		}
	}
}

func (s *Server) send(id string, chs proto.GRPCChatter_ChatServer, userName, roomShortCode string, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
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
				logger.Error(fmt.Sprintf("[ID: %s]: Failed to send message to user [%s] in chat room with short code [%s]: %s", id, userName, roomShortCode, status.Convert(err).Message()))

				sendStopCh <- struct{}{}

				return
			}

			logger.Info(fmt.Sprintf("[ID: %s]: Sent message [{Sender: %s, Body: %s}] to user [%s] in chat room with short code [%s]", id, msg.Sender, msg.Body, userName, roomShortCode))
		}
	}
}
