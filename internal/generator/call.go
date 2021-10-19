package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendRuntimeCallConstructor appends a function that constructs a
// runtime.Call implementation for an RPC method.
func appendRuntimeCallConstructor(
	code *jen.File,
	s *scope.Method,
	statements []jen.Code,
) {
	code.Commentf(
		"%s returns a new runtime.Call for the %s.%s.%s() method.",
		s.RuntimeCallConstructor(),
		s.FileDesc.GetPackage(),
		s.ServiceDesc.GetName(),
		s.MethodDesc.GetName(),
	)
	code.Func().
		Id(s.RuntimeCallConstructor()).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("service").Id(s.ServiceInterface()),
			jen.Id("interceptor").Qual(middlewarePackage, "ServerInterceptor"),
		).
		Params(
			jen.Qual(runtimePackage, "Call"),
		).
		Block(statements...)
}

// appendRuntimeCallImpl appends a generated implementation of runtime.Call to
// the output.
func appendRuntimeCallImpl(
	code *jen.File,
	s *scope.Method,
	fields, sendMethod, doneMethod, recvMethod, waitMethod, runMethod []jen.Code,
) {
	code.Commentf(
		"%s is a runtime.Call implementation for the %s.%s.%s() method.",
		s.RuntimeMethodImpl(),
		s.FileDesc.GetPackage(),
		s.ServiceDesc.GetName(),
		s.MethodDesc.GetName(),
	)
	code.Type().
		Id(s.RuntimeCallImpl()).
		Struct(fields...)

	recv := jen.Id("c").Op("*").Id(s.RuntimeCallImpl())

	code.Line()
	code.Func().
		Params(recv).
		Id("Send").
		Params(
			jen.Id("unmarshal").Qual(runtimePackage, "Unmarshaler"),
		).
		Params(
			jen.Bool(),
			jen.Error(),
		).
		Block(sendMethod...)

	code.Line()
	code.Func().
		Params(recv).
		Id("Done").
		Params().
		Block(doneMethod...)

	code.Line()
	code.Func().
		Params(recv).
		Id("Recv").
		Params().
		Params(
			jen.Qual(protoPackage, "Message"),
			jen.Bool(),
		).
		Block(recvMethod...)

	code.Line()
	code.Func().
		Params(recv).
		Id("Wait").
		Params().
		Params(
			jen.Error(),
		).
		Block(waitMethod...)

	if len(runMethod) != 0 {
		code.Line()
		code.Func().
			Params(recv).
			Id("run").
			Params().
			Block(runMethod...)
	}
}
