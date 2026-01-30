//go:build e2e.integration

// Package partition contains integration tests for network partition recovery.
// These tests validate that the blockchain can recover from network partitions
// and maintain consensus and state consistency.
package partition

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil/partition"
)

// PartitionRecoveryTestSuite tests network partition recovery scenarios.
type PartitionRecoveryTestSuite struct {
	suite.Suite

	// controller is the partition simulation controller.
	controller *partition.Controller

	// nodes is the list of simulated node IDs.
	nodes []partition.NodeID

	// ctx is the test context.
	ctx context.Context

	// cancel cancels the test context.
	cancel context.CancelFunc
}

// TestPartitionRecovery runs the partition recovery test suite.
func TestPartitionRecovery(t *testing.T) {
	suite.Run(t, new(PartitionRecoveryTestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *PartitionRecoveryTestSuite) SetupSuite() {
	// Create simulated nodes (representing validators in a real network)
	s.nodes = []partition.NodeID{
		"validator-0",
		"validator-1",
		"validator-2",
		"validator-3",
	}
	s.controller = partition.NewController(s.nodes...)
}

// SetupTest runs before each test.
func (s *PartitionRecoveryTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)
	s.controller.Metrics().Reset()
}

// TearDownTest runs after each test.
func (s *PartitionRecoveryTestSuite) TearDownTest() {
	s.controller.Heal()
	s.cancel()
}

// =============================================================================
// Basic Partition Tests
// =============================================================================

// TestControllerInitialization verifies the partition controller initializes correctly.
func (s *PartitionRecoveryTestSuite) TestControllerInitialization() {
	s.T().Log("=== Test: Controller Initialization ===")

	// Verify initial state is healthy
	require.Equal(s.T(), partition.PartitionStateHealthy, s.controller.State())

	// Verify all nodes are registered
	require.Equal(s.T(), len(s.nodes), s.controller.NodeCount())

	// Verify all nodes can reach each other
	for _, from := range s.nodes {
		for _, to := range s.nodes {
			require.True(s.T(), s.controller.CanReach(from, to),
				"Node %s should be able to reach %s in healthy state", from, to)
		}
	}

	s.T().Log("Controller initialized correctly")
}

// TestSimplePartitionAndHeal tests a simple partition and heal cycle.
func (s *PartitionRecoveryTestSuite) TestSimplePartitionAndHeal() {
	s.T().Log("=== Test: Simple Partition and Heal ===")

	// Create a simple partition
	scenario := partition.CreateSimplePartition(s.nodes)
	s.T().Logf("Applying scenario: %s - %s", scenario.Name, scenario.Description)

	s.controller.ApplyPartition(scenario.Groups)

	// Verify partition state
	require.Equal(s.T(), partition.PartitionStatePartitioned, s.controller.State())

	// Verify nodes in different groups cannot reach each other
	group0 := scenario.Groups[0].Nodes
	group1 := scenario.Groups[1].Nodes

	for _, n0 := range group0 {
		for _, n1 := range group1 {
			require.False(s.T(), s.controller.CanReach(n0, n1),
				"Node %s should NOT reach %s across partition", n0, n1)
		}
	}

	// Verify nodes in the same group CAN reach each other
	for i, n0 := range group0 {
		for j, n1 := range group0 {
			if i != j {
				require.True(s.T(), s.controller.CanReach(n0, n1),
					"Node %s should reach %s within same group", n0, n1)
			}
		}
	}

	// Heal the partition
	s.controller.Heal()
	require.Equal(s.T(), partition.PartitionStateHealthy, s.controller.State())

	// Verify all nodes can reach each other again
	for _, from := range s.nodes {
		for _, to := range s.nodes {
			require.True(s.T(), s.controller.CanReach(from, to),
				"Node %s should reach %s after heal", from, to)
		}
	}

	s.T().Log("Simple partition and heal completed successfully")
}

// TestAsymmetricPartition tests asymmetric network partitions.
func (s *PartitionRecoveryTestSuite) TestAsymmetricPartition() {
	s.T().Log("=== Test: Asymmetric Partition ===")

	// Create an asymmetric partition where A->B is blocked but B->A is allowed
	nodeA := s.nodes[0]
	nodeB := s.nodes[1]

	s.controller.ApplyAsymmetricPartition(nodeA, nodeB, true)

	require.Equal(s.T(), partition.PartitionStatePartitioned, s.controller.State())

	// A cannot reach B
	require.False(s.T(), s.controller.CanReach(nodeA, nodeB),
		"Node A should NOT reach Node B")

	// B can still reach A
	require.True(s.T(), s.controller.CanReach(nodeB, nodeA),
		"Node B should still reach Node A")

	// Other connections should be unaffected
	require.True(s.T(), s.controller.CanReach(s.nodes[0], s.nodes[2]),
		"Other connections should be unaffected")

	s.controller.Heal()
	require.True(s.T(), s.controller.CanReach(nodeA, nodeB),
		"A->B should be restored after heal")

	s.T().Log("Asymmetric partition test completed successfully")
}

// TestMajorityMinorityPartition tests partition where majority can still reach consensus.
func (s *PartitionRecoveryTestSuite) TestMajorityMinorityPartition() {
	s.T().Log("=== Test: Majority-Minority Partition ===")

	scenario := partition.CreateMajorityMinorityPartition(s.nodes)
	s.T().Logf("Applying scenario: %s - %s", scenario.Name, scenario.Description)

	s.controller.ApplyPartition(scenario.Groups)

	// Find majority and minority groups
	var majorityGroup, minorityGroup partition.PartitionGroup
	for _, g := range scenario.Groups {
		if g.Name == "majority" {
			majorityGroup = g
		} else {
			minorityGroup = g
		}
	}

	// Majority group should have > 2/3 of nodes for BFT consensus
	majorityRatio := float64(len(majorityGroup.Nodes)) / float64(len(s.nodes))
	s.T().Logf("Majority ratio: %.2f (%d/%d nodes)", majorityRatio, len(majorityGroup.Nodes), len(s.nodes))

	// In a 4-node network with 2/3+1 = 3 majority, consensus can continue
	require.GreaterOrEqual(s.T(), majorityRatio, 0.66,
		"Majority group should have at least 2/3 of nodes")

	// Verify groups are isolated
	for _, mj := range majorityGroup.Nodes {
		for _, mn := range minorityGroup.Nodes {
			require.False(s.T(), s.controller.CanReach(mj, mn),
				"Majority node should not reach minority node")
		}
	}

	s.controller.Heal()
	s.T().Log("Majority-minority partition test completed successfully")
}

// TestThreeWayPartition tests a three-way network split.
func (s *PartitionRecoveryTestSuite) TestThreeWayPartition() {
	s.T().Log("=== Test: Three-Way Partition ===")

	scenario := partition.CreateThreeWayPartition(s.nodes)
	s.T().Logf("Applying scenario: %s - %s", scenario.Name, scenario.Description)

	s.controller.ApplyPartition(scenario.Groups)

	// Verify each group is isolated from the others
	for i, group1 := range scenario.Groups {
		for j, group2 := range scenario.Groups {
			if i != j {
				for _, n1 := range group1.Nodes {
					for _, n2 := range group2.Nodes {
						require.False(s.T(), s.controller.CanReach(n1, n2),
							"Node %s (group %s) should not reach %s (group %s)",
							n1, group1.Name, n2, group2.Name)
					}
				}
			}
		}
	}

	// No group should have quorum in a 3-way split of 4 nodes
	for _, group := range scenario.Groups {
		ratio := float64(len(group.Nodes)) / float64(len(s.nodes))
		require.Less(s.T(), ratio, 0.66,
			"No group should have quorum in three-way split")
	}

	s.controller.Heal()
	s.T().Log("Three-way partition test completed successfully")
}

// TestIsolatedNodePartition tests isolation of a single node.
func (s *PartitionRecoveryTestSuite) TestIsolatedNodePartition() {
	s.T().Log("=== Test: Isolated Node Partition ===")

	scenario := partition.CreateIsolatedNodePartition(s.nodes, 0)
	s.T().Logf("Applying scenario: %s - %s", scenario.Name, scenario.Description)

	s.controller.ApplyPartition(scenario.Groups)

	isolatedNode := s.nodes[0]

	// Isolated node cannot reach others
	for _, other := range s.nodes[1:] {
		require.False(s.T(), s.controller.CanReach(isolatedNode, other),
			"Isolated node should not reach %s", other)
	}

	// Other nodes can reach each other
	for i, n1 := range s.nodes[1:] {
		for j, n2 := range s.nodes[1:] {
			if i != j {
				require.True(s.T(), s.controller.CanReach(n1, n2),
					"Non-isolated nodes should reach each other")
			}
		}
	}

	s.controller.Heal()
	s.T().Log("Isolated node partition test completed successfully")
}

// =============================================================================
// Gradual Healing Tests
// =============================================================================

// TestGradualHealing tests gradual partition healing.
func (s *PartitionRecoveryTestSuite) TestGradualHealing() {
	s.T().Log("=== Test: Gradual Healing ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	require.Equal(s.T(), partition.PartitionStatePartitioned, s.controller.State())

	// Heal gradually over 1 second
	err := s.controller.HealGradually(s.ctx, 1*time.Second)
	require.NoError(s.T(), err)

	require.Equal(s.T(), partition.PartitionStateHealthy, s.controller.State())

	// Verify full connectivity restored
	for _, from := range s.nodes {
		for _, to := range s.nodes {
			require.True(s.T(), s.controller.CanReach(from, to),
				"Full connectivity should be restored")
		}
	}

	s.T().Log("Gradual healing test completed successfully")
}

// TestGradualHealingCancellation tests cancellation of gradual healing.
func (s *PartitionRecoveryTestSuite) TestGradualHealingCancellation() {
	s.T().Log("=== Test: Gradual Healing Cancellation ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	// Create a short-lived context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to heal over a longer period than context allows
	err := s.controller.HealGradually(ctx, 10*time.Second)

	// Should return context error
	require.Error(s.T(), err)
	require.ErrorIs(s.T(), err, context.DeadlineExceeded)

	s.T().Log("Gradual healing cancellation test completed successfully")
}

// =============================================================================
// Dynamic Node Tests
// =============================================================================

// TestAddNodeDuringPartition tests adding a node during a partition.
func (s *PartitionRecoveryTestSuite) TestAddNodeDuringPartition() {
	s.T().Log("=== Test: Add Node During Partition ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	// Add a new node
	newNode := partition.NodeID("validator-new")
	s.controller.AddNode(newNode)

	require.Equal(s.T(), len(s.nodes)+1, s.controller.NodeCount())

	// New node should be able to reach all nodes (not affected by partition)
	for _, node := range s.nodes {
		require.True(s.T(), s.controller.CanReach(newNode, node),
			"New node should reach existing nodes")
	}

	// Cleanup
	s.controller.RemoveNode(newNode)
	require.Equal(s.T(), len(s.nodes), s.controller.NodeCount())

	s.T().Log("Add node during partition test completed successfully")
}

// TestRemoveNodeDuringPartition tests removing a node during a partition.
func (s *PartitionRecoveryTestSuite) TestRemoveNodeDuringPartition() {
	s.T().Log("=== Test: Remove Node During Partition ===")

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	initialCount := s.controller.NodeCount()
	removedNode := s.nodes[0]

	s.controller.RemoveNode(removedNode)

	require.Equal(s.T(), initialCount-1, s.controller.NodeCount())

	// Removed node should not appear in connectivity matrix
	matrix := s.controller.GetConnectivityMatrix()
	_, exists := matrix[removedNode]
	require.False(s.T(), exists, "Removed node should not be in connectivity matrix")

	// Re-add node for cleanup
	s.controller.AddNode(removedNode)

	s.T().Log("Remove node during partition test completed successfully")
}

// =============================================================================
// Scenario Set Tests
// =============================================================================

// TestDefaultScenarioSet tests the default scenario set.
func (s *PartitionRecoveryTestSuite) TestDefaultScenarioSet() {
	s.T().Log("=== Test: Default Scenario Set ===")

	scenarioSet := partition.DefaultScenarioSet(s.nodes)

	require.NotEmpty(s.T(), scenarioSet.Scenarios, "Should have scenarios")

	for _, scenario := range scenarioSet.Scenarios {
		s.T().Logf("Testing scenario: %s - %s", scenario.Name, scenario.Description)

		s.controller.ApplyPartition(scenario.Groups)
		require.Equal(s.T(), partition.PartitionStatePartitioned, s.controller.State())

		s.controller.Heal()
		require.Equal(s.T(), partition.PartitionStateHealthy, s.controller.State())
	}

	s.T().Logf("Tested %d scenarios from default set", len(scenarioSet.Scenarios))
}

// TestFullScenarioSet tests the full scenario set.
func (s *PartitionRecoveryTestSuite) TestFullScenarioSet() {
	s.T().Log("=== Test: Full Scenario Set ===")

	scenarioSet := partition.FullScenarioSet(s.nodes)

	require.NotEmpty(s.T(), scenarioSet.Scenarios, "Should have scenarios")
	require.GreaterOrEqual(s.T(), len(scenarioSet.Scenarios), len(partition.DefaultScenarioSet(s.nodes).Scenarios),
		"Full set should have at least as many scenarios as default")

	for _, scenario := range scenarioSet.Scenarios {
		s.T().Logf("Testing scenario: %s", scenario.Name)

		s.controller.ApplyPartition(scenario.Groups)
		s.controller.Heal()
	}

	s.T().Logf("Tested %d scenarios from full set", len(scenarioSet.Scenarios))
}

// =============================================================================
// Callback Tests
// =============================================================================

// TestPartitionCallbacks tests partition and heal callbacks.
func (s *PartitionRecoveryTestSuite) TestPartitionCallbacks() {
	s.T().Log("=== Test: Partition Callbacks ===")

	partitionCalled := false
	healCalled := false

	s.controller.SetPartitionCallback(func(groups []partition.PartitionGroup) {
		partitionCalled = true
		require.NotEmpty(s.T(), groups, "Groups should not be empty in callback")
	})

	s.controller.SetHealCallback(func() {
		healCalled = true
	})

	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)
	require.True(s.T(), partitionCalled, "Partition callback should be called")

	s.controller.Heal()
	require.True(s.T(), healCalled, "Heal callback should be called")

	s.T().Log("Partition callbacks test completed successfully")
}
