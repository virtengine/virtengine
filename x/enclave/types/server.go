package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
)

// MsgServer defines the Msg gRPC service interface
type MsgServer interface {
	RegisterEnclaveIdentity(context.Context, *MsgRegisterEnclaveIdentity) (*MsgRegisterEnclaveIdentityResponse, error)
	RotateEnclaveIdentity(context.Context, *MsgRotateEnclaveIdentity) (*MsgRotateEnclaveIdentityResponse, error)
	ProposeMeasurement(context.Context, *MsgProposeMeasurement) (*MsgProposeMeasurementResponse, error)
	RevokeMeasurement(context.Context, *MsgRevokeMeasurement) (*MsgRevokeMeasurementResponse, error)
}

// MsgRegisterEnclaveIdentityResponse is the response for MsgRegisterEnclaveIdentity
type MsgRegisterEnclaveIdentityResponse struct {
	KeyFingerprint string `json:"key_fingerprint"`
	ExpiryHeight   int64  `json:"expiry_height"`
}

// MsgRotateEnclaveIdentityResponse is the response for MsgRotateEnclaveIdentity
type MsgRotateEnclaveIdentityResponse struct {
	NewKeyFingerprint  string `json:"new_key_fingerprint"`
	OverlapStartHeight int64  `json:"overlap_start_height"`
	OverlapEndHeight   int64  `json:"overlap_end_height"`
}

// MsgProposeMeasurementResponse is the response for MsgProposeMeasurement
type MsgProposeMeasurementResponse struct {
	MeasurementHash string `json:"measurement_hash"`
}

// MsgRevokeMeasurementResponse is the response for MsgRevokeMeasurement
type MsgRevokeMeasurementResponse struct{}

// QueryServer defines the Query gRPC service interface
type QueryServer interface {
	EnclaveIdentity(context.Context, *QueryEnclaveIdentityRequest) (*QueryEnclaveIdentityResponse, error)
	ActiveValidatorEnclaveKeys(context.Context, *QueryActiveValidatorEnclaveKeysRequest) (*QueryActiveValidatorEnclaveKeysResponse, error)
	CommitteeEnclaveKeys(context.Context, *QueryCommitteeEnclaveKeysRequest) (*QueryCommitteeEnclaveKeysResponse, error)
	MeasurementAllowlist(context.Context, *QueryMeasurementAllowlistRequest) (*QueryMeasurementAllowlistResponse, error)
	Measurement(context.Context, *QueryMeasurementRequest) (*QueryMeasurementResponse, error)
	KeyRotation(context.Context, *QueryKeyRotationRequest) (*QueryKeyRotationResponse, error)
	ValidKeySet(context.Context, *QueryValidKeySetRequest) (*QueryValidKeySetResponse, error)
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	AttestedResult(context.Context, *QueryAttestedResultRequest) (*QueryAttestedResultResponse, error)
}

// RegisterMsgServer registers the MsgServer
// This is a stub implementation until proper protobuf generation is set up.
func RegisterMsgServer(s grpc.Server, impl MsgServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = impl
}

// RegisterQueryServer registers the QueryServer
// This is a stub implementation until proper protobuf generation is set up.
func RegisterQueryServer(s grpc.Server, impl QueryServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = impl
}
