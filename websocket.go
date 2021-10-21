package protean

import (
	"context"
	"fmt"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/middleware"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/dogmatiq/protean/runtime"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// webSocket manages communication for a single websocket connection.
//
// It handles the marshaling and unmarshaling of websocket frames and manages
// the lifetime of virtual channels established by the client.
type webSocket struct {
	services    map[string]runtime.Service
	conn        *websocket.Conn
	marshaler   protomime.Marshaler
	unmarshaler protomime.Unmarshaler
	interceptor middleware.ServerInterceptor
	inputCap    int
	outputCap   int

	in       chan *proteanpb.ClientEnvelope
	out      chan *proteanpb.ServerEnvelope
	channels map[uint32]*webSocketChannel
	group    *errgroup.Group
}

// webSocketChannel handles communication for a virtual websocket channel.
//
// Each channel is identified by an integer identifier supplied by the client
// and encapsulates a single RPC method invocation. It is not to be confused
// with a Go channel.
type webSocketChannel struct {
	cancel context.CancelFunc
	call   runtime.Call
	done   bool
}

// Serve communicates over the websocket until the client disconnects or an
// error occurs.
func (ws *webSocket) Serve(ctx context.Context) error {
	ws.in = make(chan *proteanpb.ClientEnvelope)
	ws.out = make(chan *proteanpb.ServerEnvelope)
	ws.channels = map[uint32]*webSocketChannel{}
	ws.group, ctx = errgroup.WithContext(ctx)

	ws.group.Go(func() error {
		return ws.readLoop(ctx)
	})

	for {
		select {
		case <-ctx.Done():
			return ws.group.Wait()
		case env := <-ws.in:
			if err := ws.handle(ctx, env); err != nil {
				return err
			}
		case env := <-ws.out:
			if err := ws.writeEnvelope(env); err != nil {
				return err
			}
		}
	}
}

func (ws *webSocket) handle(ctx context.Context, env *proteanpb.ClientEnvelope) error {
	switch fr := env.GetFrame().(type) {
	case *proteanpb.ClientEnvelope_Call:
		return ws.handleCall(ctx, env.Channel, fr)
	case *proteanpb.ClientEnvelope_Input:
		return ws.handleInput(ctx, env.Channel, fr)
	case *proteanpb.ClientEnvelope_Done:
		return ws.handleDone(ctx, env.Channel, fr)
	case *proteanpb.ClientEnvelope_Cancel:
		return ws.handleCancel(ctx, env.Channel, fr)
	default:
		return fmt.Errorf("unrecognised frame type on channel %d", env.Channel)
	}
}

// handleCall handles a "call" frame received from the client on this channel.
func (ws *webSocket) handleCall(ctx context.Context, channel uint32, fr *proteanpb.ClientEnvelope_Call) error {
	if _, ok := ws.channels[channel]; ok {
		return fmt.Errorf("unexpected call frame, channel %d is already open", channel)
	}

	method, err := ws.resolveMethod(fr.Call)
	if err != nil {
		env, err := newErrorFrame(channel, err)
		if err != nil {
			return err
		}

		return ws.writeEnvelope(env)
	}

	callCtx, cancel := context.WithCancel(ctx)
	call := method.NewCall(callCtx, runtime.CallOptions{
		Interceptor:           ws.interceptor,
		InputChannelCapacity:  ws.inputCap,
		OutputChannelCapacity: ws.outputCap,
	})

	ws.channels[channel] = &webSocketChannel{
		cancel: cancel,
		call:   call,
	}

	ws.group.Go(func() error {
		return ws.sendLoop(ctx, channel, call)
	})

	return nil
}

// handleInput handles a "input" frame received from the client on this channel.
func (ws *webSocket) handleInput(ctx context.Context, channel uint32, fr *proteanpb.ClientEnvelope_Input) error {
	ch, ok := ws.channels[channel]
	if !ok {
		return fmt.Errorf("unexpected input frame, channel %d is not open", channel)
	}

	if ch.done {
		return fmt.Errorf("unexpected input frame, channel %d does not expect additional input messages", channel)
	}

	more, err := ch.call.Send(
		func(m proto.Message) error {
			return fr.Input.UnmarshalTo(m)
		},
	)
	if err != nil {
		return err
	}

	if !more {
		ch.done = true
	}

	return nil
}

// handleDone handles a "done" frame received from the client on this channel.
func (ws *webSocket) handleDone(ctx context.Context, channel uint32, fr *proteanpb.ClientEnvelope_Done) error {
	if !fr.Done {
		// Done frame with a value of false is a no-op.
		return nil
	}

	ch, ok := ws.channels[channel]
	if !ok {
		return fmt.Errorf("unexpected input frame, channel %d is not open", channel)
	}

	if !ch.done {
		ch.done = true
		ch.call.Done()
	}

	return nil
}

// handleCancel handles a "cancel" frame received from the client on this channel.
func (ws *webSocket) handleCancel(ctx context.Context, channel uint32, fr *proteanpb.ClientEnvelope_Cancel) error {
	if !fr.Cancel {
		// Cancel frame with a value of false is a no-op.
		return nil
	}

	ch, ok := ws.channels[channel]
	if !ok {
		return fmt.Errorf("unexpected input frame, channel %d is not open", channel)
	}

	ch.done = true
	ch.cancel()

	return nil
}

// readLoop reads envelopes from the websocket connection and sends them to the
// ws.in channel.
func (ws *webSocket) readLoop(ctx context.Context) error {
	for {
		env, err := ws.readEnvelope()
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ws.in <- env:
		}
	}
}

func (ws *webSocket) sendLoop(
	ctx context.Context,
	channel uint32,
	call runtime.Call,
) error {
	for {
		out, ok := call.Recv()
		if !ok {
			break
		}

		env, err := newOutputFrame(channel, out)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ws.out <- env:
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ws.out <- &proteanpb.ServerEnvelope{
		Channel: channel,
		Frame: &proteanpb.ServerEnvelope_Done{
			Done: true,
		},
	}:
	}

	var env *proteanpb.ServerEnvelope
	if err := call.Wait(); err != nil {
		env, err = newErrorFrame(channel, err)
		if err != nil {
			return err
		}
	} else {
		env = &proteanpb.ServerEnvelope{
			Channel: channel,
			Frame: &proteanpb.ServerEnvelope_Success{
				Success: true,
			},
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ws.out <- env:
		return nil
	}
}

// readEnvelope reads the next envelope from the websocket connection.
func (ws *webSocket) readEnvelope() (*proteanpb.ClientEnvelope, error) {
	_, data, err := ws.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	env := &proteanpb.ClientEnvelope{}
	if err := ws.unmarshaler.Unmarshal(data, env); err != nil {
		return nil, err
	}

	return env, nil
}

// writeEnvelope writes an envelope to the websocket connection.
func (ws *webSocket) writeEnvelope(env *proteanpb.ServerEnvelope) error {
	messageType := websocket.TextMessage
	if ws.marshaler == protomime.BinaryMarshaler {
		messageType = websocket.BinaryMessage
	}

	data, err := ws.marshaler.Marshal(env)
	if err != nil {
		return err
	}

	if err := ws.conn.WriteMessage(messageType, data); err != nil {
		return err
	}

	if env.GetSuccess() || env.GetError() != nil {
		delete(ws.channels, env.Channel)
	}

	return nil
}

func (ws *webSocket) resolveMethod(name string) (runtime.Method, error) {
	serviceName, methodName, ok := parsePath("/" + name)
	if !ok {
		return nil, rpcerror.New(
			rpcerror.NotImplemented,
			"the method name '%s' is invalid, specify the method name using <package>/<service>/<method> syntax",
			name,
		)
	}

	service, ok := ws.services[serviceName]
	if !ok {
		return nil, rpcerror.New(
			rpcerror.NotImplemented,
			"the server does not provide the '%s' service",
			serviceName,
		)
	}

	method, ok := service.MethodByName(methodName)
	if !ok {
		return nil, rpcerror.New(
			rpcerror.NotImplemented,
			"the '%s' service does not contain an RPC method named '%s'",
			serviceName,
			methodName,
		)
	}

	return method, nil
}

func newOutputFrame(channel uint32, out proto.Message) (*proteanpb.ServerEnvelope, error) {
	any, err := anypb.New(out)
	if err != nil {
		return nil, err
	}

	return &proteanpb.ServerEnvelope{
		Channel: channel,
		Frame: &proteanpb.ServerEnvelope_Output{
			Output: any,
		},
	}, nil
}

func newErrorFrame(channel uint32, err error) (*proteanpb.ServerEnvelope, error) {
	rpcErr, ok := err.(rpcerror.Error)
	if !ok {
		rpcErr = rpcerror.New(
			rpcerror.Unknown,
			"the RPC method returned an unrecognized error",
		).WithCause(err)
	}

	var protoErr proteanpb.Error
	if err := rpcerror.ToProto(rpcErr, &protoErr); err != nil {
		return nil, err
	}

	return &proteanpb.ServerEnvelope{
		Channel: channel,
		Frame: &proteanpb.ServerEnvelope_Error{
			Error: &protoErr,
		},
	}, nil
}
