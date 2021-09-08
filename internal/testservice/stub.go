package testservice

import "context"

// Stub is a test implementation of the API interface.
type Stub struct {
	UnaryFunc               func(context.Context, *Request) (*Response, error)
	ServerStreamFunc        func(context.Context, *Request, chan<- *Response) error
	ClientStreamFunc        func(context.Context, <-chan *Request) (*Response, error)
	BidirectionalStreamFunc func(context.Context, <-chan *Request, chan<- *Response) error
}

// Unary calls s.UnaryFunc(ctx, in) if s.UnaryFunc is not nil. Otherwise, it
// returns a zero-value response.
func (s *Stub) Unary(ctx context.Context, in *Request) (*Response, error) {
	if s.UnaryFunc != nil {
		return s.UnaryFunc(ctx, in)
	}

	return &Response{}, nil
}

// ServerStream calls s.ServerStreamFunc(ctx, in, out) if s.ServerStreamFunc is
// not nil. Otherwise, it returns nil without producing any responses.
func (s *Stub) ServerStream(ctx context.Context, in *Request, out chan<- *Response) error {
	if s.ServerStreamFunc != nil {
		return s.ServerStreamFunc(ctx, in, out)
	}

	return nil
}

// ClientStream calls s.ClientStreamFunc(ctx, in) if s.ClientStreamFunc is not
// nil. Otherwise, it reads all the requests and returns an empty result.
func (s *Stub) ClientStream(ctx context.Context, in <-chan *Request) (*Response, error) {
	if s.ClientStreamFunc != nil {
		return s.ClientStreamFunc(ctx, in)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case _, ok := <-in:
			if !ok {
				return &Response{}, nil
			}
		}
	}
}

// BidirectionalStream calls s.BidirectionalStreamFunc(ctx, in, out) if
// s.BidirectionalStreamFunc is not nil. Otherwise, it reads all the requests
// without producing any responses.
func (s *Stub) BidirectionalStream(ctx context.Context, in <-chan *Request, out chan<- *Response) error {
	if s.BidirectionalStreamFunc != nil {
		return s.BidirectionalStreamFunc(ctx, in, out)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-in:
			if !ok {
				return nil
			}
		}
	}
}
