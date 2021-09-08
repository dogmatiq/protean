package generator

import (
	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateStreamingCallImpl generates an implementation of codegenapi.Call for
// a protocol buffers method that uses any kind of streaming.
func generateStreamingCallImpl(
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
		jen.Id("in").Chan().Op("*").Qual(inputPkg, inputType),
		jen.Id("out").Chan().Op("*").Qual(outputPkg, outputType),
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
			jen.Id("c").Op(":=").
				Op("&").Id(implName).
				Values(
					jen.Id("ctx"),
					jen.Id("service"),
					jen.Make(jen.Chan().Op("*").Qual(inputPkg, inputType), jen.Lit(1)),
					jen.Make(jen.Chan().Op("*").Qual(outputPkg, outputType), jen.Lit(1)),
					jen.Nil(),
				),
			jen.Go().Id("c").Dot("run").Call(),
			jen.Return(
				jen.Id("c"),
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
					jen.Id("res").Op(",").Id("ok").Op(":=").Op("<-").Id("c").Dot("out"),
				).Block(
					jen.If(jen.Id("ok")).Block(
						jen.Return(
							jen.Id("res"),
							jen.True(),
							jen.Nil(),
						),
					),
					jen.Return(
						jen.Nil(),
						jen.False(),
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
			jen.Select().Block(
				jen.Case(
					jen.Op("<-").Id("c").Dot("ctx").Dot("Done").Call(),
				).Block(
					jen.Return(
						jen.Id("c").Dot("ctx").Dot("Err").Call(),
					),
				),
				jen.Case(
					jen.Id("c").Dot("in").Op("<-").Id("req"),
				).Block(
					jen.Return(
						jen.Nil(),
					),
				),
			),
		)

	out.Line()
	out.Func().
		Params(recv).
		Id("Done").
		Params().
		Block(
			jen.Close(
				jen.Id("c").Dot("in"),
			),
		)

	out.Line()
	runMethod := out.Func().
		Params(recv).
		Id("run").
		Params()

	if m.GetClientStreaming() && m.GetServerStreaming() {
		// Generate the run() method for bidirectional streaming.
		runMethod.Block(
			jen.Defer().Close(
				jen.Id("c").Dot("out"),
			),
			jen.Id("c").Dot("err").Op("=").
				Id("c").Dot("service").Dot(m.GetName()).
				Call(
					jen.Id("c").Dot("ctx"),
					jen.Id("c").Dot("in"),
					jen.Id("c").Dot("out"),
				),
		)
	} else if m.GetClientStreaming() {
		// Generate the run() method for client streaming.
		runMethod.Block(
			jen.Defer().Close(
				jen.Id("c").Dot("out"),
			),
			jen.Line(),
			jen.Id("res").Op(",").Id("err").Op(":=").
				Id("c").Dot("service").Dot(m.GetName()).
				Call(
					jen.Id("c").Dot("ctx"),
					jen.Id("c").Dot("in"),
				),
			jen.If(jen.Id("err").Op("!=").Nil()).Block(
				jen.Id("c").Dot("err").Op("=").Id("err"),
			).Else().Block(
				jen.Id("c").Dot("out").Op("<-").Id("res").Comment("buffered, never blocks"),
			),
		)
	} else {
		// Generate the run() method for server streaming.
		runMethod.Block(
			jen.Defer().Close(
				jen.Id("c").Dot("out"),
			),
			jen.Line(),
			jen.Select().Block(
				jen.Case(
					jen.Op("<-").Id("c").Dot("ctx").Dot("Done").Call(),
				).Block(
					jen.Id("c").Dot("err").Op("=").
						Id("c").Dot("ctx").Dot("Err").Call(),
				),
				jen.Case(
					jen.Id("req").Op(",").Id("ok").Op(":=").Op("<-").Id("c").Dot("in"),
				).Block(
					jen.If(jen.Id("ok")).Block(
						jen.Id("c").Dot("err").Op("=").
							Id("c").Dot("service").Dot(m.GetName()).
							Call(
								jen.Id("c").Dot("ctx"),
								jen.Id("req"),
								jen.Id("c").Dot("out"),
							),
					).Else().Block(
						jen.Id("c").Dot("err").Op("=").
							Qual("errors", "New").
							Call(
								jen.Lit("Done() was called without sending a request"),
							),
					),
				),
			),
		)
	}
}
