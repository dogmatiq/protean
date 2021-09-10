// Code generated by protoc-gen-go-protean. DO NOT EDIT.
// versions:
// 	protoc-gen-go-protean v0.0.0+68cf482
// 	protoc                v3.17.3
// source: github.com/dogmatiq/protean/internal/testservice/service.proto

package testservice

import (
	"context"
	"errors"
	runtime "github.com/dogmatiq/protean/runtime"
	proto "google.golang.org/protobuf/proto"
)

// ProteanTestService is an interface for the protean.test.TestService service.
//
// It is used for both clients and servers.
type ProteanTestService interface {
	Unary(context.Context, *Input) (*Output, error)
	ClientStream(context.Context, <-chan *Input) (*Output, error)
	ServerStream(context.Context, *Input, chan<- *Output) error
	BidirectionalStream(context.Context, <-chan *Input, chan<- *Output) error
}

// protean_TestService_Service is an implementation of the runtime.Service
// interface for the TestService service.
type protean_TestService_Service struct {
	service ProteanTestService
}

// ProteanRegisterTestServiceServer registers a ProteanTestService service with a Protean server.
func ProteanRegisterTestServiceServer(r runtime.Registry, s ProteanTestService) {
	r.RegisterService(&protean_TestService_Service{s})
}

func (a *protean_TestService_Service) Name() string {
	return "TestService"
}

func (a *protean_TestService_Service) Package() string {
	return "protean.test"
}

func (a *protean_TestService_Service) MethodByName(name string) (runtime.Method, bool) {
	switch name {
	case "Unary":
		return &protean_TestService_Unary_Method{a.service}, true
	case "ClientStream":
		return &protean_TestService_ClientStream_Method{a.service}, true
	case "ServerStream":
		return &protean_TestService_ServerStream_Method{a.service}, true
	case "BidirectionalStream":
		return &protean_TestService_BidirectionalStream_Method{a.service}, true
	default:
		return nil, false
	}
}

// protean_TestService_Unary_Method is an implementation of the runtime.Method
// interface for the TestService.Unary() method.
type protean_TestService_Unary_Method struct {
	service ProteanTestService
}

func (m *protean_TestService_Unary_Method) Name() string {
	return "Unary"
}

func (m *protean_TestService_Unary_Method) InputIsStream() bool {
	return false
}

func (m *protean_TestService_Unary_Method) OutputIsStream() bool {
	return false
}

func (m *protean_TestService_Unary_Method) NewCall(ctx context.Context) runtime.Call {
	return newprotean_TestService_Unary_Call(ctx, m.service)
}

// protean_TestService_Unary_Call is an implementation of the runtime.Call
// interface for the TestService.Unary() method.
type protean_TestService_Unary_Call struct {
	ctx     context.Context
	service ProteanTestService
	done    chan struct{}
	res     *Output
	err     error
}

// newprotean_TestService_Unary_Call returns a new runtime.Call for the ProteanTestService.Unary() method.
func newprotean_TestService_Unary_Call(ctx context.Context, service ProteanTestService) runtime.Call {
	return &protean_TestService_Unary_Call{ctx, service, make(chan struct{}), nil, nil}
}

func (c *protean_TestService_Unary_Call) Send(unmarshal runtime.Unmarshaler) error {
	req := &Input{}
	if err := unmarshal(req); err != nil {
		return err
	}

	c.res, c.err = c.service.Unary(c.ctx, req)
	close(c.done)

	return nil
}

func (c *protean_TestService_Unary_Call) Done() {}

func (c *protean_TestService_Unary_Call) Recv() (proto.Message, bool, error) {
	select {
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	case <-c.done:
		return c.res, c.err == nil, c.err
	}
}

// protean_TestService_ClientStream_Method is an implementation of the runtime.Method
// interface for the TestService.ClientStream() method.
type protean_TestService_ClientStream_Method struct {
	service ProteanTestService
}

func (m *protean_TestService_ClientStream_Method) Name() string {
	return "ClientStream"
}

func (m *protean_TestService_ClientStream_Method) InputIsStream() bool {
	return true
}

func (m *protean_TestService_ClientStream_Method) OutputIsStream() bool {
	return false
}

func (m *protean_TestService_ClientStream_Method) NewCall(ctx context.Context) runtime.Call {
	return newprotean_TestService_ClientStream_Call(ctx, m.service)
}

// protean_TestService_ClientStream_Call is an implementation of the runtime.Call
// interface for the TestService.ClientStream() method.
type protean_TestService_ClientStream_Call struct {
	ctx     context.Context
	service ProteanTestService
	in      chan *Input
	out     chan *Output
	err     error
}

// newprotean_TestService_ClientStream_Call returns a new runtime.Call for the ProteanTestService.ClientStream() method.
func newprotean_TestService_ClientStream_Call(ctx context.Context, service ProteanTestService) runtime.Call {
	c := &protean_TestService_ClientStream_Call{ctx, service, make(chan *Input, 1), make(chan *Output, 1), nil}
	go c.run()
	return c
}

func (c *protean_TestService_ClientStream_Call) Send(unmarshal runtime.Unmarshaler) error {
	req := &Input{}
	if err := unmarshal(req); err != nil {
		return err
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.in <- req:
		return nil
	}
}

func (c *protean_TestService_ClientStream_Call) Done() {
	close(c.in)
}

func (c *protean_TestService_ClientStream_Call) Recv() (proto.Message, bool, error) {
	select {
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	case res, ok := <-c.out:
		if ok {
			return res, true, nil
		}
		return nil, false, c.err
	}
}

func (c *protean_TestService_ClientStream_Call) run() {
	defer close(c.out)

	res, err := c.service.ClientStream(c.ctx, c.in)
	if err != nil {
		c.err = err
	} else {
		c.out <- res // buffered, never blocks
	}
}

// protean_TestService_ServerStream_Method is an implementation of the runtime.Method
// interface for the TestService.ServerStream() method.
type protean_TestService_ServerStream_Method struct {
	service ProteanTestService
}

func (m *protean_TestService_ServerStream_Method) Name() string {
	return "ServerStream"
}

func (m *protean_TestService_ServerStream_Method) InputIsStream() bool {
	return false
}

func (m *protean_TestService_ServerStream_Method) OutputIsStream() bool {
	return true
}

func (m *protean_TestService_ServerStream_Method) NewCall(ctx context.Context) runtime.Call {
	return newprotean_TestService_ServerStream_Call(ctx, m.service)
}

// protean_TestService_ServerStream_Call is an implementation of the runtime.Call
// interface for the TestService.ServerStream() method.
type protean_TestService_ServerStream_Call struct {
	ctx     context.Context
	service ProteanTestService
	in      chan *Input
	out     chan *Output
	err     error
}

// newprotean_TestService_ServerStream_Call returns a new runtime.Call for the ProteanTestService.ServerStream() method.
func newprotean_TestService_ServerStream_Call(ctx context.Context, service ProteanTestService) runtime.Call {
	c := &protean_TestService_ServerStream_Call{ctx, service, make(chan *Input, 1), make(chan *Output, 1), nil}
	go c.run()
	return c
}

func (c *protean_TestService_ServerStream_Call) Send(unmarshal runtime.Unmarshaler) error {
	req := &Input{}
	if err := unmarshal(req); err != nil {
		return err
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.in <- req:
		return nil
	}
}

func (c *protean_TestService_ServerStream_Call) Done() {
	close(c.in)
}

func (c *protean_TestService_ServerStream_Call) Recv() (proto.Message, bool, error) {
	select {
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	case res, ok := <-c.out:
		if ok {
			return res, true, nil
		}
		return nil, false, c.err
	}
}

func (c *protean_TestService_ServerStream_Call) run() {
	defer close(c.out)

	select {
	case <-c.ctx.Done():
		c.err = c.ctx.Err()
	case req, ok := <-c.in:
		if ok {
			c.err = c.service.ServerStream(c.ctx, req, c.out)
		} else {
			c.err = errors.New("Done() was called before Send()")
		}
	}
}

// protean_TestService_BidirectionalStream_Method is an implementation of the runtime.Method
// interface for the TestService.BidirectionalStream() method.
type protean_TestService_BidirectionalStream_Method struct {
	service ProteanTestService
}

func (m *protean_TestService_BidirectionalStream_Method) Name() string {
	return "BidirectionalStream"
}

func (m *protean_TestService_BidirectionalStream_Method) InputIsStream() bool {
	return true
}

func (m *protean_TestService_BidirectionalStream_Method) OutputIsStream() bool {
	return true
}

func (m *protean_TestService_BidirectionalStream_Method) NewCall(ctx context.Context) runtime.Call {
	return newprotean_TestService_BidirectionalStream_Call(ctx, m.service)
}

// protean_TestService_BidirectionalStream_Call is an implementation of the runtime.Call
// interface for the TestService.BidirectionalStream() method.
type protean_TestService_BidirectionalStream_Call struct {
	ctx     context.Context
	service ProteanTestService
	in      chan *Input
	out     chan *Output
	err     error
}

// newprotean_TestService_BidirectionalStream_Call returns a new runtime.Call for the ProteanTestService.BidirectionalStream() method.
func newprotean_TestService_BidirectionalStream_Call(ctx context.Context, service ProteanTestService) runtime.Call {
	c := &protean_TestService_BidirectionalStream_Call{ctx, service, make(chan *Input, 1), make(chan *Output, 1), nil}
	go c.run()
	return c
}

func (c *protean_TestService_BidirectionalStream_Call) Send(unmarshal runtime.Unmarshaler) error {
	req := &Input{}
	if err := unmarshal(req); err != nil {
		return err
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.in <- req:
		return nil
	}
}

func (c *protean_TestService_BidirectionalStream_Call) Done() {
	close(c.in)
}

func (c *protean_TestService_BidirectionalStream_Call) Recv() (proto.Message, bool, error) {
	select {
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	case res, ok := <-c.out:
		if ok {
			return res, true, nil
		}
		return nil, false, c.err
	}
}

func (c *protean_TestService_BidirectionalStream_Call) run() {
	defer close(c.out)
	c.err = c.service.BidirectionalStream(c.ctx, c.in, c.out)
}
