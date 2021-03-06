package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendClientStreamingRuntimeCallConstructor appends a function that
// constructs a runtime.Call implementation for a client streaming RPC method.
func appendClientStreamingRuntimeCallConstructor(code *jen.File, s *scope.Method) {
	inputPkg, inputType, _ := s.GoInputType()

	appendRuntimeCallConstructor(
		code,
		s,
		[]jen.Code{
			jen.Return(
				jen.Op("&").Id(s.RuntimeCallImpl()).
					Values(
						jen.Id("ctx"),
						jen.Id("service"),
						jen.Make(
							jen.Chan().Op("*").Qual(inputPkg, inputType),
							jen.Id("options").Dot("InputChannelCapacity"),
						),
						jen.Nil(), // err
					),
			),
		},
	)
}

// appendClientStreamingRuntimeCallImpl appends an implementation of
// runtime.Call for a client streaming RPC method.
func appendClientStreamingRuntimeCallImpl(code *jen.File, s *scope.Method) {
	inputPkg, inputType, _ := s.GoInputType()
	// outputPkg, outputType, _ := s.GoOutputType()

	appendRuntimeCallImpl(
		code,
		s,

		// struct fields
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("service").Id(s.ServiceInterface()),
			jen.Id("in").Chan().Op("*").Qual(inputPkg, inputType),
			jen.Id("err").Error(),
		},

		// send method
		[]jen.Code{
			jen.Id("in").Op(":=").Op("&").Qual(inputPkg, inputType).Values(),
			jen.If(
				jen.Id("err").Op(":=").Id("unmarshal").Call(jen.Id("in")),
				jen.Id("err").Op("!=").Nil(),
			).Block(
				jen.Return(
					jen.False(),
					jen.Id("err"),
				),
			),
			jen.Line(),
			jen.Select().Block(
				jen.Case(
					jen.Op("<-").Id("c").Dot("ctx").Dot("Done").Call(),
				).Block(
					jen.Return(
						jen.False(),
						jen.Nil(),
					),
				),
				jen.Case(
					jen.Id("c").Dot("in").Op("<-").Id("in"),
				).Block(
					jen.Return(
						jen.True(),
						jen.Nil(),
					),
				),
			),
		},

		// done method
		[]jen.Code{
			jen.Close(
				jen.Id("c").Dot("in"),
			),
		},

		// recv method
		[]jen.Code{
			jen.If(
				jen.Id("c").Dot("service").Op("==").Nil(),
			).Block(
				jen.Return(
					jen.Nil(),
					jen.False(),
				),
			),
			jen.Line(),
			jen.Id("out").Op(",").Id("err").Op(":=").
				Id("c").Dot("service").Dot(s.MethodDesc.GetName()).
				Call(
					jen.Id("c").Dot("ctx"),
					jen.Id("c").Dot("in"),
				),
			jen.Line(),
			jen.Id("c").Dot("service").Op("=").Nil(),
			jen.Id("c").Dot("err").Op("=").Id("err"),
			jen.Line(),
			jen.Return(
				jen.Id("out"),
				jen.Id("err").Op("==").Nil(),
			),
		},

		// wait method
		[]jen.Code{
			jen.If(
				jen.Id("c").Dot("service").Op("==").Nil(),
			).Block(
				jen.Return(
					jen.Id("c").Dot("err"),
				),
			),
			jen.Line(),
			jen.Panic(jen.Lit("Wait() called before Recv() returned false")),
		},

		// run method
		nil,
	)
}
