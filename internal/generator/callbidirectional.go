package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendBidirectionalStreamingRuntimeCallConstructor appends a function that
// constructs a runtime.Call implementation for a bidirectional streaming RPC method.
func appendBidirectionalStreamingRuntimeCallConstructor(code *jen.File, s *scope.Method) {
	inputPkg, inputType, _ := s.GoInputType()
	outputPkg, outputType, _ := s.GoOutputType()

	appendRuntimeCallConstructor(
		code,
		s,
		[]jen.Code{
			jen.Id("c").Op(":=").
				Op("&").Id(s.RuntimeCallImpl()).
				Values(
					jen.Id("ctx"),
					jen.Id("service"),
					jen.Make(jen.Chan().Op("*").Qual(inputPkg, inputType)),
					jen.Make(jen.Chan().Op("*").Qual(outputPkg, outputType)),
					jen.Make(jen.Chan().Error(), jen.Lit(1)),
				),
			jen.Go().Id("c").Dot("run").Call(),
			jen.Return(
				jen.Id("c"),
			),
		},
	)
}

// appendBidirectionalStreamingRuntimeCallImpl appends an implementation of
// runtime.Call for a bidirectional streaming RPC method.
func appendBidirectionalStreamingRuntimeCallImpl(code *jen.File, s *scope.Method) {
	inputPkg, inputType, _ := s.GoInputType()
	outputPkg, outputType, _ := s.GoOutputType()

	appendRuntimeCallImpl(
		code,
		s,

		// struct fields
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("service").Id(s.ServiceInterface()),
			jen.Id("in").Chan().Op("*").Qual(inputPkg, inputType),
			jen.Id("out").Chan().Op("*").Qual(outputPkg, outputType),
			jen.Id("err").Chan().Error(),
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
			jen.Id("out").Op(",").Id("ok").Op(":=").Op("<-").Id("c").Dot("out"),
			jen.Return(
				jen.Id("out"),
				jen.Id("ok"),
			),
		},

		// wait method
		[]jen.Code{
			jen.Return(
				jen.Op("<-").Id("c").Dot("err"),
			),
		},
		// run method
		[]jen.Code{
			jen.Id("c").Dot("err").Op("<-").
				Id("c").Dot("service").Dot(s.MethodDesc.GetName()).
				Call(
					jen.Id("c").Dot("ctx"),
					jen.Id("c").Dot("in"),
					jen.Id("c").Dot("out"),
				),
		},
	)
}
