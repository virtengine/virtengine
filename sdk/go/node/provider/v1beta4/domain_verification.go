package v1beta4

import (
	fmt "fmt"

	proto "github.com/cosmos/gogoproto/proto"
)

// VerificationMethod defines supported domain verification methods.
type VerificationMethod int32

const (
	VERIFICATION_METHOD_UNKNOWN         VerificationMethod = 0
	VERIFICATION_METHOD_DNS_TXT         VerificationMethod = 1
	VERIFICATION_METHOD_DNS_CNAME       VerificationMethod = 2
	VERIFICATION_METHOD_HTTP_WELL_KNOWN VerificationMethod = 3
)

var VerificationMethod_name = map[int32]string{
	0: "VERIFICATION_METHOD_UNKNOWN",
	1: "VERIFICATION_METHOD_DNS_TXT",
	2: "VERIFICATION_METHOD_DNS_CNAME",
	3: "VERIFICATION_METHOD_HTTP_WELL_KNOWN",
}

var VerificationMethod_value = map[string]int32{
	"VERIFICATION_METHOD_UNKNOWN":         0,
	"VERIFICATION_METHOD_DNS_TXT":         1,
	"VERIFICATION_METHOD_DNS_CNAME":       2,
	"VERIFICATION_METHOD_HTTP_WELL_KNOWN": 3,
}

func (x VerificationMethod) String() string {
	return proto.EnumName(VerificationMethod_name, int32(x))
}

func (VerificationMethod) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{0}
}

// MsgRequestDomainVerification requests a domain verification token for a provider.
type MsgRequestDomainVerification struct {
	Owner  string             `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Domain string             `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Method VerificationMethod `protobuf:"varint,3,opt,name=method,proto3,enum=virtengine.provider.v1beta4.VerificationMethod" json:"method,omitempty"`
}

func (m *MsgRequestDomainVerification) Reset()         { *m = MsgRequestDomainVerification{} }
func (m *MsgRequestDomainVerification) String() string { return proto.CompactTextString(m) }
func (*MsgRequestDomainVerification) ProtoMessage()    {}
func (*MsgRequestDomainVerification) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{1}
}

// MsgRequestDomainVerificationResponse returns the verification token and target.
type MsgRequestDomainVerificationResponse struct {
	Token              string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	ExpiresAt          int64  `protobuf:"varint,2,opt,name=expires_at,json=expiresAt,proto3" json:"expires_at,omitempty"`
	VerificationTarget string `protobuf:"bytes,3,opt,name=verification_target,json=verificationTarget,proto3" json:"verification_target,omitempty"`
}

func (m *MsgRequestDomainVerificationResponse) Reset()         { *m = MsgRequestDomainVerificationResponse{} }
func (m *MsgRequestDomainVerificationResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRequestDomainVerificationResponse) ProtoMessage()    {}
func (*MsgRequestDomainVerificationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{2}
}

// MsgConfirmDomainVerification confirms a verification proof for a provider domain.
type MsgConfirmDomainVerification struct {
	Owner string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Proof string `protobuf:"bytes,2,opt,name=proof,proto3" json:"proof,omitempty"`
}

func (m *MsgConfirmDomainVerification) Reset()         { *m = MsgConfirmDomainVerification{} }
func (m *MsgConfirmDomainVerification) String() string { return proto.CompactTextString(m) }
func (*MsgConfirmDomainVerification) ProtoMessage()    {}
func (*MsgConfirmDomainVerification) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{3}
}

// MsgConfirmDomainVerificationResponse returns verification status.
type MsgConfirmDomainVerificationResponse struct {
	Verified   bool  `protobuf:"varint,1,opt,name=verified,proto3" json:"verified,omitempty"`
	VerifiedAt int64 `protobuf:"varint,2,opt,name=verified_at,json=verifiedAt,proto3" json:"verified_at,omitempty"`
}

func (m *MsgConfirmDomainVerificationResponse) Reset()         { *m = MsgConfirmDomainVerificationResponse{} }
func (m *MsgConfirmDomainVerificationResponse) String() string { return proto.CompactTextString(m) }
func (*MsgConfirmDomainVerificationResponse) ProtoMessage()    {}
func (*MsgConfirmDomainVerificationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{4}
}

// MsgRevokeDomainVerification revokes a verified domain.
type MsgRevokeDomainVerification struct {
	Owner string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
}

func (m *MsgRevokeDomainVerification) Reset()         { *m = MsgRevokeDomainVerification{} }
func (m *MsgRevokeDomainVerification) String() string { return proto.CompactTextString(m) }
func (*MsgRevokeDomainVerification) ProtoMessage()    {}
func (*MsgRevokeDomainVerification) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{5}
}

// MsgRevokeDomainVerificationResponse returns revoke status.
type MsgRevokeDomainVerificationResponse struct{}

func (m *MsgRevokeDomainVerificationResponse) Reset()         { *m = MsgRevokeDomainVerificationResponse{} }
func (m *MsgRevokeDomainVerificationResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRevokeDomainVerificationResponse) ProtoMessage()    {}
func (*MsgRevokeDomainVerificationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{6}
}

// EventProviderDomainVerificationRequested emitted when verification is requested.
type EventProviderDomainVerificationRequested struct {
	Owner  string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Domain string `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Method string `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	Token  string `protobuf:"bytes,4,opt,name=token,proto3" json:"token,omitempty"`
}

func (m *EventProviderDomainVerificationRequested) Reset() {
	*m = EventProviderDomainVerificationRequested{}
}
func (m *EventProviderDomainVerificationRequested) String() string { return proto.CompactTextString(m) }
func (*EventProviderDomainVerificationRequested) ProtoMessage()    {}
func (*EventProviderDomainVerificationRequested) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{7}
}

// EventProviderDomainVerificationConfirmed emitted when verification is confirmed.
type EventProviderDomainVerificationConfirmed struct {
	Owner  string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Domain string `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Method string `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
}

func (m *EventProviderDomainVerificationConfirmed) Reset() {
	*m = EventProviderDomainVerificationConfirmed{}
}
func (m *EventProviderDomainVerificationConfirmed) String() string { return proto.CompactTextString(m) }
func (*EventProviderDomainVerificationConfirmed) ProtoMessage()    {}
func (*EventProviderDomainVerificationConfirmed) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{8}
}

// EventProviderDomainVerificationRevoked emitted when verification is revoked.
type EventProviderDomainVerificationRevoked struct {
	Owner  string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Domain string `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
}

func (m *EventProviderDomainVerificationRevoked) Reset() {
	*m = EventProviderDomainVerificationRevoked{}
}
func (m *EventProviderDomainVerificationRevoked) String() string { return proto.CompactTextString(m) }
func (*EventProviderDomainVerificationRevoked) ProtoMessage()    {}
func (*EventProviderDomainVerificationRevoked) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c5d87f1117a09a3, []int{9}
}

func init() {
	proto.RegisterEnum("virtengine.provider.v1beta4.VerificationMethod", VerificationMethod_name, VerificationMethod_value)
	proto.RegisterType((*MsgRequestDomainVerification)(nil), "virtengine.provider.v1beta4.MsgRequestDomainVerification")
	proto.RegisterType((*MsgRequestDomainVerificationResponse)(nil), "virtengine.provider.v1beta4.MsgRequestDomainVerificationResponse")
	proto.RegisterType((*MsgConfirmDomainVerification)(nil), "virtengine.provider.v1beta4.MsgConfirmDomainVerification")
	proto.RegisterType((*MsgConfirmDomainVerificationResponse)(nil), "virtengine.provider.v1beta4.MsgConfirmDomainVerificationResponse")
	proto.RegisterType((*MsgRevokeDomainVerification)(nil), "virtengine.provider.v1beta4.MsgRevokeDomainVerification")
	proto.RegisterType((*MsgRevokeDomainVerificationResponse)(nil), "virtengine.provider.v1beta4.MsgRevokeDomainVerificationResponse")
	proto.RegisterType((*EventProviderDomainVerificationRequested)(nil), "virtengine.provider.v1beta4.EventProviderDomainVerificationRequested")
	proto.RegisterType((*EventProviderDomainVerificationConfirmed)(nil), "virtengine.provider.v1beta4.EventProviderDomainVerificationConfirmed")
	proto.RegisterType((*EventProviderDomainVerificationRevoked)(nil), "virtengine.provider.v1beta4.EventProviderDomainVerificationRevoked")
}

// fileDescriptor_9c5d87f1117a09a3 is a placeholder descriptor for manual messages.
var fileDescriptor_9c5d87f1117a09a3 = []byte{0x0a, 0x1b, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f}

func (x VerificationMethod) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", x.String())), nil
}
