package protomime_test

import (
	. "github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/internal/testservice"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("type IsSupportedMediaType()", func() {
	DescribeTable(
		"it returns true for supported media types",
		func(mediaType string) {
			Expect(IsSupportedMediaType(mediaType)).To(BeTrue())
		},
		Entry("binary #1", "application/vnd.google.protobuf"),
		Entry("binary #2", "application/x-protobuf"),
		Entry("JSON", "application/json"),
		Entry("text", "text/plain"),
	)

	It("returns false for unsupported media types", func() {
		Expect(IsSupportedMediaType("text/xml")).To(BeFalse())
	})
})

var _ = Describe("type IsBinary()", func() {
	DescribeTable(
		"it returns true for binary media types",
		func(mediaType string) {
			Expect(IsBinary(mediaType)).To(BeTrue())
		},
		Entry("binary #1", "application/vnd.google.protobuf"),
		Entry("binary #2", "application/x-protobuf"),
	)

	DescribeTable(
		"it returns true false for other media types",
		func(mediaType string) {
			Expect(IsBinary(mediaType)).To(BeFalse())
		},
		Entry("JSON", "application/json"),
		Entry("text", "text/plain"),
	)
})

var _ = Describe("type IsJSON()", func() {
	DescribeTable(
		"it returns true for binary media types",
		func(mediaType string) {
			Expect(IsJSON(mediaType)).To(BeTrue())
		},
		Entry("JSON", "application/json"),
	)

	DescribeTable(
		"it returns true false for other media types",
		func(mediaType string) {
			Expect(IsJSON(mediaType)).To(BeFalse())
		},
		Entry("binary #1", "application/vnd.google.protobuf"),
		Entry("binary #2", "application/x-protobuf"),
		Entry("text", "text/plain"),
	)
})

var _ = Describe("type IsText()", func() {
	DescribeTable(
		"it returns true for binary media types",
		func(mediaType string) {
			Expect(IsText(mediaType)).To(BeTrue())
		},
		Entry("text", "text/plain"),
	)

	DescribeTable(
		"it returns true false for other media types",
		func(mediaType string) {
			Expect(IsText(mediaType)).To(BeFalse())
		},
		Entry("binary #1", "application/vnd.google.protobuf"),
		Entry("binary #2", "application/x-protobuf"),
		Entry("JSON", "application/json"),
	)
})

var _ = Describe("type FormatMediaType()", func() {
	It("adds the x-proto parameter", func() {
		mediaType := FormatMediaType("application/json", &testservice.Input{})
		Expect(mediaType).To(Equal("application/json; x-proto=protean.test.Input"))
	})

	It("adds the charset parameter for text media-types", func() {
		mediaType := FormatMediaType("text/plain", &testservice.Input{})
		Expect(mediaType).To(Equal("text/plain; charset=utf-8; x-proto=protean.test.Input"))
	})
})
