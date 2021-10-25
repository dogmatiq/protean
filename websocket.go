package protean

import (
	"context"
	"fmt"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/gorilla/websocket"
)

// webSocket manages communication for a single websocket connection.
//
// It handles the marshaling and unmarshaling of websocket envelopes and manages
// the lifetime of virtual channels established by the client.
type webSocket struct {
	Conn        *websocket.Conn
	Marshaler   protomime.Marshaler
	Unmarshaler protomime.Unmarshaler
}

// Serve serves RPC requests made via the websocket connection until ctx is
// canceled or an error occurs.
func (ws *webSocket) Serve(ctx context.Context) error {
	for {
		_, err := ws.read()
		if err != nil {
			return err
		}
	}
}

// read reads the next frame from the client.
func (ws *webSocket) read() (*proteanpb.ClientEnvelope, error) {
	_, data, err := ws.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	env := &proteanpb.ClientEnvelope{}
	if err := ws.Unmarshaler.Unmarshal(data, env); err != nil {
		return nil, newWebSocketError(
			websocket.CloseInvalidFramePayloadData,
			"unable to unmarshal frame: %s",
			err,
		)
	}

	return env, nil
}

// webSocketError encapsulates a client-facing error that occured while serving
// a websocket connection.
//
// It is returned from webSocket.Serve() when the error message is intended to
// be seen by the client.
type webSocketError struct {
	CloseCode int
	Reason    string
}

// newWebSocketError returns a new webSocketError with the given code and
// message.
func newWebSocketError(code int, format string, args ...interface{}) webSocketError {
	return webSocketError{
		code,
		fmt.Sprintf(format, args...),
	}
}

func (e webSocketError) Error() string {
	return e.Reason
}
