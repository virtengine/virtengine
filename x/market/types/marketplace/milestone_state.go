// Package marketplace provides types for the marketplace on-chain module.
//
// MARKET-HW-SAFEGUARD-1: per-order milestone runtime state. A milestone can be
// held or released individually, so an order-linked dispute can act on one
// milestone without resolving the whole order.
package marketplace

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// MilestoneStatus is the release status of a single milestone.
type MilestoneStatus string

const (
	// MilestoneStatusPending means the trigger has not been satisfied yet.
	MilestoneStatusPending MilestoneStatus = "pending"

	// MilestoneStatusHeld means a dispute holds this milestone; the rest of the
	// order continues independently.
	MilestoneStatusHeld MilestoneStatus = "held"

	// MilestoneStatusReleased means the milestone amount has been released.
	MilestoneStatusReleased MilestoneStatus = "released"

	// MilestoneStatusRefunded means the milestone amount was returned to the buyer.
	MilestoneStatusRefunded MilestoneStatus = "refunded"
)

// IsValidMilestoneStatus reports whether the status is known.
func IsValidMilestoneStatus(status MilestoneStatus) bool {
	switch status {
	case MilestoneStatusPending, MilestoneStatusHeld, MilestoneStatusReleased, MilestoneStatusRefunded:
		return true
	default:
		return false
	}
}

// MilestoneState is the runtime state of one milestone on an order.
type MilestoneState struct {
	// MilestoneID links back to the Milestone in the order's schedule.
	MilestoneID string `json:"milestone_id"`

	// Status is the current release status.
	Status MilestoneStatus `json:"status"`

	// Amount is the resolved coin amount for this milestone.
	Amount string `json:"amount"`

	// HoldReason records why a dispute held this milestone.
	HoldReason string `json:"hold_reason,omitempty"`

	// DisputeID links the hold to its order-linked dispute, when held.
	DisputeID string `json:"dispute_id,omitempty"`

	// HeldAt is when the milestone was held.
	HeldAt *time.Time `json:"held_at,omitempty"`

	// ResolvedAt is when the milestone left Held.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// SatisfiedAt is when the trigger evidence was accepted.
	SatisfiedAt *time.Time `json:"satisfied_at,omitempty"`

	// ReleasedAt is when funds moved.
	ReleasedAt *time.Time `json:"released_at,omitempty"`
}

// Validate validates a milestone state.
func (s *MilestoneState) Validate() error {
	if strings.TrimSpace(s.MilestoneID) == "" {
		return fmt.Errorf("milestone state requires a milestone_id")
	}
	if !IsValidMilestoneStatus(s.Status) {
		return fmt.Errorf("milestone %s: invalid status %q", s.MilestoneID, s.Status)
	}
	if strings.TrimSpace(s.Amount) == "" {
		return fmt.Errorf("milestone %s: amount is required", s.MilestoneID)
	}
	return nil
}

// FindMilestoneState returns the state for a milestone id.
func FindMilestoneState(states []MilestoneState, milestoneID string) (*MilestoneState, bool) {
	for i := range states {
		if states[i].MilestoneID == milestoneID {
			return &states[i], true
		}
	}
	return nil, false
}

// SortMilestoneStates orders states deterministically by milestone id so that
// serialized state and digests never depend on map iteration order.
func SortMilestoneStates(states []MilestoneState) {
	sort.SliceStable(states, func(i, j int) bool { return states[i].MilestoneID < states[j].MilestoneID })
}

// MilestoneDispute is an order-linked dispute that can hold or release a single
// milestone without resolving the whole order.
type MilestoneDispute struct {
	// DisputeID is the dispute identifier.
	DisputeID string `json:"dispute_id"`

	// OrderID is the order the dispute is linked to.
	OrderID string `json:"order_id"`

	// MilestoneID is the single milestone under dispute.
	MilestoneID string `json:"milestone_id"`

	// Reason is the recorded dispute reason.
	Reason string `json:"reason"`

	// OpenedBy is the party that opened the dispute.
	OpenedBy string `json:"opened_by"`

	// VEIDSignal is the advisory identity signal attached to the dispute record.
	VEIDSignal *VEIDFraudSignal `json:"veid_signal,omitempty"`

	// Open is true while the milestone is held.
	Open bool `json:"open"`

	// OpenedAt is the block time the dispute was opened.
	OpenedAt time.Time `json:"opened_at"`

	// ResolvedAt is when the dispute was closed.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// Resolution records the outcome.
	Resolution string `json:"resolution,omitempty"`
}

// Validate validates the milestone dispute.
func (d *MilestoneDispute) Validate() error {
	if strings.TrimSpace(d.DisputeID) == "" {
		return fmt.Errorf("dispute_id is required")
	}
	if strings.TrimSpace(d.OrderID) == "" {
		return fmt.Errorf("order_id is required")
	}
	if strings.TrimSpace(d.MilestoneID) == "" {
		return fmt.Errorf("milestone_id is required")
	}
	if strings.TrimSpace(d.Reason) == "" {
		return fmt.Errorf("dispute reason is required")
	}
	if d.VEIDSignal != nil {
		if err := d.VEIDSignal.Validate(); err != nil {
			return fmt.Errorf("invalid veid signal: %w", err)
		}
	}
	return nil
}
