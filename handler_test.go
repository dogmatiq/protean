package protean_test

import (
	"bytes"
	"context"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"time"

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
		When("the URI path refers to the root", func() {
			BeforeEach(func() {
				request.URL.Path = "/"
				request.Method = http.MethodGet
			})

			When("the HTTP method is not GET", func() {
				BeforeEach(func() {
					request.Method = http.MethodPost
				})

				It("responds with an HTTP '405 Method Not Allowed' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusMethodNotAllowed,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.NotImplemented,
							"the HTTP POST method is not supported at this path, establish a websocket connection or POST to /<package>/<service>/<method>",
						),
					)

					expectStandardHeaders(response, false)

					Expect(invoked).To(BeFalse())
				})
			})

			When("the request is not a websocket upgrade", func() {
				It("responds with an HTTP '426 Upgrade Required' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusUpgradeRequired,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.NotImplemented,
							"the HTTP GET method is only supported for websocket connections, establish a websocket connection or POST to /<package>/<service>/<method>",
						),
					)

					expectStandardHeaders(response, false)

					Expect(invoked).To(BeFalse())
				})
			})
		})

		When("the URI path is not the root and does not refer to a known RPC method", func() {
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
							rpcerror.NotImplemented,
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
					"missing service & package",
					"/package",
					"POST to /<package>/<service>/<method> or establish a websocket connection",
				),
				Entry(
					"missing method",
					"/package/Service",
					"POST to /<package>/<service>/<method> or establish a websocket connection",
				),
				Entry(
					"extra segments",
					"/package/Service/Method/unknown",
					"POST to /<package>/<service>/<method> or establish a websocket connection",
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

		When("the URI path refers to a specific RPC method", func() {
			When("the HTTP method is not POST", func() {
				BeforeEach(func() {
					request.Method = http.MethodGet
				})

				It("responds with an HTTP '405 Method Not Allowed' status", func() {
					handler.ServeHTTP(response, request)

					expectError(
						response,
						http.StatusMethodNotAllowed,
						"text/plain; charset=utf-8; x-proto=protean.v1.Error",
						rpcerror.New(
							rpcerror.NotImplemented,
							"the HTTP GET method is not supported at this path, use POST or establish a websocket connection",
						),
					)

					expectStandardHeaders(response, true)

					Expect(invoked).To(BeFalse())
				})
			})
		})
	})
})

// expectStandardHeaders asserts that the response includes HTTP headers that
// should always be set for a path that refers to a value RPC method.
func expectStandardHeaders(
	response *httptest.ResponseRecorder,
	acceptPost bool,
) {
	Expect(response).To(HaveHTTPHeaderWithValue("Cache-Control", "no-store"))
	Expect(response).To(HaveHTTPHeaderWithValue("X-Content-Type-Options", "nosniff"))

	if acceptPost {
		Expect(response).To(HaveHTTPHeaderWithValue("Accept-Post", "application/vnd.google.protobuf, application/x-protobuf, application/json, text/plain"))
	} else {
		Expect(
			response.Header().Get("Accept-Post"),
		).To(
			HaveLen(0),
			"Accept-Post header should not be provided when POST is not implemented at the request path",
		)
	}
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
