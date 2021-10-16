package protean

import (
	"net/http"
	"time"

	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/runtime"
	"github.com/gorilla/websocket"
)

// defaultWebSocketProtocol is the websocket sub-protocol to use when the client
// either does not request a specific protocol, or requests an unsupported
// protocol.
var defaultWebSocketProtocol = protomime.WebSocketProtocolFromMediaType(
	protomime.JSONMediaTypes[0],
)

// serveWebSocket serves an RPC request made using a websocket.
func (h *handler) serveWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	method runtime.Method,
) {
	conn, err := h.upgrader.Upgrade(
		w,
		r,
		http.Header{
			"Sec-WebSocket-Protocol": {defaultWebSocketProtocol},
		},
	)
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
}

// extendWebSocketReadDeadline extends the read deadlone of conn according to
// the configured heartbeat interval.
func (h *handler) extendWebSocketReadDeadline(conn *websocket.Conn) {
	// Heartbeat interval is increased by 20%, allowing higher-latency clients
	// some leeway without having to send PING messages more often than the
	// h.heartbeat interval.
	_ = conn.SetReadDeadline(
		time.Now().Add(
			time.Duration(
				float64(h.heartbeat) * 1.2,
			),
		),
	)
}