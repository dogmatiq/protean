package protean_test

import (
	. "github.com/dogmatiq/protean"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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
