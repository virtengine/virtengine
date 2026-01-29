// Package types provides VEID module types.
//
// This file defines appeal-related message types for the VEID module.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants for appeal messages
const (
	TypeMsgSubmitAppeal   = "submit_appeal"
	TypeMsgClaimAppeal    = "claim_appeal"
	TypeMsgResolveAppeal  = "resolve_appeal"
	TypeMsgWithdrawAppeal = "withdraw_appeal"
)

var (
	_ sdk.Msg = &MsgSubmitAppeal{}
	_ sdk.Msg = &MsgClaimAppeal{}
	_ sdk.Msg = &MsgResolveAppeal{}
	_ sdk.Msg = &MsgWithdrawAppeal{}
)

// ============================================================================
// MsgSubmitAppeal
// ============================================================================

// MsgSubmitAppeal is the message to submit an appeal against a verification decision
type MsgSubmitAppeal struct {
	// Submitter is the account address submitting the appeal (must own the scope)
	Submitter string `json:"submitter"`

	// ScopeID is the scope whose verification decision is being appealed
	ScopeID string `json:"scope_id"`

	// Reason is the explanation for why the submitter is appealing
	Reason string `json:"reason"`

	// EvidenceHashes are hashes of supporting evidence documents (stored off-chain)
	EvidenceHashes []string `json:"evidence_hashes,omitempty"`
}

// ProtoMessage is a no-op method to satisfy the proto.Message interface
func (*MsgSubmitAppeal) ProtoMessage() {}

// Reset is a no-op method to satisfy the proto.Message interface
func (m *MsgSubmitAppeal) Reset() { *m = MsgSubmitAppeal{} }

// String returns a string representation of the message
func (m *MsgSubmitAppeal) String() string {
	return "MsgSubmitAppeal{Submitter: " + m.Submitter + ", ScopeID: " + m.ScopeID + "}"
}

// NewMsgSubmitAppeal creates a new MsgSubmitAppeal
func NewMsgSubmitAppeal(submitter, scopeID, reason string, evidenceHashes []string) *MsgSubmitAppeal {
	return &MsgSubmitAppeal{
		Submitter:      submitter,
		ScopeID:        scopeID,
		Reason:         reason,
		EvidenceHashes: evidenceHashes,
	}
}

// Route returns the route for the message
func (msg MsgSubmitAppeal) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgSubmitAppeal) Type() string { return TypeMsgSubmitAppeal }

// GetSigners returns the expected signers
func (msg MsgSubmitAppeal) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs basic validation
func (msg MsgSubmitAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return ErrInvalidAddress.Wrap("invalid submitter address")
	}

	if msg.ScopeID == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if len(msg.Reason) < int(DefaultMinAppealReasonLength) {
		return ErrInvalidAppealReason.Wrapf("reason must be at least %d characters", DefaultMinAppealReasonLength)
	}

	if len(msg.Reason) > MaxAppealReasonLength {
		return ErrInvalidAppealReason.Wrapf("reason cannot exceed %d characters", MaxAppealReasonLength)
	}

	if len(msg.EvidenceHashes) > MaxEvidenceHashes {
		return ErrInvalidAppealRecord.Wrapf("cannot exceed %d evidence hashes", MaxEvidenceHashes)
	}

	// Validate evidence hash formats (should be hex strings of proper length)
	for i, hash := range msg.EvidenceHashes {
		if len(hash) != 64 { // SHA-256 hex string
			return ErrInvalidEvidenceHash.Wrapf("evidence hash %d must be 64 hex characters", i)
		}
	}

	return nil
}

// ============================================================================
// MsgSubmitAppealResponse
// ============================================================================

// MsgSubmitAppealResponse is the response for MsgSubmitAppeal
type MsgSubmitAppealResponse struct {
	AppealID     string       `json:"appeal_id"`
	Status       AppealStatus `json:"status"`
	AppealNumber uint32       `json:"appeal_number"`
	SubmittedAt  int64        `json:"submitted_at"`
}

// ============================================================================
// MsgClaimAppeal
// ============================================================================

// MsgClaimAppeal is the message for an arbitrator to claim an appeal for review
type MsgClaimAppeal struct {
	// Reviewer is the arbitrator claiming the appeal
	Reviewer string `json:"reviewer"`

	// AppealID is the appeal being claimed
	AppealID string `json:"appeal_id"`
}

// ProtoMessage is a no-op method to satisfy the proto.Message interface
func (*MsgClaimAppeal) ProtoMessage() {}

// Reset is a no-op method to satisfy the proto.Message interface
func (m *MsgClaimAppeal) Reset() { *m = MsgClaimAppeal{} }

// String returns a string representation of the message
func (m *MsgClaimAppeal) String() string {
	return "MsgClaimAppeal{Reviewer: " + m.Reviewer + ", AppealID: " + m.AppealID + "}"
}

// NewMsgClaimAppeal creates a new MsgClaimAppeal
func NewMsgClaimAppeal(reviewer, appealID string) *MsgClaimAppeal {
	return &MsgClaimAppeal{
		Reviewer: reviewer,
		AppealID: appealID,
	}
}

// Route returns the route for the message
func (msg MsgClaimAppeal) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgClaimAppeal) Type() string { return TypeMsgClaimAppeal }

// GetSigners returns the expected signers
func (msg MsgClaimAppeal) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Reviewer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs basic validation
func (msg MsgClaimAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Reviewer); err != nil {
		return ErrInvalidAddress.Wrap("invalid reviewer address")
	}

	if msg.AppealID == "" {
		return ErrInvalidAppealRecord.Wrap("appeal_id cannot be empty")
	}

	return nil
}

// ============================================================================
// MsgClaimAppealResponse
// ============================================================================

// MsgClaimAppealResponse is the response for MsgClaimAppeal
type MsgClaimAppealResponse struct {
	AppealID  string `json:"appeal_id"`
	ClaimedAt int64  `json:"claimed_at"`
}

// ============================================================================
// MsgResolveAppeal
// ============================================================================

// MsgResolveAppeal is the message to resolve an appeal (governance/arbitrator only)
type MsgResolveAppeal struct {
	// Resolver is the arbitrator resolving the appeal
	Resolver string `json:"resolver"`

	// AppealID is the appeal being resolved
	AppealID string `json:"appeal_id"`

	// Resolution is the resolution status (approved or rejected)
	Resolution AppealStatus `json:"resolution"`

	// Reason is the explanation for the resolution decision
	Reason string `json:"reason"`

	// ScoreAdjustment is the adjustment to the verification score (if approved)
	// Can be positive (to increase score) or negative (to decrease)
	ScoreAdjustment int32 `json:"score_adjustment"`
}

// ProtoMessage is a no-op method to satisfy the proto.Message interface
func (*MsgResolveAppeal) ProtoMessage() {}

// Reset is a no-op method to satisfy the proto.Message interface
func (m *MsgResolveAppeal) Reset() { *m = MsgResolveAppeal{} }

// String returns a string representation of the message
func (m *MsgResolveAppeal) String() string {
	return "MsgResolveAppeal{Resolver: " + m.Resolver + ", AppealID: " + m.AppealID + "}"
}

// NewMsgResolveAppeal creates a new MsgResolveAppeal
func NewMsgResolveAppeal(resolver, appealID string, resolution AppealStatus, reason string, scoreAdjustment int32) *MsgResolveAppeal {
	return &MsgResolveAppeal{
		Resolver:        resolver,
		AppealID:        appealID,
		Resolution:      resolution,
		Reason:          reason,
		ScoreAdjustment: scoreAdjustment,
	}
}

// Route returns the route for the message
func (msg MsgResolveAppeal) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgResolveAppeal) Type() string { return TypeMsgResolveAppeal }

// GetSigners returns the expected signers
func (msg MsgResolveAppeal) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Resolver)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs basic validation
func (msg MsgResolveAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Resolver); err != nil {
		return ErrInvalidAddress.Wrap("invalid resolver address")
	}

	if msg.AppealID == "" {
		return ErrInvalidAppealRecord.Wrap("appeal_id cannot be empty")
	}

	if msg.Resolution != AppealStatusApproved && msg.Resolution != AppealStatusRejected {
		return ErrInvalidAppealResolution.Wrapf("resolution must be approved or rejected, got: %s", msg.Resolution.String())
	}

	if len(msg.Reason) == 0 {
		return ErrInvalidAppealResolution.Wrap("resolution reason cannot be empty")
	}

	// Score adjustment bounds check
	if msg.ScoreAdjustment < -100 || msg.ScoreAdjustment > 100 {
		return ErrInvalidScoreAdjustment.Wrapf("score adjustment must be between -100 and 100, got: %d", msg.ScoreAdjustment)
	}

	return nil
}

// ============================================================================
// MsgResolveAppealResponse
// ============================================================================

// MsgResolveAppealResponse is the response for MsgResolveAppeal
type MsgResolveAppealResponse struct {
	AppealID   string       `json:"appeal_id"`
	Resolution AppealStatus `json:"resolution"`
	ResolvedAt int64        `json:"resolved_at"`
}

// ============================================================================
// MsgWithdrawAppeal
// ============================================================================

// MsgWithdrawAppeal allows the submitter to withdraw their appeal
type MsgWithdrawAppeal struct {
	// Submitter is the original appeal submitter
	Submitter string `json:"submitter"`

	// AppealID is the appeal to withdraw
	AppealID string `json:"appeal_id"`
}

// ProtoMessage is a no-op method to satisfy the proto.Message interface
func (*MsgWithdrawAppeal) ProtoMessage() {}

// Reset is a no-op method to satisfy the proto.Message interface
func (m *MsgWithdrawAppeal) Reset() { *m = MsgWithdrawAppeal{} }

// String returns a string representation of the message
func (m *MsgWithdrawAppeal) String() string {
	return "MsgWithdrawAppeal{Submitter: " + m.Submitter + ", AppealID: " + m.AppealID + "}"
}

// NewMsgWithdrawAppeal creates a new MsgWithdrawAppeal
func NewMsgWithdrawAppeal(submitter, appealID string) *MsgWithdrawAppeal {
	return &MsgWithdrawAppeal{
		Submitter: submitter,
		AppealID:  appealID,
	}
}

// Route returns the route for the message
func (msg MsgWithdrawAppeal) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgWithdrawAppeal) Type() string { return TypeMsgWithdrawAppeal }

// GetSigners returns the expected signers
func (msg MsgWithdrawAppeal) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs basic validation
func (msg MsgWithdrawAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return ErrInvalidAddress.Wrap("invalid submitter address")
	}

	if msg.AppealID == "" {
		return ErrInvalidAppealRecord.Wrap("appeal_id cannot be empty")
	}

	return nil
}

// ============================================================================
// MsgWithdrawAppealResponse
// ============================================================================

// MsgWithdrawAppealResponse is the response for MsgWithdrawAppeal
type MsgWithdrawAppealResponse struct {
	AppealID    string `json:"appeal_id"`
	WithdrawnAt int64  `json:"withdrawn_at"`
}
