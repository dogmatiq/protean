package rpcerror_test

import (
	. "github.com/dogmatiq/protean/rpcerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Code", func() {
	Describe("func NewCode()", func() {
		It("panics if the code is zero", func() {
			Expect(func() {
				NewCode(0)
			}).To(PanicWith("error code must be positive"))
		})

		It("panics if the code is negative", func() {
			Expect(func() {
				NewCode(-1)
			}).To(PanicWith("error code must be positive"))
		})
	})

	Describe("func String()", func() {
		DescribeTable(
			"it returns a description of the pre-defined error codes",
			func(code Code, expect string) {
				Expect(code.String()).To(Equal(expect))
			},
			Entry("Unknown", Unknown, "unknown"),
			Entry("InvalidInput", InvalidInput, "invalid input"),
			Entry("Unauthenticated", Unauthenticated, "unauthenticated"),
			Entry("PermissionDenied", PermissionDenied, "permission denied"),
			Entry("NotFound", NotFound, "not found"),
			Entry("AlreadyExists", AlreadyExists, "already exists"),
			Entry("ResourceExhausted", ResourceExhausted, "resource exhausted"),
			Entry("FailedPrecondition", FailedPrecondition, "failed precondition"),
			Entry("Aborted", Aborted, "aborted"),
			Entry("Unavailable", Unavailable, "unavailable"),
			Entry("NotImplemented", NotImplemented, "not implemented"),
		)

		It("returns the numeric value of custom error codes", func() {
			code := NewCode(123)
			Expect(code.String()).To(Equal("123"))
		})
	})
})
