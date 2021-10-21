package protean

const (
	// DefaultMaxRPCInputSize is the default maximum size for RPC input
	// messages.
	//
	// The default is a conservative value of 1 megabyte. It can be overriden
	// with the WithHeatbeatInterval() handler option.
	DefaultMaxRPCInputSize = 1_000_000
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
