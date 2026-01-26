package enclave

import (
	"pkg.akt.dev/node/x/enclave/keeper"
	"pkg.akt.dev/node/x/enclave/types"
)

// Module aliases for external use
const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

// Keeper aliases
type (
	Keeper = keeper.Keeper
)

// Type aliases
type (
	// Types
	EnclaveIdentity       = types.EnclaveIdentity
	EnclaveIdentityStatus = types.EnclaveIdentityStatus
	MeasurementRecord     = types.MeasurementRecord
	KeyRotationRecord     = types.KeyRotationRecord
	KeyRotationStatus     = types.KeyRotationStatus
	AttestedScoringResult = types.AttestedScoringResult
	TEEType               = types.TEEType
	Params                = types.Params
	GenesisState          = types.GenesisState

	// Messages
	MsgRegisterEnclaveIdentity = types.MsgRegisterEnclaveIdentity
	MsgRotateEnclaveIdentity   = types.MsgRotateEnclaveIdentity
	MsgProposeMeasurement      = types.MsgProposeMeasurement
	MsgRevokeMeasurement       = types.MsgRevokeMeasurement

	// Queries
	QueryEnclaveIdentityRequest              = types.QueryEnclaveIdentityRequest
	QueryEnclaveIdentityResponse             = types.QueryEnclaveIdentityResponse
	QueryActiveValidatorEnclaveKeysRequest   = types.QueryActiveValidatorEnclaveKeysRequest
	QueryActiveValidatorEnclaveKeysResponse  = types.QueryActiveValidatorEnclaveKeysResponse
	QueryMeasurementAllowlistRequest         = types.QueryMeasurementAllowlistRequest
	QueryMeasurementAllowlistResponse        = types.QueryMeasurementAllowlistResponse
	ValidatorKeyInfo                         = types.ValidatorKeyInfo
)

// Status constants
const (
	EnclaveIdentityStatusActive   = types.EnclaveIdentityStatusActive
	EnclaveIdentityStatusPending  = types.EnclaveIdentityStatusPending
	EnclaveIdentityStatusExpired  = types.EnclaveIdentityStatusExpired
	EnclaveIdentityStatusRevoked  = types.EnclaveIdentityStatusRevoked
	EnclaveIdentityStatusRotating = types.EnclaveIdentityStatusRotating
)

// TEE type constants
const (
	TEETypeSGX       = types.TEETypeSGX
	TEETypeSEVSNP    = types.TEETypeSEVSNP
	TEETypeNitro     = types.TEETypeNitro
	TEETypeTrustZone = types.TEETypeTrustZone
)

// Function aliases
var (
	NewKeeper           = keeper.NewKeeper
	DefaultParams       = types.DefaultParams
	DefaultGenesisState = types.DefaultGenesisState
)
