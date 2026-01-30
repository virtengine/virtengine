// Package types provides type definitions for the enclave module.
//
// This file re-exports types from the generated protobuf code in the SDK.
// All struct/message types, enums, and service interfaces are defined in the SDK
// proto-generated code and should be used via these aliases.
package types

import (
	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

// Msg and Query service interfaces
type (
	// MsgServer is the server interface for the Msg service.
	MsgServer = v1.MsgServer

	// QueryServer is the server interface for the Query service.
	QueryServer = v1.QueryServer

	// MsgClient is the client interface for the Msg service.
	MsgClient = v1.MsgClient

	// QueryClient is the client interface for the Query service.
	QueryClient = v1.QueryClient
)

// Core types from proto
type (
	// EnclaveIdentity represents a validator's enclave identity record.
	EnclaveIdentity = v1.EnclaveIdentity

	// MeasurementRecord represents an approved enclave measurement in the allowlist.
	MeasurementRecord = v1.MeasurementRecord

	// KeyRotationRecord represents a key rotation event.
	KeyRotationRecord = v1.KeyRotationRecord

	// AttestedScoringResult represents an enclave-attested scoring output.
	AttestedScoringResult = v1.AttestedScoringResult

	// ValidatorKeyInfo contains key information for a validator.
	ValidatorKeyInfo = v1.ValidatorKeyInfo

	// Params defines the parameters for the enclave module.
	Params = v1.Params

	// GenesisState defines the enclave module's genesis state.
	GenesisState = v1.GenesisState
)

// Message types
type (
	// MsgRegisterEnclaveIdentity registers a new enclave identity for a validator.
	MsgRegisterEnclaveIdentity = v1.MsgRegisterEnclaveIdentity

	// MsgRegisterEnclaveIdentityResponse is the response for MsgRegisterEnclaveIdentity.
	MsgRegisterEnclaveIdentityResponse = v1.MsgRegisterEnclaveIdentityResponse

	// MsgRotateEnclaveIdentity initiates a key rotation for a validator's enclave.
	MsgRotateEnclaveIdentity = v1.MsgRotateEnclaveIdentity

	// MsgRotateEnclaveIdentityResponse is the response for MsgRotateEnclaveIdentity.
	MsgRotateEnclaveIdentityResponse = v1.MsgRotateEnclaveIdentityResponse

	// MsgProposeMeasurement proposes a new enclave measurement for the allowlist.
	MsgProposeMeasurement = v1.MsgProposeMeasurement

	// MsgProposeMeasurementResponse is the response for MsgProposeMeasurement.
	MsgProposeMeasurementResponse = v1.MsgProposeMeasurementResponse

	// MsgRevokeMeasurement revokes an enclave measurement from the allowlist.
	MsgRevokeMeasurement = v1.MsgRevokeMeasurement

	// MsgRevokeMeasurementResponse is the response for MsgRevokeMeasurement.
	MsgRevokeMeasurementResponse = v1.MsgRevokeMeasurementResponse

	// MsgUpdateParams updates the module parameters.
	MsgUpdateParams = v1.MsgUpdateParams

	// MsgUpdateParamsResponse is the response for MsgUpdateParams.
	MsgUpdateParamsResponse = v1.MsgUpdateParamsResponse
)

// Query request/response types
type (
	QueryEnclaveIdentityRequest             = v1.QueryEnclaveIdentityRequest
	QueryEnclaveIdentityResponse            = v1.QueryEnclaveIdentityResponse
	QueryActiveValidatorEnclaveKeysRequest  = v1.QueryActiveValidatorEnclaveKeysRequest
	QueryActiveValidatorEnclaveKeysResponse = v1.QueryActiveValidatorEnclaveKeysResponse
	QueryCommitteeEnclaveKeysRequest        = v1.QueryCommitteeEnclaveKeysRequest
	QueryCommitteeEnclaveKeysResponse       = v1.QueryCommitteeEnclaveKeysResponse
	QueryMeasurementAllowlistRequest        = v1.QueryMeasurementAllowlistRequest
	QueryMeasurementAllowlistResponse       = v1.QueryMeasurementAllowlistResponse
	QueryMeasurementRequest                 = v1.QueryMeasurementRequest
	QueryMeasurementResponse                = v1.QueryMeasurementResponse
	QueryKeyRotationRequest                 = v1.QueryKeyRotationRequest
	QueryKeyRotationResponse                = v1.QueryKeyRotationResponse
	QueryValidKeySetRequest                 = v1.QueryValidKeySetRequest
	QueryValidKeySetResponse                = v1.QueryValidKeySetResponse
	QueryParamsRequest                      = v1.QueryParamsRequest
	QueryParamsResponse                     = v1.QueryParamsResponse
	QueryAttestedResultRequest              = v1.QueryAttestedResultRequest
	QueryAttestedResultResponse             = v1.QueryAttestedResultResponse
)

// Enum types
type (
	// TEEType represents the type of Trusted Execution Environment.
	TEEType = v1.TEEType

	// EnclaveIdentityStatus represents the status of an enclave identity.
	EnclaveIdentityStatus = v1.EnclaveIdentityStatus

	// KeyRotationStatus represents the status of a key rotation.
	KeyRotationStatus = v1.KeyRotationStatus
)

// TEEType enum values
const (
	TEETypeUnspecified = v1.TEETypeUnspecified
	TEETypeSGX         = v1.TEETypeSGX
	TEETypeSEVSNP      = v1.TEETypeSEVSNP
	TEETypeNitro       = v1.TEETypeNitro
	TEETypeTrustZone   = v1.TEETypeTrustZone
)

// EnclaveIdentityStatus enum values
const (
	EnclaveIdentityStatusUnspecified = v1.EnclaveIdentityStatusUnspecified
	EnclaveIdentityStatusActive      = v1.EnclaveIdentityStatusActive
	EnclaveIdentityStatusPending     = v1.EnclaveIdentityStatusPending
	EnclaveIdentityStatusExpired     = v1.EnclaveIdentityStatusExpired
	EnclaveIdentityStatusRevoked     = v1.EnclaveIdentityStatusRevoked
	EnclaveIdentityStatusRotating    = v1.EnclaveIdentityStatusRotating
)

// KeyRotationStatus enum values
const (
	KeyRotationStatusUnspecified = v1.KeyRotationStatusUnspecified
	KeyRotationStatusPending     = v1.KeyRotationStatusPending
	KeyRotationStatusActive      = v1.KeyRotationStatusActive
	KeyRotationStatusCompleted   = v1.KeyRotationStatusCompleted
	KeyRotationStatusCancelled   = v1.KeyRotationStatusCancelled
)

// RegisterMsgServer registers the MsgServer with a gRPC server.
var RegisterMsgServer = v1.RegisterMsgServer

// RegisterQueryServer registers the QueryServer with a gRPC server.
var RegisterQueryServer = v1.RegisterQueryServer

// Msg_serviceDesc is the gRPC service descriptor for the Msg service.
var Msg_serviceDesc = v1.Msg_serviceDesc
