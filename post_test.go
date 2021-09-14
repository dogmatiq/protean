package protean_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/dogmatiq/iago/iotest"
	. "github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/testservice"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("type PostHandler", func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		input    *testservice.Input
		output   *testservice.Output
		handler  *PostHandler
		invoked  bool
		service  *testservice.Stub
		request  *http.Request
		response *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		format.TruncatedDiff = false

		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

		handler = &PostHandler{}

		input = &testservice.Input{
			Data: "<input>",
		}

		output = &testservice.Output{
			Data: "<output>",
		}

		invoked = false
		service = &testservice.Stub{
			UnaryFunc: func(
				_ context.Context,
				in *testservice.Input,
			) (*testservice.Output, error) {
				invoked = true
				Expect(in.GetData()).To(Equal("<input>"))
				return output, nil
			},
		}

		request = httptest.NewRequest(
			http.MethodPost,
			"/protean.test/TestService/Unary",
			nil,
		).WithContext(ctx)

		data, err := protojson.Marshal(input)
		Expect(err).ShouldNot(HaveOccurred())

		// Supply an empty JSON request by default.
		request.Header.Set("Content-Type", "application/json")
		request.Body = io.NopCloser(bytes.NewReader(data))

		response = httptest.NewRecorder()

		testservice.ProteanRegisterTestServiceServer(handler, service)
	})

	AfterEach(func() {
		format.TruncatedDiff = true
		cancel()
	})

	Describe("func ServeHTTP()", func() {
		Context("unmarshaling", func() {
			When("the request uses the binary protocol buffers format", func() {
				DescribeTable(
					"it passes the input message to the RPC method",
					func(mediaType string) {
						data, err := proto.Marshal(input)
						Expect(err).ShouldNot(HaveOccurred())

						request.Header.Set("Content-Type", mediaType)
						request.Body = io.NopCloser(bytes.NewReader(data))

						handler.ServeHTTP(response, request)

						Expect(response).To(HaveHTTPStatus(http.StatusOK))
						Expect(invoked).To(BeTrue())
					},
					Entry("preferred media type", "application/vnd.google.protobuf"),
					Entry("alternative media type", "application/x-protobuf"),
				)
			})

			When("the request uses the JSON protocol buffers format", func() {
				It("it passes the input message to the RPC method", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(invoked).To(BeTrue())
				})
			})

			When("the request uses the text protocol buffers format", func() {
				It("it passes the input message to the RPC method", func() {
					data, err := prototext.Marshal(input)
					Expect(err).ShouldNot(HaveOccurred())

					request.Header.Set("Content-Type", "text/plain")
					request.Body = io.NopCloser(bytes.NewReader(data))

					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(invoked).To(BeTrue())
				})
			})

			When("the request body can not be read", func() {
				BeforeEach(func() {
					request.Body = io.NopCloser(iotest.NewFailer(nil, nil))
				})

				It("it reponds with an HTTP '500 Internal Server Error' status", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError))
					Expect(response).To(HaveHTTPBody(
						"500 Internal Server Error\n\n" +
							"The request body could not be read.\n",
					))
					Expect(invoked).To(BeFalse())
				})
			})

			When("the input message can not be unmarshaled", func() {
				BeforeEach(func() {
					request.Header.Set("Content-Type", "application/json")
					request.Body = io.NopCloser(strings.NewReader("}"))
				})

				It("it reponds with an HTTP '400 Bad Request' status", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusBadRequest))
					Expect(response).To(HaveHTTPBody(
						"400 Bad Request\n\n" +
							"The RPC input message could not be unmarshaled from the request body.\n",
					))
					Expect(invoked).To(BeFalse())
				})
			})
		})

		Context("marshaling & content negotation", func() {
			When("the Accept header is absent", func() {
				BeforeEach(func() {
					request.Header.Del("Accept")
				})

				It("responds using the same media type as the request", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", "application/json"))

					data, err := io.ReadAll(response.Body)
					Expect(err).ShouldNot(HaveOccurred())

					var out testservice.Output
					err = protojson.Unmarshal(data, &out)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(out.GetData()).To(Equal("<output>"))
				})
			})

			When("the client prefers the binary protocol buffers format", func() {
				DescribeTable(
					"it responds using the binary protocol buffers format",
					func(mediaType string) {
						request.Header.Set(
							"Accept",
							fmt.Sprintf(
								"text/xml;q=0.1, text/plain;q=0.5, %s, application/json;q=0.75",
								mediaType,
							),
						)

						handler.ServeHTTP(response, request)

						Expect(response).To(HaveHTTPStatus(http.StatusOK))
						Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", mediaType))

						data, err := io.ReadAll(response.Body)
						Expect(err).ShouldNot(HaveOccurred())

						var out testservice.Output
						err = proto.Unmarshal(data, &out)
						Expect(err).ShouldNot(HaveOccurred())

						Expect(out.GetData()).To(Equal("<output>"))
					},
					Entry("preferred media type", "application/vnd.google.protobuf"),
					Entry("alternative media type", "application/x-protobuf"),
				)
			})

			When("the client prefers the JSON protocol buffers format", func() {
				BeforeEach(func() {
					request.Header.Set(
						"Accept",
						"text/xml;q=0.1, text/plain;q=0.5, application/vnd.google.protobuf;q=0.75, application/x-protobuf;q=0.75, application/json",
					)
				})

				It("responds using the JSON protocol buffers format", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", "application/json"))

					data, err := io.ReadAll(response.Body)
					Expect(err).ShouldNot(HaveOccurred())

					var out testservice.Output
					err = protojson.Unmarshal(data, &out)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(out.GetData()).To(Equal("<output>"))
				})
			})

			When("the client prefers the text protocol buffers format", func() {
				BeforeEach(func() {
					request.Header.Set(
						"Accept",
						"text/xml;q=0.1, text/plain, application/vnd.google.protobuf;q=0.75, application/x-protobuf;q=0.75, application/json;q=0.5",
					)
				})

				It("responds using the JSON protocol buffers format", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", "text/plain"))

					data, err := io.ReadAll(response.Body)
					Expect(err).ShouldNot(HaveOccurred())

					var out testservice.Output
					err = prototext.Unmarshal(data, &out)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(out.GetData()).To(Equal("<output>"))
				})
			})

			When("the client does not accept any of the media types supported by the server", func() {
				BeforeEach(func() {
					request.Header.Set("Accept", "text/xml")
				})

				It("responds with an HTTP '406 Not Accepted' status", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusNotAcceptable))
					Expect(response).To(HaveHTTPBody(
						"406 Not Acceptable\n\n" +
							"The client does not accept any of the media-types supported by the server.\n\n" +
							"The supported types are, in order of preference:\n" +
							"- application/vnd.google.protobuf\n" +
							"- application/x-protobuf\n" +
							"- application/json\n" +
							"- text/plain\n",
					))
					Expect(invoked).To(BeFalse())
				})
			})

			When("the Accept header is malformed", func() {
				BeforeEach(func() {
					request.Header.Set("Accept", "garbage;x")
				})

				It("responds with an HTTP '400 Bad Request' status", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusBadRequest))
					Expect(response).To(HaveHTTPBody(
						"400 Bad Request\n\n" +
							"The Accept header is invalid.\n",
					))
					Expect(invoked).To(BeFalse())
				})
			})

			When("the output message can not be marshaled", func() {
				It("it reponds with an HTTP '500 Internal Server Error' status ", func() {
					output.Data = "\xc3\x28" // invalid UTF-8

					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError))
					Expect(response).To(HaveHTTPBody(
						"500 Internal Server Error\n\n" +
							"The RPC output message could not be marshaled to the response body.\n",
					))
				})
			})
		})

		When("the URI path does not refer to a known RPC method", func() {
			DescribeTable(
				"it responds with an 'HTTP 404 Not Found' status",
				func(path, message string) {
					request.URL.Path = path

					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusNotFound))
					Expect(response).To(HaveHTTPBody(
						fmt.Sprintf(
							"404 Not Found\n\n%s\n",
							message,
						),
					))
					Expect(invoked).To(BeFalse())
				},
				Entry(
					"root",
					"/",
					"The request URI must follow the '/<package>/<service>/<method>' pattern.",
				),
				Entry(
					"missing service & package",
					"/package",
					"The request URI must follow the '/<package>/<service>/<method>' pattern.",
				),
				Entry(
					"missing method",
					"/package/Service",
					"The request URI must follow the '/<package>/<service>/<method>' pattern.",
				),
				Entry(
					"extra segments",
					"/package/Service/Method/unknown",
					"The request URI must follow the '/<package>/<service>/<method>' pattern.",
				),

				Entry(
					"unknown service",
					"/package/Service/Method",
					"The server does not provide the 'package.Service' service.",
				),
				Entry(
					"unknown method",
					"/protean.test/TestService/Method",
					"The 'protean.test.TestService' service does not contain an RPC method named 'Method'.",
				),

				Entry(
					"client streaming method",
					"/protean.test/TestService/ClientStream",
					"An RPC method named 'ClientStream' exists, but is not supported by this server because it uses streaming inputs or outputs.",
				),
				Entry(
					"server streaming method",
					"/protean.test/TestService/ServerStream",
					"An RPC method named 'ServerStream' exists, but is not supported by this server because it uses streaming inputs or outputs.",
				),
				Entry(
					"bidirectional streaming method",
					"/protean.test/TestService/BidirectionalStream",
					"An RPC method named 'BidirectionalStream' exists, but is not supported by this server because it uses streaming inputs or outputs.",
				),
			)
		})

		When("the HTTP method is not POST", func() {
			BeforeEach(func() {
				request.Method = http.MethodGet
			})

			It("responds with an 'HTTP 501 Not Implemented' status", func() {
				handler.ServeHTTP(response, request)

				Expect(response).To(HaveHTTPStatus(http.StatusNotImplemented))
				Expect(response).To(HaveHTTPBody(
					"501 Not Implemented\n\n" +
						"The HTTP method must be POST.\n",
				))
				Expect(invoked).To(BeFalse())
			})
		})

		When("the Content-Type header is missing or invalid", func() {
			DescribeTable(
				"it responds with an HTTP '400 Bad Request' status",
				func(contentType string) {
					request.Header.Set("Content-Type", contentType)

					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusBadRequest))
					Expect(response).To(HaveHTTPBody(
						"400 Bad Request\n\n" +
							"The Content-Type header is missing or invalid.\n",
					))
					Expect(invoked).To(BeFalse())
				},
				Entry("empty content type", ""),
				Entry("malformed content type", "/leading-slash"),
			)
		})

		When("the request supplies an unsupported Content-Type header", func() {
			BeforeEach(func() {
				request.Header.Set("Content-Type", "text/xml")
			})

			It("responds with an HTTP '415 Unsupported Media Type' status", func() {
				handler.ServeHTTP(response, request)

				Expect(response).To(HaveHTTPStatus(http.StatusUnsupportedMediaType))
				Expect(response).To(HaveHTTPBody(
					"415 Unsupported Media Type\n\n" +
						"The server does not support the 'text/xml' media-type supplied by the client.\n\n" +
						"The supported types are, in order of preference:\n" +
						"- application/vnd.google.protobuf\n" +
						"- application/x-protobuf\n" +
						"- application/json\n" +
						"- text/plain\n",
				))
				Expect(invoked).To(BeFalse())
			})
		})
	})
})
