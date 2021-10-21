package protean_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
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
		Context("unary RPC methods", func() {
			var (
				input  *testservice.Input
				output *testservice.Output
				conn   *websocket.Conn
			)

			BeforeEach(func() {
				input = &testservice.Input{
					Data: "<input>",
				}

				output = &testservice.Output{
					Data: "<output>",
				}

				service.UnaryFunc = func(
					_ context.Context,
					in *testservice.Input,
				) (*testservice.Output, error) {
					Expect(in.GetData()).To(Equal("<input>"))
					return output, nil
				}

				var err error
				conn, _, err = websocket.DefaultDialer.DialContext(
					ctx,
					webSocketURL,
					nil,
				)
				Expect(err).ShouldNot(HaveOccurred())

				dl, _ := ctx.Deadline()
				conn.SetReadDeadline(dl)
				conn.SetWriteDeadline(dl)
			})

			AfterEach(func() {
				if conn != nil {
					conn.Close()
				}
			})

			When("the RPC method succeeds", func() {
				It("responds with a single output message", func() {
					By("sending a call frame")

					clientEnv := &proteanpb.ClientEnvelope{
						Channel: 1,
						Frame: &proteanpb.ClientEnvelope_Call{
							Call: "protean.test/TestService/Unary",
						},
					}

					data, err := protomime.JSONMarshaler.Marshal(clientEnv)
					Expect(err).ShouldNot(HaveOccurred())

					err = conn.WriteMessage(websocket.TextMessage, data)
					Expect(err).ShouldNot(HaveOccurred())

					By("sending an input frame")

					m, err := anypb.New(input)
					Expect(err).ShouldNot(HaveOccurred())

					clientEnv = &proteanpb.ClientEnvelope{
						Channel: 1,
						Frame: &proteanpb.ClientEnvelope_Input{
							Input: m,
						},
					}

					data, err = protomime.JSONMarshaler.Marshal(clientEnv)
					Expect(err).ShouldNot(HaveOccurred())

					err = conn.WriteMessage(websocket.TextMessage, data)
					Expect(err).ShouldNot(HaveOccurred())

					By("sending a done frame")

					clientEnv = &proteanpb.ClientEnvelope{
						Channel: 1,
						Frame: &proteanpb.ClientEnvelope_Done{
							Done: true,
						},
					}

					data, err = protomime.JSONMarshaler.Marshal(clientEnv)
					Expect(err).ShouldNot(HaveOccurred())

					err = conn.WriteMessage(websocket.TextMessage, data)
					Expect(err).ShouldNot(HaveOccurred())

					By("reading the output frame")

					var serverEnv proteanpb.ServerEnvelope

					t, data, err := conn.ReadMessage()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(t).To(Equal(websocket.TextMessage))

					err = protojson.Unmarshal(data, &serverEnv)
					Expect(err).ShouldNot(HaveOccurred())

					var out testservice.Output
					err = serverEnv.GetOutput().UnmarshalTo(&out)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(out.GetData()).To(Equal("<output>"))
					Expect(serverEnv.Channel).To(BeNumerically("==", 1))

					By("reading the done frame")

					t, data, err = conn.ReadMessage()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(t).To(Equal(websocket.TextMessage))

					err = protojson.Unmarshal(data, &serverEnv)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(serverEnv.GetDone()).To(BeTrue())
					Expect(serverEnv.Channel).To(BeNumerically("==", 1))

					By("reading the success frame")

					t, data, err = conn.ReadMessage()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(t).To(Equal(websocket.TextMessage))

					err = protojson.Unmarshal(data, &serverEnv)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(serverEnv.GetSuccess()).To(BeTrue())
					Expect(serverEnv.Channel).To(BeNumerically("==", 1))
				})
			})

			// When("the RPC method returns an error", func() {
			// 	DescribeTable(
			// 		"it maps rpcerror.Code to an appropriate HTTP status code",
			// 		func(errorCode rpcerror.Code, httpCode int) {
			// 			service.UnaryFunc = func(
			// 				ctx context.Context,
			// 				in *testservice.Input,
			// 			) (*testservice.Output, error) {
			// 				return nil, rpcerror.New(errorCode, "<error>")
			// 			}

			// 			handler.ServeHTTP(response, request)

			// 			expectError(
			// 				response,
			// 				httpCode,
			// 				"application/json; x-proto=protean.v1.Error",
			// 				rpcerror.New(errorCode, "<error>"),
			// 			)
			// 			expectStandardHeaders(response)
			// 		},
			// 		Entry("Unknown", rpcerror.Unknown, http.StatusInternalServerError),
			// 		Entry("InvalidInput", rpcerror.InvalidInput, http.StatusBadRequest),
			// 		Entry("Unauthenticated", rpcerror.Unauthenticated, http.StatusUnauthorized),
			// 		Entry("PermissionDenied", rpcerror.PermissionDenied, http.StatusForbidden),
			// 		Entry("NotFound", rpcerror.NotFound, http.StatusNotFound),
			// 		Entry("AlreadyExists", rpcerror.AlreadyExists, http.StatusConflict),
			// 		Entry("ResourceExhausted", rpcerror.ResourceExhausted, http.StatusTooManyRequests),
			// 		Entry("FailedPrecondition", rpcerror.FailedPrecondition, http.StatusBadRequest),
			// 		Entry("Aborted", rpcerror.Aborted, http.StatusConflict),
			// 		Entry("Unavailable", rpcerror.Unavailable, http.StatusServiceUnavailable),
			// 		Entry("NotImplemented", rpcerror.NotImplemented, http.StatusNotImplemented),
			// 	)

			// 	It("does not include the error message from arbitrary errors", func() {
			// 		service.UnaryFunc = func(
			// 			ctx context.Context,
			// 			in *testservice.Input,
			// 		) (*testservice.Output, error) {
			// 			return nil, errors.New("<error>")
			// 		}

			// 		handler.ServeHTTP(response, request)

			// 		expectError(
			// 			response,
			// 			http.StatusInternalServerError,
			// 			"application/json; x-proto=protean.v1.Error",
			// 			rpcerror.New(
			// 				rpcerror.Unknown,
			// 				"the RPC method returned an unrecognized error",
			// 			),
			// 		)
			// 		expectStandardHeaders(response)
			// 	})
			// })
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
