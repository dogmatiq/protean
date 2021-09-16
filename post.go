package protean

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
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
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	serviceName, methodName, ok := parsePath(r.URL.Path)
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			rpcerror.New(
				rpcerror.NotFound,
				"The request URI must follow the '/<package>/<service>/<method>' pattern.",
			),
		)
		return
	}

	service, ok := h.services[serviceName]
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			rpcerror.New(
				rpcerror.NotFound,
				"The server does not provide the '%s' service.",
				serviceName,
			),
		)
		return
	}

	method, ok := service.MethodByName(methodName)
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			rpcerror.New(
				rpcerror.NotFound,
				"The '%s' service does not contain an RPC method named '%s'.",
				serviceName,
				methodName,
			),
		)
		return
	}

	if method.InputIsStream() || method.OutputIsStream() {
		httpError(
			w,
			http.StatusNotImplemented,
			rpcerror.New(
				rpcerror.NotImplemented,
				"The '%s' service does contain an RPC method named '%s', but is not supported by this server because it uses streaming inputs or outputs.",
				serviceName,
				methodName,
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
			rpcerror.New(
				rpcerror.NotImplemented,
				"The HTTP method must be POST.",
			),
		)
		return
	}

	unmarshaler, inputMediaType, ok, err := unmarshalerByNegotiation(r)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			rpcerror.New(
				rpcerror.Unknown,
				"The Content-Type header is missing or invalid.",
			),
		)
		return
	}
	if !ok {
		httpError(
			w,
			http.StatusUnsupportedMediaType,
			rpcerror.New(
				rpcerror.Unknown,
				"The server does not support the '%s' media-type supplied by the client.",
				inputMediaType,
			).WithDetails(
				&proteanpb.SupportedMediaTypes{
					MediaTypes: protomime.MediaTypes,
				},
			),
		)
		return
	}

	marshaler, outputMediaType, ok, err := marshalerByNegotiation(r)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			rpcerror.New(
				rpcerror.Unknown,
				"The Accept header is invalid.",
			),
		)
		return
	}
	if !ok {
		httpError(
			w,
			http.StatusNotAcceptable,
			rpcerror.New(
				rpcerror.Unknown,
				"The client does not accept any of the media-types supported by the server.",
			).WithDetails(
				&proteanpb.SupportedMediaTypes{
					MediaTypes: protomime.MediaTypes,
				},
			),
		)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(
			w,
			http.StatusInternalServerError,
			rpcerror.New(
				rpcerror.Unknown,
				"The request body could not be read.",
			),
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
			rpcerror.New(
				rpcerror.Unknown,
				"The RPC input message could not be unmarshaled from the request body.",
			),
		)
		return
	}

	out, _, err := call.Recv()
	if err != nil {
		if err, ok := err.(rpcerror.Error); ok {
			httpError(
				w,
				httpStatusFromErrorCode(err.Code()),
				err,
			)
		} else {
			httpError(
				w,
				http.StatusInternalServerError,
				rpcerror.New(
					rpcerror.Unknown,
					"The RPC method returned an unrecognized error.",
				),
			)
		}

		return
	}

	data, err = marshaler.Marshal(out)
	if err != nil {
		httpError(
			w,
			http.StatusInternalServerError,
			rpcerror.New(
				rpcerror.Unknown,
				"The RPC output message could not be marshaled to the response body.",
			),
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
