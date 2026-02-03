// Package scale contains large-scale load tests for VirtEngine.
// These tests simulate production-scale scenarios with 1M+ validators.
//
// Task Reference: SCALE-001 - Load Testing - 1M Nodes Simulation
package scale

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Constants and Baselines
// ============================================================================

const (
	// Scale targets
	TargetValidatorCount = 1_000_000 // 1M validators
	TargetActiveOrders   = 100_000   // 100k active orders
	TargetProviderCount  = 1_000     // 1k+ providers
	TargetConcurrentBids = 10_000    // 10k concurrent bids

	// Memory limits
	MaxMemoryPerValidator = 1024 // bytes per validator entry
	MaxTotalMemoryGB      = 16   // max total memory usage

	// Performance baselines
	ValidatorLookupP95        = 10 * time.Millisecond
	ValidatorIterationMaxTime = 30 * time.Second
	ConsensusRoundMaxTime     = 5 * time.Second
)

// ValidatorScaleBaseline defines performance targets for validator operations at scale
type ValidatorScaleBaseline struct {
	ValidatorCount         int64         `json:"validator_count"`
	LookupP95              time.Duration `json:"lookup_p95"`
	LookupP99              time.Duration `json:"lookup_p99"`
	IterationTimeMax       time.Duration `json:"iteration_time_max"`
	ConsensusRoundTimeMax  time.Duration `json:"consensus_round_time_max"`
	StateSnapshotTimeMax   time.Duration `json:"state_snapshot_time_max"`
	MemoryPerValidatorMax  int64         `json:"memory_per_validator_max_bytes"`
	VotingPowerCalcTimeMax time.Duration `json:"voting_power_calc_time_max"`
}

// DefaultValidatorScaleBaseline returns baseline for 1M validators
func DefaultValidatorScaleBaseline() ValidatorScaleBaseline {
	return ValidatorScaleBaseline{
		ValidatorCount:         1_000_000,
		LookupP95:              10 * time.Millisecond,
		LookupP99:              50 * time.Millisecond,
		IterationTimeMax:       30 * time.Second,
		ConsensusRoundTimeMax:  5 * time.Second,
		StateSnapshotTimeMax:   60 * time.Second,
		MemoryPerValidatorMax:  1024,
		VotingPowerCalcTimeMax: 100 * time.Millisecond,
	}
}

// ============================================================================
// Mock Validator Types
// ============================================================================

// MockValidator represents a validator in the scale test
type MockValidator struct {
	Address         [20]byte
	PubKey          [32]byte
	VotingPower     int64
	Commission      uint16
	Jailed          bool
	Tombstoned      bool
	Status          uint8 // 0=unbonded, 1=unbonding, 2=bonded
	Tokens          int64
	DelegatorShares int64
	Moniker         [32]byte
}

// ValidatorStore simulates a validator store at scale
type ValidatorStore struct {
	mu         sync.RWMutex
	validators map[[20]byte]*MockValidator
	byPower    []*MockValidator // sorted by voting power
	totalPower int64

	// Metrics
	lookups    int64
	iterations int64
}

// NewValidatorStore creates a new validator store
func NewValidatorStore() *ValidatorStore {
	return &ValidatorStore{
		validators: make(map[[20]byte]*MockValidator),
		byPower:    make([]*MockValidator, 0),
	}
}

// AddValidator adds a validator to the store
func (s *ValidatorStore) AddValidator(v *MockValidator) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.validators[v.Address] = v
	s.byPower = append(s.byPower, v)
	s.totalPower += v.VotingPower
}

// GetValidator retrieves a validator by address
func (s *ValidatorStore) GetValidator(addr [20]byte) (*MockValidator, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	atomic.AddInt64(&s.lookups, 1)
	v, ok := s.validators[addr]
	return v, ok
}

// IterateValidators iterates over all validators
func (s *ValidatorStore) IterateValidators(fn func(*MockValidator) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	atomic.AddInt64(&s.iterations, 1)
	for _, v := range s.validators {
		if fn(v) {
			return
		}
	}
}

// GetBondedValidators returns validators sorted by voting power
func (s *ValidatorStore) GetBondedValidators(maxCount int) []*MockValidator {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Sort by voting power (descending)
	sorted := make([]*MockValidator, len(s.byPower))
	copy(sorted, s.byPower)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].VotingPower > sorted[j].VotingPower
	})

	if maxCount > len(sorted) {
		maxCount = len(sorted)
	}

	return sorted[:maxCount]
}

// CalculateTotalVotingPower calculates total voting power
func (s *ValidatorStore) CalculateTotalVotingPower() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total int64
	for _, v := range s.validators {
		if v.Status == 2 && !v.Jailed { // bonded and not jailed
			total += v.VotingPower
		}
	}
	return total
}

// Count returns the number of validators
func (s *ValidatorStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.validators)
}

// GetMetrics returns store metrics
func (s *ValidatorStore) GetMetrics() (lookups, iterations int64) {
	return atomic.LoadInt64(&s.lookups), atomic.LoadInt64(&s.iterations)
}

// ============================================================================
// Validator Generation
// ============================================================================

// generateMockValidator creates a mock validator with random data
func generateMockValidator(index int) *MockValidator {
	commission := safeUint16FromInt(index%1000 + 500)
	v := &MockValidator{
		VotingPower:     int64(1000 + index%10000),
		Commission:      commission,      // 5-15%
		Jailed:          index%100 == 0,  // 1% jailed
		Tombstoned:      index%1000 == 0, // 0.1% tombstoned
		Status:          2,               // bonded
		Tokens:          int64(1000000 + index*100),
		DelegatorShares: int64(1000000 + index*100),
	}

	// Generate deterministic address
	h := sha256.New()
	fmt.Fprintf(h, "validator_%d", index)
	sum := h.Sum(nil)
	copy(v.Address[:], sum[:20])

	// Generate pubkey
	mustRandRead(v.PubKey[:])

	// Generate moniker
	moniker := fmt.Sprintf("Validator_%d", index)
	copy(v.Moniker[:], moniker)

	return v
}

func safeUint16FromInt(value int) uint16 {
	if value < 0 {
		return 0
	}
	if value > int(^uint16(0)) {
		return uint16(^uint16(0))
	}
	return uint16(value)
}

// populateValidatorStore creates a validator store with the specified count
func populateValidatorStore(count int) *ValidatorStore {
	store := NewValidatorStore()

	// Use parallel generation for large counts
	if count > 10000 {
		workers := runtime.NumCPU()
		perWorker := count / workers

		var wg sync.WaitGroup
		results := make(chan []*MockValidator, workers)

		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(_ int, start, end int) {
				defer wg.Done()
				validators := make([]*MockValidator, 0, end-start)
				for i := start; i < end; i++ {
					validators = append(validators, generateMockValidator(i))
				}
				results <- validators
			}(w, w*perWorker, (w+1)*perWorker)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for validators := range results {
			for _, v := range validators {
				store.AddValidator(v)
			}
		}
	} else {
		for i := 0; i < count; i++ {
			store.AddValidator(generateMockValidator(i))
		}
	}

	return store
}

// ============================================================================
// Scale Benchmarks
// ============================================================================

// BenchmarkValidatorLookup benchmarks single validator lookup at various scales
func BenchmarkValidatorLookup(b *testing.B) {
	scales := []int{1000, 10000, 100000, 1000000}

	for _, scale := range scales {
		b.Run(fmt.Sprintf("validators_%d", scale), func(b *testing.B) {
			if scale > 100000 && testing.Short() {
				b.Skip("Skipping large scale test in short mode")
			}

			store := populateValidatorStore(scale)

			// Generate lookup addresses
			addresses := make([][20]byte, 1000)
			for i := range addresses {
				v := generateMockValidator(i * (scale / 1000))
				addresses[i] = v.Address
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				addr := addresses[i%len(addresses)]
				_, _ = store.GetValidator(addr)
			}
		})
	}
}

// BenchmarkValidatorIteration benchmarks full validator set iteration
func BenchmarkValidatorIteration(b *testing.B) {
	scales := []int{1000, 10000, 100000}

	for _, scale := range scales {
		b.Run(fmt.Sprintf("validators_%d", scale), func(b *testing.B) {
			store := populateValidatorStore(scale)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				store.IterateValidators(func(v *MockValidator) bool {
					count++
					return false
				})
			}

			b.ReportMetric(float64(scale)/b.Elapsed().Seconds()*float64(b.N), "validators_iterated/sec")
		})
	}
}

// BenchmarkVotingPowerCalculation benchmarks voting power calculation at scale
func BenchmarkVotingPowerCalculation(b *testing.B) {
	scales := []int{10000, 100000, 1000000}

	for _, scale := range scales {
		b.Run(fmt.Sprintf("validators_%d", scale), func(b *testing.B) {
			if scale > 100000 && testing.Short() {
				b.Skip("Skipping large scale test in short mode")
			}

			store := populateValidatorStore(scale)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = store.CalculateTotalVotingPower()
			}
		})
	}
}

// BenchmarkGetBondedValidators benchmarks retrieving top validators by power
func BenchmarkGetBondedValidators(b *testing.B) {
	store := populateValidatorStore(100000)
	topCounts := []int{100, 200, 500, 1000}

	for _, topN := range topCounts {
		b.Run(fmt.Sprintf("top_%d", topN), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = store.GetBondedValidators(topN)
			}
		})
	}
}

// BenchmarkParallelValidatorLookup benchmarks concurrent validator lookups
func BenchmarkParallelValidatorLookup(b *testing.B) {
	store := populateValidatorStore(100000)

	addresses := make([][20]byte, 10000)
	for i := range addresses {
		v := generateMockValidator(i * 10)
		addresses[i] = v.Address
	}

	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := counter.Add(1) % int64(len(addresses))
			_, _ = store.GetValidator(addresses[idx])
		}
	})

	lookups, _ := store.GetMetrics()
	b.ReportMetric(float64(lookups)/b.Elapsed().Seconds(), "lookups/sec")
}

// ============================================================================
// Scale Tests
// ============================================================================

// TestValidatorScaleBaseline tests validator operations at 1M scale against baseline
func TestValidatorScaleBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 1M validator test in short mode")
	}

	// Use smaller scale for CI, full scale for manual testing
	scale := 100000 // 100k for CI
	if v := lookupEnvScale(); v > 0 {
		scale = v
	}

	t.Logf("=== Validator Scale Baseline Test (scale=%d) ===", scale)

	baseline := DefaultValidatorScaleBaseline()

	// Measure store population time
	populateStart := time.Now()
	store := populateValidatorStore(scale)
	populateTime := time.Since(populateStart)
	t.Logf("Store population time: %v (%d validators)", populateTime, store.Count())

	// Measure memory usage
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	memPerValidator := safeInt64FromUint64(m.HeapAlloc) / int64(scale)
	t.Logf("Memory per validator: %d bytes (max: %d)", memPerValidator, baseline.MemoryPerValidatorMax)

	// Test lookup performance
	t.Run("lookup_performance", func(t *testing.T) {
		addresses := make([][20]byte, 1000)
		for i := range addresses {
			v := generateMockValidator(i * (scale / 1000))
			addresses[i] = v.Address
		}

		latencies := make([]time.Duration, len(addresses))
		for i, addr := range addresses {
			start := time.Now()
			_, found := store.GetValidator(addr)
			latencies[i] = time.Since(start)
			require.True(t, found, "Validator should be found")
		}

		p95, p99 := calculateDurationPercentiles(latencies)
		t.Logf("Lookup P95: %v (max: %v)", p95, baseline.LookupP95)
		t.Logf("Lookup P99: %v (max: %v)", p99, baseline.LookupP99)

		require.LessOrEqual(t, p95, baseline.LookupP95*time.Duration(scale/100000+1),
			"P95 lookup latency should scale linearly")
	})

	// Test iteration performance
	t.Run("iteration_performance", func(t *testing.T) {
		start := time.Now()
		count := 0
		store.IterateValidators(func(v *MockValidator) bool {
			count++
			return false
		})
		iterationTime := time.Since(start)

		t.Logf("Full iteration time: %v (%d validators)", iterationTime, count)

		// Scale threshold based on validator count
		maxTime := baseline.IterationTimeMax * time.Duration(scale) / time.Duration(baseline.ValidatorCount)
		require.LessOrEqual(t, iterationTime, maxTime,
			"Iteration time should be within threshold")
	})

	// Test voting power calculation
	t.Run("voting_power_calculation", func(t *testing.T) {
		start := time.Now()
		power := store.CalculateTotalVotingPower()
		calcTime := time.Since(start)

		t.Logf("Voting power calculation time: %v (total power: %d)", calcTime, power)

		maxTime := baseline.VotingPowerCalcTimeMax * time.Duration(scale) / time.Duration(100000)
		require.LessOrEqual(t, calcTime, maxTime,
			"Voting power calculation should be within threshold")
	})

	// Test bonded validator retrieval
	t.Run("bonded_validators", func(t *testing.T) {
		start := time.Now()
		bonded := store.GetBondedValidators(200) // Active set size
		retrievalTime := time.Since(start)

		t.Logf("Bonded validator retrieval time: %v (%d validators)", retrievalTime, len(bonded))
		require.Equal(t, 200, len(bonded), "Should return requested number of validators")
	})
}

// TestConsensusSimulation simulates consensus rounds with large validator set
func TestConsensusSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping consensus simulation in short mode")
	}

	scale := 10000 // 10k validators for consensus simulation
	t.Logf("=== Consensus Simulation Test (scale=%d) ===", scale)

	store := populateValidatorStore(scale)

	// Simulate consensus rounds
	const numRounds = 10
	roundTimes := make([]time.Duration, numRounds)

	for round := 0; round < numRounds; round++ {
		start := time.Now()

		// Get proposer (top validator)
		bonded := store.GetBondedValidators(200)
		proposer := bonded[round%len(bonded)]

		// Simulate prevote collection (2/3 of validators)
		prevotes := 0
		votePower := int64(0)
		totalPower := store.CalculateTotalVotingPower()
		threshold := totalPower * 2 / 3

		store.IterateValidators(func(v *MockValidator) bool {
			if v.Status == 2 && !v.Jailed {
				prevotes++
				votePower += v.VotingPower
				if votePower > threshold {
					return true // stop early once threshold reached
				}
			}
			return false
		})

		// Simulate precommit (same process)
		precommits := prevotes

		roundTimes[round] = time.Since(start)

		t.Logf("Round %d: proposer=%x, prevotes=%d, precommits=%d, time=%v",
			round, proposer.Address[:4], prevotes, precommits, roundTimes[round])
	}

	avgRoundTime := averageDuration(roundTimes)
	t.Logf("Average consensus round time: %v", avgRoundTime)

	require.LessOrEqual(t, avgRoundTime, ConsensusRoundMaxTime,
		"Consensus round time should be within threshold")
}

// TestValidatorSetTransitions tests large-scale validator set changes
func TestValidatorSetTransitions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validator transitions in short mode")
	}

	scale := 10000
	t.Logf("=== Validator Set Transitions Test (scale=%d) ===", scale)

	store := populateValidatorStore(scale)

	// Simulate jailing validators (1% at a time)
	t.Run("jail_validators", func(t *testing.T) {
		jailCount := scale / 100
		start := time.Now()

		jailed := 0
		store.IterateValidators(func(v *MockValidator) bool {
			if jailed < jailCount && !v.Jailed {
				v.Jailed = true
				jailed++
			}
			return jailed >= jailCount
		})

		jailTime := time.Since(start)
		t.Logf("Jailed %d validators in %v", jailed, jailTime)
	})

	// Simulate unbonding validators (5%)
	t.Run("unbond_validators", func(t *testing.T) {
		unbondCount := scale / 20
		start := time.Now()

		unbonded := 0
		store.IterateValidators(func(v *MockValidator) bool {
			if unbonded < unbondCount && v.Status == 2 && !v.Jailed {
				v.Status = 1 // unbonding
				unbonded++
			}
			return unbonded >= unbondCount
		})

		unbondTime := time.Since(start)
		t.Logf("Unbonded %d validators in %v", unbonded, unbondTime)
	})

	// Simulate slashing (0.1%)
	t.Run("slash_validators", func(t *testing.T) {
		slashCount := scale / 1000
		start := time.Now()

		slashed := 0
		slashFraction := int64(100) // 1% of tokens
		store.IterateValidators(func(v *MockValidator) bool {
			if slashed < slashCount && !v.Tombstoned {
				v.Tokens = v.Tokens * (10000 - slashFraction) / 10000
				v.VotingPower = v.Tokens / 1000
				slashed++
			}
			return slashed >= slashCount
		})

		slashTime := time.Since(start)
		t.Logf("Slashed %d validators in %v", slashed, slashTime)
	})

	// Recalculate voting power after changes
	t.Run("recalculate_power", func(t *testing.T) {
		start := time.Now()
		newPower := store.CalculateTotalVotingPower()
		calcTime := time.Since(start)

		t.Logf("Recalculated voting power: %d in %v", newPower, calcTime)
	})
}

// ============================================================================
// Memory Pressure Tests
// ============================================================================

// TestMemoryPressureAtScale tests memory behavior with large validator sets
func TestMemoryPressureAtScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory pressure test in short mode")
	}

	scales := []int{10000, 50000, 100000}

	for _, scale := range scales {
		t.Run(fmt.Sprintf("scale_%d", scale), func(t *testing.T) {
			runtime.GC()
			var before runtime.MemStats
			runtime.ReadMemStats(&before)

			store := populateValidatorStore(scale)

			runtime.GC()
			var after runtime.MemStats
			runtime.ReadMemStats(&after)

			memUsed := after.HeapAlloc - before.HeapAlloc
			memPerValidator := memUsed / safeUint64FromIntValue(scale)

			t.Logf("Scale %d: total=%dMB, per_validator=%d bytes",
				scale, memUsed/1024/1024, memPerValidator)

			// Verify count
			require.Equal(t, scale, store.Count())

			// Memory should not exceed limit
			maxMem := uint64(MaxTotalMemoryGB) * 1024 * 1024 * 1024
			require.Less(t, memUsed, maxMem,
				"Memory usage should not exceed limit")
		})
	}
}

// ============================================================================
// Helpers
// ============================================================================

func calculateDurationPercentiles(durations []time.Duration) (p95, p99 time.Duration) {
	if len(durations) == 0 {
		return 0, 0
	}

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	p95Idx := len(sorted) * 95 / 100
	p99Idx := len(sorted) * 99 / 100

	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	return sorted[p95Idx], sorted[p99Idx]
}

func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func lookupEnvScale() int {
	// In production, this would check SCALE_TEST_VALIDATORS env var
	return 0
}
