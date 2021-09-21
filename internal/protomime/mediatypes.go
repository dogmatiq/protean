package protomime

import (
	"mime"

	"google.golang.org/protobuf/proto"
)

// MediaTypes is the set of all media types that can be used for marshaling and
// unmarshaling Protocol Buffers messages, in order of preference.
var MediaTypes []string

// BinaryMediaTypes is the set of media types that refer to the standard binary
// Protocol Buffers encoding.
var BinaryMediaTypes = []string{
	"application/vnd.google.protobuf",
	"application/x-protobuf",
}

// JSONMediaTypes is the set of media types that refer to the JSON Protocol
// Buffers encoding.
var JSONMediaTypes = []string{
	"application/json",
}

// TextMediaTypes is the set of media types that refer to the text-based
// Protocol Buffers encoding.
var TextMediaTypes = []string{
	"text/plain",
}

// IsSupportedMediaType if the given media-type is supported.
func IsSupportedMediaType(mediaType string) bool {
	for _, x := range MediaTypes {
		if x == mediaType {
			return true
		}
	}

	return false
}

// IsBinary returns true if the given media-type refers to the standard binary
// Protocol Buffers encoding.
func IsBinary(mediaType string) bool {
	for _, x := range BinaryMediaTypes {
		if x == mediaType {
			return true
		}
	}

	return false
}

// IsJSON returns true if the given media-type refers to the JSON Protocol
// Buffers encoding.
func IsJSON(mediaType string) bool {
	for _, x := range JSONMediaTypes {
		if x == mediaType {
			return true
		}
	}

	return false
}

// IsText returns true if the given media-type refers to the text-based Protocol
// Buffers encoding.
func IsText(mediaType string) bool {
	for _, x := range TextMediaTypes {
		if x == mediaType {
			return true
		}
	}

	return false
}

// FormatMediaType formats a complete media type, including parameters, to use
// when marshaling m.
func FormatMediaType(mediaType string, m proto.Message) string {
	params := map[string]string{
		"x-proto": string(proto.MessageName(m)),
	}

	if IsText(mediaType) {
		params["charset"] = "utf-8"
	}

	return mime.FormatMediaType(mediaType, params)
}

func init() {
	MediaTypes = append(MediaTypes, BinaryMediaTypes...)
	MediaTypes = append(MediaTypes, JSONMediaTypes...)
	MediaTypes = append(MediaTypes, TextMediaTypes...)
}
