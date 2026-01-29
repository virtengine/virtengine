// Package types provides VEID module types.
//
// This file defines type aliases for core VEID messages, using the proto-generated
// types from sdk/go/node/veid/v1 as the source of truth.
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// Message type constants
const (
	TypeMsgUploadScope              = "upload_scope"
	TypeMsgRevokeScope              = "revoke_scope"
	TypeMsgRequestVerification      = "request_verification"
	TypeMsgUpdateVerificationStatus = "update_verification_status"
	TypeMsgUpdateScore              = "update_score"
)

// ============================================================================
// Core Message Type Aliases - from proto-generated types
// ============================================================================

// MsgUploadScope is the message for uploading an identity scope.
// Uses the proto-generated type from sdk/go/node/veid/v1.
type MsgUploadScope = veidv1.MsgUploadScope

// MsgUploadScopeResponse is the response for MsgUploadScope.
type MsgUploadScopeResponse = veidv1.MsgUploadScopeResponse

// MsgRevokeScope is the message for revoking an identity scope.
type MsgRevokeScope = veidv1.MsgRevokeScope

// MsgRevokeScopeResponse is the response for MsgRevokeScope.
type MsgRevokeScopeResponse = veidv1.MsgRevokeScopeResponse

// MsgRequestVerification is the message for requesting verification of a scope.
type MsgRequestVerification = veidv1.MsgRequestVerification

// MsgRequestVerificationResponse is the response for MsgRequestVerification.
type MsgRequestVerificationResponse = veidv1.MsgRequestVerificationResponse

// MsgUpdateVerificationStatus is the message for validators to update verification status.
type MsgUpdateVerificationStatus = veidv1.MsgUpdateVerificationStatus

// MsgUpdateVerificationStatusResponse is the response for MsgUpdateVerificationStatus.
type MsgUpdateVerificationStatusResponse = veidv1.MsgUpdateVerificationStatusResponse

// MsgUpdateScore is the message for validators to update identity score.
type MsgUpdateScore = veidv1.MsgUpdateScore

// MsgUpdateScoreResponse is the response for MsgUpdateScore.
type MsgUpdateScoreResponse = veidv1.MsgUpdateScoreResponse

// MsgUpdateParams is the message for updating module parameters (governance only).
type MsgUpdateParams = veidv1.MsgUpdateParams

// MsgUpdateParamsResponse is the response for MsgUpdateParams.
type MsgUpdateParamsResponse = veidv1.MsgUpdateParamsResponse

// EncryptedPayloadEnvelope is the canonical encrypted payload structure.
// This alias provides backward compatibility with existing code.
type EncryptedPayloadEnvelope = veidv1.EncryptedPayloadEnvelope

// ============================================================================
// Constructor Functions
// ============================================================================

// NewMsgUploadScope creates a new MsgUploadScope.
// Note: Uses proto field names (ScopeId, ClientId, etc.)
func NewMsgUploadScope(
	sender string,
	scopeID string,
	scopeType ScopeType,
	payload EncryptedPayloadEnvelope,
	salt []byte,
	deviceFingerprint string,
	clientID string,
	clientSignature []byte,
	userSignature []byte,
	payloadHash []byte,
) *MsgUploadScope {
	return &MsgUploadScope{
		Sender:            sender,
		ScopeId:           scopeID,
		ScopeType:         ScopeTypeToProto(scopeType),
		EncryptedPayload:  payload,
		Salt:              salt,
		DeviceFingerprint: deviceFingerprint,
		ClientId:          clientID,
		ClientSignature:   clientSignature,
		UserSignature:     userSignature,
		PayloadHash:       payloadHash,
	}
}

// NewMsgRevokeScope creates a new MsgRevokeScope.
func NewMsgRevokeScope(sender, scopeID, reason string) *MsgRevokeScope {
	return &MsgRevokeScope{
		Sender:  sender,
		ScopeId: scopeID,
		Reason:  reason,
	}
}

// NewMsgRequestVerification creates a new MsgRequestVerification.
func NewMsgRequestVerification(sender, scopeID string) *MsgRequestVerification {
	return &MsgRequestVerification{
		Sender:  sender,
		ScopeId: scopeID,
	}
}

// NewMsgUpdateVerificationStatus creates a new MsgUpdateVerificationStatus.
func NewMsgUpdateVerificationStatus(sender, accountAddress, scopeID string, newStatus VerificationStatus, reason string) *MsgUpdateVerificationStatus {
	return &MsgUpdateVerificationStatus{
		Sender:         sender,
		AccountAddress: accountAddress,
		ScopeId:        scopeID,
		NewStatus:      VerificationStatusToProto(newStatus),
		Reason:         reason,
	}
}

// NewMsgUpdateScore creates a new MsgUpdateScore.
func NewMsgUpdateScore(sender, accountAddress string, newScore uint32, scoreVersion string) *MsgUpdateScore {
	return &MsgUpdateScore{
		Sender:         sender,
		AccountAddress: accountAddress,
		NewScore:       newScore,
		ScoreVersion:   scoreVersion,
	}
}
