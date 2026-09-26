// Package marketplace provides types for the marketplace on-chain module.
//
// MARKET-HW-SAFEGUARD-1: transaction safeguards for hardware and compute orders.
// VEID establishes identity, not that hardware exists, is owned by the seller, meets
// its advertised specification, or will be delivered. This file defines the staged,
// escrow-backed milestone schedule that pairs an advisory identity signal with
// delivery-shaped payment terms.
//
// Determinism: every helper here is pure integer arithmetic over ordered slices.
// No map iteration feeds an amount, a digest, or an ordering.
package marketplace

import (
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MilestoneTrigger identifies the evidence event that unlocks a milestone.
type MilestoneTrigger string

const (
	// MilestoneTriggerCapacityVerified unlocks when the seller's capacity evidence
	// has been verified (capacity report / hashed ownership attestation).
	MilestoneTriggerCapacityVerified MilestoneTrigger = "capacity_verified"

	// MilestoneTriggerDeliveryAttested unlocks when delivery or measured
	// performance has been attested by the buyer.
	MilestoneTriggerDeliveryAttested MilestoneTrigger = "delivery_performance_attested"
)

// IsValidMilestoneTrigger reports whether the trigger is known.
func IsValidMilestoneTrigger(trigger MilestoneTrigger) bool {
	switch trigger {
	case MilestoneTriggerCapacityVerified, MilestoneTriggerDeliveryAttested:
		return true
	default:
		return false
	}
}

// Milestone is a single staged release step in an escrow-backed order.
//
// The schedule is a *plan*; its runtime state (held/released) lives on the
// settlement escrow record so that dispute handling can act on one milestone
// without resolving the whole order.
type Milestone struct {
	// ID is the stable milestone identifier, unique within a schedule.
	ID string `json:"id"`

	// Description is the human-readable term shown on the listing.
	Description string `json:"description"`

	// ShareBps is this milestone's share of the order total in basis points.
	// Every milestone in a valid schedule sums to exactly 10000.
	ShareBps uint32 `json:"share_bps"`

	// Trigger is the evidence event that unlocks release.
	Trigger MilestoneTrigger `json:"trigger"`

	// Sequence is the deterministic 1-based release order.
	Sequence uint32 `json:"sequence"`
}

// Validate validates a single milestone.
func (m Milestone) Validate() error {
	if strings.TrimSpace(m.ID) == "" {
		return fmt.Errorf("milestone id is required")
	}
	if strings.TrimSpace(m.Description) == "" {
		return fmt.Errorf("milestone %s: description is required", m.ID)
	}
	if m.ShareBps == 0 {
		return fmt.Errorf("milestone %s: share_bps must be positive", m.ID)
	}
	if m.ShareBps > BpsDenominator {
		return fmt.Errorf("milestone %s: share_bps %d exceeds %d", m.ID, m.ShareBps, BpsDenominator)
	}
	if !IsValidMilestoneTrigger(m.Trigger) {
		return fmt.Errorf("milestone %s: invalid trigger %q", m.ID, m.Trigger)
	}
	if m.Sequence == 0 {
		return fmt.Errorf("milestone %s: sequence must be positive", m.ID)
	}
	return nil
}

// BpsDenominator is 100% expressed in basis points.
const BpsDenominator uint32 = 10000

// MilestoneSet is an ordered milestone schedule.
type MilestoneSet []Milestone

// DefaultMilestoneSet returns the protocol default schedule: a quarter released
// once capacity is verified, the remainder once delivery or performance is
// attested. Defaults live in params and a listing may override them.
func DefaultMilestoneSet() MilestoneSet {
	return MilestoneSet{
		{
			ID:          "capacity_verified",
			Description: "Capacity/ownership evidence verified",
			ShareBps:    2500,
			Trigger:     MilestoneTriggerCapacityVerified,
			Sequence:    1,
		},
		{
			ID:          "delivery_attested",
			Description: "Delivery or measured performance attested",
			ShareBps:    7500,
			Trigger:     MilestoneTriggerDeliveryAttested,
			Sequence:    2,
		},
	}
}

// Validate validates the schedule: non-empty, unique ids, unique contiguous
// sequences, every share positive, and shares summing to exactly 10000.
func (s MilestoneSet) Validate() error {
	if len(s) == 0 {
		return fmt.Errorf("milestone schedule must contain at least one milestone")
	}

	seenIDs := make(map[string]bool, len(s))
	seenSeq := make(map[uint32]bool, len(s))
	total := uint32(0)

	for _, m := range s {
		if err := m.Validate(); err != nil {
			return err
		}
		if seenIDs[m.ID] {
			return fmt.Errorf("duplicate milestone id %q", m.ID)
		}
		seenIDs[m.ID] = true
		if seenSeq[m.Sequence] {
			return fmt.Errorf("duplicate milestone sequence %d", m.Sequence)
		}
		seenSeq[m.Sequence] = true
		total += m.ShareBps
		if total > BpsDenominator {
			return fmt.Errorf("milestone shares exceed %d basis points", BpsDenominator)
		}
	}

	if total != BpsDenominator {
		return fmt.Errorf("milestone shares must sum to %d basis points, got %d", BpsDenominator, total)
	}

	for i := uint32(1); i <= uint32(len(s)); i++ { //nolint:gosec // G115: len(s) is bounded by the milestone count validated above
		if !seenSeq[i] {
			return fmt.Errorf("milestone sequence %d is missing; sequences must be contiguous from 1", i)
		}
	}

	return nil
}

// Ordered returns the schedule sorted by sequence, without mutating the receiver.
func (s MilestoneSet) Ordered() MilestoneSet {
	out := make(MilestoneSet, len(s))
	copy(out, s)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Sequence < out[j].Sequence })
	return out
}

// Clone returns a deep copy of the schedule.
func (s MilestoneSet) Clone() MilestoneSet {
	if len(s) == 0 {
		return nil
	}
	out := make(MilestoneSet, len(s))
	copy(out, s)
	return out
}

// Total releases exactly `total`, splitting each denomination by basis points
// with the integer remainder assigned to the final milestone. The result is
// aligned to Ordered() and is guaranteed to sum back to `total`, so no value is
// created or destroyed by the split (conservation is asserted by the tests).
func (s MilestoneSet) Total(total sdk.Coins) []sdk.Coins {
	ordered := s.Ordered()
	amounts := make([]sdk.Coins, len(ordered))

	allocated := sdk.NewCoins()
	for i, m := range ordered {
		amounts[i] = scaleCoins(total, m.ShareBps)
		allocated = allocated.Add(amounts[i]...)
	}

	// Assign any truncation remainder to the final milestone so the split is exact.
	remainder, hasNeg := total.SafeSub(allocated...)
	if !hasNeg && !remainder.IsZero() {
		last := len(amounts) - 1
		amounts[last] = amounts[last].Add(remainder...).Sort()
	}

	return amounts
}

// scaleCoins returns floor(total * bps / 10000) per denomination.
func scaleCoins(total sdk.Coins, bps uint32) sdk.Coins {
	out := sdk.NewCoins()
	denom := sdkmath.NewInt(int64(BpsDenominator))
	factor := sdkmath.NewIntFromUint64(uint64(bps))

	for _, coin := range total {
		part := coin.Amount.Mul(factor).Quo(denom)
		if part.IsPositive() {
			out = out.Add(sdk.NewCoin(coin.Denom, part))
		}
	}

	return out
}

// MilestonePolicy is the governance-controlled default staged-payment policy.
type MilestonePolicy struct {
	// Enabled enables escrow-backed milestone release for qualifying orders.
	Enabled bool `json:"enabled"`

	// DefaultMilestones is the protocol default schedule. A listing may override it.
	DefaultMilestones MilestoneSet `json:"default_milestones"`

	// MaxHoldDurationSeconds bounds how long a disputed milestone may be held
	// before governance/arbitration must resolve it.
	MaxHoldDurationSeconds uint64 `json:"max_hold_duration_seconds"`

	// RequireAttestationForHardware requires a capacity or ownership attestation
	// to be attached before an offering may accept hardware/compute orders.
	RequireAttestationForHardware bool `json:"require_attestation_for_hardware"`
}

// DefaultMilestonePolicy returns the default milestone policy.
func DefaultMilestonePolicy() MilestonePolicy {
	return MilestonePolicy{
		Enabled:                       true,
		DefaultMilestones:             DefaultMilestoneSet(),
		MaxHoldDurationSeconds:        1209600, // 14 days
		RequireAttestationForHardware: true,
	}
}

// Validate validates the milestone policy.
func (p *MilestonePolicy) Validate() error {
	if !p.Enabled {
		return nil
	}
	if err := p.DefaultMilestones.Validate(); err != nil {
		return fmt.Errorf("invalid default_milestones: %w", err)
	}
	if p.MaxHoldDurationSeconds == 0 {
		return fmt.Errorf("max_hold_duration_seconds must be positive when milestones are enabled")
	}
	return nil
}

// ResolveMilestoneSet applies a per-listing override on top of the policy
// default. An empty override yields the policy default.
func (p *MilestonePolicy) ResolveMilestoneSet(override MilestoneSet) (MilestoneSet, error) {
	if len(override) == 0 {
		return p.DefaultMilestones.Clone(), nil
	}
	if err := override.Validate(); err != nil {
		return nil, fmt.Errorf("invalid milestone override: %w", err)
	}
	return override.Clone(), nil
}

// MilestoneTriggerRequirementText returns the "what to supply" text surfaced to
// counterparties for a trigger, in the same framing as the identity display.
func MilestoneTriggerRequirementText(trigger MilestoneTrigger) string {
	switch trigger {
	case MilestoneTriggerCapacityVerified:
		return "Seller must supply capacity or ownership evidence (hashed capacity report or ownership attestation) before this milestone releases."
	case MilestoneTriggerDeliveryAttested:
		return "Buyer attests delivery or measured performance before this milestone releases."
	default:
		return "Unknown milestone trigger."
	}
}
