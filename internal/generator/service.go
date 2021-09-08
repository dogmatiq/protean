package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const harpyPackage = "github.com/dogmatiq/harpy/codegenapi"

// generateServiceImpl generates an implementation of codegenapi.Service for a
// protocol buffers service.
func generateServiceImpl(
	out *jen.File,
	req *pluginpb.CodeGeneratorRequest,
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
) {
	out.ImportAlias(harpyPackage, "harpy")

	ifaceName := interfaceName(s)
	implName := serviceImplName(s)

	out.Commentf("%s is an implementation of the harpy.Service", implName)
	out.Commentf("interface for the %s service.", s.GetName())
	out.Type().Id(implName).Struct(
		jen.Id("service").Id(ifaceName),
	)

	funcName := fmt.Sprintf("RegisterHarpy%sServer", s.GetName())
	out.Commentf("%s registers a %s service with a Harpy server.", funcName, ifaceName)
	out.Func().
		Id(funcName).
		Params(
			jen.Id("server").Qual(harpyPackage, "Server"),
			jen.Id("service").Id(ifaceName),
		).
		Block(
			jen.Id("server").Dot("RegisterService").Call(
				jen.Op("&").Id(implName).Values(
					jen.Id("service"),
				),
			),
		)

	recv := jen.Id("a").Op("*").Id(implName)

	out.Line()
	out.Func().
		Params(recv).
		Id("Name").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(s.GetName())))

	out.Line()
	out.Func().
		Params(recv).
		Id("Package").
		Params().
		Params(jen.String()).
		Block(jen.Return(jen.Lit(f.GetPackage())))

	var cases []jen.Code
	for _, m := range s.GetMethod() {
		cases = append(
			cases,
			jen.Case(
				jen.Lit(m.GetName()),
			).Block(
				jen.Return(
					jen.Op("&").Id(methodImplName(s, m)).Values(
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

	out.Line()
	out.Func().
		Params(recv).
		Id("LookupMethod").
		Params(jen.Id("name").String()).
		Params(jen.Qual(harpyPackage, "Method"), jen.Bool()).
		Block(
			jen.Switch(jen.Id("name")).Block(cases...),
		)

	for _, m := range s.GetMethod() {
		generateMethodImpl(out, req, f, s, m)

		if m.GetClientStreaming() || m.GetServerStreaming() {
			generateStreamingCallImpl(
				out,
				req,
				f,
				s,
				m,
			)
		} else {
			// If this call doesn't use any streaming at all, use an optimised
			// Call implementation that avoids starting extra goroutines
			// necessary for streaming.
			generateUnaryCallImpl(
				out,
				req,
				f,
				s,
				m,
			)
		}
	}
}

// serviceImplName returns the name to use for the type that implements
// codegenapi.Method for the given method.
func serviceImplName(
	s *descriptorpb.ServiceDescriptorProto,
) string {
	return fmt.Sprintf("harpy_%s_Service", s.GetName())
}
