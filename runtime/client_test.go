package runtime_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/testservice"
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
		It("invokes the RPC method on the server", func() {
			out, err := client.Unary(
				ctx,
				input,
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out.GetData()).To(Equal("<output>"))
		})
	})
})
