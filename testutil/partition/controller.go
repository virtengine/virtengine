// Package partition provides network partition simulation for testing.
// It allows controlled simulation of network failures, partitions, and healing
// to validate consensus recovery and state synchronization.
package partition

import (
	"context"
	"sync"
	"time"
)

// NodeID represents a unique identifier for a network node.
type NodeID string

// PartitionGroup represents a group of nodes that can communicate with each other.
type PartitionGroup struct {
	Name  string
	Nodes []NodeID
}

// PartitionState represents the current state of a network partition.
type PartitionState int

const (
	// PartitionStateHealthy indicates no partition is active.
	PartitionStateHealthy PartitionState = iota
	// PartitionStatePartitioned indicates the network is partitioned.
	PartitionStatePartitioned
	// PartitionStateHealing indicates the partition is being healed.
	PartitionStateHealing
)

// String returns a human-readable string for the partition state.
func (s PartitionState) String() string {
	switch s {
	case PartitionStateHealthy:
		return "healthy"
	case PartitionStatePartitioned:
		return "partitioned"
	case PartitionStateHealing:
		return "healing"
	default:
		return "unknown"
	}
}

// Controller manages network partition simulation for testing.
// It controls which nodes can communicate with each other and
// tracks partition events for metrics collection.
type Controller struct {
	mu sync.RWMutex

	// nodes is the set of all nodes in the network.
	nodes map[NodeID]bool

	// connectivity maps each node to the set of nodes it can reach.
	// If a node is not in this map, it can reach all nodes.
	connectivity map[NodeID]map[NodeID]bool

	// state is the current partition state.
	state PartitionState

	// metrics tracks partition-related metrics.
	metrics *Metrics

	// onPartition is called when a partition is applied.
	onPartition func(groups []PartitionGroup)

	// onHeal is called when a partition is healed.
	onHeal func()

	// filter is the message filter for simulating partitions.
	filter *MessageFilter
}

// NewController creates a new partition controller for the given nodes.
func NewController(nodeIDs ...NodeID) *Controller {
	nodes := make(map[NodeID]bool, len(nodeIDs))
	for _, id := range nodeIDs {
		nodes[id] = true
	}

	c := &Controller{
		nodes:        nodes,
		connectivity: make(map[NodeID]map[NodeID]bool),
		state:        PartitionStateHealthy,
		metrics:      NewMetrics(),
		filter:       NewMessageFilter(),
	}

	// Initially, all nodes can reach all nodes
	c.resetConnectivity()

	return c
}

// resetConnectivity sets all nodes to be able to reach all other nodes.
func (c *Controller) resetConnectivity() {
	c.connectivity = make(map[NodeID]map[NodeID]bool)
	for from := range c.nodes {
		c.connectivity[from] = make(map[NodeID]bool)
		for to := range c.nodes {
			c.connectivity[from][to] = true
		}
	}
}

// State returns the current partition state.
func (c *Controller) State() PartitionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// CanReach returns true if the 'from' node can reach the 'to' node.
func (c *Controller) CanReach(from, to NodeID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.state == PartitionStateHealthy {
		return true
	}

	if reachable, ok := c.connectivity[from]; ok {
		return reachable[to]
	}
	return true
}

// ApplyPartition applies a network partition based on the given groups.
// Nodes within the same group can communicate, but nodes in different
// groups cannot.
func (c *Controller) ApplyPartition(groups []PartitionGroup) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset connectivity
	c.resetConnectivity()

	// Build a map of node to group
	nodeToGroup := make(map[NodeID]string)
	for _, group := range groups {
		for _, node := range group.Nodes {
			nodeToGroup[node] = group.Name
		}
	}

	// Set connectivity based on groups
	for from := range c.nodes {
		for to := range c.nodes {
			fromGroup, fromHasGroup := nodeToGroup[from]
			toGroup, toHasGroup := nodeToGroup[to]

			// If both nodes are in groups and they're different, block
			if fromHasGroup && toHasGroup && fromGroup != toGroup {
				c.connectivity[from][to] = false
			}
		}
	}

	c.state = PartitionStatePartitioned
	c.metrics.RecordPartitionStart(time.Now())

	// Update filter rules
	for from := range c.nodes {
		for to := range c.nodes {
			c.filter.SetBlocked(from, to, !c.connectivity[from][to])
		}
	}

	if c.onPartition != nil {
		c.onPartition(groups)
	}
}

// ApplyAsymmetricPartition applies an asymmetric partition where
// communication is blocked in one direction only.
func (c *Controller) ApplyAsymmetricPartition(from, to NodeID, block bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == PartitionStateHealthy {
		c.metrics.RecordPartitionStart(time.Now())
	}

	c.connectivity[from][to] = !block
	c.state = PartitionStatePartitioned

	c.filter.SetBlocked(from, to, block)
}

// Heal restores full network connectivity.
func (c *Controller) Heal() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state = PartitionStateHealing
	c.resetConnectivity()
	c.filter.ClearAll()

	c.metrics.RecordPartitionEnd(time.Now())
	c.state = PartitionStateHealthy

	if c.onHeal != nil {
		c.onHeal()
	}
}

// HealGradually heals the partition gradually over the specified duration.
// This simulates a more realistic network recovery scenario.
func (c *Controller) HealGradually(ctx context.Context, duration time.Duration) error {
	c.mu.Lock()
	c.state = PartitionStateHealing
	c.mu.Unlock()

	// Get all blocked connections
	blocked := c.filter.GetBlocked()
	if len(blocked) == 0 {
		c.Heal()
		return nil
	}

	// Calculate interval between unblocking connections
	interval := duration / time.Duration(len(blocked))
	if interval < time.Millisecond {
		interval = time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	i := 0
	for pair := range blocked {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			c.mu.Lock()
			c.connectivity[pair.From][pair.To] = true
			c.filter.SetBlocked(pair.From, pair.To, false)
			c.mu.Unlock()
			i++
		}
	}

	c.mu.Lock()
	c.state = PartitionStateHealthy
	c.metrics.RecordPartitionEnd(time.Now())
	c.mu.Unlock()

	if c.onHeal != nil {
		c.onHeal()
	}

	return nil
}

// SetPartitionCallback sets a callback to be invoked when a partition is applied.
func (c *Controller) SetPartitionCallback(fn func(groups []PartitionGroup)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onPartition = fn
}

// SetHealCallback sets a callback to be invoked when a partition is healed.
func (c *Controller) SetHealCallback(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onHeal = fn
}

// Metrics returns the partition metrics collector.
func (c *Controller) Metrics() *Metrics {
	return c.metrics
}

// Filter returns the message filter for the partition controller.
func (c *Controller) Filter() *MessageFilter {
	return c.filter
}

// AddNode adds a node to the controller.
func (c *Controller) AddNode(id NodeID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.nodes[id] {
		return
	}

	c.nodes[id] = true
	c.connectivity[id] = make(map[NodeID]bool)

	// New node can reach all existing nodes
	for other := range c.nodes {
		c.connectivity[id][other] = true
		c.connectivity[other][id] = true
	}
}

// RemoveNode removes a node from the controller.
func (c *Controller) RemoveNode(id NodeID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.nodes, id)
	delete(c.connectivity, id)

	for _, reachable := range c.connectivity {
		delete(reachable, id)
	}
}

// NodeCount returns the number of nodes in the controller.
func (c *Controller) NodeCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.nodes)
}

// Nodes returns a slice of all node IDs.
func (c *Controller) Nodes() []NodeID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := make([]NodeID, 0, len(c.nodes))
	for id := range c.nodes {
		nodes = append(nodes, id)
	}
	return nodes
}

// GetConnectivityMatrix returns the current connectivity matrix.
// The outer map key is the source node, inner map key is the destination,
// and the bool value indicates whether communication is allowed.
func (c *Controller) GetConnectivityMatrix() map[NodeID]map[NodeID]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[NodeID]map[NodeID]bool)
	for from, reachable := range c.connectivity {
		result[from] = make(map[NodeID]bool)
		for to, canReach := range reachable {
			result[from][to] = canReach
		}
	}
	return result
}
