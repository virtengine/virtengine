package partition

import (
	"fmt"
)

// Scenario represents a predefined network partition scenario.
type Scenario struct {
	Name        string
	Description string
	Groups      []PartitionGroup
}

// CreateSimplePartition creates a simple two-group partition where
// the network is split into two halves.
func CreateSimplePartition(nodes []NodeID) Scenario {
	mid := len(nodes) / 2
	if mid == 0 {
		mid = 1
	}

	return Scenario{
		Name:        "simple-partition",
		Description: "Network split into two equal halves",
		Groups: []PartitionGroup{
			{
				Name:  "group-a",
				Nodes: nodes[:mid],
			},
			{
				Name:  "group-b",
				Nodes: nodes[mid:],
			},
		},
	}
}

// CreateMajorityMinorityPartition creates a partition where one group
// has a majority of nodes (can still reach consensus) and another has
// a minority (cannot reach consensus alone).
func CreateMajorityMinorityPartition(nodes []NodeID) Scenario {
	// Need at least 4 nodes for meaningful majority/minority
	if len(nodes) < 4 {
		return CreateSimplePartition(nodes)
	}

	// Majority needs 2/3+1 for BFT consensus
	majoritySize := (2*len(nodes))/3 + 1
	if majoritySize > len(nodes) {
		majoritySize = len(nodes)
	}

	return Scenario{
		Name:        "majority-minority-partition",
		Description: fmt.Sprintf("Majority group (%d nodes) can reach consensus, minority (%d nodes) cannot", majoritySize, len(nodes)-majoritySize),
		Groups: []PartitionGroup{
			{
				Name:  "majority",
				Nodes: nodes[:majoritySize],
			},
			{
				Name:  "minority",
				Nodes: nodes[majoritySize:],
			},
		},
	}
}

// CreateThreeWayPartition creates a partition where the network is
// split into three groups, none of which has consensus quorum.
func CreateThreeWayPartition(nodes []NodeID) Scenario {
	if len(nodes) < 3 {
		return CreateSimplePartition(nodes)
	}

	third := len(nodes) / 3
	if third == 0 {
		third = 1
	}

	groups := []PartitionGroup{
		{
			Name:  "group-a",
			Nodes: nodes[:third],
		},
		{
			Name:  "group-b",
			Nodes: nodes[third : 2*third],
		},
		{
			Name:  "group-c",
			Nodes: nodes[2*third:],
		},
	}

	return Scenario{
		Name:        "three-way-partition",
		Description: "Network split into three groups, no group has quorum",
		Groups:      groups,
	}
}

// CreateIsolatedNodePartition creates a partition where a single node
// is isolated from the rest of the network.
func CreateIsolatedNodePartition(nodes []NodeID, isolatedIdx int) Scenario {
	if isolatedIdx < 0 || isolatedIdx >= len(nodes) {
		isolatedIdx = 0
	}

	var mainGroup []NodeID
	for i, node := range nodes {
		if i != isolatedIdx {
			mainGroup = append(mainGroup, node)
		}
	}

	return Scenario{
		Name:        "isolated-node",
		Description: fmt.Sprintf("Node %s is isolated from the network", nodes[isolatedIdx]),
		Groups: []PartitionGroup{
			{
				Name:  "main",
				Nodes: mainGroup,
			},
			{
				Name:  "isolated",
				Nodes: []NodeID{nodes[isolatedIdx]},
			},
		},
	}
}

// CreateByzantinePartition simulates a partition that could occur
// during a Byzantine fault scenario where validators are split.
func CreateByzantinePartition(nodes []NodeID, byzantineCount int) Scenario {
	if byzantineCount >= len(nodes) {
		byzantineCount = len(nodes) / 3 // Max 1/3 Byzantine
	}
	if byzantineCount < 1 {
		byzantineCount = 1
	}

	return Scenario{
		Name:        "byzantine-partition",
		Description: fmt.Sprintf("%d Byzantine nodes isolated from %d honest nodes", byzantineCount, len(nodes)-byzantineCount),
		Groups: []PartitionGroup{
			{
				Name:  "honest",
				Nodes: nodes[byzantineCount:],
			},
			{
				Name:  "byzantine",
				Nodes: nodes[:byzantineCount],
			},
		},
	}
}

// CreateChainPartition creates a chain-like partition where each node
// can only communicate with its immediate neighbors.
func CreateChainPartition(nodes []NodeID) Scenario {
	groups := make([]PartitionGroup, len(nodes))
	for i, node := range nodes {
		neighbors := []NodeID{node}
		if i > 0 {
			neighbors = append(neighbors, nodes[i-1])
		}
		if i < len(nodes)-1 {
			neighbors = append(neighbors, nodes[i+1])
		}
		groups[i] = PartitionGroup{
			Name:  fmt.Sprintf("chain-%d", i),
			Nodes: neighbors,
		}
	}

	return Scenario{
		Name:        "chain-partition",
		Description: "Nodes can only communicate with immediate neighbors",
		Groups:      groups,
	}
}

// CreateRingPartition creates a ring topology where each node can
// communicate with two neighbors but the ring is broken at one point.
func CreateRingPartition(nodes []NodeID, breakPoint int) Scenario {
	if len(nodes) < 3 {
		return CreateSimplePartition(nodes)
	}

	if breakPoint < 0 || breakPoint >= len(nodes) {
		breakPoint = len(nodes) - 1
	}

	// Split into two arcs
	arc1 := make([]NodeID, 0, len(nodes)/2+1)
	arc2 := make([]NodeID, 0, len(nodes)/2+1)

	for i := 0; i <= breakPoint; i++ {
		arc1 = append(arc1, nodes[i])
	}
	for i := breakPoint + 1; i < len(nodes); i++ {
		arc2 = append(arc2, nodes[i])
	}

	return Scenario{
		Name:        "ring-partition",
		Description: fmt.Sprintf("Ring topology broken at position %d", breakPoint),
		Groups: []PartitionGroup{
			{
				Name:  "arc-1",
				Nodes: arc1,
			},
			{
				Name:  "arc-2",
				Nodes: arc2,
			},
		},
	}
}

// CreateGeoPartition simulates a geographic partition where nodes
// in different "regions" cannot communicate.
func CreateGeoPartition(nodes []NodeID, regionsCount int) Scenario {
	if regionsCount < 2 {
		regionsCount = 2
	}
	if regionsCount > len(nodes) {
		regionsCount = len(nodes)
	}

	nodesPerRegion := len(nodes) / regionsCount
	if nodesPerRegion < 1 {
		nodesPerRegion = 1
	}

	groups := make([]PartitionGroup, 0, regionsCount)
	for i := 0; i < regionsCount; i++ {
		start := i * nodesPerRegion
		end := start + nodesPerRegion
		if i == regionsCount-1 {
			end = len(nodes) // Last region gets remaining nodes
		}
		if end > len(nodes) {
			end = len(nodes)
		}
		if start >= end {
			continue
		}

		groups = append(groups, PartitionGroup{
			Name:  fmt.Sprintf("region-%d", i+1),
			Nodes: nodes[start:end],
		})
	}

	return Scenario{
		Name:        "geo-partition",
		Description: fmt.Sprintf("Network split into %d geographic regions", len(groups)),
		Groups:      groups,
	}
}

// ScenarioSet represents a collection of scenarios for comprehensive testing.
type ScenarioSet struct {
	Scenarios []Scenario
}

// DefaultScenarioSet returns a set of common partition scenarios.
func DefaultScenarioSet(nodes []NodeID) ScenarioSet {
	scenarios := []Scenario{
		CreateSimplePartition(nodes),
		CreateMajorityMinorityPartition(nodes),
	}

	if len(nodes) >= 3 {
		scenarios = append(scenarios, CreateThreeWayPartition(nodes))
		scenarios = append(scenarios, CreateIsolatedNodePartition(nodes, 0))
	}

	if len(nodes) >= 4 {
		scenarios = append(scenarios, CreateByzantinePartition(nodes, 1))
		scenarios = append(scenarios, CreateGeoPartition(nodes, 2))
	}

	return ScenarioSet{Scenarios: scenarios}
}

// FullScenarioSet returns all available partition scenarios.
func FullScenarioSet(nodes []NodeID) ScenarioSet {
	scenarios := DefaultScenarioSet(nodes).Scenarios

	if len(nodes) >= 4 {
		scenarios = append(scenarios, CreateChainPartition(nodes))
		scenarios = append(scenarios, CreateRingPartition(nodes, len(nodes)/2))
		scenarios = append(scenarios, CreateGeoPartition(nodes, 3))
	}

	return ScenarioSet{Scenarios: scenarios}
}
