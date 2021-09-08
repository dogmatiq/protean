package harpy

import (
	"net/http"

	"github.com/dogmatiq/harpy/codegenapi"
)

// handleJSONRPC handles JSON-RPC requests made to a specific service.
func (h *Handler) handleJSONRPC(
	w http.ResponseWriter,
	r *http.Request,
	s codegenapi.Service,
) {
	// ensure post
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
