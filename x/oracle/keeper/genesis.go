// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	otypes "github.com/virtengine/virtengine/x/oracle/types"
)

// InitGenesis initializes the Oracle module's state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper IKeeper, data *types.GenesisState) {
	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Store all prices from genesis
	store := ctx.KVStore(keeper.StoreKey())
	cdc := keeper.Codec()

	for _, priceData := range data.Prices {
		key := otypes.PriceDataKey(
			priceData.ID.Source,
			priceData.ID.Denom,
			priceData.ID.BaseDenom,
			priceData.ID.Height,
		)
		bz := cdc.MustMarshal(&priceData)
		store.Set(key, bz)
	}

	// Store latest height entries
	for _, latestID := range data.LatestHeight {
		key := otypes.LatestPriceDataKey(
			latestID.Source,
			latestID.Denom,
			latestID.BaseDenom,
		)
		bz := cdc.MustMarshal(&latestID)
		store.Set(key, bz)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper IKeeper) *types.GenesisState {
	params := keeper.GetParams(ctx)

	// Export all prices
	prices, _ := keeper.GetPrices(ctx, types.PricesFilter{})

	// Build latest height entries
	latestHeight := make([]types.PriceDataID, 0)
	seenPairs := make(map[string]bool)

	for _, priceData := range prices {
		key := priceData.ID.Denom + "/" + priceData.ID.BaseDenom + "/" + string(rune(priceData.ID.Source))
		if !seenPairs[key] {
			seenPairs[key] = true
			latestHeight = append(latestHeight, types.PriceDataID{
				Source:    priceData.ID.Source,
				Denom:     priceData.ID.Denom,
				BaseDenom: priceData.ID.BaseDenom,
			})
		}
	}

	return &types.GenesisState{
		Params:       params,
		Prices:       prices,
		LatestHeight: latestHeight,
	}
}
