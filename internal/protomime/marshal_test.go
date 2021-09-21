package protomime_test

import (
	. "github.com/dogmatiq/protean/internal/protomime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("func MarshalerForMediaType()", func() {
	DescribeTable(
		"it returns the expected marshaler for each supported media type",
		func(mediaType string, marshaler Marshaler) {
			m, ok := MarshalerForMediaType(mediaType)
			Expect(m).To(BeIdenticalTo(marshaler))
			Expect(ok).To(BeTrue())
		},
		Entry("binary #1", "application/vnd.google.protobuf", BinaryMarshaler),
		Entry("binary #2", "application/x-protobuf", BinaryMarshaler),
		Entry("JSON", "application/json", JSONMarshaler),
		Entry("text", "text/plain", TextMarshaler),
	)

	It("returns false for unsupported media types", func() {
		_, ok := MarshalerForMediaType("text/xml")
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("func UnmarshalerForMediaType()", func() {
	DescribeTable(
		"it returns the expected unmarshaler for each supported media type",
		func(mediaType string, unmarshaler Unmarshaler) {
			u, ok := UnmarshalerForMediaType(mediaType)
			Expect(u).To(BeIdenticalTo(unmarshaler))
			Expect(ok).To(BeTrue())
		},
		Entry("binary #1", "application/vnd.google.protobuf", BinaryUnmarshaler),
		Entry("binary #2", "application/x-protobuf", BinaryUnmarshaler),
		Entry("JSON", "application/json", JSONUnmarshaler),
		Entry("text", "text/plain", TextUnmarshaler),
	)

	It("returns false for unsupported media types", func() {
		_, ok := UnmarshalerForMediaType("text/xml")
		Expect(ok).To(BeFalse())
	})
})
