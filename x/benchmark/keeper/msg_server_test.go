// Package keeper implements tests for the Benchmark module MsgServer.
//
// VE-2016: MsgServer tests for benchmark module
package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
	"github.com/virtengine/virtengine/x/benchmark/types"
)

// setupMsgServerTest creates a test MsgServer with mocked dependencies
func setupMsgServerTest(t *testing.T) (types.MsgServer, Keeper, sdk.Context, *mockProviderKeeper, *mockRolesKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	mockProvider := newMockProviderKeeper()
	mockRoles := newMockRolesKeeper()

	k := NewKeeper(
		cdc,
		storeKey,
		mockProvider,
		mockRoles,
		"authority",
	)

	// Set default params
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now(),
		Height: 1,
	}, false, log.NewNopLogger())

	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	msgServer := NewMsgServerImpl(k)

	return msgServer, k, ctx, mockProvider, mockRoles
}

// generateMsgServerTestKeyPair generates a test ed25519 key pair for MsgServer tests
func generateMsgServerTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	return pub, priv
}

// signMsgServerTestReport signs a benchmark report for MsgServer tests
func signMsgServerTestReport(t *testing.T, report *types.BenchmarkReport, priv ed25519.PrivateKey) {
	t.Helper()
	hash, err := report.Hash()
	if err != nil {
		t.Fatalf("failed to hash report: %v", err)
	}
	sig := ed25519.Sign(priv, []byte(hash))
	report.Signature = hex.EncodeToString(sig)
}

// createMsgServerTestReport creates a valid test benchmark report for MsgServer tests
func createMsgServerTestReport(t *testing.T, providerAddr string, pub ed25519.PublicKey, priv ed25519.PrivateKey) types.BenchmarkReport {
	t.Helper()

	report := types.BenchmarkReport{
		ReportID:        "test-report-msg-1",
		ProviderAddress: providerAddr,
		ClusterID:       "test-cluster-1",
		NodeMetadata: types.NodeMetadata{
			NodeID: "node-1",
			Region: "us-east-1",
			OSType: "linux",
		},
		SuiteVersion: "1.0.0",
		SuiteHash:    "abc123",
		Metrics: types.BenchmarkMetrics{
			SchemaVersion: types.MetricSchemaVersion,
			CPU: types.CPUMetrics{
				SingleCoreScore:  5000,
				MultiCoreScore:   8000,
				CoreCount:        8,
				ThreadCount:      16,
				BaseFrequencyMHz: 3000,
			},
			Memory: types.MemoryMetrics{
				TotalGB:       64,
				BandwidthMBps: 50000,
				LatencyNs:     100,
				Score:         7000,
			},
			Disk: types.DiskMetrics{
				ReadIOPS:            100000,
				WriteIOPS:           80000,
				ReadThroughputMBps:  3000,
				WriteThroughputMBps: 2500,
				TotalStorageGB:      1000,
				Score:               6500,
			},
			Network: types.NetworkMetrics{
				ThroughputMbps:    10000,
				LatencyMs:         5000,
				PacketLossRate:    100,
				ReferenceEndpoint: "benchmark.virtengine.com",
				Score:             8500,
			},
		},
		SummaryScore: 7000,
		Timestamp:    time.Now().UTC(),
		PublicKey:    hex.EncodeToString(pub),
	}

	signMsgServerTestReport(t, &report, priv)
	return report
}

// convertReportsToResults converts BenchmarkReport to BenchmarkResult for msg testing
func convertReportsToResults(reports []types.BenchmarkReport) []benchmarkv1.BenchmarkResult {
	results := make([]benchmarkv1.BenchmarkResult, len(reports))
	for i, r := range reports {
		results[i] = benchmarkv1.BenchmarkResult{
			BenchmarkType: "compute",
			Score:         fmt.Sprintf("%d", r.SummaryScore),
			Timestamp:     r.Timestamp.Unix(),
		}
	}
	return results
}

func TestMsgServer_SubmitBenchmarks(t *testing.T) {
	msgServer, k, ctx, mockProvider, mockRoles := setupMsgServerTest(t)

	providerAddr := bech32AddrBenchmark(t)
	moderatorAddr := bech32AddrBenchmark(t)
	pub, priv := generateMsgServerTestKeyPair(t)

	// Add provider to mock
	mockProvider.AddProvider(providerAddr, pub)
	// Add moderator for flagging test
	mockRoles.AddModerator(moderatorAddr)

	t.Run("valid submission", func(t *testing.T) {
		report := createMsgServerTestReport(t, providerAddr, pub, priv)
		report.ReportID = "valid-submission-1"
		signMsgServerTestReport(t, &report, priv)
		msg := &types.MsgSubmitBenchmarks{
			Provider:  providerAddr,
			ClusterId: "test-cluster-1",
			Results:   convertReportsToResults([]types.BenchmarkReport{report}),
			Signature: []byte(report.Signature),
		}
		_, err := msgServer.SubmitBenchmarks(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("invalid provider address", func(t *testing.T) {
		msg := &types.MsgSubmitBenchmarks{
			Provider: "invalid",
			Results:  []benchmarkv1.BenchmarkResult{},
		}
		_, err := msgServer.SubmitBenchmarks(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid provider address")
	})

	t.Run("flagged provider", func(t *testing.T) {
		// Flag the provider first (using a moderator address that exists in mockRoles)
		flag := &types.ProviderFlag{
			ProviderAddress: providerAddr,
			Active:          true,
			Reason:          "test flag",
			FlaggedBy:       moderatorAddr,
			FlaggedAt:       ctx.BlockTime(),
		}
		err := k.FlagProvider(ctx, flag)
		require.NoError(t, err)

		report := createMsgServerTestReport(t, providerAddr, pub, priv)
		report.ReportID = "flagged-submission-1"
		signMsgServerTestReport(t, &report, priv)
		msg := &types.MsgSubmitBenchmarks{
			Provider:  providerAddr,
			ClusterId: "test-cluster-1",
			Results:   convertReportsToResults([]types.BenchmarkReport{report}),
			Signature: []byte(report.Signature),
		}
		_, err = msgServer.SubmitBenchmarks(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "flagged")

		// Cleanup
		moderatorAccAddr, _ := sdk.AccAddressFromBech32(moderatorAddr)
		_ = k.UnflagProvider(ctx, providerAddr, moderatorAccAddr)
	})
}

func TestMsgServer_RequestChallenge(t *testing.T) {
	msgServer, _, ctx, mockProvider, _ := setupMsgServerTest(t)

	providerAddr := bech32AddrBenchmark(t)
	requesterAddr := bech32AddrBenchmark(t)
	pub, _ := generateMsgServerTestKeyPair(t)

	// Add provider to mock
	mockProvider.AddProvider(providerAddr, pub)

	tests := []struct {
		name    string
		msg     *types.MsgRequestChallenge
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid challenge request",
			msg: &types.MsgRequestChallenge{
				Requester:     requesterAddr,
				Provider:      providerAddr,
				BenchmarkType: "compute",
			},
			wantErr: false,
		},
		{
			name: "invalid requester address",
			msg: &types.MsgRequestChallenge{
				Requester:     "invalid",
				Provider:      providerAddr,
				BenchmarkType: "compute",
			},
			wantErr: true,
			errMsg:  "invalid requester address",
		},
		{
			name: "invalid provider address",
			msg: &types.MsgRequestChallenge{
				Requester:     requesterAddr,
				Provider:      "invalid",
				BenchmarkType: "compute",
			},
			wantErr: true,
			errMsg:  "invalid provider address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.RequestChallenge(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.ChallengeId)
			}
		})
	}
}

func TestMsgServer_FlagProvider(t *testing.T) {
	msgServer, _, ctx, mockProvider, mockRoles := setupMsgServerTest(t)

	providerAddr := bech32AddrBenchmark(t)
	moderatorAddr := bech32AddrBenchmark(t)
	pub, _ := generateMsgServerTestKeyPair(t)

	// Add provider and moderator to mocks
	mockProvider.AddProvider(providerAddr, pub)
	mockRoles.AddModerator(moderatorAddr)

	tests := []struct {
		name    string
		msg     *types.MsgFlagProvider
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid flag",
			msg: &types.MsgFlagProvider{
				Reporter: moderatorAddr,
				Provider: providerAddr,
				Reason:   "Performance issues detected",
			},
			wantErr: false,
		},
		{
			name: "invalid reporter address",
			msg: &types.MsgFlagProvider{
				Reporter: "invalid",
				Provider: providerAddr,
				Reason:   "test",
			},
			wantErr: true,
			errMsg:  "invalid moderator address",
		},
		{
			name: "non-moderator",
			msg: &types.MsgFlagProvider{
				Reporter: bech32AddrBenchmark(t),
				Provider: providerAddr,
				Reason:   "test",
			},
			wantErr: true,
			errMsg:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.FlagProvider(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_UnflagProvider(t *testing.T) {
	msgServer, k, ctx, mockProvider, mockRoles := setupMsgServerTest(t)

	providerAddr := bech32AddrBenchmark(t)
	moderatorAddr := bech32AddrBenchmark(t)
	pub, _ := generateMsgServerTestKeyPair(t)

	// Add provider and moderator to mocks
	mockProvider.AddProvider(providerAddr, pub)
	mockRoles.AddModerator(moderatorAddr)

	// First, flag the provider
	flag := &types.ProviderFlag{
		ProviderAddress: providerAddr,
		Active:          true,
		Reason:          "test flag",
		FlaggedBy:       moderatorAddr,
		FlaggedAt:       ctx.BlockTime(),
	}
	err := k.FlagProvider(ctx, flag)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     *types.MsgUnflagProvider
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid unflag",
			msg: &types.MsgUnflagProvider{
				Authority: moderatorAddr,
				Provider:  providerAddr,
			},
			wantErr: false,
		},
		{
			name: "invalid authority address",
			msg: &types.MsgUnflagProvider{
				Authority: "invalid",
				Provider:  providerAddr,
			},
			wantErr: true,
			errMsg:  "invalid moderator address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.UnflagProvider(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_ResolveAnomalyFlag(t *testing.T) {
	msgServer, k, ctx, _, mockRoles := setupMsgServerTest(t)

	moderatorAddr := bech32AddrBenchmark(t)
	providerAddr := bech32AddrBenchmark(t)
	missingProviderAddr := bech32AddrBenchmark(t)
	mockRoles.AddModerator(moderatorAddr)

	// Create an anomaly flag to resolve
	anomalyFlag := &types.AnomalyFlag{
		FlagID:          "anomaly-1",
		ReportID:        "report-1",
		ProviderAddress: providerAddr,
		Type:            types.AnomalyTypeSuddenJump,
		Severity:        types.AnomalySeverityMedium,
		Description:     "Test anomaly",
		CreatedAt:       ctx.BlockTime(),
	}
	err := k.CreateAnomalyFlag(ctx, anomalyFlag)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     *types.MsgResolveAnomalyFlag
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid resolution",
			msg: &types.MsgResolveAnomalyFlag{
				Authority:  moderatorAddr,
				Provider:   providerAddr,
				Resolution: "Issue investigated and resolved",
			},
			wantErr: false,
		},
		{
			name: "invalid authority address",
			msg: &types.MsgResolveAnomalyFlag{
				Authority:  "invalid",
				Provider:   providerAddr,
				Resolution: "test",
			},
			wantErr: true,
			errMsg:  "invalid moderator address",
		},
		{
			name: "non-existent provider",
			msg: &types.MsgResolveAnomalyFlag{
				Authority:  moderatorAddr,
				Provider:   missingProviderAddr,
				Resolution: "test",
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.ResolveAnomalyFlag(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
