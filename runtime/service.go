package runtime

import (
	"context"

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

	// LookupMethod returns information about a specific RPC method within the
	// service.
	//
	// If no such method exists, ok is false.
	LookupMethod(name string) (_ Method, ok bool)
}

// Method encapsulates information about an RPC method.
type Method interface {
	// Name returns the name of the RPC method.
	Name() string

	// ClientStreaming returns true if the method uses "client streaming", that
	// is, streams of requests.
	ClientStreaming() bool

	// ServerStreaming returns true fi the method uses "server streaming", that
	// is, streams of responses.
	ServerStreaming() bool

	// NewCall starts a new call to the method.
	//
	// ctx is the context for the lifetime of the call, including the time taken
	// to stream requests and responses.
	NewCall(context.Context) Call
}

// Call encapsulates the state of a single invocation of an RPC method.
type Call interface {
	// Recv returns the next response from the call.
	//
	// If there are no more responses, ok is false.
	Recv() (_ proto.Message, ok bool, _ error)

	// Send sends the next request to the call.
	Send(Unmarshaler) error

	// Done is called to indicate that no more requests will be sent.
	Done()
}

// Unmarshaler is a function that unmarshals a protocol buffers message into m.
type Unmarshaler func(m proto.Message) error
