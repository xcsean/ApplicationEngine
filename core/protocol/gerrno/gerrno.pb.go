// Code generated by protoc-gen-go. DO NOT EDIT.
// source: gerrno.proto

package gerrno

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type E int32

const (
	E_OK               E = 0
	E_SYSINTERNALERROR E = 65001
)

var E_name = map[int32]string{
	0:     "OK",
	65001: "SYSINTERNALERROR",
}

var E_value = map[string]int32{
	"OK":               0,
	"SYSINTERNALERROR": 65001,
}

func (x E) String() string {
	return proto.EnumName(E_name, int32(x))
}

func (E) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_026ca002533abd27, []int{0}
}

func init() {
	proto.RegisterEnum("gerrno.E", E_name, E_value)
}

func init() { proto.RegisterFile("gerrno.proto", fileDescriptor_026ca002533abd27) }

var fileDescriptor_026ca002533abd27 = []byte{
	// 84 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x49, 0x4f, 0x2d, 0x2a,
	0xca, 0xcb, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x83, 0xf0, 0xb4, 0x94, 0xb9, 0x18,
	0x5d, 0x85, 0xd8, 0xb8, 0x98, 0xfc, 0xbd, 0x05, 0x18, 0x84, 0xc4, 0xb8, 0x04, 0x82, 0x23, 0x83,
	0x3d, 0xfd, 0x42, 0x5c, 0x83, 0xfc, 0x1c, 0x7d, 0x5c, 0x83, 0x82, 0xfc, 0x83, 0x04, 0x5e, 0xfe,
	0x66, 0x4e, 0x62, 0x03, 0xeb, 0x31, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xfe, 0x18, 0x6d, 0xf7,
	0x43, 0x00, 0x00, 0x00,
}