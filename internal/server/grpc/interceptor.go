package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/service"
	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/MSSkowron/GRPCChatter/pkg/wrapper"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	errMsgMissingHeaders       = "Missing gRPC headers: [%s]. Please include your authentication token in the [%s] gRPC header."
	errMsgTokenMissing         = "Authentication token missing in gRPC headers. Please include your token in the [%s] gRPC header."
	errMsgInvalidToken         = "Invalid authentication token. Please provide a valid token."
	errMsgNoPermissionToAccess = "No permission to access chat room with short code [%s]."
)

func (s *Server) unaryLogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	id := uuid.New().String()

	logger.Info(fmt.Sprintf("Received Unary RPC [ID: %s] [Method: %s] with [%v]", id, info.FullMethod, req))

	ctx = context.WithValue(ctx, contextKeyRPCID, id)

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
	id := uuid.New().String()

	logger.Info(fmt.Sprintf("Received Stream RPC [ID: %s] [Method: %s] call with [%v]", id, info.FullMethod, srv))

	ctx := context.WithValue(ss.Context(), contextKeyRPCID, id)

	wrapped := wrapper.WrapServerStream(ss)
	wrapped.SetContext(ctx)

	return handler(srv, wrapped)
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
		return "", "", status.Errorf(codes.Unauthenticated, errMsgMissingHeaders, grpcHeaderTokenKey, grpcHeaderTokenKey)
	}

	tokens := md.Get(grpcHeaderTokenKey)
	if len(tokens) == 0 {
		return "", "", status.Errorf(codes.Unauthenticated, errMsgTokenMissing, grpcHeaderTokenKey)
	}
	userToken := tokens[0]

	if err := s.chatTokenService.ValidateToken(userToken); err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return "", "", status.Error(codes.Unauthenticated, errMsgInvalidToken)
		}

		return "", "", status.Errorf(codes.Internal, errMsgInternalServer, "validating token")
	}

	shortCode, err := s.chatTokenService.GetShortCodeFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return "", "", status.Error(codes.Unauthenticated, errMsgInvalidToken)
		}

		return "", "", status.Errorf(codes.Internal, errMsgInternalServer, "retrieving short code from token")
	}

	userName, err := s.chatTokenService.GetUserNameFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return "", "", status.Error(codes.Unauthenticated, errMsgInvalidToken)
		}

		return "", "", status.Errorf(codes.Internal, errMsgInternalServer, "retrieving user name from token")
	}

	if !s.roomService.RoomExists(shortCode) {
		return "", "", status.Errorf(codes.NotFound, errMsgChatRoomNotFound, shortCode)
	}

	is, err := s.roomService.IsUserInRoom(shortCode, userName)
	if !is {
		return "", "", status.Errorf(codes.PermissionDenied, errMsgNoPermissionToAccess, shortCode)
	}
	if err != nil {
		if errors.Is(err, service.ErrRoomDoesNotExist) {
			return "", "", status.Errorf(codes.NotFound, errMsgChatRoomNotFound, shortCode)
		}

		return "", "", status.Errorf(codes.Internal, errMsgInternalServer, "checking user presence in chat room")
	}

	return shortCode, userName, nil
}

func (s *Server) authorizeUserToken(ctx context.Context) (int, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, "", status.Errorf(codes.Unauthenticated, errMsgMissingHeaders, grpcHeaderTokenKey, grpcHeaderTokenKey)
	}

	tokens := md.Get(grpcHeaderTokenKey)
	if len(tokens) == 0 {
		return 0, "", status.Errorf(codes.Unauthenticated, errMsgTokenMissing, grpcHeaderTokenKey)
	}
	userToken := tokens[0]

	if err := s.userTokenService.ValidateToken(userToken); err != nil {
		if errors.Is(err, service.ErrInvalidUserToken) {
			return 0, "", status.Error(codes.Unauthenticated, errMsgInvalidToken)
		}

		return 0, "", status.Errorf(codes.Internal, errMsgInternalServer, "validating token")
	}

	userID, err := s.userTokenService.GetUserIDFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserToken) {
			return 0, "", status.Error(codes.Unauthenticated, errMsgInvalidToken)
		}

		return 0, "", status.Errorf(codes.Internal, errMsgInternalServer, "retrieving user ID from token")
	}

	userName, err := s.userTokenService.GetUserNameFromToken(userToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidChatToken) {
			return 0, "", status.Error(codes.Unauthenticated, errMsgInvalidToken)
		}

		return 0, "", status.Errorf(codes.Internal, errMsgInternalServer, "retrieving user name from token")
	}

	return userID, userName, nil
}
