package types

import "fmt"

// BiometricHashProto is the stored representation of a biometric hash.
// Uses gogoproto tags for proper serialization with Cosmos SDK codec.
type BiometricHashProto struct {
	HashId          string   `protobuf:"bytes,1,opt,name=hash_id,json=hashId,proto3" json:"hash_id"`
	TemplateType    int32    `protobuf:"varint,2,opt,name=template_type,json=templateType,proto3" json:"template_type"`
	HashValue       []byte   `protobuf:"bytes,3,opt,name=hash_value,json=hashValue,proto3" json:"hash_value"`
	Salt            []byte   `protobuf:"bytes,4,opt,name=salt,proto3" json:"salt"`
	Version         uint32   `protobuf:"varint,5,opt,name=version,proto3" json:"version"`
	MatchThreshold  float64  `protobuf:"fixed64,6,opt,name=match_threshold,json=matchThreshold,proto3" json:"match_threshold"`
	LshHashes       [][]byte `protobuf:"bytes,7,rep,name=lsh_hashes,json=lshHashes,proto3" json:"lsh_hashes"`
	CreatedAt       int64    `protobuf:"varint,8,opt,name=created_at,json=createdAt,proto3" json:"created_at"`
	CreatedAtHeight int64    `protobuf:"varint,9,opt,name=created_at_height,json=createdAtHeight,proto3" json:"created_at_height"`
}

// BiometricAuditProto is the stored representation of a biometric audit entry.
// Uses gogoproto tags for proper serialization with Cosmos SDK codec.
type BiometricAuditProto struct {
	Operation    string `protobuf:"bytes,1,opt,name=operation,proto3" json:"operation"`
	HashId       string `protobuf:"bytes,2,opt,name=hash_id,json=hashId,proto3" json:"hash_id"`
	TemplateType int32  `protobuf:"varint,3,opt,name=template_type,json=templateType,proto3" json:"template_type"`
	Timestamp    int64  `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp"`
	BlockHeight  int64  `protobuf:"varint,5,opt,name=block_height,json=blockHeight,proto3" json:"block_height"`
	Address      string `protobuf:"bytes,6,opt,name=address,proto3" json:"address"`
	Success      bool   `protobuf:"varint,7,opt,name=success,proto3" json:"success"`
	ErrorMessage string `protobuf:"bytes,8,opt,name=error_message,json=errorMessage,proto3" json:"error_message"`
}

// Proto.Message interface stubs for BiometricHashProto.
// ProtoMessage is a marker method required by the proto.Message interface.
func (m *BiometricHashProto) ProtoMessage() { /* marker method - no implementation needed */ }
func (m *BiometricHashProto) Reset()        { *m = BiometricHashProto{} }
func (m *BiometricHashProto) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for BiometricAuditProto.
// ProtoMessage is a marker method required by the proto.Message interface.
func (m *BiometricAuditProto) ProtoMessage() { /* marker method - no implementation needed */ }
func (m *BiometricAuditProto) Reset()        { *m = BiometricAuditProto{} }
func (m *BiometricAuditProto) String() string { return fmt.Sprintf("%+v", *m) }
