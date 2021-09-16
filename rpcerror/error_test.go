package rpcerror_test

import (
	"errors"

	"github.com/dogmatiq/protean/internal/proteanpb"
	. "github.com/dogmatiq/protean/rpcerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("type Error", func() {
	Describe("func New()", func() {
		It("returns an error with the given code and message", func() {
			err := New(NotFound, "<message>")
			Expect(err.Code()).To(Equal(NotFound))
			Expect(err.Message()).To(Equal("<message>"))
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
	})
})
