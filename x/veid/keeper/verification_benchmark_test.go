// Package keeper_test contains VEID verification latency benchmarks.
// Task Reference: PERF-001 - VEID Verification Latency Benchmarks (target: <100ms)
package keeper_test

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
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

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// VEID Verification Latency Benchmarks
// Target: <100ms per verification
// ============================================================================

// VerificationLatencyBaseline defines baseline metrics for verification latency
type VerificationLatencyBaseline struct {
	TargetLatency        time.Duration `json:"target_latency"`
	MaxLatencyP95        time.Duration `json:"max_latency_p95"`
	MaxLatencyP99        time.Duration `json:"max_latency_p99"`
	MinThroughput        float64       `json:"min_throughput"` // verifications per second
	MaxDecryptionLatency time.Duration `json:"max_decryption_latency"`
	MaxScoringLatency    time.Duration `json:"max_scoring_latency"`
}

// DefaultVerificationBaseline returns baseline metrics for verification
func DefaultVerificationBaseline() VerificationLatencyBaseline {
	return VerificationLatencyBaseline{
		TargetLatency:        100 * time.Millisecond,
		MaxLatencyP95:        150 * time.Millisecond,
		MaxLatencyP99:        250 * time.Millisecond,
		MinThroughput:        10.0, // 10 verifications per second
		MaxDecryptionLatency: 30 * time.Millisecond,
		MaxScoringLatency:    50 * time.Millisecond,
	}
}

// MockDecryptedScope represents a decrypted scope for benchmarking
type MockDecryptedScope struct {
	ScopeID      string
	ContentHash  []byte
	Features     []float64
	BiometricRef []byte
	Timestamp    time.Time
	Valid        bool
}

// MockMLScorer simulates ML scoring for benchmarking
type MockMLScorer struct {
	modelVersion    string
	processingDelay time.Duration
}

// NewMockMLScorer creates a new mock ML scorer
func NewMockMLScorer() *MockMLScorer {
	return &MockMLScorer{
		modelVersion:    "v1.0.0-benchmark",
		processingDelay: 10 * time.Millisecond,
	}
}

// Score computes a mock score for decrypted scopes
func (s *MockMLScorer) Score(scopes []MockDecryptedScope) (int32, string, []byte, error) {
	time.Sleep(s.processingDelay)

	// Simulate feature extraction and scoring
	h := sha256.New()
	for _, scope := range scopes {
		h.Write(scope.ContentHash)
		for _, f := range scope.Features {
			h.Write([]byte(fmt.Sprintf("%f", f)))
		}
	}
	inputHash := h.Sum(nil)

	// Return a score based on number of valid scopes
	validCount := 0
	for _, scope := range scopes {
		if scope.Valid {
			validCount++
		}
	}

	// Score calculation (50-100 based on valid scopes)
	score := int32(50 + (validCount * 10))
	if score > 100 {
		score = 100
	}

	return score, s.modelVersion, inputHash, nil
}

// MockDecryptor simulates scope decryption for benchmarking
type MockDecryptor struct {
	decryptionDelay time.Duration
}

// NewMockDecryptor creates a new mock decryptor
func NewMockDecryptor() *MockDecryptor {
	return &MockDecryptor{
		decryptionDelay: 5 * time.Millisecond,
	}
}

// Decrypt decrypts mock scopes
func (d *MockDecryptor) Decrypt(scopeIDs []string) ([]MockDecryptedScope, error) {
	scopes := make([]MockDecryptedScope, 0, len(scopeIDs))

	for _, id := range scopeIDs {
		time.Sleep(d.decryptionDelay)

		// Generate mock decrypted data
		contentHash := make([]byte, 32)
		rand.Read(contentHash)

		features := make([]float64, 128) // 128-dim feature vector
		for i := range features {
			features[i] = float64(i) / 128.0
		}

		biometricRef := make([]byte, 64)
		rand.Read(biometricRef)

		scopes = append(scopes, MockDecryptedScope{
			ScopeID:      id,
			ContentHash:  contentHash,
			Features:     features,
			BiometricRef: biometricRef,
			Timestamp:    time.Now().UTC(),
			Valid:        true,
		})
	}

	return scopes, nil
}

// MockVerificationPipeline simulates the full verification pipeline
type MockVerificationPipeline struct {
	decryptor *MockDecryptor
	scorer    *MockMLScorer
	mu        sync.Mutex
	latencies []time.Duration
}

// NewMockVerificationPipeline creates a new mock verification pipeline
func NewMockVerificationPipeline() *MockVerificationPipeline {
	return &MockVerificationPipeline{
		decryptor: NewMockDecryptor(),
		scorer:    NewMockMLScorer(),
		latencies: make([]time.Duration, 0, 10000),
	}
}

// VerificationResult represents the result of a verification
type VerificationResult struct {
	Score         int32
	ModelVersion  string
	InputHash     []byte
	Duration      time.Duration
	DecryptDur    time.Duration
	ScoringDur    time.Duration
	ScopeCount    int
	ValidCount    int
	Success       bool
	Error         error
}

// Verify performs a full verification
func (p *MockVerificationPipeline) Verify(accountAddr string, scopeIDs []string) *VerificationResult {
	start := time.Now()
	result := &VerificationResult{
		ScopeCount: len(scopeIDs),
	}

	// Step 1: Decrypt scopes
	decryptStart := time.Now()
	scopes, err := p.decryptor.Decrypt(scopeIDs)
	result.DecryptDur = time.Since(decryptStart)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Step 2: Score using ML
	scoringStart := time.Now()
	score, modelVersion, inputHash, err := p.scorer.Score(scopes)
	result.ScoringDur = time.Since(scoringStart)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Count valid scopes
	for _, scope := range scopes {
		if scope.Valid {
			result.ValidCount++
		}
	}

	result.Score = score
	result.ModelVersion = modelVersion
	result.InputHash = inputHash
	result.Duration = time.Since(start)
	result.Success = true

	// Record latency
	p.mu.Lock()
	p.latencies = append(p.latencies, result.Duration)
	p.mu.Unlock()

	return result
}

// GetLatencies returns all recorded latencies
func (p *MockVerificationPipeline) GetLatencies() []time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.latencies
}

// BenchmarkScopeDecryption benchmarks scope decryption
func BenchmarkScopeDecryption(b *testing.B) {
	decryptor := NewMockDecryptor()
	scopeIDs := []string{"scope_1", "scope_2", "scope_3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decryptor.Decrypt(scopeIDs)
	}
}

// BenchmarkMLScoring benchmarks ML scoring
func BenchmarkMLScoring(b *testing.B) {
	scorer := NewMockMLScorer()
	scopes := []MockDecryptedScope{
		{ScopeID: "scope_1", ContentHash: make([]byte, 32), Features: make([]float64, 128), Valid: true},
		{ScopeID: "scope_2", ContentHash: make([]byte, 32), Features: make([]float64, 128), Valid: true},
		{ScopeID: "scope_3", ContentHash: make([]byte, 32), Features: make([]float64, 128), Valid: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = scorer.Score(scopes)
	}
}

// BenchmarkFullVerification benchmarks full verification pipeline
func BenchmarkFullVerification(b *testing.B) {
	pipeline := NewMockVerificationPipeline()
	scopeIDs := []string{"scope_1", "scope_2", "scope_3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := pipeline.Verify("cosmos1testaddr", scopeIDs)
		if !result.Success {
			b.Fatalf("verification failed: %v", result.Error)
		}
	}

	b.ReportMetric(float64(time.Second)/float64(b.Elapsed()/time.Duration(b.N)), "verifications/sec")
}

// BenchmarkVerificationWithVaryingScopes benchmarks verification with varying scope counts
func BenchmarkVerificationWithVaryingScopes(b *testing.B) {
	scopeCounts := []int{1, 2, 3, 5, 10}

	for _, count := range scopeCounts {
		b.Run(fmt.Sprintf("scopes_%d", count), func(b *testing.B) {
			pipeline := NewMockVerificationPipeline()
			scopeIDs := make([]string, count)
			for i := 0; i < count; i++ {
				scopeIDs[i] = fmt.Sprintf("scope_%d", i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := pipeline.Verify("cosmos1testaddr", scopeIDs)
				if !result.Success {
					b.Fatalf("verification failed: %v", result.Error)
				}
			}

			b.ReportMetric(float64(count), "scope_count")
		})
	}
}

// BenchmarkVerificationParallel benchmarks parallel verification
func BenchmarkVerificationParallel(b *testing.B) {
	pipeline := NewMockVerificationPipeline()
	scopeIDs := []string{"scope_1", "scope_2", "scope_3"}
	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := counter.Add(1)
			addr := fmt.Sprintf("cosmos1testaddr%d", i%1000)
			result := pipeline.Verify(addr, scopeIDs)
			if !result.Success {
				b.Fatalf("verification failed: %v", result.Error)
			}
		}
	})
}

// TestVerificationLatencyBaseline tests verification latency against baseline
func TestVerificationLatencyBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping verification latency baseline test in short mode")
	}

	baseline := DefaultVerificationBaseline()
	pipeline := NewMockVerificationPipeline()

	const iterations = 100
	scopeIDs := []string{"scope_1", "scope_2", "scope_3"}

	var totalDecrypt, totalScoring time.Duration
	var successCount int

	for i := 0; i < iterations; i++ {
		addr := fmt.Sprintf("cosmos1testaddr%d", i)
		result := pipeline.Verify(addr, scopeIDs)

		if result.Success {
			successCount++
			totalDecrypt += result.DecryptDur
			totalScoring += result.ScoringDur
		}
	}

	latencies := pipeline.GetLatencies()
	p95, p99 := calculateVerificationPercentiles(latencies)

	avgLatency := time.Duration(0)
	if len(latencies) > 0 {
		var total time.Duration
		for _, l := range latencies {
			total += l
		}
		avgLatency = total / time.Duration(len(latencies))
	}

	avgDecrypt := totalDecrypt / time.Duration(successCount)
	avgScoring := totalScoring / time.Duration(successCount)

	t.Logf("=== VEID Verification Latency Baseline Test ===")
	t.Logf("Iterations: %d", iterations)
	t.Logf("Success Count: %d", successCount)
	t.Logf("Average Latency: %v (target: %v)", avgLatency, baseline.TargetLatency)
	t.Logf("P95 Latency: %v (max: %v)", p95, baseline.MaxLatencyP95)
	t.Logf("P99 Latency: %v (max: %v)", p99, baseline.MaxLatencyP99)
	t.Logf("Avg Decryption: %v (max: %v)", avgDecrypt, baseline.MaxDecryptionLatency)
	t.Logf("Avg Scoring: %v (max: %v)", avgScoring, baseline.MaxScoringLatency)

	// Assertions against baseline
	require.LessOrEqual(t, avgLatency, baseline.TargetLatency,
		"Average latency should meet target")
	require.LessOrEqual(t, p95, baseline.MaxLatencyP95,
		"P95 latency should be within acceptable limit")
	require.LessOrEqual(t, p99, baseline.MaxLatencyP99,
		"P99 latency should be within acceptable limit")
}

// calculateVerificationPercentiles calculates P95 and P99 latencies
func calculateVerificationPercentiles(latencies []time.Duration) (p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return 0, 0
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// Simple insertion sort for small arrays
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	p95Idx := int(float64(len(sorted)) * 0.95)
	p99Idx := int(float64(len(sorted)) * 0.99)

	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	return sorted[p95Idx], sorted[p99Idx]
}

// ============================================================================
// Identity Record Benchmarks
// ============================================================================

// BenchmarkIdentityRecordCreate benchmarks identity record creation
func BenchmarkIdentityRecordCreate(b *testing.B) {
	k, ctx := setupVEIDKeeperForBenchmark(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := sdk.AccAddress([]byte(fmt.Sprintf("bench-addr-%d", i)))
		record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
		err := k.SetIdentityRecord(ctx, *record)
		if err != nil {
			b.Fatalf("failed to create identity record: %v", err)
		}
	}
}

// BenchmarkIdentityRecordGet benchmarks identity record retrieval
func BenchmarkIdentityRecordGet(b *testing.B) {
	k, ctx := setupVEIDKeeperForBenchmark(b)

	// Pre-create records
	const numRecords = 1000
	addrs := make([]sdk.AccAddress, numRecords)
	for i := 0; i < numRecords; i++ {
		addr := sdk.AccAddress([]byte(fmt.Sprintf("bench-addr-%d", i)))
		addrs[i] = addr
		record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
		_ = k.SetIdentityRecord(ctx, *record)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := addrs[i%numRecords]
		_, _ = k.GetIdentityRecord(ctx, addr)
	}
}

// BenchmarkScoreUpdate benchmarks score updates
func BenchmarkScoreUpdate(b *testing.B) {
	k, ctx := setupVEIDKeeperForBenchmark(b)

	// Pre-create identity
	addr := sdk.AccAddress([]byte("bench-score-addr"))
	record := types.NewIdentityRecord(addr.String(), ctx.BlockTime())
	_ = k.SetIdentityRecord(ctx, *record)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		score := int32(50 + (i % 50))
		err := k.SetScore(ctx, addr.String(), score, "v1.0.0")
		if err != nil {
			b.Fatalf("failed to set score: %v", err)
		}
	}
}

// BenchmarkIdentityRecordSerialization benchmarks record serialization
func BenchmarkIdentityRecordSerialization(b *testing.B) {
	now := time.Now().UTC()
	record := types.NewIdentityRecord("cosmos1testaddr", now)
	record.CurrentScore = 75
	record.Tier = types.IdentityTierStandard

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(record)
		if err != nil {
			b.Fatalf("failed to serialize: %v", err)
		}
	}
}

// BenchmarkIdentityRecordDeserialization benchmarks record deserialization
func BenchmarkIdentityRecordDeserialization(b *testing.B) {
	now := time.Now().UTC()
	record := types.NewIdentityRecord("cosmos1testaddr", now)
	record.CurrentScore = 75
	data, _ := json.Marshal(record)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var r types.IdentityRecord
		err := json.Unmarshal(data, &r)
		if err != nil {
			b.Fatalf("failed to deserialize: %v", err)
		}
	}
}

// ============================================================================
// Hash Computation Benchmarks
// ============================================================================

// BenchmarkInputHashComputation benchmarks input hash computation
func BenchmarkInputHashComputation(b *testing.B) {
	data := make([]byte, 4096)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := sha256.Sum256(data)
		_ = hex.EncodeToString(h[:])
	}
}

// BenchmarkFeatureVectorHash benchmarks feature vector hashing
func BenchmarkFeatureVectorHash(b *testing.B) {
	features := make([]float64, 512) // 512-dim feature vector
	for i := range features {
		features[i] = float64(i) / 512.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := sha256.New()
		for _, f := range features {
			h.Write([]byte(fmt.Sprintf("%.6f", f)))
		}
		_ = h.Sum(nil)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func setupVEIDKeeperForBenchmark(b *testing.B) (keeper.Keeper, sdk.Context) {
	b.Helper()

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		b.Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	k := keeper.NewKeeper(cdc, storeKey, "authority")

	// Set default params
	err = k.SetParams(ctx, types.DefaultParams())
	if err != nil {
		b.Fatalf("failed to set params: %v", err)
	}

	return k, ctx
}
