package protean

import (
	"io"
	"net/http"
	"strconv"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/dogmatiq/protean/runtime"
	"google.golang.org/protobuf/proto"
)

// servePOST serves an RPC request made using the HTTP POST method.
func (h *handler) servePOST(
	w http.ResponseWriter,
	r *http.Request,
	method runtime.Method,
) {
	contentLength, ok := h.parseContentLength(w, r)
	if !ok {
		return
	}

	unmarshaler, inputMediaType, ok, err := unmarshalerByNegotiation(r)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the Content-Type header is missing or invalid",
			),
		)
		return
	}
	if !ok {
		httpError(
			w,
			http.StatusUnsupportedMediaType,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the server does not support the '%s' media-type supplied by the client",
				inputMediaType,
			).WithDetails(
				&proteanpb.SupportedMediaTypes{
					MediaTypes: protomime.MediaTypes,
				},
			),
		)
		return
	}

	marshaler, outputMediaType, ok, err := marshalerByNegotiation(r)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the Accept header is invalid",
			),
		)
		return
	}
	if !ok {
		httpError(
			w,
			http.StatusNotAcceptable,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the client does not accept any of the media-types supported by the server",
			).WithDetails(
				&proteanpb.SupportedMediaTypes{
					MediaTypes: protomime.MediaTypes,
				},
			),
		)
		return
	}

	// Setup a read limit. If we don't know the content-length we allow reading
	// up to the maximum input size.
	readLimit := contentLength
	if readLimit <= 0 {
		readLimit = h.maxInputSize
	}

	// Read the request body, up to the limit determined above, plus one byte.
	// The extra byte lets us detect if the actual content length is longer than
	// expected, or exceeds the maximum.
	data, err := io.ReadAll(
		io.LimitReader(
			r.Body,
			int64(readLimit)+1,
		),
	)
	if err != nil {
		httpError(
			w,
			http.StatusInternalServerError,
			outputMediaType,
			marshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the request body could not be read",
			),
		)
		return
	}

	if contentLength != 0 && len(data) != contentLength {
		httpError(
			w,
			http.StatusBadRequest,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the RPC input message length does not match the length specified by the Content-Length header",
			),
		)
		return
	}

	if len(data) > h.maxInputSize {
		httpError(
			w,
			http.StatusRequestEntityTooLarge,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the RPC input message length exceeds the maximum allowable size",
			),
		)
		return
	}

	call := method.NewCall(
		r.Context(),
		runtime.CallOptions{
			Interceptor: h.interceptor,
		},
	)
	defer call.Done()

	// Send never blocks on unary RPC methods.
	if _, err := call.Send(func(in proto.Message) error {
		return unmarshaler.Unmarshal(data, in)
	}); err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			outputMediaType,
			marshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the RPC input message could not be unmarshaled from the request body",
			),
		)
		return
	}

	out, _ := call.Recv()

	if err := call.Wait(); err != nil {
		if err, ok := err.(rpcerror.Error); ok {
			httpError(
				w,
				httpStatusFromErrorCode(err.Code()),
				outputMediaType,
				marshaler,
				err,
			)
		} else {
			httpError(
				w,
				http.StatusInternalServerError,
				outputMediaType,
				marshaler,
				rpcerror.New(
					rpcerror.Unknown,
					"the RPC method returned an unrecognized error",
				),
			)
		}

		return
	}

	data, err = marshaler.Marshal(out)
	if err != nil {
		httpError(
			w,
			http.StatusInternalServerError,
			outputMediaType,
			marshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the RPC output message could not be marshaled to the response body",
			),
		)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Add("Content-Type", protomime.FormatMediaType(outputMediaType, out))
	w.Header().Add("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// parseContentLength parses the Content-Length header, if present.
//
// It returns false if the Content-Length header is invalid or too large, in
// which case an error response has already been written to w.
//
// If no header is present it returns (0, true).
func (h *handler) parseContentLength(
	w http.ResponseWriter,
	r *http.Request,
) (int, bool) {
	header := r.Header.Get("Content-Length")
	if header == "" {
		return 0, true
	}

	contentLength, err := strconv.Atoi(header)
	if err != nil {
		httpError(
			w,
			http.StatusBadRequest,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the Content-Length header is invalid",
			),
		)
		return 0, false
	}

	if contentLength > h.maxInputSize {
		httpError(
			w,
			http.StatusRequestEntityTooLarge,
			protomime.TextMediaTypes[0],
			protomime.TextMarshaler,
			rpcerror.New(
				rpcerror.Unknown,
				"the length specified by the Content-Length header exceeds the maximum allowable size",
			),
		)
		return 0, false
	}

	return contentLength, true
}
