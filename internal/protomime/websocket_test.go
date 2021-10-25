package protomime_test

import (
	. "github.com/dogmatiq/protean/internal/protomime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("func MediaTypeFromWebSocketProtocol()", func() {
	DescribeTable(
		"it returns a media type when the protocol is well-formed",
		func(protocol, mediaType string) {
			t, ok := MediaTypeFromWebSocketProtocol(protocol)
			Expect(t).To(Equal(mediaType))
			Expect(ok).To(BeTrue())
		},
		Entry("binary #1", "protean.v1+application.vnd.google.protobuf", "application/vnd.google.protobuf"),
		Entry("binary #2", "protean.v1+application.x-protobuf", "application/x-protobuf"),
		Entry("JSON", "protean.v1+application.json", "application/json"),
		Entry("text", "protean.v1+text.plain", "text/plain"),
		Entry("other", "protean.v1+application.cbor", "application/cbor"),
	)

	DescribeTable(
		"it returns a false when the protocol is not well-formed",
		func(protocol string) {
			_, ok := MediaTypeFromWebSocketProtocol(protocol)
			Expect(ok).To(BeFalse())
		},
		Entry("empty", ""),
		Entry("no prefix separator", "v12.stomp"),
		Entry("incorrect prefix", "protean.v2+text.plain"),
		Entry("no dot after prefix", "protean.v1+garbage"),
	)

	It("returns a media type for each of the protocols in the WebSocketProtocols variable", func() {
		Expect(WebSocketProtocols).To(HaveLen(len(MediaTypes)))

		for _, p := range WebSocketProtocols {
			mediaType, ok := MediaTypeFromWebSocketProtocol(p)
			Expect(ok).To(BeTrue())
			Expect(IsSupportedMediaType(mediaType)).To(BeTrue())
		}
	})
})
