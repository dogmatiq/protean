package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateFile generates a Go file for the given input .proto file.
func generateFile(
	s *scope.File,
	version string,
) (*pluginpb.CodeGeneratorResponse_File, error) {
	pkgPath, pkgName, err := s.GoPackage()
	if err != nil {
		return nil, err
	}

	code := jen.NewFilePathName(pkgPath, pkgName)
	code.HeaderComment("Code generated by protoc-gen-go-protean. DO NOT EDIT.")
	code.HeaderComment("versions:")
	code.HeaderComment(fmt.Sprintf("// 	protoc-gen-go-protean v%s", version))
	code.HeaderComment(fmt.Sprintf("// 	protoc                v%s", formatProtocVersion(s.GenRequest)))
	code.HeaderComment(fmt.Sprintf("// source: %s", s.FileDesc.GetName()))

	for _, d := range s.FileDesc.GetService() {
		if err := appendServiceExported(code, s.EnterService(d)); err != nil {
			return nil, err
		}
	}

	code.Comment(strings.Repeat("-", 117))
	code.Line()

	for _, d := range s.FileDesc.GetService() {
		if err := appendServiceUnexported(code, s.EnterService(d)); err != nil {
			return nil, err
		}
	}

	var w strings.Builder
	if err := code.Render(&w); err != nil {
		return nil, err
	}

	return &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(s.OutputFileName()),
		Content: proto.String(w.String()),
	}, nil
}

// formatProtocVersion formats the protoc version provided in the request for
// use in a file header.
func formatProtocVersion(req *pluginpb.CodeGeneratorRequest) string {
	v := req.GetCompilerVersion()

	s := fmt.Sprintf(
		"%d.%d.%d",
		v.GetMajor(),
		v.GetMinor(),
		v.GetPatch(),
	)

	if suffix := v.GetSuffix(); suffix != "" {
		s += "-" + suffix
	}

	return s
}
