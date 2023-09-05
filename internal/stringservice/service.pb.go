// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.2
// source: github.com/dogmatiq/protean/internal/stringservice/service.proto

package stringservice

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ToUpperRequest is the RPC input message for the StringService.ToUpper() RPC
// method.
type ToUpperRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OriginalString string `protobuf:"bytes,1,opt,name=original_string,json=originalString,proto3" json:"original_string,omitempty"`
}

func (x *ToUpperRequest) Reset() {
	*x = ToUpperRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToUpperRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToUpperRequest) ProtoMessage() {}

func (x *ToUpperRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToUpperRequest.ProtoReflect.Descriptor instead.
func (*ToUpperRequest) Descriptor() ([]byte, []int) {
	return file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescGZIP(), []int{0}
}

func (x *ToUpperRequest) GetOriginalString() string {
	if x != nil {
		return x.OriginalString
	}
	return ""
}

// ToUpperResponse is the RPC output message for the StringService.ToUpper() RPC
// method.
type ToUpperResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UppercaseString string `protobuf:"bytes,1,opt,name=uppercase_string,json=uppercaseString,proto3" json:"uppercase_string,omitempty"`
}

func (x *ToUpperResponse) Reset() {
	*x = ToUpperResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToUpperResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToUpperResponse) ProtoMessage() {}

func (x *ToUpperResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToUpperResponse.ProtoReflect.Descriptor instead.
func (*ToUpperResponse) Descriptor() ([]byte, []int) {
	return file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescGZIP(), []int{1}
}

func (x *ToUpperResponse) GetUppercaseString() string {
	if x != nil {
		return x.UppercaseString
	}
	return ""
}

var File_github_com_dogmatiq_protean_internal_stringservice_service_proto protoreflect.FileDescriptor

var file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDesc = []byte{
	0x0a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6f, 0x67,
	0x6d, 0x61, 0x74, 0x69, 0x71, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x65, 0x61, 0x6e, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0e, 0x70, 0x72, 0x6f, 0x74, 0x65, 0x61, 0x6e, 0x2e, 0x73, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x22, 0x39, 0x0a, 0x0e, 0x54, 0x6f, 0x55, 0x70, 0x70, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c,
	0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x6f,
	0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x22, 0x3c, 0x0a,
	0x0f, 0x54, 0x6f, 0x55, 0x70, 0x70, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x29, 0x0a, 0x10, 0x75, 0x70, 0x70, 0x65, 0x72, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x75, 0x70, 0x70, 0x65,
	0x72, 0x63, 0x61, 0x73, 0x65, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x32, 0x5b, 0x0a, 0x0d, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x07,
	0x54, 0x6f, 0x55, 0x70, 0x70, 0x65, 0x72, 0x12, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x65, 0x61,
	0x6e, 0x2e, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x54, 0x6f, 0x55, 0x70, 0x70, 0x65, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x65, 0x61,
	0x6e, 0x2e, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x54, 0x6f, 0x55, 0x70, 0x70, 0x65, 0x72,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6f, 0x67, 0x6d, 0x61, 0x74, 0x69, 0x71, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x65, 0x61, 0x6e, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescOnce sync.Once
	file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescData = file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDesc
)

func file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescGZIP() []byte {
	file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescOnce.Do(func() {
		file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescData)
	})
	return file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDescData
}

var file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_dogmatiq_protean_internal_stringservice_service_proto_goTypes = []interface{}{
	(*ToUpperRequest)(nil),  // 0: protean.string.ToUpperRequest
	(*ToUpperResponse)(nil), // 1: protean.string.ToUpperResponse
}
var file_github_com_dogmatiq_protean_internal_stringservice_service_proto_depIdxs = []int32{
	0, // 0: protean.string.StringService.ToUpper:input_type -> protean.string.ToUpperRequest
	1, // 1: protean.string.StringService.ToUpper:output_type -> protean.string.ToUpperResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_dogmatiq_protean_internal_stringservice_service_proto_init() }
func file_github_com_dogmatiq_protean_internal_stringservice_service_proto_init() {
	if File_github_com_dogmatiq_protean_internal_stringservice_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToUpperRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToUpperResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_com_dogmatiq_protean_internal_stringservice_service_proto_goTypes,
		DependencyIndexes: file_github_com_dogmatiq_protean_internal_stringservice_service_proto_depIdxs,
		MessageInfos:      file_github_com_dogmatiq_protean_internal_stringservice_service_proto_msgTypes,
	}.Build()
	File_github_com_dogmatiq_protean_internal_stringservice_service_proto = out.File
	file_github_com_dogmatiq_protean_internal_stringservice_service_proto_rawDesc = nil
	file_github_com_dogmatiq_protean_internal_stringservice_service_proto_goTypes = nil
	file_github_com_dogmatiq_protean_internal_stringservice_service_proto_depIdxs = nil
}
