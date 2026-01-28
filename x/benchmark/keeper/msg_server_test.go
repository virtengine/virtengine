// Package keeper implements tests for the Benchmark module MsgServer.
//
// VE-2016: MsgServer tests for benchmark module
package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
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

func TestMsgServer_SubmitBenchmarks(t *testing.T) {
	msgServer, k, ctx, mockProvider, mockRoles := setupMsgServerTest(t)

	providerAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	moderatorAddr := "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
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
			ProviderAddress: providerAddr,
			Reports:         []types.BenchmarkReport{report},
		}
		_, err := msgServer.SubmitBenchmarks(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("invalid provider address", func(t *testing.T) {
		msg := &types.MsgSubmitBenchmarks{
			ProviderAddress: "invalid",
			Reports:         []types.BenchmarkReport{},
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
			ProviderAddress: providerAddr,
			Reports:         []types.BenchmarkReport{report},
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

	providerAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	requesterAddr := "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
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
				Requester:       requesterAddr,
				ProviderAddress: providerAddr,
				ClusterID:       "cluster-1",
				SuiteVersion:    "1.0.0",
				DeadlineSeconds: 3600, // 1 hour
			},
			wantErr: false,
		},
		{
			name: "invalid requester address",
			msg: &types.MsgRequestChallenge{
				Requester:       "invalid",
				ProviderAddress: providerAddr,
				ClusterID:       "cluster-1",
				SuiteVersion:    "1.0.0",
				DeadlineSeconds: 3600,
			},
			wantErr: true,
			errMsg:  "invalid requester address",
		},
		{
			name: "invalid provider address",
			msg: &types.MsgRequestChallenge{
				Requester:       requesterAddr,
				ProviderAddress: "invalid",
				ClusterID:       "cluster-1",
				SuiteVersion:    "1.0.0",
				DeadlineSeconds: 3600,
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
				require.NotEmpty(t, resp.ChallengeID)
			}
		})
	}
}

func TestMsgServer_FlagProvider(t *testing.T) {
	msgServer, _, ctx, mockProvider, mockRoles := setupMsgServerTest(t)

	providerAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	moderatorAddr := "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
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
				Moderator:       moderatorAddr,
				ProviderAddress: providerAddr,
				Reason:          "Performance issues detected",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &types.MsgFlagProvider{
				Moderator:       "invalid",
				ProviderAddress: providerAddr,
				Reason:          "test",
			},
			wantErr: true,
			errMsg:  "invalid moderator address",
		},
		{
			name: "non-moderator",
			msg: &types.MsgFlagProvider{
				Moderator:       "cosmos1w3jhxap3gempvr46xzaqf7ajj5drrjqcpd8n8g",
				ProviderAddress: providerAddr,
				Reason:          "test",
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

	providerAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	moderatorAddr := "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
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
				Moderator:       moderatorAddr,
				ProviderAddress: providerAddr,
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &types.MsgUnflagProvider{
				Moderator:       "invalid",
				ProviderAddress: providerAddr,
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

	moderatorAddr := "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
	mockRoles.AddModerator(moderatorAddr)

	// Create an anomaly flag to resolve
	anomalyFlag := &types.AnomalyFlag{
		FlagID:          "anomaly-1",
		ReportID:        "report-1",
		ProviderAddress: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
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
				Moderator:  moderatorAddr,
				FlagID:     "anomaly-1",
				Resolution: "Issue investigated and resolved",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &types.MsgResolveAnomalyFlag{
				Moderator:  "invalid",
				FlagID:     "anomaly-1",
				Resolution: "test",
			},
			wantErr: true,
			errMsg:  "invalid moderator address",
		},
		{
			name: "non-existent flag",
			msg: &types.MsgResolveAnomalyFlag{
				Moderator:  moderatorAddr,
				FlagID:     "non-existent",
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
