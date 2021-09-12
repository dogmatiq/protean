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

	if err := appendRuntimeServiceImpl(code, s); err != nil {
		return err
	}

	for _, m := range s.ServiceDesc.GetMethod() {
		if err := appendMethod(code, s.EnterMethod(m)); err != nil {
			return err
		}
	}

	return nil
}

// appendServiceRegisterFunction appends generated code for the user-facing
// function that registers a service with a registry.
func appendServiceRegisterFunction(code *jen.File, s *scope.Service) {
	params := []jen.Code{}

	for _, m := range s.ServiceDesc.GetMethod() {
		s := s.EnterMethod(m)
		params = append(
			params,
			jen.Line().Id(s.RuntimeMethodImpl()).Values(
				jen.Id("s"),
			),
		)
	}

	params = append(params, jen.Line())

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
					params...,
				),
			),
		)

}

// appendRuntimeServiceImpl appends a generated implementation of
// runtime.Service to the output.
func appendRuntimeServiceImpl(code *jen.File, s *scope.Service) error {
	var fields, methodByNameCases []jen.Code

	for _, m := range s.ServiceDesc.GetMethod() {
		s := s.EnterMethod(m)

		fields = append(
			fields,
			jen.Id(
				s.RuntimeMethodField(),
			).Id(
				s.RuntimeMethodImpl(),
			),
		)

		methodByNameCases = append(
			methodByNameCases,
			jen.Case(
				jen.Lit(m.GetName()),
			).Block(
				jen.Return(
					jen.Op("&").Id("s").Dot(s.RuntimeMethodField()),
					jen.True(),
				),
			),
		)
	}

	code.Commentf(
		"%s is a runtime.Service implementation for the %s.%s service.",
		s.RuntimeServiceImpl(),
		s.FileDesc.GetPackage(),
		s.ServiceDesc.GetName(),
	)
	code.Type().
		Id(s.RuntimeServiceImpl()).
		Struct(fields...)

	recv := jen.Id("s").Op("*").Id(s.RuntimeServiceImpl())

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
			jen.Switch(jen.Id("name")).Block(methodByNameCases...),
			jen.Return(
				jen.Nil(),
				jen.False(),
			),
		)

	// statements, err := genMethodByRouteLogic(s)
	// if err != nil {
	// 	return err
	// }

	// code.Line()
	// code.Func().
	// 	Params(recv).
	// 	Id("MethodByRoute").
	// 	Params(
	// 		jen.Id("path").String(),
	// 		jen.Id("params").Qual("net/url", "Values"),
	// 	).
	// 	Params(
	// 		jen.Qual(runtimePackage, "Method"),
	// 		jen.Qual(runtimePackage, "Unmarshaler"),
	// 		jen.Bool(),
	// 	).
	// 	Block(statements...)

	return nil
}
