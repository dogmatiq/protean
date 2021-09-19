package middleware_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/protean/internal/stringservice"
	"github.com/dogmatiq/protean/internal/testservice"
	. "github.com/dogmatiq/protean/middleware"
	"github.com/dogmatiq/protean/rpcerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("type Validator", func() {
	var validator Validator

	Describe("func InterceptUnaryRPC()", func() {
		When("the messages implement ValidatableMessage", func() {
			It("behaves normally if both messages are valid", func() {
				expect := &testservice.Output{
					Data: "<data>",
				}

				out, err := validator.InterceptUnaryRPC(
					context.Background(),
					UnaryServerInfo{},
					&testservice.Input{
						Data: "<data>",
					},
					func(ctx context.Context) (proto.Message, error) {
						return expect, nil
					},
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(out).To(BeIdenticalTo(expect))
			})

			It("does not call next() if the input message is invalid", func() {
				_, err := validator.InterceptUnaryRPC(
					context.Background(),
					UnaryServerInfo{},
					&testservice.Input{
						Data: "", // invalid
					},
					func(ctx context.Context) (proto.Message, error) {
						Fail("unexpected call")
						return nil, nil
					},
				)

				Expect(err).To(Equal(
					rpcerror.New(
						rpcerror.InvalidInput,
						"the RPC input message is invalid: input data must not be empty",
					),
				))
			})

			It("returns an error if the output message is invalid", func() {
				_, err := validator.InterceptUnaryRPC(
					context.Background(),
					UnaryServerInfo{},
					&testservice.Input{
						Data: "<data>",
					},
					func(ctx context.Context) (proto.Message, error) {
						return &testservice.Output{
							Data: "", // invalid
						}, nil
					},
				)

				Expect(err).To(Equal(
					rpcerror.New(
						rpcerror.Unknown,
						"the server produced an invalid RPC output message",
					).WithCause(
						errors.New("output data must not be empty"),
					),
				))
			})
		})

		When("the messages do not implement ValidatableMessage", func() {
			It("calls next() and returns its result", func() {
				expect := &stringservice.ToUpperResponse{}

				out, err := validator.InterceptUnaryRPC(
					context.Background(),
					UnaryServerInfo{},
					&stringservice.ToUpperRequest{},
					func(ctx context.Context) (proto.Message, error) {
						return expect, nil
					},
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(out).To(BeIdenticalTo(expect))
			})
		})

		It("it returns the error returned by next()", func() {
			_, err := validator.InterceptUnaryRPC(
				context.Background(),
				UnaryServerInfo{},
				&stringservice.ToUpperRequest{},
				func(ctx context.Context) (proto.Message, error) {
					return nil, errors.New("<error>")
				},
			)

			Expect(err).To(MatchError("<error>"))
		})
	})
})
