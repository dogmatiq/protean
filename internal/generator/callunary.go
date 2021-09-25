package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendUnaryRuntimeCallConstructor appends a function that constructs a
// runtime.Call implementation for a unary RPC method.
func appendUnaryRuntimeCallConstructor(code *jen.File, s *scope.Method) {
	inputPkg, inputType, _ := s.GoInputType()

	appendRuntimeCallConstructor(
		code,
		s,
		[]jen.Code{
			jen.Return(
				jen.Op("&").Id(s.RuntimeCallImpl()).Values(
					jen.Id("ctx"),
					jen.Id("service"),
					jen.Id("interceptor"),
					jen.Make(jen.Chan().Op("*").Qual(inputPkg, inputType), jen.Lit(1)),
				),
			),
		},
	)
}

// appendUnaryRuntimeCallImpl appends an implementation of runtime.Call for a
// unary RPC method.
func appendUnaryRuntimeCallImpl(code *jen.File, s *scope.Method) {
	inputPkg, inputType, _ := s.GoInputType()

	appendRuntimeCallImpl(
		code,
		s,

		// struct fields
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("service").Id(s.ServiceInterface()),
			jen.Id("interceptor").Qual(middlewarePackage, "ServerInterceptor"),
			jen.Id("in").Chan().Op("*").Qual(inputPkg, inputType),
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
			jen.Id("c").Dot("in").Op("<-").Id("in"),
			jen.Close(jen.Id("c").Dot("in")),
			jen.Line(),
			jen.Return(
				jen.False(),
				jen.Nil(),
			),
		},

		// done method
		[]jen.Code{},

		// recv method
		[]jen.Code{
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
					jen.Id("in").Op(",").Id("ok").Op(":=").Op("<-").Id("c").Dot("in"),
				).Block(
					jen.If(jen.Op("!").Id("ok")).Block(
						jen.Return(
							jen.Nil(),
							jen.False(),
							jen.Nil(),
						),
					),
					jen.Line(),
					jen.Id("out").Op(",").Id("err").Op(":=").
						Id("c").Dot("interceptor").Dot("InterceptUnaryRPC").Call(
						jen.Line().Id("c").Dot("ctx"),
						jen.Line().Qual(middlewarePackage, "UnaryServerInfo").Values(
							jen.Dict{
								jen.Id("Package"): jen.Lit(s.FileDesc.GetPackage()),
								jen.Id("Service"): jen.Lit(s.ServiceDesc.GetName()),
								jen.Id("Method"):  jen.Lit(s.MethodDesc.GetName()),
							},
						),
						jen.Line().Id("in"),
						jen.Line().Func().
							Params(
								jen.Id("ctx").Qual("context", "Context"),
							).
							Params(
								jen.Qual(protoPackage, "Message"),
								jen.Error(),
							).
							Block(
								jen.Return(
									jen.Id("c").Dot("service").Dot(s.MethodDesc.GetName()).
										Call(
											jen.Id("ctx"),
											jen.Id("in"),
										),
								),
							),
						jen.Line(),
					),
					jen.Line(),
					jen.Return(
						jen.Id("out"),
						jen.Id("err").Op("==").Nil(),
						jen.Id("err"),
					),
				),
			),
		},

		// run method
		nil,
	)
}
