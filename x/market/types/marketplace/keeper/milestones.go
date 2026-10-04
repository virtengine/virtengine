// Package keeper implements the market module keeper.
//
// MARKET-HW-SAFEGUARD-1: staged, escrow-backed payment for hardware and compute
// orders, an advisory VEID fraud signal, hashed capacity/ownership attestations,
// and an order-linked dispute path that can hold a single milestone.
//
// Determinism (consensus-critical): all ordering here is over ordered slices or
// explicitly sorted keys. Block time comes from ctx.BlockTime(). No map iteration
// feeds an amount, an ordering, or a stored digest.
package keeper

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// MilestoneEscrowKeeper is the narrow settlement/escrow boundary used to move
// funds for one milestone. Keeping it this small means the marketplace module
// cannot reach into escrow beyond a single, auditable release call.
type MilestoneEscrowKeeper interface {
	// ReleaseMilestoneEscrow releases the funds held for one milestone of an
	// order. Implementations must be idempotent per (orderID, milestoneID).
	ReleaseMilestoneEscrow(ctx sdk.Context, orderID, milestoneID string, amount sdk.Coins, reason string) error

	// RefundMilestoneEscrow returns the funds held for one milestone to the buyer.
	RefundMilestoneEscrow(ctx sdk.Context, orderID, milestoneID string, amount sdk.Coins, reason string) error
}

// SetMilestoneEscrowKeeper wires the escrow boundary after construction, so the
// keeper signature stays stable for existing callers.
func (k *Keeper) SetMilestoneEscrowKeeper(escrow MilestoneEscrowKeeper) {
	k.milestoneEscrow = escrow
}

// ResolveOrderMilestones returns the milestone schedule for an order against an
// offering: the per-listing override when present, otherwise the params default.
func (k Keeper) ResolveOrderMilestones(ctx sdk.Context, offering *marketplace.Offering) (marketplace.MilestoneSet, error) {
	params := k.GetParams(ctx)
	if !params.MilestonePolicy.Enabled {
		return nil, nil
	}

	var override marketplace.MilestoneSet
	if offering != nil {
		override = offering.MilestoneOverride
	}

	set, err := params.MilestonePolicy.ResolveMilestoneSet(override)
	if err != nil {
		return nil, err
	}
	return set, nil
}

// RequireAttestationForOrder enforces the params-level rule that hardware and
// compute listings must carry capacity/ownership evidence before accepting orders.
func (k Keeper) RequireAttestationForOrder(ctx sdk.Context, offering *marketplace.Offering) error {
	params := k.GetParams(ctx)
	if !params.MilestonePolicy.Enabled || !params.MilestonePolicy.RequireAttestationForHardware {
		return nil
	}
	if offering == nil {
		return nil
	}
	if !isHardwareOrComputeCategory(offering.Category) {
		return nil
	}
	if offering.Attestation == nil {
		return marketplace.ErrAttestationRequired.Wrapf(
			"offering %s is a hardware/compute listing and must carry a capacity or ownership attestation", offering.ID.String(),
		)
	}
	if err := offering.Attestation.Validate(); err != nil {
		return marketplace.ErrAttestationRequired.Wrapf("invalid attestation: %s", err.Error())
	}
	return nil
}

// isHardwareOrComputeCategory reports whether a category is in scope for the
// hardware/compute safeguard.
func isHardwareOrComputeCategory(category marketplace.OfferingCategory) bool {
	switch category {
	case marketplace.OfferingCategoryCompute,
		marketplace.OfferingCategoryHPC,
		marketplace.OfferingCategoryGPU,
		marketplace.OfferingCategoryStorage:
		return true
	default:
		return false
	}
}

// RecordOrderVEIDSignal attaches the advisory VEID signal to an order and
// resolves its milestone schedule. The signal is recorded, never enforced.
func (k Keeper) RecordOrderVEIDSignal(
	ctx sdk.Context,
	orderID marketplace.OrderID,
	signal *marketplace.VEIDFraudSignal,
) error {
	order, found := k.GetOrder(ctx, orderID)
	if !found {
		return marketplace.ErrOrderNotFound
	}

	if signal != nil {
		if err := signal.Validate(); err != nil {
			return err
		}
		// Guard the invariant at the write boundary: a VEID signal may never be
		// recorded as determinative.
		if marketplace.VEIDIsDeterminative() || !signal.Advisory || !signal.NotDeterminative {
			return fmt.Errorf("veid signal must remain advisory and non-determinative")
		}
	}

	order.VEIDSignal = signal
	return k.UpdateOrder(ctx, order)
}

// InitializeOrderMilestones resolves and stores the milestone schedule plus the
// initial per-milestone runtime state for an order.
func (k Keeper) InitializeOrderMilestones(ctx sdk.Context, orderID marketplace.OrderID, total sdk.Coins) error {
	order, found := k.GetOrder(ctx, orderID)
	if !found {
		return marketplace.ErrOrderNotFound
	}

	offering, found := k.GetOffering(ctx, order.OfferingID)
	if !found {
		return marketplace.ErrOfferingNotFound
	}

	// Hardware/compute listings must carry capacity or ownership evidence before
	// staged payment is initialised for them (MARKET-HW-SAFEGUARD-1).
	if err := k.RequireAttestationForOrder(ctx, offering); err != nil {
		return err
	}

	set, err := k.ResolveOrderMilestones(ctx, offering)
	if err != nil {
		return err
	}
	if len(set) == 0 {
		return nil
	}

	order.Milestones = set
	order.MilestoneStates = buildMilestoneStates(set, total)
	return k.UpdateOrder(ctx, order)
}

// buildMilestoneStates expands a schedule over an order total into pending state.
func buildMilestoneStates(set marketplace.MilestoneSet, total sdk.Coins) []marketplace.MilestoneState {
	ordered := set.Ordered()
	amounts := set.Total(total)

	states := make([]marketplace.MilestoneState, 0, len(ordered))
	for i, m := range ordered {
		amount := ""
		if i < len(amounts) {
			amount = amounts[i].String()
		}
		states = append(states, marketplace.MilestoneState{
			MilestoneID: m.ID,
			Status:      marketplace.MilestoneStatusPending,
			Amount:      amount,
		})
	}
	// Deterministic order, independent of schedule declaration order.
	sort.SliceStable(states, func(i, j int) bool { return states[i].MilestoneID < states[j].MilestoneID })
	return states
}

// ReleaseOrderMilestone evaluates the milestone settlement gate and, when it
// permits, releases that single milestone's escrow.
//
// The gate reads milestone evidence only. The VEID signal is never consulted
// here — see marketplace.EvaluateMilestoneSettlement.
func (k Keeper) ReleaseOrderMilestone(
	ctx sdk.Context,
	orderID marketplace.OrderID,
	milestoneID string,
	evidence marketplace.MilestoneEvidence,
) (marketplace.SettlementDecision, error) {
	order, found := k.GetOrder(ctx, orderID)
	if !found {
		return marketplace.SettlementDecision{}, marketplace.ErrOrderNotFound
	}

	evidence.MilestoneID = milestoneID
	decision, err := marketplace.EvaluateMilestoneSettlement(order.Milestones, order.MilestoneStates, evidence)
	if err != nil {
		return decision, err
	}
	if !decision.Release {
		return decision, nil
	}

	state, found := marketplace.FindMilestoneState(order.MilestoneStates, milestoneID)
	if !found {
		return decision, fmt.Errorf("milestone %q has no runtime state", milestoneID)
	}

	amount, err := sdk.ParseCoinsNormalized(state.Amount)
	if err != nil {
		return decision, fmt.Errorf("milestone %s has unparseable amount %q: %w", milestoneID, state.Amount, err)
	}

	now := ctx.BlockTime().UTC()

	if k.milestoneEscrow != nil && !amount.IsZero() {
		if err := k.milestoneEscrow.ReleaseMilestoneEscrow(ctx, orderID.String(), milestoneID, amount, "milestone_trigger_satisfied"); err != nil {
			return decision, err
		}
	}

	state.Status = marketplace.MilestoneStatusReleased
	state.SatisfiedAt = &now
	state.ReleasedAt = &now
	return decision, k.UpdateOrder(ctx, order)
}

// HoldOrderMilestone opens an order-linked dispute that holds exactly one
// milestone. The rest of the order continues unreleased but unaffected: the
// order is not resolved, and other milestones keep their own status.
func (k Keeper) HoldOrderMilestone(
	ctx sdk.Context,
	orderID marketplace.OrderID,
	milestoneID string,
	disputeID string,
	openedBy string,
	reason string,
	signal *marketplace.VEIDFraudSignal,
) (*marketplace.MilestoneDispute, error) {
	order, found := k.GetOrder(ctx, orderID)
	if !found {
		return nil, marketplace.ErrOrderNotFound
	}

	state, found := marketplace.FindMilestoneState(order.MilestoneStates, milestoneID)
	if !found {
		return nil, fmt.Errorf("milestone %q has no runtime state on order %s", milestoneID, orderID.String())
	}
	if state.Status == marketplace.MilestoneStatusReleased {
		return nil, fmt.Errorf("milestone %s has already released and cannot be held", milestoneID)
	}

	now := ctx.BlockTime().UTC()

	dispute := &marketplace.MilestoneDispute{
		DisputeID:   disputeID,
		OrderID:     orderID.String(),
		MilestoneID: milestoneID,
		Reason:      strings.TrimSpace(reason),
		OpenedBy:    openedBy,
		VEIDSignal:  signal,
		Open:        true,
		OpenedAt:    now,
	}
	if err := dispute.Validate(); err != nil {
		return nil, err
	}

	// Scope the hold to this milestone only. The order is deliberately left in
	// its current state.
	state.Status = marketplace.MilestoneStatusHeld
	state.HoldReason = dispute.Reason
	state.DisputeID = disputeID
	state.HeldAt = &now

	if err := k.SetMilestoneDispute(ctx, dispute); err != nil {
		return nil, err
	}
	if err := k.UpdateOrder(ctx, order); err != nil {
		return nil, err
	}
	return dispute, nil
}

// ResolveOrderMilestoneDispute resolves a milestone-scoped dispute, either
// releasing the held milestone (release=true) or refunding it (release=false).
func (k Keeper) ResolveOrderMilestoneDispute(
	ctx sdk.Context,
	disputeID string,
	release bool,
	resolution string,
) error {
	dispute, found := k.GetMilestoneDispute(ctx, disputeID)
	if !found {
		return marketplace.ErrDisputeNotFound
	}
	if !dispute.Open {
		return fmt.Errorf("dispute %s is already resolved", disputeID)
	}

	orderID, err := marketplace.ParseOrderID(dispute.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order id on dispute %s: %w", disputeID, err)
	}
	order, found := k.GetOrder(ctx, orderID)
	if !found {
		return marketplace.ErrOrderNotFound
	}

	state, found := marketplace.FindMilestoneState(order.MilestoneStates, dispute.MilestoneID)
	if !found {
		return fmt.Errorf("milestone %q has no runtime state", dispute.MilestoneID)
	}

	now := ctx.BlockTime().UTC()
	amount, err := sdk.ParseCoinsNormalized(state.Amount)
	if err != nil {
		return fmt.Errorf("milestone %s has unparseable amount %q: %w", dispute.MilestoneID, state.Amount, err)
	}

	if k.milestoneEscrow != nil && !amount.IsZero() {
		if release {
			if err := k.milestoneEscrow.ReleaseMilestoneEscrow(ctx, dispute.OrderID, dispute.MilestoneID, amount, "dispute_resolved_release"); err != nil {
				return err
			}
		} else {
			if err := k.milestoneEscrow.RefundMilestoneEscrow(ctx, dispute.OrderID, dispute.MilestoneID, amount, "dispute_resolved_refund"); err != nil {
				return err
			}
		}
	}

	if release {
		state.Status = marketplace.MilestoneStatusReleased
		state.SatisfiedAt = &now
		state.ReleasedAt = &now
	} else {
		state.Status = marketplace.MilestoneStatusRefunded
	}
	state.ResolvedAt = &now
	state.HoldReason = ""
	state.DisputeID = disputeID

	dispute.Open = false
	dispute.ResolvedAt = &now
	dispute.Resolution = strings.TrimSpace(resolution)

	if err := k.SetMilestoneDispute(ctx, dispute); err != nil {
		return err
	}
	return k.UpdateOrder(ctx, order)
}

// GetOrderMilestoneSummary returns per-order stage progress for a UI or dispute view.
func (k Keeper) GetOrderMilestoneSummary(ctx sdk.Context, orderID marketplace.OrderID) (marketplace.MilestoneOrderSummary, error) {
	order, found := k.GetOrder(ctx, orderID)
	if !found {
		return marketplace.MilestoneOrderSummary{}, marketplace.ErrOrderNotFound
	}
	return marketplace.SummarizeMilestones(order.Milestones, order.MilestoneStates), nil
}

// GetOfferingListing is the listing query. It reports whether evidence is
//
// It returns no raw evidence: only the hash is on chain (Constitution 39.1-39.2).
func (k Keeper) GetOfferingListing(ctx sdk.Context, id marketplace.OfferingID) (*marketplace.OfferingListingView, bool) {
	offering, found := k.GetOffering(ctx, id)
	if !found {
		return nil, false
	}

	// Expected evidence class: ownership entitlement is the stronger claim for
	// hardware, capacity is what a compute seller owes.
	expected := marketplace.AttestationKindCapacity
	if isHardwareOrComputeCategory(offering.Category) {
		expected = marketplace.AttestationKindOwnership
	}

	view := &marketplace.OfferingListingView{
		OfferingID:              id.String(),
		AttestationRequirement:  marketplace.AttestationRequirementText(expected),
		AttestationDoesNotProve: marketplace.AttestationDoesNotProveText(expected),
		IdentityRequirementText: identityRequirementAdvisoryText(&offering.IdentityRequirement),
	}

	if offering.Attestation != nil {
		view.AttestationPresent = true
		view.AttestationKind = offering.Attestation.Kind
		view.AttestationEvidenceHash = offering.Attestation.EvidenceHash
		view.AttestationProviderSupplied = offering.Attestation.ProviderSupplied
	}

	if set, err := k.ResolveOrderMilestones(ctx, offering); err == nil {
		view.Milestones = set
	}

	return view, true
}

// identityRequirementAdvisoryText renders the identity requirement as an
// advisory statement rather than a guarantee.
func identityRequirementAdvisoryText(req *marketplace.IdentityRequirement) string {
	parts := make([]string, 0, 4)
	if req.MinScore > 0 {
		parts = append(parts, fmt.Sprintf("minimum identity score %d", req.MinScore))
	}
	if req.RequiredStatus != "" {
		parts = append(parts, fmt.Sprintf("identity status %q", req.RequiredStatus))
	}
	if req.RequireVerifiedEmail {
		parts = append(parts, "verified email")
	}
	if req.RequireVerifiedDomain {
		parts = append(parts, "verified domain")
	}
	if req.RequireMFA {
		parts = append(parts, "multi-factor authentication enabled")
	}
	if len(parts) == 0 {
		return "No identity requirement is set for this listing. VEID does not prove that hardware or capacity exists, is owned by the seller, matches the advertised specification, or will be delivered; it is one advisory fraud signal only."
	}
	return fmt.Sprintf(
		"Requires %s. VEID confirms an identity signal only; it does not prove the hardware or capacity exists, is owned by the seller, matches the advertised specification, or will be delivered.",
		strings.Join(parts, ", "),
	)
}
