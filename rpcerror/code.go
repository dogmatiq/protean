package rpcerror

import "strconv"

// Code is a numeric code that identifies the general class of an RPC error.
type Code struct{ n int32 }

var (
	// Unknown is an error code used when no information is available about the
	// error.
	//
	// The Unknown code is used whenever the server behaves incorrectly. Even
	// though the reason may be known on the server, it is inappropriate to
	// provide granular information about these errors to the client.
	Unknown = Code{0}

	// DeadlineExceeded is an error code that indicates the operation took too
	// long.
	//
	// This error may be returned even if the operation has completed
	// successfully.
	DeadlineExceeded = Code{-1}

	// Canceled is an error code that indicates execution of an RPC method was
	// forcibly terminated.
	Canceled = Code{-2}

	// InvalidInput is an error code that indicates that the input message to
	// the RPC method is invalid according to some application-defined rules.
	//
	// InvalidInput means the input is inherently problematic.
	//
	// FailedPrecondition is the appropriate code to use if the input is only
	// invalid due to the current state of the application.
	InvalidInput = Code{-3}

	// Unauthenticated is an error code that indicates that the client has
	// attempted to perform some action that requires authentication, but valid
	// authentication credentials have not been provided.
	Unauthenticated = Code{-4}

	// PermissionDenied is an error code that indicates that the caller does not
	// have permission to perform some action.
	//
	// It differs from Unauthenticated, which indicates that valid credentials
	// have not been supplied at all.
	PermissionDenied = Code{-5}

	// NotFound is an error code that indicates that the client has requested
	// some entity that was not found.
	NotFound = Code{-6}

	// AlreadyExists is an error code that indicates that client has attempted
	// to create some entity that already exists.
	AlreadyExists = Code{-7}

	// ResourceExhausted is an error code that indicates that some resource has
	// been exhausted, such as a rate limit.
	ResourceExhausted = Code{-8}

	// FailedPrecondition is an error code that indicates the application is not
	// in the required state to perform some action.
	//
	// The client should not retry until the application state has been
	// explicitly changed.
	//
	// Use Unavailable if the client can safely re-send the failed RPC request,
	// or Aborted if the client should retry at a higher level, such as by
	// beginning some business process again.
	FailedPrecondition = Code{-9}

	// Aborted is an error code that indicates some action was aborted.
	//
	// The client may retry the operation by restarting whatever higher level
	// process it belongs to, but should not simply re-send the same RPC
	// request.
	Aborted = Code{-10}

	// Unavailable is an error code that indicates that the server is
	// temporarily unable to fulfill a request.
	//
	// The client may safely retry the RPC call by re-sending the request,
	// typically after some delay.
	Unavailable = Code{-11}

	// NotImplemented is an error code that indicates an RPC method is not
	// implemented or otherwise unsupported by the server.
	NotImplemented = Code{-12}
)

// NewCode returns a new application-defined error code.
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
//
// If custom codes are necessary, it is recommended they be treated like an
// enumeration by assigning the result of NewCode() to global variables with
// meaningful names.
func NewCode(c int32) Code {
	if c <= 0 {
		panic("error code must be positive")
	}

	return Code{c}
}

// NumericValue returns the numeric value of the error code.
func (c Code) NumericValue() int32 {
	return c.n
}

func (c Code) String() string {
	switch c {
	case Unknown:
		return "unknown"
	case DeadlineExceeded:
		return "deadline exceeded"
	case Canceled:
		return "canceled"
	case InvalidInput:
		return "invalid input"
	case Unauthenticated:
		return "unauthenticated"
	case PermissionDenied:
		return "permission denied"
	case NotFound:
		return "not found"
	case AlreadyExists:
		return "already exists"
	case ResourceExhausted:
		return "resource exhausted"
	case FailedPrecondition:
		return "failed precondition"
	case Aborted:
		return "aborted"
	case Unavailable:
		return "unavailable"
	case NotImplemented:
		return "not implemented"
	}

	return strconv.FormatInt(
		int64(c.n),
		10,
	)
}
