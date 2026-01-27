// Package keeper implements tests for the Benchmark module keeper.
//
// VE-600 through VE-603: Benchmark keeper tests
package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

	"github.com/virtengine/virtengine/x/benchmark/types"
)

// mockProviderKeeper is a mock implementation of ProviderKeeper
type mockProviderKeeper struct {
	providers map[string]bool
	pubKeys   map[string][]byte
}

func newMockProviderKeeper() *mockProviderKeeper {
	return &mockProviderKeeper{
		providers: make(map[string]bool),
		pubKeys:   make(map[string][]byte),
	}
}

func (m *mockProviderKeeper) ProviderExists(_ sdk.Context, addr sdk.AccAddress) bool {
	return m.providers[addr.String()]
}

func (m *mockProviderKeeper) GetProviderPublicKey(_ sdk.Context, addr sdk.AccAddress) ([]byte, bool) {
	pk, ok := m.pubKeys[addr.String()]
	return pk, ok
}

func (m *mockProviderKeeper) AddProvider(addr string, pubKey []byte) {
	m.providers[addr] = true
	m.pubKeys[addr] = pubKey
}

// mockRolesKeeper is a mock implementation of RolesKeeper
type mockRolesKeeper struct {
	moderators map[string]bool
	admins     map[string]bool
}

func newMockRolesKeeper() *mockRolesKeeper {
	return &mockRolesKeeper{
		moderators: make(map[string]bool),
		admins:     make(map[string]bool),
	}
}

func (m *mockRolesKeeper) IsModerator(_ sdk.Context, addr sdk.AccAddress) bool {
	return m.moderators[addr.String()]
}

func (m *mockRolesKeeper) IsAdmin(_ sdk.Context, addr sdk.AccAddress) bool {
	return m.admins[addr.String()]
}

func (m *mockRolesKeeper) AddModerator(addr string) {
	m.moderators[addr] = true
}

func (m *mockRolesKeeper) AddAdmin(addr string) {
	m.admins[addr] = true
}

// setupKeeper creates a test keeper
func setupKeeper(t *testing.T) (Keeper, sdk.Context, *mockProviderKeeper, *mockRolesKeeper) {
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

	keeper := NewKeeper(cdc, storeKey, mockProvider, mockRoles, "authority")

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	// Set default params
	_ = keeper.SetParams(ctx, types.DefaultParams())

	return keeper, ctx, mockProvider, mockRoles
}

// generateTestKeyPair generates a test ed25519 key pair
func generateTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	return pub, priv
}

// signReport signs a benchmark report
func signReport(t *testing.T, report *types.BenchmarkReport, priv ed25519.PrivateKey) {
	t.Helper()
	hash, err := report.Hash()
	if err != nil {
		t.Fatalf("failed to hash report: %v", err)
	}
	sig := ed25519.Sign(priv, []byte(hash))
	report.Signature = hex.EncodeToString(sig)
}

// createTestReport creates a valid test benchmark report
func createTestReport(t *testing.T, providerAddr string, pub ed25519.PublicKey, priv ed25519.PrivateKey) types.BenchmarkReport {
	t.Helper()

	report := types.BenchmarkReport{
		ReportID:        "test-report-1",
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
				SingleCoreScore: 5000,
				MultiCoreScore:  8000,
				CoreCount:       8,
				ThreadCount:     16,
				BaseFrequencyMHz: 3000,
			},
			Memory: types.MemoryMetrics{
				TotalGB:       64,
				BandwidthMBps: 50000,
				LatencyNs:     100,
				Score:         7000,
			},
			Disk: types.DiskMetrics{
				ReadIOPS:           100000,
				WriteIOPS:          80000,
				ReadThroughputMBps: 3000,
				WriteThroughputMBps: 2500,
				TotalStorageGB:     1000,
				Score:              6500,
			},
			Network: types.NetworkMetrics{
				ThroughputMbps:    10000,
				LatencyMs:         5000, // 5ms in fixed-point
				PacketLossRate:    100,  // 0.0001%
				ReferenceEndpoint: "benchmark.virtengine.com",
				Score:             8500,
			},
		},
		SummaryScore: 7000,
		Timestamp:    time.Now().UTC(),
		PublicKey:    hex.EncodeToString(pub),
	}

	signReport(t, &report, priv)
	return report
}

func TestSubmitBenchmarks(t *testing.T) {
	keeper, ctx, mockProvider, _ := setupKeeper(t)

	pub, priv := generateTestKeyPair(t)
	providerAddr := "cosmos1test123456789"
	mockProvider.AddProvider(providerAddr, pub)

	report := createTestReport(t, providerAddr, pub, priv)

	// Test successful submission
	err := keeper.SubmitBenchmarks(ctx, []types.BenchmarkReport{report})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify report was stored
	storedReport, found := keeper.GetBenchmarkReport(ctx, report.ReportID)
	if !found {
		t.Fatal("expected report to be found")
	}
	if storedReport.SummaryScore != report.SummaryScore {
		t.Errorf("expected summary score %d, got %d", report.SummaryScore, storedReport.SummaryScore)
	}
}

func TestSubmitBenchmarks_DuplicateRejection(t *testing.T) {
	keeper, ctx, mockProvider, _ := setupKeeper(t)

	pub, priv := generateTestKeyPair(t)
	providerAddr := "cosmos1test123456789"
	mockProvider.AddProvider(providerAddr, pub)

	report := createTestReport(t, providerAddr, pub, priv)

	// Submit first time
	err := keeper.SubmitBenchmarks(ctx, []types.BenchmarkReport{report})
	if err != nil {
		t.Fatalf("first submission should succeed: %v", err)
	}

	// Submit duplicate
	err = keeper.SubmitBenchmarks(ctx, []types.BenchmarkReport{report})
	if err == nil {
		t.Fatal("expected duplicate rejection error")
	}
}

func TestSubmitBenchmarks_InvalidSignature(t *testing.T) {
	keeper, ctx, mockProvider, _ := setupKeeper(t)

	pub, priv := generateTestKeyPair(t)
	providerAddr := "cosmos1test123456789"
	mockProvider.AddProvider(providerAddr, pub)

	report := createTestReport(t, providerAddr, pub, priv)
	report.Signature = "invalidsignature"

	err := keeper.SubmitBenchmarks(ctx, []types.BenchmarkReport{report})
	if err == nil {
		t.Fatal("expected signature verification error")
	}
}

func TestSubmitBenchmarks_UnknownProvider(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	pub, priv := generateTestKeyPair(t)
	providerAddr := "cosmos1unknown123456789"
	// Not adding provider to mock

	report := createTestReport(t, providerAddr, pub, priv)

	err := keeper.SubmitBenchmarks(ctx, []types.BenchmarkReport{report})
	if err == nil {
		t.Fatal("expected unknown provider error")
	}
}

func TestReliabilityScore(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	providerAddr := "cosmos1test123456789"

	inputs := types.ReliabilityScoreInputs{
		BenchmarkSummary:        7000,
		ProvisioningSuccessRate: 950000, // 95%
		ProvisioningAttempts:    100,
		ProvisioningSuccesses:   95,
		MeanTimeToProvision:     120,
		MeanTimeBetweenFailures: 86400 * 7, // 7 days
		TotalUptimeSeconds:      86400 * 30,
		TotalDowntimeSeconds:    3600, // 1 hour
		DisputeCount:            1,
		DisputesLost:            0,
		AnomalyFlagCount:        0,
	}

	err := keeper.UpdateReliabilityScore(ctx, providerAddr, inputs)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	score, found := keeper.GetReliabilityScore(ctx, providerAddr)
	if !found {
		t.Fatal("expected score to be found")
	}

	if score.Score <= 0 || score.Score > 10000 {
		t.Errorf("score out of expected range: %d", score.Score)
	}

	if score.ScoreVersion != types.ScoreVersion {
		t.Errorf("expected score version %s, got %s", types.ScoreVersion, score.ScoreVersion)
	}
}

func TestComputeReliabilityScore_Deterministic(t *testing.T) {
	inputs := types.ReliabilityScoreInputs{
		BenchmarkSummary:        7500,
		ProvisioningSuccessRate: 980000,
		ProvisioningAttempts:    1000,
		ProvisioningSuccesses:   980,
		MeanTimeToProvision:     60,
		MeanTimeBetweenFailures: 86400 * 30,
		TotalUptimeSeconds:      86400 * 90,
		TotalDowntimeSeconds:    3600,
		DisputeCount:            0,
		DisputesLost:            0,
		AnomalyFlagCount:        0,
	}

	// Compute multiple times to verify determinism
	score1, components1 := types.ComputeReliabilityScore(inputs)
	score2, components2 := types.ComputeReliabilityScore(inputs)

	if score1 != score2 {
		t.Errorf("scores not deterministic: %d != %d", score1, score2)
	}

	if components1.PerformanceScore != components2.PerformanceScore {
		t.Error("performance score not deterministic")
	}
	if components1.UptimeScore != components2.UptimeScore {
		t.Error("uptime score not deterministic")
	}
	if components1.ProvisioningScore != components2.ProvisioningScore {
		t.Error("provisioning score not deterministic")
	}
	if components1.TrustScore != components2.TrustScore {
		t.Error("trust score not deterministic")
	}
}

func TestChallenge(t *testing.T) {
	keeper, ctx, mockProvider, _ := setupKeeper(t)

	pub, priv := generateTestKeyPair(t)
	providerAddr := "cosmos1test123456789"
	mockProvider.AddProvider(providerAddr, pub)

	challenge := &types.BenchmarkChallenge{
		ProviderAddress:      providerAddr,
		ClusterID:            "test-cluster-1",
		RequiredSuiteVersion: "1.0.0",
		SuiteHash:            "abc123",
		Deadline:             time.Now().Add(24 * time.Hour),
	}

	err := keeper.CreateChallenge(ctx, challenge)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify challenge was stored
	storedChallenge, found := keeper.GetChallenge(ctx, challenge.ChallengeID)
	if !found {
		t.Fatal("expected challenge to be found")
	}
	if storedChallenge.State != types.ChallengeStatePending {
		t.Errorf("expected state pending, got %s", storedChallenge.State)
	}

	// Respond to challenge
	report := createTestReport(t, providerAddr, pub, priv)
	report.ReportID = "challenge-response-1"

	err = keeper.RespondToChallenge(ctx, challenge.ChallengeID, report, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify challenge was updated
	storedChallenge, _ = keeper.GetChallenge(ctx, challenge.ChallengeID)
	if storedChallenge.State != types.ChallengeStateCompleted {
		t.Errorf("expected state completed, got %s", storedChallenge.State)
	}
}

func TestChallenge_Expired(t *testing.T) {
	keeper, ctx, mockProvider, _ := setupKeeper(t)

	pub, priv := generateTestKeyPair(t)
	providerAddr := "cosmos1test123456789"
	mockProvider.AddProvider(providerAddr, pub)

	challenge := &types.BenchmarkChallenge{
		ProviderAddress:      providerAddr,
		ClusterID:            "test-cluster-1",
		RequiredSuiteVersion: "1.0.0",
		Deadline:             time.Now().Add(-1 * time.Hour), // Already expired
	}

	err := keeper.CreateChallenge(ctx, challenge)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Try to respond to expired challenge
	report := createTestReport(t, providerAddr, pub, priv)
	report.ReportID = "expired-response-1"

	err = keeper.RespondToChallenge(ctx, challenge.ChallengeID, report, "")
	if err == nil {
		t.Fatal("expected expired challenge error")
	}
}

func TestAnomalyDetection_SuddenJump(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	providerAddr := "cosmos1test123456789"

	previousReport := types.BenchmarkReport{
		ReportID:        "prev-1",
		ProviderAddress: providerAddr,
		ClusterID:       "cluster-1",
		SummaryScore:    5000,
		Timestamp:       time.Now().Add(-1 * time.Hour),
		Metrics: types.BenchmarkMetrics{
			SchemaVersion: types.MetricSchemaVersion,
			CPU:           types.CPUMetrics{SingleCoreScore: 5000, MultiCoreScore: 5000, CoreCount: 8, ThreadCount: 16},
			Memory:        types.MemoryMetrics{TotalGB: 64, Score: 5000, BandwidthMBps: 10000},
			Disk:          types.DiskMetrics{Score: 5000},
			Network:       types.NetworkMetrics{Score: 5000},
		},
	}

	currentReport := types.BenchmarkReport{
		ReportID:        "curr-1",
		ProviderAddress: providerAddr,
		ClusterID:       "cluster-1",
		SummaryScore:    9000, // 80% jump
		Timestamp:       time.Now(),
		Metrics: types.BenchmarkMetrics{
			SchemaVersion: types.MetricSchemaVersion,
			CPU:           types.CPUMetrics{SingleCoreScore: 9000, MultiCoreScore: 9000, CoreCount: 8, ThreadCount: 16},
			Memory:        types.MemoryMetrics{TotalGB: 64, Score: 9000, BandwidthMBps: 10000},
			Disk:          types.DiskMetrics{Score: 9000},
			Network:       types.NetworkMetrics{Score: 9000},
		},
	}

	anomalies := keeper.DetectAnomalies(ctx, currentReport, []types.BenchmarkReport{previousReport})

	if len(anomalies) == 0 {
		t.Fatal("expected at least one anomaly to be detected")
	}

	foundJump := false
	for _, a := range anomalies {
		if a.Type == types.AnomalyTypeSuddenJump {
			foundJump = true
			break
		}
	}

	if !foundJump {
		t.Error("expected sudden jump anomaly to be detected")
	}
}

func TestAnomalyDetection_RepeatedOutput(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	providerAddr := "cosmos1test123456789"

	metrics := types.BenchmarkMetrics{
		SchemaVersion: types.MetricSchemaVersion,
		CPU:           types.CPUMetrics{SingleCoreScore: 5000, MultiCoreScore: 8000, CoreCount: 8, ThreadCount: 16},
		Memory:        types.MemoryMetrics{TotalGB: 64, BandwidthMBps: 50000, Score: 7000},
		Disk:          types.DiskMetrics{ReadIOPS: 100000, WriteIOPS: 80000, Score: 6500},
		Network:       types.NetworkMetrics{ThroughputMbps: 10000, Score: 8500},
	}

	// Create multiple identical reports
	var previousReports []types.BenchmarkReport
	for i := 0; i < 5; i++ {
		previousReports = append(previousReports, types.BenchmarkReport{
			ReportID:        "prev-" + string(rune('0'+i)),
			ProviderAddress: providerAddr,
			ClusterID:       "cluster-1",
			SummaryScore:    7000,
			Timestamp:       time.Now().Add(-time.Duration(i+1) * time.Hour),
			Metrics:         metrics,
		})
	}

	currentReport := types.BenchmarkReport{
		ReportID:        "current",
		ProviderAddress: providerAddr,
		ClusterID:       "cluster-1",
		SummaryScore:    7000,
		Timestamp:       time.Now(),
		Metrics:         metrics,
	}

	anomalies := keeper.DetectAnomalies(ctx, currentReport, previousReports)

	foundRepeated := false
	for _, a := range anomalies {
		if a.Type == types.AnomalyTypeRepeatedOutput {
			foundRepeated = true
			break
		}
	}

	if !foundRepeated {
		t.Error("expected repeated output anomaly to be detected")
	}
}

func TestProviderFlag(t *testing.T) {
	keeper, ctx, _, mockRoles := setupKeeper(t)

	moderatorAddr := "cosmos1moderator123"
	mockRoles.AddModerator(moderatorAddr)

	providerAddr := "cosmos1provider123"

	flag := &types.ProviderFlag{
		ProviderAddress: providerAddr,
		Reason:          "Suspected fraudulent benchmarks",
		FlaggedBy:       moderatorAddr,
	}

	err := keeper.FlagProvider(ctx, flag)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify flag
	if !keeper.IsProviderFlagged(ctx, providerAddr) {
		t.Fatal("expected provider to be flagged")
	}

	// Unflag
	modAddr, _ := sdk.AccAddressFromBech32(moderatorAddr)
	err = keeper.UnflagProvider(ctx, providerAddr, modAddr)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify unflagged
	if keeper.IsProviderFlagged(ctx, providerAddr) {
		t.Fatal("expected provider to be unflagged")
	}
}

func TestPruneOldReports(t *testing.T) {
	keeper, ctx, mockProvider, _ := setupKeeper(t)

	pub, priv := generateTestKeyPair(t)
	providerAddr := "cosmos1test123456789"
	mockProvider.AddProvider(providerAddr, pub)

	// Set a low retention limit
	params := types.DefaultParams()
	params.RetentionCount = 3
	_ = keeper.SetParams(ctx, params)

	// Submit more reports than retention limit
	for i := 0; i < 5; i++ {
		report := createTestReport(t, providerAddr, pub, priv)
		report.ReportID = "report-" + string(rune('0'+i))
		report.Timestamp = time.Now().Add(-time.Duration(4-i) * time.Hour)
		signReport(t, &report, priv)

		// Store directly to avoid auto-pruning during submit
		_ = keeper.SetBenchmarkReport(ctx, report)
	}

	// Prune
	pruned, err := keeper.PruneOldReports(ctx, providerAddr, "test-cluster-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if pruned != 2 {
		t.Errorf("expected 2 reports pruned, got %d", pruned)
	}
}

func TestParams(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	params := types.Params{
		RetentionCount:                  50,
		DefaultChallengeDeadlineSeconds: 43200,
		MinBenchmarkInterval:            600,
		MaxReportsPerSubmission:         5,
		AnomalyThresholdJumpPercent:     75,
		AnomalyThresholdRepeatCount:     5,
	}

	err := keeper.SetParams(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	storedParams := keeper.GetParams(ctx)
	if storedParams.RetentionCount != params.RetentionCount {
		t.Errorf("expected retention count %d, got %d", params.RetentionCount, storedParams.RetentionCount)
	}
}

func TestSequences(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	// Test report sequence
	seq1 := keeper.GetNextReportSequence(ctx)
	keeper.SetNextReportSequence(ctx, seq1+1)
	seq2 := keeper.GetNextReportSequence(ctx)
	if seq2 != seq1+1 {
		t.Errorf("expected sequence %d, got %d", seq1+1, seq2)
	}

	// Test challenge sequence
	cSeq1 := keeper.GetNextChallengeSequence(ctx)
	keeper.SetNextChallengeSequence(ctx, cSeq1+1)
	cSeq2 := keeper.GetNextChallengeSequence(ctx)
	if cSeq2 != cSeq1+1 {
		t.Errorf("expected sequence %d, got %d", cSeq1+1, cSeq2)
	}

	// Test anomaly sequence
	aSeq1 := keeper.GetNextAnomalySequence(ctx)
	keeper.SetNextAnomalySequence(ctx, aSeq1+1)
	aSeq2 := keeper.GetNextAnomalySequence(ctx)
	if aSeq2 != aSeq1+1 {
		t.Errorf("expected sequence %d, got %d", aSeq1+1, aSeq2)
	}
}

// Suppress unused import warning
var _ = json.Marshal
