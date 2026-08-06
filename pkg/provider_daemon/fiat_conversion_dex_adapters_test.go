// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
)

type providerBoundPoolState struct {
	response dex.BoundOsmosisPoolResponse
}

func (p providerBoundPoolState) PoolState(context.Context, string) (dex.BoundOsmosisPoolResponse, error) {
	return p.response, nil
}

func (p providerBoundPoolState) PoolStates(context.Context, int) (dex.BoundOsmosisPoolResponse, error) {
	return p.response, nil
}

type providerChainEvidence struct{ observation dex.ChainObservation }

func (p providerChainEvidence) SourceID() string { return p.observation.SourceID }

func (p providerChainEvidence) LatestObservation() (dex.ChainObservation, error) {
	return p.observation, nil
}

func (p providerChainEvidence) BlockHash(uint64) (string, error) {
	return p.observation.BlockHash, nil
}

type providerExecutionVerifier struct{}

func (providerExecutionVerifier) VerifySignedExecution(context.Context, []byte, []byte) error {
	return nil
}

func TestNewBoundOsmosisAdapterRequiresBoundPoolProvider(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	profile := providerOsmosisProfile()
	cfg := dex.AdapterConfig{
		Name: "provider-osmosis", Type: "osmosis", Enabled: true, EngineeringTestMode: true,
		RESTEndpoint: "http://example.invalid", RouteProfile: &profile, Now: func() time.Time { return now },
	}
	observation := dex.ChainObservation{ChainID: dex.OsmosisChainIDMainnet, SourceID: "node-a", Height: 100, BlockHash: "AABB", ObservedAt: now}
	chain := providerChainEvidence{observation: observation}
	oracle := OraclePriceFunc(func(string, string, uint64) (sdkmath.LegacyDec, error) { return sdkmath.LegacyOneDec(), nil })

	_, err := NewBoundOsmosisAdapter(cfg, nil, chain, oracle, providerExecutionVerifier{})
	require.ErrorContains(t, err, "bound Osmosis pool")

	payload, err := json.Marshal(dex.OsmosisPoolsResponse{})
	require.NoError(t, err)
	poolState := providerBoundPoolState{response: dex.BoundOsmosisPoolResponse{
		Payload: payload, ResponseHeight: observation.Height, ResponseBlockHash: observation.BlockHash, ResponseSourceID: observation.SourceID, Observation: observation,
	}}
	adapter, err := NewBoundOsmosisAdapter(cfg, poolState, chain, oracle, providerExecutionVerifier{})
	require.NoError(t, err)
	require.IsType(t, &dex.RealOsmosisAdapter{}, adapter)
}

func providerOsmosisProfile() dex.DEXRouteProfile {
	return dex.DEXRouteProfile{
		ID: "provider-osmosis-engineering", State: dex.RouteEngineeringCompleteExternalBlocked,
		Network: "osmosis", ChainID: dex.OsmosisChainIDMainnet, Environment: dex.EnvironmentDevelopment,
		DEX: "osmosis", Version: dex.OsmosisAdapterVersion, AllowedPoolIDs: []string{"1"},
		Tokens:              []dex.RouteToken{{Symbol: "OSMO", Denom: "uosmo", Decimals: 6}, {Symbol: "ATOM", Denom: "uatom", Decimals: 6}},
		FinalityBlocks:      2,
		MaxObservationAge:   time.Minute,
		MaxHeightLag:        2,
		MaxHops:             1,
		MinLiquidity:        sdkmath.NewInt(1),
		MinReserve:          sdkmath.NewInt(1),
		MaxAmount:           sdkmath.NewInt(1_000_000),
		MaxPriceImpact:      sdkmath.LegacyMustNewDecFromStr("0.5"),
		MaxOracleDeviation:  sdkmath.LegacyMustNewDecFromStr("0.5"),
		QuoteTTL:            time.Minute,
		CustodyMode:         "test-custody",
		OracleSource:        "test-oracle",
		EngineeringTestOnly: true,
	}
}

var _ dex.OsmosisPoolStateProvider = providerBoundPoolState{}
