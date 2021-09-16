package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendClientConstructor appends a function that constructs a client
// implementation for an RPC service.
func appendClientConstructor(
	code *jen.File,
	s *scope.Service,
) {
	code.Commentf(
		"%s returns a new client for the %s.%s service.",
		s.ClientConstructor(),
		s.FileDesc.GetPackage(),
		s.ServiceDesc.GetName(),
	)
	code.Func().
		Id(s.ClientConstructor()).
		Params(
			jen.Id("baseURL").Op("*").Qual("net/url", "URL"),
			jen.Id("options").Op("...").Qual(rootPackage, "ClientOption"),
		).
		Params(
			jen.Id(
				s.ServiceInterface(),
			),
		).
		Block(
			jen.Var().Id("opts").Qual(runtimePackage, "ClientOptions"),
			jen.For(
				jen.Id("_").Op(",").Id("opt").Op(":=").Range().Id("options"),
			).Block(
				jen.Id("opt").Call(
					jen.Op("&").Id("opts"),
				),
			),
			jen.Line(),
			jen.Return(
				jen.Op("&").Id(
					s.ClientImpl(),
				).Values(
					jen.Qual(runtimePackage, "NewClient").Call(
						jen.Id("baseURL"),
						jen.Id("opts"),
					),
				),
			),
		)
}

// appendClientImpl appends a generated implementation of an RPC client to the
// output.
func appendClientImpl(
	code *jen.File,
	s *scope.Service,
) error {
	code.Commentf(
		"%s is an implementation of the %s interface that is an RPC client.",
		s.ClientImpl(),
		s.ServiceInterface(),
	)
	code.Type().
		Id(s.ClientImpl()).
		Struct(
			jen.Id("client").Op("*").Qual(runtimePackage, "Client"),
		)

	for _, m := range s.ServiceDesc.GetMethod() {
		if err := appendClientMethod(code, s.EnterMethod(m)); err != nil {
			return err
		}
	}

	return nil
}

func appendClientMethod(
	code *jen.File,
	s *scope.Method,
) error {
	inputs, err := genInterfaceMethodInputs(s)
	if err != nil {
		return err
	}

	outputs, err := genInterfaceMethodOutputs(s)
	if err != nil {
		return err
	}

	recv := jen.Id("c").Op("*").Id(s.ClientImpl())

	var statements []jen.Code

	if s.MethodDesc.GetClientStreaming() || s.MethodDesc.GetServerStreaming() {
		newError := jen.Qual("errors", "New").Call(
			jen.Lit("This client does not support streaming RPC methods."),
		)

		if len(outputs) == 2 {
			statements = append(
				statements,
				jen.Return(
					jen.Nil(),
					newError,
				),
			)
		} else {
			statements = append(
				statements,
				jen.Return(newError),
			)
		}
	} else {
		outputPkg, outputType, err := s.GoOutputType()
		if err != nil {
			return err
		}

		statements = append(
			statements,
			jen.Id("out").Op(":=").Op("&").Qual(outputPkg, outputType).Values(),
			jen.Return(
				jen.Id("out"),
				jen.Id("c").Dot("client").Dot("CallUnary").Call(
					jen.Id("ctx"),
					jen.Lit(
						fmt.Sprintf(
							"/%s/%s/%s",
							s.FileDesc.GetPackage(),
							s.ServiceDesc.GetName(),
							s.MethodDesc.GetName(),
						),
					),
					jen.Id("in"),
					jen.Id("out"),
				),
			),
		)
	}

	code.Line()
	code.Func().
		Params(recv).
		Id(s.MethodDesc.GetName()).
		Params(inputs...).
		Params(outputs...).
		Block(statements...)

	return nil
}
