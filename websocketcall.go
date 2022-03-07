package protean

import (
	"context"
	"time"

	"github.com/dogmatiq/protean/runtime"
	"github.com/gorilla/websocket"
)

// webSocketCall manages communication for a single RPC call made over a
// websocket connection.
type webSocketCall struct {
	ID              uint32
	Method          runtime.Method
	ProtocolTimeout time.Duration
}

// Serve a single this RPC request until ctx is canceled or an error occurs.
func (c *webSocketCall) Serve(ctx context.Context) error {
	if c.Method.InputIsStream() {
		return nil
	}

	time.Sleep(c.ProtocolTimeout)

	return newWebSocketError(
		websocket.CloseProtocolError,
		"expected 'send' frame within %s of 'call' frame (%d)",
		c.ProtocolTimeout,
		c.ID,
	)
}
