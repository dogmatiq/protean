package generator

import (
	"fmt"
	"strings"
)

// parameters encapsulates the options passed to the generator via the
// --harpy_opt flag.
//
// The are referred to as "options" on the protoc command line, but "parameters"
// within the Protocol Buffers plugin system.
type parameters struct {
	Module string
}

// parseParameters parses the parameters passed to the generator.
func parseParameters(params string) (parameters, error) {
	var p parameters

	for _, k := range strings.Split(params, ",") {
		if k == "" {
			continue
		}

		var v string
		if i := strings.Index(k, "="); i != -1 {
			v = k[i+1:]
			k = k[:i]
		}

		switch k {
		case "module":
			p.Module = v
		default:
			return parameters{}, fmt.Errorf("unrecognized option: %s", k)
		}
	}

	return p, nil
}
