// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.31.1
// source: github.com/dogmatiq/protean/internal/testservice/service.proto

package testservice

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Input is the message used as inputs to all of the RPC methods in the test
// service.
type Input struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Data          string                 `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Input) Reset() {
	*x = Input{}
	mi := &file_github_com_dogmatiq_protean_internal_testservice_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Input) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Input) ProtoMessage() {}

func (x *Input) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dogmatiq_protean_internal_testservice_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Input.ProtoReflect.Descriptor instead.
func (*Input) Descriptor() ([]byte, []int) {
	return file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescGZIP(), []int{0}
}

func (x *Input) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Input) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

// Output is the message used as outputs from all of the RPC methods in the test
// service.
type Output struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Data          string                 `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Output) Reset() {
	*x = Output{}
	mi := &file_github_com_dogmatiq_protean_internal_testservice_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Output) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Output) ProtoMessage() {}

func (x *Output) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dogmatiq_protean_internal_testservice_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Output.ProtoReflect.Descriptor instead.
func (*Output) Descriptor() ([]byte, []int) {
	return file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescGZIP(), []int{1}
}

func (x *Output) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Output) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

var File_github_com_dogmatiq_protean_internal_testservice_service_proto protoreflect.FileDescriptor

const file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDesc = "" +
	"\n" +
	">github.com/dogmatiq/protean/internal/testservice/service.proto\x12\fprotean.test\"+\n" +
	"\x05Input\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04data\x18\x02 \x01(\tR\x04data\",\n" +
	"\x06Output\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04data\x18\x02 \x01(\tR\x04data2\x81\x02\n" +
	"\vTestService\x122\n" +
	"\x05Unary\x12\x13.protean.test.Input\x1a\x14.protean.test.Output\x12;\n" +
	"\fClientStream\x12\x13.protean.test.Input\x1a\x14.protean.test.Output(\x01\x12;\n" +
	"\fServerStream\x12\x13.protean.test.Input\x1a\x14.protean.test.Output0\x01\x12D\n" +
	"\x13BidirectionalStream\x12\x13.protean.test.Input\x1a\x14.protean.test.Output(\x010\x01B2Z0github.com/dogmatiq/protean/internal/testserviceb\x06proto3"

var (
	file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescOnce sync.Once
	file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescData []byte
)

func file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescGZIP() []byte {
	file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescOnce.Do(func() {
		file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDesc), len(file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDesc)))
	})
	return file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDescData
}

var file_github_com_dogmatiq_protean_internal_testservice_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_dogmatiq_protean_internal_testservice_service_proto_goTypes = []any{
	(*Input)(nil),  // 0: protean.test.Input
	(*Output)(nil), // 1: protean.test.Output
}
var file_github_com_dogmatiq_protean_internal_testservice_service_proto_depIdxs = []int32{
	0, // 0: protean.test.TestService.Unary:input_type -> protean.test.Input
	0, // 1: protean.test.TestService.ClientStream:input_type -> protean.test.Input
	0, // 2: protean.test.TestService.ServerStream:input_type -> protean.test.Input
	0, // 3: protean.test.TestService.BidirectionalStream:input_type -> protean.test.Input
	1, // 4: protean.test.TestService.Unary:output_type -> protean.test.Output
	1, // 5: protean.test.TestService.ClientStream:output_type -> protean.test.Output
	1, // 6: protean.test.TestService.ServerStream:output_type -> protean.test.Output
	1, // 7: protean.test.TestService.BidirectionalStream:output_type -> protean.test.Output
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_dogmatiq_protean_internal_testservice_service_proto_init() }
func file_github_com_dogmatiq_protean_internal_testservice_service_proto_init() {
	if File_github_com_dogmatiq_protean_internal_testservice_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDesc), len(file_github_com_dogmatiq_protean_internal_testservice_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_com_dogmatiq_protean_internal_testservice_service_proto_goTypes,
		DependencyIndexes: file_github_com_dogmatiq_protean_internal_testservice_service_proto_depIdxs,
		MessageInfos:      file_github_com_dogmatiq_protean_internal_testservice_service_proto_msgTypes,
	}.Build()
	File_github_com_dogmatiq_protean_internal_testservice_service_proto = out.File
	file_github_com_dogmatiq_protean_internal_testservice_service_proto_goTypes = nil
	file_github_com_dogmatiq_protean_internal_testservice_service_proto_depIdxs = nil
}
