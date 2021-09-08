package handler

import (
	"net/http"

	"github.com/dogmatiq/harpy/codegenapi"
)

// jsonRPCHandler is an implementation of http.Handler that handles JSON-RPC
// requests for all methods within a single service.
type jsonRPCHandler struct {
	Service codegenapi.Service
}

func (h *jsonRPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ensure post
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
