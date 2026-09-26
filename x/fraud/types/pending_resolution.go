// Package types contains types for the Fraud module.
//
// A suspension or termination resolution removes an account's access
// network-wide and is effectively permanent in the eyes of the affected party.
// Neither may be driven by one moderator alone, so those resolutions are
// recorded as a PendingResolution and only take effect once a second, distinct
// moderator-or-above identity confirms them within a bounded window.
//
// The record is stored as JSON (the pattern this module already uses for
// FraudReport), so no protobuf change is required to land it.
package types

import "fmt"

// Pending resolution review windows.
const (
	// MinPendingResolutionWindowSeconds is the shortest review window (1h).
	MinPendingResolutionWindowSeconds int64 = 60 * 60

	// DefaultPendingResolutionWindowSeconds is the default review window (72h).
	DefaultPendingResolutionWindowSeconds int64 = 72 * 60 * 60

	// MaxPendingResolutionWindowSeconds is the longest review window (7 days).
	MaxPendingResolutionWindowSeconds int64 = 7 * 24 * 60 * 60
)

// PendingResolution is a proposed suspension or termination awaiting review by
// a second, distinct moderator.
//
// It deliberately has no effect on the report or the subject account: that is
// the whole point. If the window lapses before review, the proposal lapses with
// it and the report is left as it was, free to be re-proposed.
type PendingResolution struct {
	// ReportID is the report the resolution applies to.
	ReportID string `json:"report_id"`

	// ProposedBy is the moderator who proposed the resolution.
	ProposedBy string `json:"proposed_by"`

	// Resolution is the proposed (not yet effective) resolution.
	Resolution ResolutionType `json:"resolution"`

	// Notes is the proposing moderator's justification.
	Notes string `json:"notes,omitempty"`

	// ProposedAt is the block time (Unix seconds) of the proposal.
	ProposedAt int64 `json:"proposed_at"`

	// ExpiresAt is the block time (Unix seconds) the proposal lapses.
	ExpiresAt int64 `json:"expires_at"`

	// BlockHeight is the height at which the proposal was recorded.
	BlockHeight int64 `json:"block_height"`
}

// NewPendingResolution builds a pending resolution with a bounded window.
func NewPendingResolution(
	reportID string,
	proposedBy string,
	resolution ResolutionType,
	notes string,
	blockTime int64,
	blockHeight int64,
) PendingResolution {
	return PendingResolution{
		ReportID:    reportID,
		ProposedBy:  proposedBy,
		Resolution:  resolution,
		Notes:       notes,
		ProposedAt:  blockTime,
		ExpiresAt:   blockTime + DefaultPendingResolutionWindowSeconds,
		BlockHeight: blockHeight,
	}
}

// Validate checks the pending resolution's internal consistency.
func (p PendingResolution) Validate() error {
	if p.ReportID == "" {
		return fmt.Errorf("report_id is required")
	}
	if p.ProposedBy == "" {
		return fmt.Errorf("proposed_by is required")
	}
	if !p.Resolution.RequiresSecondReviewer() {
		return fmt.Errorf("resolution %s does not require a second reviewer", p.Resolution)
	}
	if p.ProposedAt <= 0 {
		return fmt.Errorf("proposed_at is required")
	}
	if p.ExpiresAt <= p.ProposedAt {
		return fmt.Errorf("expires_at must be after proposed_at")
	}
	return nil
}

// IsExpired reports whether the review window has closed at the given time.
func (p PendingResolution) IsExpired(now int64) bool {
	return p.ExpiresAt > 0 && p.ExpiresAt <= now
}

// IsConfirmedBy reports whether the given address is a distinct, valid
// confirming reviewer for this proposal.
func (p PendingResolution) IsConfirmedBy(reviewer string) bool {
	return reviewer != "" && reviewer != p.ProposedBy
}
