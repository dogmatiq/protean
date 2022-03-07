package protean

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/runtime"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

// webSocket manages communication for a single websocket connection.
//
// It handles the marshaling and unmarshaling of websocket envelopes and manages
// the RPC calls made by the client.
type webSocket struct {
	Conn            *websocket.Conn
	Services        map[string]runtime.Service
	Marshaler       protomime.Marshaler
	Unmarshaler     protomime.Unmarshaler
	ProtocolTimeout time.Duration

	minCallID uint32
	calls     *errgroup.Group
}

// Serve serves RPC requests made via the websocket connection until ctx is
// canceled or an error occurs.
func (ws *webSocket) Serve(ctx context.Context) error {
	// The read-loop is NOT run within the same errgroup as the calls because we
	// have no way to unblock it without closing the connection.
	//
	// Instead we let Serve() return when all the calls are done and let
	// readLoop() run until the websocket connection is closed. The readErr
	// channel is used to terminate Serve() if the read-loop itself fails.
	read := make(chan *proteanpb.ClientEnvelope)
	readErr := make(chan error, 1)
	go func() {
		readErr <- ws.readLoop(ctx, read)
	}()

	ws.calls, ctx = errgroup.WithContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return ws.calls.Wait()

		case env := <-read:
			if err := ws.handle(ctx, env); err != nil {
				return err
			}

		case err := <-readErr:
			if err != nil {
				return err
			}
		}
	}
}

// handle processes a data frame received from the client.
func (ws *webSocket) handle(
	ctx context.Context,
	env *proteanpb.ClientEnvelope,
) error {
	switch fr := env.Frame.(type) {
	case *proteanpb.ClientEnvelope_Call:
		return ws.handleCall(ctx, env.CallId, fr)
	case *proteanpb.ClientEnvelope_Send:
		return ws.handleSend(ctx, env.CallId, fr)
	case *proteanpb.ClientEnvelope_Close:
		return ws.handleClose(ctx, env.CallId, fr)
	case *proteanpb.ClientEnvelope_Cancel:
		return ws.handleCancel(ctx, env.CallId, fr)
	default:
		return newWebSocketError(
			websocket.CloseProtocolError,
			"unrecognized frame type",
		)
	}
}

// handleCall handles a "call" frame.
func (ws *webSocket) handleCall(
	ctx context.Context,
	id uint32,
	fr *proteanpb.ClientEnvelope_Call,
) error {
	if id < ws.minCallID {
		return newWebSocketError(
			websocket.CloseProtocolError,
			"out-of-sequence call ID in 'call' frame (%d), expected >=%d",
			id,
			ws.minCallID,
		)
	}

	ws.minCallID = id + 1

	ws.calls.Go(func() error {
		c := &webSocketCall{
			Conn:            ws.Conn,
			Services:        ws.Services,
			Marshaler:       ws.Marshaler,
			ProtocolTimeout: ws.ProtocolTimeout,
			ID:              id,
			MethodName:      fr.Call,
		}

		return c.Serve(ctx)
	})

	return nil
}

// handleSend handles a "send" frame.
func (ws *webSocket) handleSend(
	ctx context.Context,
	id uint32,
	fr *proteanpb.ClientEnvelope_Send,
) error {
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
func (ws *webSocket) handleClose(
	ctx context.Context,
	id uint32,
	fr *proteanpb.ClientEnvelope_Close,
) error {
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

// handleCancel handles a "cancel" frame.
func (ws *webSocket) handleCancel(
	ctx context.Context,
	id uint32,
	fr *proteanpb.ClientEnvelope_Cancel,
) error {
	if !fr.Cancel {
		return nil
	}

	if id >= ws.minCallID {
		return newWebSocketError(
			websocket.CloseProtocolError,
			"out-of-sequence call ID in 'cancel' frame (%d), expected <%d",
			id,
			ws.minCallID,
		)
	}

	return nil
}

// read reads message from the websocket and pipes them to the ws.in channel.
func (ws *webSocket) readLoop(
	ctx context.Context,
	envelopes chan<- *proteanpb.ClientEnvelope,
) error {
	for {
		env, err := ws.readNext()
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case envelopes <- env:
			continue // shows test coverage
		}
	}
}

// readNext reads the next frame from the client.
func (ws *webSocket) readNext() (*proteanpb.ClientEnvelope, error) {
	_, data, err := ws.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	env := &proteanpb.ClientEnvelope{}
	if err := ws.Unmarshaler.Unmarshal(data, env); err != nil {
		return nil, newWebSocketError(
			websocket.CloseInvalidFramePayloadData,
			"could not unmarshal envelope",
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
//
// The message is packed into a websocket control frame, and as such has a
// length limit of 120 bytes.
func newWebSocketError(code int, format string, args ...interface{}) webSocketError {
	reason := fmt.Sprintf(format, args...)
	if len(reason) > 120 {
		panic("websocket error message is too long")
	}

	return webSocketError{
		code,
		reason,
	}
}

func (e webSocketError) Error() string {
	return e.Reason
}
