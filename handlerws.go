package protean

import (
	"context"
	"net/http"
	"time"

	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/gorilla/websocket"
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

	m, u := resolveWebSocketProtocol(conn)

	ws := &webSocket{
		services:    h.services,
		conn:        conn,
		marshaler:   m,
		unmarshaler: u,
		interceptor: h.interceptor,
		inputCap:    10, // TODO: make this configurable
		outputCap:   10, // TODO: make this configurable
	}

	if err := ws.Serve(ctx); err != nil {
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseProtocolError,
				err.Error(),
			),
			time.Now().Add(1*time.Second),
		)
	}
}

// resolveWebSocketProtocol resolves a websocket sub-protocol to the MIME
// media-type it implies, and returns the marshaler/unmarshaler to use for that
// media type.
//
// It panics if the sub-protocol is not supported, as it should have already
// been negotiated as part of the websocket upgrade process.
func resolveWebSocketProtocol(conn *websocket.Conn) (protomime.Marshaler, protomime.Unmarshaler) {
	p := conn.Subprotocol()
	if p == "" {
		// Fall back to the default.
		//
		// An empty string means negotation failed, but gorilla/websocket will
		// have sent the default Sec-WebSocket-Protocol we configured above
		// (when upgrading to a websocket).
		p = defaultWebSocketProtocol
	}

	mediaType, ok := protomime.MediaTypeFromWebSocketProtocol(p)
	if !ok {
		panic("unsupported websocket sub-protocol")
	}

	marshaler, ok := protomime.MarshalerForMediaType(mediaType)
	if !ok {
		panic("unsupported websocket sub-protocol")
	}

	unmarshaler, ok := protomime.UnmarshalerForMediaType(mediaType)
	if !ok {
		panic("unsupported websocket sub-protocol")
	}

	return marshaler, unmarshaler
}
