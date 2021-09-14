package scope

import (
	"fmt"

	"google.golang.org/protobuf/types/descriptorpb"
)

// Service enscapsulates the generator scope for a single service within a file.
type Service struct {
	*File

	ServiceDesc *descriptorpb.ServiceDescriptorProto
}

// ServiceRegisterFunc returns the name of the function that registers an
// implementation of this service with a registry.
func (s *Service) ServiceRegisterFunc() string {
	return fmt.Sprintf(
		"ProteanRegister%sServer",
		s.ServiceDesc.GetName(),
	)
}

// ServiceInterface returns the name of the user-facing interface for this
// service.
func (s *Service) ServiceInterface() string {
	return fmt.Sprintf(
		"Protean%s",
		s.ServiceDesc.GetName(),
	)
}

// RuntimeServiceImpl returns the name of the runtime.Service implementation for
// this service.
func (s *Service) RuntimeServiceImpl() string {
	return fmt.Sprintf(
		"proteanService_%s",
		s.ServiceDesc.GetName(),
	)
}

// EnterMethod returns a scope for a method within this service.
func (s *Service) EnterMethod(d *descriptorpb.MethodDescriptorProto) *Method {
	return &Method{s, d}
}