package protomime

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Marshaler is an interface for marshaling Protocol Buffers messages to a byte
// slice.
type Marshaler interface {
	Marshal(proto.Message) ([]byte, error)
}

// Unmarshaler is an interface for unmarshaling Protocol Buffers messages from a
// byte slice.
type Unmarshaler interface {
	Unmarshal([]byte, proto.Message) error
}

var (
	// BinaryMarshaler is a Marshaler that marshals messages to the binary
	// Protocol Buffers encoding.
	BinaryMarshaler Marshaler = proto.MarshalOptions{}

	// JSONMarshaler is a Marshaler that marshals messages to the JSON Protocol
	// Buffers encoding.
	JSONMarshaler Marshaler = protojson.MarshalOptions{}

	// TextMarshaler is a Marshaler that marshals messages to the text-based
	// Protocol Buffers encoding.
	TextMarshaler Marshaler = prototext.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
)

var (
	// BinaryUnmarshaler is an Unmarshaler that unmarshals messages from the
	// binary Protocol Buffers encoding.
	BinaryUnmarshaler Unmarshaler = proto.UnmarshalOptions{}

	// JSONUnmarshaler is an Unmarshaler that unmarshals messages from the JSON
	// Protocol Buffers encoding.
	JSONUnmarshaler Unmarshaler = protojson.UnmarshalOptions{}

	// TextUnmarshaler is an Unmarshaler that unmarshals messages from the
	// text-based Protocol Buffers encoding.
	TextUnmarshaler Unmarshaler = prototext.UnmarshalOptions{}
)

// MarshalerForMediaType returns the marshaler to use for the given media type.
func MarshalerForMediaType(mediaType string) (Marshaler, bool) {
	if IsBinary(mediaType) {
		return BinaryMarshaler, true
	}

	if IsJSON(mediaType) {
		return JSONMarshaler, true
	}

	if IsText(mediaType) {
		return TextMarshaler, true
	}

	return nil, false
}

// UnmarshalerForMediaType returns the unmarshaler to use for the given media
// type.
func UnmarshalerForMediaType(mediaType string) (Unmarshaler, bool) {
	if IsBinary(mediaType) {
		return BinaryUnmarshaler, true
	}

	if IsJSON(mediaType) {
		return JSONUnmarshaler, true
	}

	if IsText(mediaType) {
		return TextUnmarshaler, true
	}

	return nil, false
}
