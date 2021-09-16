package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendServiceInterface appends generated code for a user-facing Go interface
// for an RPC service.
func appendServiceInterface(code *jen.File, s *scope.Service) error {
	methods, err := genServiceInterfaceMethods(s)
	if err != nil {
		return err
	}

	code.Commentf(
		"%s is an interface for the %s.%s service.",
		s.ServiceInterface(),
		s.FileDesc.GetPackage(),
		s.ServiceDesc.GetName(),
	)
	code.Type().
		Id(s.ServiceInterface()).
		Interface(methods...)

	return nil
}

// genServiceInterfaceMethods returns the methods to be included in the service
// interface.
func genServiceInterfaceMethods(s *scope.Service) ([]jen.Code, error) {
	var methods []jen.Code

	for _, m := range s.ServiceDesc.GetMethod() {
		s := s.EnterMethod(m)

		inputs, err := genInterfaceMethodInputs(s)
		if err != nil {
			return nil, err
		}

		outputs, err := genInterfaceMethodOutputs(s)
		if err != nil {
			return nil, err
		}

		methods = append(
			methods,
			jen.Id(m.GetName()).
				Params(inputs...).
				Params(outputs...),
		)
	}

	return methods, nil
}

// genInterfaceMethodInputs returns the input parameters for a method in the
// service interface.
func genInterfaceMethodInputs(s *scope.Method) ([]jen.Code, error) {
	params := []jen.Code{
		jen.Id("ctx").Qual("context", "Context"),
	}

	// Add the input message parameter.
	{
		pkgPath, typeName, err := s.GoInputType()
		if err != nil {
			return nil, err
		}

		if s.MethodDesc.GetClientStreaming() {
			params = append(
				params,
				jen.Id("inputs").Op("<-").Chan().Op("*").Qual(pkgPath, typeName),
			)
		} else {
			params = append(
				params,
				jen.Id("in").Op("*").Qual(pkgPath, typeName),
			)
		}
	}

	// Add the output message channel for streaming responses.
	if s.MethodDesc.GetServerStreaming() {
		pkgPath, typeName, err := s.GoOutputType()
		if err != nil {
			return nil, err
		}

		params = append(
			params,
			jen.Id("outputs").Chan().Op("<-").Op("*").Qual(pkgPath, typeName),
		)
	}

	return params, nil
}

// genInterfaceMethodOutputs returns the input parameters for a method in the
// service interface.
func genInterfaceMethodOutputs(s *scope.Method) ([]jen.Code, error) {
	var params []jen.Code

	if !s.MethodDesc.GetServerStreaming() {
		pkgPath, typeName, err := s.GoOutputType()
		if err != nil {
			return nil, err
		}

		params = append(
			params,
			jen.Op("*").Qual(pkgPath, typeName),
		)
	}

	params = append(
		params,
		jen.Id("error"),
	)

	return params, nil
}
