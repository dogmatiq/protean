package proteanpb

import (
	sync "sync"

	"github.com/dogmatiq/protean/internal/protomime"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type EnvelopeMarshaler struct {
	d       protoreflect.MessageDescriptor
	corrID  protoreflect.FieldDescriptor
	message protoreflect.FieldDescriptor
	error   protoreflect.FieldDescriptor
}

func (em *EnvelopeMarshaler) Marshal(
	m protomime.Marshaler,
	corrID uint32,
	message proto.Message,
) ([]byte, error) {
	env := dynamicpb.NewMessage(em.d)

	env.Set(
		em.corrID,
		protoreflect.ValueOfUint32(corrID),
	)

	env.Set(
		em.message,
		protoreflect.ValueOfMessage(message.ProtoReflect()),
	)

	return m.Marshal(env)
}

func (em *EnvelopeMarshaler) MarshalError(
	m protomime.Marshaler,
	corrID uint32,
	err *Error,
) ([]byte, error) {
	env := dynamicpb.NewMessage(em.d)

	env.Set(
		em.corrID,
		protoreflect.ValueOfUint32(corrID),
	)

	env.Set(
		em.error,
		protoreflect.ValueOfMessage(err.ProtoReflect()),
	)

	return m.Marshal(env)
}

func (em *EnvelopeMarshaler) Unmarshal(
	u protomime.Unmarshaler,
	data []byte,
) (
	corrID uint32,
	message proto.Message,
	rpcErr *Error,
	err error,
) {
	env := dynamicpb.NewMessage(em.d)

	if err := u.Unmarshal(data, env); err != nil {
		return 0, nil, nil, err
	}

	corrID = uint32(env.Get(em.corrID).Uint())

	if env.Has(em.message) {
		message = env.Get(em.message).Message().Interface()
	} else if env.Has(em.error) {
		rpcErr = env.Get(em.message).Message().Interface().(*Error)
	}

	return corrID, message, rpcErr, nil
}

func NewEnvelopeMarshaler(t protoreflect.MessageType) *EnvelopeMarshaler {
	d := envelopeDescriptorFor(t.Descriptor())

	return &EnvelopeMarshaler{
		d,
		d.Fields().ByName("correlation_id"),
		d.Fields().ByName("message"),
		d.Fields().ByName("error"),
	}
}

var envelopeDescriptors sync.Map

func envelopeDescriptorFor(md protoreflect.MessageDescriptor) protoreflect.MessageDescriptor {
	env, ok := envelopeDescriptors.Load(md.FullName())
	if !ok {
		env, _ = envelopeDescriptors.LoadOrStore(
			md.FullName(),
			generateEnvelopeDescriptorFor(md),
		)
	}

	return env.(protoreflect.MessageDescriptor)
}

func generateEnvelopeDescriptorFor(md protoreflect.MessageDescriptor) protoreflect.MessageDescriptor {
	correlationIDType := descriptorpb.FieldDescriptorProto_TYPE_UINT32

	rpcErr := (&Error{}).ProtoReflect().Descriptor()

	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("_.proto"),
		Package: proto.String("protean.v1.envelope." + string(md.FullName().Parent())),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String(string(md.Name())),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("correlation_id"),
						Number:   proto.Int32(1),
						Type:     &correlationIDType,
						JsonName: proto.String("correlation_id"),
					},
					{
						Name:       proto.String("message"),
						Number:     proto.Int32(2),
						TypeName:   proto.String("." + string(md.FullName())),
						OneofIndex: proto.Int32(0),
					},
					{
						Name:       proto.String("error"),
						Number:     proto.Int32(3),
						TypeName:   proto.String("." + string(rpcErr.FullName())),
						OneofIndex: proto.Int32(0),
					},
				},
				OneofDecl: []*descriptorpb.OneofDescriptorProto{
					{Name: proto.String("payload")},
				},
			},
		},
		Dependency: []string{
			md.ParentFile().Path(),
			rpcErr.ParentFile().Path(),
		},
	}

	fdr, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		panic(err)
	}

	return fdr.Messages().Get(0)
}
