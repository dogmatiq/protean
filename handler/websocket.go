package handler

import (
	"net/http"

	"github.com/dogmatiq/harpy/runtime"
)

// webSocketHandler is an implementation of http.Handler that handles
// websocket-based requests for a specific RPC method.
//
// A websocket MUST be used for RPC methods that use client streaming (streams
// of requests), and MAY be used for any other RPC type.
//
// If an RPC call that does NOT use client streaming receives multiple requests
// over the websocket, the connection is closed by the server.
//
// If the RPC call is unary (no streaming involved), the connection is closed
// after the response is sent.
type webSocketHandler struct {
	Service runtime.Service
	Method  runtime.Method
}

func (h *webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
