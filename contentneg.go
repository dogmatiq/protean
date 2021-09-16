package protean

import (
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/elnormous/contenttype"
)

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
