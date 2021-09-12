package runtime

import "strings"

// NextPathSegment returns the next path segment of the path p.
func NextPathSegment(p string) (rest, seg string, ok bool) {
	if p == "" {
		return "", "", false
	}

	n := strings.IndexByte(p, '/')
	if n == -1 {
		return "", p, true
	}

	return p[n+1:], p[:n], true
}
