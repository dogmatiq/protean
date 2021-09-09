package handler

import (
	"net/http"

	"github.com/dogmatiq/harpy/runtime"
)

// handleSSE handles requests that consume server-streaming RPC calls using
// server-sent events.
//
// It is assumed that the client is a browser using the EventStream API, and
// therefore the RPC responses are encoded using JSON.
func (h *Handler) handleSSE(
	w http.ResponseWriter,
	r *http.Request,
	s runtime.Service,
	m runtime.Method,
) {
	// ensure get, parse query
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
