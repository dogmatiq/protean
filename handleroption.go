package protean

import "time"

const (
	// DefaultMaxRPCInputSize is the default maximum size for RPC input
	// messages.
	//
	// The default is a conservative value of 1 megabyte. It can be overriden
	// with the WithHeatbeatInterval() handler option.
	DefaultMaxRPCInputSize = 1_000_000

	// DefaultWebSocketProtocolTimeout is the maximum amount of time the handler
	// will wait for a mandatory websocket frame for the client before closing
	// the websocket connection.
	//
	// The default timeout is fairly generous, but it may be desirable to
	// increase this timeout if the API is expected to be called over
	// high-latency or high-packet-loss connections.
	DefaultWebSocketProtocolTimeout = 500 * time.Millisecond
)

// HandlerOption is an option that changes the behavior of an HTTP handler.
type HandlerOption func(*handler)

// WithMaxRPCInputSize is a HandlerOption that sets the maximum size of RPC
// input messages that the handler will accept, in bytes.
//
// If this option is not provided, DefaultMaxRPCInputSize is used.
func WithMaxRPCInputSize(n int) HandlerOption {
	if n <= 0 {
		panic("maximum input size must be postive")
	}

	return func(h *handler) {
		h.maxInputSize = n
	}
}

// WithWebSocketProtocolTimeout is a HandlerOption that sets the maximum amount
// of time the handler will wait for a mandatory websocket frame for the client
// before closing the websocket connection.
//
// If this option is not provided, DefaultWebSocketProtocolTimeout is used.
func WithWebSocketProtocolTimeout(t time.Duration) HandlerOption {
	if t <= 0 {
		panic("timeout duration postive")
	}

	return func(h *handler) {
		h.webSocketProtocolTimeout = t
	}
}
