package runtime

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Server is an interface for registering service handlers with a central
// server.
type Server interface {
	RegisterService(Service)
}

// Service is a genericized interface to a protocol buffers service. The
// implementation is provided by generated code.
type Service interface {
	Name() string
	Package() string
	LookupMethod(name string) (Method, bool)
}

// Method encapsulates information about a single method.
type Method interface {
	Name() string
	ClientStreaming() bool
	ServerStreaming() bool
	NewCall(context.Context) Call
}

// Call encapsulates the state of a single call to an RPC method.
type Call interface {
	Recv() (proto.Message, bool, error)
	Send(RawMessage) error
	Done()
}

// RawMessage represents a message that has not been unmarshaled yet.
//
// It is a function that, when invoked, unmarshals the message into m.
type RawMessage func(m proto.Message) error
