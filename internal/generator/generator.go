package generator

import (
	"fmt"

	"github.com/dogmatiq/protean/internal/generator/scope"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	rootPackage    = "github.com/dogmatiq/protean"
	runtimePackage = rootPackage + "/runtime"
)

// Generator produces a code generation response from a request.
type Generator struct {
	Version string
}

// Generate produces a code generation response for the given request.
func (g *Generator) Generate(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	params, err := parseParameters(req.GetParameter())
	if err != nil {
		return nil, err
	}

	s := &scope.Request{
		GenRequest: req,
		GoModule:   params.Module,
	}

	res := &pluginpb.CodeGeneratorResponse{}

	for _, n := range req.GetFileToGenerate() {
		for _, d := range req.GetProtoFile() {
			if d.GetName() != n {
				continue
			}

			if len(d.GetService()) == 0 {
				continue
			}

			fr, err := generateFile(s.EnterFile(d), g.Version)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", d.GetName(), err)
			}

			res.File = append(res.File, fr)
		}
	}

	return res, nil
}
