package handler

import (
	"fmt"
	"net/http"

	"github.com/dogmatiq/protean/runtime"
	"github.com/elnormous/contenttype"
	"google.golang.org/protobuf/proto"
)

// unaryMediaTypes are the accepted media types when calling a unary RPC
// operation without using a websocket.
var unaryMediaTypes = []contenttype.MediaType{
	contenttype.NewMediaType("text/plain"),
	contenttype.NewMediaType("application/vnd.google.protobuf"),
	contenttype.NewMediaType("application/x-protobuf"),
	contenttype.NewMediaType("application/json"),
}

// unaryHandler is an implementation of http.Handler that handles "unary"
// (non-streaming) requests for a specific RPC method.
type unaryHandler struct {
	Service runtime.Service
	Method  runtime.Method
}

func (h *unaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mediaType, ok := negotiateMediaType(w, r, unaryMediaTypes)
	if !ok {
		return
	}

	var marshal func(proto.Message) ([]byte, error)

	switch mediaType {
	case "application/vnd.google.protobuf", "application/x-protobuf":
		marshal = nativeMarshaler.Marshal
	case "application/json":
		marshal = jsonMarshaler.Marshal
	case "text/plain":
		marshal = textMarshaler.Marshal
	default:
		panic(fmt.Sprintf("missing switch case: %s", mediaType))
	}

	var u runtime.Unmarshaler

	switch r.Method {
	case http.MethodGet:
		u = func(proto.Message) error {
			// TODO: parse query parameters
			return nil
		}

	case http.MethodPost:
		panic("not implemented")

	default:
		http.Error(
			w,
			"must use HTTP GET or POST method",
			http.StatusNotImplemented,
		)
	}

	call := h.Method.NewCall(r.Context())

	_, err := call.Send(u)
	call.Done()

	if err != nil {
		// TODO: handle err
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	res, ok, err := call.Recv()
	if err != nil {
		// TODO: handle err
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
	if !ok {
		// TODO: handle err
		http.Error(
			w,
			"no response received",
			http.StatusInternalServerError,
		)
		return
	}

	data, err := marshal(res)
	if err != nil {
		// TODO: handle err
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", mediaType)
	_, _ = w.Write(data) // TODO: gzip, etc
}
