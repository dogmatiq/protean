package route

import (
	"errors"
	"fmt"
	"strings"
)

// Pattern is a parsed route pattern.
type Pattern []Segment

// Segment represents one segment of the route, that is, the part between
// slashes.
type Segment struct {
	Value         string
	IsPlaceholder bool
}

// ConflictsWith returns true if p conflicts with x, such that the route would
// be ambiguous.
func (p Pattern) ConflictsWith(x Pattern) bool {
	if len(p) != len(x) {
		return false
	}

	for i, pseg := range p {
		xseg := x[i]

		if pseg.IsPlaceholder || xseg.IsPlaceholder {
			continue
		}

		if pseg.Value != xseg.Value {
			return false
		}
	}

	return true
}

// ParsePattern parses a route pattern.
func ParsePattern(pattern string) (Pattern, error) {
	if len(pattern) == 0 {
		return nil, errors.New("pattern must not be empty")
	}

	if pattern[0] != '/' {
		return nil, errors.New("patterns must begin with a slash")
	}

	var result Pattern
	placeholders := map[string]struct{}{}

	for _, seg := range strings.Split(pattern[1:], "/") {
		if seg == "" {
			return nil, errors.New("path segment can not be empty")
		}

		if seg[0] == ':' {
			id := seg[1:]
			if id == "" {
				return nil, errors.New("placeholder identity can not be empty (nothing after colon)")
			}

			if _, ok := placeholders[id]; ok {
				return nil, fmt.Errorf("multiple uses of the :%s placeholder", id)
			}
			placeholders[id] = struct{}{}
			result = append(result, Segment{id, true})
		} else {
			result = append(result, Segment{seg, false})
		}
	}

	return result, nil
}
