package scope

import (
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// Request enscapsulates the generator scope for an entire generation request.
type Request struct {
	GenRequest *pluginpb.CodeGeneratorRequest
	GoModule   string
}

// EnterFile returns a scope for a file within this request.
func (s *Request) EnterFile(d *descriptorpb.FileDescriptorProto) *File {
	return &File{s, d}
}
