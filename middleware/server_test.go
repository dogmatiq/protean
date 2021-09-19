package middleware_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/protean/internal/testservice"
	. "github.com/dogmatiq/protean/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("type ServerChain", func() {
	Describe("func InterceptUnaryRPC()", func() {
		It("calls each interceptor in the chain", func() {
			info := UnaryServerInfo{
				Package: "<package>",
				Service: "<service>",
				Method:  "<method>",
			}

			chain := ServerChain{
				&serverStub{
					InterceptUnaryRPCFunc: func(
						ctx context.Context,
						i UnaryServerInfo,
						in proto.Message,
						next func(ctx context.Context) (out proto.Message, err error),
					) (proto.Message, error) {
						Expect(i).To(Equal(info))

						{
							m, ok := in.(*testservice.Input)
							Expect(ok).To(BeTrue())
							Expect(m.GetData()).To(Equal("<input one>"))
							m.Data = "<input two>"
						}

						out, err := next(ctx)
						Expect(err).To(MatchError("<error two>"))

						{
							m, ok := out.(*testservice.Output)
							Expect(ok).To(BeTrue())
							Expect(m.GetData()).To(Equal("<output two>"))
						}

						return &testservice.Output{
							Data: "<output three>",
						}, errors.New("<error three>")
					},
				},
				&serverStub{
					InterceptUnaryRPCFunc: func(
						ctx context.Context,
						i UnaryServerInfo,
						in proto.Message,
						next func(ctx context.Context) (out proto.Message, err error),
					) (proto.Message, error) {
						Expect(i).To(Equal(info))

						{
							m, ok := in.(*testservice.Input)
							Expect(ok).To(BeTrue())
							Expect(m.GetData()).To(Equal("<input two>"))
							m.Data = "<input three>"
						}

						out, err := next(ctx)
						Expect(err).To(MatchError("<error one>"))

						{
							m, ok := out.(*testservice.Output)
							Expect(ok).To(BeTrue())
							Expect(m.GetData()).To(Equal("<output one>"))
						}

						return &testservice.Output{
							Data: "<output two>",
						}, errors.New("<error two>")
					},
				},
			}

			out, err := chain.InterceptUnaryRPC(
				context.Background(),
				info,
				&testservice.Input{
					Data: "<input one>",
				},
				func(ctx context.Context) (out proto.Message, err error) {
					return &testservice.Output{
						Data: "<output one>",
					}, errors.New("<error one>")
				},
			)

			Expect(err).To(MatchError("<error three>"))

			m, ok := out.(*testservice.Output)
			Expect(ok).To(BeTrue())
			Expect(m.GetData()).To(Equal("<output three>"))
		})
	})
})

type serverStub struct {
	InterceptUnaryRPCFunc func(
		ctx context.Context,
		info UnaryServerInfo,
		in proto.Message,
		next func(ctx context.Context) (out proto.Message, err error),
	) (proto.Message, error)
}

func (s *serverStub) InterceptUnaryRPC(
	ctx context.Context,
	info UnaryServerInfo,
	in proto.Message,
	next func(ctx context.Context) (out proto.Message, err error),
) (proto.Message, error) {
	if s.InterceptUnaryRPCFunc != nil {
		return s.InterceptUnaryRPCFunc(ctx, info, in, next)
	}

	return next(ctx)
}
