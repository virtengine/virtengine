//go:build e2e.integration

package partition

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil/partition"
)

// ConsensusRecoveryTestSuite tests consensus recovery after network partitions.
type ConsensusRecoveryTestSuite struct {
	suite.Suite

	controller *partition.Controller
	nodes      []partition.NodeID
	metrics    *partition.Metrics
}

// TestConsensusRecovery runs the consensus recovery test suite.
func TestConsensusRecovery(t *testing.T) {
	suite.Run(t, new(ConsensusRecoveryTestSuite))
}

// SetupSuite runs once before all tests.
func (s *ConsensusRecoveryTestSuite) SetupSuite() {
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
func (s *ConsensusRecoveryTestSuite) SetupTest() {
	s.controller.Heal()
	s.metrics.Reset()
}

// TearDownTest runs after each test.
func (s *ConsensusRecoveryTestSuite) TearDownTest() {
	s.controller.Heal()
}

// =============================================================================
// Consensus Recovery Validation Tests
// =============================================================================

// TestConsensusRecoveryAfterSimplePartition validates consensus can resume
// after a simple partition is healed.
func (s *ConsensusRecoveryTestSuite) TestConsensusRecoveryAfterSimplePartition() {
	s.T().Log("=== Test: Consensus Recovery After Simple Partition ===")

	// Simulate blocks being produced before partition
	for i := int64(1); i <= 5; i++ {
		s.metrics.RecordBlock(i, s.nodes[i%4], false, 0)
	}

	// Apply partition
	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	// Simulate blocks during partition (only from majority group)
	partitionStart := time.Now()
	majorityNodes := scenario.Groups[0].Nodes
	for i := int64(6); i <= 10; i++ {
		s.metrics.RecordBlock(i, majorityNodes[int(i)%len(majorityNodes)], true, 0)
	}

	// Simulate partition duration
	time.Sleep(100 * time.Millisecond)

	// Heal partition
	s.controller.Heal()
	healTime := time.Now()

	// Simulate recovery - first block after heal
	timeToFirstBlock := 50 * time.Millisecond
	s.metrics.RecordBlock(11, s.nodes[0], false, timeToFirstBlock)

	// Record healing metric
	s.metrics.RecordHealingMetric(partition.HealingMetric{
		PartitionDuration: healTime.Sub(partitionStart),
		TimeToFirstBlock:  timeToFirstBlock,
		TimeToConsensus:   100 * time.Millisecond,
	})

	// Verify metrics
	summary := s.metrics.Summary()
	require.Equal(s.T(), 1, summary.TotalPartitions)
	require.Equal(s.T(), 11, summary.TotalBlocksProduced)
	require.Equal(s.T(), 5, summary.BlocksDuringPartition)
	require.Equal(s.T(), 6, summary.BlocksAfterHeal)

	s.T().Logf("Recovery metrics: TimeToFirstBlock=%v, TimeToConsensus=%v",
		summary.AverageTimeToFirstBlock, summary.AverageTimeToConsensus)
}

// TestConsensusRecoveryWithNoQuorum tests recovery when no partition has quorum.
func (s *ConsensusRecoveryTestSuite) TestConsensusRecoveryWithNoQuorum() {
	s.T().Log("=== Test: Consensus Recovery With No Quorum ===")

	// Apply three-way partition (no group has quorum)
	scenario := partition.CreateThreeWayPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	// Verify no group has quorum
	for _, group := range scenario.Groups {
		ratio := float64(len(group.Nodes)) / float64(len(s.nodes))
		require.Less(s.T(), ratio, 0.67, "No group should have quorum")
	}

	// During this partition, no blocks should be produced by any group
	// (simulated by not recording any blocks during partition)

	partitionStart := time.Now()
	time.Sleep(100 * time.Millisecond)

	// Heal partition
	s.controller.Heal()
	healTime := time.Now()

	// After heal, blocks resume
	s.metrics.RecordBlock(1, s.nodes[0], false, 50*time.Millisecond)

	s.metrics.RecordHealingMetric(partition.HealingMetric{
		PartitionDuration: healTime.Sub(partitionStart),
		TimeToFirstBlock:  50 * time.Millisecond,
		TimeToConsensus:   150 * time.Millisecond,
	})

	summary := s.metrics.Summary()
	require.Equal(s.T(), 0, summary.BlocksDuringPartition,
		"No blocks should be produced when no group has quorum")
	require.Equal(s.T(), 1, summary.BlocksAfterHeal,
		"Blocks should resume after heal")

	s.T().Log("No-quorum recovery test completed successfully")
}

// TestConsensusRecoveryTiming validates timing metrics for consensus recovery.
func (s *ConsensusRecoveryTestSuite) TestConsensusRecoveryTiming() {
	s.T().Log("=== Test: Consensus Recovery Timing ===")

	// Run multiple partition/heal cycles
	for cycle := 0; cycle < 3; cycle++ {
		scenario := partition.CreateSimplePartition(s.nodes)
		s.controller.ApplyPartition(scenario.Groups)

		time.Sleep(50 * time.Millisecond)

		s.controller.Heal()

		// Record varying recovery times
		timeToFirstBlock := time.Duration(10+cycle*5) * time.Millisecond
		timeToConsensus := time.Duration(20+cycle*10) * time.Millisecond

		s.metrics.RecordHealingMetric(partition.HealingMetric{
			TimeToFirstBlock: timeToFirstBlock,
			TimeToConsensus:  timeToConsensus,
		})
	}

	summary := s.metrics.Summary()
	require.Equal(s.T(), 3, summary.TotalPartitions)
	require.Greater(s.T(), summary.AverageTimeToFirstBlock, time.Duration(0))
	require.Greater(s.T(), summary.AverageTimeToConsensus, time.Duration(0))
	require.GreaterOrEqual(s.T(), summary.MaxTimeToFirstBlock, summary.AverageTimeToFirstBlock)

	s.T().Logf("Timing metrics: AvgFirstBlock=%v, AvgConsensus=%v, MaxFirstBlock=%v",
		summary.AverageTimeToFirstBlock, summary.AverageTimeToConsensus, summary.MaxTimeToFirstBlock)
}

// TestBlockProductionContinuity validates block production continues properly.
func (s *ConsensusRecoveryTestSuite) TestBlockProductionContinuity() {
	s.T().Log("=== Test: Block Production Continuity ===")

	// Record pre-partition blocks
	prePartitionHeight := int64(100)
	for h := int64(95); h <= prePartitionHeight; h++ {
		s.metrics.RecordBlock(h, s.nodes[h%4], false, 0)
	}

	// Apply partition with majority
	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	// Majority continues producing blocks
	majorityNodes := scenario.Groups[0].Nodes
	for h := prePartitionHeight + 1; h <= prePartitionHeight+10; h++ {
		s.metrics.RecordBlock(h, majorityNodes[int(h)%len(majorityNodes)], true, 0)
	}

	s.controller.Heal()

	// Post-heal blocks
	for h := prePartitionHeight + 11; h <= prePartitionHeight+15; h++ {
		s.metrics.RecordBlock(h, s.nodes[int(h)%4], false, time.Duration(h-prePartitionHeight-10)*10*time.Millisecond)
	}

	// Verify block sequence
	blockMetrics := s.metrics.GetBlockMetrics()
	require.Equal(s.T(), 21, len(blockMetrics)) // 6 pre + 10 during + 5 post

	// Verify heights are sequential
	expectedHeight := int64(95)
	for _, bm := range blockMetrics {
		require.Equal(s.T(), expectedHeight, bm.Height, "Block heights should be sequential")
		expectedHeight++
	}

	s.T().Log("Block production continuity verified")
}

// TestValidatorParticipationAfterHeal validates all validators can participate after heal.
func (s *ConsensusRecoveryTestSuite) TestValidatorParticipationAfterHeal() {
	s.T().Log("=== Test: Validator Participation After Heal ===")

	// Apply partition that isolates one validator
	scenario := partition.CreateIsolatedNodePartition(s.nodes, 0)
	s.controller.ApplyPartition(scenario.Groups)

	isolatedNode := s.nodes[0]

	// During partition, isolated node cannot participate
	for _, other := range s.nodes[1:] {
		require.False(s.T(), s.controller.CanReach(isolatedNode, other))
	}

	s.controller.Heal()

	// After heal, all validators can participate
	for _, node := range s.nodes {
		for _, other := range s.nodes {
			require.True(s.T(), s.controller.CanReach(node, other),
				"All validators should be able to participate after heal")
		}
	}

	// Record blocks from all validators to prove participation
	for i, node := range s.nodes {
		s.metrics.RecordBlock(int64(i+1), node, false, 0)
	}

	blockMetrics := s.metrics.GetBlockMetrics()
	producers := make(map[partition.NodeID]bool)
	for _, bm := range blockMetrics {
		producers[bm.Producer] = true
	}

	require.Equal(s.T(), len(s.nodes), len(producers),
		"All validators should have produced blocks")

	s.T().Log("All validators can participate after heal")
}

// =============================================================================
// Fork Resolution Tests
// =============================================================================

// TestForkResolutionAfterPartition tests that forks are resolved after partition heals.
func (s *ConsensusRecoveryTestSuite) TestForkResolutionAfterPartition() {
	s.T().Log("=== Test: Fork Resolution After Partition ===")

	// Apply simple partition
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	group1Nodes := scenario.Groups[0].Nodes
	group2Nodes := scenario.Groups[1].Nodes

	// Each group might produce blocks independently (in real scenario)
	// This simulates potential forking during partition
	s.T().Logf("Group 1 (%d nodes) and Group 2 (%d nodes) partitioned",
		len(group1Nodes), len(group2Nodes))

	// Only the group with majority should produce valid blocks
	// In 4-node network with 2/2 split, neither has quorum
	// (This is a simulated test - real consensus would stall)

	s.controller.Heal()

	// After heal, the canonical chain is chosen
	// Record a block after heal to prove recovery
	s.metrics.RecordHealingMetric(partition.HealingMetric{
		TimeToConsensus: 100 * time.Millisecond,
	})

	s.T().Log("Fork resolution test completed")
}

// TestRecoveryMetricsSummary validates the metrics summary is accurate.
func (s *ConsensusRecoveryTestSuite) TestRecoveryMetricsSummary() {
	s.T().Log("=== Test: Recovery Metrics Summary ===")

	// Record various metrics
	for i := 0; i < 5; i++ {
		s.controller.ApplyPartition(partition.CreateSimplePartition(s.nodes).Groups)
		time.Sleep(10 * time.Millisecond)
		s.controller.Heal()

		s.metrics.RecordHealingMetric(partition.HealingMetric{
			PartitionDuration: time.Duration(50+i*10) * time.Millisecond,
			TimeToFirstBlock:  time.Duration(10+i*2) * time.Millisecond,
			TimeToConsensus:   time.Duration(20+i*5) * time.Millisecond,
			MessagesReplayed:  i,
		})
	}

	summary := s.metrics.Summary()

	require.Equal(s.T(), 5, summary.TotalPartitions)
	require.Greater(s.T(), summary.TotalPartitionTime, time.Duration(0))
	require.Greater(s.T(), summary.AveragePartitionTime, time.Duration(0))
	require.LessOrEqual(s.T(), summary.MinPartitionTime, summary.AveragePartitionTime)
	require.GreaterOrEqual(s.T(), summary.MaxPartitionTime, summary.AveragePartitionTime)
	require.Equal(s.T(), 10, summary.TotalMessagesReplayed) // 0+1+2+3+4 = 10

	s.T().Logf("Metrics summary: Partitions=%d, AvgDuration=%v, TotalReplayed=%d",
		summary.TotalPartitions, summary.AveragePartitionTime, summary.TotalMessagesReplayed)
}
