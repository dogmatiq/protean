package protean

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"

	"github.com/dogmatiq/protean/runtime"
	"google.golang.org/protobuf/proto"
)

// PostHandler is an http.Handler that handles RPC calls made by posting to an
// RPC method endpoint.
type PostHandler struct {
	serviceByPath map[string]runtime.Service
}

var _ runtime.Registry = (*PostHandler)(nil)

// RegisterService adds a service to this handler.
func (h *PostHandler) RegisterService(s runtime.Service) {
	prefix := fmt.Sprintf(
		"/%s/%s/",
		s.Package(),
		s.Name(),
	)

	if h.serviceByPath == nil {
		h.serviceByPath = map[string]runtime.Service{}
	}

	h.serviceByPath[prefix] = s
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
//   - application/vnd.google.protobuf (native binary format, preferred)
//   - application/x-protobuf (equivalent to application/vnd.google.protobuf)
//   - application/json (as per google.golang.org/protobuf/encoding/protojson)
//   - text/plain (as per google.golang.org/protobuf/encoding/prototext)
//
// The RPC output message is written to the response body, encoded as per the
// request's Accept header, which need not be the same as the input encoding.
func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(
			w,
			http.StatusNotImplemented,
			"The request must use the POST HTTP method.",
		)
		return
	}

	servicePath, methodName := path.Split(r.URL.Path)

	service, ok := h.serviceByPath[servicePath]
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			"The server does not provide a service named '%s'.",
			servicePath[1:],
		)
		return
	}

	method, ok := service.MethodByName(methodName)
	if !ok {
		httpError(
			w,
			http.StatusNotFound,
			"The '%s' service does not contain an RPC method named '%s'.",
			servicePath[1:],
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
	}

	inputMediaType := r.Header.Get("Content-Type")
	inputMediaType, _, err := mime.ParseMediaType(inputMediaType)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			"The Content-Type header specifies an invalid media type.",
		)
		return
	}

	unmarshaler, ok := unmarshalerByMediaType(inputMediaType)
	if !ok {
		httpErrorUnsupportedMedia(w, protoMediaTypes)
		return
	}

	marshaler, outputMediaType, ok := marshalerByAcceptHeaders(r)
	if !ok {
		httpErrorNotAcceptable(w, protoMediaTypes)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			"Unable to read request body.",
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
			"Unable to parse input message from request body.",
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
			"Unable to marshal the RPC output message.",
		)
		return
	}

	w.Header().Add("Content-Type", outputMediaType)
	w.Header().Add("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(data)
}
