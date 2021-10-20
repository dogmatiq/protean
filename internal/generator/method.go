package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/scope"
)

// appendMethod appends all generated code for an RPC method to the output.
func appendMethod(code *jen.File, s *scope.Method) {
	appendRuntimeMethodImpl(code, s)

	if s.MethodDesc.GetClientStreaming() && s.MethodDesc.GetServerStreaming() {
		appendBidirectionalStreamingRuntimeCallConstructor(code, s)
		appendBidirectionalStreamingRuntimeCallImpl(code, s)
	} else if s.MethodDesc.GetClientStreaming() {
		appendClientStreamingRuntimeCallConstructor(code, s)
		appendClientStreamingRuntimeCallImpl(code, s)
	} else if s.MethodDesc.GetServerStreaming() {
		appendServerStreamingRuntimeCallConstructor(code, s)
		appendServerStreamingRuntimeCallImpl(code, s)
	} else {
		appendUnaryRuntimeCallConstructor(code, s)
		appendUnaryRuntimeCallImpl(code, s)
	}
}

// appendRuntimeMethodImpl appends a generated implementation of runtime.Method
// to the output.
func appendRuntimeMethodImpl(code *jen.File, s *scope.Method) {
	code.Commentf(
		"%s is a runtime.Method implementation for the %s.%s.%s() method.",
		s.RuntimeMethodImpl(),
		s.FileDesc.GetPackage(),
		s.ServiceDesc.GetName(),
		s.MethodDesc.GetName(),
	)
	code.Type().
		Id(s.RuntimeMethodImpl()).
		Struct(
			jen.Id("service").Id(s.ServiceInterface()),
		)

	recv := jen.Id("m").Op("*").Id(s.RuntimeMethodImpl())

	code.Line()
	code.Func().
		Params(recv).
		Id("Name").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(s.MethodDesc.GetName())))

	code.Line()
	code.Func().
		Params(recv).
		Id("InputIsStream").
		Params().
		Params(jen.Bool()).
		Block(jen.Return(jen.Lit(s.MethodDesc.GetClientStreaming())))

	code.Line()
	code.Func().
		Params(recv).
		Id("OutputIsStream").
		Params().
		Params(jen.Bool()).
		Block(jen.Return(jen.Lit(s.MethodDesc.GetServerStreaming())))

	code.Line()
	code.Func().
		Params(recv).
		Id("NewCall").
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("options").Qual(runtimePackage, "CallOptions"),
		).
		Params(
			jen.Qual(runtimePackage, "Call"),
		).
		Block(jen.Return(
			jen.Id(s.RuntimeCallConstructor()).Call(
				jen.Id("ctx"),
				jen.Id("m").Dot("service"),
				jen.Id("options"),
			),
		))
}
