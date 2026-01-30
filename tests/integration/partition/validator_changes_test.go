//go:build e2e.integration

package partition

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil/partition"
)

// ValidatorChangesTestSuite tests validator set changes during network partitions.
type ValidatorChangesTestSuite struct {
	suite.Suite

	controller *partition.Controller
	nodes      []partition.NodeID
	metrics    *partition.Metrics
}

// TestValidatorChanges runs the validator changes test suite.
func TestValidatorChanges(t *testing.T) {
	suite.Run(t, new(ValidatorChangesTestSuite))
}

// SetupSuite runs once before all tests.
func (s *ValidatorChangesTestSuite) SetupSuite() {
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
func (s *ValidatorChangesTestSuite) SetupTest() {
	s.controller.Heal()
	s.metrics.Reset()
}

// TearDownTest runs after each test.
func (s *ValidatorChangesTestSuite) TearDownTest() {
	s.controller.Heal()
}

// =============================================================================
// Validator Join During Partition Tests
// =============================================================================

// TestValidatorJoinDuringPartition tests adding a validator during partition.
func (s *ValidatorChangesTestSuite) TestValidatorJoinDuringPartition() {
	s.T().Log("=== Test: Validator Join During Partition ===")

	// Apply partition
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	require.Equal(s.T(), len(s.nodes), s.controller.NodeCount())

	// New validator joins during partition
	newValidator := partition.NodeID("validator-new")
	s.controller.AddNode(newValidator)

	require.Equal(s.T(), len(s.nodes)+1, s.controller.NodeCount())

	// New validator should be able to reach all nodes
	// (since it wasn't in the original partition groups)
	for _, node := range s.nodes {
		require.True(s.T(), s.controller.CanReach(newValidator, node),
			"New validator should reach %s", node)
	}

	// Heal partition
	s.controller.Heal()

	// After heal, verify full connectivity
	allNodes := append(s.nodes, newValidator)
	for _, from := range allNodes {
		for _, to := range allNodes {
			require.True(s.T(), s.controller.CanReach(from, to),
				"Full connectivity should be restored")
		}
	}

	// Cleanup
	s.controller.RemoveNode(newValidator)

	s.T().Log("Validator join during partition test completed")
}

// TestValidatorJoinMajorityGroup tests a validator joining the majority group.
func (s *ValidatorChangesTestSuite) TestValidatorJoinMajorityGroup() {
	s.T().Log("=== Test: Validator Join Majority Group ===")

	// Create majority/minority partition
	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	majorityGroup := scenario.Groups[0]
	minorityGroup := scenario.Groups[1]

	s.T().Logf("Majority: %d nodes, Minority: %d nodes",
		len(majorityGroup.Nodes), len(minorityGroup.Nodes))

	// New validator joins (by default can reach all)
	newValidator := partition.NodeID("validator-joining-majority")
	s.controller.AddNode(newValidator)

	// Simulate the new validator joining the majority group
	// by blocking connections to minority
	for _, minNode := range minorityGroup.Nodes {
		s.controller.ApplyAsymmetricPartition(newValidator, minNode, true)
		s.controller.ApplyAsymmetricPartition(minNode, newValidator, true)
	}

	// Verify new validator can reach majority
	for _, majNode := range majorityGroup.Nodes {
		require.True(s.T(), s.controller.CanReach(newValidator, majNode),
			"New validator should reach majority node %s", majNode)
	}

	// Verify new validator cannot reach minority
	for _, minNode := range minorityGroup.Nodes {
		require.False(s.T(), s.controller.CanReach(newValidator, minNode),
			"New validator should NOT reach minority node %s", minNode)
	}

	// This effectively increases the majority group size
	effectiveMajority := len(majorityGroup.Nodes) + 1
	totalNodes := len(s.nodes) + 1
	s.T().Logf("Effective majority now: %d/%d", effectiveMajority, totalNodes)

	// Cleanup
	s.controller.RemoveNode(newValidator)
	s.T().Log("Validator join majority group test completed")
}

// TestMultipleValidatorsJoinDuringPartition tests multiple validators joining.
func (s *ValidatorChangesTestSuite) TestMultipleValidatorsJoinDuringPartition() {
	s.T().Log("=== Test: Multiple Validators Join During Partition ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	newValidators := []partition.NodeID{
		"validator-new-1",
		"validator-new-2",
		"validator-new-3",
	}

	for _, v := range newValidators {
		s.controller.AddNode(v)
	}

	require.Equal(s.T(), len(s.nodes)+len(newValidators), s.controller.NodeCount())

	s.controller.Heal()

	// All nodes including new ones should have full connectivity
	allNodes := s.controller.Nodes()
	for _, from := range allNodes {
		for _, to := range allNodes {
			require.True(s.T(), s.controller.CanReach(from, to))
		}
	}

	// Cleanup
	for _, v := range newValidators {
		s.controller.RemoveNode(v)
	}

	s.T().Logf("Added and removed %d validators during partition", len(newValidators))
}

// =============================================================================
// Validator Leave During Partition Tests
// =============================================================================

// TestValidatorLeaveDuringPartition tests a validator leaving during partition.
func (s *ValidatorChangesTestSuite) TestValidatorLeaveDuringPartition() {
	s.T().Log("=== Test: Validator Leave During Partition ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	initialCount := s.controller.NodeCount()

	// Remove a validator from one of the groups
	leavingValidator := scenario.Groups[0].Nodes[0]
	s.controller.RemoveNode(leavingValidator)

	require.Equal(s.T(), initialCount-1, s.controller.NodeCount())

	// Leaving validator should not be in connectivity matrix
	matrix := s.controller.GetConnectivityMatrix()
	_, exists := matrix[leavingValidator]
	require.False(s.T(), exists, "Leaving validator should not be in matrix")

	// Heal partition
	s.controller.Heal()

	// Remaining nodes should have full connectivity
	remainingNodes := s.controller.Nodes()
	for _, from := range remainingNodes {
		for _, to := range remainingNodes {
			require.True(s.T(), s.controller.CanReach(from, to))
		}
	}

	// Re-add validator
	s.controller.AddNode(leavingValidator)

	s.T().Log("Validator leave during partition test completed")
}

// TestValidatorLeaveMajorityAffectsQuorum tests quorum impact when validator leaves.
func (s *ValidatorChangesTestSuite) TestValidatorLeaveMajorityAffectsQuorum() {
	s.T().Log("=== Test: Validator Leave Affects Quorum ===")

	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	majorityGroup := scenario.Groups[0]
	originalMajoritySize := len(majorityGroup.Nodes)

	// If a majority node leaves, it might affect consensus capability
	leavingNode := majorityGroup.Nodes[0]
	s.controller.RemoveNode(leavingNode)

	newMajoritySize := originalMajoritySize - 1
	totalNodes := len(s.nodes) - 1
	quorumRatio := float64(newMajoritySize) / float64(totalNodes)

	s.T().Logf("After leaving: majority=%d/%d (%.2f%%)",
		newMajoritySize, totalNodes, quorumRatio*100)

	// Record this as a significant event
	// In production, this would trigger quorum alerts
	if quorumRatio < 0.67 {
		s.T().Log("WARNING: Majority group may no longer have quorum")
	}

	// Cleanup
	s.controller.AddNode(leavingNode)
	s.controller.Heal()

	s.T().Log("Quorum impact test completed")
}

// TestAllMinorityValidatorsLeave tests what happens when all minority nodes leave.
func (s *ValidatorChangesTestSuite) TestAllMinorityValidatorsLeave() {
	s.T().Log("=== Test: All Minority Validators Leave ===")

	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	minorityNodes := scenario.Groups[1].Nodes

	// All minority nodes leave
	for _, node := range minorityNodes {
		s.controller.RemoveNode(node)
	}

	remainingCount := s.controller.NodeCount()
	require.Equal(s.T(), len(s.nodes)-len(minorityNodes), remainingCount)

	// Heal should work with remaining nodes
	s.controller.Heal()

	remainingNodes := s.controller.Nodes()
	for _, from := range remainingNodes {
		for _, to := range remainingNodes {
			require.True(s.T(), s.controller.CanReach(from, to))
		}
	}

	// Re-add minority nodes
	for _, node := range minorityNodes {
		s.controller.AddNode(node)
	}

	s.T().Logf("Removed and restored %d minority validators", len(minorityNodes))
}

// =============================================================================
// Validator Rejoin After Partition Tests
// =============================================================================

// TestValidatorRejoinAfterPartition tests a validator leaving and rejoining.
func (s *ValidatorChangesTestSuite) TestValidatorRejoinAfterPartition() {
	s.T().Log("=== Test: Validator Rejoin After Partition ===")

	scenario := partition.CreateIsolatedNodePartition(s.nodes, 0)
	s.controller.ApplyPartition(scenario.Groups)

	isolatedNode := s.nodes[0]

	// Isolated node "leaves" (e.g., crashes)
	s.controller.RemoveNode(isolatedNode)

	// Main network continues
	time.Sleep(50 * time.Millisecond)

	// Heal partition (for remaining nodes)
	s.controller.Heal()

	// Node rejoins
	s.controller.AddNode(isolatedNode)

	// Rejoined node should be able to sync with network
	for _, other := range s.nodes[1:] {
		require.True(s.T(), s.controller.CanReach(isolatedNode, other),
			"Rejoined node should reach %s", other)
	}

	// Record state sync
	s.metrics.RecordStateSync(isolatedNode, 100, 150, 50*time.Millisecond, true)

	s.T().Log("Validator rejoin after partition test completed")
}

// TestValidatorSetTransition tests a complete validator set transition.
func (s *ValidatorChangesTestSuite) TestValidatorSetTransition() {
	s.T().Log("=== Test: Validator Set Transition ===")

	// This simulates a scenario where the validator set is gradually replaced

	newValidators := []partition.NodeID{
		"new-validator-0",
		"new-validator-1",
		"new-validator-2",
		"new-validator-3",
	}

	// Add new validators one by one
	for i, nv := range newValidators {
		s.controller.AddNode(nv)

		// Remove old validator
		if i < len(s.nodes) {
			s.controller.RemoveNode(s.nodes[i])
		}

		currentCount := s.controller.NodeCount()
		s.T().Logf("Transition step %d: %d active validators", i+1, currentCount)
	}

	// Verify final state
	require.Equal(s.T(), len(newValidators), s.controller.NodeCount())

	// All new validators should have full connectivity
	for _, from := range newValidators {
		for _, to := range newValidators {
			require.True(s.T(), s.controller.CanReach(from, to))
		}
	}

	// Restore original validators
	for _, nv := range newValidators {
		s.controller.RemoveNode(nv)
	}
	for _, v := range s.nodes {
		s.controller.AddNode(v)
	}

	s.T().Log("Validator set transition test completed")
}

// TestPartitionDuringValidatorSetChange tests partition occurring during set change.
func (s *ValidatorChangesTestSuite) TestPartitionDuringValidatorSetChange() {
	s.T().Log("=== Test: Partition During Validator Set Change ===")

	// Add a new validator
	newValidator := partition.NodeID("new-validator")
	s.controller.AddNode(newValidator)

	// Partition occurs during the transition
	allNodes := append(s.nodes, newValidator)
	scenario := partition.CreateSimplePartition(allNodes)
	s.controller.ApplyPartition(scenario.Groups)

	require.Equal(s.T(), partition.PartitionStatePartitioned, s.controller.State())

	// Simulate validator set update being propagated
	// Some nodes might not know about the new validator yet
	s.T().Log("Partition occurred during validator set change")

	// Heal partition
	s.controller.Heal()

	// After heal, all nodes including new validator should be connected
	for _, from := range allNodes {
		for _, to := range allNodes {
			require.True(s.T(), s.controller.CanReach(from, to),
				"All nodes should be connected after heal")
		}
	}

	// Cleanup
	s.controller.RemoveNode(newValidator)

	s.T().Log("Partition during validator set change test completed")
}
