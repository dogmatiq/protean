package protean

import (
	"fmt"
	"strconv"

	"google.golang.org/protobuf/proto"
)

// ErrorCode is a numeric code that identifies the general class of an RPC
// error.
type ErrorCode struct{ n int32 }

var (
	// ErrorCodeUnknown is an error code used when no information is available
	// about the error.
	ErrorCodeUnknown = ErrorCode{0}

	// ErrorCodeInvalidInput indicates that the input message to the RPC method
	// is invalid according to some application-defined rules.
	//
	// ErrorCodeInvalidInput means the input is inherently problematic.
	// ErrorCodeFailedPrecondition is the appropriate code to use if the input
	// is only invalid due to the current state of the application.
	ErrorCodeInvalidInput = ErrorCode{-1}

	// ErrorCodeUnauthenticated indicates that the client has attempted to
	// perform some action that requires authentication, but valid
	// authentication credentials have not been provided.
	ErrorCodeUnauthenticated = ErrorCode{-2}

	// ErrorCodePermissionDenied indicates that the caller does not have
	// permission to perform some action.
	//
	// It differs from ErrorCodeUnauthenticated, which indicates that valid
	// credentials have not been supplied at all.
	ErrorCodePermissionDenied = ErrorCode{-3}

	// ErrorCodeNotFound indicates that the client has requested some entity
	// that was not found.
	ErrorCodeNotFound = ErrorCode{-4}

	// ErrorCodeAlreadyExists indicates that client has attempted to create some
	// entity that already exists.
	ErrorCodeAlreadyExists = ErrorCode{-5}

	// ErrorCodeResourceExhausted indicates that some resource has been
	// exhausted, such as a rate limit.
	ErrorCodeResourceExhausted = ErrorCode{-6}

	// ErrorCodeFailedPrecondition indicates the application is not in the
	// required state to perform some action.
	//
	// The client should not retry until the application state has been
	// explicitly changed.
	//
	// Use ErrorCodeUnavailable if the client can safely re-send the failed RPC
	// request, or ErrorCodeAborted if the client should retry at a higher
	// level, such as by beginning some business process again.
	ErrorCodeFailedPrecondition = ErrorCode{-7}

	// ErrorCodeAborted indicates some action was aborted.
	//
	// The client may retry the operation by restarting whatever higher level
	// process it belongs to, but should not simply re-send the same RPC
	// request.
	ErrorCodeAborted = ErrorCode{-8}

	// ErrorCodeUnavailable indicates that the server is temporarily unable to
	// fulfill a request.
	//
	// The client may safely retry the RPC call by re-sending the request,
	// typically after some delay.
	ErrorCodeUnavailable = ErrorCode{-9}

	// ErrorCodeUnimplemented indicates an RPC method is not implemented or
	// otherwise unsupported by the server.
	ErrorCodeUnimplemented = ErrorCode{-10}
)

// CustomErrorCode returns a new application-defined error code.
//
// c is the numeric value of the application-defined error code, it must be a
// positive integer.
//
// Error codes should be used to organize related errors into broad categories
// based on their general meaning. This allows RPC clients to handle the error
// without being able to identify the specific cause.
//
// Where possible, server implementations should favour using the pre-defined
// error codes from this package over defining custom error codes.
func CustomErrorCode(c int32) ErrorCode {
	if c <= 0 {
		panic("error code must be positive")
	}

	return ErrorCode{c}
}

func (c ErrorCode) String() string {
	switch c {
	case ErrorCodeUnknown:
		return "unknown"
	case ErrorCodeInvalidInput:
		return "invalid input"
	case ErrorCodeUnauthenticated:
		return "unauthenticated"
	case ErrorCodePermissionDenied:
		return "permission denied"
	case ErrorCodeNotFound:
		return "not found"
	case ErrorCodeAlreadyExists:
		return "already exists"
	case ErrorCodeResourceExhausted:
		return "resource exhausted"
	case ErrorCodeFailedPrecondition:
		return "failed precondition"
	case ErrorCodeAborted:
		return "aborted"
	case ErrorCodeUnavailable:
		return "unavailable"
	case ErrorCodeUnimplemented:
		return "unimplemented"
	}

	return strconv.FormatInt(
		int64(c.n),
		10,
	)
}

// Error is an error produced by an RPC server that is intended to be received
// by the client.
//
// These errors form part of the service's public API, as opposed to "runtime
// errors" (such as network timeouts, etc) which are unexpected and meaningless
// within the context of the application's business domain.
type Error struct {
	code    ErrorCode
	message string
	details proto.Message
	cause   error
}

// NewError returns an error that will be sent from the server to the client.
//
// c is the error code that best describes the error.
//
// The error message is produced by performing sprintf-style interpolation on
// format and args.
//
// The error message should be understood by technical users that maintain or
// operate the software making the RPC request. These people are typically not
// the end-users of the software.
func NewError(c ErrorCode, format string, args ...interface{}) Error {
	return Error{
		code:    c,
		message: fmt.Sprintf(format, args...),
	}
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

	e.details = d

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

// Code returns the error's code.
//
// Clients should use the code to decide how best to handle the error if no
// better determination can be made by examining the error's application-defined
// details value.
func (e Error) Code() ErrorCode {
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
// ok is false if no details were provided.
func (e Error) Details() (details proto.Message, ok bool) {
	return e.details, e.details != nil
}

func (e Error) Error() string {
	if e.details != nil {
		return fmt.Sprintf(
			"%s: %s (%s)",
			e.code,
			e.message,
			proto.MessageName(e.details),
		)
	}

	return fmt.Sprintf(
		"%s: %s",
		e.code,
		e.message,
	)
}
