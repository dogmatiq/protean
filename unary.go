package harpy

import (
	"fmt"
	"net/http"

	"github.com/dogmatiq/harpy/codegenapi"
)

// handleUnary handles non-streaming events that are made using traditional GET
// or POST requests.
//
// Note that the HTTP method is not necessarily used to distinguish between read
// and write operations (as protocol buffers services do not distinguish between
// the two). However, use of the GET method for write operations is generally
// discouraged.
func (h *Handler) handleUnary(
	w http.ResponseWriter,
	r *http.Request,
	s codegenapi.Service,
	m codegenapi.Method,
) {
	mediaType, ok := negotiateMediaType(w, r, methodUnaryMediaTypes)
	if !ok {
		return
	}

	switch mediaType {
	case "text/plain":
	case "application/vnd.google.protobuf", "application/x-protobuf":
	case "application/json":
	default:
		panic(fmt.Sprintf("missing switch case: %s", mediaType))
	}

	// allow get, parse query
	// allow post
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
