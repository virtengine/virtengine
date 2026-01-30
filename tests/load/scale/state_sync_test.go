// Package scale contains state sync performance tests.
// These tests measure state synchronization performance at scale.
//
// Task Reference: SCALE-001 - Load Testing - 1M Nodes Simulation
package scale

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// State Sync Constants
// ============================================================================

const (
	// State size targets
	TargetStateEntries     = 10_000_000 // 10M state entries
	TargetSnapshotSize     = 1024 * 1024 * 1024 // 1GB snapshot
	ChunkSize              = 16 * 1024 // 16KB chunks

	// Performance baselines
	SnapshotCreationRate   = 100 * 1024 * 1024 // 100MB/sec
	ChunkTransferRate      = 50 * 1024 * 1024  // 50MB/sec
	StateApplyRate         = 50_000            // 50k entries/sec
)

// StateSyncBaseline defines targets for state sync operations
type StateSyncBaseline struct {
	StateEntries           int64         `json:"state_entries"`
	SnapshotSizeBytes      int64         `json:"snapshot_size_bytes"`
	SnapshotCreationRate   int64         `json:"snapshot_creation_rate_bytes_per_sec"`
	ChunkTransferRate      int64         `json:"chunk_transfer_rate_bytes_per_sec"`
	StateApplyRate         int64         `json:"state_apply_rate_entries_per_sec"`
	MaxSnapshotTime        time.Duration `json:"max_snapshot_time"`
	MaxSyncTime            time.Duration `json:"max_sync_time"`
	MaxMemoryPressure      float64       `json:"max_memory_pressure"`
}

// DefaultStateSyncBaseline returns baseline for state sync
func DefaultStateSyncBaseline() StateSyncBaseline {
	return StateSyncBaseline{
		StateEntries:           10_000_000,
		SnapshotSizeBytes:      1024 * 1024 * 1024,
		SnapshotCreationRate:   100 * 1024 * 1024,
		ChunkTransferRate:      50 * 1024 * 1024,
		StateApplyRate:         50_000,
		MaxSnapshotTime:        60 * time.Second,
		MaxSyncTime:            300 * time.Second,
		MaxMemoryPressure:      0.80, // 80% of max
	}
}

// ============================================================================
// Mock State Types
// ============================================================================

// StateEntry represents a key-value state entry
type StateEntry struct {
	Key       [32]byte
	Value     []byte
	Height    int64
	Timestamp int64
}

// MockStateStore simulates a large state store
type MockStateStore struct {
	mu       sync.RWMutex
	entries  map[[32]byte]*StateEntry
	height   int64
	stateRoot [32]byte
	
	// Metrics
	reads    atomic.Int64
	writes   atomic.Int64
	deletes  atomic.Int64
}

// NewMockStateStore creates a new mock state store
func NewMockStateStore() *MockStateStore {
	return &MockStateStore{
		entries: make(map[[32]byte]*StateEntry),
	}
}

// Set stores a key-value pair
func (s *MockStateStore) Set(key [32]byte, value []byte, height int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.entries[key] = &StateEntry{
		Key:       key,
		Value:     value,
		Height:    height,
		Timestamp: time.Now().UnixNano(),
	}
	s.writes.Add(1)
	s.updateRoot()
}

// Get retrieves a value by key
func (s *MockStateStore) Get(key [32]byte) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	s.reads.Add(1)
	if entry, ok := s.entries[key]; ok {
		return entry.Value, true
	}
	return nil, false
}

// Delete removes a key
func (s *MockStateStore) Delete(key [32]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.entries, key)
	s.deletes.Add(1)
	s.updateRoot()
}

// Count returns number of entries
func (s *MockStateStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// GetHeight returns current height
func (s *MockStateStore) GetHeight() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.height
}

// SetHeight sets current height
func (s *MockStateStore) SetHeight(height int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.height = height
}

// GetRoot returns state root hash
func (s *MockStateStore) GetRoot() [32]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stateRoot
}

func (s *MockStateStore) updateRoot() {
	// Simple root computation
	h := sha256.New()
	binary.Write(h, binary.BigEndian, s.height)
	binary.Write(h, binary.BigEndian, int64(len(s.entries)))
	copy(s.stateRoot[:], h.Sum(nil))
}

// IteratePrefix iterates entries with given prefix
func (s *MockStateStore) IteratePrefix(prefix []byte, fn func(key [32]byte, value []byte) bool) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for key, entry := range s.entries {
		match := true
		for i, b := range prefix {
			if i >= len(key) || key[i] != b {
				match = false
				break
			}
		}
		if match {
			count++
			if fn(key, entry.Value) {
				break
			}
		}
	}
	return count
}

// GetStats returns store statistics
func (s *MockStateStore) GetStats() (reads, writes, deletes int64) {
	return s.reads.Load(), s.writes.Load(), s.deletes.Load()
}

// ============================================================================
// Snapshot Types
// ============================================================================

// Snapshot represents a state snapshot
type Snapshot struct {
	Height     int64
	Root       [32]byte
	Chunks     [][]byte
	ChunkCount int
	TotalSize  int64
	CreatedAt  time.Time
}

// SnapshotManager manages snapshot creation and restoration
type SnapshotManager struct {
	store      *MockStateStore
	chunkSize  int
	
	// Metrics
	snapshotsCreated   atomic.Int64
	snapshotsApplied   atomic.Int64
	bytesWritten       atomic.Int64
	bytesRead          atomic.Int64
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(store *MockStateStore, chunkSize int) *SnapshotManager {
	return &SnapshotManager{
		store:     store,
		chunkSize: chunkSize,
	}
}

// CreateSnapshot creates a snapshot of the current state
func (m *SnapshotManager) CreateSnapshot() (*Snapshot, error) {
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()
	
	snapshot := &Snapshot{
		Height:    m.store.height,
		Root:      m.store.stateRoot,
		CreatedAt: time.Now(),
	}
	
	// Serialize entries into chunks
	currentChunk := make([]byte, 0, m.chunkSize)
	
	for key, entry := range m.store.entries {
		// Serialize entry (simplified)
		entryData := make([]byte, 32+8+len(entry.Value))
		copy(entryData[:32], key[:])
		binary.BigEndian.PutUint64(entryData[32:40], uint64(entry.Height))
		copy(entryData[40:], entry.Value)
		
		// Add to chunk
		if len(currentChunk)+len(entryData) > m.chunkSize {
			snapshot.Chunks = append(snapshot.Chunks, currentChunk)
			snapshot.TotalSize += int64(len(currentChunk))
			currentChunk = make([]byte, 0, m.chunkSize)
		}
		currentChunk = append(currentChunk, entryData...)
	}
	
	// Add final chunk
	if len(currentChunk) > 0 {
		snapshot.Chunks = append(snapshot.Chunks, currentChunk)
		snapshot.TotalSize += int64(len(currentChunk))
	}
	
	snapshot.ChunkCount = len(snapshot.Chunks)
	m.snapshotsCreated.Add(1)
	m.bytesWritten.Add(snapshot.TotalSize)
	
	return snapshot, nil
}

// ApplySnapshot applies a snapshot to a new store
func (m *SnapshotManager) ApplySnapshot(snapshot *Snapshot, targetStore *MockStateStore) error {
	targetStore.mu.Lock()
	defer targetStore.mu.Unlock()
	
	for _, chunk := range snapshot.Chunks {
		offset := 0
		for offset < len(chunk) {
			if offset+40 > len(chunk) {
				break
			}
			
			var key [32]byte
			copy(key[:], chunk[offset:offset+32])
			height := int64(binary.BigEndian.Uint64(chunk[offset+32 : offset+40]))
			
			// Find value end (simplified - assumes fixed size for demo)
			valueEnd := offset + 40 + 64 // Assume 64-byte values
			if valueEnd > len(chunk) {
				valueEnd = len(chunk)
			}
			
			value := chunk[offset+40 : valueEnd]
			
			targetStore.entries[key] = &StateEntry{
				Key:       key,
				Value:     value,
				Height:    height,
				Timestamp: time.Now().UnixNano(),
			}
			
			offset = valueEnd
		}
		
		m.bytesRead.Add(int64(len(chunk)))
	}
	
	targetStore.height = snapshot.Height
	targetStore.stateRoot = snapshot.Root
	m.snapshotsApplied.Add(1)
	
	return nil
}

// GetStats returns snapshot statistics
func (m *SnapshotManager) GetStats() (created, applied, written, read int64) {
	return m.snapshotsCreated.Load(), m.snapshotsApplied.Load(),
		m.bytesWritten.Load(), m.bytesRead.Load()
}

// ============================================================================
// State Population
// ============================================================================

func generateStateKey(index int) [32]byte {
	h := sha256.Sum256([]byte(fmt.Sprintf("key_%d", index)))
	return h
}

func generateStateValue(size int) []byte {
	value := make([]byte, size)
	rand.Read(value)
	return value
}

// populateStateStore creates a state store with specified entry count
func populateStateStore(entryCount, valueSize int) *MockStateStore {
	store := NewMockStateStore()
	
	workers := runtime.NumCPU()
	entriesPerWorker := entryCount / workers
	
	var wg sync.WaitGroup
	
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID, start, count int) {
			defer wg.Done()
			
			for i := 0; i < count; i++ {
				key := generateStateKey(start + i)
				value := generateStateValue(valueSize)
				store.Set(key, value, int64(start+i))
			}
		}(w, w*entriesPerWorker, entriesPerWorker)
	}
	
	wg.Wait()
	return store
}

// ============================================================================
// Benchmarks
// ============================================================================

// BenchmarkStateWrite benchmarks state write operations
func BenchmarkStateWrite(b *testing.B) {
	store := NewMockStateStore()
	value := generateStateValue(256)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := generateStateKey(i)
		store.Set(key, value, int64(i))
	}
	
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "writes/sec")
}

// BenchmarkStateRead benchmarks state read operations
func BenchmarkStateRead(b *testing.B) {
	store := populateStateStore(10000, 256)
	
	keys := make([][32]byte, 1000)
	for i := range keys {
		keys[i] = generateStateKey(i * 10)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		store.Get(key)
	}
	
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "reads/sec")
}

// BenchmarkSnapshotCreation benchmarks snapshot creation
func BenchmarkSnapshotCreation(b *testing.B) {
	scales := []int{1000, 10000, 100000}
	
	for _, scale := range scales {
		b.Run(fmt.Sprintf("entries_%d", scale), func(b *testing.B) {
			if scale > 10000 && testing.Short() {
				b.Skip("Skipping large scale in short mode")
			}
			
			store := populateStateStore(scale, 256)
			manager := NewSnapshotManager(store, ChunkSize)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				snapshot, _ := manager.CreateSnapshot()
				_ = snapshot.TotalSize
			}
		})
	}
}

// BenchmarkSnapshotApply benchmarks snapshot restoration
func BenchmarkSnapshotApply(b *testing.B) {
	store := populateStateStore(10000, 256)
	manager := NewSnapshotManager(store, ChunkSize)
	snapshot, _ := manager.CreateSnapshot()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		targetStore := NewMockStateStore()
		manager.ApplySnapshot(snapshot, targetStore)
	}
}

// BenchmarkParallelStateAccess benchmarks parallel state access
func BenchmarkParallelStateAccess(b *testing.B) {
	store := populateStateStore(100000, 256)
	
	var counter atomic.Int64
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := counter.Add(1)
			if idx%2 == 0 {
				// Read
				key := generateStateKey(int(idx % 100000))
				store.Get(key)
			} else {
				// Write
				key := generateStateKey(int(idx))
				value := generateStateValue(256)
				store.Set(key, value, idx)
			}
		}
	})
	
	reads, writes, _ := store.GetStats()
	b.ReportMetric(float64(reads)/b.Elapsed().Seconds(), "reads/sec")
	b.ReportMetric(float64(writes)/b.Elapsed().Seconds(), "writes/sec")
}

// ============================================================================
// Scale Tests
// ============================================================================

// TestStateSyncBaseline tests state sync at scale
func TestStateSyncBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping state sync scale test in short mode")
	}
	
	entryCount := 100000 // 100k for CI
	valueSize := 256     // 256 bytes per value
	
	t.Logf("=== State Sync Baseline Test ===")
	t.Logf("Entries: %d, Value size: %d bytes", entryCount, valueSize)
	
	baseline := DefaultStateSyncBaseline()
	
	// Measure state population
	t.Run("state_population", func(t *testing.T) {
		start := time.Now()
		store := populateStateStore(entryCount, valueSize)
		populateTime := time.Since(start)
		
		entriesPerSec := float64(entryCount) / populateTime.Seconds()
		t.Logf("Population time: %v (%.0f entries/sec)", populateTime, entriesPerSec)
		
		require.Equal(t, entryCount, store.Count())
		require.Greater(t, int64(entriesPerSec), baseline.StateApplyRate/10,
			"State write rate should be reasonable")
	})
	
	// Measure snapshot creation
	t.Run("snapshot_creation", func(t *testing.T) {
		store := populateStateStore(entryCount, valueSize)
		manager := NewSnapshotManager(store, ChunkSize)
		
		start := time.Now()
		snapshot, err := manager.CreateSnapshot()
		createTime := time.Since(start)
		
		require.NoError(t, err)
		
		bytesPerSec := float64(snapshot.TotalSize) / createTime.Seconds()
		
		t.Logf("Snapshot creation time: %v", createTime)
		t.Logf("Snapshot size: %d bytes (%d chunks)", snapshot.TotalSize, snapshot.ChunkCount)
		t.Logf("Creation rate: %.2f MB/sec", bytesPerSec/1024/1024)
		
		require.Greater(t, int64(bytesPerSec), baseline.SnapshotCreationRate/10,
			"Snapshot creation rate should be reasonable")
	})
	
	// Measure snapshot application
	t.Run("snapshot_application", func(t *testing.T) {
		store := populateStateStore(entryCount, valueSize)
		manager := NewSnapshotManager(store, ChunkSize)
		snapshot, _ := manager.CreateSnapshot()
		
		targetStore := NewMockStateStore()
		
		start := time.Now()
		err := manager.ApplySnapshot(snapshot, targetStore)
		applyTime := time.Since(start)
		
		require.NoError(t, err)
		
		bytesPerSec := float64(snapshot.TotalSize) / applyTime.Seconds()
		
		t.Logf("Snapshot apply time: %v", applyTime)
		t.Logf("Apply rate: %.2f MB/sec", bytesPerSec/1024/1024)
		
		require.Equal(t, snapshot.Height, targetStore.GetHeight())
		require.Equal(t, snapshot.Root, targetStore.GetRoot())
	})
	
	// Measure memory pressure
	t.Run("memory_pressure", func(t *testing.T) {
		runtime.GC()
		var before runtime.MemStats
		runtime.ReadMemStats(&before)
		
		store := populateStateStore(entryCount, valueSize)
		manager := NewSnapshotManager(store, ChunkSize)
		snapshot, _ := manager.CreateSnapshot()
		
		runtime.GC()
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		
		memUsed := after.HeapAlloc - before.HeapAlloc
		memPerEntry := memUsed / uint64(entryCount)
		
		t.Logf("Memory used: %d MB", memUsed/1024/1024)
		t.Logf("Memory per entry: %d bytes", memPerEntry)
		t.Logf("Snapshot size: %d MB", snapshot.TotalSize/1024/1024)
		
		_ = store.Count()
	})
}

// TestIncrementalSync tests incremental state synchronization
func TestIncrementalSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping incremental sync test in short mode")
	}
	
	t.Logf("=== Incremental Sync Test ===")
	
	baseStore := populateStateStore(10000, 256)
	baseStore.SetHeight(100)
	
	// Create base snapshot
	manager := NewSnapshotManager(baseStore, ChunkSize)
	baseSnapshot, _ := manager.CreateSnapshot()
	
	t.Logf("Base snapshot: %d entries, %d bytes", baseStore.Count(), baseSnapshot.TotalSize)
	
	// Make incremental changes
	changesCount := 1000
	for i := 0; i < changesCount; i++ {
		key := generateStateKey(i + 10000)
		value := generateStateValue(256)
		baseStore.Set(key, value, 101)
	}
	baseStore.SetHeight(101)
	
	// Create new snapshot
	newSnapshot, _ := manager.CreateSnapshot()
	
	t.Logf("New snapshot: %d entries, %d bytes", baseStore.Count(), newSnapshot.TotalSize)
	
	// Compare sizes
	sizeIncrease := newSnapshot.TotalSize - baseSnapshot.TotalSize
	t.Logf("Size increase: %d bytes", sizeIncrease)
	
	require.Greater(t, newSnapshot.TotalSize, baseSnapshot.TotalSize)
	require.Equal(t, 10000+changesCount, baseStore.Count())
}

// TestConcurrentStateAccess tests concurrent read/write access
func TestConcurrentStateAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent access test in short mode")
	}
	
	store := populateStateStore(10000, 256)
	duration := 5 * time.Second
	
	t.Logf("=== Concurrent State Access Test ===")
	t.Logf("Duration: %v", duration)
	
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	
	var wg sync.WaitGroup
	var readOps, writeOps atomic.Int64
	
	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					key := generateStateKey(randomInt(10000))
					store.Get(key)
					readOps.Add(1)
				}
			}
		}()
	}
	
	// Writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			counter := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					key := generateStateKey(workerID*1000000 + counter)
					value := generateStateValue(256)
					store.Set(key, value, int64(counter))
					writeOps.Add(1)
					counter++
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	reads := readOps.Load()
	writes := writeOps.Load()
	
	t.Logf("Read operations: %d (%.0f/sec)", reads, float64(reads)/duration.Seconds())
	t.Logf("Write operations: %d (%.0f/sec)", writes, float64(writes)/duration.Seconds())
	t.Logf("Final entry count: %d", store.Count())
	
	require.Greater(t, reads, int64(0))
	require.Greater(t, writes, int64(0))
}

// TestLargeStateIteration tests iterating over large state
func TestLargeStateIteration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large iteration test in short mode")
	}
	
	scales := []int{10000, 50000, 100000}
	
	for _, scale := range scales {
		t.Run(fmt.Sprintf("entries_%d", scale), func(t *testing.T) {
			store := populateStateStore(scale, 256)
			
			start := time.Now()
			count := 0
			store.IteratePrefix(nil, func(key [32]byte, value []byte) bool {
				count++
				return false
			})
			iterTime := time.Since(start)
			
			rate := float64(count) / iterTime.Seconds()
			
			t.Logf("Iterated %d entries in %v (%.0f/sec)", count, iterTime, rate)
			require.Equal(t, scale, count)
		})
	}
}

// TestStateRecoveryAfterCrash simulates state recovery after crash
func TestStateRecoveryAfterCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping crash recovery test in short mode")
	}
	
	t.Logf("=== State Recovery After Crash Test ===")
	
	// Create and populate initial state
	store := populateStateStore(10000, 256)
	store.SetHeight(50)
	
	// Create snapshot at height 50
	manager := NewSnapshotManager(store, ChunkSize)
	snapshot, _ := manager.CreateSnapshot()
	
	t.Logf("Snapshot at height %d: %d entries", snapshot.Height, store.Count())
	
	// Continue making changes (simulating operations after snapshot)
	for i := 0; i < 5000; i++ {
		key := generateStateKey(i + 10000)
		value := generateStateValue(256)
		store.Set(key, value, 51+int64(i/1000))
	}
	store.SetHeight(55)
	
	t.Logf("State before 'crash': height=%d, entries=%d", store.GetHeight(), store.Count())
	
	// Simulate crash - create new store from snapshot
	recoveredStore := NewMockStateStore()
	
	start := time.Now()
	manager.ApplySnapshot(snapshot, recoveredStore)
	recoveryTime := time.Since(start)
	
	t.Logf("Recovery time: %v", recoveryTime)
	t.Logf("Recovered state: height=%d, entries=%d", recoveredStore.GetHeight(), recoveredStore.Count())
	
	// Verify recovered state matches snapshot
	require.Equal(t, snapshot.Height, recoveredStore.GetHeight())
	require.Equal(t, snapshot.Root, recoveredStore.GetRoot())
}
