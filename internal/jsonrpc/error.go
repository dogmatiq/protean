package jsonrpc

const (
	// ErrParseError is the JSON-RPC error code used when the client supplies a
	// request that is not valid JSON.
	ErrParseError = -32700

	// ErrInvalidRequest is the JSON-RPC error code used when the client
	// supplies a valid JSON, but that JSON to does not conform to the schema of
	// a JSON-RPC request object.
	ErrInvalidRequest = -32600

	// ErrMethodNotFound is the JSON-RPC error code used when the client
	// attempts to invoke an RPC method that does not exist.
	ErrMethodNotFound = -32601

	// ErrInvalidParams is the JSON-RPC error code used when the client supplies
	// a valid JSON-RPC request, but it's content is not considered valid for
	// the method being called.
	ErrInvalidParams = -32602

	// ErrInternalError is the JSON-RPC error code used when some unspecified
	// internal server error occurs.
	ErrInternalError = -32603
)
