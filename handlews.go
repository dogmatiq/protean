package protean

import (
	"net/http"

	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/dogmatiq/protean/runtime"
)

// serveWebSocket serves an RPC request made using a websocket.
func (h *handler) serveWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	method runtime.Method,
) {
	panic("not implemented")
}

// webSocketError writes information about an HTTP error that was produced by
// the websocket.Upgrader to w.
func webSocketError(
	w http.ResponseWriter,
	r *http.Request,
	code int,
	reason error,
) {
	httpError(
		w,
		http.StatusNotImplemented,
		protomime.TextMediaTypes[0],
		protomime.TextMarshaler,
		rpcerror.New(
			rpcerror.Unknown,
			reason.Error(),
		),
	)
}

// webSocketSubProtocols is the set of supported websocket subprotocols, in
// order of preference.
var webSocketSubProtocols = []string{
	"protean.v1+binary",
	"protean.v1+json",
	"protean.v1+text",
}
