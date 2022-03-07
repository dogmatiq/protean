package protean

import (
	"context"
	"time"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/dogmatiq/protean/runtime"
	"github.com/gorilla/websocket"
)

// webSocketCall manages communication for a single RPC call made over a
// websocket connection.
type webSocketCall struct {
	Conn            *websocket.Conn
	Services        map[string]runtime.Service
	Marshaler       protomime.Marshaler
	ProtocolTimeout time.Duration
	ID              uint32
	MethodName      string
}

// Serve a single this RPC request until ctx is canceled or an error occurs.
func (c *webSocketCall) Serve(ctx context.Context) error {
	method, err := c.lookupMethod(ctx)
	if err != nil {
		return c.sendError(err)
	}

	if method.InputIsStream() {
		return nil
	}

	time.Sleep(c.ProtocolTimeout)

	// TODO: use an RPC error instead of closing the connection
	return newWebSocketError(
		websocket.CloseProtocolError,
		"expected 'send' frame within %s of 'call' frame (%d)",
		c.ProtocolTimeout,
		c.ID,
	)
}

// lookupMethod returns the method referenced by the "call" frame that started
// this call.
func (c *webSocketCall) lookupMethod(ctx context.Context) (runtime.Method, error) {
	// Using parsePath() guarantees that the method name parsing behaves
	// identically to the HTTP-request-based transports.
	serviceName, methodName, ok := parsePath("/" + c.MethodName)
	if !ok {
		return nil, rpcerror.New(
			rpcerror.NotImplemented,
			"method name must be in '<package>/<service>/<method>' format",
		)
	}

	service, ok := c.Services[serviceName]
	if !ok {
		return nil, unimplementedServiceError(serviceName)
	}

	method, ok := service.MethodByName(methodName)
	if !ok {
		return nil, unimplementedMethodError(serviceName, methodName)
	}

	return method, nil
}

// sendError sends an "error" frame in response to this call.
func (c *webSocketCall) sendError(err error) error {
	rpcErr := err.(rpcerror.Error)

	protoErr := &proteanpb.Error{}
	if err := rpcerror.ToProto(rpcErr, protoErr); err != nil {
		panic(err)
	}

	env := &proteanpb.ServerEnvelope{
		CallId: c.ID,
		Frame: &proteanpb.ServerEnvelope_Error{
			Error: protoErr,
		},
	}

	data, err := c.Marshaler.Marshal(env)
	if err != nil {
		panic(err)
	}

	// TODO: hack
	messageType := websocket.TextMessage
	if c.Marshaler == protomime.BinaryMarshaler {
		messageType = websocket.BinaryMessage
	}

	return c.Conn.WriteMessage(messageType, data)
}
