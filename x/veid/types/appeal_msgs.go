// Package types provides VEID module types.
//
// This file defines type aliases for appeal-related VEID messages, using the
// proto-generated types from sdk/go/node/veid/v1 as the source of truth.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// Message type constants for appeal messages
const (
	TypeMsgSubmitAppeal   = "submit_appeal"
	TypeMsgClaimAppeal    = "claim_appeal"
	TypeMsgResolveAppeal  = "resolve_appeal"
	TypeMsgWithdrawAppeal = "withdraw_appeal"
)

// ============================================================================
// Appeal Message Type Aliases - from proto-generated types
// ============================================================================

// MsgSubmitAppeal is the message to submit an appeal against a verification decision.
type MsgSubmitAppeal = veidv1.MsgSubmitAppeal

// MsgSubmitAppealResponse is the response for MsgSubmitAppeal.
type MsgSubmitAppealResponse = veidv1.MsgSubmitAppealResponse

// MsgClaimAppeal is the message for an arbitrator to claim an appeal for review.
type MsgClaimAppeal = veidv1.MsgClaimAppeal

// MsgClaimAppealResponse is the response for MsgClaimAppeal.
type MsgClaimAppealResponse = veidv1.MsgClaimAppealResponse

// MsgResolveAppeal is the message to resolve an appeal (governance/arbitrator only).
type MsgResolveAppeal = veidv1.MsgResolveAppeal

// MsgResolveAppealResponse is the response for MsgResolveAppeal.
type MsgResolveAppealResponse = veidv1.MsgResolveAppealResponse

// MsgWithdrawAppeal allows the submitter to withdraw their appeal.
type MsgWithdrawAppeal = veidv1.MsgWithdrawAppeal

// MsgWithdrawAppealResponse is the response for MsgWithdrawAppeal.
type MsgWithdrawAppealResponse = veidv1.MsgWithdrawAppealResponse

// ============================================================================
// Appeal Constructor Functions
// ============================================================================

// NewMsgSubmitAppeal creates a new MsgSubmitAppeal.
func NewMsgSubmitAppeal(submitter, scopeID, reason string, evidenceHashes []string) *MsgSubmitAppeal {
	return &MsgSubmitAppeal{
		Submitter:      submitter,
		ScopeId:        scopeID,
		Reason:         reason,
		EvidenceHashes: evidenceHashes,
	}
}

// NewMsgClaimAppeal creates a new MsgClaimAppeal.
func NewMsgClaimAppeal(reviewer, appealID string) *MsgClaimAppeal {
	return &MsgClaimAppeal{
		Reviewer: reviewer,
		AppealId: appealID,
	}
}

// NewMsgResolveAppeal creates a new MsgResolveAppeal.
func NewMsgResolveAppeal(resolver, appealID string, resolution AppealStatus, reason string, scoreAdjustment int32) *MsgResolveAppeal {
	return &MsgResolveAppeal{
		Resolver:        resolver,
		AppealId:        appealID,
		Resolution:      AppealStatusToProto(resolution),
		Reason:          reason,
		ScoreAdjustment: scoreAdjustment,
	}
}

// NewMsgWithdrawAppeal creates a new MsgWithdrawAppeal.
func NewMsgWithdrawAppeal(submitter, appealID string) *MsgWithdrawAppeal {
	return &MsgWithdrawAppeal{
		Submitter: submitter,
		AppealId:  appealID,
	}
}
