package protean

import (
	"net/http"

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
