// Package types provides message types for the VEID module.
//
// This file defines type aliases to the buf-generated protobuf types in
// sdk/go/node/veid/v1. The generated types already implement:
// - proto.Message interface (ProtoMessage, Reset, String)
// - sdk.Msg interface (Route, Type, ValidateBasic, GetSigners)
//
// Using type aliases ensures deterministic serialization across nodes
// and avoids maintaining duplicate hand-written proto stubs.
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// Message type constants - kept for backwards compatibility
const (
	TypeMsgUploadScope                  = "upload_scope"
	TypeMsgRevokeScope                  = "revoke_scope"
	TypeMsgRequestVerification          = "request_verification"
	TypeMsgUpdateVerificationStatus     = "update_verification_status"
	TypeMsgUpdateScore                  = "update_score"
	TypeMsgSubmitSSOVerificationProof   = "submit_sso_verification_proof"
	TypeMsgSubmitEmailVerificationProof = "submit_email_verification_proof"
	TypeMsgSubmitSMSVerificationProof   = "submit_sms_verification_proof"
	TypeMsgSubmitSocialMediaScope       = "submit_social_media_scope"
)

// ============================================================================
// Core VEID Message Type Aliases
// These types alias the buf-generated protobuf types which already implement
// proto.Message and sdk.Msg interfaces.
// ============================================================================

// MsgUploadScope is the message for uploading an identity scope
type MsgUploadScope = veidv1.MsgUploadScope

// MsgUploadScopeResponse is the response for MsgUploadScope
type MsgUploadScopeResponse = veidv1.MsgUploadScopeResponse

// MsgRevokeScope is the message for revoking an identity scope
type MsgRevokeScope = veidv1.MsgRevokeScope

// MsgRevokeScopeResponse is the response for MsgRevokeScope
type MsgRevokeScopeResponse = veidv1.MsgRevokeScopeResponse

// MsgRequestVerification is the message for requesting verification of a scope
type MsgRequestVerification = veidv1.MsgRequestVerification

// MsgRequestVerificationResponse is the response for MsgRequestVerification
type MsgRequestVerificationResponse = veidv1.MsgRequestVerificationResponse

// MsgUpdateVerificationStatus is the message for validators to update verification status
type MsgUpdateVerificationStatus = veidv1.MsgUpdateVerificationStatus

// MsgUpdateVerificationStatusResponse is the response for MsgUpdateVerificationStatus
type MsgUpdateVerificationStatusResponse = veidv1.MsgUpdateVerificationStatusResponse

// MsgUpdateScore is the message for validators to update identity score
type MsgUpdateScore = veidv1.MsgUpdateScore

// MsgUpdateScoreResponse is the response for MsgUpdateScore
type MsgUpdateScoreResponse = veidv1.MsgUpdateScoreResponse

// MsgSubmitSSOVerificationProof is the message for SSO verification proof submission
type MsgSubmitSSOVerificationProof = veidv1.MsgSubmitSSOVerificationProof

// MsgSubmitSSOVerificationProofResponse is the response for SSO proof submission
type MsgSubmitSSOVerificationProofResponse = veidv1.MsgSubmitSSOVerificationProofResponse

// MsgSubmitEmailVerificationProof is the message for email verification proof submission
type MsgSubmitEmailVerificationProof = veidv1.MsgSubmitEmailVerificationProof

// MsgSubmitEmailVerificationProofResponse is the response for email proof submission
type MsgSubmitEmailVerificationProofResponse = veidv1.MsgSubmitEmailVerificationProofResponse

// MsgSubmitSMSVerificationProof is the message for SMS verification proof submission
type MsgSubmitSMSVerificationProof = veidv1.MsgSubmitSMSVerificationProof

// MsgSubmitSMSVerificationProofResponse is the response for SMS proof submission
type MsgSubmitSMSVerificationProofResponse = veidv1.MsgSubmitSMSVerificationProofResponse

// MsgSubmitSocialMediaScope is the message for social media scope submission
type MsgSubmitSocialMediaScope = veidv1.MsgSubmitSocialMediaScope

// MsgSubmitSocialMediaScopeResponse is the response for social media scope submission
type MsgSubmitSocialMediaScopeResponse = veidv1.MsgSubmitSocialMediaScopeResponse

// EncryptedPayloadEnvelope is the encrypted payload type from proto
type EncryptedPayloadEnvelope = veidv1.EncryptedPayloadEnvelope

// MsgUpdateParams is the message for updating module params
type MsgUpdateParams = veidv1.MsgUpdateParams

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse = veidv1.MsgUpdateParamsResponse

// ============================================================================
// Message Constructors
// ============================================================================

// NewMsgUploadScope creates a new MsgUploadScope.
// Accepts encryptiontypes.EncryptedPayloadEnvelope and converts to proto type.
func NewMsgUploadScope(
	sender string,
	scopeID string,
	scopeType ScopeType,
	encryptedPayload encryptiontypes.EncryptedPayloadEnvelope,
	salt []byte,
	deviceFingerprint string,
	clientID string,
	clientSignature []byte,
	userSignature []byte,
	payloadHash []byte,
) *MsgUploadScope {
	// Convert encryptiontypes envelope to proto envelope
	protoEnvelope := EncryptedPayloadEnvelope{
		Version:             encryptedPayload.Version,
		AlgorithmId:         encryptedPayload.AlgorithmID,
		AlgorithmVersion:    encryptedPayload.AlgorithmVersion,
		RecipientKeyIds:     encryptedPayload.RecipientKeyIDs,
		RecipientPublicKeys: encryptedPayload.RecipientPublicKeys,
		EncryptedKeys:       encryptedPayload.EncryptedKeys,
		Nonce:               encryptedPayload.Nonce,
		Ciphertext:          encryptedPayload.Ciphertext,
		SenderSignature:     encryptedPayload.SenderSignature,
		SenderPubKey:        encryptedPayload.SenderPubKey,
		Metadata:            encryptedPayload.Metadata,
	}

	return &MsgUploadScope{
		Sender:            sender,
		ScopeId:           scopeID,
		ScopeType:         ScopeTypeToProto(scopeType),
		EncryptedPayload:  protoEnvelope,
		Salt:              salt,
		DeviceFingerprint: deviceFingerprint,
		ClientId:          clientID,
		ClientSignature:   clientSignature,
		UserSignature:     userSignature,
		PayloadHash:       payloadHash,
	}
}

// NewMsgUpdateScore creates a new MsgUpdateScore.
func NewMsgUpdateScore(
	sender string,
	accountAddress string,
	newScore uint32,
	scoreVersion string,
) *MsgUpdateScore {
	return &MsgUpdateScore{
		Sender:         sender,
		AccountAddress: accountAddress,
		NewScore:       newScore,
		ScoreVersion:   scoreVersion,
	}
}
