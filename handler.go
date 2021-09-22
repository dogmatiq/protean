package protean

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/middleware"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/dogmatiq/protean/runtime"
)

// Handler is an http.Handler that maps HTTP requests to RPC calls.
type Handler interface {
	http.Handler
	runtime.Registry
}

// handler is an implementation of Handler that handles RPC method calls made
// via HTTP POST requests and "method-scoped" websocket connections.
type handler struct {
	services map[string]runtime.Service
}

// HandlerOption is an option that changes the behavior of an HTTP handler.
type HandlerOption func(*handler)

// NewHandler returns a new HTTP handler that maps HTTP requests to RPC calls.
func NewHandler(options ...HandlerOption) Handler {
	h := &handler{}

	for _, opt := range options {
		opt(h)
	}

	return h
}

// RegisterService adds a service to this handler.
func (h *handler) RegisterService(s runtime.Service) {
	prefix := fmt.Sprintf(
		"%s.%s",
		s.Package(),
		s.Name(),
	)

	if h.services == nil {
		h.services = map[string]runtime.Service{}
	}

	h.services[prefix] = s
}

// ServeHTTP handles an HTTP request.
//
// The request must use the POST HTTP method.
//
// The request URL path is mapped to an RPC method using the following pattern:
// /<package>/<service>/<method>, where <package> is the Protocol Buffers
// package that contains the service definition, <service> is the service's
// name, and <method> is the name of the RPC method.
//
// The request body is the RPC input message, which is a Protocol Buffers
// message encoded in one of the following media types:
//   - application/vnd.google.protobuf (binary format, preferred)
//   - application/x-protobuf (equivalent to application/vnd.google.protobuf)
//   - application/json (as per google.golang.org/protobuf/encoding/protojson)
//   - text/plain (as per google.golang.org/protobuf/encoding/prototext)
//
// The RPC output message is written to the response body, encoded as per the
// request's Accept header, which need not be the same as the input encoding.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, method, ok := h.resolveMethod(w, r)
	if !ok {
		return
	}

	if method.InputIsStream() || method.OutputIsStream() {
		httpError(
			w,
			http.StatusNotImplemented,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.NotImplemented,
				"the '%s.%s' service does contain an RPC method named '%s', but is not supported by this server because it uses streaming inputs or outputs",
				service.Package(),
				service.Name(),
				method.Name(),
			),
		)
		return
	}

	// Set the Accept-Post header only once we've verified that the requested
	// method exists and is supported.
	w.Header().Set("Accept-Post", acceptPostHeader)

	if r.Method != http.MethodPost {
		httpError(
			w,
			http.StatusNotImplemented,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.NotImplemented,
				"the HTTP method must be POST",
			),
		)
		return
	}

	h.servePOST(w, r, method)
}

// resolveMethod looks up the RPC method based on the request URL.
//
// It returns false if the RPC method can not be found, in which case a 404 Not
// Found error has already been written to w.
func (h *handler) resolveMethod(
	w http.ResponseWriter,
	r *http.Request,
) (runtime.Service, runtime.Method, bool) {
	serviceName, methodName, ok := parsePath(r.URL.Path)
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.NotFound,
				"the request URI must follow the '/<package>/<service>/<method>' pattern",
			),
		)

		return nil, nil, false
	}

	service, ok := h.services[serviceName]
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.NotFound,
				"the server does not provide the '%s' service",
				serviceName,
			),
		)

		return nil, nil, false
	}

	method, ok := service.MethodByName(methodName)
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.NotFound,
				"the '%s' service does not contain an RPC method named '%s'",
				serviceName,
				methodName,
			),
		)

		return nil, nil, false
	}

	return service, method, true
}

// newRPCCall starts a new RPC call to the given method.
func (h *handler) newRPCCall(r *http.Request, method runtime.Method) runtime.Call {
	return method.NewCall(r.Context(), middleware.Validator{})
}

// parsePath parses the URI path p and returns the names of the service
// and method that it maps to.
func parsePath(p string) (service, method string, ok bool) {
	pkg, p, ok := nextPathSegment(p)
	if !ok {
		return "", "", false
	}

	service, p, ok = nextPathSegment(p)
	if !ok {
		return "", "", false
	}

	method, p, ok = nextPathSegment(p)
	if !ok {
		return "", "", false
	}

	// ensure there are no more segments
	_, _, ok = nextPathSegment(p)
	if !ok {
		return pkg + "." + service, method, true
	}

	return "", "", false
}

// nextPathSegment returns the next segment of the path p.
func nextPathSegment(p string) (seg, rest string, ok bool) {
	if p == "" {
		return "", "", false
	}

	p = p[1:] // trim leading slash
	if p == "" {
		return "", "", false
	}

	if i := strings.IndexByte(p, '/'); i != -1 {
		return p[:i], p[i:], true
	}

	return p, "", true
}

// httpError writes information about an HTTP error to w.
func httpError(
	w http.ResponseWriter,
	status int,
	mediaType string,
	marshaler protomime.Marshaler,
	rpcErr rpcerror.Error,
) {
	var protoErr proteanpb.Error
	if err := rpcerror.ToProto(rpcErr, &protoErr); err != nil {
		panic(err)
	}

	data, err := marshaler.Marshal(&protoErr)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", protomime.FormatMediaType(mediaType, &protoErr))
	w.Header().Add("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

// httpStatusFromErrorCode returns the default HTTP status to send when an error
// with the given code occurs.
func httpStatusFromErrorCode(c rpcerror.Code) int {
	switch c {
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
