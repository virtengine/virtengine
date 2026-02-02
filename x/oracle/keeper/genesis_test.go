// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
)

func TestGenesisStateSerialization(t *testing.T) {
	// Test that genesis state can be properly serialized and deserialized
	genesis := types.DefaultGenesisState()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.Empty(t, genesis.Prices)
	require.Empty(t, genesis.LatestHeight)
}

func TestGenesisStateWithPrices(t *testing.T) {
	now := time.Now()
	price := math.LegacyNewDec(100)

	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Prices: []types.PriceData{
			{
				ID: types.PriceDataRecordID{
					Source:    0,
					Denom:     "uve",
					BaseDenom: "usd",
					Height:    100,
				},
				State: types.PriceDataState{
					Price:     price,
					Timestamp: now,
				},
			},
		},
		LatestHeight: []types.PriceDataID{
			{
				Source:    0,
				Denom:     "uve",
				BaseDenom: "usd",
			},
		},
	}

	require.NotNil(t, genesis)
	require.Len(t, genesis.Prices, 1)
	require.Len(t, genesis.LatestHeight, 1)
	require.Equal(t, "uve", genesis.Prices[0].ID.Denom)
	require.Equal(t, "usd", genesis.Prices[0].ID.BaseDenom)
	require.True(t, genesis.Prices[0].State.Price.Equal(price))
}

func TestGenesisStateValidationWithInvalidParams(t *testing.T) {
	genesis := types.DefaultGenesisState()

	// Default params should be valid
	err := genesis.Validate()
	require.NoError(t, err)

	// Modify to invalid params
	genesis.Params.MinPriceSources = 0
	err = genesis.Validate()
	require.Error(t, err)
}
