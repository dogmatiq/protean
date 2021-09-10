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
	NewCall(ctx context.Context) Call
}

// Call represents a single invocation of an RPC method.
type Call interface {
	// Send sends an input message to the call.
	//
	// u is an unmarshaler that unmarshals the input message.
	Send(u Unmarshaler) error

	// Done is called to indicate that no more input messages will be sent.
	Done()

	// Recv returns the next output message produced by this call.
	//
	// If ok is true, out is the next output message. Otherwise, there are no
	// more output messages to be received, and out is nil.
	//
	// err is the error returned by the RPC method, if any.
	Recv() (out proto.Message, ok bool, err error)
}

// Unmarshaler is a function that unmarshals a protocol buffers message into m.
type Unmarshaler func(m proto.Message) error
