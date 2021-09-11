package handler

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var nativeMarshaler = proto.MarshalOptions{}

var textMarshaler = prototext.MarshalOptions{
	Multiline: true,
	Indent:    "  ",
}

var jsonMarshaler = protojson.MarshalOptions{}
