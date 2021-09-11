package scope

import (
	"path"
	"strings"

	"github.com/dogmatiq/protean/internal/generator/descriptorutil"
	"google.golang.org/protobuf/types/descriptorpb"
)

// File enscapsulates the generator scope for a single file within a generation
// request.
type File struct {
	*Request

	FileDesc *descriptorpb.FileDescriptorProto
}

// GoPackage returns the import path and unqualified package name for the file
// being generated.
func (s *File) GoPackage() (string, string, error) {
	return descriptorutil.GoPackage(s.FileDesc)
}

// OutputFileName returns the name of the file that is generated based on this
// file.
func (s *File) OutputFileName() string {
	n := strings.TrimPrefix(s.FileDesc.GetName(), s.GoModule)

	if ext := path.Ext(n); ext == ".proto" || ext == ".protodevel" {
		n = strings.TrimSuffix(n, ext)
	}

	return n + "_protean.pb.go"
}

// EnterService returns a scope for a service within this file.
func (s *File) EnterService(d *descriptorpb.ServiceDescriptorProto) *Service {
	return &Service{s, d}
}
