package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendService appends all generated code for an RPC service to the output.
func appendService(code *jen.File, s *scope.Service) error {
	if err := appendServiceInterface(code, s); err != nil {
		return err
	}

	appendServiceRegisterFunction(code, s)
	appendRuntimeServiceImpl(code, s)

	for _, m := range s.ServiceDesc.GetMethod() {
		appendMethod(code, s.EnterMethod(m))
	}

	return nil
}

// appendServiceRegisterFunction appends generated code for the user-facing
// function that registers a service with a registry.
func appendServiceRegisterFunction(code *jen.File, s *scope.Service) {
	code.Commentf(
		"%s registers a %s service with a Protean registry.",
		s.ServiceRegisterFunc(),
		s.ServiceInterface(),
	)
	code.Func().
		Id(s.ServiceRegisterFunc()).
		Params(
			jen.Id("r").Qual(runtimePackage, "Registry"),
			jen.Id("s").Id(s.ServiceInterface()),
		).
		Block(
			jen.Id("r").Dot("RegisterService").Call(
				jen.Op("&").Id(s.RuntimeServiceImpl()).Values(
					jen.Id("s"),
				),
			),
		)
}

// appendRuntimeServiceImpl appends a generated implementation of
// runtime.Service to the output.
func appendRuntimeServiceImpl(code *jen.File, s *scope.Service) {
	code.Commentf(
		"%s is a runtime.Service implementation for the %s.%s service.",
		s.RuntimeServiceImpl(),
		s.FileDesc.GetPackage(),
		s.ServiceDesc.GetName(),
	)
	code.Type().
		Id(s.RuntimeServiceImpl()).
		Struct(
			jen.Id("service").Id(s.ServiceInterface()),
		)

	recv := jen.Id("a").Op("*").Id(s.RuntimeServiceImpl())

	code.Line()
	code.Func().
		Params(recv).
		Id("Name").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(s.ServiceDesc.GetName())))

	code.Line()
	code.Func().
		Params(recv).
		Id("Package").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(s.FileDesc.GetPackage())))

	var cases []jen.Code
	for _, m := range s.ServiceDesc.GetMethod() {
		s := s.EnterMethod(m)

		cases = append(
			cases,
			jen.Case(
				jen.Lit(m.GetName()),
			).Block(
				jen.Return(
					jen.Op("&").Id(s.RuntimeMethodImpl()).Values(
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

	// TODO: avoid constructing new method instances each time
	code.Line()
	code.Func().
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
