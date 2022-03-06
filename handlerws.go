package protean

import (
	"net/http"
	"time"

	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/gorilla/websocket"
)

// serveWebSocket serves an HTTP connection that establishes a websocket
// connection.
//
// The websocket can be used to make multiple concurrent RPC calls using the
// framing formats defined in the proteanpb package.
func (h *handler) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := webSocketUpgrader.Upgrade(w, r, http.Header{
		// Specify the default websocket protocol in the response headers. This
		// response header is overridden if the client requests one of the other
		// supported protocols.
		"Sec-WebSocket-Protocol": {defaultWebSocketProtocol},
	})
	if err != nil {
		// This error has already been written to w via the upgrader.Error()
		// callback function.
		return
	}
	defer conn.Close()

	m, u := resolveWebSocketProtocol(conn)
	ws := &webSocket{
		Conn:            conn,
		Marshaler:       m,
		Unmarshaler:     u,
		ProtocolTimeout: h.webSocketProtocolTimeout,
	}

	// TODO: is r.Context() still appropriate now that the connection has been
	// hijacked?
	if err := ws.Serve(r.Context()); err != nil {
		code := websocket.CloseInternalServerErr
		reason := "internal server error"

		if err, ok := err.(webSocketError); ok {
			code = err.CloseCode
			reason = err.Reason
		}

		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason),
			time.Now().Add(1*time.Second),
		)
	}
}

var (
	// webSocketUpgrader is the upgrader used to hijack an HTTP request when
	// establishing a websocket connection.
	webSocketUpgrader = websocket.Upgrader{
		Subprotocols: protomime.WebSocketProtocols,
		Error: func(
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
		},
		EnableCompression: true,
	}

	// defaultWebSocketProtocol is the websocket sub-protocol to use when the client
	// either does not request a specific protocol, or requests an unsupported
	// protocol.
	defaultWebSocketProtocol = protomime.WebSocketProtocolFromMediaType(
		protomime.JSONMediaTypes[0],
	)
)

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
