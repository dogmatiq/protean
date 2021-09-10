package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateServiceRegisterFunction generates a function for registering a
// service with a registry.
func generateServiceRegisterFunction(
	out *jen.File,
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
) {
	ifaceName := interfaceName(s)
	implName := serviceImplName(s)
	funcName := fmt.Sprintf("ProteanRegister%sServer", s.GetName())

	out.Commentf("%s registers a %s service with a Protean registry.", funcName, ifaceName)
	out.Func().
		Id(funcName).
		Params(
			jen.Id("r").Qual(runtimePackage, "Registry"),
			jen.Id("s").Id(ifaceName),
		).
		Block(
			jen.Id("r").Dot("RegisterService").Call(
				jen.Op("&").Id(implName).Values(
					jen.Id("s"),
				),
			),
		)
}

// generateServiceImpl generates an implementation of runtime.Service for a
// protocol buffers service.
func generateServiceImpl(
	out *jen.File,
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
) {
	ifaceName := interfaceName(s)
	implName := serviceImplName(s)

	out.Commentf("%s is an implementation of the runtime.Service", implName)
	out.Commentf("interface for the %s service.", s.GetName())
	out.Type().Id(implName).Struct(
		jen.Id("service").Id(ifaceName),
	)

	recv := jen.Id("a").Op("*").Id(implName)

	out.Line()
	out.Func().
		Params(recv).
		Id("Name").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(s.GetName())))

	out.Line()
	out.Func().
		Params(recv).
		Id("Package").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(f.GetPackage())))

	var cases []jen.Code
	for _, m := range s.GetMethod() {
		cases = append(
			cases,
			jen.Case(
				jen.Lit(m.GetName()),
			).Block(
				jen.Return(
					jen.Op("&").Id(methodImplName(s, m)).Values(
						jen.Id("a").Dot("service"),
					),
					jen.True(),
				),
			),
		)
	}

	cases = append(
		cases,
		jen.Default().Block(
			jen.Return(
				jen.Nil(),
				jen.False(),
			),
		),
	)

	out.Line()
	out.Func().
		Params(recv).
		Id("MethodByName").
		Params(
			jen.Id("name").String(),
		).
		Params(
			jen.Qual(runtimePackage, "Method"),
			jen.Bool(),
		).
		Block(
			jen.Switch(jen.Id("name")).Block(cases...),
		)

	// out.Line()
	// out.Func().
	// 	Params(recv).
	// 	Id("MethodByURL").
	// 	Params(
	// 		jen.Id("u").Op("*").Qual("net/url", "URL"),
	// 	).
	// 	Params(
	// 		jen.Qual(runtimePackage, "Method"),
	// 		jen.Qual(runtimePackage, "Unmarshaler"),
	// 		jen.Bool(),
	// 	).
	// 	Block()
}

// serviceImplName returns the name to use for the type that implements
// runtime.Method for the given method.
func serviceImplName(
	s *descriptorpb.ServiceDescriptorProto,
) string {
	return fmt.Sprintf("protean_%s_Service", s.GetName())
}
