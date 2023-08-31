package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/service"
	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/MSSkowron/GRPCChatter/pkg/wrapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) unaryLogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	logger.Info(fmt.Sprintf("Received Unary RPC [%s] call with [%v]", info.FullMethod, req))

	return handler(ctx, req)
}

func (s *Server) unaryAuthorizationInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, exists := s.authorizedChatTokenUnaryMethods[info.FullMethod]; exists {
		shortCode, userName, err := s.authorizeChatToken(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, contextKeyShortCode, shortCode)
		ctx = context.WithValue(ctx, contextKeyUserName, userName)
		return handler(ctx, req)
	}
	if _, exists := s.authorizedUserTokenUnaryMethods[info.FullMethod]; exists {
		userID, userName, err := s.authorizeUserToken(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, contextKeyUserID, userID)
		ctx = context.WithValue(ctx, contextKeyUserName, userName)
		return handler(ctx, req)
	}

	return handler(ctx, req)
}

func (s *Server) streamLogInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logger.Info(fmt.Sprintf("Received Stream RPC [%s] call with [%v]", info.FullMethod, srv))

	return handler(srv, ss)
}

func (s *Server) streamAuthorizationInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if _, exists := s.authorizedChatTokenStreamMethods[info.FullMethod]; exists {
		shortCode, userName, err := s.authorizeChatToken(ss.Context())
		if err != nil {
			return err
		}

		newCtx := context.WithValue(ss.Context(), contextKeyShortCode, shortCode)
		newCtx = context.WithValue(newCtx, contextKeyUserName, userName)

		wrapped := wrapper.WrapServerStream(ss)
		wrapped.SetContext(newCtx)

		return handler(srv, wrapped)
	}
	if _, exists := s.authorizedUserTokenStreamMethods[info.FullMethod]; exists {
		userID, userName, err := s.authorizeUserToken(ss.Context())
		if err != nil {
			return err
		}

		newCtx := context.WithValue(ss.Context(), contextKeyUserID, userID)
		newCtx = context.WithValue(newCtx, contextKeyUserName, userName)

		wrapped := wrapper.WrapServerStream(ss)
		wrapped.SetContext(newCtx)

		return handler(srv, wrapped)
	}

	return handler(srv, ss)
}

func (s *Server) authorizeChatToken(ctx context.Context) (string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", status.Errorf(codes.Unauthenticated, "Missing gRPC headers: %s. Please include your authentication token in the '%s' gRPC header.", grpcHeaderTokenKey, grpcHeaderTokenKey)
	}

	tokens := md.Get(grpcHeaderTokenKey)
	if len(tokens) == 0 {
		return "", "", status.Errorf(codes.Unauthenticated, "Authentication token missing in gRPC headers. Please include your token in the '%s' gRPC header.", grpcHeaderTokenKey)
	}
	userToken := tokens[0]

	if err := s.chatTokenService.ValidateToken(userToken); err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return "", "", status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return "", "", status.Error(codes.Internal, "Internal server error while validating token.")
	}

	shortCode, err := s.chatTokenService.GetShortCodeFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return "", "", status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return "", "", status.Error(codes.Internal, "Internal server error while retrieving short code from token.")
	}

	userName, err := s.chatTokenService.GetUserNameFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return "", "", status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return "", "", status.Error(codes.Internal, "Internal server error while retrieving user name from token.")
	}

	if !s.roomService.RoomExists(shortCode) {
		return "", "", status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", shortCode)
	}

	is, err := s.roomService.IsUserInRoom(shortCode, userName)
	if !is {
		return "", "", status.Error(codes.PermissionDenied, "No permission to access this room. You do not have permission to participate in this chat room.")
	}
	if err != nil {
		if errors.Is(err, service.ErrRoomDoesNotExist) {
			return "", "", status.Errorf(codes.NotFound, "Chat room with short code [%s] not found. Please check the provided short code.", shortCode)
		}

		return "", "", status.Error(codes.Internal, "Internal server error while adding user to chat room.")
	}

	return shortCode, userName, nil
}

func (s *Server) authorizeUserToken(ctx context.Context) (int, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, "", status.Errorf(codes.Unauthenticated, "Missing gRPC headers: %s. Please include your authentication token in the '%s' gRPC header.", grpcHeaderTokenKey, grpcHeaderTokenKey)
	}

	tokens := md.Get(grpcHeaderTokenKey)
	if len(tokens) == 0 {
		return 0, "", status.Errorf(codes.Unauthenticated, "Authentication token missing in gRPC headers. Please include your token in the '%s' gRPC header.", grpcHeaderTokenKey)
	}
	userToken := tokens[0]

	if err := s.userTokenService.ValidateToken(userToken); err != nil {
		if errors.Is(err, service.ErrInvalidUserToken) {
			return 0, "", status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return 0, "", status.Error(codes.Internal, "Internal server error while validating token.")
	}

	userID, err := s.userTokenService.GetUserIDFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserToken) {
			return 0, "", status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return 0, "", status.Error(codes.Internal, "Internal server error while retrieving short code from token.")
	}

	userName, err := s.userTokenService.GetUserNameFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return 0, "", status.Error(codes.Unauthenticated, "Invalid authentication token. Please provide a valid token.")
		}

		return 0, "", status.Error(codes.Internal, "Internal server error while retrieving user name from token.")
	}

	return userID, userName, nil
}
