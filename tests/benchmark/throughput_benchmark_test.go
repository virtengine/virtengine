// Package benchmark contains performance benchmarks for VirtEngine.
// Task Reference: PERF-001 - Performance Benchmarking Suite
package benchmark

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Transaction Throughput Benchmarks
// Target: 10k TPS
// ============================================================================

// TransactionThroughputBaseline defines the baseline metrics for throughput tests
type TransactionThroughputBaseline struct {
	TargetTPS         int64         `json:"target_tps"`
	MinAcceptableTPS  int64         `json:"min_acceptable_tps"`
	MaxLatencyP95     time.Duration `json:"max_latency_p95"`
	MaxLatencyP99     time.Duration `json:"max_latency_p99"`
	MaxErrorRate      float64       `json:"max_error_rate"`
}

// DefaultTransactionBaseline returns baseline metrics for transaction throughput
func DefaultTransactionBaseline() TransactionThroughputBaseline {
	return TransactionThroughputBaseline{
		TargetTPS:         10000,
		MinAcceptableTPS:  5000,
		MaxLatencyP95:     100 * time.Millisecond,
		MaxLatencyP99:     250 * time.Millisecond,
		MaxErrorRate:      0.001, // 0.1%
	}
}

// MockTransaction simulates a transaction for benchmarking
type MockTransaction struct {
	ID        string
	Sender    string
	Receiver  string
	Amount    int64
	Nonce     uint64
	Signature []byte
	Hash      []byte
	GasLimit  uint64
	Timestamp time.Time
}

// NewMockTransaction creates a new mock transaction
func NewMockTransaction(nonce uint64) *MockTransaction {
	tx := &MockTransaction{
		ID:        generateTxID(),
		Sender:    "cosmos1sender" + fmt.Sprintf("%d", nonce%1000),
		Receiver:  "cosmos1receiver" + fmt.Sprintf("%d", nonce%1000),
		Amount:    int64(1000 + nonce%10000),
		Nonce:     nonce,
		GasLimit:  200000,
		Timestamp: time.Now().UTC(),
	}
	tx.Signature = make([]byte, 64)
	rand.Read(tx.Signature)
	tx.computeHash()
	return tx
}

func (tx *MockTransaction) computeHash() {
	h := sha256.New()
	h.Write([]byte(tx.ID))
	h.Write([]byte(tx.Sender))
	h.Write([]byte(tx.Receiver))
	h.Write([]byte(fmt.Sprintf("%d", tx.Amount)))
	h.Write([]byte(fmt.Sprintf("%d", tx.Nonce)))
	tx.Hash = h.Sum(nil)
}

func generateTxID() string {
	id := make([]byte, 16)
	rand.Read(id)
	return hex.EncodeToString(id)
}

// MockTransactionProcessor simulates transaction processing
type MockTransactionProcessor struct {
	mu              sync.Mutex
	processedCount  int64
	failedCount     int64
	latencies       []time.Duration
	stateRoot       []byte
	validationDelay time.Duration
	executionDelay  time.Duration
}

// NewMockTransactionProcessor creates a new mock processor
func NewMockTransactionProcessor() *MockTransactionProcessor {
	return &MockTransactionProcessor{
		latencies:       make([]time.Duration, 0, 100000),
		stateRoot:       make([]byte, 32),
		validationDelay: 10 * time.Microsecond,
		executionDelay:  50 * time.Microsecond,
	}
}

// ProcessTransaction processes a single transaction
func (p *MockTransactionProcessor) ProcessTransaction(tx *MockTransaction) error {
	start := time.Now()

	// Simulate transaction validation
	if err := p.validateTransaction(tx); err != nil {
		atomic.AddInt64(&p.failedCount, 1)
		return err
	}

	// Simulate transaction execution
	if err := p.executeTransaction(tx); err != nil {
		atomic.AddInt64(&p.failedCount, 1)
		return err
	}

	// Record latency
	latency := time.Since(start)
	p.mu.Lock()
	p.latencies = append(p.latencies, latency)
	p.mu.Unlock()

	atomic.AddInt64(&p.processedCount, 1)
	return nil
}

func (p *MockTransactionProcessor) validateTransaction(tx *MockTransaction) error {
	// Simulate signature verification
	time.Sleep(p.validationDelay)

	// Verify transaction fields
	if tx.Amount <= 0 {
		return fmt.Errorf("invalid amount")
	}
	if len(tx.Signature) != 64 {
		return fmt.Errorf("invalid signature length")
	}
	return nil
}

func (p *MockTransactionProcessor) executeTransaction(tx *MockTransaction) error {
	// Simulate state update
	time.Sleep(p.executionDelay)

	// Update state root
	h := sha256.New()
	h.Write(p.stateRoot)
	h.Write(tx.Hash)
	p.mu.Lock()
	p.stateRoot = h.Sum(nil)
	p.mu.Unlock()

	return nil
}

// GetStats returns processing statistics
func (p *MockTransactionProcessor) GetStats() (processed, failed int64, latencies []time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.processedCount, p.failedCount, p.latencies
}

// BenchmarkTransactionValidation benchmarks transaction validation
func BenchmarkTransactionValidation(b *testing.B) {
	processor := NewMockTransactionProcessor()
	tx := NewMockTransaction(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.validateTransaction(tx)
	}
}

// BenchmarkTransactionExecution benchmarks transaction execution
func BenchmarkTransactionExecution(b *testing.B) {
	processor := NewMockTransactionProcessor()
	tx := NewMockTransaction(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.executeTransaction(tx)
	}
}

// BenchmarkTransactionProcessing benchmarks full transaction processing
func BenchmarkTransactionProcessing(b *testing.B) {
	processor := NewMockTransactionProcessor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewMockTransaction(uint64(i))
		_ = processor.ProcessTransaction(tx)
	}
}

// BenchmarkTransactionProcessingParallel benchmarks parallel transaction processing
func BenchmarkTransactionProcessingParallel(b *testing.B) {
	processor := NewMockTransactionProcessor()
	var nonce atomic.Uint64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := nonce.Add(1)
			tx := NewMockTransaction(n)
			_ = processor.ProcessTransaction(tx)
		}
	})

	processed, failed, _ := processor.GetStats()
	b.ReportMetric(float64(processed)/b.Elapsed().Seconds(), "tps")
	b.ReportMetric(float64(failed)/float64(processed+failed)*100, "error_rate_%")
}

// BenchmarkTransactionHashComputation benchmarks transaction hash computation
func BenchmarkTransactionHashComputation(b *testing.B) {
	tx := NewMockTransaction(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.computeHash()
	}
}

// BenchmarkBlockProcessing benchmarks processing a full block of transactions
func BenchmarkBlockProcessing(b *testing.B) {
	processor := NewMockTransactionProcessor()
	const txPerBlock = 1000

	// Pre-generate transactions
	txs := make([]*MockTransaction, txPerBlock)
	for i := range txs {
		txs[i] = NewMockTransaction(uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tx := range txs {
			_ = processor.ProcessTransaction(tx)
		}
	}

	b.ReportMetric(float64(txPerBlock), "txs_per_block")
}

// TestTransactionThroughputBaseline tests transaction throughput against baseline
func TestTransactionThroughputBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping throughput baseline test in short mode")
	}

	baseline := DefaultTransactionBaseline()
	processor := NewMockTransactionProcessor()

	// Run throughput test
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const workerCount = 100
	var wg sync.WaitGroup
	var nonce atomic.Uint64

	start := time.Now()

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n := nonce.Add(1)
					tx := NewMockTransaction(n)
					_ = processor.ProcessTransaction(tx)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	processed, failed, latencies := processor.GetStats()
	tps := float64(processed) / elapsed.Seconds()
	errorRate := float64(failed) / float64(processed+failed)

	// Calculate percentiles
	p95, p99 := calculateLatencyPercentiles(latencies)

	t.Logf("=== Transaction Throughput Baseline Test ===")
	t.Logf("Duration: %v", elapsed)
	t.Logf("Transactions Processed: %d", processed)
	t.Logf("Transactions Failed: %d", failed)
	t.Logf("TPS: %.2f (target: %d, min: %d)", tps, baseline.TargetTPS, baseline.MinAcceptableTPS)
	t.Logf("P95 Latency: %v (max: %v)", p95, baseline.MaxLatencyP95)
	t.Logf("P99 Latency: %v (max: %v)", p99, baseline.MaxLatencyP99)
	t.Logf("Error Rate: %.4f%% (max: %.4f%%)", errorRate*100, baseline.MaxErrorRate*100)

	// Assertions against baseline
	require.GreaterOrEqual(t, tps, float64(baseline.MinAcceptableTPS),
		"TPS should meet minimum acceptable threshold")
	require.LessOrEqual(t, errorRate, baseline.MaxErrorRate,
		"Error rate should be within acceptable limit")
}

// calculateLatencyPercentiles calculates P95 and P99 latencies
func calculateLatencyPercentiles(latencies []time.Duration) (p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return 0, 0
	}

	// Sort using insertion sort for small arrays
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
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
// State Write Benchmarks
// ============================================================================

// MockStateStore simulates state storage for benchmarking
type MockStateStore struct {
	mu    sync.RWMutex
	store map[string][]byte
}

// NewMockStateStore creates a new mock state store
func NewMockStateStore() *MockStateStore {
	return &MockStateStore{
		store: make(map[string][]byte),
	}
}

// Set stores a key-value pair
func (s *MockStateStore) Set(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
}

// Get retrieves a value by key
func (s *MockStateStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.store[key]
	return v, ok
}

// BenchmarkStateWrite benchmarks state write operations
func BenchmarkStateWrite(b *testing.B) {
	store := NewMockStateStore()
	value := make([]byte, 256)
	rand.Read(value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		store.Set(key, value)
	}
}

// BenchmarkStateRead benchmarks state read operations
func BenchmarkStateRead(b *testing.B) {
	store := NewMockStateStore()
	const numKeys = 10000
	value := make([]byte, 256)
	rand.Read(value)

	// Pre-populate store
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key_%d", i)
		store.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%numKeys)
		store.Get(key)
	}
}

// BenchmarkStateWriteParallel benchmarks parallel state write operations
func BenchmarkStateWriteParallel(b *testing.B) {
	store := NewMockStateStore()
	value := make([]byte, 256)
	rand.Read(value)

	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := counter.Add(1)
			key := fmt.Sprintf("key_%d", i)
			store.Set(key, value)
		}
	})
}

// ============================================================================
// Concurrent Access Benchmarks
// ============================================================================

// BenchmarkConcurrentProcessing benchmarks processing with varying concurrency
func BenchmarkConcurrentProcessing(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8, 16, 32, 64}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			runtime.GOMAXPROCS(concurrency)
			processor := NewMockTransactionProcessor()
			var nonce atomic.Uint64

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					n := nonce.Add(1)
					tx := NewMockTransaction(n)
					_ = processor.ProcessTransaction(tx)
				}
			})
		})
	}
}
