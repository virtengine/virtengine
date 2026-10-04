package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

type fakeProviderKeeper struct {
	allowed map[string]bool
}

func (f fakeProviderKeeper) IsProvider(_ sdk.Context, addr sdk.AccAddress) bool {
	return f.allowed[addr.String()]
}

func (f fakeProviderKeeper) GetProvider(_ sdk.Context, _ sdk.AccAddress) (interface{}, bool) {
	return nil, false
}

func setupResolutionKeeper(t *testing.T, providers ...string) (*Keeper, sdk.Context) {
	t.Helper()
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	key := storetypes.NewKVStoreKey(marketplace.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 10, Time: time.Unix(100, 0).UTC()}, false, log.NewNopLogger())

	allowed := make(map[string]bool, len(providers))
	for _, p := range providers {
		allowed[p] = true
	}
	k := NewKeeper(cdc, key, sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String(), nil, nil, fakeProviderKeeper{allowed: allowed})
	params := marketplace.DefaultParams()
	params.EnableIdentityGating = false
	params.EnableAutoResolution = true
	require.NoError(t, k.SetParams(ctx, params))
	return k, ctx
}

func providerAddr(seed byte) string {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20)).String()
}

func seedOffering(t *testing.T, k *Keeper, ctx sdk.Context, offering marketplace.Offering) {
	t.Helper()
	bz, err := json.Marshal(offering)
	require.NoError(t, err)
	ctx.KVStore(k.skey).Set(marketplace.OfferingKey(offering.ID), bz)
}

func seedOrder(t *testing.T, k *Keeper, ctx sdk.Context, order marketplace.Order) {
	t.Helper()
	bz, err := json.Marshal(order)
	require.NoError(t, err)
	ctx.KVStore(k.skey).Set(marketplace.OrderKey(order.ID), bz)
}

func seedBid(t *testing.T, k *Keeper, ctx sdk.Context, bid marketplace.MarketplaceBid) {
	t.Helper()
	bz, err := json.Marshal(bid)
	require.NoError(t, err)
	ctx.KVStore(k.skey).Set(marketplace.BidKey(bid.ID), bz)
}

func fixedOffering(provider string, seq uint64, price uint64) marketplace.Offering {
	offering := marketplace.NewOfferingAt(
		marketplace.OfferingID{ProviderAddress: provider, Sequence: seq},
		"offering",
		marketplace.OfferingCategoryCompute,
		marketplace.PricingInfo{Model: marketplace.PricingModelHourly, Currency: "uvirt", BasePrice: price},
		time.Unix(1, 0).UTC(),
	)
	offering.State = marketplace.OfferingStateActive
	offering.Visibility = marketplace.OfferingVisibilityPublic
	return *offering
}

func TestResolveDirectOrderPicksCheapestListing(t *testing.T) {
	cheap := providerAddr(2)
	expensive := providerAddr(3)
	k, ctx := setupResolutionKeeper(t, cheap, expensive)

	seedOffering(t, k, ctx, fixedOffering(expensive, 1, 900))
	seedOffering(t, k, ctx, fixedOffering(cheap, 2, 100))

	order := marketplace.Order{
		ID:                marketplace.OrderID{CustomerAddress: providerAddr(8), Sequence: 1},
		State:             marketplace.OrderStateOpen,
		AcquisitionMode:   marketplace.AcquisitionModeDirect,
		Selector:          &marketplace.OfferSelector{Category: marketplace.OfferingCategoryCompute, MinSpecs: map[string]uint64{}},
		RequestedQuantity: 1,
		MaxBidPrice:       1000,
		CreatedAt:         time.Unix(1, 0).UTC(),
		UpdatedAt:         time.Unix(1, 0).UTC(),
	}
	seedOrder(t, k, ctx, order)

	resolved, err := k.ResolveOpenOrders(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, resolved)

	stored, found := k.GetOrder(ctx, order.ID)
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateMatched, stored.State)
	require.Equal(t, cheap, stored.AllocatedProviderAddress)
	require.Equal(t, uint64(100), stored.AcceptedPrice)
}

func TestResolveBidOrderPicksLowestBid(t *testing.T) {
	low := providerAddr(4)
	high := providerAddr(5)
	k, ctx := setupResolutionKeeper(t, low, high)

	orderID := marketplace.OrderID{CustomerAddress: providerAddr(8), Sequence: 2}
	deadline := ctx.BlockTime().Add(-time.Minute)
	order := marketplace.Order{
		ID:                orderID,
		State:             marketplace.OrderStateOpen,
		AcquisitionMode:   marketplace.AcquisitionModeBid,
		Selector:          &marketplace.OfferSelector{},
		MatchingDeadline:  &deadline,
		RequestedQuantity: 1,
		MaxBidPrice:       1000,
		CreatedAt:         time.Unix(1, 0).UTC(),
		UpdatedAt:         time.Unix(1, 0).UTC(),
	}
	seedOrder(t, k, ctx, order)

	seedBid(t, k, ctx, marketplace.MarketplaceBid{
		ID:        marketplace.BidID{OrderID: orderID, ProviderAddress: high, Sequence: 1},
		Price:     800,
		State:     marketplace.BidStateOpen,
		CreatedAt: time.Unix(1, 0).UTC(),
		UpdatedAt: time.Unix(1, 0).UTC(),
	})
	lowBidID := marketplace.BidID{OrderID: orderID, ProviderAddress: low, Sequence: 2}
	seedBid(t, k, ctx, marketplace.MarketplaceBid{
		ID:        lowBidID,
		Price:     400,
		State:     marketplace.BidStateOpen,
		CreatedAt: time.Unix(1, 0).UTC(),
		UpdatedAt: time.Unix(1, 0).UTC(),
	})

	resolved, err := k.ResolveOpenOrders(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, resolved)

	stored, found := k.GetOrder(ctx, orderID)
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateMatched, stored.State)
	require.Equal(t, low, stored.AllocatedProviderAddress)

	accepted, found := k.GetBid(ctx, lowBidID)
	require.True(t, found)
	require.Equal(t, marketplace.BidStateAccepted, accepted.State)

	loser, found := k.GetBid(ctx, marketplace.BidID{OrderID: orderID, ProviderAddress: high, Sequence: 1})
	require.True(t, found)
	require.Equal(t, marketplace.BidStateRejected, loser.State)
}

func TestResolveBidOrderWithoutBidsFailsAtDeadline(t *testing.T) {
	k, ctx := setupResolutionKeeper(t)
	deadline := ctx.BlockTime().Add(-time.Minute)
	order := marketplace.Order{
		ID:                marketplace.OrderID{CustomerAddress: providerAddr(8), Sequence: 3},
		State:             marketplace.OrderStateOpen,
		AcquisitionMode:   marketplace.AcquisitionModeBid,
		Selector:          &marketplace.OfferSelector{Category: marketplace.OfferingCategoryGPU},
		MatchingDeadline:  &deadline,
		RequestedQuantity: 1,
		MaxBidPrice:       1000,
		CreatedAt:         time.Unix(1, 0).UTC(),
		UpdatedAt:         time.Unix(1, 0).UTC(),
	}
	seedOrder(t, k, ctx, order)

	resolved, err := k.ResolveOpenOrders(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, resolved)

	stored, found := k.GetOrder(ctx, order.ID)
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateFailed, stored.State)
}

func TestIngestWaldurOffering(t *testing.T) {
	provider := providerAddr(6)
	k, ctx := setupResolutionKeeper(t, provider)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	require.NoError(t, k.SetWaldurSource(ctx, &marketplace.WaldurSource{
		InstanceID:   "waldur-eu",
		PublicKey:    hex.EncodeToString(pub),
		RegisteredAt: ctx.BlockTime(),
		Active:       true,
	}))

	imp := &marketplace.WaldurOfferingImport{
		UUID:         "offer-uuid-1",
		InstanceID:   "waldur-eu",
		Name:         "Ingested Compute",
		Description:  "ingested",
		Type:         "VirtEngine.Compute",
		State:        "Active",
		CustomerUUID: "cust-1",
		Components: []marketplace.WaldurPricingComponent{
			{Name: "hourly", Type: "usage", BillingType: "usage", MeasuredUnit: "hour", Price: "0.50"},
		},
		Attributes: map[string]interface{}{"ve_provider": provider},
	}
	att := marketplace.SignWaldurOfferingAttestation("waldur-eu", imp.UUID, imp.IngestChecksum(), 1, priv)

	result, err := k.IngestWaldurOffering(ctx, imp, att)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, marketplace.IngestActionCreate, result.Action)

	offeringID := marketplace.OfferingID{ProviderAddress: provider, Sequence: 1}
	stored, found := k.GetOffering(ctx, offeringID)
	require.True(t, found)
	require.Equal(t, marketplace.OfferingSourceWaldur, stored.Source)
	require.NotNil(t, stored.Waldur)
	require.Equal(t, "offer-uuid-1", stored.Waldur.OfferingUUID)

	// Replay of the same snapshot is skipped.
	replay, err := k.IngestWaldurOffering(ctx, imp, att)
	require.NoError(t, err)
	require.Equal(t, marketplace.IngestActionSkip, replay.Action)
}

func TestIngestWaldurOfferingRejectsBadSignature(t *testing.T) {
	provider := providerAddr(6)
	k, ctx := setupResolutionKeeper(t, provider)

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	_, otherPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	require.NoError(t, k.SetWaldurSource(ctx, &marketplace.WaldurSource{
		InstanceID: "waldur-eu",
		PublicKey:  hex.EncodeToString(pub),
		Active:     true,
	}))

	imp := &marketplace.WaldurOfferingImport{
		UUID:         "offer-uuid-2",
		InstanceID:   "waldur-eu",
		Name:         "Bad Sig",
		Type:         "VirtEngine.Compute",
		State:        "Active",
		CustomerUUID: "cust-1",
		Attributes:   map[string]interface{}{"ve_provider": provider},
	}
	att := marketplace.SignWaldurOfferingAttestation("waldur-eu", imp.UUID, imp.IngestChecksum(), 1, otherPriv)

	_, err = k.IngestWaldurOffering(ctx, imp, att)
	require.Error(t, err)
}

func TestUnifiedCatalogFiltersVisibilityAndSource(t *testing.T) {
	provider := providerAddr(7)
	k, ctx := setupResolutionKeeper(t)

	pub := fixedOffering(provider, 1, 100)
	pub.Visibility = marketplace.OfferingVisibilityPublic
	seedOffering(t, k, ctx, pub)

	hidden := fixedOffering(provider, 2, 100)
	hidden.Visibility = marketplace.OfferingVisibilityPrivate
	seedOffering(t, k, ctx, hidden)

	catalog := k.UnifiedCatalog(ctx, CatalogFilter{})
	require.Len(t, catalog, 1)
	require.Equal(t, uint64(1), catalog[0].ID.Sequence)

	all := k.UnifiedCatalog(ctx, CatalogFilter{IncludeUnlisted: true})
	require.Len(t, all, 2)
}

func TestEnqueueAndAckWaldurCommand(t *testing.T) {
	provider := providerAddr(7)
	k, ctx := setupResolutionKeeper(t, provider)

	offering := fixedOffering(provider, 1, 100)
	offering.Source = marketplace.OfferingSourceWaldur
	offering.Waldur = &marketplace.WaldurOfferingRef{InstanceID: "waldur-eu", OfferingUUID: "uuid-1"}
	seedOffering(t, k, ctx, offering)

	order := marketplace.Order{
		ID:                marketplace.OrderID{CustomerAddress: providerAddr(8), Sequence: 5},
		OfferingID:        offering.ID,
		State:             marketplace.OrderStateOpen,
		AcquisitionMode:   marketplace.AcquisitionModeDirect,
		RequestedQuantity: 1,
		MaxBidPrice:       1000,
		CreatedAt:         time.Unix(1, 0).UTC(),
		UpdatedAt:         time.Unix(1, 0).UTC(),
	}
	seedOrder(t, k, ctx, order)

	_, err := k.ResolveOpenOrders(ctx)
	require.NoError(t, err)

	var commands []marketplace.WaldurCommand
	k.WithWaldurCommands(ctx, func(c marketplace.WaldurCommand) bool {
		commands = append(commands, c)
		return false
	})
	require.Len(t, commands, 1)
	require.Equal(t, marketplace.WaldurCommandCreateOrder, commands[0].Kind)

	require.NoError(t, k.AckWaldurCommand(ctx, commands[0].ID))
	acked, found := k.GetWaldurCommand(ctx, commands[0].ID)
	require.True(t, found)
	require.True(t, acked.Acked)
}
