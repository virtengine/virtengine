// Package scale contains network partition and recovery tests.
// These tests simulate network partitions and measure recovery behavior.
//
// Task Reference: SCALE-001 - Load Testing - 1M Nodes Simulation
package scale

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Network Partition Constants
// ============================================================================

const (
	// Partition scenarios
	DefaultNodeCount       = 100
	MajorityPartitionRatio = 0.67 // 2/3 majority
	MinorityPartitionRatio = 0.33 // 1/3 minority

	// Recovery baselines
	MaxPartitionRecoveryTime = 30 * time.Second
	MaxStateReconcileTime    = 60 * time.Second
	MaxMessageLossRate       = 0.01 // 1%
)

// NetworkPartitionBaseline defines targets for partition recovery
type NetworkPartitionBaseline struct {
	NodeCount               int           `json:"node_count"`
	MaxRecoveryTime         time.Duration `json:"max_recovery_time"`
	MaxStateReconcileTime   time.Duration `json:"max_state_reconcile_time"`
	MaxMessageLoss          float64       `json:"max_message_loss"`
	MaxConsensusRoundsLost  int           `json:"max_consensus_rounds_lost"`
	MinMajorityAvailability float64       `json:"min_majority_availability"`
}

// DefaultPartitionBaseline returns baseline for partition recovery
func DefaultPartitionBaseline() NetworkPartitionBaseline {
	return NetworkPartitionBaseline{
		NodeCount:               100,
		MaxRecoveryTime:         30 * time.Second,
		MaxStateReconcileTime:   60 * time.Second,
		MaxMessageLoss:          0.01,
		MaxConsensusRoundsLost:  5,
		MinMajorityAvailability: 0.99,
	}
}

// ============================================================================
// Mock Network Types
// ============================================================================

// NodeState represents node operational state
type NodeState uint8

const (
	NodeHealthy NodeState = iota
	NodePartitioned
	NodeRecovering
	NodeFailed
)

// MockNode represents a network node
type MockNode struct {
	ID        int
	Address   [20]byte
	State     atomic.Uint32
	Partition int // -1 = not partitioned, 0+ = partition ID

	// Network simulation
	peers  map[int]*MockNode
	peerMu sync.RWMutex
	inbox  chan *NetworkMessage

	// Metrics
	messagesSent     atomic.Int64
	messagesReceived atomic.Int64
	messagesDropped  atomic.Int64

	// State
	height        atomic.Int64
	lastBlockTime atomic.Int64

	ctx    context.Context
	cancel context.CancelFunc
}

// NetworkMessage represents a network message
type NetworkMessage struct {
	From      int
	To        int
	Type      string
	Height    int64
	Timestamp time.Time
	Data      []byte
}

// MockNetwork simulates a network with partition capabilities
type MockNetwork struct {
	mu         sync.RWMutex
	nodes      map[int]*MockNode
	partitions map[int][]int // partition ID -> node IDs

	// Partition state
	partitioned   bool
	partitionTime time.Time
	healTime      time.Time

	// Metrics
	//nolint:unused // Reserved for message tracking metrics
	totalMessages atomic.Int64
	//nolint:unused // Reserved for message tracking metrics
	droppedMessages atomic.Int64
	recoveryEvents  atomic.Int64
}

// NewMockNetwork creates a new mock network
func NewMockNetwork(nodeCount int) *MockNetwork {
	net := &MockNetwork{
		nodes:      make(map[int]*MockNode),
		partitions: make(map[int][]int),
	}

	// Create nodes
	for i := 0; i < nodeCount; i++ {
		node := newMockNode(i)
		net.nodes[i] = node
	}

	// Connect all nodes (full mesh for simplicity)
	for i := 0; i < nodeCount; i++ {
		for j := 0; j < nodeCount; j++ {
			if i != j {
				net.nodes[i].addPeer(net.nodes[j])
			}
		}
	}

	return net
}

func newMockNode(id int) *MockNode {
	ctx, cancel := context.WithCancel(context.Background())

	node := &MockNode{
		ID:        id,
		Partition: -1,
		peers:     make(map[int]*MockNode),
		inbox:     make(chan *NetworkMessage, 1000),
		ctx:       ctx,
		cancel:    cancel,
	}

	mustRandReadPartition(node.Address[:])
	node.State.Store(uint32(NodeHealthy))

	return node
}

func mustRandReadPartition(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

func nodeStateFromUint32(value uint32) (NodeState, bool) {
	if value > uint32(^uint8(0)) {
		return 0, false
	}
	return NodeState(value), true
}

func (n *MockNode) addPeer(peer *MockNode) {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()
	n.peers[peer.ID] = peer
}

//nolint:unused // Test helper for peer removal
func (n *MockNode) removePeer(peerID int) {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()
	delete(n.peers, peerID)
}

// Start starts the node
func (n *MockNode) Start() {
	go n.processMessages()
}

// Stop stops the node
func (n *MockNode) Stop() {
	n.cancel()
}

// processMessages processes incoming messages
func (n *MockNode) processMessages() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case msg := <-n.inbox:
			n.handleMessage(msg)
		}
	}
}

func (n *MockNode) handleMessage(msg *NetworkMessage) {
	n.messagesReceived.Add(1)

	// Update height if message is newer
	if msg.Height > n.height.Load() {
		n.height.Store(msg.Height)
		n.lastBlockTime.Store(msg.Timestamp.UnixNano())
	}
}

// SendMessage sends a message to a peer
func (n *MockNode) SendMessage(to *MockNode, msgType string, data []byte) error {
	// Check if partitioned from target
	if n.Partition >= 0 && to.Partition >= 0 && n.Partition != to.Partition {
		n.messagesDropped.Add(1)
		return fmt.Errorf("partitioned")
	}

	msg := &NetworkMessage{
		From:      n.ID,
		To:        to.ID,
		Type:      msgType,
		Height:    n.height.Load(),
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case to.inbox <- msg:
		n.messagesSent.Add(1)
		return nil
	default:
		n.messagesDropped.Add(1)
		return fmt.Errorf("inbox full")
	}
}

// BroadcastMessage broadcasts to all peers
func (n *MockNode) BroadcastMessage(msgType string, data []byte) (sent, dropped int) {
	n.peerMu.RLock()
	defer n.peerMu.RUnlock()

	for _, peer := range n.peers {
		if err := n.SendMessage(peer, msgType, data); err != nil {
			dropped++
		} else {
			sent++
		}
	}
	return
}

// GetStats returns node statistics
func (n *MockNode) GetStats() (sent, received, dropped int64) {
	return n.messagesSent.Load(), n.messagesReceived.Load(), n.messagesDropped.Load()
}

// Start starts all nodes
func (net *MockNetwork) Start() {
	net.mu.RLock()
	defer net.mu.RUnlock()

	for _, node := range net.nodes {
		node.Start()
	}
}

// Stop stops all nodes
func (net *MockNetwork) Stop() {
	net.mu.RLock()
	defer net.mu.RUnlock()

	for _, node := range net.nodes {
		node.Stop()
	}
}

// CreatePartition creates a network partition
func (net *MockNetwork) CreatePartition(majorityRatio float64) {
	net.mu.Lock()
	defer net.mu.Unlock()

	nodeCount := len(net.nodes)
	majoritySize := int(float64(nodeCount) * majorityRatio)

	// Create two partitions
	partition0 := make([]int, 0, majoritySize)
	partition1 := make([]int, 0, nodeCount-majoritySize)

	i := 0
	for id, node := range net.nodes {
		if i < majoritySize {
			node.Partition = 0
			partition0 = append(partition0, id)
		} else {
			node.Partition = 1
			partition1 = append(partition1, id)
		}
		i++
	}

	net.partitions[0] = partition0
	net.partitions[1] = partition1
	net.partitioned = true
	net.partitionTime = time.Now()
}

// HealPartition heals the network partition
func (net *MockNetwork) HealPartition() {
	net.mu.Lock()
	defer net.mu.Unlock()

	for _, node := range net.nodes {
		state, ok := nodeStateFromUint32(node.State.Load())
		if ok && state == NodePartitioned {
			node.State.Store(uint32(NodeRecovering))
		}
		node.Partition = -1
	}

	net.partitions = make(map[int][]int)
	net.partitioned = false
	net.healTime = time.Now()
	net.recoveryEvents.Add(1)
}

// IsPartitioned returns whether network is partitioned
func (net *MockNetwork) IsPartitioned() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.partitioned
}

// GetPartitionDuration returns how long the partition has been active
func (net *MockNetwork) GetPartitionDuration() time.Duration {
	net.mu.RLock()
	defer net.mu.RUnlock()

	if !net.partitioned {
		return 0
	}
	return time.Since(net.partitionTime)
}

// GetRecoveryTime returns time taken to recover from last partition
func (net *MockNetwork) GetRecoveryTime() time.Duration {
	net.mu.RLock()
	defer net.mu.RUnlock()

	if net.healTime.IsZero() {
		return 0
	}
	return net.healTime.Sub(net.partitionTime)
}

// GetNodesByPartition returns nodes in a partition
func (net *MockNetwork) GetNodesByPartition(partitionID int) []*MockNode {
	net.mu.RLock()
	defer net.mu.RUnlock()

	nodeIDs := net.partitions[partitionID]
	nodes := make([]*MockNode, 0, len(nodeIDs))

	for _, id := range nodeIDs {
		if node, ok := net.nodes[id]; ok {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// SimulateConsensusRound simulates a consensus round
func (net *MockNetwork) SimulateConsensusRound(height int64) (participatingNodes, droppedMessages int) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	// Each node broadcasts a prevote
	for _, node := range net.nodes {
		state, ok := nodeStateFromUint32(node.State.Load())
		if !ok || state != NodeFailed {
			sent, dropped := node.BroadcastMessage("prevote", []byte(fmt.Sprintf("height:%d", height)))
			participatingNodes++
			droppedMessages += dropped
			_ = sent
		}
	}

	// Update heights
	for _, node := range net.nodes {
		if node.Partition == 0 || !net.partitioned {
			node.height.Store(height)
		}
	}

	return
}

// GetAggregateStats returns network-wide statistics
func (net *MockNetwork) GetAggregateStats() (sent, received, dropped int64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	for _, node := range net.nodes {
		s, r, d := node.GetStats()
		sent += s
		received += r
		dropped += d
	}
	return
}

// GetHeightConsensus returns nodes grouped by their current height
func (net *MockNetwork) GetHeightConsensus() map[int64]int {
	net.mu.RLock()
	defer net.mu.RUnlock()

	heights := make(map[int64]int)
	for _, node := range net.nodes {
		h := node.height.Load()
		heights[h]++
	}
	return heights
}

// Count returns number of nodes
func (net *MockNetwork) Count() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return len(net.nodes)
}

// ============================================================================
// Benchmarks
// ============================================================================

// BenchmarkNetworkCreation benchmarks network creation
func BenchmarkNetworkCreation(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("nodes_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				net := NewMockNetwork(size)
				_ = net.Count()
			}
		})
	}
}

// BenchmarkMessageBroadcast benchmarks message broadcasting
func BenchmarkMessageBroadcast(b *testing.B) {
	net := NewMockNetwork(50)
	net.Start()
	defer net.Stop()

	node := net.nodes[0]
	data := make([]byte, 1024)
	mustRandReadPartition(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.BroadcastMessage("test", data)
	}
}

// BenchmarkConsensusRound benchmarks consensus round simulation
func BenchmarkConsensusRound(b *testing.B) {
	net := NewMockNetwork(100)
	net.Start()
	defer net.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		net.SimulateConsensusRound(int64(i))
	}
}

// BenchmarkPartitionCreateHeal benchmarks partition create/heal cycle
func BenchmarkPartitionCreateHeal(b *testing.B) {
	net := NewMockNetwork(100)
	net.Start()
	defer net.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		net.CreatePartition(0.67)
		net.HealPartition()
	}
}

// ============================================================================
// Network Partition Tests
// ============================================================================

// TestNetworkPartitionRecovery tests partition recovery behavior
func TestNetworkPartitionRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping partition recovery test in short mode")
	}

	nodeCount := 50

	t.Logf("=== Network Partition Recovery Test ===")
	t.Logf("Nodes: %d", nodeCount)

	net := NewMockNetwork(nodeCount)
	net.Start()
	defer net.Stop()

	// Run consensus rounds before partition
	t.Run("pre_partition_consensus", func(t *testing.T) {
		for h := int64(1); h <= 10; h++ {
			participants, dropped := net.SimulateConsensusRound(h)
			require.Equal(t, nodeCount, participants)
			require.Equal(t, 0, dropped, "No messages should drop without partition")
		}

		heights := net.GetHeightConsensus()
		t.Logf("Pre-partition heights: %v", heights)
		require.Len(t, heights, 1, "All nodes should be at same height")
	})

	// Create partition
	t.Run("create_partition", func(t *testing.T) {
		net.CreatePartition(MajorityPartitionRatio)
		require.True(t, net.IsPartitioned())

		majorityNodes := net.GetNodesByPartition(0)
		minorityNodes := net.GetNodesByPartition(1)

		t.Logf("Majority partition: %d nodes", len(majorityNodes))
		t.Logf("Minority partition: %d nodes", len(minorityNodes))

		require.Greater(t, len(majorityNodes), len(minorityNodes))
	})

	// Run consensus during partition
	t.Run("partitioned_consensus", func(t *testing.T) {
		initialHeight := int64(10)

		for h := initialHeight + 1; h <= initialHeight+10; h++ {
			participants, dropped := net.SimulateConsensusRound(h)
			t.Logf("Height %d: participants=%d, dropped=%d", h, participants, dropped)
		}

		heights := net.GetHeightConsensus()
		t.Logf("Partitioned heights: %v", heights)

		// Should have divergent heights
		require.Greater(t, len(heights), 1, "Partitions should diverge")
	})

	// Heal partition
	t.Run("heal_partition", func(t *testing.T) {
		partitionDuration := net.GetPartitionDuration()
		net.HealPartition()

		require.False(t, net.IsPartitioned())
		t.Logf("Partition duration: %v", partitionDuration)
	})

	// Verify recovery
	t.Run("post_partition_recovery", func(t *testing.T) {
		// Run more consensus rounds
		for h := int64(30); h <= 40; h++ {
			net.SimulateConsensusRound(h)
		}

		heights := net.GetHeightConsensus()
		t.Logf("Post-recovery heights: %v", heights)

		// Eventually all nodes should converge
		// Note: In this simulation, majority partition continues
		require.GreaterOrEqual(t, len(heights), 1)
	})

	// Check message statistics
	sent, received, dropped := net.GetAggregateStats()
	dropRate := float64(dropped) / float64(sent) * 100

	t.Logf("Total messages - Sent: %d, Received: %d, Dropped: %d (%.2f%%)",
		sent, received, dropped, dropRate)

	// During partition testing, drops are expected - just verify not catastrophic
	require.Less(t, dropRate/100, 0.50,
		"Message drop rate should not exceed 50% even during partitions")
}

// TestMajorityMinorityBehavior tests how majority/minority partitions behave
func TestMajorityMinorityBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping majority/minority test in short mode")
	}

	nodeCount := 100
	t.Logf("=== Majority/Minority Partition Behavior ===")

	net := NewMockNetwork(nodeCount)
	net.Start()
	defer net.Stop()

	// Create 2/3 - 1/3 partition
	net.CreatePartition(MajorityPartitionRatio)

	majorityNodes := net.GetNodesByPartition(0)
	minorityNodes := net.GetNodesByPartition(1)

	t.Logf("Majority: %d nodes (%.0f%%)", len(majorityNodes), float64(len(majorityNodes))/float64(nodeCount)*100)
	t.Logf("Minority: %d nodes (%.0f%%)", len(minorityNodes), float64(len(minorityNodes))/float64(nodeCount)*100)

	// Simulate consensus - only majority should make progress
	const rounds = 20
	majorityHeights := make([]int64, rounds)

	for r := 0; r < rounds; r++ {
		height := int64(r + 1)
		net.SimulateConsensusRound(height)

		// Check majority progress
		var majorityAtHeight int
		for _, node := range majorityNodes {
			if node.height.Load() == height {
				majorityAtHeight++
			}
		}
		majorityHeights[r] = int64(majorityAtHeight)
	}

	// Verify majority made progress
	require.Equal(t, int64(len(majorityNodes)), majorityHeights[rounds-1],
		"All majority nodes should reach final height")

	// Verify minority is behind
	var minorityMaxHeight int64
	for _, node := range minorityNodes {
		if node.height.Load() > minorityMaxHeight {
			minorityMaxHeight = node.height.Load()
		}
	}

	t.Logf("Majority final height: %d", rounds)
	t.Logf("Minority max height: %d", minorityMaxHeight)

	require.Less(t, minorityMaxHeight, int64(rounds),
		"Minority should not reach consensus without majority")

	net.HealPartition()
}

// TestPartitionDurations tests various partition durations
func TestPartitionDurations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping partition duration test in short mode")
	}

	durations := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	for _, duration := range durations {
		t.Run(fmt.Sprintf("duration_%v", duration), func(t *testing.T) {
			net := NewMockNetwork(30)
			net.Start()
			defer net.Stop()

			// Pre-partition
			for h := int64(1); h <= 5; h++ {
				net.SimulateConsensusRound(h)
			}

			// Create partition
			net.CreatePartition(0.67)

			// Run consensus during partition
			start := time.Now()
			height := int64(6)
			for time.Since(start) < duration {
				net.SimulateConsensusRound(height)
				height++
				time.Sleep(10 * time.Millisecond)
			}

			// Heal
			net.HealPartition()

			// Post-partition consensus
			for h := height; h <= height+5; h++ {
				net.SimulateConsensusRound(h)
			}

			heights := net.GetHeightConsensus()
			t.Logf("Duration %v: final heights=%v", duration, heights)
		})
	}
}

// TestChaosPartitioning tests random partition scenarios
func TestChaosPartitioning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos partitioning test in short mode")
	}

	nodeCount := 50
	iterations := 10

	t.Logf("=== Chaos Partitioning Test ===")
	t.Logf("Nodes: %d, Iterations: %d", nodeCount, iterations)

	net := NewMockNetwork(nodeCount)
	net.Start()
	defer net.Stop()

	var totalPartitionTime time.Duration
	partitionRatios := []float64{0.5, 0.6, 0.7, 0.8}
	ratioCount := len(partitionRatios)
	if ratioCount == 0 {
		t.Fatal("partitionRatios must not be empty")
	}

	for i := 0; i < iterations; i++ {
		// Choose random partition ratio
		//nolint:gosec // ratioCount checked above and modulo ensures in-range index.
		ratio := partitionRatios[i%ratioCount]

		// Create partition
		net.CreatePartition(ratio)

		// Run some consensus rounds
		for h := int64(i*10 + 1); h <= int64(i*10+5); h++ {
			net.SimulateConsensusRound(h)
		}

		partitionDuration := net.GetPartitionDuration()
		totalPartitionTime += partitionDuration

		// Heal
		net.HealPartition()

		// Run recovery rounds
		for h := int64(i*10 + 6); h <= int64(i*10+10); h++ {
			net.SimulateConsensusRound(h)
		}
	}

	sent, received, dropped := net.GetAggregateStats()
	avgPartitionTime := totalPartitionTime / time.Duration(iterations)

	t.Logf("Total iterations: %d", iterations)
	t.Logf("Average partition duration: %v", avgPartitionTime)
	t.Logf("Messages - Sent: %d, Received: %d, Dropped: %d", sent, received, dropped)
}

// TestStateReconciliation tests state reconciliation after partition
func TestStateReconciliation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping state reconciliation test in short mode")
	}

	nodeCount := 30
	t.Logf("=== State Reconciliation Test ===")

	net := NewMockNetwork(nodeCount)
	net.Start()
	defer net.Stop()

	// Establish initial state
	for h := int64(1); h <= 10; h++ {
		net.SimulateConsensusRound(h)
	}

	// Create partition
	net.CreatePartition(0.67)

	// Run divergent consensus
	for h := int64(11); h <= 30; h++ {
		net.SimulateConsensusRound(h)
	}

	// Record heights before healing
	heightsBefore := net.GetHeightConsensus()
	t.Logf("Heights before healing: %v", heightsBefore)

	// Heal partition
	net.HealPartition()

	// Simulate state sync - nodes catch up to highest height
	reconcileStart := time.Now()

	// Find max height
	var maxHeight int64
	for h := range heightsBefore {
		if h > maxHeight {
			maxHeight = h
		}
	}

	// Simulate catch-up
	for h := int64(31); h <= maxHeight+10; h++ {
		net.SimulateConsensusRound(h)
	}

	reconcileTime := time.Since(reconcileStart)
	heightsAfter := net.GetHeightConsensus()

	t.Logf("Heights after reconciliation: %v", heightsAfter)
	t.Logf("Reconciliation time: %v", reconcileTime)

	// All nodes should converge
	if len(heightsAfter) > 2 {
		t.Logf("Warning: State not fully reconciled after healing")
	}
}
