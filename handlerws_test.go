package protean_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	. "github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/internal/testservice"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("type Handler (websocket)", func() {
	var (
		ctx          context.Context
		cancel       context.CancelFunc
		handler      Handler
		service      *testservice.Stub
		server       *httptest.Server
		webSocketURL string
	)

	BeforeEach(func() {
		format.TruncatedDiff = false

		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

		handler = NewHandler()

		service = &testservice.Stub{}
		testservice.RegisterProteanTestService(handler, service)

		server = httptest.NewServer(handler)

		webSocketURL = strings.Replace(server.URL, "http", "ws", -1)
	})

	AfterEach(func() {
		format.TruncatedDiff = true
		cancel()
		server.Close()
	})

	Describe("func ServeHTTP()", func() {
		When("the websocket connection is established successfully", func() {
			var (
				conn *websocket.Conn
			)

			BeforeEach(func() {
				var err error
				conn, _, err = websocket.DefaultDialer.DialContext(
					ctx,
					webSocketURL,
					nil,
				)
				Expect(err).ShouldNot(HaveOccurred())
			})

			AfterEach(func() {
				if conn != nil {
					conn.Close()
				}
			})

			When("the client sends a frame that can not be unmarshaled", func() {
				It("closes the connection with a 'invalid frame payload' code", func() {
					err := conn.WriteMessage(websocket.TextMessage, []byte("}"))
					Expect(err).ShouldNot(HaveOccurred())

					_, _, err = conn.ReadMessage()
					Expect(err).To(MatchError(MatchRegexp(
						`websocket: close 1007 \(invalid payload data\): proto:.+syntax error \(line 1:1\): unexpected token }`,
					)))
				})
			})

			When("the client sends a 'call' frame with an unexpected call ID", func() {
				DescribeTable(
					"it closes the connection with a 'protocol error' code",
					func(callID int) {
						stringCallID := strconv.Itoa(callID)

						err := conn.WriteMessage(websocket.TextMessage, []byte(
							`{ "call_id": 456, "call": "protean.test/TestService/Unary" }`,
						))
						Expect(err).ShouldNot(HaveOccurred())

						err = conn.WriteMessage(websocket.TextMessage, []byte(
							`{ "call_id": `+stringCallID+`, "call": "protean.test/TestService/Unary" }`,
						))
						Expect(err).ShouldNot(HaveOccurred())

						for {
							_, _, err = conn.ReadMessage()

							// We may receive valid frames from the first call
							// before the connection is closed as expected.
							if err != nil {
								Expect(err).To(MatchError(
									`websocket: close 1002 (protocol error): out-of-sequence call ID in 'call' frame (` + stringCallID + `), expected >=457`,
								))
								break
							}
						}
					},
					Entry("lower than previous call ID", 123),
					Entry("same as previous call ID", 456),
				)
			})
		})

		Context("sub-protocol negotiation", func() {
			When("the client does specifies a supported sub-protocol", func() {
				DescribeTable(
					"it uses that sub-protocol",
					func(protocol string) {
						conn, res, err := websocket.DefaultDialer.DialContext(
							ctx,
							webSocketURL,
							http.Header{
								"Sec-WebSocket-Protocol": {protocol},
							},
						)
						Expect(err).ShouldNot(HaveOccurred())
						defer conn.Close()

						Expect(res).To(HaveHTTPHeaderWithValue("Sec-WebSocket-Protocol", protocol))
					},
					Entry("binary #1", "protean.v1+application.vnd.google.protobuf"),
					Entry("binary #2", "protean.v1+application.x-protobuf"),
					Entry("JSON", "protean.v1+application.json"),
					Entry("text", "protean.v1+text.plain"),
				)
			})

			When("the client does not specify a sub-protocol", func() {
				It("defaults to the JSON sub-protocol", func() {
					conn, res, err := websocket.DefaultDialer.DialContext(
						ctx,
						webSocketURL,
						nil,
					)
					Expect(err).ShouldNot(HaveOccurred())
					defer conn.Close()

					Expect(res).To(HaveHTTPHeaderWithValue("Sec-WebSocket-Protocol", "protean.v1+application.json"))
				})
			})

			When("the client specifies an unsupported sub-protocol", func() {
				It("defaults to the JSON sub-protocol", func() {
					conn, res, err := websocket.DefaultDialer.DialContext(
						ctx,
						webSocketURL,
						http.Header{
							"Sec-WebSocket-Protocol": {"garbage"},
						},
					)
					Expect(err).ShouldNot(HaveOccurred())
					defer conn.Close()

					Expect(res).To(HaveHTTPHeaderWithValue("Sec-WebSocket-Protocol", "protean.v1+application.json"))
				})
			})
		})

		When("the websocket upgrade fails", func() {
			It("responds with a text-based RPC error", func() {
				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodPost, // expects GET
					server.URL,
					nil,
				)
				Expect(err).ShouldNot(HaveOccurred())

				// Fool the handler into thinking this is a real websocket
				// upgrade reqeuest.
				req.Header.Set("Connection", "upgrade")
				req.Header.Set("Upgrade", "websocket")

				res, err := http.DefaultClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())

				expectWebSocketError(res, "websocket: the client is not using the websocket protocol: request method is not GET")
			})
		})
	})
})

// expectWebSocketError asserts that the response describes the expected error.
func expectWebSocketError(
	response *http.Response,
	message string,
) {
	Expect(response).To(HaveHTTPStatus(http.StatusMethodNotAllowed))
	Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", "text/plain; charset=utf-8; x-proto=protean.v1.Error"))

	data, err := io.ReadAll(response.Body)
	Expect(err).ShouldNot(HaveOccurred())

	var protoErr proteanpb.Error
	err = protomime.TextUnmarshaler.Unmarshal(data, &protoErr)
	Expect(err).ShouldNot(HaveOccurred())

	actual, err := rpcerror.FromProto(&protoErr)
	Expect(err).ShouldNot(HaveOccurred())

	Expect(actual.Code()).To(Equal(rpcerror.Unknown))
	Expect(actual.Message()).To(Equal(message))
}
