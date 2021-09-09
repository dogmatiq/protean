package generator

import (
	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateUnaryCallImpl generates an implementation of runtime.Call for a
// unary protocol buffers method.
func generateUnaryCallImpl(
	out *jen.File,
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) {
	ifaceName := interfaceName(s)
	implName := callImplName(s, m)

	inputPkg, inputType, _ := goType(req, m.GetInputType())
	outputPkg, outputType, _ := goType(req, m.GetOutputType())

	out.Commentf("%s is an implementation of the harpy.Call", implName)
	out.Commentf("interface for the %s.%s() method.", s.GetName(), m.GetName())
	out.Type().Id(implName).Struct(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("service").Id(ifaceName),
		jen.Id("done").Chan().Struct(),
		jen.Id("res").Op("*").Qual(outputPkg, outputType),
		jen.Id("err").Id("error"),
	)

	funcName := newCallFuncName(s, m)
	out.Commentf("%s returns a new harpy.Call for the %s.%s() method.", funcName, ifaceName, m.GetName())
	out.Func().
		Id(funcName).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("service").Id(ifaceName),
		).
		Params(
			jen.Qual(harpyPackage, "Call"),
		).
		Block(
			jen.Return(
				jen.Op("&").Id(implName).Values(
					jen.Id("ctx"),
					jen.Id("service"),
					jen.Make(jen.Chan().Struct()),
					jen.Nil(),
					jen.Nil(),
				),
			),
		)

	recv := jen.Id("c").Op("*").Id(implName)

	out.Line()
	out.Func().
		Params(recv).
		Id("Recv").
		Params().
		Params(
			jen.Qual("google.golang.org/protobuf/proto", "Message"),
			jen.Bool(),
			jen.Error(),
		).
		Block(
			jen.Select().Block(
				jen.Case(
					jen.Op("<-").Id("c").Dot("ctx").Dot("Done").Call(),
				).Block(
					jen.Return(
						jen.Nil(),
						jen.False(),
						jen.Id("c").Dot("ctx").Dot("Err").Call(),
					),
				),
				jen.Case(
					jen.Op("<-").Id("c").Dot("done"),
				).Block(
					jen.Return(
						jen.Id("c").Dot("res"),
						jen.Id("c").Dot("err").Op("==").Nil(),
						jen.Id("c").Dot("err"),
					),
				),
			),
		)

	out.Line()
	out.Func().
		Params(recv).
		Id("Send").
		Params(
			jen.Id("unmarshal").Qual(harpyPackage, "RawMessage"),
		).
		Params(
			jen.Error(),
		).
		Block(
			jen.Id("req").Op(":=").Op("&").Qual(inputPkg, inputType).Values(),
			jen.If(
				jen.Id("err").Op(":=").Id("unmarshal").Call(jen.Id("req")),
				jen.Id("err").Op("!=").Nil(),
			).Block(
				jen.Return(
					jen.Id("err"),
				),
			),
			jen.Line(),
			jen.Id("c").Dot("res").Op(",").Id("c").Dot("err").Op("=").
				Id("c").Dot("service").Dot(m.GetName()).
				Call(
					jen.Id("c").Dot("ctx"),
					jen.Id("req"),
				),
			jen.Close(jen.Id("c").Dot("done")),
			jen.Line(),
			jen.Return(
				jen.Nil(),
			),
		)

	out.Line()
	out.Func().
		Params(recv).
		Id("Done").
		Params().
		Block()
}
