package protean

import (
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
	nativeMarshaler marshaler = proto.MarshalOptions{}
	jsonMarshaler   marshaler = protojson.MarshalOptions{}
	textMarshaler   marshaler = prototext.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
)

var (
	nativeUnmarshaler unmarshaler = proto.UnmarshalOptions{}
	jsonUnmarshaler   unmarshaler = protojson.UnmarshalOptions{}
	textUnmarshaler   unmarshaler = prototext.UnmarshalOptions{}
)

// unmarshalerByMediaType returns the unmarshaler to use for unmarshaling the
// given media type.
//
// If the media type is not supported, ok is false.
func unmarshalerByMediaType(mediaType string) (_ unmarshaler, ok bool) {
	switch mediaType {
	case "application/vnd.google.protobuf", "application/x-protobuf":
		return nativeUnmarshaler, true
	case "application/json":
		return jsonUnmarshaler, true
	case "text/plain":
		return textUnmarshaler, true
	}

	return nil, false
}

// marshalerByAcceptHeaders returns the marshaler to use for marshaling
// responses based on the given accept headers.
//
// If none of the supported media types are accepted, ok is false.
func marshalerByAcceptHeaders(r *http.Request) (_ marshaler, mediaType string, ok bool) {
	t, _, err := contenttype.GetAcceptableMediaType(r, protoAcceptMediaTypes)
	if err != nil {
		return nil, "", false
	}

	mediaType = t.String()

	switch mediaType {
	case "application/vnd.google.protobuf", "application/x-protobuf":
		return nativeMarshaler, mediaType, true
	case "application/json":
		return jsonMarshaler, mediaType, true
	case "text/plain":
		return textMarshaler, mediaType, true
	}

	return nil, "", false
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
