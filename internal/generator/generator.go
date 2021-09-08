package generator

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// Generator produces a code generation response from a request.
type Generator struct {
}

// Generate produces a code generation response for the given request.
func (g *Generator) Generate(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	params, err := parseParameters(req.GetParameter())
	if err != nil {
		return nil, err
	}

	res := &pluginpb.CodeGeneratorResponse{}

	for _, n := range req.GetFileToGenerate() {
		for _, desc := range req.GetProtoFile() {
			if desc.GetName() == n {
				f, ok, err := g.generateFile(params, desc)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", desc.GetName(), err)
				}

				if ok {
					res.File = append(res.File, f)
				}
			}
		}
	}

	return res, nil
}

// generateFile generates a Harpy Go file for the given input protobuffers file.
func (g *Generator) generateFile(
	params parameters,
	desc *descriptorpb.FileDescriptorProto,
) (*pluginpb.CodeGeneratorResponse_File, bool, error) {
	services := desc.GetService()
	if len(services) == 0 {
		return nil, false, nil
	}

	pkg, err := packageName(desc)
	if err != nil {
		return nil, false, err
	}

	f := jen.NewFile(pkg)

	return &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(outputFileName(params, desc)),
		Content: proto.String(f.GoString()),
	}, true, nil
}

// outputFileName returns the name of the file to be generated from the given
// input file.
func outputFileName(params parameters, desc *descriptorpb.FileDescriptorProto) string {
	n := strings.TrimPrefix(desc.GetName(), params.Module)

	if ext := path.Ext(n); ext == ".proto" || ext == ".protodevel" {
		n = strings.TrimSuffix(n, ext)
	}

	return n + "_harpy.pb.go"
}

// parsePackageOption parses the "go_package" option in the given file and
// returns the (unqualified) package name.
func packageName(desc *descriptorpb.FileDescriptorProto) (string, error) {
	pkg := desc.GetOptions().GetGoPackage()
	if pkg == "" {
		return "", errors.New("no 'go_package' option was specified")
	}

	// If a semi-colon is present, the part after the semi-colon is the actual
	// package name. Used when the import path and package name differ.
	//
	// Use of this option is discouraged. See
	// https://developers.google.com/protocol-buffers/docs/reference/go-generated
	if i := strings.Index(pkg, ";"); i != -1 {
		return pkg[i+1:], nil
	}

	return path.Base(pkg), nil
}
