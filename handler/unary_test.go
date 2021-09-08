package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/dogmatiq/harpy/handler"
	"github.com/dogmatiq/harpy/internal/testservice"
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

		testservice.RegisterHarpyTestServiceServer(handler, service)
	})

	AfterEach(func() {
		cancel()
	})

	Describe("func ServeHTTP()", func() {
		When("the request uses the HTTP GET method", func() {
			It("invokes the service with a zero-valued request", func() {
				request := httptest.NewRequest(
					http.MethodGet,
					"/harpy.test.TestService/Unary",
					nil,
				).WithContext(ctx)

				handler.ServeHTTP(response, request)

				Expect(response.Code).To(
					Equal(http.StatusOK),
					response.Body.String(),
				)
			})
		})
	})
})
