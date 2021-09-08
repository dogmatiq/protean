package httptransport

import (
	"io"
	"net/http"

	"github.com/elnormous/contenttype"
)

// notAcceptable writes a "not acceptable" HTTP error response.
func notAcceptable(w http.ResponseWriter, acceptable []contenttype.MediaType) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotAcceptable)

	_, _ = io.WriteString(w, "The client does not accept any of the media-types supported by the server:\n\n")

	for _, t := range acceptable {
		_, _ = io.WriteString(w, "- ")
		_, _ = io.WriteString(w, t.String())
		_, _ = io.WriteString(w, "\n")
	}
}

// negotiateMediaType returns the first accepted media-type that is supported by
// the client, or otherwise writes a "not acceptable" HTTP error response.
func negotiateMediaType(
	w http.ResponseWriter,
	r *http.Request,
	acceptable []contenttype.MediaType,
) (string, bool) {
	mediaType, _, err := contenttype.GetAcceptableMediaType(r, serviceMediaTypes)
	if err != nil {
		notAcceptable(w, acceptable)
		return "", false
	}

	mediaType.Parameters = nil
	return mediaType.String(), true
}
