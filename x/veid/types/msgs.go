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
)

// Message type constants - kept for backwards compatibility
const (
	TypeMsgUploadScope              = "upload_scope"
	TypeMsgRevokeScope              = "revoke_scope"
	TypeMsgRequestVerification      = "request_verification"
	TypeMsgUpdateVerificationStatus = "update_verification_status"
	TypeMsgUpdateScore              = "update_score"
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

// EncryptedPayloadEnvelope is the encrypted payload type from proto
type EncryptedPayloadEnvelope = veidv1.EncryptedPayloadEnvelope
