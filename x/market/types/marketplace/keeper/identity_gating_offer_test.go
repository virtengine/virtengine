package keeper

import (
	"bytes"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	marketplace "github.com/virtengine/virtengine/x/market/types/marketplace"
)

// stubVEID is a controllable VEIDKeeper for per-listing gating tests.
type stubVEID struct {
	score  uint32
	status string
	locked bool
}

func (s stubVEID) GetIdentityScore(_ sdk.Context, _ sdk.AccAddress) (uint32, bool) {
	return s.score, true
}

func (s stubVEID) GetIdentityStatus(_ sdk.Context, _ sdk.AccAddress) (string, bool) {
	return s.status, true
}

func (s stubVEID) IsEmailVerified(_ sdk.Context, _ sdk.AccAddress) bool { return true }

func (s stubVEID) IsDomainVerified(_ sdk.Context, _ sdk.AccAddress) bool { return true }

func (s stubVEID) IsIdentityLocked(_ sdk.Context, _ sdk.AccAddress) bool { return s.locked }

func (s stubVEID) IsComplianceCleared(_ sdk.Context, _ sdk.AccAddress) (bool, bool) {
	return true, true
}

// stubMFA reports MFA as enabled so MFA is never the reason a test fails.
type stubMFA struct{ enabled bool }

func (s stubMFA) HasActiveFactors(_ sdk.Context, _ sdk.AccAddress) bool { return s.enabled }

func (s stubMFA) GetLastMFAVerification(_ sdk.Context, _ sdk.AccAddress) (*time.Time, bool) {
	return nil, false
}

func (s stubMFA) IsTrustedDevice(_ sdk.Context, _ sdk.AccAddress, _ string) bool { return false }

func (s stubMFA) CreateChallenge(_ sdk.Context, _ sdk.AccAddress, _ string) (string, error) {
	return "", nil
}

func (s stubMFA) VerifyChallenge(_ sdk.Context, _ string, _ interface{}) (bool, error) {
	return true, nil
}

func setupGatingKeeper(t *testing.T, veid stubVEID) (*Keeper, sdk.Context) {
	t.Helper()

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	key := storetypes.NewKVStoreKey(marketplace.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1, Time: time.Unix(1, 0).UTC()}, false, log.NewNopLogger())

	return NewKeeper(cdc, key, sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String(), veid, stubMFA{enabled: true}, nil), ctx
}

// newActiveOffering builds an active offering with the supplied requirement.
func newActiveOffering(t *testing.T, ctx sdk.Context, req marketplace.IdentityRequirement) *marketplace.Offering {
	t.Helper()

	offering := marketplace.NewOfferingAt(
		marketplace.OfferingID{ProviderAddress: sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String(), Sequence: 1},
		"gated-catalog",
		marketplace.OfferingCategoryCompute,
		marketplace.PricingInfo{Model: marketplace.PricingModelFixed, BasePrice: 10, Currency: "uve"},
		ctx.BlockTime(),
	)
	offering.State = marketplace.OfferingStateActive
	offering.IdentityRequirement = req
	return offering
}

func newOrderFor(ctx sdk.Context, offering *marketplace.Offering, customer string) *marketplace.Order {
	order := marketplace.NewOrderAt(
		marketplace.OrderID{CustomerAddress: customer, Sequence: 1},
		offering.ID,
		100,
		1,
		ctx.BlockTime(),
	)
	order.State = marketplace.OrderStateOpen
	return order
}

// DONE-WHEN 1(a): a default offering creates an order with NO identity check,
// even when the buyer has no identity record at all (score 0 / unverified).
func TestCreateOrder_DefaultOfferingRequiresNoIdentity(t *testing.T) {
	// Zero score and unverified status would fail any real requirement.
	k, ctx := setupGatingKeeper(t, stubVEID{score: 0, status: "unverified", locked: true})

	offering := newActiveOffering(t, ctx, marketplace.DefaultIdentityRequirement())
	require.NoError(t, k.CreateOffering(ctx, offering))

	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	order := newOrderFor(ctx, offering, customer)

	require.NoError(t, k.CreateOrder(ctx, order), "default listing must not gate the buyer")

	stored, found := k.GetOrder(ctx, order.ID)
	require.True(t, found, "order must be persisted when no requirement is declared")
	require.Equal(t, marketplace.OrderStateOpen, stored.State)
}

// DONE-WHEN 1(b): an offering that declares a requirement rejects a
// non-matching identity with a typed error, and the order is not persisted.
func TestCreateOrder_OfferingRequirementRejectsNonMatchingIdentity(t *testing.T) {
	k, ctx := setupGatingKeeper(t, stubVEID{score: 10, status: "pending"})

	offering := newActiveOffering(t, ctx, marketplace.IdentityRequirement{
		MinScore:       80,
		RequiredStatus: "verified",
	})
	require.NoError(t, k.CreateOffering(ctx, offering))

	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	order := newOrderFor(ctx, offering, customer)

	err := k.CreateOrder(ctx, order)
	require.Error(t, err, "order from a non-matching identity must be rejected")

	// The rejection is a typed, structured error so a client can render exactly
	// which proof is missing.
	var gatingErr *marketplace.IdentityGatingError
	require.ErrorAs(t, err, &gatingErr, "rejection must be the typed identity gating error")
	require.True(t, gatingErr.HasErrors())
	require.Equal(t, offering.ID, gatingErr.OfferingID)

	_, found := k.GetOrder(ctx, order.ID)
	require.False(t, found, "rejected order must not be persisted")
}

// The matching buyer for the same listing must still be accepted: the gate is
// scoped to the declared requirement, not a blanket block.
func TestCreateOrder_OfferingRequirementAcceptsMatchingIdentity(t *testing.T) {
	k, ctx := setupGatingKeeper(t, stubVEID{score: 90, status: "verified"})

	offering := newActiveOffering(t, ctx, marketplace.IdentityRequirement{
		MinScore:       80,
		RequiredStatus: "verified",
	})
	require.NoError(t, k.CreateOffering(ctx, offering))

	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	require.NoError(t, k.CreateOrder(ctx, newOrderFor(ctx, offering, customer)))
}

// DONE-WHEN 1(c): a requirement set to ONLY RequireUnlockedIdentity is not
// treated as zero, and it actually gates a locked identity.
func TestCreateOrder_UnlockedOnlyRequirementIsNotZero(t *testing.T) {
	req := marketplace.IdentityRequirement{RequireUnlockedIdentity: true}
	require.False(t, req.IsZero(), "unlocked-only requirement must not be zero")

	// Locked buyer must be rejected by an unlocked-only listing.
	k, ctx := setupGatingKeeper(t, stubVEID{score: 100, status: "verified", locked: true})
	offering := newActiveOffering(t, ctx, req)
	require.NoError(t, k.CreateOffering(ctx, offering))

	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	err := k.CreateOrder(ctx, newOrderFor(ctx, offering, customer))
	require.Error(t, err, "locked identity must be rejected by an unlocked-only listing")

	var gatingErr *marketplace.IdentityGatingError
	require.ErrorAs(t, err, &gatingErr, "rejection must be the typed identity gating error")

	// An unlocked buyer passes the same listing.
	k2, ctx2 := setupGatingKeeper(t, stubVEID{score: 100, status: "verified", locked: false})
	offering2 := newActiveOffering(t, ctx2, req)
	require.NoError(t, k2.CreateOffering(ctx2, offering2))
	customer2 := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	require.NoError(t, k2.CreateOrder(ctx2, newOrderFor(ctx2, offering2, customer2)))
}

// DONE-WHEN 2 (partial): the bid path resolves listing terms through the order
// it is bound to. This pins the plumbing the bid path relies on — an order that
// exists for a gated offering is accepted, and a bid with no order fails rather
// than silently passing.
//
// NOTE: buyer identity requirements are already enforced when the order is
// created, and a bidder is a provider, not the buyer. Provider-side gating for
// the bid path is tracked separately (see PR body); this test does not claim to
// cover it.
func TestCreateBid_ResolvesOrderBoundToGatedOffering(t *testing.T) {
	k, ctx := setupGatingKeeper(t, stubVEID{score: 90, status: "verified"})

	provider := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	bid := marketplace.MarketplaceBid{
		ID:         marketplace.BidID{OrderID: marketplace.OrderID{CustomerAddress: sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String(), Sequence: 1}, ProviderAddress: provider, Sequence: 1},
		OfferingID: marketplace.OfferingID{ProviderAddress: provider, Sequence: 1},
		Price:      10,
	}

	// Requirement declared but the order does not exist yet: the bid path must
	// not silently pass a gate it cannot evaluate.
	err := k.CreateBid(ctx, &bid)
	require.Error(t, err, "bid without a matching order must fail")

	offering := newActiveOffering(t, ctx, marketplace.IdentityRequirement{MinScore: 80})
	require.NoError(t, k.CreateOffering(ctx, offering))
	require.NoError(t, k.CreateOrder(ctx, newOrderFor(ctx, offering, bid.ID.OrderID.CustomerAddress)))

	bid.OfferingID = offering.ID
	require.NoError(t, k.CreateBid(ctx, &bid), "matching identity must be allowed to bid")
}

// DONE-WHEN 3: the query surface returns the effective requirement, including
// the "none" case, for an offering.
func TestGetEffectiveIdentityRequirement_ReturnsNoneAndDeclared(t *testing.T) {
	k, ctx := setupGatingKeeper(t, stubVEID{score: 0, status: "unverified"})

	// Absent offering reports not-found rather than a zero requirement.
	_, found := k.GetEffectiveIdentityRequirement(ctx, marketplace.OfferingID{ProviderAddress: sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String(), Sequence: 99})
	require.False(t, found)

	// Default listing: requirement is explicitly "none".
	plain := newActiveOffering(t, ctx, marketplace.DefaultIdentityRequirement())
	require.NoError(t, k.CreateOffering(ctx, plain))

	req, found := k.GetEffectiveIdentityRequirement(ctx, plain.ID)
	require.True(t, found)
	require.True(t, req.IsZero(), "a default listing must report no identity requirement")

	// Declared listing reports the exact proof requested.
	gated := newActiveOffering(t, ctx, marketplace.IdentityRequirement{MinScore: 75, RequireUnlockedIdentity: true})
	gated.ID.Sequence = 2
	require.NoError(t, k.CreateOffering(ctx, gated))

	req, found = k.GetEffectiveIdentityRequirement(ctx, gated.ID)
	require.True(t, found)
	require.Equal(t, uint32(75), req.MinScore)
	require.True(t, req.RequireUnlockedIdentity)
	require.False(t, req.IsZero())
}

// The checker must not gate anything when the listing declares nothing, even
// with a buyer whose identity is entirely absent.
func TestCheckIdentityGating_ZeroRequirementGatesNobody(t *testing.T) {
	k, ctx := setupGatingKeeper(t, stubVEID{score: 0, status: "unverified", locked: true})

	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20))
	require.NoError(t, k.CheckIdentityGating(ctx, newActiveOffering(t, ctx, marketplace.DefaultIdentityRequirement()), customer))
}

// A nil offering is a programming error and must be reported, never treated as
// "no requirements".
func TestCheckIdentityGating_NilOffering(t *testing.T) {
	k, ctx := setupGatingKeeper(t, stubVEID{score: 100, status: "verified"})
	require.ErrorIs(t, k.CheckIdentityGating(ctx, nil, sdk.AccAddress(bytes.Repeat([]byte{8}, 20))), marketplace.ErrOfferingNotFound)
}
