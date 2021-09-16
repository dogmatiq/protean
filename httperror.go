package protean

import (
	"fmt"
	"net/http"

	"github.com/dogmatiq/protean/rpcerror"
)

// httpError information aboubt an HTTP error to w.
func httpError(
	w http.ResponseWriter,
	status int,
	rpcErr rpcerror.Error,
) {
	data, err := rpcErr.MarshalText()
	if err != nil {
		// The proteanpb.Error value itself can not be marshaled. This can only
		// fail if we've misconfigured the marshaler we're using (which are
		// hardcoded into this library), or the server is attempting to use the
		// "version 1" Go Protocol Buffers library, which is not supported.
		panic(fmt.Sprintf("unable to marshal error: %s", err))
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

// httpStatusFromErrorCode returns the default HTTP status to send when an error
// with the given code occurs.
func httpStatusFromErrorCode(c rpcerror.Code) int {
	switch c {
	case rpcerror.Unknown:
		return http.StatusInternalServerError
	case rpcerror.InvalidInput:
		return http.StatusBadRequest
	case rpcerror.Unauthenticated:
		return http.StatusUnauthorized
	case rpcerror.PermissionDenied:
		return http.StatusForbidden
	case rpcerror.NotFound:
		return http.StatusNotFound
	case rpcerror.AlreadyExists:
		return http.StatusConflict
	case rpcerror.ResourceExhausted:
		return http.StatusTooManyRequests
	case rpcerror.FailedPrecondition:
		return http.StatusBadRequest
	case rpcerror.Aborted:
		return http.StatusConflict
	case rpcerror.Unavailable:
		return http.StatusServiceUnavailable
	case rpcerror.NotImplemented:
		return http.StatusNotImplemented
	}

	return http.StatusInternalServerError
}
