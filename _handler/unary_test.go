package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/dogmatiq/protean/handler"
	"github.com/dogmatiq/protean/internal/testservice"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type unaryHandler (via Handler)", func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		handler  *Handler
		service  *testservice.Stub
		response *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

		handler = &Handler{}
		service = &testservice.Stub{}
		response = httptest.NewRecorder()

		testservice.ProteanRegisterTestServiceServer(handler, service)
	})

	AfterEach(func() {
		cancel()
	})

	Describe("func ServeHTTP()", func() {
		When("the request uses the HTTP GET method", func() {
			XIt("invokes the service with a zero-valued request", func() {
				request := httptest.NewRequest(
					http.MethodGet,
					"/protean.test.TestService/Unary",
					nil,
				).WithContext(ctx)

				service.UnaryFunc = func(
					context.Context,
					*testservice.Input,
				) (*testservice.Output, error) {
					return &testservice.Output{
						Id:   "<id>",
						Data: "<data>",
					}, nil
				}

				handler.ServeHTTP(response, request)

				Expect(response.Header().Get("Content-Type")).To(Equal("text/plain; proto=protean.test.Response"))
				Expect(response.Body.String()).To(Equal("xxx"))
				Expect(response.Code).To(Equal(http.StatusOK))
			})
		})
	})
})
