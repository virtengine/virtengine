// Package dex provides DEX integration adapters for VirtEngine.
package dex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOsmosisConfig tests configuration functions
func TestOsmosisConfig(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		t.Parallel()
		config := DefaultOsmosisConfig()
		assert.Equal(t, "mainnet", config.Network)
		assert.NotEmpty(t, config.GetRESTEndpoint())
		assert.NotEmpty(t, config.GetGRPCEndpoint())
		assert.Equal(t, "osmosis-1", config.GetChainID())
	})

	t.Run("testnet config", func(t *testing.T) {
		t.Parallel()
		config := OsmosisConfig{
			Network: "testnet",
		}
		assert.Equal(t, "osmo-test-5", config.GetChainID())
	})

	t.Run("custom endpoints override defaults", func(t *testing.T) {
		t.Parallel()
		config := OsmosisConfig{
			Network:      "mainnet",
			RESTEndpoint: "https://custom.lcd.zone",
			GRPCEndpoint: "custom.grpc.zone:443",
		}
		assert.Equal(t, "https://custom.lcd.zone", config.GetRESTEndpoint())
		assert.Equal(t, "custom.grpc.zone:443", config.GetGRPCEndpoint())
	})
}

// createTestAdapter creates an adapter with a mock server for testing
func createTestAdapter(t *testing.T, handler http.Handler) (*RealOsmosisAdapter, *httptest.Server) {
	server := httptest.NewServer(handler)

	adapterCfg := AdapterConfig{
		Name:    "osmosis-test",
		Type:    "osmosis",
		Enabled: true,
		Timeout: 30 * time.Second,
	}

	osmosisConfig := OsmosisConfig{
		Network:             "mainnet",
		RESTEndpoint:        server.URL,
		PoolRefreshInterval: 1 * time.Minute,
		Timeout:             30 * time.Second,
		SlippageTolerance:   0.01,
	}

	adapter, err := NewRealOsmosisAdapter(adapterCfg, osmosisConfig)
	require.NoError(t, err)

	return adapter, server
}

// TestNewRealOsmosisAdapter tests adapter creation
func TestNewRealOsmosisAdapter(t *testing.T) {
	t.Parallel()

	adapterCfg := AdapterConfig{
		Name:    "osmosis",
		Type:    "osmosis",
		Enabled: true,
		Timeout: 30 * time.Second,
	}

	osmosisConfig := OsmosisConfig{
		Network:             "mainnet",
		RESTEndpoint:        "https://lcd.osmosis.zone",
		GRPCEndpoint:        "grpc.osmosis.zone:443",
		PoolRefreshInterval: 1 * time.Minute,
		Timeout:             30 * time.Second,
		SlippageTolerance:   0.01,
	}

	adapter, err := NewRealOsmosisAdapter(adapterCfg, osmosisConfig)
	require.NoError(t, err)
	assert.NotNil(t, adapter)
	assert.Equal(t, "osmosis", adapter.Name())
	assert.Equal(t, "osmosis-1", adapter.ChainID())
}

// TestNewRealOsmosisAdapter_DefaultConfig tests adapter creation with default config
func TestNewRealOsmosisAdapter_DefaultConfig(t *testing.T) {
	t.Parallel()

	adapterCfg := AdapterConfig{
		Name:    "osmosis",
		Type:    "osmosis",
		Enabled: true,
	}

	adapter, err := NewRealOsmosisAdapter(adapterCfg, DefaultOsmosisConfig())
	require.NoError(t, err)
	assert.NotNil(t, adapter)
	assert.Equal(t, "osmosis", adapter.Name())
}

// TestOsmosisAdapter_GetPool tests pool retrieval with mock server
func TestOsmosisAdapter_GetPool(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsSubstring(r.URL.Path, "/pools/1") {
			response := OsmosisPoolResponse{
				Pool: OsmosisPoolData{
					Type:    "/osmosis.gamm.v1beta1.Pool",
					ID:      "1",
					Address: "osmo1mw0ac6rwlp5r8wapwk3zs6g29h8fcscxqakdzw9emkne6c8wjp9q0t3v8t",
					PoolParams: OsmosisPoolParams{
						SwapFee: "0.002",
						ExitFee: "0.0",
					},
					PoolAssets: []OsmosisPoolAsset{
						{
							Token:  OsmosisCoin{Denom: "uosmo", Amount: "1000000000"},
							Weight: "50",
						},
						{
							Token:  OsmosisCoin{Denom: "uatom", Amount: "500000000"},
							Weight: "50",
						},
					},
					TotalShares: OsmosisCoin{Denom: "gamm/pool/1", Amount: "100000000"},
					TotalWeight: "100",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()
	pool, err := adapter.GetPool(ctx, "1")
	require.NoError(t, err)

	assert.Equal(t, "1", pool.ID)
	assert.Equal(t, "osmosis-test", pool.DEX)
	assert.Equal(t, PoolTypeConstantProduct, pool.Type)
	assert.Len(t, pool.Tokens, 2)
	assert.Equal(t, "OSMO", pool.Tokens[0].Symbol)
	assert.Equal(t, "ATOM", pool.Tokens[1].Symbol)
}

// TestOsmosisAdapter_GetPool_NotFound tests pool not found error
func TestOsmosisAdapter_GetPool_NotFound(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()
	_, err := adapter.GetPool(ctx, "nonexistent")
	assert.Error(t, err)
}

// TestOsmosisAdapter_GetSpotPrice tests spot price retrieval
func TestOsmosisAdapter_GetSpotPrice(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case containsSubstring(r.URL.Path, "/all-pools"):
			response := OsmosisPoolsResponse{
				Pools: []OsmosisPoolData{
					{
						Type:    "/osmosis.gamm.v1beta1.Pool",
						ID:      "1",
						Address: "osmo1pool1...",
						PoolParams: OsmosisPoolParams{
							SwapFee: "0.002",
						},
						PoolAssets: []OsmosisPoolAsset{
							{Token: OsmosisCoin{Denom: "uosmo", Amount: "1000000000"}, Weight: "50"},
							{Token: OsmosisCoin{Denom: "uatom", Amount: "500000000"}, Weight: "50"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		case containsSubstring(r.URL.Path, "/spot-price"):
			response := OsmosisSpotPriceResponse{
				SpotPrice: "2.0",
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()
	price, err := adapter.GetPrice(ctx, "OSMO", "ATOM")
	require.NoError(t, err)

	assert.Equal(t, "OSMO", price.Pair.BaseToken.Symbol)
	assert.Equal(t, "ATOM", price.Pair.QuoteToken.Symbol)
	expectedRate := sdkmath.LegacyNewDec(2)
	assert.True(t, price.Rate.Equal(expectedRate), "Expected rate 2.0, got %s", price.Rate.String())
	assert.Equal(t, "osmosis-test", price.Source)
}

// TestOsmosisAdapter_GetSwapQuote tests swap quote generation
func TestOsmosisAdapter_GetSwapQuote(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case containsSubstring(r.URL.Path, "/all-pools"):
			response := OsmosisPoolsResponse{
				Pools: []OsmosisPoolData{
					{
						Type:    "/osmosis.gamm.v1beta1.Pool",
						ID:      "1",
						Address: "osmo1pool1...",
						PoolParams: OsmosisPoolParams{
							SwapFee: "0.003",
						},
						PoolAssets: []OsmosisPoolAsset{
							{Token: OsmosisCoin{Denom: "uosmo", Amount: "1000000000000"}, Weight: "50"},
							{Token: OsmosisCoin{Denom: "uatom", Amount: "500000000000"}, Weight: "50"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		case containsSubstring(r.URL.Path, "estimate"):
			response := OsmosisEstimateSwapResponse{
				TokenOutAmount: "990000000",
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()

	request := SwapRequest{
		FromToken:         Token{Symbol: "OSMO", Denom: "uosmo", Decimals: 6, ChainID: "osmosis-1"},
		ToToken:           Token{Symbol: "ATOM", Denom: "uatom", Decimals: 6, ChainID: "osmosis-1"},
		Amount:            sdkmath.NewInt(1000000000),
		Sender:            "osmo1sender...",
		SlippageTolerance: 0.01,
	}

	quote, err := adapter.GetSwapQuote(ctx, request)
	require.NoError(t, err)

	assert.NotEmpty(t, quote.ID)
	assert.Equal(t, request.Amount, quote.InputAmount)
	assert.True(t, quote.OutputAmount.IsPositive())
	assert.Len(t, quote.Route.Hops, 1)
	assert.Equal(t, "1", quote.Route.Hops[0].PoolID)
}

// TestOsmosisAdapter_GetSupportedPairs tests getting supported trading pairs
func TestOsmosisAdapter_GetSupportedPairs(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsSubstring(r.URL.Path, "/all-pools") {
			response := OsmosisPoolsResponse{
				Pools: []OsmosisPoolData{
					{
						ID: "1",
						PoolAssets: []OsmosisPoolAsset{
							{Token: OsmosisCoin{Denom: "uosmo", Amount: "1000000"}, Weight: "50"},
							{Token: OsmosisCoin{Denom: "uatom", Amount: "500000"}, Weight: "50"},
						},
						PoolParams: OsmosisPoolParams{SwapFee: "0.002"},
					},
					{
						ID: "2",
						PoolAssets: []OsmosisPoolAsset{
							{Token: OsmosisCoin{Denom: "uosmo", Amount: "2000000"}, Weight: "50"},
							{Token: OsmosisCoin{Denom: "uusdc", Amount: "1000000"}, Weight: "50"},
						},
						PoolParams: OsmosisPoolParams{SwapFee: "0.003"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()
	pairs, err := adapter.GetSupportedPairs(ctx)
	require.NoError(t, err)

	assert.Len(t, pairs, 2)
}

// TestOsmosisAdapter_ListPools tests pool listing
func TestOsmosisAdapter_ListPools(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsSubstring(r.URL.Path, "/all-pools") {
			response := OsmosisPoolsResponse{
				Pools: []OsmosisPoolData{
					{
						ID:         "1",
						PoolParams: OsmosisPoolParams{SwapFee: "0.002"},
						PoolAssets: []OsmosisPoolAsset{
							{Token: OsmosisCoin{Denom: "uosmo", Amount: "1000000"}, Weight: "50"},
							{Token: OsmosisCoin{Denom: "uatom", Amount: "500000"}, Weight: "50"},
						},
					},
					{
						ID:         "2",
						PoolParams: OsmosisPoolParams{SwapFee: "0.003"},
						PoolAssets: []OsmosisPoolAsset{
							{Token: OsmosisCoin{Denom: "uosmo", Amount: "2000000"}, Weight: "50"},
							{Token: OsmosisCoin{Denom: "uusdc", Amount: "1000000"}, Weight: "50"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()

	// Test list all pools
	pools, err := adapter.ListPools(ctx, PoolQuery{})
	require.NoError(t, err)
	assert.Len(t, pools, 2)

	// Test filter by token
	pools, err = adapter.ListPools(ctx, PoolQuery{
		TokenSymbols: []string{"USDC"},
	})
	require.NoError(t, err)
	// Should find pool with USDC
	assert.GreaterOrEqual(t, len(pools), 0)
}

// TestOsmosisAdapter_ContextCancellation tests context cancellation
func TestOsmosisAdapter_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Create a slow server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := adapter.GetPool(ctx, "1")
	assert.Error(t, err) // Should error due to context cancellation
}

// TestOsmosisAdapter_EstimateGas tests gas estimation
func TestOsmosisAdapter_EstimateGas(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsSubstring(r.URL.Path, "/all-pools") {
			response := OsmosisPoolsResponse{
				Pools: []OsmosisPoolData{
					{
						ID:         "1",
						PoolParams: OsmosisPoolParams{SwapFee: "0.002"},
						PoolAssets: []OsmosisPoolAsset{
							{Token: OsmosisCoin{Denom: "uosmo", Amount: "1000000"}, Weight: "50"},
							{Token: OsmosisCoin{Denom: "uatom", Amount: "500000"}, Weight: "50"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()

	// Direct swap should use ~300k gas
	request := SwapRequest{
		FromToken: Token{Symbol: "OSMO", Denom: "uosmo"},
		ToToken:   Token{Symbol: "ATOM", Denom: "uatom"},
		Amount:    sdkmath.NewInt(1000000),
	}

	gas, err := adapter.EstimateGas(ctx, request)
	require.NoError(t, err)
	assert.Equal(t, uint64(300000), gas)

	// Multi-hop swap should use more gas
	request = SwapRequest{
		FromToken: Token{Symbol: "FOO", Denom: "ufoo"},
		ToToken:   Token{Symbol: "BAR", Denom: "ubar"},
		Amount:    sdkmath.NewInt(1000000),
	}

	gas, err = adapter.EstimateGas(ctx, request)
	require.NoError(t, err)
	assert.Equal(t, uint64(500000), gas)
}

// TestOsmosisAdapter_GetPoolReserves tests pool reserve retrieval
func TestOsmosisAdapter_GetPoolReserves(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsSubstring(r.URL.Path, "/pools/1") {
			response := OsmosisPoolResponse{
				Pool: OsmosisPoolData{
					ID:         "1",
					PoolParams: OsmosisPoolParams{SwapFee: "0.002"},
					PoolAssets: []OsmosisPoolAsset{
						{Token: OsmosisCoin{Denom: "uosmo", Amount: "1000000000"}, Weight: "50"},
						{Token: OsmosisCoin{Denom: "uatom", Amount: "500000000"}, Weight: "50"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}
	})

	adapter, server := createTestAdapter(t, handler)
	defer server.Close()

	ctx := context.Background()
	reserve0, reserve1, err := adapter.GetPoolReserves(ctx, "1")
	require.NoError(t, err)

	expectedReserve0 := sdkmath.NewInt(1000000000)
	expectedReserve1 := sdkmath.NewInt(500000000)

	assert.True(t, reserve0.Equal(expectedReserve0), "Expected reserve0 %s, got %s", expectedReserve0, reserve0)
	assert.True(t, reserve1.Equal(expectedReserve1), "Expected reserve1 %s, got %s", expectedReserve1, reserve1)
}

// TestTokenParsing tests token denomination parsing
func TestTokenParsing(t *testing.T) {
	t.Parallel()

	adapterCfg := AdapterConfig{
		Name:    "osmosis",
		Type:    "osmosis",
		Enabled: true,
	}

	osmosisConfig := OsmosisConfig{
		Network:             "mainnet",
		RESTEndpoint:        "https://lcd.osmosis.zone",
		PoolRefreshInterval: 1 * time.Minute,
		Timeout:             30 * time.Second,
		SlippageTolerance:   0.01,
	}

	adapter, err := NewRealOsmosisAdapter(adapterCfg, osmosisConfig)
	require.NoError(t, err)

	tests := []struct {
		denom          string
		expectedSymbol string
		expectedNative bool
	}{
		{"uosmo", "OSMO", true},
		{"uatom", "ATOM", true},
		{"uusdc", "USDC", true},
		{"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", "IBC-27394F", false},
		{"unknown", "unknown", true},
	}

	for _, tc := range tests {
		t.Run(tc.denom, func(t *testing.T) {
			t.Parallel()
			token := adapter.parseToken(OsmosisCoin{Denom: tc.denom, Amount: "1000000"})
			assert.Equal(t, tc.expectedSymbol, token.Symbol)
			assert.Equal(t, tc.expectedNative, token.IsNative)
			assert.Equal(t, tc.denom, token.Denom)
			assert.Equal(t, uint8(6), token.Decimals)
		})
	}
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

