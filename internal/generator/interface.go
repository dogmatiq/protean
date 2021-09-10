package generator

import (
	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateInterface generates a Go interface for a protocol buffers service.
func generateInterface(
	out *jen.File,
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
) error {
	typeName := interfaceName(s)

	var methods []jen.Code

	for _, m := range s.GetMethod() {
		inputs, err := methodInputs(req, f, m)
		if err != nil {
			return err
		}

		outputs, err := methodOutputs(req, f, m)
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
		"%s is an interface for the %s.%s service.",
		typeName,
		f.GetPackage(),
		s.GetName(),
	)
	out.Comment("")
	out.Comment("It is used for both clients and servers.")

	out.Type().Id(typeName).Interface(methods...)

	return nil
}

// methodInputs returns the input parameters for a method of a generated server
// interface.
func methodInputs(
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) ([]jen.Code, error) {
	params := []jen.Code{
		jen.Qual("context", "Context"),
	}

	// Add the input message parameter.
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
				jen.Op("<-").Chan().Op("*").Qual(pkgPath, typeName),
			)
		} else {
			params = append(
				params,
				jen.Op("*").Qual(pkgPath, typeName),
			)
		}
	}

	// Add the output message parameter for streaming responses.
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
			jen.Chan().Op("<-").Op("*").Qual(pkgPath, typeName),
		)
	}

	return params, nil
}

// methodOutputs returns the output parameters for a method of a generated
// server interface.
func methodOutputs(
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
			jen.Op("*").Qual(pkgPath, typeName),
		)
	}

	params = append(
		params,
		jen.Id("error"),
	)

	return params, nil
}

// interfaceName returns the name to use for the service interface for
// the given service.
func interfaceName(s *descriptorpb.ServiceDescriptorProto) string {
	return "Protean" + s.GetName()
}
