package partition

import (
	"sort"
	"sync"
	"time"
)

// Metrics collects and reports network partition metrics.
type Metrics struct {
	mu sync.RWMutex

	// partitionEvents tracks all partition events.
	partitionEvents []PartitionEvent

	// healingMetrics tracks healing performance.
	healingMetrics []HealingMetric

	// blockMetrics tracks block production metrics.
	blockMetrics []BlockMetric

	// stateMetrics tracks state synchronization metrics.
	stateMetrics []StateMetric

	// currentPartition tracks the current partition start time.
	currentPartitionStart *time.Time
}

// PartitionEvent represents a partition event.
type PartitionEvent struct {
	Type      PartitionEventType
	Timestamp time.Time
	Duration  time.Duration
	Groups    int
	Nodes     int
}

// PartitionEventType indicates the type of partition event.
type PartitionEventType int

const (
	PartitionEventStart PartitionEventType = iota
	PartitionEventEnd
	PartitionEventHealStart
	PartitionEventHealComplete
)

func (t PartitionEventType) String() string {
	switch t {
	case PartitionEventStart:
		return "partition_start"
	case PartitionEventEnd:
		return "partition_end"
	case PartitionEventHealStart:
		return "heal_start"
	case PartitionEventHealComplete:
		return "heal_complete"
	default:
		return "unknown"
	}
}

// HealingMetric captures metrics about partition healing.
type HealingMetric struct {
	// PartitionDuration is how long the partition lasted.
	PartitionDuration time.Duration

	// HealingDuration is how long healing took.
	HealingDuration time.Duration

	// TimeToFirstBlock is time from heal to first new block.
	TimeToFirstBlock time.Duration

	// TimeToConsensus is time from heal to full consensus.
	TimeToConsensus time.Duration

	// MessagesReplayed is the number of replayed messages detected.
	MessagesReplayed int

	// StateSyncDuration is time spent synchronizing state.
	StateSyncDuration time.Duration
}

// BlockMetric captures block production metrics during partitions.
type BlockMetric struct {
	Timestamp       time.Time
	Height          int64
	Producer        NodeID
	DuringPartition bool
	TimeSinceHeal   time.Duration
}

// StateMetric captures state synchronization metrics.
type StateMetric struct {
	NodeID        NodeID
	Timestamp     time.Time
	HeightBefore  int64
	HeightAfter   int64
	SyncDuration  time.Duration
	StateHashMatch bool
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		partitionEvents: make([]PartitionEvent, 0),
		healingMetrics:  make([]HealingMetric, 0),
		blockMetrics:    make([]BlockMetric, 0),
		stateMetrics:    make([]StateMetric, 0),
	}
}

// RecordPartitionStart records the start of a partition.
func (m *Metrics) RecordPartitionStart(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentPartitionStart = &t
	m.partitionEvents = append(m.partitionEvents, PartitionEvent{
		Type:      PartitionEventStart,
		Timestamp: t,
	})
}

// RecordPartitionEnd records the end of a partition.
func (m *Metrics) RecordPartitionEnd(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var duration time.Duration
	if m.currentPartitionStart != nil {
		duration = t.Sub(*m.currentPartitionStart)
	}

	m.partitionEvents = append(m.partitionEvents, PartitionEvent{
		Type:      PartitionEventEnd,
		Timestamp: t,
		Duration:  duration,
	})
	m.currentPartitionStart = nil
}

// RecordHealingMetric records a healing metric.
func (m *Metrics) RecordHealingMetric(metric HealingMetric) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healingMetrics = append(m.healingMetrics, metric)
}

// RecordBlock records a block production event.
func (m *Metrics) RecordBlock(height int64, producer NodeID, duringPartition bool, timeSinceHeal time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blockMetrics = append(m.blockMetrics, BlockMetric{
		Timestamp:       time.Now(),
		Height:          height,
		Producer:        producer,
		DuringPartition: duringPartition,
		TimeSinceHeal:   timeSinceHeal,
	})
}

// RecordStateSync records a state synchronization event.
func (m *Metrics) RecordStateSync(nodeID NodeID, heightBefore, heightAfter int64, syncDuration time.Duration, hashMatch bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stateMetrics = append(m.stateMetrics, StateMetric{
		NodeID:         nodeID,
		Timestamp:      time.Now(),
		HeightBefore:   heightBefore,
		HeightAfter:    heightAfter,
		SyncDuration:   syncDuration,
		StateHashMatch: hashMatch,
	})
}

// Summary returns a summary of all collected metrics.
type MetricsSummary struct {
	TotalPartitions      int
	TotalPartitionTime   time.Duration
	AveragePartitionTime time.Duration
	MaxPartitionTime     time.Duration
	MinPartitionTime     time.Duration

	AverageTimeToFirstBlock time.Duration
	AverageTimeToConsensus  time.Duration
	MaxTimeToFirstBlock     time.Duration
	MaxTimeToConsensus      time.Duration

	TotalBlocksProduced     int
	BlocksDuringPartition   int
	BlocksAfterHeal         int
	AverageStateSyncTime    time.Duration
	StateSyncSuccessRate    float64
	TotalMessagesReplayed   int
}

// Summary computes and returns a summary of the metrics.
func (m *Metrics) Summary() MetricsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := MetricsSummary{}

	// Count partitions and calculate durations
	var partitionDurations []time.Duration
	for _, event := range m.partitionEvents {
		if event.Type == PartitionEventEnd && event.Duration > 0 {
			summary.TotalPartitions++
			partitionDurations = append(partitionDurations, event.Duration)
			summary.TotalPartitionTime += event.Duration
		}
	}

	if len(partitionDurations) > 0 {
		sort.Slice(partitionDurations, func(i, j int) bool {
			return partitionDurations[i] < partitionDurations[j]
		})
		summary.MinPartitionTime = partitionDurations[0]
		summary.MaxPartitionTime = partitionDurations[len(partitionDurations)-1]
		summary.AveragePartitionTime = summary.TotalPartitionTime / time.Duration(len(partitionDurations))
	}

	// Calculate healing metrics
	var timeToFirstBlocks, timeToConsensus []time.Duration
	for _, hm := range m.healingMetrics {
		if hm.TimeToFirstBlock > 0 {
			timeToFirstBlocks = append(timeToFirstBlocks, hm.TimeToFirstBlock)
		}
		if hm.TimeToConsensus > 0 {
			timeToConsensus = append(timeToConsensus, hm.TimeToConsensus)
		}
		summary.TotalMessagesReplayed += hm.MessagesReplayed
	}

	if len(timeToFirstBlocks) > 0 {
		var total time.Duration
		for _, d := range timeToFirstBlocks {
			total += d
			if d > summary.MaxTimeToFirstBlock {
				summary.MaxTimeToFirstBlock = d
			}
		}
		summary.AverageTimeToFirstBlock = total / time.Duration(len(timeToFirstBlocks))
	}

	if len(timeToConsensus) > 0 {
		var total time.Duration
		for _, d := range timeToConsensus {
			total += d
			if d > summary.MaxTimeToConsensus {
				summary.MaxTimeToConsensus = d
			}
		}
		summary.AverageTimeToConsensus = total / time.Duration(len(timeToConsensus))
	}

	// Calculate block metrics
	summary.TotalBlocksProduced = len(m.blockMetrics)
	for _, bm := range m.blockMetrics {
		if bm.DuringPartition {
			summary.BlocksDuringPartition++
		} else {
			summary.BlocksAfterHeal++
		}
	}

	// Calculate state sync metrics
	syncDurations := make([]time.Duration, 0, len(m.stateMetrics))
	successCount := 0
	for _, sm := range m.stateMetrics {
		syncDurations = append(syncDurations, sm.SyncDuration)
		if sm.StateHashMatch {
			successCount++
		}
	}

	if len(syncDurations) > 0 {
		var total time.Duration
		for _, d := range syncDurations {
			total += d
		}
		summary.AverageStateSyncTime = total / time.Duration(len(syncDurations))
		summary.StateSyncSuccessRate = float64(successCount) / float64(len(m.stateMetrics))
	}

	return summary
}

// Reset clears all collected metrics.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.partitionEvents = make([]PartitionEvent, 0)
	m.healingMetrics = make([]HealingMetric, 0)
	m.blockMetrics = make([]BlockMetric, 0)
	m.stateMetrics = make([]StateMetric, 0)
	m.currentPartitionStart = nil
}

// GetPartitionEvents returns all partition events.
func (m *Metrics) GetPartitionEvents() []PartitionEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]PartitionEvent, len(m.partitionEvents))
	copy(events, m.partitionEvents)
	return events
}

// GetHealingMetrics returns all healing metrics.
func (m *Metrics) GetHealingMetrics() []HealingMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make([]HealingMetric, len(m.healingMetrics))
	copy(metrics, m.healingMetrics)
	return metrics
}

// GetBlockMetrics returns all block metrics.
func (m *Metrics) GetBlockMetrics() []BlockMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make([]BlockMetric, len(m.blockMetrics))
	copy(metrics, m.blockMetrics)
	return metrics
}

// GetStateMetrics returns all state sync metrics.
func (m *Metrics) GetStateMetrics() []StateMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make([]StateMetric, len(m.stateMetrics))
	copy(metrics, m.stateMetrics)
	return metrics
}

// CurrentPartitionDuration returns the duration of the current partition.
// Returns 0 if no partition is active.
func (m *Metrics) CurrentPartitionDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentPartitionStart == nil {
		return 0
	}
	return time.Since(*m.currentPartitionStart)
}

// IsPartitioned returns true if a partition is currently active.
func (m *Metrics) IsPartitioned() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentPartitionStart != nil
}
