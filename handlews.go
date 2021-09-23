package protean

import (
	"net/http"
	"time"

	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/dogmatiq/protean/runtime"
	"github.com/gorilla/websocket"
)

// serveWebSocket serves an RPC request made using a websocket.
func (h *handler) serveWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	method runtime.Method,
) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Error has already been written to w via h.upgrader.Error.
		return
	}
	defer conn.Close()

	h.extendWebSocketReadDeadline(conn)

	conn.SetPongHandler(
		func(string) error {
			h.extendWebSocketReadDeadline(conn)
			return nil
		},
	)

	conn.SetReadLimit(int64(h.maxInputSize))

	panic("not implemented")
}

// extendWebSocketReadDeadline extends the read deadlone of conn according to
// the configured heartbeat interval.
func (h *handler) extendWebSocketReadDeadline(conn *websocket.Conn) {
	// Heartbeat interval is increased by 20%, allowing higher-latency clients
	// some leeway without having to send PING messages more often than the
	// h.heartbeat interval.
	conn.SetReadDeadline(
		time.Now().Add(
			time.Duration(
				float64(h.heartbeat) * 1.2,
			),
		),
	)
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
		code,
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
