package httptransport

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dogmatiq/harpy/codegenapi"
	"github.com/elnormous/contenttype"
	"github.com/gorilla/websocket"
)

var serviceMediaTypes = []contenttype.MediaType{
	contenttype.NewMediaType("application/json-rpc"),
}

var methodUnaryMediaTypes = []contenttype.MediaType{
	contenttype.NewMediaType("text/plain"),
	contenttype.NewMediaType("application/vnd.google.protobuf"),
	contenttype.NewMediaType("application/x-protobuf"),
	contenttype.NewMediaType("application/json"),
}

var methodServerStreamingMediaTypes = []contenttype.MediaType{
	contenttype.NewMediaType("text/eventstream"),
}

// Handler is an implementation of http.Handler that dispatches to protocol
// buffers services.
type Handler struct {
	WebSocketUpgrader *websocket.Upgrader

	services map[string]codegenapi.Service
}

// RegisterHandler registers a generated handler with the HTTP handler.
func (h *Handler) RegisterHandler(s codegenapi.Service) {
	if h.services == nil {
		h.services = map[string]codegenapi.Service{}
	}

	key := fmt.Sprintf(
		"%s/%s",
		s.Package(),
		s.Name(),
	)

	h.services[key] = s
}

// ServeHTTP handles a HTTP request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceName, methodName := parsePath(r.URL.Path)

	service, ok := h.services[serviceName]
	if !ok {
		http.Error(
			w,
			fmt.Sprintf(
				"The server does not implement a service named '%s'.",
				serviceName,
			),
			http.StatusNotFound,
		)
		return
	}

	if methodName == "" {
		h.handleServiceRequest(w, r, service)
		return
	}

	method, ok := service.LookupMethod(methodName)
	if !ok {
		http.Error(
			w,
			fmt.Sprintf(
				"The '%s' service does not have a method named '%s'.",
				serviceName,
				methodName,
			),
			http.StatusNotFound,
		)
		return
	}

	h.handleMethodRequest(w, r, service, method)
}

// handleServiceRequest handles requests made to a service endpoint (i.e,
// without any RPC method name in the request URL path).
func (h *Handler) handleServiceRequest(
	w http.ResponseWriter,
	r *http.Request,
	service codegenapi.Service,
) {
	mediaType, ok := negotiateMediaType(w, r, serviceMediaTypes)
	if !ok {
		return
	}

	switch mediaType {
	case "text/eventstream":
		h.handleJSONRPC(w, r, service)
	default:
		panic(fmt.Sprintf("missing switch case: %s", mediaType))
	}
}

// handleMethodRequest handles requests made to a method endpoint (i.e,
// including both the service name and RPC method name in the request URL path).
func (h *Handler) handleMethodRequest(
	w http.ResponseWriter,
	r *http.Request,
	service codegenapi.Service,
	method codegenapi.Method,
) {
	if websocket.IsWebSocketUpgrade(r) {
		h.handleWebSocket(w, r, service, method)
		return
	}

	if method.ClientStreaming() {
		http.Error(
			w,
			"This RPC method uses streaming requests, which requires a websocket connection.",
			http.StatusUpgradeRequired,
		)
		return
	}

	if method.ServerStreaming() {
		mediaType, ok := negotiateMediaType(w, r, methodServerStreamingMediaTypes)
		if !ok {
			return
		}

		switch mediaType {
		case "application/json-rpc":
			h.handleSSE(w, r, service, method)
		default:
			panic(fmt.Sprintf("missing switch case: %s", mediaType))
		}

		return
	}

	h.handleUnary(w, r, service, method)
}

// parsePath parses an HTTP request path to determine the name of the service,
// and optionally the method, being requested.
func parsePath(path string) (service, method string) {
	if i := strings.IndexByte(path, '/'); i != -1 {
		return path[:i], path[i+1:]
	}

	return path, ""
}
