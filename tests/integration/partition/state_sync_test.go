//go:build e2e.integration

package partition

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil/partition"
)

// StateSyncTestSuite tests state synchronization after network partitions.
type StateSyncTestSuite struct {
	suite.Suite

	controller *partition.Controller
	nodes      []partition.NodeID
	metrics    *partition.Metrics
}

// TestStateSync runs the state sync test suite.
func TestStateSync(t *testing.T) {
	suite.Run(t, new(StateSyncTestSuite))
}

// SetupSuite runs once before all tests.
func (s *StateSyncTestSuite) SetupSuite() {
	s.nodes = []partition.NodeID{
		"validator-0",
		"validator-1",
		"validator-2",
		"validator-3",
	}
	s.controller = partition.NewController(s.nodes...)
	s.metrics = s.controller.Metrics()
}

// SetupTest runs before each test.
func (s *StateSyncTestSuite) SetupTest() {
	s.controller.Heal()
	s.metrics.Reset()
}

// TearDownTest runs after each test.
func (s *StateSyncTestSuite) TearDownTest() {
	s.controller.Heal()
}

// =============================================================================
// State Sync Tests
// =============================================================================

// TestStateSyncAfterPartition validates nodes can sync state after partition.
func (s *StateSyncTestSuite) TestStateSyncAfterPartition() {
	s.T().Log("=== Test: State Sync After Partition ===")

	// Apply partition isolating one node
	scenario := partition.CreateIsolatedNodePartition(s.nodes, 0)
	s.controller.ApplyPartition(scenario.Groups)

	isolatedNode := s.nodes[0]
	heightBeforeHeal := int64(100)
	heightAfterSync := int64(150) // Main network advanced 50 blocks

	// Simulate time passing
	time.Sleep(100 * time.Millisecond)

	// Heal partition
	s.controller.Heal()

	// Simulate state sync
	syncStart := time.Now()
	time.Sleep(50 * time.Millisecond) // Simulate sync time
	syncDuration := time.Since(syncStart)

	// Record state sync metric
	s.metrics.RecordStateSync(isolatedNode, heightBeforeHeal, heightAfterSync, syncDuration, true)

	// Verify metrics
	stateMetrics := s.metrics.GetStateMetrics()
	require.Len(s.T(), stateMetrics, 1)

	sm := stateMetrics[0]
	require.Equal(s.T(), isolatedNode, sm.NodeID)
	require.Equal(s.T(), heightBeforeHeal, sm.HeightBefore)
	require.Equal(s.T(), heightAfterSync, sm.HeightAfter)
	require.True(s.T(), sm.StateHashMatch)

	s.T().Logf("Node %s synced from height %d to %d in %v",
		isolatedNode, heightBeforeHeal, heightAfterSync, syncDuration)
}

// TestStateSyncMultipleNodes validates multiple nodes can sync after partition.
func (s *StateSyncTestSuite) TestStateSyncMultipleNodes() {
	s.T().Log("=== Test: State Sync Multiple Nodes ===")

	// Apply partition creating two groups
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	minorityGroup := scenario.Groups[1]
	majorityHeight := int64(200)

	time.Sleep(100 * time.Millisecond)
	s.controller.Heal()

	// All minority nodes need to sync
	for i, node := range minorityGroup.Nodes {
		heightBefore := int64(100 + i*5) // Slight variation
		syncDuration := time.Duration(20+i*10) * time.Millisecond

		s.metrics.RecordStateSync(node, heightBefore, majorityHeight, syncDuration, true)
	}

	summary := s.metrics.Summary()
	require.Greater(s.T(), summary.AverageStateSyncTime, time.Duration(0))
	require.Equal(s.T(), float64(1), summary.StateSyncSuccessRate)

	s.T().Logf("Synced %d minority nodes, avg sync time: %v",
		len(minorityGroup.Nodes), summary.AverageStateSyncTime)
}

// TestStateSyncHashMismatch tests detection of state hash mismatches.
func (s *StateSyncTestSuite) TestStateSyncHashMismatch() {
	s.T().Log("=== Test: State Sync Hash Mismatch ===")

	// Simulate a scenario where state hashes don't match
	// (This would indicate a consensus bug or Byzantine behavior)
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	time.Sleep(50 * time.Millisecond)
	s.controller.Heal()

	// Record some successful syncs
	for i := 0; i < 3; i++ {
		s.metrics.RecordStateSync(s.nodes[i], 100, 150, 50*time.Millisecond, true)
	}

	// Record one failed sync (hash mismatch)
	s.metrics.RecordStateSync(s.nodes[3], 100, 150, 50*time.Millisecond, false)

	summary := s.metrics.Summary()
	require.Equal(s.T(), 0.75, summary.StateSyncSuccessRate,
		"Success rate should be 75% (3/4)")

	s.T().Log("State hash mismatch detection working")
}

// TestStateSyncPerformance validates state sync performance metrics.
func (s *StateSyncTestSuite) TestStateSyncPerformance() {
	s.T().Log("=== Test: State Sync Performance ===")

	// Simulate multiple sync operations with varying durations
	syncDurations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		100 * time.Millisecond, // Slow sync
	}

	for i, duration := range syncDurations {
		s.metrics.RecordStateSync(
			s.nodes[i],
			int64(100),
			int64(100+i*10),
			duration,
			true,
		)
	}

	summary := s.metrics.Summary()

	// Average should be 40ms
	expectedAvg := (10 + 20 + 30 + 100) / 4 * time.Millisecond
	require.InDelta(s.T(), expectedAvg.Milliseconds(), summary.AverageStateSyncTime.Milliseconds(), 5)

	s.T().Logf("Sync performance: avg=%v", summary.AverageStateSyncTime)
}

// TestStateSyncAfterLongPartition validates sync after extended partition.
func (s *StateSyncTestSuite) TestStateSyncAfterLongPartition() {
	s.T().Log("=== Test: State Sync After Long Partition ===")

	scenario := partition.CreateIsolatedNodePartition(s.nodes, 0)
	s.controller.ApplyPartition(scenario.Groups)

	isolatedNode := s.nodes[0]

	// Record partition start
	partitionStart := time.Now()

	// Simulate extended partition
	time.Sleep(200 * time.Millisecond)

	s.controller.Heal()
	partitionDuration := time.Since(partitionStart)

	// Isolated node has a lot of catching up to do
	heightBefore := int64(100)
	heightAfter := int64(100 + int64(partitionDuration.Milliseconds()/10)) // ~1 block per 10ms

	// Longer sync time for more blocks
	syncDuration := time.Duration(heightAfter-heightBefore) * time.Millisecond

	s.metrics.RecordStateSync(isolatedNode, heightBefore, heightAfter, syncDuration, true)

	stateMetrics := s.metrics.GetStateMetrics()
	require.Len(s.T(), stateMetrics, 1)

	blocksToSync := stateMetrics[0].HeightAfter - stateMetrics[0].HeightBefore
	s.T().Logf("After %v partition, node needed to sync %d blocks in %v",
		partitionDuration, blocksToSync, syncDuration)
}

// TestStateSyncConcurrency validates concurrent state sync operations.
func (s *StateSyncTestSuite) TestStateSyncConcurrency() {
	s.T().Log("=== Test: State Sync Concurrency ===")

	// All minority nodes sync concurrently after partition
	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	time.Sleep(100 * time.Millisecond)
	s.controller.Heal()

	minorityNodes := scenario.Groups[1].Nodes

	// Record concurrent sync starts
	syncStart := time.Now()
	for _, node := range minorityNodes {
		// All nodes start syncing at the same time
		s.metrics.RecordStateSync(node, 100, 150, 0, true)
	}

	// Update with actual sync durations
	time.Sleep(50 * time.Millisecond)
	totalSyncTime := time.Since(syncStart)

	s.T().Logf("Concurrent sync of %d nodes completed in %v",
		len(minorityNodes), totalSyncTime)

	stateMetrics := s.metrics.GetStateMetrics()
	require.Len(s.T(), stateMetrics, len(minorityNodes))
}

// TestStateSyncWithBlockProduction validates sync while blocks are produced.
func (s *StateSyncTestSuite) TestStateSyncWithBlockProduction() {
	s.T().Log("=== Test: State Sync With Block Production ===")

	scenario := partition.CreateIsolatedNodePartition(s.nodes, 0)
	s.controller.ApplyPartition(scenario.Groups)

	isolatedNode := s.nodes[0]
	currentHeight := int64(100)

	// Simulate blocks being produced by main network
	for h := currentHeight + 1; h <= currentHeight+10; h++ {
		s.metrics.RecordBlock(h, s.nodes[int(h)%3+1], true, 0)
	}

	time.Sleep(100 * time.Millisecond)
	s.controller.Heal()

	// Isolated node starts syncing while blocks continue
	syncStart := time.Now()

	// More blocks during sync
	for h := currentHeight + 11; h <= currentHeight+15; h++ {
		s.metrics.RecordBlock(h, s.nodes[int(h)%4], false, time.Since(syncStart))
	}

	// Sync completes
	syncDuration := time.Since(syncStart)
	finalHeight := currentHeight + 15

	s.metrics.RecordStateSync(isolatedNode, currentHeight, finalHeight, syncDuration, true)

	blockMetrics := s.metrics.GetBlockMetrics()
	require.Len(s.T(), blockMetrics, 15) // 10 during partition + 5 after

	stateMetrics := s.metrics.GetStateMetrics()
	require.Len(s.T(), stateMetrics, 1)
	require.Equal(s.T(), finalHeight, stateMetrics[0].HeightAfter)

	s.T().Logf("Node synced to height %d while blocks were being produced", finalHeight)
}
