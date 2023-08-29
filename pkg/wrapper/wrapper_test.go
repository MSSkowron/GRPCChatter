package wrapper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	someKey struct{}
	other   struct{}
)

func TestContext(t *testing.T) {
	fakeStream := &fakeServerStream{}
	wrapped := WrapServerStream(fakeStream)

	t.Run("SetAndGetValues", func(t *testing.T) {
		newContext := context.WithValue(context.TODO(), someKey, 1)
		wrapped.SetContext(newContext)

		assert.Equal(t, wrapped.Context(), newContext, "Context not equal to the newly set context")
		assert.NotNil(t, wrapped.Context().Value(someKey), "Values from fake must propagate to wrapper")
	})

	t.Run("SetAndGetValuesFromWrapper", func(t *testing.T) {
		newContext := context.WithValue(wrapped.Context(), other, 2)
		wrapped.SetContext(newContext)

		assert.Equal(t, wrapped.Context(), newContext, "Context not equal to the newly set context")
		assert.NotNil(t, wrapped.Context().Value(other), "Values from wrapper must be set")
	})
}

func TestWrapServerStream(t *testing.T) {
	fakeStream := &fakeServerStream{}
	wrapped := WrapServerStream(fakeStream)

	t.Run("ServerStreamEquality", func(t *testing.T) {
		assert.Equal(t, wrapped.ServerStream, fakeStream, "Expected ServerStream to be the same as input")
	})

	t.Run("WrappedContextEquality", func(t *testing.T) {
		assert.Equal(t, wrapped.WrappedContext, fakeStream.ctx, "Expected WrappedContext to be ServerStream.Context()")
	})
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx         context.Context
	recvMessage any
	sentMessage any
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}

func (f *fakeServerStream) SendMsg(m any) error {
	if f.sentMessage != nil {
		return status.Error(codes.AlreadyExists, "fakeServerStream only takes one message, sorry")
	}
	f.sentMessage = m
	return nil
}

func (f *fakeServerStream) RecvMsg(m any) error {
	if f.recvMessage == nil {
		return status.Error(codes.NotFound, "fakeServerStream has no message, sorry")
	}
	return nil
}
