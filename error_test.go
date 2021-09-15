package protean_test

import (
	"errors"

	. "github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/proteanpb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("type ErrorCode", func() {
	Describe("func CustomErrorCode()", func() {
		It("panics if the code is zero", func() {
			Expect(func() {
				CustomErrorCode(0)
			}).To(PanicWith("error code must be positive"))
		})

		It("panics if the code is negative", func() {
			Expect(func() {
				CustomErrorCode(-1)
			}).To(PanicWith("error code must be positive"))
		})
	})

	Describe("func String()", func() {
		DescribeTable(
			"it returns a description of the pre-defined error codes",
			func(code ErrorCode, expect string) {
				Expect(code.String()).To(Equal(expect))
			},
			Entry("ErrorCodeUnknown", ErrorCodeUnknown, "unknown"),
			Entry("ErrorCodeInvalidInput", ErrorCodeInvalidInput, "invalid input"),
			Entry("ErrorCodeUnauthenticated", ErrorCodeUnauthenticated, "unauthenticated"),
			Entry("ErrorCodePermissionDenied", ErrorCodePermissionDenied, "permission denied"),
			Entry("ErrorCodeNotFound", ErrorCodeNotFound, "not found"),
			Entry("ErrorCodeAlreadyExists", ErrorCodeAlreadyExists, "already exists"),
			Entry("ErrorCodeResourceExhausted", ErrorCodeResourceExhausted, "resource exhausted"),
			Entry("ErrorCodeFailedPrecondition", ErrorCodeFailedPrecondition, "failed precondition"),
			Entry("ErrorCodeAborted", ErrorCodeAborted, "aborted"),
			Entry("ErrorCodeUnavailable", ErrorCodeUnavailable, "unavailable"),
			Entry("ErrorCodeNotImplemented", ErrorCodeNotImplemented, "not implemented"),
		)

		It("returns the numeric value of custom error codes", func() {
			code := CustomErrorCode(123)
			Expect(code.String()).To(Equal("123"))
		})
	})
})

var _ = Describe("type Error", func() {
	Describe("func NewError()", func() {
		It("returns an error with the given code and message", func() {
			err := NewError(ErrorCodeNotFound, "<message>")
			Expect(err.Code()).To(Equal(ErrorCodeNotFound))
			Expect(err.Message()).To(Equal("<message>"))
		})
	})

	Describe("func WithDetails()", func() {
		It("panics if the error already has details", func() {
			details := &proteanpb.SupportedMediaTypes{}
			err := NewError(ErrorCodeUnknown, "<message>").
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

			err := NewError(ErrorCodeUnknown, "<message>").
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
			err := NewError(ErrorCodeUnknown, "<message>").
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
			err := NewError(ErrorCodeNotFound, "<message>")
			Expect(err.Error()).To(Equal("not found: <message>"))

			err = err.WithDetails(&proteanpb.SupportedMediaTypes{})
			Expect(err.Error()).To(Equal("not found [protean.v1.SupportedMediaTypes]: <message>"))
		})
	})
})
