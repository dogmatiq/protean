package rpcerror_test

import (
	"errors"

	"github.com/dogmatiq/protean/internal/proteanpb"
	. "github.com/dogmatiq/protean/rpcerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ = Describe("type Error", func() {
	Describe("func New()", func() {
		It("returns an error with the given code and message", func() {
			err := New(NotFound, "<message>")
			Expect(err.Code()).To(Equal(NotFound))
			Expect(err.Message()).To(Equal("<message>"))
		})
	})

	Describe("func Details()", func() {
		It("returns the details value", func() {
			details := &proteanpb.SupportedMediaTypes{
				MediaTypes: []string{"text/plain"},
			}

			err := New(Unknown, "<message>").
				WithDetails(details)

			d, ok, detailsErr := err.Details()
			Expect(detailsErr).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(proto.Equal(d, details)).To(BeTrue(), "error details do not match")
		})

		It("returns false if there is no details value", func() {
			err := New(Unknown, "<message>")

			_, ok, detailsErr := err.Details()
			Expect(detailsErr).ShouldNot(HaveOccurred())
			Expect(ok).To(BeFalse())
		})
	})

	Describe("func WithDetails()", func() {
		It("panics if the error already has details", func() {
			details := &proteanpb.SupportedMediaTypes{}
			err := New(Unknown, "<message>").
				WithDetails(details)

			Expect(func() {
				err.WithDetails(details)
			}).To(PanicWith("error details have already been provided"))
		})
	})

	Describe("func WithCause()", func() {
		It("panics if the error already has a cause", func() {
			cause := errors.New("<cause>")
			err := New(Unknown, "<message>").
				WithCause(cause)

			Expect(func() {
				err.WithCause(cause)
			}).To(PanicWith("error cause has already been provided"))
		})
	})

	Describe("func Error()", func() {
		BeforeEach(func() {
			format.TruncatedDiff = false
		})

		AfterEach(func() {
			format.TruncatedDiff = true
		})

		It("returns a human readable description of the error", func() {
			err := New(NotFound, "<message>")
			Expect(err.Error()).To(Equal("not found: <message>"))

			err = err.WithDetails(&proteanpb.SupportedMediaTypes{})
			Expect(err.Error()).To(Equal("not found [protean.v1.SupportedMediaTypes]: <message>"))
		})

		It("adds a message if none is provided", func() {
			err := New(Unknown, "")
			Expect(err.Error()).To(Equal("unknown: <no message provided>"))
		})
	})

	Describe("func ToProto()", func() {
		It("constructs an error from its protocol buffers representation", func() {
			details := &proteanpb.SupportedMediaTypes{}

			var protoErr proteanpb.Error
			err := ToProto(
				New(NotFound, "<message>").
					WithDetails(details),
				&protoErr,
			)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(protoErr.GetCode()).To(Equal(NotFound.NumericValue()))
			Expect(protoErr.GetMessage()).To(Equal("<message>"))

			d, err := protoErr.GetData().UnmarshalNew()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(proto.Equal(d, details)).To(BeTrue(), "error details do not match")
		})

		It("returns an error if the protocol buffers message type is not supported", func() {
			err := ToProto(
				New(NotFound, "<message>"),
				&proteanpb.SupportedMediaTypes{},
			)
			Expect(err).To(MatchError("unsupported protocol buffers message type"))
		})
	})

	Describe("func FromProto()", func() {
		It("constructs an error from its protocol buffers representation", func() {
			details := &proteanpb.SupportedMediaTypes{}

			protoErr := &proteanpb.Error{
				Code:    NotFound.NumericValue(),
				Message: "<message>",
				Data:    &anypb.Any{},
			}

			err := protoErr.Data.MarshalFrom(details)
			Expect(err).ShouldNot(HaveOccurred())

			rpcErr, err := FromProto(protoErr)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(rpcErr.Code()).To(Equal(NotFound))
			Expect(rpcErr.Message()).To(Equal("<message>"))

			d, ok, err := rpcErr.Details()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(proto.Equal(d, details)).To(BeTrue(), "error details do not match")
		})

		It("returns an error if the protocol buffers message type is not supported", func() {
			_, err := FromProto(&proteanpb.SupportedMediaTypes{})
			Expect(err).To(MatchError("unsupported protocol buffers message type"))
		})
	})
})
