package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/dogmatiq/protean/internal/generator/descriptorutil"
	"github.com/dogmatiq/protean/internal/generator/route"
	"github.com/dogmatiq/protean/internal/generator/scope"
	"google.golang.org/protobuf/types/descriptorpb"
)

// appendMethod appends all generated code for an RPC method to the output.
func appendMethod(code *jen.File, s *scope.Method) error {
	if err := appendRuntimeMethodImpl(code, s); err != nil {
		return err
	}

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

	return nil
}

// appendRuntimeMethodImpl appends a generated implementation of runtime.Method
// to the output.
func appendRuntimeMethodImpl(code *jen.File, s *scope.Method) error {
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
		).
		Params(
			jen.Qual(runtimePackage, "Call"),
		).
		Block(jen.Return(
			jen.Id(s.RuntimeCallConstructor()).Call(
				jen.Id("ctx"),
				jen.Id("m").Dot("service"),
			),
		))

	patternString := s.MethodOptions().GetHttpRoute()
	if patternString == "" {
		return nil
	}

	pattern, err := route.ParsePattern(patternString)
	if err != nil {
		return fmt.Errorf(
			"%s.%s: http_route option '%s': %w",
			s.ServiceDesc.GetName(),
			s.MethodDesc.GetName(),
			patternString,
			err,
		)
	}

	_, protoInputType, err := descriptorutil.FindType(
		s.GenRequest.GetProtoFile(),
		s.MethodDesc.GetInputType(),
	)
	if err != nil {
		return fmt.Errorf(
			"%s.%s: http_route option '%s': %w",
			s.ServiceDesc.GetName(),
			s.MethodDesc.GetName(),
			patternString,
			err,
		)
	}

	inputPkg, inputType, _ := s.GoInputType()

	params := []jen.Code{
		jen.Id("params").Qual("net/url", "Values"),
	}

	statements := []jen.Code{
		jen.Id("in").Op(":=").Id("m").Assert(
			jen.Op("*").Qual(inputPkg, inputType),
		),
		jen.Line(),
	}

	for _, field := range protoInputType.GetField() {
		var assignStatements []jen.Code

		switch field.GetType() {
		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
			assignStatements = append(
				assignStatements,
				jen.Id("in").Dot(
					descriptorutil.GoFieldName(field.GetName()),
				).Op("=").Id("v"),
			)
			// default:
			// return fmt.Errorf(
			// 	"%s.%s: http_route option '%s': can not use :%s placeholder, population of %s fields is not supported",
			// 	s.ServiceDesc.GetName(),
			// 	s.MethodDesc.GetName(),
			// 	patternString,
			// 	seg.Value,
			// 	field.GetType(),
			// )
		}

		statements = append(
			statements,
			jen.For(
				jen.Id("_").Op(",").Id("v").Op(":=").Range().Id("params").Index(
					jen.Lit(field.GetName()),
				),
			).Block(assignStatements...),
			jen.Line(),
		)

	}

	for _, seg := range pattern {
		if !seg.IsPlaceholder {
			continue
		}

		param := "x_" + seg.Value
		params = append(
			params,
			jen.Id(param).String(),
		)

		field, err := descriptorutil.FindField(protoInputType, seg.Value)
		if err != nil {
			return fmt.Errorf(
				"%s.%s: http_route option '%s': %w",
				s.ServiceDesc.GetName(),
				s.MethodDesc.GetName(),
				patternString,
				err,
			)
		}

		switch field.GetType() {
		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
			statements = append(
				statements,
				jen.Id("in").Dot(
					descriptorutil.GoFieldName(seg.Value),
				).Op("=").Id(param),
			)
		default:
			return fmt.Errorf(
				"%s.%s: http_route option '%s': can not use :%s placeholder, population of %s fields is not supported",
				s.ServiceDesc.GetName(),
				s.MethodDesc.GetName(),
				patternString,
				seg.Value,
				field.GetType(),
			)
		}
	}

	statements = append(
		statements,
		jen.Line(),
		jen.Return(
			jen.Nil(),
		),
	)

	code.Line()
	code.Func().
		Params(recv).
		Id("newRouteUnmarshaler").
		Params(
			params...,
		).
		Params(
			jen.Qual(runtimePackage, "Unmarshaler"),
		).
		Block(
			jen.Return(
				jen.Func().
					Params(
						jen.Id("m").Qual(protoPackage, "Message"),
					).
					Params(
						jen.Error(),
					).Block(statements...),
			),
		)

	return nil
}
