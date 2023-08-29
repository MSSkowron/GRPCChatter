package wrapper

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying context.
type WrappedServerStream struct {
	grpc.ServerStream
	WrappedContext context.Context
}

// SetContext sets the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *WrappedServerStream) SetContext(ctx context.Context) {
	w.WrappedContext = ctx
}

// Context returns the wrapper's WrappedContext.
func (w *WrappedServerStream) Context() context.Context {
	if w.WrappedContext != nil {
		return w.WrappedContext
	}
	return w.ServerStream.Context()
}

// WrapServerStream returns a ServerStream that has the ability to overwrite context.
func WrapServerStream(stream grpc.ServerStream) *WrappedServerStream {
	if existing, ok := stream.(*WrappedServerStream); ok {
		return existing
	}
	return &WrappedServerStream{
		ServerStream:   stream,
		WrappedContext: stream.Context(),
	}
}