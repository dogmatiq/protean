package runtime_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/testservice"
	"github.com/dogmatiq/protean/rpcerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("type Client", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		input   *testservice.Input
		output  *testservice.Output
		handler protean.Handler
		service *testservice.Stub
		server  *httptest.Server
		client  testservice.ProteanTestService
	)

	BeforeEach(func() {
		format.TruncatedDiff = false

		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

		handler = protean.NewHandler()

		input = &testservice.Input{
			Data: "<input>",
		}

		output = &testservice.Output{
			Data: "<output>",
		}

		service = &testservice.Stub{
			UnaryFunc: func(
				_ context.Context,
				in *testservice.Input,
			) (*testservice.Output, error) {
				defer GinkgoRecover()
				Expect(in.GetData()).To(Equal("<input>"))
				return output, nil
			},
		}

		testservice.RegisterProteanTestService(handler, service)

		server = httptest.NewServer(handler)

		baseURL, err := url.Parse(server.URL)
		Expect(err).ShouldNot(HaveOccurred())

		client = testservice.NewProteanTestServiceClient(baseURL)
	})

	AfterEach(func() {
		format.TruncatedDiff = true
		cancel()

		server.Close()
	})

	Describe("func Unary()", func() {
		When("the server behaves correctly", func() {
			When("the RPC method succeeds", func() {
				It("returns the RPC output message", func() {
					out, err := client.Unary(ctx, input)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(out.GetData()).To(Equal("<output>"))
				})
			})

			When("the RPC method returns an error", func() {
				It("returns an rpcerror.Error", func() {
					expect := rpcerror.New(
						rpcerror.PermissionDenied,
						"<error>",
					)

					service.UnaryFunc = func(
						ctx context.Context,
						in *testservice.Input,
					) (*testservice.Output, error) {
						return nil, expect
					}

					_, err := client.Unary(ctx, input)
					Expect(err).To(Equal(expect))
				})
			})

			When("the RPC input message can not be marshaled", func() {
				BeforeEach(func() {
					input.Data = "\xc3\x28" // invalid UTF-8
				})

				It("returns an error", func() {
					_, err := client.Unary(ctx, input)
					Expect(err).To(MatchError("unable to marshal RPC input message: string field contains invalid UTF-8"))
				})
			})

			When("the HTTP request can not be constructed", func() {
				It("returns an error", func() {
					_, err := client.Unary(nil, input)
					Expect(err).To(MatchError("unable to create HTTP request: net/http: nil Context"))
				})
			})

			When("the HTTP request can not be performed", func() {
				BeforeEach(func() {
					server.Close()
				})

				It("returns an error", func() {
					_, err := client.Unary(ctx, input)
					Expect(err).To(MatchError(MatchRegexp(
						"unable to perform HTTP request: Post .+: dial tcp .+: connect: connection refused",
					)))
				})
			})
		})

		When("the server does not function as expected", func() {
			When("the HTTP request times out", func() {
				BeforeEach(func() {
					server.Config.Handler = http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							time.Sleep(150 * time.Millisecond)
						},
					)
				})

				It("returns an error", func() {
					ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
					defer cancel()

					_, err := client.Unary(ctx, input)
					Expect(err).To(Equal(context.DeadlineExceeded))
				})
			})

			When("the reading the HTTP response body times out", func() {
				BeforeEach(func() {
					server.Config.Handler = http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)

							for {
								if ctx.Err() != nil {
									return
								}

								w.Write([]byte{0})
							}
						},
					)
				})

				It("returns an error", func() {
					ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
					defer cancel()

					_, err := client.Unary(ctx, input)
					Expect(err).To(Equal(context.DeadlineExceeded))
				})
			})

			When("the HTTP response has no Content-Type header", func() {
				BeforeEach(func() {
					server.Config.Handler = http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {},
					)
				})

				It("returns an error", func() {
					_, err := client.Unary(ctx, input)
					Expect(err).To(MatchError("unable to unmarshal RPC output message: response has no Content-Type header"))
				})
			})

			When("the HTTP response has an invalid Content-Type header", func() {
				BeforeEach(func() {
					server.Config.Handler = http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							w.Header().Add("Content-Type", "garbage;x")
						},
					)
				})

				It("returns an error", func() {
					_, err := client.Unary(ctx, input)
					Expect(err).To(MatchError("unable to unmarshal RPC output message: Content-Type header is invalid: mime: invalid media parameter"))
				})
			})

			When("the HTTP response uses an unsupported media type for an RPC output message", func() {
				BeforeEach(func() {
					server.Config.Handler = http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							w.Header().Add("Content-Type", "text/xml")
						},
					)
				})

				It("returns an error", func() {
					_, err := client.Unary(ctx, input)
					Expect(err).To(MatchError("unable to unmarshal RPC output message: unsupported media type (text/xml)"))
				})
			})

			When("the HTTP response uses an unsupported media type for an RPC error", func() {
				BeforeEach(func() {
					server.Config.Handler = http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							w.Header().Add("Content-Type", "text/xml")
							w.WriteHeader(http.StatusNotFound)
						},
					)
				})

				It("returns an error", func() {
					_, err := client.Unary(ctx, input)
					Expect(err).To(MatchError("unable to unmarshal RPC error: unsupported media type (text/xml)"))
				})
			})
		})
	})
})
