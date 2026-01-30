// Package types provides VEID module types.
//
// This file defines type aliases for borderline-related VEID messages, using the
// proto-generated types from sdk/go/node/veid/v1 as the source of truth.
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

const (
	// TypeMsgCompleteBorderlineFallback is the type for MsgCompleteBorderlineFallback
	TypeMsgCompleteBorderlineFallback = "complete_borderline_fallback"

	// TypeMsgUpdateBorderlineParams is the type for MsgUpdateBorderlineParams
	TypeMsgUpdateBorderlineParams = "update_borderline_params"
)

// ============================================================================
// Borderline Message Type Aliases - from proto-generated types
// ============================================================================

// MsgCompleteBorderlineFallback is the message for completing a borderline fallback.
type MsgCompleteBorderlineFallback = veidv1.MsgCompleteBorderlineFallback

// MsgCompleteBorderlineFallbackResponse is the response for MsgCompleteBorderlineFallback.
type MsgCompleteBorderlineFallbackResponse = veidv1.MsgCompleteBorderlineFallbackResponse

// MsgUpdateBorderlineParams is the message for updating borderline parameters.
type MsgUpdateBorderlineParams = veidv1.MsgUpdateBorderlineParams

// MsgUpdateBorderlineParamsResponse is the response for MsgUpdateBorderlineParams.
type MsgUpdateBorderlineParamsResponse = veidv1.MsgUpdateBorderlineParamsResponse

// ============================================================================
// Borderline Constructor Functions
// ============================================================================

// NewMsgCompleteBorderlineFallback creates a new MsgCompleteBorderlineFallback.
func NewMsgCompleteBorderlineFallback(
	sender string,
	challengeID string,
	factorsSatisfied []string,
) *MsgCompleteBorderlineFallback {
	return &MsgCompleteBorderlineFallback{
		Sender:           sender,
		ChallengeId:      challengeID,
		FactorsSatisfied: factorsSatisfied,
	}
}

// NewMsgUpdateBorderlineParams creates a new MsgUpdateBorderlineParams.
func NewMsgUpdateBorderlineParams(authority string, params BorderlineParams) *MsgUpdateBorderlineParams {
	return &MsgUpdateBorderlineParams{
		Authority: authority,
		Params:    params,
	}
}

// BorderlineParams is the proto-generated borderline parameters type.
type BorderlineParams = veidv1.BorderlineParams
