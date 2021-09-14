package protean

import (
	"errors"
	"mime"
	"net/http"

	"github.com/elnormous/contenttype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// marshaler is an interface for marshaling Protocol Buffers messages to a byte
// slice.
type marshaler interface {
	Marshal(proto.Message) ([]byte, error)
}

// unmarshaler is an interface for unmarshaling Protocol Buffers messages from a
// byte slice.
type unmarshaler interface {
	Unmarshal([]byte, proto.Message) error
}

var (
	binaryMarshaler marshaler = proto.MarshalOptions{}
	jsonMarshaler   marshaler = protojson.MarshalOptions{}
	textMarshaler   marshaler = prototext.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
)

var (
	binaryUnmarshaler unmarshaler = proto.UnmarshalOptions{}
	jsonUnmarshaler   unmarshaler = protojson.UnmarshalOptions{}
	textUnmarshaler   unmarshaler = prototext.UnmarshalOptions{}
)

// unmarshalerByNegotiation returns the unmarshaler to use for unmarshaling the
// given request based on its Content-Type header.
//
// If the media types is not supported, ok is false.
func unmarshalerByNegotiation(r *http.Request) (_ unmarshaler, mediaType string, ok bool, err error) {
	mediaType = r.Header.Get("Content-Type")
	if mediaType == "" {
		return nil, "", false, errors.New("content type is empty")
	}

	// Parse media type both to validate and to strip the parameters.
	mediaType, _, err = mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, "", false, err
	}

	switch mediaType {
	case "application/vnd.google.protobuf", "application/x-protobuf":
		return binaryUnmarshaler, mediaType, true, nil
	case "application/json":
		return jsonUnmarshaler, mediaType, true, nil
	case "text/plain":
		return textUnmarshaler, mediaType, true, nil
	}

	return nil, mediaType, false, nil
}

// marshalerByNegotiation returns the marshaler to use for marshaling
// responses to the given request based on its Accept headers.
//
// If none of the supported media types are accepted, ok is false.
func marshalerByNegotiation(r *http.Request) (_ marshaler, mediaType string, ok bool, err error) {
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

	switch mediaType {
	case "application/vnd.google.protobuf", "application/x-protobuf":
		return binaryMarshaler, mediaType, true, nil
	case "application/json":
		return jsonMarshaler, mediaType, true, nil
	case "text/plain":
		return textMarshaler, mediaType, true, nil
	}

	return nil, "", false, nil
}

// protoMediaTypes is the set of media types that can be used for marshaling and
// unmarshaling protobuf messages.
var protoMediaTypes = []string{
	"application/vnd.google.protobuf",
	"application/x-protobuf",
	"application/json",
	"text/plain",
}

// protoAcceptMediaTypes is the set of media types that can be used for
// marshaling protobuf messages, in the format consumed by the
// github.com/elnormous/contenttype package.
var protoAcceptMediaTypes []contenttype.MediaType

func init() {
	for _, mediaType := range protoMediaTypes {
		protoAcceptMediaTypes = append(
			protoAcceptMediaTypes,
			contenttype.NewMediaType(mediaType),
		)
	}
}
