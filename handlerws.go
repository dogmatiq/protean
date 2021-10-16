package protean

import (
	"context"
	"net/http"
	"time"

	"github.com/dogmatiq/protean/internal/protomime"
)

// defaultWebSocketProtocol is the websocket sub-protocol to use when the client
// either does not request a specific protocol, or requests an unsupported
// protocol.
var defaultWebSocketProtocol = protomime.WebSocketProtocolFromMediaType(
	protomime.JSONMediaTypes[0],
)

// serveWebSocket services a websocket request.
func (h *handler) serveWebSocket(w http.ResponseWriter, r *http.Request) {
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

	extendDeadline := func() {
		dl := time.Now().Add(
			time.Duration(
				float64(h.heartbeat) * 1.2,
			),
		)

		_ = conn.SetReadDeadline(dl)
		_ = conn.SetWriteDeadline(dl)
	}

	extendDeadline()

	// TODO: is r.Context() still applicable for a websocket connection?
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Force close the connection if the context is canceled.
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// Cancel the context if the connection is closed first.
	conn.SetCloseHandler(
		func(int, string) error {
			cancel()
			return nil
		},
	)

	// Extend the deadline whenever we get a PING message.
	conn.SetPongHandler(
		func(string) error {
			extendDeadline()
			return nil
		},
	)

	conn.SetReadLimit(int64(h.maxInputSize))
}
