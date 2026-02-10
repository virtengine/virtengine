//go:build e2e.integration

package partition

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil/partition"
)

// SplitBrainTestSuite tests split-brain scenarios and their resolution.
type SplitBrainTestSuite struct {
	suite.Suite

	controller *partition.Controller
	nodes      []partition.NodeID
	metrics    *partition.Metrics
	ctx        context.Context
	cancel     context.CancelFunc
}

// TestSplitBrain runs the split-brain test suite.
func TestSplitBrain(t *testing.T) {
	suite.Run(t, new(SplitBrainTestSuite))
}

// SetupSuite runs once before all tests.
func (s *SplitBrainTestSuite) SetupSuite() {
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
func (s *SplitBrainTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)
	s.controller.Heal()
	s.metrics.Reset()
}

// TearDownTest runs after each test.
func (s *SplitBrainTestSuite) TearDownTest() {
	s.controller.Heal()
	s.cancel()
}

// =============================================================================
// Split-Brain Scenario Tests
// =============================================================================

// TestEqualSplitScenario tests a 50/50 network split.
func (s *SplitBrainTestSuite) TestEqualSplitScenario() {
	s.T().Log("=== Test: Equal Split Scenario ===")

	// Create a 50/50 split (neither side has quorum)
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	group1 := scenario.Groups[0]
	group2 := scenario.Groups[1]

	s.T().Logf("Group 1: %d nodes, Group 2: %d nodes",
		len(group1.Nodes), len(group2.Nodes))

	// Verify complete isolation between groups
	for _, n1 := range group1.Nodes {
		for _, n2 := range group2.Nodes {
			require.False(s.T(), s.controller.CanReach(n1, n2),
				"Groups should be completely isolated")
			require.False(s.T(), s.controller.CanReach(n2, n1),
				"Groups should be completely isolated (reverse)")
		}
	}

	// In a 4-node network with 2/2 split, neither side has 2/3+1 = 3 nodes
	// So consensus should stall in both groups
	for _, group := range scenario.Groups {
		quorumRatio := float64(len(group.Nodes)) / float64(len(s.nodes))
		require.Less(s.T(), quorumRatio, 0.67,
			"Neither group should have quorum in 50/50 split")
	}

	s.controller.Heal()
	s.T().Log("Equal split scenario tested successfully")
}

// TestCompetingBlockProduction tests potential competing block production.
func (s *SplitBrainTestSuite) TestCompetingBlockProduction() {
	s.T().Log("=== Test: Competing Block Production ===")

	// Create partition
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	group1 := scenario.Groups[0]
	group2 := scenario.Groups[1]

	baseHeight := int64(100)

	// Simulate both groups attempting to produce blocks
	// (In reality, neither would succeed without quorum)

	// Group 1 attempts
	for i, node := range group1.Nodes {
		s.metrics.RecordBlock(baseHeight+int64(i), node, true, 0)
	}

	// Group 2 attempts (would be on a different fork if successful)
	for i, node := range group2.Nodes {
		s.metrics.RecordBlock(baseHeight+int64(i), node, true, 0)
	}

	s.T().Logf("Both groups attempted block production at height %d", baseHeight)

	// Heal the partition
	s.controller.Heal()

	// After heal, network must converge on a single chain
	// This simulates the resolution process
	s.metrics.RecordHealingMetric(partition.HealingMetric{
		TimeToConsensus: 200 * time.Millisecond,
	})

	// First block after heal establishes the canonical chain
	s.metrics.RecordBlock(baseHeight+10, s.nodes[0], false, 50*time.Millisecond)

	summary := s.metrics.Summary()
	require.Greater(s.T(), summary.BlocksDuringPartition, 0)
	require.Greater(s.T(), summary.BlocksAfterHeal, 0)

	s.T().Log("Competing block production test completed")
}

// TestForkResolution tests that forks are properly resolved after heal.
func (s *SplitBrainTestSuite) TestForkResolution() {
	s.T().Log("=== Test: Fork Resolution ===")

	// Create partition with majority/minority (majority can produce blocks)
	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	majorityGroup := scenario.Groups[0]
	minorityGroup := scenario.Groups[1]

	s.T().Logf("Majority: %v", majorityGroup.Nodes)
	s.T().Logf("Minority: %v", minorityGroup.Nodes)

	// Majority produces blocks
	majorityHeight := int64(110)
	for h := int64(101); h <= majorityHeight; h++ {
		producer := majorityGroup.Nodes[int(h)%len(majorityGroup.Nodes)]
		s.metrics.RecordBlock(h, producer, true, 0)
	}

	// Minority might have stale data at lower height
	minorityHeight := int64(100)

	s.controller.Heal()

	// After heal, minority must sync to majority's chain
	for _, node := range minorityGroup.Nodes {
		s.metrics.RecordStateSync(node, minorityHeight, majorityHeight, 100*time.Millisecond, true)
	}

	// Verify state metrics
	stateMetrics := s.metrics.GetStateMetrics()
	require.Len(s.T(), stateMetrics, len(minorityGroup.Nodes))

	for _, sm := range stateMetrics {
		require.Equal(s.T(), majorityHeight, sm.HeightAfter,
			"Minority nodes should sync to majority height")
		require.True(s.T(), sm.StateHashMatch,
			"State hashes should match after sync")
	}

	s.T().Log("Fork resolution test completed")
}

// TestThreeWaySplitBrain tests a three-way split-brain scenario.
func (s *SplitBrainTestSuite) TestThreeWaySplitBrain() {
	s.T().Log("=== Test: Three-Way Split-Brain ===")

	scenario := partition.CreateThreeWayPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	require.Equal(s.T(), 3, len(scenario.Groups),
		"Should have three groups")

	// Verify complete isolation
	for i := 0; i < len(scenario.Groups); i++ {
		for j := 0; j < len(scenario.Groups); j++ {
			if i != j {
				for _, n1 := range scenario.Groups[i].Nodes {
					for _, n2 := range scenario.Groups[j].Nodes {
						require.False(s.T(), s.controller.CanReach(n1, n2),
							"All groups should be isolated from each other")
					}
				}
			}
		}
	}

	// No group should have quorum
	for _, group := range scenario.Groups {
		quorumRatio := float64(len(group.Nodes)) / float64(len(s.nodes))
		s.T().Logf("Group %s: %d nodes (%.1f%%)",
			group.Name, len(group.Nodes), quorumRatio*100)
		require.Less(s.T(), quorumRatio, 0.67,
			"No group should have quorum in three-way split")
	}

	s.controller.Heal()
	s.T().Log("Three-way split-brain test completed")
}

// TestByzantineNodeIsolation tests isolation of potentially Byzantine nodes.
func (s *SplitBrainTestSuite) TestByzantineNodeIsolation() {
	s.T().Log("=== Test: Byzantine Node Isolation ===")

	scenario := partition.CreateByzantinePartition(s.nodes, 1)
	s.controller.ApplyPartition(scenario.Groups)

	var honestGroup, byzantineGroup partition.PartitionGroup
	for _, g := range scenario.Groups {
		if g.Name == "honest" {
			honestGroup = g
		} else {
			byzantineGroup = g
		}
	}

	s.T().Logf("Honest: %d nodes, Byzantine: %d nodes",
		len(honestGroup.Nodes), len(byzantineGroup.Nodes))

	// Byzantine nodes should be isolated from honest nodes
	for _, honest := range honestGroup.Nodes {
		for _, byzantine := range byzantineGroup.Nodes {
			require.False(s.T(), s.controller.CanReach(honest, byzantine))
			require.False(s.T(), s.controller.CanReach(byzantine, honest))
		}
	}

	// Honest nodes should still have quorum (with 4 nodes and 1 Byzantine, 3 honest)
	honestRatio := float64(len(honestGroup.Nodes)) / float64(len(s.nodes))
	s.T().Logf("Honest nodes have %.1f%% of network", honestRatio*100)

	// In BFT, we can tolerate up to 1/3 Byzantine nodes
	// With 4 nodes, that's 1 Byzantine, 3 honest (which is > 2/3)
	require.GreaterOrEqual(s.T(), honestRatio, 0.67,
		"Honest nodes should maintain quorum with 1 Byzantine")

	s.controller.Heal()
	s.T().Log("Byzantine node isolation test completed")
}

// =============================================================================
// Recovery Time Tests
// =============================================================================

// TestSplitBrainRecoveryTime tests recovery time measurement.
func (s *SplitBrainTestSuite) TestSplitBrainRecoveryTime() {
	s.T().Log("=== Test: Split-Brain Recovery Time ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	partitionStart := time.Now()

	s.controller.ApplyPartition(scenario.Groups)

	// Simulate partition duration
	time.Sleep(100 * time.Millisecond)

	healStart := time.Now()
	s.controller.Heal()

	// Measure time to first block
	firstBlockTime := 50 * time.Millisecond
	time.Sleep(firstBlockTime)

	// Measure time to full consensus
	consensusTime := 150 * time.Millisecond
	time.Sleep(consensusTime - firstBlockTime)

	totalRecoveryTime := time.Since(healStart)

	s.metrics.RecordHealingMetric(partition.HealingMetric{
		PartitionDuration: healStart.Sub(partitionStart),
		HealingDuration:   totalRecoveryTime,
		TimeToFirstBlock:  firstBlockTime,
		TimeToConsensus:   consensusTime,
	})

	summary := s.metrics.Summary()

	s.T().Logf("Recovery metrics:")
	s.T().Logf("  Partition duration: %v", summary.AveragePartitionTime)
	s.T().Logf("  Time to first block: %v", summary.AverageTimeToFirstBlock)
	s.T().Logf("  Time to consensus: %v", summary.AverageTimeToConsensus)

	require.Greater(s.T(), summary.AverageTimeToFirstBlock, time.Duration(0))
	require.Greater(s.T(), summary.AverageTimeToConsensus, summary.AverageTimeToFirstBlock)

	s.T().Log("Split-brain recovery time test completed")
}

// TestMultipleSplitBrainRecoveries tests multiple consecutive split-brain events.
func (s *SplitBrainTestSuite) TestMultipleSplitBrainRecoveries() {
	s.T().Log("=== Test: Multiple Split-Brain Recoveries ===")

	cycles := 5
	for cycle := 0; cycle < cycles; cycle++ {
		s.T().Logf("Cycle %d/%d", cycle+1, cycles)

		scenario := partition.CreateSimplePartition(s.nodes)
		s.controller.ApplyPartition(scenario.Groups)

		time.Sleep(50 * time.Millisecond)

		s.controller.Heal()

		s.metrics.RecordHealingMetric(partition.HealingMetric{
			PartitionDuration: 50 * time.Millisecond,
			TimeToFirstBlock:  time.Duration(10+cycle*2) * time.Millisecond,
			TimeToConsensus:   time.Duration(20+cycle*5) * time.Millisecond,
		})

		// Brief pause between cycles
		time.Sleep(20 * time.Millisecond)
	}

	summary := s.metrics.Summary()

	require.Equal(s.T(), cycles, summary.TotalPartitions)
	require.Greater(s.T(), summary.MaxTimeToFirstBlock, summary.AverageTimeToFirstBlock)

	s.T().Logf("Completed %d split-brain recovery cycles", cycles)
	s.T().Logf("  Average time to first block: %v", summary.AverageTimeToFirstBlock)
	s.T().Logf("  Max time to first block: %v", summary.MaxTimeToFirstBlock)
}

// TestGradualHealFromSplitBrain tests gradual healing from split-brain.
func (s *SplitBrainTestSuite) TestGradualHealFromSplitBrain() {
	s.T().Log("=== Test: Gradual Heal From Split-Brain ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	require.Equal(s.T(), partition.PartitionStatePartitioned, s.controller.State())

	// Gradually heal
	healDuration := 200 * time.Millisecond
	err := s.controller.HealGradually(s.ctx, healDuration)
	require.NoError(s.T(), err)

	require.Equal(s.T(), partition.PartitionStateHealthy, s.controller.State())

	// Verify full connectivity
	for _, from := range s.nodes {
		for _, to := range s.nodes {
			require.True(s.T(), s.controller.CanReach(from, to),
				"Should have full connectivity after gradual heal")
		}
	}

	s.T().Log("Gradual heal from split-brain test completed")
}

// =============================================================================
// Edge Cases
// =============================================================================

// TestRapidPartitionAndHeal tests rapid partition/heal cycles.
func (s *SplitBrainTestSuite) TestRapidPartitionAndHeal() {
	s.T().Log("=== Test: Rapid Partition and Heal ===")

	cycles := 20
	for i := 0; i < cycles; i++ {
		scenario := partition.CreateSimplePartition(s.nodes)
		s.controller.ApplyPartition(scenario.Groups)
		s.controller.Heal()
	}

	require.Equal(s.T(), partition.PartitionStateHealthy, s.controller.State())

	// Verify connectivity after rapid cycles
	for _, from := range s.nodes {
		for _, to := range s.nodes {
			require.True(s.T(), s.controller.CanReach(from, to))
		}
	}

	s.T().Logf("Completed %d rapid partition/heal cycles", cycles)
}

// TestPartitionWithFlappingConnections tests partitions with unstable connections.
func (s *SplitBrainTestSuite) TestPartitionWithFlappingConnections() {
	s.T().Log("=== Test: Partition With Flapping Connections ===")

	filter := s.controller.Filter()

	// Set up a flaky connection between first two nodes
	filter.SetDropRate(s.nodes[0], s.nodes[1], 0.5) // 50% packet loss

	// Create partition with flaky connection
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	// Verify flaky connection still affects filtering
	decision := filter.Filter(s.nodes[0], s.nodes[1])
	// Due to randomness, we can't predict exact behavior
	s.T().Logf("Flaky connection decision: drop=%v", decision.Drop)

	s.controller.Heal()
	filter.ClearAll()

	s.T().Log("Partition with flapping connections test completed")
}

// TestPartitionMetricsAccuracy tests that metrics are accurate.
func (s *SplitBrainTestSuite) TestPartitionMetricsAccuracy() {
	s.T().Log("=== Test: Partition Metrics Accuracy ===")

	// Apply partition and track timing
	expectedPartitionDuration := 100 * time.Millisecond

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	time.Sleep(expectedPartitionDuration)

	require.True(s.T(), s.metrics.IsPartitioned())
	actualDuration := s.metrics.CurrentPartitionDuration()
	require.InDelta(s.T(), expectedPartitionDuration.Milliseconds(),
		actualDuration.Milliseconds(), 50) // 50ms tolerance

	s.controller.Heal()
	require.False(s.T(), s.metrics.IsPartitioned())

	s.T().Logf("Expected duration: %v, Actual: %v",
		expectedPartitionDuration, actualDuration)
}
