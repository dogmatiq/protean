package rpcerror

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Error is an error produced by an RPC server that is intended to be received
// by the client.
//
// These errors form part of the service's public API, as opposed to "runtime
// errors" (such as network timeouts, etc) which are unexpected and meaningless
// within the context of the application's business domain.
type Error struct {
	code    Code
	message string
	details *anypb.Any
	cause   error
}

// New returns an error that will be sent from the server to the client.
//
// c is the error code that best describes the error.
//
// The error message is produced by performing sprintf-style interpolation on
// format and args.
//
// The error message should be understood by technical users that maintain or
// operate the software making the RPC request. These people are typically not
// the end-users of the software.
func New(c Code, format string, args ...interface{}) Error {
	return Error{
		c,
		fmt.Sprintf(format, args...),
		nil,
		nil,
	}
}

// Code returns the error's code.
//
// Clients should use the code to decide how best to handle the error if no
// better determination can be made by examining the error's application-defined
// details value.
func (e Error) Code() Code {
	return e.code
}

// Message returns a human-readable description of the message.
//
// This message is intended for technical users that maintain or operate the
// software making the RPC request, and should not be shown to end-users.
func (e Error) Message() string {
	return e.message
}

// Details returns application-defined information about this error.
//
// The client may use this information to notify the end-user about the error in
// whatever language or user interface may be appropriate.
//
// It returns an error if the details can not be unmarshaled.
//
// ok is true if error details are present in the error, even if an error
// occurs.
func (e Error) Details() (details proto.Message, ok bool, err error) {
	if e.details == nil {
		return nil, false, nil
	}

	d, err := e.details.UnmarshalNew()
	return d, true, err
}

// Unwrap returns the error that caused this RPC error, if known.
func (e Error) Unwrap() error {
	return e.cause
}

// WithDetails returns a copy of e that includes some application-defined
// information about the error.
//
// These details provide more specific information than can be conveyed by the
// error code.
//
// It is best practice to define a distinct Protocol Buffers message type for
// each error that the client is expected to handle in some unique way.
//
// The server should avoid including human-readable messages within the details
// value. Instead, include key information about the error that the client can
// use to notify the end-user about the error in whatever language or user
// interface may be appropriate.
func (e Error) WithDetails(d proto.Message) Error {
	if e.details != nil {
		panic("error details have already been provided")
	}

	e.details = &anypb.Any{}

	if err := e.details.MarshalFrom(d); err != nil {
		panic(err)
	}

	return e
}

// WithCause returns a copy of e that records err as the initial cause of the
// error.
//
// err is typically some unexpected runtime error that is important to the
// people who maintain the RPC server implementation, but not to the caller.
//
// Information about err is never sent to the client.
func (e Error) WithCause(err error) Error {
	if e.cause != nil {
		panic("error cause has already been provided")
	}

	e.cause = err

	return e
}

func (e Error) Error() string {
	message := e.message
	if message == "" {
		message = "<no message provided>"
	}

	if e.details == nil {
		return fmt.Sprintf(
			"%s: %s",
			e.code,
			message,
		)
	}

	detailsType := strings.TrimPrefix(
		e.details.GetTypeUrl(),
		"type.googleapis.com/",
	)

	return fmt.Sprintf(
		"%s [%s]: %s",
		e.code,
		detailsType,
		message,
	)
}

// ToProto returns the Protocol Buffers representation of an error.
func ToProto(err Error, m proto.Message) error {
	pb, ok := m.(*proteanpb.Error)
	if !ok {
		return errors.New("unsupported protocol buffers message type")
	}

	pb.Code = err.code.n
	pb.Message = err.message
	pb.Data = err.details

	return nil
}

// FromProto returns a new Error constructed from a Protocol Buffers
// representation.
func FromProto(m proto.Message) (Error, error) {
	pb, ok := m.(*proteanpb.Error)
	if !ok {
		return Error{}, errors.New("unsupported protocol buffers message type")
	}

	return Error{
		Code{pb.GetCode()},
		pb.GetMessage(),
		pb.GetData(),
		nil,
	}, nil
}
