package testservice

import "context"

// Stub is a test implementation of the API interface.
type Stub struct {
	UnaryFunc               func(context.Context, *Input) (*Output, error)
	ServerStreamFunc        func(context.Context, *Input, chan<- *Output) error
	ClientStreamFunc        func(context.Context, <-chan *Input) (*Output, error)
	BidirectionalStreamFunc func(context.Context, <-chan *Input, chan<- *Output) error
}

// Unary calls s.UnaryFunc(ctx, in) if s.UnaryFunc is not nil. Otherwise, it
// returns a zero-value output message.
func (s *Stub) Unary(ctx context.Context, in *Input) (*Output, error) {
	if s.UnaryFunc != nil {
		return s.UnaryFunc(ctx, in)
	}

	return &Output{}, nil
}

// ServerStream calls s.ServerStreamFunc(ctx, in, out) if s.ServerStreamFunc is
// not nil. Otherwise, it returns nil without producing any output messages.
func (s *Stub) ServerStream(ctx context.Context, in *Input, out chan<- *Output) error {
	if s.ServerStreamFunc != nil {
		return s.ServerStreamFunc(ctx, in, out)
	}

	close(out)
	return nil
}

// ClientStream calls s.ClientStreamFunc(ctx, in) if s.ClientStreamFunc is not
// nil. Otherwise, it reads all the input messages and returns an empty output
// message.
func (s *Stub) ClientStream(ctx context.Context, in <-chan *Input) (*Output, error) {
	if s.ClientStreamFunc != nil {
		return s.ClientStreamFunc(ctx, in)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case _, ok := <-in:
			if !ok {
				return &Output{}, nil
			}
		}
	}
}

// BidirectionalStream calls s.BidirectionalStreamFunc(ctx, in, out) if
// s.BidirectionalStreamFunc is not nil. Otherwise, it reads all of the input
// messages without producing any output messages.
func (s *Stub) BidirectionalStream(ctx context.Context, in <-chan *Input, out chan<- *Output) error {
	if s.BidirectionalStreamFunc != nil {
		return s.BidirectionalStreamFunc(ctx, in, out)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-in:
			if !ok {
				close(out)
				return nil
			}
		}
	}
}
