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

// Find returns the descriptor for the given protocol buffers type.
func Find(
	candidates []*descriptorpb.FileDescriptorProto,
	protoName string,
) (*descriptorpb.FileDescriptorProto, *descriptorpb.DescriptorProto, error) {
	i := strings.LastIndexByte(protoName, '.')
	pkg := protoName[1:i] // also trim leading .
	name := protoName[i+1:]

	for _, f := range candidates {
		if f.GetPackage() != pkg {
			continue
		}

		for _, m := range f.GetMessageType() {
			if m.GetName() == name {
				return f, m, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("none of the known files contain a definition for %s", protoName)
}

// GoType returns the package path and type name for the Go type that represents
// the given protocol buffers type.
func GoType(
	candidates []*descriptorpb.FileDescriptorProto,
	protoName string,
) (string, string, error) {
	f, t, err := Find(candidates, protoName)
	if err != nil {
		return "", "", err
	}

	pkgPath, _, err := GoPackage(f)
	if err != nil {
		return "", "", err
	}

	return pkgPath, camelCase(t.GetName()), nil
}
