// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: github.com/dogmatiq/protean/proto/options.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var file_github_com_dogmatiq_protean_proto_options_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         58080,
		Name:          "protean.http_route",
		Tag:           "bytes,58080,opt,name=http_route",
		Filename:      "github.com/dogmatiq/protean/proto/options.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// optional string http_route = 58080;
	E_HttpRoute = &file_github_com_dogmatiq_protean_proto_options_proto_extTypes[0]
)

var File_github_com_dogmatiq_protean_proto_options_proto protoreflect.FileDescriptor

var file_github_com_dogmatiq_protean_proto_options_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6f, 0x67,
	0x6d, 0x61, 0x74, 0x69, 0x71, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x65, 0x61, 0x6e, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x07, 0x70, 0x72, 0x6f, 0x74, 0x65, 0x61, 0x6e, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3a, 0x3f, 0x0a, 0x0a,
	0x68, 0x74, 0x74, 0x70, 0x5f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe0, 0xc5, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x68, 0x74, 0x74, 0x70, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x42, 0x23, 0x5a,
	0x21, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6f, 0x67, 0x6d,
	0x61, 0x74, 0x69, 0x71, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x65, 0x61, 0x6e, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_github_com_dogmatiq_protean_proto_options_proto_goTypes = []interface{}{
	(*descriptorpb.MethodOptions)(nil), // 0: google.protobuf.MethodOptions
}
var file_github_com_dogmatiq_protean_proto_options_proto_depIdxs = []int32{
	0, // 0: protean.http_route:extendee -> google.protobuf.MethodOptions
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	0, // [0:1] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_dogmatiq_protean_proto_options_proto_init() }
func file_github_com_dogmatiq_protean_proto_options_proto_init() {
	if File_github_com_dogmatiq_protean_proto_options_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_dogmatiq_protean_proto_options_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_github_com_dogmatiq_protean_proto_options_proto_goTypes,
		DependencyIndexes: file_github_com_dogmatiq_protean_proto_options_proto_depIdxs,
		ExtensionInfos:    file_github_com_dogmatiq_protean_proto_options_proto_extTypes,
	}.Build()
	File_github_com_dogmatiq_protean_proto_options_proto = out.File
	file_github_com_dogmatiq_protean_proto_options_proto_rawDesc = nil
	file_github_com_dogmatiq_protean_proto_options_proto_goTypes = nil
	file_github_com_dogmatiq_protean_proto_options_proto_depIdxs = nil
}
