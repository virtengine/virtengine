// Package marketplace provides types for the marketplace on-chain module.
//
// MARKET-HW-SAFEGUARD-1: the settlement gate. This is the single place that
// decides whether an escrowed milestone-backed order may release funds.
//
// The gate deliberately reads only milestone evidence. The VEID signal is
// advisory context carried alongside the order and MUST NOT be read here;
// `TestSettlementGateIgnoresVEIDSignal` pins that invariant.
package marketplace

import "fmt"

// MilestoneEvidence reports the realised evidence for one milestone trigger.
type MilestoneEvidence struct {
	// MilestoneID links the evidence to a schedule entry.
	MilestoneID string `json:"milestone_id"`

	// Satisfied reports whether the trigger evidence was accepted.
	Satisfied bool `json:"satisfied"`
}

// SettlementDecision is the outcome of the milestone settlement gate.
type SettlementDecision struct {
	// Release reports whether the milestone may release funds.
	Release bool `json:"release"`

	// MilestoneID is the milestone evaluated.
	MilestoneID string `json:"milestone_id"`

	// Reason explains the decision in human-readable terms.
	Reason string `json:"reason"`
}

// VEIDIsDeterminative reports whether a VEID signal may, on its own, authorise
// or block settlement. It is always false by design.
func VEIDIsDeterminative() bool { return false }

// EvaluateMilestoneSettlement decides whether a single milestone may release.
//
// Signature note: no VEID parameter is accepted. That is the guarantee, not an
// omission — a caller cannot express a VEID-driven settlement here.
func EvaluateMilestoneSettlement(
	set MilestoneSet,
	states []MilestoneState,
	evidence MilestoneEvidence,
) (SettlementDecision, error) {
	decision := SettlementDecision{MilestoneID: evidence.MilestoneID}

	milestone, found := findMilestone(set, evidence.MilestoneID)
	if !found {
		return decision, fmt.Errorf("milestone %q is not part of the order schedule", evidence.MilestoneID)
	}

	state, found := FindMilestoneState(states, evidence.MilestoneID)
	if !found {
		return decision, fmt.Errorf("milestone %q has no runtime state", evidence.MilestoneID)
	}

	// A held milestone never releases, regardless of evidence.
	if state.Status == MilestoneStatusHeld {
		decision.Release = false
		decision.Reason = fmt.Sprintf("milestone %s is held by dispute %s", state.MilestoneID, state.DisputeID)
		return decision, nil
	}

	// Already terminal: report the existing outcome rather than double-releasing.
	if state.Status == MilestoneStatusReleased {
		decision.Release = false
		decision.Reason = fmt.Sprintf("milestone %s has already released", state.MilestoneID)
		return decision, nil
	}
	if state.Status == MilestoneStatusRefunded {
		decision.Release = false
		decision.Reason = fmt.Sprintf("milestone %s was refunded", state.MilestoneID)
		return decision, nil
	}

	if !evidence.Satisfied {
		decision.Release = false
		decision.Reason = fmt.Sprintf(
			"trigger %q not satisfied: %s",
			milestone.Trigger,
			MilestoneTriggerRequirementText(milestone.Trigger),
		)
		return decision, nil
	}

	decision.Release = true
	decision.Reason = fmt.Sprintf("trigger %q satisfied", milestone.Trigger)
	return decision, nil
}

// findMilestone locates a milestone in a schedule by id.
func findMilestone(set MilestoneSet, milestoneID string) (Milestone, bool) {
	for _, m := range set {
		if m.ID == milestoneID {
			return m, true
		}
	}
	return Milestone{}, false
}

// MilestoneOrderSummary reports stage progress for an order, so a UI or dispute
// view can show partial completion without reading raw state.
type MilestoneOrderSummary struct {
	// Total is the number of milestones in the schedule.
	Total int `json:"total"`

	// Released is the number of released milestones.
	Released int `json:"released"`

	// Held is the number of milestones currently held.
	Held int `json:"held"`

	// Pending is the number of milestones still awaiting evidence.
	Pending int `json:"pending"`

	// Refunded is the number of refunded milestones.
	Refunded int `json:"refunded"`

	// FullyReleased reports whether every milestone released.
	FullyReleased bool `json:"fully_released"`
}

// SummarizeMilestones counts states by status in a deterministic pass.
func SummarizeMilestones(set MilestoneSet, states []MilestoneState) MilestoneOrderSummary {
	summary := MilestoneOrderSummary{Total: len(set)}
	for _, state := range states {
		switch state.Status {
		case MilestoneStatusReleased:
			summary.Released++
		case MilestoneStatusHeld:
			summary.Held++
		case MilestoneStatusRefunded:
			summary.Refunded++
		default:
			summary.Pending++
		}
	}
	summary.FullyReleased = summary.Total > 0 && summary.Released == summary.Total
	return summary
}
