// Code generated by protoc-gen-go-protean. DO NOT EDIT.
// versions:
// 	protoc-gen-go-protean v0.0.0+3ef1194
// 	protoc                v3.17.3
// source: github.com/dogmatiq/protean/internal/testservice/service.proto

package testservice

import (
	"context"
	runtime "github.com/dogmatiq/protean/runtime"
	proto "google.golang.org/protobuf/proto"
)

// ProteanTestService is an interface for the protean.test.TestService service.
type ProteanTestService interface {
	Unary(context.Context, *Input) (*Output, error)
	ClientStream(context.Context, <-chan *Input) (*Output, error)
	ServerStream(context.Context, *Input, chan<- *Output) error
	BidirectionalStream(context.Context, <-chan *Input, chan<- *Output) error
}

// ProteanRegisterTestServiceServer registers a ProteanTestService service with a Protean registry.
func ProteanRegisterTestServiceServer(r runtime.Registry, s ProteanTestService) {
	r.RegisterService(&proteanService_TestService{s})
}

// proteanService_TestService is a runtime.Service implementation for the protean.test.TestService service.
type proteanService_TestService struct {
	service ProteanTestService
}

func (a *proteanService_TestService) Name() string {
	return "TestService"
}

func (a *proteanService_TestService) Package() string {
	return "protean.test"
}

func (a *proteanService_TestService) MethodByName(name string) (runtime.Method, bool) {
	switch name {
	case "Unary":
		return &proteanMethod_TestService_Unary{a.service}, true
	case "ClientStream":
		return &proteanMethod_TestService_ClientStream{a.service}, true
	case "ServerStream":
		return &proteanMethod_TestService_ServerStream{a.service}, true
	case "BidirectionalStream":
		return &proteanMethod_TestService_BidirectionalStream{a.service}, true
	default:
		return nil, false
	}
}

// proteanMethod_TestService_Unary is a runtime.Method implementation for the protean.test.TestService.Unary() method.
type proteanMethod_TestService_Unary struct {
	service ProteanTestService
}

func (m *proteanMethod_TestService_Unary) Name() string {
	return "Unary"
}

func (m *proteanMethod_TestService_Unary) InputIsStream() bool {
	return false
}

func (m *proteanMethod_TestService_Unary) OutputIsStream() bool {
	return false
}

func (m *proteanMethod_TestService_Unary) NewCall(ctx context.Context) runtime.Call {
	return newProteanCall_TestService_Unary(ctx, m.service)
}

// newProteanCall_TestService_Unary returns a new runtime.Call for the protean.test.TestService.Unary() method.
func newProteanCall_TestService_Unary(ctx context.Context, service ProteanTestService) runtime.Call {
	return &proteanCall_TestService_Unary{ctx, service, make(chan *Input, 1)}
}

// proteanMethod_TestService_Unary is a runtime.Call implementation for the protean.test.TestService.Unary() method.
type proteanCall_TestService_Unary struct {
	ctx     context.Context
	service ProteanTestService
	in      chan *Input
}

func (c *proteanCall_TestService_Unary) Send(unmarshal runtime.Unmarshaler) (bool, error) {
	in := &Input{}
	if err := unmarshal(in); err != nil {
		return false, err
	}

	c.in <- in
	close(c.in)

	return false, nil
}

func (c *proteanCall_TestService_Unary) Done() {}

func (c *proteanCall_TestService_Unary) Recv() (proto.Message, bool, error) {
	select {
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	case in, ok := <-c.in:
		if !ok {
			return nil, false, nil
		}

		out, err := c.service.Unary(c.ctx, in)
		return out, false, err
	}
}

// proteanMethod_TestService_ClientStream is a runtime.Method implementation for the protean.test.TestService.ClientStream() method.
type proteanMethod_TestService_ClientStream struct {
	service ProteanTestService
}

func (m *proteanMethod_TestService_ClientStream) Name() string {
	return "ClientStream"
}

func (m *proteanMethod_TestService_ClientStream) InputIsStream() bool {
	return true
}

func (m *proteanMethod_TestService_ClientStream) OutputIsStream() bool {
	return false
}

func (m *proteanMethod_TestService_ClientStream) NewCall(ctx context.Context) runtime.Call {
	return newProteanCall_TestService_ClientStream(ctx, m.service)
}

// newProteanCall_TestService_ClientStream returns a new runtime.Call for the protean.test.TestService.ClientStream() method.
func newProteanCall_TestService_ClientStream(ctx context.Context, service ProteanTestService) runtime.Call {
	c := &proteanCall_TestService_ClientStream{ctx, service, make(chan *Input)}
	return c
}

// proteanMethod_TestService_ClientStream is a runtime.Call implementation for the protean.test.TestService.ClientStream() method.
type proteanCall_TestService_ClientStream struct {
	ctx     context.Context
	service ProteanTestService
	in      chan *Input
}

func (c *proteanCall_TestService_ClientStream) Send(unmarshal runtime.Unmarshaler) (bool, error) {
	in := &Input{}
	if err := unmarshal(in); err != nil {
		return false, err
	}

	select {
	case <-c.ctx.Done():
		return false, nil
	case c.in <- in:
		return true, nil
	}
}

func (c *proteanCall_TestService_ClientStream) Done() {
	close(c.in)
}

func (c *proteanCall_TestService_ClientStream) Recv() (proto.Message, bool, error) {
	out, err := c.service.ClientStream(c.ctx, c.in)
	return out, false, err
}

// proteanMethod_TestService_ServerStream is a runtime.Method implementation for the protean.test.TestService.ServerStream() method.
type proteanMethod_TestService_ServerStream struct {
	service ProteanTestService
}

func (m *proteanMethod_TestService_ServerStream) Name() string {
	return "ServerStream"
}

func (m *proteanMethod_TestService_ServerStream) InputIsStream() bool {
	return false
}

func (m *proteanMethod_TestService_ServerStream) OutputIsStream() bool {
	return true
}

func (m *proteanMethod_TestService_ServerStream) NewCall(ctx context.Context) runtime.Call {
	return newProteanCall_TestService_ServerStream(ctx, m.service)
}

// newProteanCall_TestService_ServerStream returns a new runtime.Call for the protean.test.TestService.ServerStream() method.
func newProteanCall_TestService_ServerStream(ctx context.Context, service ProteanTestService) runtime.Call {
	c := &proteanCall_TestService_ServerStream{ctx, service, make(chan *Input, 1), make(chan *Output, 1), nil}
	go c.run()
	return c
}

// proteanMethod_TestService_ServerStream is a runtime.Call implementation for the protean.test.TestService.ServerStream() method.
type proteanCall_TestService_ServerStream struct {
	ctx     context.Context
	service ProteanTestService
	in      chan *Input
	out     chan *Output
	err     error
}

func (c *proteanCall_TestService_ServerStream) Send(unmarshal runtime.Unmarshaler) (bool, error) {
	in := &Input{}
	if err := unmarshal(in); err != nil {
		return false, err
	}

	c.in <- in
	close(c.in)

	return false, nil
}

func (c *proteanCall_TestService_ServerStream) Done() {}

func (c *proteanCall_TestService_ServerStream) Recv() (proto.Message, bool, error) {
	if out, ok := <-c.out; ok {
		return out, true, nil
	}
	return nil, false, c.err
}

func (c *proteanCall_TestService_ServerStream) run() {
	defer close(c.out)

	select {
	case <-c.ctx.Done():
		c.err = c.ctx.Err()
	case in := <-c.in:
		c.err = c.service.ServerStream(c.ctx, in, c.out)
	}
}

// proteanMethod_TestService_BidirectionalStream is a runtime.Method implementation for the protean.test.TestService.BidirectionalStream() method.
type proteanMethod_TestService_BidirectionalStream struct {
	service ProteanTestService
}

func (m *proteanMethod_TestService_BidirectionalStream) Name() string {
	return "BidirectionalStream"
}

func (m *proteanMethod_TestService_BidirectionalStream) InputIsStream() bool {
	return true
}

func (m *proteanMethod_TestService_BidirectionalStream) OutputIsStream() bool {
	return true
}

func (m *proteanMethod_TestService_BidirectionalStream) NewCall(ctx context.Context) runtime.Call {
	return newProteanCall_TestService_BidirectionalStream(ctx, m.service)
}

// newProteanCall_TestService_BidirectionalStream returns a new runtime.Call for the protean.test.TestService.BidirectionalStream() method.
func newProteanCall_TestService_BidirectionalStream(ctx context.Context, service ProteanTestService) runtime.Call {
	c := &proteanCall_TestService_BidirectionalStream{ctx, service, make(chan *Input), make(chan *Output), nil}
	go c.run()
	return c
}

// proteanMethod_TestService_BidirectionalStream is a runtime.Call implementation for the protean.test.TestService.BidirectionalStream() method.
type proteanCall_TestService_BidirectionalStream struct {
	ctx     context.Context
	service ProteanTestService
	in      chan *Input
	out     chan *Output
	err     error
}

func (c *proteanCall_TestService_BidirectionalStream) Send(unmarshal runtime.Unmarshaler) (bool, error) {
	in := &Input{}
	if err := unmarshal(in); err != nil {
		return false, err
	}

	select {
	case <-c.ctx.Done():
		return false, nil
	case c.in <- in:
		return true, nil
	}
}

func (c *proteanCall_TestService_BidirectionalStream) Done() {
	close(c.in)
}

func (c *proteanCall_TestService_BidirectionalStream) Recv() (proto.Message, bool, error) {
	if out, ok := <-c.out; ok {
		return out, true, nil
	}
	return nil, false, c.err
}

func (c *proteanCall_TestService_BidirectionalStream) run() {
	c.err = c.service.BidirectionalStream(c.ctx, c.in, c.out)
	close(c.out)
}
