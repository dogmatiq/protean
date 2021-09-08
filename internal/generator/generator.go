package generator

import (
	"fmt"
	"path"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
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

	res := &pluginpb.CodeGeneratorResponse{}

	for _, n := range req.GetFileToGenerate() {
		for _, f := range req.GetProtoFile() {
			if f.GetName() == n {
				fres, ok, err := generateFile(
					req,
					f,
					params,
					g.Version,
				)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", f.GetName(), err)
				}

				if ok {
					res.File = append(res.File, fres)
				}
			}
		}
	}

	return res, nil
}

// goPackage parses the "go_package" option in the given file and returns
// the import path and unqualified package name.
func goPackage(f *descriptorpb.FileDescriptorProto) (string, string, error) {
	pkg := f.GetOptions().GetGoPackage()
	if pkg == "" {
		return "", "", fmt.Errorf("no 'go_package' option was specified in %s", f.GetName())
	}

	// If a semi-colon is present, the part after the semi-colon is the actual
	// package name. Used when the import path and package name differ.
	//
	// Use of this option is discouraged. See
	// https://developers.google.com/protocol-buffers/docs/reference/go-generated
	if i := strings.Index(pkg, ";"); i != -1 {
		return pkg[:i], pkg[i+1:], nil
	}

	return pkg, path.Base(pkg), nil
}

// goType returns the package path and type name for the Go type that represents
// the given protocol buffers type.
func goType(
	req *pluginpb.CodeGeneratorRequest,
	protoName string,
) (string, string, error) {
	f, t := findDescriptor(req, protoName)

	pkgPath, _, err := goPackage(f)
	if err != nil {
		return "", "", err
	}

	return pkgPath, camelCase(t.GetName()), nil
}

// findDescriptor returns the descriptor for the given protocol buffers type.
func findDescriptor(
	req *pluginpb.CodeGeneratorRequest,
	protoName string,
) (*descriptorpb.FileDescriptorProto, *descriptorpb.DescriptorProto) {
	i := strings.LastIndexByte(protoName, '.')
	pkg := protoName[1:i] // also trim leading .
	name := protoName[i+1:]

	for _, f := range req.GetProtoFile() {
		if f.GetPackage() != pkg {
			continue
		}

		for _, m := range f.GetMessageType() {
			if m.GetName() == name {
				return f, m
			}
		}
	}

	panic(fmt.Sprintf("no definition for type: %s", protoName))
}
