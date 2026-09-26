package keeper

import (
	"bytes"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// mockMilestoneEscrow records each milestone release/refund so the tests can
// assert that funds move exactly once, for exactly the right amount.
type mockMilestoneEscrow struct {
	released []escrowCall
	refunded []escrowCall
}

type escrowCall struct {
	orderID     string
	milestoneID string
	amount      sdk.Coins
	reason      string
}

func (m *mockMilestoneEscrow) ReleaseMilestoneEscrow(_ sdk.Context, orderID, milestoneID string, amount sdk.Coins, reason string) error {
	m.released = append(m.released, escrowCall{orderID, milestoneID, amount, reason})
	return nil
}

func (m *mockMilestoneEscrow) RefundMilestoneEscrow(_ sdk.Context, orderID, milestoneID string, amount sdk.Coins, reason string) error {
	m.refunded = append(m.refunded, escrowCall{orderID, milestoneID, amount, reason})
	return nil
}

// newHardwareFixture builds provider/customer addresses, a compute offering that
// carries an ownership attestation, and an order against it.
func newHardwareFixture(t *testing.T) (*Keeper, sdk.Context, *mockMilestoneEscrow, marketplace.OfferingID, marketplace.OrderID) {
	t.Helper()

	k, ctx := setupKeeper(t)

	mock := &mockMilestoneEscrow{}
	k.SetMilestoneEscrowKeeper(mock)

	params := marketplace.DefaultParams()
	// Identity gating is exercised elsewhere; disable it so these tests isolate
	// the transaction-safeguard behaviour.
	params.EnableIdentityGating = false
	require.NoError(t, k.SetParams(ctx, params))

	provider := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()

	offering := marketplace.NewOfferingAt(
		marketplace.OfferingID{ProviderAddress: provider, Sequence: 1},
		"bare-metal-gpu",
		marketplace.OfferingCategoryGPU,
		marketplace.PricingInfo{Model: marketplace.PricingModelFixed, BasePrice: 1, Currency: "uve"},
		ctx.BlockTime(),
	)
	offering.ListingTerms = "Delivery within 14 days; capacity verifiable on demand."
	offering.Attestation = marketplace.NewOfferingAttestation(
		marketplace.AttestationKindOwnership,
		"provider-self",
		[]byte("ownership-deed-document-bytes"),
		true,
		ctx.BlockTime(),
	)
	require.NoError(t, k.CreateOffering(ctx, offering))

	order := marketplace.NewOrderAt(
		marketplace.OrderID{CustomerAddress: customer, Sequence: 1},
		offering.ID,
		1_000_000,
		1,
		ctx.BlockTime(),
	)
	order.State = marketplace.OrderStateOpen
	require.NoError(t, k.CreateOrder(ctx, order))

	return k, ctx, mock, offering.ID, order.ID
}

// orderTotal is the escrow total used across these tests.
func orderTotal() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1_000_000)))
}

// DONE WHEN 1: params define the default milestone set, and a test order moves
// through at least two milestones with escrow release at each.
func TestOrderMovesThroughTwoMilestonesWithEscrowReleaseAtEach(t *testing.T) {
	k, ctx, mock, _, orderID := newHardwareFixture(t)

	// The default milestone set comes from params.
	params := k.GetParams(ctx)
	require.True(t, params.MilestonePolicy.Enabled)
	require.Equal(t, marketplace.DefaultMilestoneSet(), params.MilestonePolicy.DefaultMilestones)
	require.NoError(t, params.MilestonePolicy.DefaultMilestones.Validate())

	require.NoError(t, k.InitializeOrderMilestones(ctx, orderID, orderTotal()))

	order, found := k.GetOrder(ctx, orderID)
	require.True(t, found)
	require.Len(t, order.Milestones, 2)
	require.Len(t, order.MilestoneStates, 2)

	summary, err := k.GetOrderMilestoneSummary(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, 2, summary.Total)
	require.Equal(t, 2, summary.Pending)
	require.False(t, summary.FullyReleased)

	// Milestone 1: capacity verified -> 25% (250000uve).
	decision, err := k.ReleaseOrderMilestone(ctx, orderID, "capacity_verified", marketplace.MilestoneEvidence{Satisfied: true})
	require.NoError(t, err)
	require.True(t, decision.Release, decision.Reason)

	require.Len(t, mock.released, 1)
	require.Equal(t, "capacity_verified", mock.released[0].milestoneID)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(250_000))), mock.released[0].amount)

	// Milestone 2: delivery attested -> remainder (750000uve).
	decision, err = k.ReleaseOrderMilestone(ctx, orderID, "delivery_attested", marketplace.MilestoneEvidence{Satisfied: true})
	require.NoError(t, err)
	require.True(t, decision.Release, decision.Reason)

	require.Len(t, mock.released, 2)
	require.Equal(t, "delivery_attested", mock.released[1].milestoneID)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(750_000))), mock.released[1].amount)

	// The two releases sum to exactly the escrow total: nothing created or lost.
	total := mock.released[0].amount.Add(mock.released[1].amount...)
	require.True(t, total.Equal(orderTotal()), "released %s must equal order total %s", total, orderTotal())

	summary, err = k.GetOrderMilestoneSummary(ctx, orderID)
	require.NoError(t, err)
	require.True(t, summary.FullyReleased)
	require.Equal(t, 2, summary.Released)

	// Releasing again is refused, so funds cannot be double-spent.
	decision, err = k.ReleaseOrderMilestone(ctx, orderID, "capacity_verified", marketplace.MilestoneEvidence{Satisfied: true})
	require.NoError(t, err)
	require.False(t, decision.Release)
	require.Len(t, mock.released, 2, "an already-released milestone must not move funds again")
}

// Milestone shares must partition the total exactly, including awkward amounts.
func TestMilestoneSplitIsExactAndConservesValue(t *testing.T) {
	set := marketplace.DefaultMilestoneSet()
	require.NoError(t, set.Validate())

	for _, amount := range []int64{1, 3, 7, 99, 1_000_001} {
		total := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(amount)))
		parts := set.Total(total)

		require.Len(t, parts, 2)
		sum := parts[0].Add(parts[1]...)
		require.True(t, sum.Equal(total), "amount %d: parts %s sum to %s", amount, parts, sum)
	}
}

// A schedule whose shares do not sum to 100% is rejected.
func TestMilestoneSetRejectsInvalidShares(t *testing.T) {
	bad := marketplace.MilestoneSet{
		{ID: "a", Description: "a", ShareBps: 2500, Trigger: marketplace.MilestoneTriggerCapacityVerified, Sequence: 1},
		{ID: "b", Description: "b", ShareBps: 2500, Trigger: marketplace.MilestoneTriggerDeliveryAttested, Sequence: 2},
	}
	require.ErrorContains(t, bad.Validate(), "must sum to 10000")

	// Non-contiguous sequences are rejected too.
	gapped := marketplace.MilestoneSet{
		{ID: "a", Description: "a", ShareBps: 5000, Trigger: marketplace.MilestoneTriggerCapacityVerified, Sequence: 1},
		{ID: "b", Description: "b", ShareBps: 5000, Trigger: marketplace.MilestoneTriggerDeliveryAttested, Sequence: 3},
	}
	require.ErrorContains(t, gapped.Validate(), "sequence 2 is missing")
}

// DONE WHEN 2: the order record carries the VEID signal, and settlement does not
// branch on it. A deliberately failing VEID signal must not block a satisfied
// milestone, and a passing VEID signal must not release an unsatisfied one.
func TestSettlementGateIgnoresVEIDSignal(t *testing.T) {
	k, ctx, mock, _, orderID := newHardwareFixture(t)
	require.NoError(t, k.InitializeOrderMilestones(ctx, orderID, orderTotal()))

	// The VEID signal is recorded as advisory only.
	failingSignal := marketplace.NewVEIDFraudSignal(0, 0, false, false, ctx.BlockTime())
	require.NoError(t, k.RecordOrderVEIDSignal(ctx, orderID, failingSignal))

	order, found := k.GetOrder(ctx, orderID)
	require.True(t, found)
	require.NotNil(t, order.VEIDSignal)
	require.True(t, order.VEIDSignal.Advisory)
	require.True(t, order.VEIDSignal.NotDeterminative)
	require.False(t, marketplace.VEIDIsDeterminative())

	// A *failing* VEID signal does not block a milestone whose own trigger is met.
	decision, err := k.ReleaseOrderMilestone(ctx, orderID, "capacity_verified", marketplace.MilestoneEvidence{Satisfied: true})
	require.NoError(t, err)
	require.True(t, decision.Release, "VEID must not gate a satisfied milestone: %s", decision.Reason)
	require.Len(t, mock.released, 1)

	// A *passing* VEID signal does not release a milestone whose trigger is unmet.
	passingSignal := marketplace.NewVEIDFraudSignal(100, 3, true, true, ctx.BlockTime())
	require.NoError(t, k.RecordOrderVEIDSignal(ctx, orderID, passingSignal))

	decision, err = k.ReleaseOrderMilestone(ctx, orderID, "delivery_attested", marketplace.MilestoneEvidence{Satisfied: false})
	require.NoError(t, err)
	require.False(t, decision.Release, "VEID must not authorise an unsatisfied milestone")
	require.Len(t, mock.released, 1, "no funds may move on VEID evidence alone")

	// The gate's signature admits no VEID input at all; this is the structural
	// guarantee behind the two assertions above.
	require.Contains(t, failingSignal.NotDeterminativeText(), "does not prove")
}

// DONE WHEN 3: the listing query returns attestation presence and requirement text.
func TestListingQueryReturnsAttestationPresenceAndRequirementText(t *testing.T) {
	k, ctx, _, offeringID, _ := newHardwareFixture(t)

	view, found := k.GetOfferingListing(ctx, offeringID)
	require.True(t, found)
	require.True(t, view.AttestationPresent)
	require.Equal(t, marketplace.AttestationKindOwnership, view.AttestationKind)
	require.True(t, view.AttestationProviderSupplied)
	require.Len(t, view.AttestationEvidenceHash, 64)
	require.NotEmpty(t, view.AttestationRequirement)
	require.NotEmpty(t, view.AttestationDoesNotProve)
	require.NotEmpty(t, view.IdentityRequirementText)
	require.Len(t, view.Milestones, 2)

	// The "what this does not prove" framing must be present, not implied.
	require.Contains(t, view.AttestationDoesNotProve, "does not prove")
	require.Contains(t, view.IdentityRequirementText, "does not prove")

	// Only a hash is exposed; the source document never reaches the query.
	require.NotContains(t, view.AttestationEvidenceHash, "ownership-deed")
}

// A hardware/compute listing with no attestation is refused by the guard.
func TestHardwareListingWithoutAttestationIsRefused(t *testing.T) {
	k, ctx, _, _, _ := newHardwareFixture(t)

	bare := marketplace.NewOfferingAt(
		marketplace.OfferingID{ProviderAddress: sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String(), Sequence: 2},
		"unattested-gpu",
		marketplace.OfferingCategoryGPU,
		marketplace.PricingInfo{Model: marketplace.PricingModelFixed, BasePrice: 1, Currency: "uve"},
		ctx.BlockTime(),
	)
	require.NoError(t, k.CreateOffering(ctx, bare))

	require.ErrorIs(t, k.RequireAttestationForOrder(ctx, bare), marketplace.ErrAttestationRequired)
}

// DONE WHEN 4: a single milestone can be held without resolving the whole order.
func TestSingleMilestoneCanBeHeldWithoutResolvingWholeOrder(t *testing.T) {
	k, ctx, mock, _, orderID := newHardwareFixture(t)
	require.NoError(t, k.InitializeOrderMilestones(ctx, orderID, orderTotal()))

	before, found := k.GetOrder(ctx, orderID)
	require.True(t, found)
	stateBefore := before.State

	// Hold only the delivery milestone.
	dispute, err := k.HoldOrderMilestone(
		ctx, orderID, "delivery_attested", "dispute-1", "customer-addr",
		"hardware did not match the advertised specification", nil,
	)
	require.NoError(t, err)
	require.True(t, dispute.Open)
	require.Equal(t, "delivery_attested", dispute.MilestoneID)

	// The whole order is NOT resolved: its lifecycle state is unchanged.
	after, found := k.GetOrder(ctx, orderID)
	require.True(t, found)
	require.Equal(t, stateBefore, after.State, "holding one milestone must not resolve the order")

	summary, err := k.GetOrderMilestoneSummary(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, 1, summary.Held)
	require.Equal(t, 1, summary.Pending)

	// The held milestone refuses to release even with satisfied evidence.
	decision, err := k.ReleaseOrderMilestone(ctx, orderID, "delivery_attested", marketplace.MilestoneEvidence{Satisfied: true})
	require.NoError(t, err)
	require.False(t, decision.Release)
	require.Contains(t, decision.Reason, "held")
	require.Empty(t, mock.released)

	// The *other* milestone is unaffected and can still release.
	decision, err = k.ReleaseOrderMilestone(ctx, orderID, "capacity_verified", marketplace.MilestoneEvidence{Satisfied: true})
	require.NoError(t, err)
	require.True(t, decision.Release, decision.Reason)
	require.Len(t, mock.released, 1)
	require.Equal(t, "capacity_verified", mock.released[0].milestoneID)

	// Resolving the dispute refunds only that milestone.
	require.NoError(t, k.ResolveOrderMilestoneDispute(ctx, "dispute-1", false, "refunded: spec mismatch"))
	require.Len(t, mock.refunded, 1)
	require.Equal(t, "delivery_attested", mock.refunded[0].milestoneID)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(750_000))), mock.refunded[0].amount)

	summary, err = k.GetOrderMilestoneSummary(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, 1, summary.Released)
	require.Equal(t, 1, summary.Refunded)
	require.Equal(t, 0, summary.Held)
	require.False(t, summary.FullyReleased)

	stored, found := k.GetMilestoneDispute(ctx, "dispute-1")
	require.True(t, found)
	require.False(t, stored.Open)
	require.NotNil(t, stored.ResolvedAt)
}

// The dispute record carries the advisory VEID signal without acting on it.
func TestMilestoneDisputeCarriesAdvisoryVEIDSignal(t *testing.T) {
	k, ctx, _, _, orderID := newHardwareFixture(t)
	require.NoError(t, k.InitializeOrderMilestones(ctx, orderID, orderTotal()))

	signal := marketplace.NewVEIDFraudSignal(42, 1, false, false, ctx.BlockTime())
	dispute, err := k.HoldOrderMilestone(
		ctx, orderID, "capacity_verified", "dispute-2", "customer-addr", "capacity could not be verified", signal,
	)
	require.NoError(t, err)
	require.NotNil(t, dispute.VEIDSignal)
	require.True(t, dispute.VEIDSignal.Advisory)
	require.True(t, dispute.VEIDSignal.NotDeterminative)
	require.Equal(t, uint32(42), dispute.VEIDSignal.Score)
}

// A per-listing override replaces the params default for that listing.
func TestPerListingMilestoneOverrideIsHonoured(t *testing.T) {
	k, ctx, _, offeringID, orderID := newHardwareFixture(t)

	offering, found := k.GetOffering(ctx, offeringID)
	require.True(t, found)
	offering.MilestoneOverride = marketplace.MilestoneSet{
		{ID: "deposit", Description: "booking deposit", ShareBps: 1000, Trigger: marketplace.MilestoneTriggerCapacityVerified, Sequence: 1},
		{ID: "final", Description: "delivery", ShareBps: 9000, Trigger: marketplace.MilestoneTriggerDeliveryAttested, Sequence: 2},
	}
	require.NoError(t, k.UpdateOffering(ctx, offering))

	require.NoError(t, k.InitializeOrderMilestones(ctx, orderID, orderTotal()))
	order, found := k.GetOrder(ctx, orderID)
	require.True(t, found)
	require.Len(t, order.Milestones, 2)
	require.Equal(t, "deposit", order.Milestones[0].ID)
}

// Block time is the only clock, and recorded timestamps come from it.
func TestMilestoneStateUsesBlockTime(t *testing.T) {
	k, ctx, mock, _, orderID := newHardwareFixture(t)
	require.NoError(t, k.InitializeOrderMilestones(ctx, orderID, orderTotal()))

	blockTime := time.Unix(1_800_000_000, 0).UTC()
	ctx = ctx.WithBlockTime(blockTime)

	_, err := k.ReleaseOrderMilestone(ctx, orderID, "capacity_verified", marketplace.MilestoneEvidence{Satisfied: true})
	require.NoError(t, err)
	require.Len(t, mock.released, 1)

	order, found := k.GetOrder(ctx, orderID)
	require.True(t, found)
	state, found := marketplace.FindMilestoneState(order.MilestoneStates, "capacity_verified")
	require.True(t, found)
	require.NotNil(t, state.ReleasedAt)
	require.True(t, state.ReleasedAt.Equal(blockTime), "released_at must equal block time")
}
