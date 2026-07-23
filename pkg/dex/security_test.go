package dex

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/require"
)

type rejectingRouteAuthorizer struct{}

func (rejectingRouteAuthorizer) AuthorizeDEXRoute(DEXRouteProfile) error {
	return errors.New("untrusted profile")
}

type staticBoundPoolProvider struct {
	discovery BoundOsmosisPoolResponse
	pool      BoundOsmosisPoolResponse
	err       error
}

func (p staticBoundPoolProvider) PoolState(context.Context, string) (BoundOsmosisPoolResponse, error) {
	if p.err != nil {
		return BoundOsmosisPoolResponse{}, p.err
	}
	return p.pool, nil
}

func (p staticBoundPoolProvider) PoolStates(context.Context, int) (BoundOsmosisPoolResponse, error) {
	if p.err != nil {
		return BoundOsmosisPoolResponse{}, p.err
	}
	return p.discovery, nil
}

func TestProductionFactoryUsesRealOsmosisAndRejectsStubs(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	profile := testOsmosisProfile()
	cfg := AdapterConfig{
		Name: "production-factory", Type: dexOsmosis, Enabled: true, RESTEndpoint: "https://lcd.example.invalid",
		RouteProfile: profile, EngineeringTestMode: true,
		PoolState:     newRESTBoundPoolStateProvider("https://lcd.example.invalid", http.DefaultClient, ChainObservation{ChainID: OsmosisChainIDMainnet, SourceID: testPoolStateSourceID, Height: 100, BlockHash: "AABB", ObservedAt: now}),
		ChainEvidence: staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, SourceID: testPoolStateSourceID, Height: 100, BlockHash: "AABB", ObservedAt: now}},
		Oracle:        staticOracle{price: sdkmath.LegacyOneDec()}, ExecutionVerifier: staticExecutionVerifier{}, Now: func() time.Time { return now },
	}
	adapter, err := CreateAdapter(cfg)
	require.NoError(t, err)
	require.IsType(t, &RealOsmosisAdapter{}, adapter)

	for _, adapterType := range []string{"uniswap_v2", "curve"} {
		cfg.Type = adapterType
		_, err = CreateAdapter(cfg)
		require.Error(t, err)
	}

	cfg.Type = dexOsmosis
	cfg.EngineeringTestMode = false
	_, err = CreateAdapter(cfg)
	require.ErrorIs(t, err, ErrRouteNotCertified)
}

func TestServiceConfigurationCannotReachPlaceholderAdapter(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Adapters = []AdapterConfig{{Name: "placeholder", Type: "uniswap_v2", Enabled: true, ContractAddresses: map[string]string{"factory": "0x1", "router": "0x2"}}}
	_, err := NewService(cfg)
	require.Error(t, err)
}

func TestRuntimeServiceRejectsManualPlaceholderRegistration(t *testing.T) {
	svc, err := NewService(DefaultConfig())
	require.NoError(t, err)
	placeholder, err := NewOsmosisAdapter(AdapterConfig{Name: "placeholder", Type: dexOsmosis, Enabled: true})
	require.NoError(t, err)
	require.Error(t, svc.RegisterAdapter(placeholder))

	testSvc, err := NewTestService(DefaultConfig())
	require.NoError(t, err)
	require.NoError(t, testSvc.RegisterAdapter(placeholder))
}

func TestDEXRouteProfileFailClosedMatrix(t *testing.T) {
	profile := testOsmosisProfile()
	require.NoError(t, profile.Validate(RouteValidationEngineering))
	require.Error(t, profile.Validate(RouteValidationRuntime))

	profile.State = RouteCertifiedEnabled
	profile.Environment = EnvironmentMainnet
	profile.EngineeringTestOnly = false
	require.Error(t, profile.Validate(RouteValidationRuntime), "certified rows require complete evidence")
	profile.Evidence = RouteEvidence{
		EngineeringEvidence: "sha256:engineering", NetworkEvidence: "sha256:network", LiquidityEvidence: "sha256:liquidity",
		OracleEvidence: "sha256:oracle", CustodyEvidence: "sha256:custody", GovernanceEvidence: "proposal:1",
		EngineeringOwners: []string{"dex-team"}, OperationsOwners: []string{"operations"}, SecurityOwners: []string{"security"},
	}
	require.NoError(t, profile.Validate(RouteValidationRuntime))

	profile.ChainID = "TBD"
	require.Error(t, profile.Validate(RouteValidationRuntime))

	profile = testOsmosisProfile()
	profile.AllowedPoolIDs = []string{"01"}
	require.Error(t, profile.Validate(RouteValidationEngineering))
}

func TestRuntimeOsmosisRequiresTrustedProfileAuthority(t *testing.T) {
	profile := *testOsmosisProfile()
	profile.State = RouteCertifiedEnabled
	profile.Environment = EnvironmentMainnet
	profile.EngineeringTestOnly = false
	profile.Evidence = RouteEvidence{
		EngineeringEvidence: "evidence:engineering", NetworkEvidence: "evidence:network", LiquidityEvidence: "evidence:liquidity",
		OracleEvidence: "evidence:oracle", CustodyEvidence: "evidence:custody", GovernanceEvidence: "evidence:governance",
		EngineeringOwners: []string{"engineering"}, OperationsOwners: []string{"operations"}, SecurityOwners: []string{"security"},
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	config := DefaultOsmosisConfig()
	config.RouteProfile = &profile
	config.PoolState = newRESTBoundPoolStateProvider(config.GetRESTEndpoint(), http.DefaultClient, ChainObservation{ChainID: OsmosisChainIDMainnet, SourceID: testPoolStateSourceID, Height: 100, BlockHash: "AABB", ObservedAt: now})
	config.ChainEvidence = staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, SourceID: testPoolStateSourceID, Height: 100, BlockHash: "AABB", ObservedAt: now}}
	config.Oracle = staticOracle{price: sdkmath.LegacyOneDec()}
	config.ExecutionVerifier = staticExecutionVerifier{}
	config.Now = func() time.Time { return now }
	_, err := NewRealOsmosisAdapter(AdapterConfig{Name: "runtime", Type: dexOsmosis, Enabled: true}, config)
	require.Error(t, err)
	config.RouteAuthorizer = rejectingRouteAuthorizer{}
	_, err = NewRealOsmosisAdapter(AdapterConfig{Name: "runtime", Type: dexOsmosis, Enabled: true}, config)
	require.Error(t, err)
}

func TestConstantProductExactInGolden(t *testing.T) {
	output, err := ConstantProductExactIn(
		sdkmath.NewInt(1_000_000_000), sdkmath.NewInt(500_000_000), sdkmath.NewInt(10_000_000),
		sdkmath.LegacyMustNewDecFromStr("0.003"),
	)
	require.NoError(t, err)
	require.Equal(t, "4935790", output.String())
}

func TestRealOsmosisRejectsUnsupportedWeightedPoolVariant(t *testing.T) {
	pool := securePool("1", "1000000000", "1000000000")
	pool.PoolAssets[1].Weight = "20"
	endpoint := secureQuoteServer(t, pool)
	adapter, server, now := newSecureQuoteAdapter(t, endpoint, sdkmath.LegacyOneDec())
	defer server.Close()
	_, err := adapter.GetSwapQuote(context.Background(), secureSwapRequest(now))
	require.ErrorIs(t, err, ErrOsmosisInvalidPool)
}

func TestDeterministicQuoteDigestAndPayload(t *testing.T) {
	adapter, server, now := newSecureQuoteAdapter(t, secureQuoteServer(t, securePool("1", "1000000000", "1000000000")), sdkmath.LegacyMustNewDecFromStr("0.987158034397061298"))
	defer server.Close()
	request := secureSwapRequest(now)

	first, err := adapter.GetSwapQuote(context.Background(), request)
	require.NoError(t, err)
	second, err := adapter.GetSwapQuote(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, first.QuoteDigest, second.QuoteDigest)
	timeChanged := first
	timeChanged.CreatedAt = timeChanged.CreatedAt.Add(time.Second)
	timeChanged.ExpiresAt = timeChanged.ExpiresAt.Add(time.Second)
	digest, err := QuoteDigest(timeChanged)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, digest)

	firstPayload, err := BuildExecutionPayload(first)
	require.NoError(t, err)
	secondPayload, err := BuildExecutionPayload(second)
	require.NoError(t, err)
	require.Equal(t, firstPayload, secondPayload)

	tampered := first
	tampered.MinOutputAmount = tampered.MinOutputAmount.AddRaw(1)
	_, err = BuildExecutionPayload(tampered)
	require.ErrorIs(t, err, ErrExecutionPayload)
}

func TestSecureQuoteRejectsDecimalsLiquidityPoolImpactOracleAndAmount(t *testing.T) {
	tests := []struct {
		name     string
		pool     OsmosisPoolData
		mutate   func(*DEXRouteProfile, *SwapRequest)
		oracle   sdkmath.LegacyDec
		expected error
	}{
		{name: "decimal mismatch", pool: securePool("1", "1000000000", "1000000000"), mutate: func(_ *DEXRouteProfile, req *SwapRequest) { req.FromToken.Decimals = 8 }, oracle: sdkmath.LegacyOneDec(), expected: ErrTokenDecimals},
		{name: "zero liquidity", pool: securePool("1", "0", "1000000000"), oracle: sdkmath.LegacyOneDec(), expected: ErrOsmosisInvalidPool},
		{name: "unallowlisted pool", pool: securePool("9", "1000000000", "1000000000"), oracle: sdkmath.LegacyOneDec(), expected: ErrOsmosisPoolNotFound},
		{name: "amount limit", pool: securePool("1", "1000000000", "1000000000"), mutate: func(profile *DEXRouteProfile, _ *SwapRequest) { profile.MaxAmount = sdkmath.NewInt(1) }, oracle: sdkmath.LegacyOneDec(), expected: ErrAmountTooLarge},
		{name: "high impact", pool: securePool("1", "1000000", "1000000"), mutate: func(profile *DEXRouteProfile, _ *SwapRequest) {
			profile.MaxPriceImpact = sdkmath.LegacyMustNewDecFromStr("0.01")
			profile.MinReserve = sdkmath.NewInt(1)
			profile.MinLiquidity = sdkmath.NewInt(1)
		}, oracle: sdkmath.LegacyOneDec(), expected: ErrPriceImpactExceeded},
		{name: "oracle deviation", pool: securePool("1", "1000000000", "1000000000"), mutate: func(profile *DEXRouteProfile, _ *SwapRequest) {
			profile.MaxOracleDeviation = sdkmath.LegacyMustNewDecFromStr("0.01")
		}, oracle: sdkmath.LegacyMustNewDecFromStr("2"), expected: ErrOracleDeviation},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Unix(1_700_000_000, 0).UTC()
			profile := testOsmosisProfile()
			profile.MaxPriceImpact = sdkmath.LegacyMustNewDecFromStr("0.99")
			profile.MaxOracleDeviation = sdkmath.LegacyMustNewDecFromStr("0.99")
			request := secureSwapRequest(now)
			if tc.mutate != nil {
				tc.mutate(profile, &request)
			}
			adapter, server := secureAdapterWithProfile(t, secureQuoteServer(t, tc.pool), now, profile, tc.oracle, staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, Height: 100, BlockHash: "AABB", ObservedAt: now}})
			defer server.Close()
			_, err := adapter.GetSwapQuote(context.Background(), request)
			require.ErrorIs(t, err, tc.expected)
		})
	}
}

func TestSecureQuoteRejectsWrongOrStaleChainEvidence(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	pool := securePool("1", "1000000000", "1000000000")
	profile := testOsmosisProfile()
	profile.MaxPriceImpact = sdkmath.LegacyMustNewDecFromStr("0.99")

	t.Run("wrong chain", func(t *testing.T) {
		adapter, server := secureAdapterWithProfile(t, secureQuoteServer(t, pool), now, profile, sdkmath.LegacyOneDec(), staticChainEvidence{observation: ChainObservation{ChainID: "wrong-chain", Height: 100, BlockHash: "AABB", ObservedAt: now}})
		defer server.Close()
		_, err := adapter.GetSwapQuote(context.Background(), secureSwapRequest(now))
		require.ErrorIs(t, err, ErrWrongChain)
	})

	t.Run("stale observation", func(t *testing.T) {
		adapter, server := secureAdapterWithProfile(t, secureQuoteServer(t, pool), now, profile, sdkmath.LegacyOneDec(), staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, Height: 100, BlockHash: "AABB", ObservedAt: now.Add(-2 * time.Minute)}})
		defer server.Close()
		_, err := adapter.GetSwapQuote(context.Background(), secureSwapRequest(now))
		require.ErrorIs(t, err, ErrPoolStateStale)
	})
}

func TestBoundOsmosisPoolStateEvidenceDiscoveryAndSinglePool(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	observation := ChainObservation{ChainID: OsmosisChainIDMainnet, SourceID: "node-a", Height: 100, BlockHash: "AABB", ObservedAt: now}
	pool := securePool("1", "1000000000", "500000000")
	discoveryPayload, err := json.Marshal(OsmosisPoolsResponse{Pools: []OsmosisPoolData{pool}})
	require.NoError(t, err)
	poolPayload, err := json.Marshal(OsmosisPoolResponse{Pool: pool})
	require.NoError(t, err)
	bound := staticBoundPoolProvider{
		discovery: BoundOsmosisPoolResponse{Payload: discoveryPayload, ResponseHeight: 100, ResponseBlockHash: "AABB", ResponseSourceID: "node-a", Observation: observation},
		pool:      BoundOsmosisPoolResponse{Payload: poolPayload, ResponseHeight: 100, ResponseBlockHash: "AABB", ResponseSourceID: "node-a", Observation: observation},
	}
	adapter := secureAdapterWithBoundProvider(t, now, bound, staticChainEvidence{observation: observation})

	pools, err := adapter.ListPools(context.Background(), PoolQuery{})
	require.NoError(t, err)
	require.Len(t, pools, 1)
	require.Equal(t, uint64(100), pools[0].Height)
	require.Equal(t, "AABB", pools[0].BlockHash)
	require.Equal(t, "node-a", pools[0].SourceID)

	singleAdapter := secureAdapterWithBoundProvider(t, now, bound, staticChainEvidence{observation: observation})
	single, err := singleAdapter.GetPool(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, pools[0].Height, single.Height)
	require.Equal(t, pools[0].BlockHash, single.BlockHash)
	require.Equal(t, pools[0].SourceID, single.SourceID)
}

func TestBoundOsmosisPoolStateEvidenceRejectsInvalidBindings(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	base := ChainObservation{ChainID: OsmosisChainIDMainnet, SourceID: "node-a", Height: 100, BlockHash: "AABB", ObservedAt: now}
	pool := securePool("1", "1000000000", "500000000")
	discoveryPayload, err := json.Marshal(OsmosisPoolsResponse{Pools: []OsmosisPoolData{pool}})
	require.NoError(t, err)
	poolPayload, err := json.Marshal(OsmosisPoolResponse{Pool: pool})
	require.NoError(t, err)

	tests := []struct {
		name     string
		mutate   func(*BoundOsmosisPoolResponse)
		latest   ChainObservation
		expected error
	}{
		{name: "unbound", mutate: func(result *BoundOsmosisPoolResponse) { result.ResponseHeight = 0; result.ResponseSourceID = "" }, latest: base, expected: ErrPoolStateEvidence},
		{name: "height mismatch", mutate: func(result *BoundOsmosisPoolResponse) { result.ResponseHeight = 99 }, latest: base, expected: ErrPoolStateEvidence},
		{name: "hash mismatch", mutate: func(result *BoundOsmosisPoolResponse) { result.ResponseBlockHash = "CCDD" }, latest: base, expected: ErrPoolStateEvidence},
		{name: "stale", mutate: func(result *BoundOsmosisPoolResponse) { result.Observation.ObservedAt = now.Add(-2 * time.Minute) }, latest: base, expected: ErrPoolStateStale},
		{name: "cross node", mutate: func(result *BoundOsmosisPoolResponse) {
			result.ResponseSourceID = "node-b"
			result.Observation.SourceID = "node-b"
		}, latest: base, expected: ErrPoolStateEvidence},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			discovery := BoundOsmosisPoolResponse{Payload: discoveryPayload, ResponseHeight: 100, ResponseBlockHash: "AABB", ResponseSourceID: "node-a", Observation: base}
			single := BoundOsmosisPoolResponse{Payload: poolPayload, ResponseHeight: 100, ResponseBlockHash: "AABB", ResponseSourceID: "node-a", Observation: base}
			tc.mutate(&discovery)
			tc.mutate(&single)
			adapter := secureAdapterWithBoundProvider(t, now, staticBoundPoolProvider{discovery: discovery, pool: single}, staticChainEvidence{observation: tc.latest})
			_, err := adapter.ListPools(context.Background(), PoolQuery{})
			require.ErrorIs(t, err, tc.expected, "discovery path")
			_, err = adapter.GetPool(context.Background(), "1")
			require.ErrorIs(t, err, tc.expected, "single-pool path")
		})
	}
}

func secureAdapterWithBoundProvider(t *testing.T, now time.Time, poolState OsmosisPoolStateProvider, evidence ChainEvidenceProvider) *RealOsmosisAdapter {
	t.Helper()
	profile := testOsmosisProfile()
	config := DefaultOsmosisConfig()
	config.RESTEndpoint = "http://example.invalid"
	config.RouteProfile = profile
	config.ValidationMode = RouteValidationEngineering
	config.PoolState = poolState
	config.ChainEvidence = evidence
	config.Oracle = staticOracle{price: sdkmath.LegacyOneDec()}
	config.ExecutionVerifier = staticExecutionVerifier{}
	config.Now = func() time.Time { return now }
	adapter, err := NewRealOsmosisAdapter(AdapterConfig{Name: "bound-test", Type: dexOsmosis, Enabled: true}, config)
	require.NoError(t, err)
	return adapter
}

func TestRouteValidationRejectsCyclesAndExcessiveHops(t *testing.T) {
	profile := testOsmosisProfile()
	profile.AllowedPoolIDs = []string{"1", "2", "3"}
	a := Token{Symbol: "A", Denom: "a"}
	b := Token{Symbol: "B", Denom: "b"}
	c := Token{Symbol: "C", Denom: "c"}
	cycle := SwapRoute{Hops: []SwapHop{{PoolID: "1", FromToken: a, ToToken: b, AmountIn: sdkmath.NewInt(10), AmountOut: sdkmath.NewInt(9)}, {PoolID: "2", FromToken: b, ToToken: a, AmountIn: sdkmath.NewInt(9), AmountOut: sdkmath.NewInt(8)}}}
	require.ErrorIs(t, validateBoundedRoute(cycle, 3, *profile), ErrRouteCycle)
	excessive := SwapRoute{Hops: []SwapHop{{PoolID: "1", FromToken: a, ToToken: b}, {PoolID: "2", FromToken: b, ToToken: c}, {PoolID: "3", FromToken: c, ToToken: Token{Denom: "d"}}}}
	require.ErrorIs(t, validateBoundedRoute(excessive, 2, *profile), ErrRouteHopsExceeded)
}

func TestExecutionRejectsExpiredTamperedPayloadAndMinimumOutput(t *testing.T) {
	adapter, server, now := newSecureQuoteAdapter(t, secureQuoteServer(t, securePool("1", "1000000000", "1000000000")), sdkmath.LegacyMustNewDecFromStr("0.987158034397061298"))
	defer server.Close()
	quote, err := adapter.GetSwapQuote(context.Background(), secureSwapRequest(now))
	require.NoError(t, err)
	payload, err := BuildExecutionPayload(quote)
	require.NoError(t, err)
	envelope, err := MarshalSignedExecutionEnvelope(payload, []byte("custody-signed-transaction"))
	require.NoError(t, err)

	tamperedEnvelope := envelope
	var decoded SignedExecutionEnvelope
	require.NoError(t, json.Unmarshal(tamperedEnvelope, &decoded))
	decoded.Payload = append([]byte(nil), decoded.Payload...)
	decoded.Payload[0] ^= 1
	tamperedEnvelope, err = json.Marshal(decoded)
	require.NoError(t, err)
	_, err = adapter.ExecuteSwap(context.Background(), quote, tamperedEnvelope)
	require.ErrorIs(t, err, ErrExecutionPayload)

	expired := quote
	expired.ExpiresAt = now.Add(-time.Second)
	expired.ID, err = QuoteDigest(expired)
	require.NoError(t, err)
	expired.QuoteDigest = expired.ID
	_, err = adapter.ExecuteSwap(context.Background(), expired, envelope)
	require.ErrorIs(t, err, ErrQuoteExpired)

	belowMinimum := quote
	belowMinimum.MinOutputAmount = belowMinimum.OutputAmount.AddRaw(1)
	belowMinimum.ID, err = QuoteDigest(belowMinimum)
	require.NoError(t, err)
	belowMinimum.QuoteDigest = belowMinimum.ID
	_, err = adapter.ExecuteSwap(context.Background(), belowMinimum, envelope)
	require.ErrorIs(t, err, ErrMinimumOutput)

	adapter.executionVerifier = rejectingExecutionVerifier{}
	_, err = adapter.ExecuteSwap(context.Background(), quote, envelope)
	require.ErrorIs(t, err, ErrExecutionPayload)
}

func TestExecuteSwapUsesBase64TxRawAndConfirmedOutput(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	pool := securePool("1", "1000000000", "1000000000")
	var receivedTxBytes string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "all-pools"):
			require.NoError(t, json.NewEncoder(w).Encode(OsmosisPoolsResponse{Pools: []OsmosisPoolData{pool}}))
		case r.Method == http.MethodPost && r.URL.Path == "/cosmos/tx/v1beta1/txs":
			var request struct {
				TxBytes string `json:"tx_bytes"`
				Mode    string `json:"mode"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			receivedTxBytes = request.TxBytes
			require.Equal(t, "BROADCAST_MODE_SYNC", request.Mode)
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"tx_response": map[string]any{"txhash": "ABC123", "code": 0}}))
		case r.Method == http.MethodGet && r.URL.Path == "/cosmos/tx/v1beta1/txs/ABC123":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"tx_response": map[string]any{
				"txhash": "ABC123", "code": 0, "height": "98", "gas_used": "12345",
				"events": []any{map[string]any{"type": "token_swapped", "attributes": []any{map[string]string{"key": "tokens_out", "value": "9871580uatom"}}}},
			}}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	profile := testOsmosisProfile()
	profile.MaxPriceImpact = sdkmath.LegacyMustNewDecFromStr("0.99")
	profile.MaxOracleDeviation = sdkmath.LegacyMustNewDecFromStr("0.99")
	adapter, noop := secureAdapterWithProfile(t, server.URL, now, profile, sdkmath.LegacyMustNewDecFromStr("0.987158034397061298"), staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, Height: 100, BlockHash: "AABB", ObservedAt: now}})
	defer noop.Close()
	quote, err := adapter.GetSwapQuote(context.Background(), secureSwapRequest(now))
	require.NoError(t, err)
	payload, err := BuildExecutionPayload(quote)
	require.NoError(t, err)
	rawTx, err := (&tx.TxRaw{BodyBytes: []byte("body"), AuthInfoBytes: []byte("auth"), Signatures: [][]byte{[]byte("signature")}}).Marshal()
	require.NoError(t, err)
	envelope, err := MarshalSignedExecutionEnvelope(payload, rawTx)
	require.NoError(t, err)
	result, err := adapter.ExecuteSwap(context.Background(), quote, envelope)
	require.NoError(t, err)
	require.Equal(t, "9871580", result.OutputAmount.String())
	decoded, err := base64.StdEncoding.DecodeString(receivedTxBytes)
	require.NoError(t, err)
	require.Equal(t, rawTx, decoded)

	badEnvelope, err := MarshalSignedExecutionEnvelope(payload, []byte("not-a-txraw"))
	require.NoError(t, err)
	_, err = adapter.ExecuteSwap(context.Background(), quote, badEnvelope)
	require.ErrorIs(t, err, ErrExecutionPayload)
}

type rejectingExecutionVerifier struct{}

func (rejectingExecutionVerifier) VerifySignedExecution(context.Context, []byte, []byte) error {
	return errors.New("signature does not authorize payload")
}

func TestRemoteVerificationRejectsStatusOversizeAndPoolIdentity(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	profile := testOsmosisProfile()

	t.Run("status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "unavailable", http.StatusBadGateway) }))
		defer server.Close()
		adapter, remote := secureAdapterWithProfile(t, server.URL, now, profile, sdkmath.LegacyOneDec(), staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, Height: 100, BlockHash: "AABB", ObservedAt: now}})
		defer remote.Close()
		_, err := adapter.GetPool(context.Background(), "1")
		require.Error(t, err)
	})

	t.Run("oversize", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Repeat(" ", int(maxOsmosisResponseBytes)) + "{}"))
		}))
		defer server.Close()
		adapter, remote := secureAdapterWithProfile(t, server.URL, now, profile, sdkmath.LegacyOneDec(), staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, Height: 100, BlockHash: "AABB", ObservedAt: now}})
		defer remote.Close()
		_, err := adapter.GetPool(context.Background(), "1")
		require.Error(t, err)
	})

	t.Run("pool identity", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(OsmosisPoolResponse{Pool: securePool("2", "100", "100")})
		}))
		defer server.Close()
		adapter, remote := secureAdapterWithProfile(t, server.URL, now, profile, sdkmath.LegacyOneDec(), staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, Height: 100, BlockHash: "AABB", ObservedAt: now}})
		defer remote.Close()
		_, err := adapter.GetPool(context.Background(), "1")
		require.ErrorIs(t, err, ErrPoolStateEvidence)
	})
}

func secureSwapRequest(now time.Time) SwapRequest {
	return SwapRequest{
		FromToken: Token{Symbol: "OSMO", Denom: "uosmo", Decimals: 6, ChainID: OsmosisChainIDMainnet},
		ToToken:   Token{Symbol: "ATOM", Denom: "uatom", Decimals: 6, ChainID: OsmosisChainIDMainnet},
		Amount:    sdkmath.NewInt(10_000_000), Type: SwapTypeExactIn, SlippageTolerance: 0.01, SlippageToleranceExact: sdkmath.LegacyMustNewDecFromStr("0.01"),
		Deadline: now.Add(time.Minute), Sender: "osmo1sender", Recipient: "osmo1recipient",
	}
}

func securePool(id, reserveIn, reserveOut string) OsmosisPoolData {
	return OsmosisPoolData{
		Type: "/osmosis.gamm.v1beta1.Pool", ID: id, Address: "osmo1pool",
		PoolParams:  OsmosisPoolParams{SwapFee: "0.003"},
		PoolAssets:  []OsmosisPoolAsset{{Token: OsmosisCoin{Denom: "uosmo", Amount: reserveIn}, Weight: "50"}, {Token: OsmosisCoin{Denom: "uatom", Amount: reserveOut}, Weight: "50"}},
		TotalWeight: "100",
	}
}

func secureQuoteServer(t *testing.T, pool OsmosisPoolData) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "all-pools"):
			require.NoError(t, json.NewEncoder(w).Encode(OsmosisPoolsResponse{Pools: []OsmosisPoolData{pool}}))
		case strings.Contains(r.URL.Path, "/pools/"):
			require.NoError(t, json.NewEncoder(w).Encode(OsmosisPoolResponse{Pool: pool}))
		case strings.Contains(r.URL.Path, "/txs"):
			require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"tx_response": map[string]interface{}{"txhash": "ABC123", "code": 0}}))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func newSecureQuoteAdapter(t *testing.T, endpoint string, oracle sdkmath.LegacyDec) (*RealOsmosisAdapter, *httptest.Server, time.Time) {
	t.Helper()
	now := time.Unix(1_700_000_000, 0).UTC()
	profile := testOsmosisProfile()
	profile.MaxPriceImpact = sdkmath.LegacyMustNewDecFromStr("0.99")
	profile.MaxOracleDeviation = sdkmath.LegacyMustNewDecFromStr("0.99")
	adapter, server := secureAdapterWithProfile(t, endpoint, now, profile, oracle, staticChainEvidence{observation: ChainObservation{ChainID: OsmosisChainIDMainnet, Height: 100, BlockHash: "AABB", ObservedAt: now}})
	return adapter, server, now
}

func secureAdapterWithProfile(t *testing.T, endpoint string, now time.Time, profile *DEXRouteProfile, oracle sdkmath.LegacyDec, evidence ChainEvidenceProvider) (*RealOsmosisAdapter, *httptest.Server) {
	t.Helper()
	// The returned server is a no-op compatibility handle; endpoint ownership is
	// managed by the caller or secureQuoteServer cleanup.
	noop := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(noop.Close)
	cfg := AdapterConfig{Name: "osmosis-secure", Type: dexOsmosis, Enabled: true}
	osmosisCfg := DefaultOsmosisConfig()
	osmosisCfg.RESTEndpoint = endpoint
	osmosisCfg.RouteProfile = profile
	osmosisCfg.ValidationMode = RouteValidationEngineering
	observation, err := evidence.LatestObservation()
	require.NoError(t, err)
	if observation.SourceID == "" {
		observation.SourceID = testPoolStateSourceID
		evidence = staticChainEvidence{observation: observation}
	}
	osmosisCfg.PoolState = newRESTBoundPoolStateProvider(endpoint, http.DefaultClient, observation)
	osmosisCfg.ChainEvidence = evidence
	osmosisCfg.Oracle = staticOracle{price: oracle}
	osmosisCfg.ExecutionVerifier = staticExecutionVerifier{}
	osmosisCfg.Now = func() time.Time { return now }
	adapter, err := NewRealOsmosisAdapter(cfg, osmosisCfg)
	require.NoError(t, err)
	return adapter, noop
}

func TestPoolStateDigestDetectsTampering(t *testing.T) {
	evidence := PoolStateEvidence{ChainID: "chain", SourceID: "node", ProfileID: "profile", PoolID: "1", Height: 1, BlockHash: "AA", ObservedAt: time.Unix(1, 0), FromDenom: "a", ToDenom: "b", ReserveIn: sdkmath.NewInt(10), ReserveOut: sdkmath.NewInt(20), SwapFee: sdkmath.LegacyMustNewDecFromStr("0.003")}
	digest, err := canonicalPoolStateDigest(evidence)
	require.NoError(t, err)
	evidence.ReserveIn = sdkmath.NewInt(11)
	tampered, err := canonicalPoolStateDigest(evidence)
	require.NoError(t, err)
	require.NotEqual(t, digest, tampered)
	decoded, err := hex.DecodeString(digest)
	require.NoError(t, err)
	require.Len(t, decoded, sha256.Size)
}

var _ = errors.Is
