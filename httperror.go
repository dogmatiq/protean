package protean

import (
	"fmt"
	"net/http"
)

// httpError writes an error response to w in plain text format.
func httpError(
	w http.ResponseWriter,
	status int,
	format string,
	args ...interface{},
) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	fmt.Fprintf(w, "%d %s\n\n", status, http.StatusText(status))
	fmt.Fprintf(w, format+"\n", args...)
}

// httpErrorUnsupportedMedia writes an HTTP 415 "Unsupported Media Type" error
// to w in plain text format.
func httpErrorUnsupportedMedia(
	w http.ResponseWriter,
	mediaType string,
	acceptable []string,
) {
	httpError(
		w,
		http.StatusUnsupportedMediaType,
		"The server does not support the '%s' media-type supplied by the client.",
		mediaType,
	)

	fmt.Fprintln(w, "\nThe supported types are, in order of preference:")
	for _, t := range acceptable {
		fmt.Fprintf(w, "- %s\n", t)
	}
}

// httpErrorNotAcceptable writes an HTTP 406 "Not Acceptable" error to w in
// plain text format.
func httpErrorNotAcceptable(w http.ResponseWriter, acceptable []string) {
	httpError(
		w,
		http.StatusNotAcceptable,
		"The client does not accept any of the media-types supported by the server.",
	)

	fmt.Fprintln(w, "\nThe supported types are, in order of preference:")
	for _, t := range acceptable {
		fmt.Fprintf(w, "- %s\n", t)
	}
}
