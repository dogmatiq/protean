package protomime

import "strings"

// WebSocketProtocols is the set of supported websocket sub-protocols, in
// order of preference.
var WebSocketProtocols []string

// webSocketProtocolPrefix is the prefix to add to the supported media types to
// produce the websocket sub-protocol name.
const webSocketProtocolPrefix = "protean.v1"

// MediaTypeFromWebSocketProtocol returns the MIME media-type implied by the
// given websocket sub-protocol.
//
// It returns false if p is not a well-formed Protean protocol name.
func MediaTypeFromWebSocketProtocol(p string) (string, bool) {
	n := strings.IndexByte(p, '+')
	if n == -1 {
		return "", false
	}

	if p[:n] != webSocketProtocolPrefix {
		return "", false
	}

	p = p[n+1:]

	n = strings.IndexByte(p, '.')
	if n == -1 {
		return "", false
	}

	return p[:n] + "/" + p[n+1:], true
}

// WebSocketProtocolFromMediaType returns the websocket sub-protocol name to use
// to transport messages of the given media type.
func WebSocketProtocolFromMediaType(mediaType string) string {
	return webSocketProtocolPrefix + "+" + strings.Replace(mediaType, "/", ".", -1)
}

func init() {
	for _, mediaType := range MediaTypes {
		WebSocketProtocols = append(
			WebSocketProtocols,
			WebSocketProtocolFromMediaType(mediaType),
		)
	}
}
