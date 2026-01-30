// Package types provides type aliases and adapters for the generated proto types.
//
// This file bridges the local encryption types in x/encryption/types with the generated
// protobuf types in sdk/go/node/encryption/v1. It provides type aliases for proto message
// types that are used for gRPC and Cosmos SDK message routing.
package types

import (
	"context"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
)

// =============================================================================
// Proto Type Aliases
// =============================================================================

// Proto message type aliases for Msg types
type (
	// MsgRegisterRecipientKeyPB is the generated proto type
	MsgRegisterRecipientKeyPB = encryptionv1.MsgRegisterRecipientKey
	// MsgRegisterRecipientKeyResponsePB is the generated proto type
	MsgRegisterRecipientKeyResponsePB = encryptionv1.MsgRegisterRecipientKeyResponse
	// MsgRevokeRecipientKeyPB is the generated proto type
	MsgRevokeRecipientKeyPB = encryptionv1.MsgRevokeRecipientKey
	// MsgRevokeRecipientKeyResponsePB is the generated proto type
	MsgRevokeRecipientKeyResponsePB = encryptionv1.MsgRevokeRecipientKeyResponse
	// MsgUpdateKeyLabelPB is the generated proto type
	MsgUpdateKeyLabelPB = encryptionv1.MsgUpdateKeyLabel
	// MsgUpdateKeyLabelResponsePB is the generated proto type
	MsgUpdateKeyLabelResponsePB = encryptionv1.MsgUpdateKeyLabelResponse
)

// Proto message type aliases for core types
type (
	// GenesisStatePB is the generated proto type for GenesisState
	GenesisStatePB = encryptionv1.GenesisState
	// ParamsPB is the generated proto type for Params
	ParamsPB = encryptionv1.Params
	// RecipientKeyRecordPB is the generated proto type for RecipientKeyRecord
	RecipientKeyRecordPB = encryptionv1.RecipientKeyRecord
	// AlgorithmInfoPB is the generated proto type for AlgorithmInfo
	AlgorithmInfoPB = encryptionv1.AlgorithmInfo
	// EncryptedPayloadEnvelopePB is the generated proto type for EncryptedPayloadEnvelope
	EncryptedPayloadEnvelopePB = encryptionv1.EncryptedPayloadEnvelope
	// WrappedKeyEntryPB is the generated proto type for WrappedKeyEntry
	WrappedKeyEntryPB = encryptionv1.WrappedKeyEntry
	// MultiRecipientEnvelopePB is the generated proto type for MultiRecipientEnvelope
	MultiRecipientEnvelopePB = encryptionv1.MultiRecipientEnvelope
)

// Proto message type aliases for event types
type (
	// EventKeyRegisteredPB is the generated proto type for EventKeyRegistered
	EventKeyRegisteredPB = encryptionv1.EventKeyRegistered
	// EventKeyRevokedPB is the generated proto type for EventKeyRevoked
	EventKeyRevokedPB = encryptionv1.EventKeyRevoked
	// EventKeyUpdatedPB is the generated proto type for EventKeyUpdated
	EventKeyUpdatedPB = encryptionv1.EventKeyUpdated
)

// Proto message type aliases for query types
type (
	// QueryRecipientKeyRequestPB is the generated proto type
	QueryRecipientKeyRequestPB = encryptionv1.QueryRecipientKeyRequest
	// QueryRecipientKeyResponsePB is the generated proto type
	QueryRecipientKeyResponsePB = encryptionv1.QueryRecipientKeyResponse
	// QueryKeyByFingerprintRequestPB is the generated proto type
	QueryKeyByFingerprintRequestPB = encryptionv1.QueryKeyByFingerprintRequest
	// QueryKeyByFingerprintResponsePB is the generated proto type
	QueryKeyByFingerprintResponsePB = encryptionv1.QueryKeyByFingerprintResponse
	// QueryParamsRequestPB is the generated proto type
	QueryParamsRequestPB = encryptionv1.QueryParamsRequest
	// QueryParamsResponsePB is the generated proto type
	QueryParamsResponsePB = encryptionv1.QueryParamsResponse
	// QueryAlgorithmsRequestPB is the generated proto type
	QueryAlgorithmsRequestPB = encryptionv1.QueryAlgorithmsRequest
	// QueryAlgorithmsResponsePB is the generated proto type
	QueryAlgorithmsResponsePB = encryptionv1.QueryAlgorithmsResponse
	// QueryValidateEnvelopeRequestPB is the generated proto type
	QueryValidateEnvelopeRequestPB = encryptionv1.QueryValidateEnvelopeRequest
	// QueryValidateEnvelopeResponsePB is the generated proto type
	QueryValidateEnvelopeResponsePB = encryptionv1.QueryValidateEnvelopeResponse
)

// Proto enum type aliases
type (
	// RecipientModePB is the generated proto enum for RecipientMode
	RecipientModePB = encryptionv1.RecipientMode
)

// Proto enum value constants
const (
	RecipientModePBUnspecified       = encryptionv1.RecipientMode_RECIPIENT_MODE_UNSPECIFIED
	RecipientModePBFullValidatorSet  = encryptionv1.RecipientMode_RECIPIENT_MODE_FULL_VALIDATOR_SET
	RecipientModePBCommittee         = encryptionv1.RecipientMode_RECIPIENT_MODE_COMMITTEE
	RecipientModePBSpecific          = encryptionv1.RecipientMode_RECIPIENT_MODE_SPECIFIC
)

// =============================================================================
// MsgServer Adapter
// =============================================================================

// msgServerAdapter adapts the local MsgServer interface to the generated proto MsgServer.
type msgServerAdapter struct {
	local MsgServerLocal
}

// MsgServerLocal is the local message server interface
type MsgServerLocal interface {
	RegisterRecipientKey(ctx context.Context, msg *MsgRegisterRecipientKeyPB) (*MsgRegisterRecipientKeyResponsePB, error)
	RevokeRecipientKey(ctx context.Context, msg *MsgRevokeRecipientKeyPB) (*MsgRevokeRecipientKeyResponsePB, error)
	UpdateKeyLabel(ctx context.Context, msg *MsgUpdateKeyLabelPB) (*MsgUpdateKeyLabelResponsePB, error)
}

// NewMsgServerAdapter creates a new adapter that wraps a local MsgServer
func NewMsgServerAdapter(local MsgServerLocal) encryptionv1.MsgServer {
	return &msgServerAdapter{local: local}
}

func (a *msgServerAdapter) RegisterRecipientKey(ctx context.Context, req *encryptionv1.MsgRegisterRecipientKey) (*encryptionv1.MsgRegisterRecipientKeyResponse, error) {
	return a.local.RegisterRecipientKey(ctx, req)
}

func (a *msgServerAdapter) RevokeRecipientKey(ctx context.Context, req *encryptionv1.MsgRevokeRecipientKey) (*encryptionv1.MsgRevokeRecipientKeyResponse, error) {
	return a.local.RevokeRecipientKey(ctx, req)
}

func (a *msgServerAdapter) UpdateKeyLabel(ctx context.Context, req *encryptionv1.MsgUpdateKeyLabel) (*encryptionv1.MsgUpdateKeyLabelResponse, error) {
	return a.local.UpdateKeyLabel(ctx, req)
}

// =============================================================================
// QueryServer Adapter
// =============================================================================

// queryServerAdapter adapts the local QueryServer interface to the generated proto QueryServer.
type queryServerAdapter struct {
	local QueryServerLocal
}

// QueryServerLocal is the local query server interface
type QueryServerLocal interface {
	RecipientKey(ctx context.Context, req *QueryRecipientKeyRequestPB) (*QueryRecipientKeyResponsePB, error)
	KeyByFingerprint(ctx context.Context, req *QueryKeyByFingerprintRequestPB) (*QueryKeyByFingerprintResponsePB, error)
	Params(ctx context.Context, req *QueryParamsRequestPB) (*QueryParamsResponsePB, error)
	Algorithms(ctx context.Context, req *QueryAlgorithmsRequestPB) (*QueryAlgorithmsResponsePB, error)
	ValidateEnvelope(ctx context.Context, req *QueryValidateEnvelopeRequestPB) (*QueryValidateEnvelopeResponsePB, error)
}

// NewQueryServerAdapter creates a new adapter that wraps a local QueryServer
func NewQueryServerAdapter(local QueryServerLocal) encryptionv1.QueryServer {
	return &queryServerAdapter{local: local}
}

func (a *queryServerAdapter) RecipientKey(ctx context.Context, req *encryptionv1.QueryRecipientKeyRequest) (*encryptionv1.QueryRecipientKeyResponse, error) {
	return a.local.RecipientKey(ctx, req)
}

func (a *queryServerAdapter) KeyByFingerprint(ctx context.Context, req *encryptionv1.QueryKeyByFingerprintRequest) (*encryptionv1.QueryKeyByFingerprintResponse, error) {
	return a.local.KeyByFingerprint(ctx, req)
}

func (a *queryServerAdapter) Params(ctx context.Context, req *encryptionv1.QueryParamsRequest) (*encryptionv1.QueryParamsResponse, error) {
	return a.local.Params(ctx, req)
}

func (a *queryServerAdapter) Algorithms(ctx context.Context, req *encryptionv1.QueryAlgorithmsRequest) (*encryptionv1.QueryAlgorithmsResponse, error) {
	return a.local.Algorithms(ctx, req)
}

func (a *queryServerAdapter) ValidateEnvelope(ctx context.Context, req *encryptionv1.QueryValidateEnvelopeRequest) (*encryptionv1.QueryValidateEnvelopeResponse, error) {
	return a.local.ValidateEnvelope(ctx, req)
}
