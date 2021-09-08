package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateService generates code for a single service definition.
func (g *Generator) generateServerInterface(
	out *jen.File,
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
) error {
	ident := fmt.Sprintf("Harpy%sServer", s.GetName())

	var methods []jen.Code

	for _, m := range s.GetMethod() {
		inputs, err := serverInputs(req, f, m)
		if err != nil {
			return err
		}

		outputs, err := serverOutputs(req, f, m)
		if err != nil {
			return err
		}

		methods = append(
			methods,
			jen.Id(m.GetName()).
				Params(inputs...).
				Params(outputs...),
		)
	}

	out.Commentf(
		"%s is an interface for types that implement the %s service.",
		ident,
		s.GetName(),
	)
	out.Type().Id(ident).Interface(methods...)

	return nil
}

// serverInputs returns the input parameters for a method of a generated server
// interface.
func serverInputs(
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) ([]jen.Code, error) {
	params := []jen.Code{
		jen.Qual("context", "Context"),
	}

	// Add the request parameter.
	{
		pkgPath, typeName, err := goType(
			req,
			m.GetInputType(),
		)
		if err != nil {
			return nil, err
		}

		if m.GetClientStreaming() {
			params = append(
				params,
				jen.Op("<-").Chan().Qual(pkgPath, typeName),
			)
		} else {
			params = append(
				params,
				jen.Qual(pkgPath, typeName),
			)
		}
	}

	// Add the response parameter for streaming responses.
	if m.GetServerStreaming() {
		pkgPath, typeName, err := goType(
			req,
			m.GetOutputType(),
		)
		if err != nil {
			return nil, err
		}

		params = append(
			params,
			jen.Chan().Op("<-").Qual(pkgPath, typeName),
		)
	}

	return params, nil
}

// serverOutpus returns the output parameters for a method of a generated server
// interface.
func serverOutputs(
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) ([]jen.Code, error) {
	var params []jen.Code

	if !m.GetServerStreaming() {
		pkgPath, typeName, err := goType(
			req,
			m.GetOutputType(),
		)
		if err != nil {
			return nil, err
		}

		params = append(
			params,
			jen.Qual(pkgPath, typeName),
		)
	}

	params = append(
		params,
		jen.Id("error"),
	)

	return params, nil
}
