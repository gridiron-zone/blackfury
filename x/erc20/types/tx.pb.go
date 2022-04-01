// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: merlion/erc20/v1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("merlion/erc20/v1/tx.proto", fileDescriptor_bbb170594d423c80) }

var fileDescriptor_bbb170594d423c80 = []byte{
	// 195 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x8e, 0x31, 0x8e, 0xc2, 0x30,
	0x10, 0x00, 0x13, 0x9d, 0xee, 0x8a, 0x54, 0xa7, 0xd3, 0x35, 0x44, 0xc8, 0x3f, 0x88, 0x17, 0x87,
	0x1f, 0xd0, 0x50, 0xf1, 0x01, 0x3a, 0xdb, 0xb2, 0x8c, 0xa5, 0xc4, 0x1b, 0xc5, 0x4b, 0x14, 0x78,
	0x05, 0xcf, 0xa2, 0x4c, 0x49, 0x89, 0x92, 0x8f, 0x20, 0x12, 0xd3, 0xcd, 0xee, 0xac, 0x56, 0x93,
	0xad, 0x6a, 0xd3, 0x56, 0x0e, 0x3d, 0x98, 0x56, 0x97, 0x1b, 0xe8, 0x04, 0x50, 0xcf, 0x9b, 0x16,
	0x09, 0xff, 0x7e, 0xa3, 0xe2, 0xb3, 0xe2, 0x9d, 0xc8, 0xd7, 0x16, 0xd1, 0x56, 0x06, 0x64, 0xe3,
	0x40, 0x7a, 0x8f, 0x24, 0xc9, 0xa1, 0x0f, 0xcb, 0x7d, 0xfe, 0x6f, 0xd1, 0xe2, 0x8c, 0xf0, 0xa6,
	0xb8, 0x65, 0x1a, 0x43, 0x8d, 0x01, 0x94, 0x0c, 0x06, 0x3a, 0xa1, 0x0c, 0x49, 0x01, 0x1a, 0x9d,
	0x5f, 0x7c, 0xf9, 0x9d, 0x7d, 0x1d, 0x82, 0xdd, 0xed, 0xef, 0x23, 0x4b, 0x87, 0x91, 0xa5, 0xcf,
	0x91, 0xa5, 0xb7, 0x89, 0x25, 0xc3, 0xc4, 0x92, 0xc7, 0xc4, 0x92, 0x63, 0x61, 0x1d, 0x9d, 0xce,
	0x8a, 0x6b, 0xac, 0x21, 0x16, 0x15, 0x57, 0xf4, 0xe6, 0x33, 0x40, 0x1f, 0xdb, 0xe9, 0xd2, 0x98,
	0xa0, 0x7e, 0xe6, 0xb7, 0xdb, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0x4b, 0xcf, 0x82, 0xa5, 0xd9,
	0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "merlion.erc20.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "merlion/erc20/v1/tx.proto",
}