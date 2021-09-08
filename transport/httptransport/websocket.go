package httptransport

import (
	"net/http"

	"github.com/dogmatiq/harpy/codegenapi"
)

// handleWebSocket handles requests to upgrade to a websocket to invoke a
// specific method.
//
// A websocket is required for RPC methods that use client streaming (streams of
// requests), and MAY be used for any other RPC type.
//
// If an RPC call that does NOT use client streaming receives multiple requests
// over the websocket, the connection is closed by the server.
//
// If the RPC call is unary (no streaming involved), the connection is closed
// after the response is sent.
func (h *Handler) handleWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	s codegenapi.Service,
	m codegenapi.Method,
) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
