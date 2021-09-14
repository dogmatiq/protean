package protean

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dogmatiq/protean/runtime"
	"google.golang.org/protobuf/proto"
)

// PostHandler is an http.Handler that handles RPC calls made by posting to an
// RPC method endpoint.
type PostHandler struct {
	services map[string]runtime.Service
}

var _ runtime.Registry = (*PostHandler)(nil)

// RegisterService adds a service to this handler.
func (h *PostHandler) RegisterService(s runtime.Service) {
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
func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceName, methodName, ok := parsePath(r.URL.Path)
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			"The request URI must follow the '/<package>/<service>/<method>' pattern.",
		)
		return
	}

	service, ok := h.services[serviceName]
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			"The server does not provide the '%s' service.",
			serviceName,
		)
		return
	}

	method, ok := service.MethodByName(methodName)
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			"The '%s' service does not contain an RPC method named '%s'.",
			serviceName,
			methodName,
		)
		return
	}

	if method.InputIsStream() || method.OutputIsStream() {
		httpError(
			w,
			http.StatusNotFound,
			"An RPC method named '%s' exists, but is not supported by this server because it uses streaming inputs or outputs.",
			methodName,
		)
		return
	}

	if r.Method != http.MethodPost {
		httpError(
			w,
			http.StatusNotImplemented,
			"The HTTP method must be POST.",
		)
		return
	}

	unmarshaler, inputMediaType, ok, err := unmarshalerByNegotiation(r)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			"The Content-Type header is missing or invalid.",
		)
		return
	}
	if !ok {
		httpErrorUnsupportedMedia(w, inputMediaType, protoMediaTypes)
		return
	}

	marshaler, outputMediaType, ok, err := marshalerByNegotiation(r)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			"The Accept header is invalid.",
		)
		return
	}
	if !ok {
		httpErrorNotAcceptable(w, protoMediaTypes)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(
			w,
			http.StatusInternalServerError,
			"The request body could not be read.",
		)
		return
	}

	call := method.NewCall(r.Context())
	defer call.Done()

	// Send never blocks on unary RPC methods.
	if _, err := call.Send(func(m proto.Message) error {
		return unmarshaler.Unmarshal(data, m)
	}); err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			"The RPC input message could not be unmarshaled from the request body.",
		)
		return
	}

	out, _, err := call.Recv()
	if err != nil {
		// TODO: we need an error system!
		httpError(
			w,
			http.StatusInternalServerError,
			"The RPC method produced an unrecognized error.",
		)
		return
	}

	data, err = marshaler.Marshal(out)
	if err != nil {
		httpError(
			w,
			http.StatusInternalServerError,
			"The RPC output message could not be marshaled to the response body.",
		)
		return
	}

	w.Header().Add("Content-Type", outputMediaType)
	w.Header().Add("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(data)
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