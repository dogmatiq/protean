package protean_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/dogmatiq/protean"
	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/internal/testservice"
	"github.com/dogmatiq/protean/rpcerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("type Handler (HTTP POST)", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		// input   *testservice.Input
		output  *testservice.Output
		handler Handler
		// invoked bool
		service *testservice.Stub
		server  *httptest.Server
	)

	BeforeEach(func() {
		format.TruncatedDiff = false

		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

		handler = NewHandler()

		// input = &testservice.Input{
		// 	Data: "<input>",
		// }

		output = &testservice.Output{
			Data: "<output>",
		}

		// invoked = false
		service = &testservice.Stub{
			UnaryFunc: func(
				_ context.Context,
				in *testservice.Input,
			) (*testservice.Output, error) {
				// invoked = true
				Expect(in.GetData()).To(Equal("<input>"))
				return output, nil
			},
		}

		testservice.RegisterProteanTestService(handler, service)

		server = httptest.NewServer(handler)
	})

	AfterEach(func() {
		format.TruncatedDiff = true
		cancel()
		server.Close()
	})

	Describe("func ServeHTTP()", func() {
		When("the websocket upgrade fails", func() {
			It("responds with a text-based RPC error", func() {
				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodPost, // expects GET
					server.URL+"/protean.test/TestService/Unary",
					nil,
				)
				Expect(err).ShouldNot(HaveOccurred())

				// Fool the handler into thinking this is a real websocket
				// upgrade reqeuest.
				req.Header.Set("Connection", "upgrade")
				req.Header.Set("Upgrade", "websocket")

				res, err := http.DefaultClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(res).To(HaveHTTPStatus(http.StatusMethodNotAllowed))
				Expect(res).To(HaveHTTPHeaderWithValue("Content-Type", "text/plain; charset=utf-8; x-proto=protean.v1.Error"))

				data, err := io.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())

				var protoErr proteanpb.Error
				err = protomime.TextUnmarshaler.Unmarshal(data, &protoErr)
				Expect(err).ShouldNot(HaveOccurred())

				actual, err := rpcerror.FromProto(&protoErr)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(actual.Code()).To(Equal(rpcerror.Unknown))
				Expect(actual.Message()).To(Equal("websocket: the client is not using the websocket protocol: request method is not GET"))
			})
		})
	})
})
