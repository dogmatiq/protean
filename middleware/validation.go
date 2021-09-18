package middleware

import (
	"context"

	"github.com/dogmatiq/protean/rpcerror"
	"google.golang.org/protobuf/proto"
)

// ValidatableMessage is an RPC input or output message that provides its own
// validation.
type ValidatableMessage interface {
	proto.Message

	// Validate returns an error if the message is invalid.
	//
	// The error message may be sent to the RPC client, and as such should not
	// contain any sensitive information.
	Validate() error
}

// Validator is an implementation of ServerInterceptor that validates RPC input
// and messages by calling their Validate() method, if present.
//
// The Validator interceptor is installed by default.
type Validator struct{}

// InterceptUnaryRPC returns an error if any RPC input or output message that
// implements ValidatableMessage is invalid.
func (Validator) InterceptUnaryRPC(
	ctx context.Context,
	info UnaryServerInfo,
	in proto.Message,
	next func(ctx context.Context) (out proto.Message, err error),
) (proto.Message, error) {
	if in, ok := in.(ValidatableMessage); ok {
		if err := in.Validate(); err != nil {
			return nil, rpcerror.New(
				rpcerror.InvalidInput,
				err.Error(),
			)
		}
	}

	out, err := next(ctx)
	if err != nil {
		return nil, err
	}

	if out, ok := out.(ValidatableMessage); ok {
		if err := out.Validate(); err != nil {
			return nil, rpcerror.New(
				rpcerror.Unknown,
				"the server produced invalid RPC output",
			).WithCause(err)
		}
	}

	return out, nil
}
