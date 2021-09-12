package descriptorutil

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// FindType returns the descriptor for the given protocol buffers type.
func FindType(
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

// FindField returns the descriptor for the field with the given name.
func FindField(
	d *descriptorpb.DescriptorProto,
	name string,
) (*descriptorpb.FieldDescriptorProto, error) {
	for _, f := range d.GetField() {
		if f.GetName() == name {
			return f, nil
		}
	}

	return nil, fmt.Errorf(
		"%s does not contain a field named '%s'",
		d.GetName(),
		name,
	)
}
