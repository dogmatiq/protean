package protean_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/dogmatiq/iago/iotest"
	. "github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/testservice"
	"github.com/dogmatiq/protean/rpcerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("type Handler (HTTP POST)", func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		input    *testservice.Input
		output   *testservice.Output
		handler  Handler
		invoked  bool
		service  *testservice.Stub
		request  *http.Request
		response *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		format.TruncatedDiff = false

		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

		handler = NewHandler()

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

		response = httptest.NewRecorder()

		testservice.RegisterProteanTestService(handler, service)
	})

	// Use JustBeforeEach to set the body so that we can manipulate the input
	// message in BeforeEach blocks before it is marshaled.
	JustBeforeEach(func() {
		data, err := protojson.Marshal(input)
		Expect(err).ShouldNot(HaveOccurred())

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Length", strconv.Itoa(len(data)))
		request.Body = io.NopCloser(bytes.NewReader(data))
	})

	AfterEach(func() {
		format.TruncatedDiff = true
		cancel()
	})

	Describe("func ServeHTTP()", func() {
		When("the RPC method succeeds", func() {
			It("responds with the the RPC output message", func() {
				handler.ServeHTTP(response, request)

				Expect(response).To(HaveHTTPStatus(http.StatusOK))
				expectStandardHeaders(response, true)

				data, err := io.ReadAll(response.Body)
				Expect(err).ShouldNot(HaveOccurred())

				var out testservice.Output
				err = protojson.Unmarshal(data, &out)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(out.GetData()).To(Equal("<output>"))
			})
		})

		When("the RPC method returns an error", func() {
			DescribeTable(
				"it maps rpcerror.Code to an appropriate HTTP status code",
				func(errorCode rpcerror.Code, httpCode int) {
					service.UnaryFunc = func(
						ctx context.Context,
						in *testservice.Input,
					) (*testservice.Output, error) {
						return nil, rpcerror.New(errorCode, "<error>")
					}

					handler.ServeHTTP(response, request)

					expectError(
						response,
						httpCode,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(errorCode, "<error>"),
					)
					expectStandardHeaders(response, true)
				},
				Entry("Unknown", rpcerror.Unknown, http.StatusInternalServerError),
				Entry("InvalidInput", rpcerror.InvalidInput, http.StatusBadRequest),
				Entry("Unauthenticated", rpcerror.Unauthenticated, http.StatusUnauthorized),
				Entry("PermissionDenied", rpcerror.PermissionDenied, http.StatusForbidden),
				Entry("NotFound", rpcerror.NotFound, http.StatusNotFound),
				Entry("AlreadyExists", rpcerror.AlreadyExists, http.StatusConflict),
				Entry("ResourceExhausted", rpcerror.ResourceExhausted, http.StatusTooManyRequests),
				Entry("FailedPrecondition", rpcerror.FailedPrecondition, http.StatusBadRequest),
				Entry("Aborted", rpcerror.Aborted, http.StatusConflict),
				Entry("Unavailable", rpcerror.Unavailable, http.StatusServiceUnavailable),
				Entry("NotImplemented", rpcerror.NotImplemented, http.StatusNotImplemented),
			)

			It("does not include the error message from arbitrary errors", func() {
				service.UnaryFunc = func(
					ctx context.Context,
					in *testservice.Input,
				) (*testservice.Output, error) {
					return nil, errors.New("<error>")
				}

				handler.ServeHTTP(response, request)

				expectError(
					response,
					http.StatusInternalServerError,
					"application/json; x-proto=protean.v1.Error",
					rpcerror.New(
						rpcerror.Unknown,
						"the RPC method returned an unrecognized error",
					),
				)
				expectStandardHeaders(response, true)
			})
		})

		Context("unmarshaling", func() {
			When("the request uses the binary protocol buffers format", func() {
				DescribeTable(
					"it passes the input message to the RPC method",
					func(mediaType string) {
						data, err := proto.Marshal(input)
						Expect(err).ShouldNot(HaveOccurred())

						request.Header.Set("Content-Type", mediaType)
						request.Header.Set("Content-Length", strconv.Itoa(len(data)))
						request.Body = io.NopCloser(bytes.NewReader(data))

						handler.ServeHTTP(response, request)

						Expect(response).To(HaveHTTPStatus(http.StatusOK))
						expectStandardHeaders(response, true)

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
					expectStandardHeaders(response, true)

					Expect(invoked).To(BeTrue())
				})
			})

			When("the request uses the text protocol buffers format", func() {
				JustBeforeEach(func() {
					data, err := prototext.Marshal(input)
					Expect(err).ShouldNot(HaveOccurred())

					request.Header.Set("Content-Type", "text/plain")
					request.Header.Set("Content-Length", strconv.Itoa(len(data)))
					request.Body = io.NopCloser(bytes.NewReader(data))
				})

				It("it passes the input message to the RPC method", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					expectStandardHeaders(response, true)

					Expect(invoked).To(BeTrue())
				})
			})

			When("the request body can not be read", func() {
				JustBeforeEach(func() {
					request.Body = io.NopCloser(iotest.NewFailer(nil, nil))
				})

				It("it reponds with an HTTP '500 Internal Server Error' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusInternalServerError,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the request body could not be read",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the input message is invalid", func() {
				BeforeEach(func() {
					input.Data = ""
				})

				It("it reponds with an HTTP '400 Bad Request' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusBadRequest,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.InvalidInput,
							"the RPC input message is invalid: input data must not be empty",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("there is no Content-Length header", func() {
				JustBeforeEach(func() {
					request.Header.Del("Content-Length")
				})

				It("responds as normal", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					expectStandardHeaders(response, true)

					Expect(invoked).To(BeTrue())
				})
			})

			When("the Content-Length header specifies a length greater than the maximum RPC input size", func() {
				JustBeforeEach(func() {
					request.Header.Set(
						"Content-Length",
						strconv.Itoa(DefaultMaxRPCInputSize+1),
					)
				})

				It("responds with an HTTP '413 Request Entity Too Large' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusRequestEntityTooLarge,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the length specified by the Content-Length header exceeds the maximum allowable size",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the Content-Length header is invalid", func() {
				JustBeforeEach(func() {
					request.Header.Set("Content-Length", "1.12")
				})

				It("responds with an HTTP '400 Bad Request' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusBadRequest,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the Content-Length header is invalid",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the actual content length is different to the length specified by the Content-Length header", func() {
				JustBeforeEach(func() {
					data := make([]byte, 5)
					request.Body = io.NopCloser(bytes.NewReader(data))
				})

				It("responds with an HTTP '400 Bad Request' status when the actual content length is shorter", func() {
					request.Header.Set("Content-Length", "10")
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusBadRequest,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC input message length does not match the length specified by the Content-Length header",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})

				It("responds with an HTTP '400 Bad Request' status when the actual content length is longer", func() {
					request.Header.Set("Content-Length", "2")
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusBadRequest,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC input message length does not match the length specified by the Content-Length header",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the actual content length is greater than the maximum RPC input size", func() {
				JustBeforeEach(func() {
					data := make([]byte, DefaultMaxRPCInputSize+1)
					request.Header.Del("Content-Length")
					request.Body = io.NopCloser(bytes.NewReader(data))
				})

				It("responds with an HTTP '413 Request Entity Too Large' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusRequestEntityTooLarge,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC input message length exceeds the maximum allowable size",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the Content-Type header is missing or invalid", func() {
				DescribeTable(
					"it responds with an HTTP '400 Bad Request' status",
					func(contentType string) {
						request.Header.Set("Content-Type", contentType)

						handler.ServeHTTP(response, request)

						expectError(
							response,
							http.StatusBadRequest,
							"text/plain; charset=utf-8; x-proto=protean.v1.Error",
							rpcerror.New(
								rpcerror.Unknown,
								"the Content-Type header is missing or invalid",
							),
						)

						expectStandardHeaders(response, true)

						Expect(invoked).To(BeFalse())
					},
					Entry("empty content type", ""),
					Entry("malformed content type", "/leading-slash"),
				)
			})

			When("the request supplies an unsupported Content-Type header", func() {
				JustBeforeEach(func() {
					request.Header.Set("Content-Type", "text/xml")
				})

				It("responds with an HTTP '415 Unsupported Media Type' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusUnsupportedMediaType,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the server does not support the 'text/xml' media-type supplied by the client",
						).WithDetails(
							&proteanpb.SupportedMediaTypes{
								MediaTypes: []string{
									"application/vnd.google.protobuf",
									"application/x-protobuf",
									"application/json",
									"text/plain",
								},
							},
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the input message can not be unmarshaled", func() {
				JustBeforeEach(func() {
					request.Header.Set("Content-Type", "application/json")
					request.Header.Set("Content-Length", "1")
					request.Body = io.NopCloser(strings.NewReader("}"))
				})

				It("it reponds with an HTTP '400 Bad Request' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusBadRequest,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC input message could not be unmarshaled from the request body",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})
		})

		Context("marshaling & content negotation", func() {
			When("the Accept header is absent", func() {
				JustBeforeEach(func() {
					request.Header.Del("Accept")
				})

				It("encodes RPC output messages using the same media type as the RPC input message", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", "application/json; x-proto=protean.test.Output"))
					expectStandardHeaders(response, true)

					data, err := io.ReadAll(response.Body)
					Expect(err).ShouldNot(HaveOccurred())

					var out testservice.Output
					err = protojson.Unmarshal(data, &out)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(out.GetData()).To(Equal("<output>"))
				})

				It("encodes RPC errors using the same media type as the RPC input message", func() {
					service.UnaryFunc = func(
						ctx context.Context,
						in *testservice.Input,
					) (*testservice.Output, error) {
						return nil, errors.New("<error>")
					}

					handler.ServeHTTP(response, request)

					expectStandardHeaders(response, true)

					expectError(
						response,
						http.StatusInternalServerError,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC method returned an unrecognized error",
						),
					)
				})
			})

			When("the client prefers the binary protocol buffers format", func() {
				DescribeTable(
					"it encodes RPC output messages using the binary protocol buffers format",
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
						Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", mediaType+"; x-proto=protean.test.Output"))
						expectStandardHeaders(response, true)

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

				DescribeTable(
					"it encodes RPC errors using the binary protocol buffers format",
					func(mediaType string) {
						request.Header.Set(
							"Accept",
							fmt.Sprintf(
								"text/xml;q=0.1, text/plain;q=0.5, %s, application/json;q=0.75",
								mediaType,
							),
						)

						service.UnaryFunc = func(
							ctx context.Context,
							in *testservice.Input,
						) (*testservice.Output, error) {
							return nil, errors.New("<error>")
						}

						handler.ServeHTTP(response, request)

						expectError(
							response,
							http.StatusInternalServerError,
							mediaType+"; x-proto=protean.v1.Error",
							rpcerror.New(
								rpcerror.Unknown,
								"the RPC method returned an unrecognized error",
							),
						)
						expectStandardHeaders(response, true)
					},
					Entry("preferred media type", "application/vnd.google.protobuf"),
					Entry("alternative media type", "application/x-protobuf"),
				)
			})

			When("the client prefers the JSON protocol buffers format", func() {
				JustBeforeEach(func() {
					request.Header.Set(
						"Accept",
						"text/xml;q=0.1, text/plain;q=0.5, application/vnd.google.protobuf;q=0.75, application/x-protobuf;q=0.75, application/json",
					)
				})

				It("encodes RPC output messages using the JSON protocol buffers format", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", "application/json; x-proto=protean.test.Output"))
					expectStandardHeaders(response, true)

					data, err := io.ReadAll(response.Body)
					Expect(err).ShouldNot(HaveOccurred())

					var out testservice.Output
					err = protojson.Unmarshal(data, &out)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(out.GetData()).To(Equal("<output>"))
				})

				It("encodes RPC errors using the JSON protocol buffers format", func() {
					service.UnaryFunc = func(
						ctx context.Context,
						in *testservice.Input,
					) (*testservice.Output, error) {
						return nil, errors.New("<error>")
					}

					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusInternalServerError,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC method returned an unrecognized error",
						),
					)
					expectStandardHeaders(response, true)
				})
			})

			When("the client prefers the text protocol buffers format", func() {
				JustBeforeEach(func() {
					request.Header.Set(
						"Accept",
						"text/xml;q=0.1, text/plain, application/vnd.google.protobuf;q=0.75, application/x-protobuf;q=0.75, application/json;q=0.5",
					)
				})

				It("encodes RPC output messages using the text-based protocol buffers format", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", "text/plain; charset=utf-8; x-proto=protean.test.Output"))
					expectStandardHeaders(response, true)

					data, err := io.ReadAll(response.Body)
					Expect(err).ShouldNot(HaveOccurred())

					var out testservice.Output
					err = prototext.Unmarshal(data, &out)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(out.GetData()).To(Equal("<output>"))
				})

				It("encodes RPC errors using the text-based protocol buffers format", func() {
					service.UnaryFunc = func(
						ctx context.Context,
						in *testservice.Input,
					) (*testservice.Output, error) {
						return nil, errors.New("<error>")
					}

					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusInternalServerError,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC method returned an unrecognized error",
						),
					)
					expectStandardHeaders(response, true)
				})
			})

			When("the client does not accept any of the media types supported by the server", func() {
				JustBeforeEach(func() {
					request.Header.Set("Accept", "text/xml")
				})

				It("responds with an HTTP '406 Not Accepted' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusNotAcceptable,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the client does not accept any of the media-types supported by the server",
						).WithDetails(
							&proteanpb.SupportedMediaTypes{
								MediaTypes: []string{
									"application/vnd.google.protobuf",
									"application/x-protobuf",
									"application/json",
									"text/plain",
								},
							},
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the Accept header is malformed", func() {
				JustBeforeEach(func() {
					request.Header.Set("Accept", "garbage;x")
				})

				It("responds with an HTTP '400 Bad Request' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusBadRequest,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the Accept header is invalid",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the output message is invalid", func() {
				BeforeEach(func() {
					output.Data = ""
				})

				It("it reponds with an HTTP '500 Internal Server Error' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusInternalServerError,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the server produced an invalid RPC output message",
						),
					)

					expectStandardHeaders(response, true)
				})
			})

			When("the output message can not be marshaled", func() {
				BeforeEach(func() {
					output.Data = "\xc3\x28" // invalid UTF-8
				})

				It("it reponds with an HTTP '500 Internal Server Error' status ", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusInternalServerError,
						"application/json; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.Unknown,
							"the RPC output message could not be marshaled to the response body",
						),
					)

					expectStandardHeaders(response, true)
				})
			})
		})

		When("the URI path refers to a streaming RPC method", func() {
			DescribeTable(
				"it responds with an HTTP '501 Not Implemented' status",
				func(path, message string) {
					request.URL.Path = path

					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusNotImplemented,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.NotImplemented,
							message,
						),
					)

					expectStandardHeaders(response, false)

					Expect(invoked).To(BeFalse())
				},
				Entry(
					"client streaming method",
					"/protean.test/TestService/ClientStream",
					"the 'protean.test.TestService' service contains an RPC method named 'ClientStream', but it requires streaming and therefore must be called by establishing a websocket connection at /",
				),
				Entry(
					"server streaming method",
					"/protean.test/TestService/ServerStream",
					"the 'protean.test.TestService' service contains an RPC method named 'ServerStream', but it requires streaming and therefore must be called by establishing a websocket connection at /",
				),
				Entry(
					"bidirectional streaming method",
					"/protean.test/TestService/BidirectionalStream",
					"the 'protean.test.TestService' service contains an RPC method named 'BidirectionalStream', but it requires streaming and therefore must be called by establishing a websocket connection at /",
				),
			)
		})
	})
})
