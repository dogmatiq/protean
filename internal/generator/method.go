package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateMethodImpl generates an implementation of runtime.Method for a
// protocol buffers method.
func generateMethodImpl(
	out *jen.File,
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) {
	ifaceName := interfaceName(s)
	implName := methodImplName(s, m)

	out.Commentf("%s is an implementation of the runtime.Method", implName)
	out.Commentf("interface for the %s.%s() method.", s.GetName(), m.GetName())
	out.Type().Id(implName).Struct(
		jen.Id("service").Id(ifaceName),
	)

	recv := jen.Id("m").Op("*").Id(implName)

	out.Line()
	out.Func().
		Params(recv).
		Id("Name").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(m.GetName())))

	out.Line()
	out.Func().
		Params(recv).
		Id("InputIsStream").
		Params().
		Params(jen.Bool()).
		Block(jen.Return(jen.Lit(m.GetClientStreaming())))

	out.Line()
	out.Func().
		Params(recv).
		Id("OutputIsStream").
		Params().
		Params(jen.Bool()).
		Block(jen.Return(jen.Lit(m.GetServerStreaming())))

	out.Line()
	out.Func().
		Params(recv).
		Id("NewCall").
		Params(
			jen.Id("ctx").Qual("context", "Context"),
		).
		Params(
			jen.Qual(runtimePackage, "Call"),
		).
		Block(jen.Return(
			jen.Id(newCallFuncName(s, m)).Call(
				jen.Id("ctx"),
				jen.Id("m").Dot("service"),
			),
		))
}

// methodImplName returns the name to use for the type that implements
// runtime.Method for the given method.
func methodImplName(
	s *descriptorpb.ServiceDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) string {
	return fmt.Sprintf(
		"protean_%s_%s_Method",
		s.GetName(),
		m.GetName(),
	)
}

// callImplName returns the name to use for the type that implements
// runtime.Call for the given method.
func callImplName(
	s *descriptorpb.ServiceDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) string {
	return fmt.Sprintf(
		"protean_%s_%s_Call",
		s.GetName(),
		m.GetName(),
	)
}

// newCallFuncName returns the name to use for the function that returns a new
// runtime.Call for the given method.
func newCallFuncName(
	s *descriptorpb.ServiceDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) string {
	return fmt.Sprintf(
		"newprotean_%s_%s_Call",
		s.GetName(),
		m.GetName(),
	)
}
