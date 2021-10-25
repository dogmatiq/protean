package protean

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
	"github.com/dogmatiq/protean/runtime"
	"github.com/elnormous/contenttype"
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

	marshaler, unmarshaler, outputMediaType, ok := negotiateMediaTypes(w, r)
	if !ok {
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

// negotiateMediaTypes negotiates the media types used to marshal and unmarshal
// RPC input/output messages.
//
// ok is false if the media-types can not be negotiated, in which case an error
// response has already been written to w.
func negotiateMediaTypes(
	w http.ResponseWriter,
	r *http.Request,
) (
	marshaler protomime.Marshaler,
	unmarshaler protomime.Unmarshaler,
	outputMediaType string,
	ok bool,
) {
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
		return nil, nil, "", false
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
		return nil, nil, "", false
	}

	marshaler, outputMediaType, ok, err = marshalerByNegotiation(r)
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
		return nil, nil, "", false
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

		return nil, nil, "", false
	}

	return marshaler, unmarshaler, outputMediaType, true
}

// unmarshalerByNegotiation returns the unmarshaler to use for unmarshaling the
// given request based on its Content-Type header.
//
// If the media types is not supported, ok is false.
func unmarshalerByNegotiation(r *http.Request) (u protomime.Unmarshaler, mediaType string, ok bool, err error) {
	mediaType = r.Header.Get("Content-Type")
	if mediaType == "" {
		return nil, "", false, errors.New("content type is empty")
	}

	// Parse media type both to validate and to strip the parameters.
	mediaType, _, err = mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, "", false, err
	}

	u, ok = protomime.UnmarshalerForMediaType(mediaType)

	return u, mediaType, ok, nil
}

// marshalerByNegotiation returns the marshaler to use for marshaling
// responses to the given request based on its Accept headers.
//
// If none of the supported media types are accepted, ok is false.
func marshalerByNegotiation(r *http.Request) (m protomime.Marshaler, mediaType string, ok bool, err error) {
	if len(r.Header.Values("Accept")) == 0 {
		// If no Accept header is provided, respond using the same content type
		// that the client supplied for the RPC input method.
		mediaType = r.Header.Get("Content-Type")
	} else {
		t, _, err := contenttype.GetAcceptableMediaType(r, protoAcceptMediaTypes)
		if err != nil && err != contenttype.ErrNoAcceptableTypeFound {
			return nil, "", false, err
		}

		mediaType = t.String()
	}

	m, ok = protomime.MarshalerForMediaType(mediaType)

	return m, mediaType, ok, nil
}

// protoAcceptMediaTypes is the set of media types that can be used for
// marshaling protobuf messages, in the format consumed by the
// github.com/elnormous/contenttype package.
var protoAcceptMediaTypes []contenttype.MediaType

// acceptPostHeader is the value to use for the Accept-Post header in HTTP
// responses.
var acceptPostHeader = strings.Join(
	protomime.MediaTypes,
	", ",
)

func init() {
	for _, mediaType := range protomime.MediaTypes {
		protoAcceptMediaTypes = append(
			protoAcceptMediaTypes,
			contenttype.NewMediaType(mediaType),
		)
	}
}

// httpStatusFromErrorCode returns the default HTTP status to send when an error
// with the given code occurs.
func httpStatusFromErrorCode(c rpcerror.Code) int {
	switch c {
	case rpcerror.InvalidInput:
		return http.StatusBadRequest
	case rpcerror.Unauthenticated:
		return http.StatusUnauthorized
	case rpcerror.PermissionDenied:
		return http.StatusForbidden
	case rpcerror.NotFound:
		return http.StatusNotFound
	case rpcerror.AlreadyExists:
		return http.StatusConflict
	case rpcerror.ResourceExhausted:
		return http.StatusTooManyRequests
	case rpcerror.FailedPrecondition:
		return http.StatusBadRequest
	case rpcerror.Aborted:
		return http.StatusConflict
	case rpcerror.Unavailable:
		return http.StatusServiceUnavailable
	case rpcerror.NotImplemented:
		return http.StatusNotImplemented
	}

	return http.StatusInternalServerError
}
