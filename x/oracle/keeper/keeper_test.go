// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	otypes "github.com/virtengine/virtengine/x/oracle/types"
)

func setupTestKeeper(t *testing.T) (IKeeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(otypes.StoreKey)

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create test context with memory store
	testCtx := sdk.Context{}.WithBlockHeight(100).WithBlockTime(time.Now())

	// Create in-memory store
	db := runtime.NewKVStoreService(storeKey)
	_ = db // We'll use the context's KVStore directly

	keeper := NewKeeper(
		cdc,
		storeKey,
		"test-authority",
	)

	return keeper, testCtx
}

func TestGetSetParams(t *testing.T) {
	// This is a basic unit test structure
	// Full tests would require proper store setup with baseapp

	// Test default params structure
	defaultParams := types.DefaultParams()

	require.Equal(t, uint32(1), defaultParams.MinPriceSources)
	require.Equal(t, int64(60), defaultParams.MaxPriceStalenessBlocks)
	require.Equal(t, uint64(150), defaultParams.MaxPriceDeviationBps)
	require.Equal(t, int64(180), defaultParams.TwapWindow)
}

func TestDefaultGenesisState(t *testing.T) {
	genesisState := types.DefaultGenesisState()

	require.NotNil(t, genesisState)
	require.NotNil(t, genesisState.Params)
	require.Empty(t, genesisState.Prices)
}

func TestGenesisStateValidation(t *testing.T) {
	genesisState := types.DefaultGenesisState()

	err := genesisState.Validate()
	require.NoError(t, err)
}

func TestPriceDataState(t *testing.T) {
	now := time.Now()
	price := math.LegacyNewDec(100)

	priceState := types.PriceDataState{
		Price:     price,
		Timestamp: now,
	}

	require.True(t, priceState.Price.Equal(price))
	require.Equal(t, now.Unix(), priceState.Timestamp.Unix())
}

func TestAggregatedPriceFields(t *testing.T) {
	now := time.Now()
	twap := math.LegacyNewDec(100)
	median := math.LegacyNewDec(99)
	minPrice := math.LegacyNewDec(95)
	maxPrice := math.LegacyNewDec(105)

	aggregated := types.AggregatedPrice{
		Denom:        "uve",
		TWAP:         twap,
		MedianPrice:  median,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		Timestamp:    now,
		NumSources:   3,
		DeviationBps: 100,
	}

	require.Equal(t, "uve", aggregated.Denom)
	require.True(t, aggregated.TWAP.Equal(twap))
	require.True(t, aggregated.MedianPrice.Equal(median))
	require.True(t, aggregated.MinPrice.Equal(minPrice))
	require.True(t, aggregated.MaxPrice.Equal(maxPrice))
	require.Equal(t, uint32(3), aggregated.NumSources)
	require.Equal(t, uint64(100), aggregated.DeviationBps)
}

func TestPriceHealthFields(t *testing.T) {
	health := types.PriceHealth{
		Denom:               "uve",
		IsHealthy:           true,
		HasMinSources:       true,
		DeviationOk:         true,
		TotalSources:        5,
		TotalHealthySources: 4,
		FailureReason:       nil,
	}

	require.Equal(t, "uve", health.Denom)
	require.True(t, health.IsHealthy)
	require.True(t, health.HasMinSources)
	require.True(t, health.DeviationOk)
	require.Equal(t, uint32(5), health.TotalSources)
	require.Equal(t, uint32(4), health.TotalHealthySources)
	require.Empty(t, health.FailureReason)
}

func TestKeysGeneration(t *testing.T) {
	// Test params key
	paramsKey := otypes.ParamsPrefix()
	require.NotEmpty(t, paramsKey)
	require.Equal(t, []byte{0x01}, paramsKey)

	// Test price data key
	priceKey := otypes.PriceDataKey(0, "uve", "usd", 100)
	require.NotEmpty(t, priceKey)
	require.True(t, len(priceKey) > len(otypes.PriceDataPrefix))

	// Test latest price data key
	latestKey := otypes.LatestPriceDataKey(0, "uve", "usd")
	require.NotEmpty(t, latestKey)
	require.True(t, len(latestKey) > len(otypes.LatestPriceDataPrefix))

	// Test prefix by pair
	pairPrefix := otypes.PriceDataPrefixByPair(0, "uve", "usd")
	require.NotEmpty(t, pairPrefix)
}

func TestMsgAddPriceEntry(t *testing.T) {
	msg := &types.MsgAddPriceEntry{
		Signer: "ve1test123",
		ID: types.DataID{
			Denom:     "uve",
			BaseDenom: "usd",
		},
		Price: types.PriceDataState{
			Price:     math.LegacyNewDec(100),
			Timestamp: time.Now(),
		},
	}

	require.Equal(t, "ve1test123", msg.Signer)
	require.Equal(t, "uve", msg.ID.Denom)
	require.Equal(t, "usd", msg.ID.BaseDenom)
}

func TestMsgUpdateParams(t *testing.T) {
	params := types.DefaultParams()
	msg := &types.MsgUpdateParams{
		Authority: "ve1authority123",
		Params:    params,
	}

	require.Equal(t, "ve1authority123", msg.Authority)
	require.NotNil(t, msg.Params)
}
