package middleware

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// UnaryServerInfo encapsulates information about a call to unary RPC method and
// makes it available to a ServerInterceptor implementation.
type UnaryServerInfo struct {
	// Package is the name of the Protocol Buffers package that contains the
	// service definition.
	Package string

	// Service is the name of the RPC service.
	Service string

	// Method is the name of the RPC method being invoked.
	Method string
}

// ServerInterceptor is an interface intercepting RPC method calls on the
// server-side.
// server.
type ServerInterceptor interface {
	// InterceptUnaryRPC is called before the RPC method is invoked.
	//
	// It must call next() to forward the call to the next interceptor in the
	// chain, or ultimately to the application-defined server implementation.
	//
	// It returns the output that should be sent to the client.
	//
	// The RPC input message may be mutated in place. The output message
	// returned by next() must not be modified. To produce different RPC output,
	// return a new output message or error.
	InterceptUnaryRPC(
		ctx context.Context,
		info UnaryServerInfo,
		in proto.Message,
		next func(ctx context.Context) (out proto.Message, err error),
	) (proto.Message, error)
}

// ServerChain is a ServerInterceptor that chains multiple interceptors to be
// applied sequentially.
type ServerChain []ServerInterceptor

// InterceptUnaryRPC is called before the RPC method is invoked.
//
// It must call next() to forward the call to the next interceptor in the
// chain, or ultimately to the application-defined server implementation.
//
// It returns the output that should be sent to the client.
//
// The RPC input message may be mutated in place. The output message
// returned by next() must not be modified. To produce different RPC output,
// return a new output message or error.
func (c ServerChain) InterceptUnaryRPC(
	ctx context.Context,
	info UnaryServerInfo,
	in proto.Message,
	next func(ctx context.Context) (out proto.Message, err error),
) (proto.Message, error) {
	if len(c) == 0 {
		return next(ctx)
	}

	head, tail := c[0], c[1:]

	return head.InterceptUnaryRPC(
		ctx,
		info,
		in,
		func(ctx context.Context) (out proto.Message, err error) {
			return tail.InterceptUnaryRPC(ctx, info, in, next)
		},
	)
}
