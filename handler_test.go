package protean_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/dogmatiq/iago/iotest"
	. "github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
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

var _ = Describe("type Handler", func() {
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
				expectStandardHeaders(response)

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
					expectStandardHeaders(response)
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
				expectStandardHeaders(response)
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
						request.Body = io.NopCloser(bytes.NewReader(data))

						handler.ServeHTTP(response, request)

						Expect(response).To(HaveHTTPStatus(http.StatusOK))
						expectStandardHeaders(response)

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
					expectStandardHeaders(response)

					Expect(invoked).To(BeTrue())
				})
			})

			When("the request uses the text protocol buffers format", func() {
				JustBeforeEach(func() {
					data, err := prototext.Marshal(input)
					Expect(err).ShouldNot(HaveOccurred())

					request.Header.Set("Content-Type", "text/plain")
					request.Body = io.NopCloser(bytes.NewReader(data))
				})

				It("it passes the input message to the RPC method", func() {
					handler.ServeHTTP(response, request)

					Expect(response).To(HaveHTTPStatus(http.StatusOK))
					expectStandardHeaders(response)

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

					expectStandardHeaders(response)

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

					expectStandardHeaders(response)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the input message can not be unmarshaled", func() {
				JustBeforeEach(func() {
					request.Header.Set("Content-Type", "application/json")
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

					expectStandardHeaders(response)

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
					expectStandardHeaders(response)

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

					expectStandardHeaders(response)

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
						expectStandardHeaders(response)

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
						expectStandardHeaders(response)
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
					expectStandardHeaders(response)

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
					expectStandardHeaders(response)
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
					expectStandardHeaders(response)

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
					expectStandardHeaders(response)
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

					expectStandardHeaders(response)

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

					expectStandardHeaders(response)

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

					expectStandardHeaders(response)
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

					expectStandardHeaders(response)
				})
			})
		})

		When("the URI path does not refer to a known RPC method", func() {
			DescribeTable(
				"it responds with an HTTP '404 Not Found' status",
				func(path, message string) {
					request.URL.Path = path

					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusNotFound,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.NotFound,
							message,
						),
					)

					Expect(
						response.Header().Get("Accept-Post"),
					).To(
						HaveLen(0),
						"Accept-Post header should not be provided when the request path does not refer to an RPC method",
					)

					Expect(invoked).To(BeFalse())
				},
				Entry(
					"root",
					"/",
					"the request URI must follow the '/<package>/<service>/<method>' pattern",
				),
				Entry(
					"missing service & package",
					"/package",
					"the request URI must follow the '/<package>/<service>/<method>' pattern",
				),
				Entry(
					"missing method",
					"/package/Service",
					"the request URI must follow the '/<package>/<service>/<method>' pattern",
				),
				Entry(
					"extra segments",
					"/package/Service/Method/unknown",
					"the request URI must follow the '/<package>/<service>/<method>' pattern",
				),
				Entry(
					"unknown service",
					"/package/Service/Method",
					"the server does not provide the 'package.Service' service",
				),
				Entry(
					"unknown method",
					"/protean.test/TestService/Method",
					"the 'protean.test.TestService' service does not contain an RPC method named 'Method'",
				),
			)
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

					Expect(
						response.Header().Get("Accept-Post"),
					).To(
						HaveLen(0),
						"Accept-Post header should not be provided when POST is not implemented at the request path",
					)

					Expect(invoked).To(BeFalse())
				},
				Entry(
					"client streaming method",
					"/protean.test/TestService/ClientStream",
					"the 'protean.test.TestService' service does contain an RPC method named 'ClientStream', but is not supported by this server because it uses streaming inputs or outputs",
				),
				Entry(
					"server streaming method",
					"/protean.test/TestService/ServerStream",
					"the 'protean.test.TestService' service does contain an RPC method named 'ServerStream', but is not supported by this server because it uses streaming inputs or outputs",
				),
				Entry(
					"bidirectional streaming method",
					"/protean.test/TestService/BidirectionalStream",
					"the 'protean.test.TestService' service does contain an RPC method named 'BidirectionalStream', but is not supported by this server because it uses streaming inputs or outputs",
				),
			)
		})

		When("the HTTP method is not POST", func() {
			BeforeEach(func() {
				request.Method = http.MethodGet
			})

			It("responds with an HTTP '501 Not Implemented' status", func() {
				handler.ServeHTTP(response, request)

				expectError(
					response,
					http.StatusNotImplemented,
					"text/plain; charset=utf-8; x-proto=protean.v1.Error",
					rpcerror.New(
						rpcerror.NotImplemented,
						"the HTTP method must be POST",
					),
				)

				expectStandardHeaders(response)

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

					expectStandardHeaders(response)

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

				expectStandardHeaders(response)

				Expect(invoked).To(BeFalse())
			})
		})
	})
})

// expectStandardHeaders asserts that the response includes HTTP headers that
// should always be set for a path that refers to a value RPC method.
func expectStandardHeaders(
	response *httptest.ResponseRecorder,
) {
	Expect(response).To(HaveHTTPHeaderWithValue("Cache-Control", "no-store"))
	Expect(response).To(HaveHTTPHeaderWithValue("X-Content-Type-Options", "nosniff"))
	Expect(response).To(HaveHTTPHeaderWithValue("Accept-Post", "application/vnd.google.protobuf, application/x-protobuf, application/json, text/plain"))
}

// expectError asserts that the response describes the expected error.
func expectError(
	response *httptest.ResponseRecorder,
	status int,
	mediaType string,
	expect rpcerror.Error,
) {
	var protoErr proteanpb.Error

	Expect(response).To(HaveHTTPStatus(status))
	Expect(response).To(HaveHTTPHeaderWithValue("Content-Type", mediaType))

	data, err := io.ReadAll(response.Body)
	Expect(err).ShouldNot(HaveOccurred())

	mediaType, _, _ = mime.ParseMediaType(mediaType)
	unmarshaler, ok := protomime.UnmarshalerForMediaType(mediaType)
	Expect(ok).To(BeTrue())

	err = unmarshaler.Unmarshal(data, &protoErr)
	Expect(err).ShouldNot(HaveOccurred())

	actual, err := rpcerror.FromProto(&protoErr)
	Expect(err).ShouldNot(HaveOccurred())

	Expect(actual.Code()).To(Equal(expect.Code()))
	Expect(actual.Message()).To(Equal(expect.Message()))

	expectDetails, ok, err := expect.Details()
	Expect(err).ShouldNot(HaveOccurred())

	if ok {
		actualDetails, ok, err := actual.Details()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(BeTrue())
		Expect(proto.Equal(expectDetails, actualDetails)).To(BeTrue(), "error details do not match")
	}
}
