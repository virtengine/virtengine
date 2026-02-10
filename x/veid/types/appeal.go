// Package types provides VEID module types.
//
// This file defines the appeal system types for contesting verification decisions.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// Appeal Constants
// ============================================================================

// Appeal configuration defaults
const (
	// DefaultAppealWindowBlocks is how long after rejection can user appeal (7 days @ 6s blocks)
	DefaultAppealWindowBlocks int64 = 100800

	// DefaultMaxAppealsPerScope is the maximum appeals allowed per scope
	DefaultMaxAppealsPerScope uint32 = 3

	// DefaultMinAppealReasonLength is the minimum characters for appeal reason
	DefaultMinAppealReasonLength uint32 = 50

	// DefaultAppealReviewTimeoutBlocks is how long an appeal can stay in reviewing status
	DefaultAppealReviewTimeoutBlocks int64 = 50400 // 3.5 days @ 6s blocks

	// MaxAppealReasonLength is the maximum length of an appeal reason
	MaxAppealReasonLength = 2000

	// MaxEvidenceHashes is the maximum number of evidence hashes per appeal
	MaxEvidenceHashes = 10
)

// ============================================================================
// Appeal Status
// ============================================================================

// AppealStatus represents the current state of an appeal
type AppealStatus int32

const (
	// AppealStatusUnspecified is the default unspecified status
	AppealStatusUnspecified AppealStatus = 0

	// AppealStatusPending indicates the appeal has been submitted and awaits review
	AppealStatusPending AppealStatus = 1

	// AppealStatusReviewing indicates an arbitrator has claimed the appeal for review
	AppealStatusReviewing AppealStatus = 2

	// AppealStatusApproved indicates the appeal was approved and verification should be reconsidered
	AppealStatusApproved AppealStatus = 3

	// AppealStatusRejected indicates the appeal was rejected and the original decision stands
	AppealStatusRejected AppealStatus = 4

	// AppealStatusWithdrawn indicates the submitter withdrew their appeal
	AppealStatusWithdrawn AppealStatus = 5

	// AppealStatusExpired indicates the appeal expired without resolution
	AppealStatusExpired AppealStatus = 6
)

// String returns the string representation of the appeal status
func (s AppealStatus) String() string {
	switch s {
	case AppealStatusUnspecified:
		return "UNSPECIFIED"
	case AppealStatusPending:
		return "PENDING"
	case AppealStatusReviewing:
		return "REVIEWING"
	case AppealStatusApproved:
		return "APPROVED"
	case AppealStatusRejected:
		return "REJECTED"
	case AppealStatusWithdrawn:
		return "WITHDRAWN"
	case AppealStatusExpired:
		return "EXPIRED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}

// IsTerminal returns true if the appeal is in a terminal state
func (s AppealStatus) IsTerminal() bool {
	switch s {
	case AppealStatusApproved, AppealStatusRejected, AppealStatusWithdrawn, AppealStatusExpired:
		return true
	default:
		return false
	}
}

// IsActive returns true if the appeal is still active (pending or reviewing)
func (s AppealStatus) IsActive() bool {
	return s == AppealStatusPending || s == AppealStatusReviewing
}

// ============================================================================
// Appeal Record
// ============================================================================

// AppealRecord tracks an appeal against a verification decision
type AppealRecord struct {
	// AppealID is the unique identifier for this appeal
	AppealID string `json:"appeal_id"`

	// AccountAddress is the address of the account filing the appeal
	AccountAddress string `json:"account_address"`

	// ScopeID is the scope whose verification decision is being appealed
	ScopeID string `json:"scope_id"`

	// OriginalStatus is the verification status that prompted the appeal
	OriginalStatus string `json:"original_status"`

	// OriginalScore is the verification score at the time of appeal (if applicable)
	OriginalScore uint32 `json:"original_score,omitempty"`

	// AppealReason is the user's explanation for why they are appealing
	AppealReason string `json:"appeal_reason"`

	// EvidenceHashes are hashes of supporting evidence documents
	EvidenceHashes []string `json:"evidence_hashes,omitempty"`

	// SubmittedAt is the block height when the appeal was submitted
	SubmittedAt int64 `json:"submitted_at"`

	// SubmittedAtTime is the timestamp when the appeal was submitted
	SubmittedAtTime int64 `json:"submitted_at_time"`

	// Status is the current status of the appeal
	Status AppealStatus `json:"status"`

	// ReviewerAddress is the address of the arbitrator reviewing the appeal
	ReviewerAddress string `json:"reviewer_address,omitempty"`

	// ClaimedAt is the block height when the appeal was claimed for review
	ClaimedAt int64 `json:"claimed_at,omitempty"`

	// ReviewedAt is the block height when the appeal was resolved
	ReviewedAt int64 `json:"reviewed_at,omitempty"`

	// ReviewedAtTime is the timestamp when the appeal was resolved
	ReviewedAtTime int64 `json:"reviewed_at_time,omitempty"`

	// ResolutionReason is the arbitrator's explanation for the decision
	ResolutionReason string `json:"resolution_reason,omitempty"`

	// ScoreAdjustment is the adjustment to the verification score (can be positive or negative)
	ScoreAdjustment int32 `json:"score_adjustment,omitempty"`

	// AppealNumber tracks which appeal this is for the scope (1st, 2nd, 3rd)
	AppealNumber uint32 `json:"appeal_number"`
}

// NewAppealRecord creates a new AppealRecord
func NewAppealRecord(
	appealID string,
	accountAddress string,
	scopeID string,
	originalStatus string,
	originalScore uint32,
	appealReason string,
	evidenceHashes []string,
	submittedAt int64,
	submittedAtTime time.Time,
	appealNumber uint32,
) *AppealRecord {
	return &AppealRecord{
		AppealID:        appealID,
		AccountAddress:  accountAddress,
		ScopeID:         scopeID,
		OriginalStatus:  originalStatus,
		OriginalScore:   originalScore,
		AppealReason:    appealReason,
		EvidenceHashes:  evidenceHashes,
		SubmittedAt:     submittedAt,
		SubmittedAtTime: submittedAtTime.Unix(),
		Status:          AppealStatusPending,
		AppealNumber:    appealNumber,
	}
}

// GenerateAppealID generates a unique appeal ID from components
func GenerateAppealID(accountAddress string, scopeID string, blockHeight int64) string {
	data := fmt.Sprintf("%s:%s:%d", accountAddress, scopeID, blockHeight)
	hash := sha256.Sum256([]byte(data))
	return "appeal_" + hex.EncodeToString(hash[:8])
}

// Validate validates the appeal record
func (a *AppealRecord) Validate() error {
	if a.AppealID == "" {
		return ErrInvalidAppealRecord.Wrap("appeal_id cannot be empty")
	}

	if a.AccountAddress == "" {
		return ErrInvalidAppealRecord.Wrap("account_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(a.AccountAddress); err != nil {
		return ErrInvalidAppealRecord.Wrapf("invalid account_address: %v", err)
	}

	if a.ScopeID == "" {
		return ErrInvalidAppealRecord.Wrap("scope_id cannot be empty")
	}

	if len(a.AppealReason) < int(DefaultMinAppealReasonLength) {
		return ErrInvalidAppealReason.Wrapf("appeal reason must be at least %d characters", DefaultMinAppealReasonLength)
	}

	if len(a.AppealReason) > MaxAppealReasonLength {
		return ErrInvalidAppealReason.Wrapf("appeal reason cannot exceed %d characters", MaxAppealReasonLength)
	}

	if len(a.EvidenceHashes) > MaxEvidenceHashes {
		return ErrInvalidAppealRecord.Wrapf("cannot exceed %d evidence hashes", MaxEvidenceHashes)
	}

	if a.SubmittedAt <= 0 {
		return ErrInvalidAppealRecord.Wrap("submitted_at must be positive")
	}

	return nil
}

// SetReviewing transitions the appeal to reviewing status
func (a *AppealRecord) SetReviewing(reviewerAddress string, claimedAt int64) error {
	if a.Status != AppealStatusPending {
		return ErrAppealNotPending.Wrapf("cannot claim appeal in status %s", a.Status.String())
	}
	a.Status = AppealStatusReviewing
	a.ReviewerAddress = reviewerAddress
	a.ClaimedAt = claimedAt
	return nil
}

// Resolve resolves the appeal with the given resolution
func (a *AppealRecord) Resolve(
	resolution AppealStatus,
	reason string,
	scoreAdjustment int32,
	reviewedAt int64,
	reviewedAtTime time.Time,
) error {
	if !a.Status.IsActive() {
		return ErrAppealNotPending.Wrapf("cannot resolve appeal in status %s", a.Status.String())
	}

	if resolution != AppealStatusApproved && resolution != AppealStatusRejected {
		return ErrInvalidAppealResolution.Wrapf("invalid resolution status: %s", resolution.String())
	}

	a.Status = resolution
	a.ResolutionReason = reason
	a.ScoreAdjustment = scoreAdjustment
	a.ReviewedAt = reviewedAt
	a.ReviewedAtTime = reviewedAtTime.Unix()
	return nil
}

// Withdraw withdraws the appeal
func (a *AppealRecord) Withdraw() error {
	if !a.Status.IsActive() {
		return ErrAppealNotPending.Wrapf("cannot withdraw appeal in status %s", a.Status.String())
	}
	a.Status = AppealStatusWithdrawn
	return nil
}

// ============================================================================
// Appeal Params
// ============================================================================

// AppealParams defines the parameters for the appeal system
type AppealParams struct {
	// AppealWindowBlocks is how long after rejection can user appeal
	AppealWindowBlocks int64 `json:"appeal_window_blocks"`

	// MaxAppealsPerScope is the maximum appeals allowed per scope
	MaxAppealsPerScope uint32 `json:"max_appeals_per_scope"`

	// MinAppealReasonLength is the minimum characters for appeal reason
	MinAppealReasonLength uint32 `json:"min_appeal_reason_length"`

	// ReviewTimeoutBlocks is how long an appeal can stay in reviewing status
	ReviewTimeoutBlocks int64 `json:"review_timeout_blocks"`

	// Enabled indicates whether the appeal system is active
	Enabled bool `json:"enabled"`

	// RequireEscrowDeposit indicates whether appeals require a deposit
	RequireEscrowDeposit bool `json:"require_escrow_deposit"`

	// EscrowDepositAmount is the deposit amount required (in base units)
	EscrowDepositAmount int64 `json:"escrow_deposit_amount,omitempty"`
}

// DefaultAppealParams returns the default appeal parameters
func DefaultAppealParams() AppealParams {
	return AppealParams{
		AppealWindowBlocks:    DefaultAppealWindowBlocks,
		MaxAppealsPerScope:    DefaultMaxAppealsPerScope,
		MinAppealReasonLength: DefaultMinAppealReasonLength,
		ReviewTimeoutBlocks:   DefaultAppealReviewTimeoutBlocks,
		Enabled:               true,
		RequireEscrowDeposit:  false,
		EscrowDepositAmount:   0,
	}
}

// Validate validates the appeal parameters
func (p AppealParams) Validate() error {
	if p.AppealWindowBlocks <= 0 && p.Enabled {
		return ErrInvalidParams.Wrap("appeal_window_blocks must be positive when enabled")
	}

	if p.MaxAppealsPerScope == 0 && p.Enabled {
		return ErrInvalidParams.Wrap("max_appeals_per_scope cannot be zero when enabled")
	}

	if p.MinAppealReasonLength == 0 && p.Enabled {
		return ErrInvalidParams.Wrap("min_appeal_reason_length cannot be zero when enabled")
	}

	if p.ReviewTimeoutBlocks <= 0 && p.Enabled {
		return ErrInvalidParams.Wrap("review_timeout_blocks must be positive when enabled")
	}

	if p.RequireEscrowDeposit && p.EscrowDepositAmount <= 0 {
		return ErrInvalidParams.Wrap("escrow_deposit_amount must be positive when require_escrow_deposit is true")
	}

	return nil
}

// IsWithinAppealWindow checks if the current block is within the appeal window
func (p AppealParams) IsWithinAppealWindow(rejectionBlock, currentBlock int64) bool {
	return currentBlock <= rejectionBlock+p.AppealWindowBlocks
}

// ============================================================================
// Appeal Summary
// ============================================================================

// AppealSummary provides a summary of appeals for an account
type AppealSummary struct {
	TotalAppeals     uint32 `json:"total_appeals"`
	PendingAppeals   uint32 `json:"pending_appeals"`
	ApprovedAppeals  uint32 `json:"approved_appeals"`
	RejectedAppeals  uint32 `json:"rejected_appeals"`
	WithdrawnAppeals uint32 `json:"withdrawn_appeals"`
}

// NewAppealSummary creates a new appeal summary
func NewAppealSummary() *AppealSummary {
	return &AppealSummary{}
}

// AddAppeal increments the appropriate counter based on appeal status
func (s *AppealSummary) AddAppeal(status AppealStatus) {
	s.TotalAppeals++
	switch status {
	case AppealStatusPending, AppealStatusReviewing:
		s.PendingAppeals++
	case AppealStatusApproved:
		s.ApprovedAppeals++
	case AppealStatusRejected:
		s.RejectedAppeals++
	case AppealStatusWithdrawn:
		s.WithdrawnAppeals++
	}
}
