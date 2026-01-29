package marketplace

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	marketplacekeeper "github.com/virtengine/virtengine/x/market/types/marketplace/keeper"
)

// InitGenesis initializes module state from genesis.
func InitGenesis(ctx sdk.Context, k marketplacekeeper.IKeeper, gs *marketplacetypes.GenesisState) {
	if gs == nil {
		gs = marketplacetypes.DefaultGenesisState()
	}

	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	for _, offering := range gs.Offerings {
		off := offering
		if err := k.CreateOffering(ctx, &off); err != nil {
			panic(err)
		}
	}

	for _, order := range gs.Orders {
		or := order
		if err := k.CreateOrder(ctx, &or); err != nil {
			panic(err)
		}
	}

	for _, bid := range gs.Bids {
		b := bid
		if err := k.CreateBid(ctx, &b); err != nil {
			panic(err)
		}
	}

	for _, allocation := range gs.Allocations {
		alloc := allocation
		if err := k.CreateAllocation(ctx, &alloc); err != nil {
			panic(err)
		}
	}

	for address, settings := range gs.ProviderSettings {
		cfg := settings
		if err := k.SetProviderIdentitySettings(ctx, address, &cfg); err != nil {
			panic(err)
		}
	}

	for _, mfaConfig := range gs.MFAConfigs {
		cfg := mfaConfig
		if err := k.SetMFAActionConfig(ctx, &cfg); err != nil {
			panic(err)
		}
	}

	store := ctx.KVStore(k.StoreKey())
	bz, _ := json.Marshal(gs.EventSequence)
	store.Set(marketplacetypes.EventSequenceKey(), bz)
}

// ExportGenesis exports module state to genesis.
func ExportGenesis(ctx sdk.Context, k marketplacekeeper.IKeeper) *marketplacetypes.GenesisState {
	genesis := marketplacetypes.DefaultGenesisState()

	genesis.Params = k.GetParams(ctx)

	k.WithOfferings(ctx, func(offering marketplacetypes.Offering) bool {
		genesis.Offerings = append(genesis.Offerings, offering)
		return false
	})

	k.WithOrders(ctx, func(order marketplacetypes.Order) bool {
		genesis.Orders = append(genesis.Orders, order)
		return false
	})

	k.WithBids(ctx, func(bid marketplacetypes.MarketplaceBid) bool {
		genesis.Bids = append(genesis.Bids, bid)
		return false
	})

	k.WithAllocations(ctx, func(allocation marketplacetypes.Allocation) bool {
		genesis.Allocations = append(genesis.Allocations, allocation)
		return false
	})

	genesis.ProviderSettings = make(map[string]marketplacetypes.ProviderIdentitySettings)
	k.WithProviderSettings(ctx, func(address string, settings marketplacetypes.ProviderIdentitySettings) bool {
		genesis.ProviderSettings[address] = settings
		return false
	})

	k.WithMFAActionConfigs(ctx, func(config marketplacetypes.MFAActionConfig) bool {
		genesis.MFAConfigs = append(genesis.MFAConfigs, config)
		return false
	})

	genesis.EventSequence = k.GetEventSequence(ctx)
	return genesis
}
