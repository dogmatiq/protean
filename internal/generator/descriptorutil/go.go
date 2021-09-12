package descriptorutil

import (
	"fmt"
	"path"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// GoPackage parses the "go_package" option in the file and returns
// the import path and unqualified package name.
func GoPackage(f *descriptorpb.FileDescriptorProto) (string, string, error) {
	pkg := f.GetOptions().GetGoPackage()
	if pkg == "" {
		return "", "", fmt.Errorf("%s does not specify a 'go_package' option", f.GetName())
	}

	// If a semi-colon is present, the part after the semi-colon is the actual
	// package name. Used when the import path and package name differ.
	//
	// Use of this option is discouraged. See
	// https://developers.google.com/protocol-buffers/docs/reference/go-generated
	if i := strings.Index(pkg, ";"); i != -1 {
		return pkg[:i], pkg[i+1:], nil
	}

	return pkg, path.Base(pkg), nil
}

// GoType returns the package path and type name for the Go type that represents
// the given protocol buffers type.
func GoType(
	candidates []*descriptorpb.FileDescriptorProto,
	protoName string,
) (string, string, error) {
	f, t, err := FindType(candidates, protoName)
	if err != nil {
		return "", "", err
	}

	pkgPath, _, err := GoPackage(f)
	if err != nil {
		return "", "", err
	}

	return pkgPath, camelCase(t.GetName()), nil
}

// GoFieldName returns the Go struct field name that corresponds to the given
// Protocol Buffers field.
func GoFieldName(n string) string {
	return camelCase(n)
}
