package scope

import (
	"fmt"

	"github.com/dogmatiq/protean/internal/generator/descriptorutil"
	"github.com/dogmatiq/protean/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Method enscapsulates the generator scope for a single method of a service.
type Method struct {
	*Service

	MethodDesc *descriptorpb.MethodDescriptorProto
}

// GoInputType returns the package path and type name for the Go type that is
// used as the input message for this method.
func (s *Method) GoInputType() (string, string, error) {
	return descriptorutil.GoType(
		s.GenRequest.GetProtoFile(),
		s.MethodDesc.GetInputType(),
	)
}

// GoOutputType returns the package path and type name for the Go type that is
// used as the output message for this method.
func (s *Method) GoOutputType() (string, string, error) {
	return descriptorutil.GoType(
		s.GenRequest.GetProtoFile(),
		s.MethodDesc.GetOutputType(),
	)
}

// RuntimeMethodImpl returns the name of the runtime.Method implementation for this
// method.
func (s *Method) RuntimeMethodImpl() string {
	return fmt.Sprintf(
		"proteanMethod_%s_%s",
		s.ServiceDesc.GetName(),
		s.MethodDesc.GetName(),
	)
}

// RuntimeMethodField returns the name of the field within the runtime.Service
// implementation that contains the runtime.Method implementation for this
// method.
func (s *Method) RuntimeMethodField() string {
	return "method" + s.MethodDesc.GetName()
}

// RuntimeCallImpl returns the name of the runtime.Call implementation for this method.
func (s *Method) RuntimeCallImpl() string {
	return fmt.Sprintf(
		"proteanCall_%s_%s",
		s.ServiceDesc.GetName(),
		s.MethodDesc.GetName(),
	)
}

// RuntimeCallConstructor returns the name of the function that returns a new
// runtime.Call for this method.
func (s *Method) RuntimeCallConstructor() string {
	return fmt.Sprintf(
		"newProteanCall_%s_%s",
		s.ServiceDesc.GetName(),
		s.MethodDesc.GetName(),
	)
}

// MethodOptions returns the Protean method options for this method.
func (s *Method) MethodOptions() *options.MethodOptions {
	opts := s.MethodDesc.GetOptions()

	if proto.HasExtension(opts, options.E_Method) {
		return proto.GetExtension(opts, options.E_Method).(*options.MethodOptions)
	}

	return nil
}
