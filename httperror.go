package protean

import (
	"fmt"
	"net/http"
)

// httpError information aboubt an HTTP error to w.
func httpError(
	w http.ResponseWriter,
	status int,
	responseErr Error,
) {
	data, err := responseErr.MarshalText()
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
func httpStatusFromErrorCode(c ErrorCode) int {
	switch c {
	case ErrorCodeUnknown:
		return http.StatusInternalServerError
	case ErrorCodeInvalidInput:
		return http.StatusBadRequest
	case ErrorCodeUnauthenticated:
		return http.StatusUnauthorized
	case ErrorCodePermissionDenied:
		return http.StatusForbidden
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeAlreadyExists:
		return http.StatusConflict
	case ErrorCodeResourceExhausted:
		return http.StatusTooManyRequests
	case ErrorCodeFailedPrecondition:
		return http.StatusBadRequest
	case ErrorCodeAborted:
		return http.StatusConflict
	case ErrorCodeUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeNotImplemented:
		return http.StatusNotImplemented
	}

	return http.StatusInternalServerError
}
