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

func TestGenesisRejectsMutableNonOwnerLifecycle(t *testing.T) {
	provider := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	offering := marketplace.NewOfferingAt(marketplace.OfferingID{ProviderAddress: provider, Sequence: 1}, "catalog", marketplace.OfferingCategoryCompute, marketplace.PricingInfo{Model: marketplace.PricingModelFixed, BasePrice: 1, Currency: "uve"}, time.Unix(1, 0).UTC())
	offering.State = marketplace.OfferingStateActive
	order := marketplace.NewOrderAt(marketplace.OrderID{CustomerAddress: customer, Sequence: 1}, offering.ID, 1, 1, time.Unix(1, 0).UTC())
	order.State = marketplace.OrderStateOpen

	genesis := marketplace.DefaultGenesisState()
	genesis.CanonicalLifecycleActive = true
	genesis.Offerings = append(genesis.Offerings, *offering)
	genesis.Orders = append(genesis.Orders, *order)
	require.ErrorContains(t, genesis.Validate(), "must be terminal at canonical activation")

	genesis.CanonicalLifecycleActive = false
	require.NoError(t, genesis.Validate(), "historical pre-upgrade genesis remains replayable")
}

func TestDefaultGenesisActivatesCurrentProtocolForNewChains(t *testing.T) {
	require.True(t, marketplace.DefaultGenesisState().CanonicalLifecycleActive)
}

func TestCanonicalActivationRejectsLifecycleWritesAndPreservesCatalog(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.ActivateCanonicalLifecycle(ctx)

	provider := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	offering := marketplace.NewOfferingAt(marketplace.OfferingID{ProviderAddress: provider, Sequence: 1}, "catalog", marketplace.OfferingCategoryCompute, marketplace.PricingInfo{Model: marketplace.PricingModelFixed, BasePrice: 1, Currency: "uve"}, ctx.BlockTime())
	offering.State = marketplace.OfferingStateActive
	require.NoError(t, k.CreateOffering(ctx, offering), "supply catalog writes remain compatible")

	order := marketplace.NewOrderAt(marketplace.OrderID{CustomerAddress: customer, Sequence: 1}, offering.ID, 1, 1, ctx.BlockTime())
	order.State = marketplace.OrderStateOpen
	require.ErrorIs(t, k.CreateOrder(ctx, order), marketplace.ErrLifecycleDeprecated)

	bid := marketplace.MarketplaceBid{ID: marketplace.BidID{OrderID: order.ID, ProviderAddress: provider, Sequence: 1}, OfferingID: offering.ID, Price: 1}
	require.ErrorIs(t, k.CreateBid(ctx, &bid), marketplace.ErrLifecycleDeprecated)

	allocation := marketplace.NewAllocationAt(marketplace.AllocationID{OrderID: order.ID, Sequence: 1}, offering.ID, provider, bid.ID, 1, ctx.BlockTime())
	require.ErrorIs(t, k.CreateAllocation(ctx, allocation), marketplace.ErrLifecycleDeprecated)

	callback := marketplace.NewWaldurCallbackAt(marketplace.ActionTypeStatusUpdate, "waldur-1", marketplace.SyncTypeAllocation, allocation.ID.String(), ctx.BlockTime())
	require.ErrorIs(t, k.ProcessWaldurCallback(ctx, callback), marketplace.ErrLifecycleDeprecated)
	require.ErrorIs(t, k.RequestUsageUpdate(ctx, allocation.ID, "periodic"), marketplace.ErrLifecycleDeprecated)
}

func TestPreActivationLifecycleWriteRemainsReplayable(t *testing.T) {
	k, ctx := setupKeeper(t)
	params := marketplace.DefaultParams()
	params.EnableIdentityGating = false
	require.NoError(t, k.SetParams(ctx, params))
	provider := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	offering := marketplace.NewOfferingAt(marketplace.OfferingID{ProviderAddress: provider, Sequence: 1}, "catalog", marketplace.OfferingCategoryCompute, marketplace.PricingInfo{Model: marketplace.PricingModelFixed, BasePrice: 1, Currency: "uve"}, ctx.BlockTime())
	offering.State = marketplace.OfferingStateActive
	require.NoError(t, k.CreateOffering(ctx, offering))

	order := marketplace.NewOrderAt(marketplace.OrderID{CustomerAddress: customer, Sequence: 1}, offering.ID, 1, 1, ctx.BlockTime())
	order.State = marketplace.OrderStateOpen
	require.NoError(t, k.CreateOrder(ctx, order))
	stored, found := k.GetOrder(ctx, order.ID)
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateOpen, stored.State)
}

func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	key := storetypes.NewKVStoreKey(marketplace.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1, Time: time.Unix(1, 0).UTC()}, false, log.NewNopLogger())
	return NewKeeper(cdc, key, sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String(), nil, nil, nil), ctx
}
