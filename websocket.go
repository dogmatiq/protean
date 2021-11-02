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

	minCallID uint32
}

// Serve serves RPC requests made via the websocket connection until ctx is
// canceled or an error occurs.
func (ws *webSocket) Serve(ctx context.Context) error {
	for {
		env, err := ws.read()
		if err != nil {
			return err
		}

		if err := ws.handle(env); err != nil {
			return err
		}
	}
}

// handle processes a data frame received from the client.
func (ws *webSocket) handle(env *proteanpb.ClientEnvelope) error {
	switch fr := env.Frame.(type) {
	case *proteanpb.ClientEnvelope_Call:
		return ws.handleCall(env.CallId, fr)
	case *proteanpb.ClientEnvelope_Send:
		return ws.handleSend(env.CallId, fr)
	case *proteanpb.ClientEnvelope_Close:
		return ws.handleClose(env.CallId, fr)
	default:
		return newWebSocketError(
			websocket.CloseProtocolError,
			"unrecognized frame type",
		)
	}
}

// handleCall handles a "call" frame.
func (ws *webSocket) handleCall(id uint32, fr *proteanpb.ClientEnvelope_Call) error {
	if id < ws.minCallID {
		return newWebSocketError(
			websocket.CloseProtocolError,
			"out-of-sequence call ID in 'call' frame (%d), expected >=%d",
			id,
			ws.minCallID,
		)
	}

	ws.minCallID = id + 1

	return nil
}

// handleSend handles a "send" frame.
func (ws *webSocket) handleSend(id uint32, fr *proteanpb.ClientEnvelope_Send) error {
	if id >= ws.minCallID {
		return newWebSocketError(
			websocket.CloseProtocolError,
			"out-of-sequence call ID in 'send' frame (%d), expected <%d",
			id,
			ws.minCallID,
		)
	}

	return nil
}

// handleClose handles a "close" frame.
func (ws *webSocket) handleClose(id uint32, fr *proteanpb.ClientEnvelope_Close) error {
	if !fr.Close {
		return nil
	}

	if id >= ws.minCallID {
		return newWebSocketError(
			websocket.CloseProtocolError,
			"out-of-sequence call ID in 'close' frame (%d), expected <%d",
			id,
			ws.minCallID,
		)
	}

	return nil
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
			"%s",
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
