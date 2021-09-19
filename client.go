package protean

import (
	"net/http"

	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/runtime"
)

// ClientOption is an option that changes the behavior of an RPC client.
type ClientOption func(*runtime.ClientOptions)

// WithHTTPClient is a ClientOption that sets the HTTP client to use when making
// RPC requests.
func WithHTTPClient(c *http.Client) ClientOption {
	return func(options *runtime.ClientOptions) {
		options.HTTPClient = c
	}
}

// WithMediaType is a ClientOption that sets the preferred media-type to use for
// both RPC input and output messages.
func WithMediaType(mediaType string) ClientOption {
	if !protomime.IsSupportedMediaType(mediaType) {
		panic("unsupported media type")
	}

	return func(options *runtime.ClientOptions) {
		options.InputMediaType = mediaType
		options.OutputMediaType = mediaType
	}
}

// WithInputMediaType is a ClientOption that sets the preferred media-type that
// the client should use when encoding RPC input messages in HTTP requests.
func WithInputMediaType(mediaType string) ClientOption {
	if !protomime.IsSupportedMediaType(mediaType) {
		panic("unsupported media type")
	}

	return func(options *runtime.ClientOptions) {
		options.InputMediaType = mediaType
	}
}

// WithOutputMediaType is a ClientOption that sets the preferred media-type that
// the server should use when encoding RPC output messages in HTTP responses.
func WithOutputMediaType(mediaType string) ClientOption {
	if !protomime.IsSupportedMediaType(mediaType) {
		panic("unsupported media type")
	}

	return func(options *runtime.ClientOptions) {
		options.OutputMediaType = mediaType
	}
}
