// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: virtengine/base/v1beta1/attribute.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
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

// Attribute represents key value pair
type Attribute struct {
	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty" yaml:"key"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty" yaml:"value"`
}

func (m *Attribute) Reset()      { *m = Attribute{} }
func (*Attribute) ProtoMessage() {}
func (*Attribute) Descriptor() ([]byte, []int) {
	return fileDescriptor_77b8b6164cda3d93, []int{0}
}
func (m *Attribute) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Attribute) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Attribute.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Attribute) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Attribute.Merge(m, src)
}
func (m *Attribute) XXX_Size() int {
	return m.Size()
}
func (m *Attribute) XXX_DiscardUnknown() {
	xxx_messageInfo_Attribute.DiscardUnknown(m)
}

var xxx_messageInfo_Attribute proto.InternalMessageInfo

// SignedBy represents validation accounts that tenant expects signatures for provider attributes
// AllOf has precedence i.e. if there is at least one entry AnyOf is ignored regardless to how many
// entries there
// this behaviour to be discussed
type SignedBy struct {
	// all_of all keys in this list must have signed attributes
	AllOf []string `protobuf:"bytes,1,rep,name=all_of,json=allOf,proto3" json:"all_of" yaml:"allOf"`
	// any_of at least of of the keys from the list must have signed attributes
	AnyOf []string `protobuf:"bytes,2,rep,name=any_of,json=anyOf,proto3" json:"any_of" yaml:"anyOf"`
}

func (m *SignedBy) Reset()      { *m = SignedBy{} }
func (*SignedBy) ProtoMessage() {}
func (*SignedBy) Descriptor() ([]byte, []int) {
	return fileDescriptor_77b8b6164cda3d93, []int{1}
}
func (m *SignedBy) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SignedBy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SignedBy.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SignedBy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignedBy.Merge(m, src)
}
func (m *SignedBy) XXX_Size() int {
	return m.Size()
}
func (m *SignedBy) XXX_DiscardUnknown() {
	xxx_messageInfo_SignedBy.DiscardUnknown(m)
}

var xxx_messageInfo_SignedBy proto.InternalMessageInfo

// PlacementRequirements
type PlacementRequirements struct {
	// SignedBy list of keys that tenants expect to have signatures from
	SignedBy SignedBy `protobuf:"bytes,1,opt,name=signed_by,json=signedBy,proto3" json:"signed_by" yaml:"signed_by"`
	// Attribute list of attributes tenant expects from the provider
	Attributes []Attribute `protobuf:"bytes,2,rep,name=attributes,proto3" json:"attributes" yaml:"attributes"`
}

func (m *PlacementRequirements) Reset()      { *m = PlacementRequirements{} }
func (*PlacementRequirements) ProtoMessage() {}
func (*PlacementRequirements) Descriptor() ([]byte, []int) {
	return fileDescriptor_77b8b6164cda3d93, []int{2}
}
func (m *PlacementRequirements) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PlacementRequirements) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PlacementRequirements.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PlacementRequirements) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PlacementRequirements.Merge(m, src)
}
func (m *PlacementRequirements) XXX_Size() int {
	return m.Size()
}
func (m *PlacementRequirements) XXX_DiscardUnknown() {
	xxx_messageInfo_PlacementRequirements.DiscardUnknown(m)
}

var xxx_messageInfo_PlacementRequirements proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Attribute)(nil), "virtengine.base.v1beta1.Attribute")
	proto.RegisterType((*SignedBy)(nil), "virtengine.base.v1beta1.SignedBy")
	proto.RegisterType((*PlacementRequirements)(nil), "virtengine.base.v1beta1.PlacementRequirements")
}

func init() {
	proto.RegisterFile("virtengine/base/v1beta1/attribute.proto", fileDescriptor_77b8b6164cda3d93)
}

var fileDescriptor_77b8b6164cda3d93 = []byte{
	// 390 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0x31, 0x6f, 0xe2, 0x30,
	0x18, 0x86, 0x13, 0x38, 0x10, 0x31, 0xa7, 0x13, 0x17, 0xdd, 0xe9, 0x10, 0x83, 0xcd, 0x59, 0x2a,
	0x30, 0x25, 0x82, 0x6e, 0x6c, 0xcd, 0xd4, 0xad, 0x55, 0xba, 0xb5, 0x03, 0x72, 0xa8, 0x49, 0x23,
	0x42, 0x42, 0x13, 0x27, 0x52, 0xb6, 0x8e, 0x1d, 0xfb, 0x13, 0xfa, 0x73, 0x18, 0x19, 0x99, 0xa2,
	0x36, 0x6c, 0x6c, 0xcd, 0x2f, 0xa8, 0xe2, 0x84, 0x90, 0x85, 0xc9, 0x9f, 0x5e, 0x3f, 0xaf, 0xbf,
	0xef, 0xb5, 0x0d, 0x86, 0xa1, 0xe5, 0x31, 0xea, 0x98, 0x96, 0x43, 0x55, 0x83, 0xf8, 0x54, 0x0d,
	0xc7, 0x06, 0x65, 0x64, 0xac, 0x12, 0xc6, 0x3c, 0xcb, 0x08, 0x18, 0x55, 0xd6, 0x9e, 0xcb, 0x5c,
	0xf9, 0xdf, 0x09, 0x54, 0x32, 0x50, 0x29, 0xc0, 0xde, 0x1f, 0xd3, 0x35, 0x5d, 0xce, 0xa8, 0x59,
	0x95, 0xe3, 0xf8, 0x01, 0x48, 0x57, 0xc7, 0x13, 0xe4, 0x3e, 0xa8, 0x2f, 0x69, 0xd4, 0x15, 0xfb,
	0xe2, 0x48, 0xd2, 0x7e, 0xa5, 0x31, 0x02, 0x11, 0x59, 0xd9, 0x53, 0xbc, 0xa4, 0x11, 0xd6, 0xb3,
	0x2d, 0x79, 0x00, 0x1a, 0x21, 0xb1, 0x03, 0xda, 0xad, 0x71, 0xa6, 0x93, 0xc6, 0xe8, 0x67, 0xce,
	0x70, 0x19, 0xeb, 0xf9, 0xf6, 0xf4, 0xc7, 0xeb, 0x3b, 0x12, 0x70, 0x08, 0x5a, 0x77, 0x96, 0xe9,
	0xd0, 0x47, 0x2d, 0x92, 0xc7, 0xa0, 0x49, 0x6c, 0x7b, 0xe6, 0x2e, 0xba, 0x62, 0xbf, 0x3e, 0x92,
	0xb4, 0xde, 0x21, 0x46, 0x85, 0x72, 0x3a, 0x84, 0xd8, 0xf6, 0xcd, 0x02, 0xeb, 0x0d, 0xbe, 0x72,
	0x8b, 0x13, 0x65, 0x96, 0x5a, 0xc5, 0xc2, 0x95, 0x8a, 0xc5, 0x89, 0x72, 0x4b, 0xb6, 0x16, 0x7d,
	0xbf, 0x44, 0xf0, 0xf7, 0xd6, 0x26, 0x73, 0xba, 0xa2, 0x0e, 0xd3, 0xe9, 0x73, 0x60, 0x79, 0xbc,
	0xf4, 0x65, 0x0a, 0x24, 0x9f, 0x4f, 0x34, 0x33, 0xf2, 0x9c, 0xed, 0xc9, 0x7f, 0xe5, 0xcc, 0x8d,
	0x29, 0xc7, 0xd9, 0xb5, 0x8b, 0x4d, 0x8c, 0x84, 0x43, 0x8c, 0x4e, 0xde, 0x34, 0x46, 0x9d, 0xbc,
	0x7f, 0x29, 0x61, 0xbd, 0xe5, 0x1f, 0xc3, 0x2e, 0x01, 0x28, 0xdf, 0xc5, 0xe7, 0xd3, 0xb7, 0x27,
	0xf8, 0x6c, 0x9f, 0xf2, 0x01, 0xb4, 0x61, 0xd1, 0xa8, 0xe2, 0x4e, 0x63, 0xf4, 0xbb, 0x48, 0x5a,
	0x6a, 0x58, 0xaf, 0x00, 0x79, 0x66, 0xed, 0x7a, 0xf7, 0x09, 0x85, 0x97, 0x04, 0x0a, 0x9b, 0x04,
	0x8a, 0xdb, 0x04, 0x8a, 0x1f, 0x09, 0x14, 0xdf, 0xf6, 0x50, 0xd8, 0xee, 0xa1, 0xb0, 0xdb, 0x43,
	0xe1, 0x7e, 0x60, 0x5a, 0xec, 0x29, 0x30, 0x94, 0xb9, 0xbb, 0x52, 0x2b, 0xbf, 0xa9, 0x52, 0xb2,
	0x68, 0x4d, 0x7d, 0xa3, 0xc9, 0x7f, 0xc6, 0xe5, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xda, 0x17,
	0x8a, 0x31, 0x73, 0x02, 0x00, 0x00,
}

func (m *Attribute) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Attribute) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Attribute) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Value) > 0 {
		i -= len(m.Value)
		copy(dAtA[i:], m.Value)
		i = encodeVarintAttribute(dAtA, i, uint64(len(m.Value)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Key) > 0 {
		i -= len(m.Key)
		copy(dAtA[i:], m.Key)
		i = encodeVarintAttribute(dAtA, i, uint64(len(m.Key)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SignedBy) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SignedBy) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SignedBy) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AnyOf) > 0 {
		for iNdEx := len(m.AnyOf) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AnyOf[iNdEx])
			copy(dAtA[i:], m.AnyOf[iNdEx])
			i = encodeVarintAttribute(dAtA, i, uint64(len(m.AnyOf[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.AllOf) > 0 {
		for iNdEx := len(m.AllOf) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AllOf[iNdEx])
			copy(dAtA[i:], m.AllOf[iNdEx])
			i = encodeVarintAttribute(dAtA, i, uint64(len(m.AllOf[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *PlacementRequirements) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PlacementRequirements) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PlacementRequirements) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Attributes) > 0 {
		for iNdEx := len(m.Attributes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Attributes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintAttribute(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.SignedBy.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintAttribute(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintAttribute(dAtA []byte, offset int, v uint64) int {
	offset -= sovAttribute(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Attribute) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Key)
	if l > 0 {
		n += 1 + l + sovAttribute(uint64(l))
	}
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovAttribute(uint64(l))
	}
	return n
}

func (m *SignedBy) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.AllOf) > 0 {
		for _, s := range m.AllOf {
			l = len(s)
			n += 1 + l + sovAttribute(uint64(l))
		}
	}
	if len(m.AnyOf) > 0 {
		for _, s := range m.AnyOf {
			l = len(s)
			n += 1 + l + sovAttribute(uint64(l))
		}
	}
	return n
}

func (m *PlacementRequirements) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.SignedBy.Size()
	n += 1 + l + sovAttribute(uint64(l))
	if len(m.Attributes) > 0 {
		for _, e := range m.Attributes {
			l = e.Size()
			n += 1 + l + sovAttribute(uint64(l))
		}
	}
	return n
}

func sovAttribute(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAttribute(x uint64) (n int) {
	return sovAttribute(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Attribute) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAttribute
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Attribute: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Attribute: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAttribute
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAttribute
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Key = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAttribute
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAttribute
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAttribute(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAttribute
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SignedBy) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAttribute
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SignedBy: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignedBy: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllOf", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAttribute
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAttribute
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AllOf = append(m.AllOf, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AnyOf", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAttribute
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAttribute
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AnyOf = append(m.AnyOf, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAttribute(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAttribute
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PlacementRequirements) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAttribute
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PlacementRequirements: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PlacementRequirements: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SignedBy", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAttribute
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAttribute
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SignedBy.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Attributes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAttribute
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAttribute
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Attributes = append(m.Attributes, Attribute{})
			if err := m.Attributes[len(m.Attributes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAttribute(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAttribute
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipAttribute(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAttribute
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAttribute
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthAttribute
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAttribute
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAttribute
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAttribute        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAttribute          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAttribute = fmt.Errorf("proto: unexpected end of group")
)
