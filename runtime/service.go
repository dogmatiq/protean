package runtime

import (
	"context"

	"github.com/dogmatiq/protean/middleware"
	"google.golang.org/protobuf/proto"
)

// Registry is a type that allows services to be registered.
type Registry interface {
	RegisterService(Service)
}

// Service is a generalized interface to a Protocol Buffers service.
//
// The implementation is provided by the code generated from the service
// definition.
type Service interface {
	// Name returns the unqualified service name.
	Name() string

	// Package returns the Protocol Buffers package name in which the service is
	// defined.
	Package() string

	// MethodByName returns the method with the given name.
	//
	// If no such method exists, ok is false.
	MethodByName(name string) (_ Method, ok bool)

	// // MethodByURL returns the method that matches the given HTTP request URL
	// // based on the protean.method.http_route option.
	// //
	// // un an Unmarshaler that populates the method's input message based on path
	// // segments and query parameters in the given URL.
	// //
	// // If no method matches this route, ok is false.
	// MethodByURL(u *url.URL) (_ Method, un Unmarshaler, ok bool)
}

// Method encapsulates information about an RPC method.
type Method interface {
	// Name returns the name of the RPC method.
	Name() string

	// InputIsStream returns true if the method accepts a stream of input
	// messages, as opposed to a single input message.
	InputIsStream() bool

	// OutputIsStream returns true if the method produces a stream of output
	// messages, as opposed to a single output message.
	OutputIsStream() bool

	// NewCall starts a new call to the method.
	//
	// ctx is the context for the lifetime of the call, including any time taken
	// to stream input and output messages.
	NewCall(ctx context.Context, options CallOptions) Call
}

// CallOptions encapsulates the options used when making a new call.
type CallOptions struct {
	// Interceptor is a hook that intercepts calls to RPC methods and
	// input/output values.
	Interceptor middleware.ServerInterceptor

	// InputChannelCapacity is the capacity of the "inputs" channel for RPC
	// methods that use client-streaming.
	InputChannelCapacity int

	// OutputChannelCapcity is the capacity of the "outputs" channel for RPC
	// methods that use server-streaming.
	OutputChannelCapacity int
}

// Call represents a single invocation of an RPC method.
type Call interface {
	// Send sends an input message to the call.
	//
	// u is an unmarshaler that produces the input message.
	// err is the error produced by the unmarshaler.
	//
	// more is true if the call can accept additional input messages.
	Send(u Unmarshaler) (more bool, err error)

	// Done is called to indicate that no more input messages will be sent.
	Done()

	// Recv returns the next output message produced by this call.
	//
	// If ok is true then out is an RPC output message, and Recv() should be
	// called again.
	Recv() (out proto.Message, ok bool)

	// Wait blocks until the RPC method returns.
	//
	// It must not be called until after Recv() has returned false.
	//
	// If err is non-nil, the RPC method has returned an error or the call's
	// context has been canceled.
	Wait() error
}

// Unmarshaler is a function that unmarshals a protocol buffers message into m.
type Unmarshaler func(m proto.Message) error
