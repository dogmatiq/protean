package generator

// import (
// 	"fmt"
// 	"sort"

// 	"github.com/dave/jennifer/jen"
// 	"github.com/dogmatiq/protean/internal/generator/descriptorutil"
// 	"github.com/dogmatiq/protean/internal/generator/route"
// 	"github.com/dogmatiq/protean/internal/generator/scope"
// 	"google.golang.org/protobuf/types/descriptorpb"
// )

// // genMethodByRouteLogic generates the statements that comprise the
// // runtime.Service.MethodByRoute() method.
// func genMethodByRouteLogic(s *scope.Service) ([]jen.Code, error) {
// 	var routes []methodRoute

// 	for _, m := range s.ServiceDesc.GetMethod() {
// 		s := s.EnterMethod(m)

// 		pattern := s.MethodOptions().GetHttpRoute()
// 		if pattern == "" {
// 			continue
// 		}

// 		r, err := buildMethodRoute(s, routes, pattern)
// 		if err != nil {
// 			return nil, fmt.Errorf(
// 				"%s.%s: http_route option '%s': %w",
// 				s.ServiceDesc.GetName(),
// 				s.MethodDesc.GetName(),
// 				pattern,
// 				err,
// 			)
// 		}

// 		routes = append(routes, r)
// 	}

// 	var code []jen.Code

// 	code = append(
// 		code,
// 		jen.If(
// 			jen.Id("path").Op("==").Lit("").Op("||").
// 				Id("path").Index(jen.Lit(0)).Op("!=").LitRune('/'),
// 		).Block(
// 			jen.Return(
// 				jen.Nil(),
// 				jen.Nil(),
// 				jen.False(),
// 			),
// 		),
// 		jen.Line(),
// 		jen.Id("path").Op("=").Id("path").Index(jen.Lit(1).Op(":")),
// 		jen.Var().Id("seg").String(),
// 		jen.Var().Id("ok").Bool(),
// 		jen.Line(),
// 	)

// 	root := buildRouteTree(routes)

// 	code = append(
// 		code,
// 		genNextRouteSegment(root, "/")...,
// 	)

// 	code = append(
// 		code,
// 		walkRouteTree(root, "/", 0)...,
// 	)

// 	code = append(
// 		code,
// 		jen.Line(),
// 		jen.Return(
// 			jen.Nil(),
// 			jen.Nil(),
// 			jen.False(),
// 		),
// 	)

// 	return code, nil
// }

// func walkRouteTree(node *routeNode, path string, placeholderCount int) []jen.Code {
// 	var code []jen.Code

// 	if len(node.static) != 0 {
// 		code = append(
// 			code,
// 			genStaticRouteSwitch(node.static, path, placeholderCount)...,
// 		)
// 	}

// 	if node.placeholder != nil {
// 		code = append(
// 			code,
// 			genPlaceholderRoute(node.placeholder, path, placeholderCount)...,
// 		)
// 	}

// 	return code
// }

// func genStaticRouteSwitch(nodes []*routeNode, path string, placeholderCount int) []jen.Code {
// 	var cases []jen.Code
// 	for _, n := range nodes {
// 		casePath := path + n.match + "/"

// 		var statements []jen.Code

// 		statements = append(
// 			statements,
// 			genNextRouteSegment(n, casePath)...,
// 		)

// 		statements = append(
// 			statements,
// 			walkRouteTree(n, casePath, placeholderCount)...,
// 		)

// 		cases = append(
// 			cases,
// 			jen.Line(),
// 			jen.Comment("path: "+casePath),
// 			jen.Case(
// 				jen.Lit(n.match),
// 			).Block(statements...),
// 		)
// 	}

// 	var code []jen.Code

// 	code = append(
// 		code,
// 		jen.Line(),
// 		jen.Switch(
// 			jen.Id("seg"),
// 		).Block(cases...),
// 	)

// 	return code
// }

// func genPlaceholderRoute(node *routeNode, path string, placeholderCount int) []jen.Code {
// 	path += "*/"
// 	var code []jen.Code

// 	code = append(
// 		code,
// 		jen.Line(),
// 		jen.Id(
// 			fmt.Sprintf("placeholder%d", placeholderCount),
// 		).Op(":=").Id("seg"),
// 	)

// 	code = append(
// 		code,
// 		walkRouteTree(node, path, placeholderCount+1)...,
// 	)

// 	if node.leaf != nil {
// 		code = append(
// 			code,
// 			jen.Line(),
// 			jen.Comment("path: "+path),
// 		)

// 		code = append(
// 			code,
// 			genRouteLeaf(node, path)...,
// 		)
// 	}

// 	return code
// }

// func genNextRouteSegment(node *routeNode, path string) []jen.Code {
// 	var code []jen.Code

// 	code = append(
// 		code,
// 		jen.Id("path").Op(",").Id("seg").Op(",").Id("ok").Op("=").
// 			Qual(runtimePackage, "NextPathSegment").Call(jen.Id("path")),
// 	)

// 	if node.leaf != nil {
// 		code = append(
// 			code,
// 			jen.
// 				If(jen.Op("!").Id("ok")).
// 				Block(
// 					genRouteLeaf(node, path)...,
// 				),
// 		)
// 	} else {
// 		code = append(
// 			code,
// 			jen.
// 				If(jen.Op("!").Id("ok")).
// 				Block(
// 					jen.Return(
// 						jen.Nil(),
// 						jen.Nil(),
// 						jen.False(),
// 					),
// 				),
// 		)
// 	}

// 	return code
// }

// func genRouteLeaf(node *routeNode, path string) []jen.Code {
// 	inputPkg, inputType, _ := node.leaf.Scope.GoInputType()

// 	var code []jen.Code

// 	var statements []jen.Code

// 	statements = append(
// 		statements,
// 		jen.Id("in").Op(":=").Id("m").Assert(
// 			jen.Op("*").Qual(inputPkg, inputType),
// 		),
// 	)

// 	statements = append(
// 		statements,
// 		jen.Return(
// 			jen.Nil(),
// 		),
// 	)

// 	code = append(
// 		code,
// 		jen.Return(
// 			jen.Op("&").Id("s").Dot(node.leaf.Scope.RuntimeMethodField()),
// 			jen.Func().
// 				Params(
// 					jen.Id("m").Qual(protoPackage, "Message"),
// 				).
// 				Params(
// 					jen.Error(),
// 				).
// 				Block(statements...),
// 			jen.True(),
// 		),
// 	)

// 	return code
// }

// // methodRoute encapsulates information about an RPC method that uses the
// // Protean "http_route" option.
// type methodRoute struct {
// 	Scope   *scope.Method
// 	Pattern route.Pattern
// 	Fields  map[string]*descriptorpb.FieldDescriptorProto
// }

// // buildMethodRoute returns the methodRoute for the given method.
// func buildMethodRoute(
// 	s *scope.Method,
// 	routes []methodRoute,
// 	patternString string,
// ) (methodRoute, error) {
// 	_, inputType, err := descriptorutil.FindType(
// 		s.GenRequest.GetProtoFile(),
// 		s.MethodDesc.GetInputType(),
// 	)
// 	if err != nil {
// 		return methodRoute{}, err
// 	}

// 	pattern, err := route.ParsePattern(patternString)
// 	if err != nil {
// 		return methodRoute{}, err
// 	}

// 	if err := checkForRouteConflicts(routes, pattern); err != nil {
// 		return methodRoute{}, err
// 	}

// 	// fields, err := resolveRoutePlaceholders(inputType, pattern)
// 	// if err != nil {
// 	// 	return methodRoute{}, err
// 	// }

// 	return methodRoute{
// 		s,
// 		pattern,
// 		nil, // fields,
// 	}, nil
// }

// // checkForRouteConflicts returns an error if p conflicts with any of the given
// // routes.
// func checkForRouteConflicts(routes []methodRoute, p route.Pattern) error {
// 	for _, r := range routes {
// 		if p.ConflictsWith(r.Pattern) {
// 			return fmt.Errorf(
// 				"conflicts with the route for the '%s' method",
// 				r.Scope.MethodDesc.GetName(),
// 			)
// 		}
// 	}

// 	return nil
// }

// type routeNode struct {
// 	match       string
// 	static      []*routeNode
// 	placeholder *routeNode
// 	leaf        *methodRoute
// }

// func buildRouteTree(routes []methodRoute) *routeNode {
// 	root := &routeNode{}

// 	for _, r := range routes {
// 		r := r // capture loop variable
// 		node := root

// 		for i, seg := range r.Pattern {
// 			isLastSeg := i == len(r.Pattern)-1

// 			if seg.IsPlaceholder {
// 				if node.placeholder == nil {
// 					node.placeholder = &routeNode{}
// 				}

// 				if isLastSeg {
// 					node.placeholder.leaf = &r
// 				} else {
// 					node = node.placeholder
// 				}
// 			} else {
// 				var n *routeNode
// 				for _, x := range node.static {
// 					if x.match == seg.Value {
// 						n = x
// 						break
// 					}
// 				}

// 				if n == nil {
// 					n = &routeNode{
// 						match: seg.Value,
// 					}
// 					node.static = append(node.static, n)

// 					sort.Slice(node.static, func(i, j int) bool {
// 						a := node.static[i].match
// 						b := node.static[j].match

// 						if len(a) == len(b) {
// 							return a < b
// 						}

// 						return len(a) < len(b)
// 					})
// 				}

// 				if isLastSeg {
// 					n.leaf = &r
// 				} else {
// 					node = n
// 				}
// 			}
// 		}
// 	}

// 	return root
// }
