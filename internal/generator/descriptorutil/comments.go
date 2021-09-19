package descriptorutil

import (
	"reflect"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// These constants refer to the Protocol Buffers field tag within a specific
// descriptor type.
const (
	// descriptorpb.FileDescriptor field tags.
	tagFileDescriptorService = 6

	// descriptorpb.MethodDescriptor field tags.
	tagServiceDescriptorMethod = 2
)

// ServiceComments returns the comments on an RPC method definition.
func ServiceComments(
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
) []string {
	if f.GetSourceCodeInfo() == nil {
		return nil
	}

	for serviceIndex, service := range f.GetService() {
		if service != s {
			continue
		}

		return commentsAtPath(
			f,
			tagFileDescriptorService,
			int32(serviceIndex),
		)
	}

	return nil
}

// MethodComments returns the comments on an RPC method definition.
func MethodComments(
	f *descriptorpb.FileDescriptorProto,
	s *descriptorpb.ServiceDescriptorProto,
	m *descriptorpb.MethodDescriptorProto,
) []string {
	if f.GetSourceCodeInfo() == nil {
		return nil
	}

	for serviceIndex, service := range f.GetService() {
		if service != s {
			continue
		}

		for methodIndex, method := range s.GetMethod() {
			if method != m {
				continue
			}

			return commentsAtPath(
				f,
				tagFileDescriptorService,
				int32(serviceIndex),
				tagServiceDescriptorMethod,
				int32(methodIndex),
			)
		}
	}

	return nil
}

// commentsAtPath returns the lines of comment text at the location with the
// given path.
func commentsAtPath(f *descriptorpb.FileDescriptorProto, path ...int32) []string {
	for _, loc := range f.SourceCodeInfo.GetLocation() {
		if reflect.DeepEqual(path, loc.Path) {
			return splitComments(loc.GetLeadingComments())
		}
	}

	return nil
}

// splitComments splits comment text into separate lines and removes whitespace
// that does not actually appear in the comments as written by the user.
func splitComments(c string) []string {
	c = strings.TrimSuffix(c, "\n")

	if strings.TrimSpace(c) == "" {
		return nil
	}

	var lines []string
	for _, line := range strings.Split(c, "\n") {
		lines = append(
			lines,
			strings.TrimPrefix(line, " "),
		)
	}

	return lines
}
