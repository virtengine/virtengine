package partition

import (
	"context"
	"math/rand" //nolint:gosec // G404: simulation uses weak random for non-security network conditions
	"sync"
	"time"
)

// NodePair represents a pair of nodes for filtering.
type NodePair struct {
	From NodeID
	To   NodeID
}

// FilterRule defines how messages between nodes should be filtered.
type FilterRule struct {
	// Blocked indicates all messages are blocked.
	Blocked bool

	// DropRate is the probability (0.0-1.0) of dropping a message.
	DropRate float64

	// DelayMin is the minimum delay to add to messages.
	DelayMin time.Duration

	// DelayMax is the maximum delay to add to messages.
	DelayMax time.Duration

	// DuplicateRate is the probability of duplicating a message.
	DuplicateRate float64

	// ReorderWindow is the time window for potential message reordering.
	ReorderWindow time.Duration
}

// MessageFilter manages message filtering rules for partition simulation.
type MessageFilter struct {
	mu sync.RWMutex

	// rules maps node pairs to filter rules.
	rules map[NodePair]FilterRule

	// globalRule applies to all connections not in rules.
	globalRule *FilterRule

	// messageHistory tracks message hashes for replay detection.
	messageHistory map[string]time.Time

	// historyRetention is how long to keep message history.
	historyRetention time.Duration

	// rand is used for probabilistic filtering.
	rand *rand.Rand
}

// NewMessageFilter creates a new message filter.
func NewMessageFilter() *MessageFilter {
	return &MessageFilter{
		rules:            make(map[NodePair]FilterRule),
		messageHistory:   make(map[string]time.Time),
		historyRetention: 10 * time.Minute,
		rand:             rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec // G404: simulation randomness for fault injection
	}
}

// SetBlocked sets whether messages between two nodes are blocked.
func (f *MessageFilter) SetBlocked(from, to NodeID, blocked bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	pair := NodePair{From: from, To: to}
	rule := f.rules[pair]
	rule.Blocked = blocked
	f.rules[pair] = rule
}

// SetDropRate sets the message drop rate between two nodes.
func (f *MessageFilter) SetDropRate(from, to NodeID, rate float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	pair := NodePair{From: from, To: to}
	rule := f.rules[pair]
	rule.DropRate = rate
	f.rules[pair] = rule
}

// SetDelay sets the message delay range between two nodes.
func (f *MessageFilter) SetDelay(from, to NodeID, min, max time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	pair := NodePair{From: from, To: to}
	rule := f.rules[pair]
	rule.DelayMin = min
	rule.DelayMax = max
	f.rules[pair] = rule
}

// SetDuplicateRate sets the message duplication rate between two nodes.
func (f *MessageFilter) SetDuplicateRate(from, to NodeID, rate float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	pair := NodePair{From: from, To: to}
	rule := f.rules[pair]
	rule.DuplicateRate = rate
	f.rules[pair] = rule
}

// SetGlobalRule sets a global filter rule for all connections.
func (f *MessageFilter) SetGlobalRule(rule *FilterRule) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.globalRule = rule
}

// ClearRule removes the filter rule for a specific node pair.
func (f *MessageFilter) ClearRule(from, to NodeID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rules, NodePair{From: from, To: to})
}

// ClearAll removes all filter rules and optionally clears message history.
func (f *MessageFilter) ClearAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rules = make(map[NodePair]FilterRule)
	f.globalRule = nil
}

// Reset clears all filter rules AND message history. Use for test cleanup.
func (f *MessageFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rules = make(map[NodePair]FilterRule)
	f.globalRule = nil
	f.messageHistory = make(map[string]time.Time)
}

// GetBlocked returns all blocked node pairs.
func (f *MessageFilter) GetBlocked() map[NodePair]bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	blocked := make(map[NodePair]bool)
	for pair, rule := range f.rules {
		if rule.Blocked {
			blocked[pair] = true
		}
	}
	return blocked
}

// ShouldFilter determines if a message should be filtered and returns
// the filtering decision with optional delay.
type FilterDecision struct {
	// Drop indicates the message should be dropped.
	Drop bool

	// Delay is the delay to apply before delivering the message.
	Delay time.Duration

	// Duplicate indicates the message should be duplicated.
	Duplicate bool

	// Reason provides context for why the message was filtered.
	Reason string
}

// Filter evaluates filter rules for a message between two nodes.
func (f *MessageFilter) Filter(from, to NodeID) FilterDecision {
	f.mu.RLock()
	defer f.mu.RUnlock()

	pair := NodePair{From: from, To: to}
	rule, ok := f.rules[pair]
	if !ok && f.globalRule != nil {
		rule = *f.globalRule
	}

	decision := FilterDecision{}

	// Check if blocked
	if rule.Blocked {
		decision.Drop = true
		decision.Reason = "connection blocked"
		return decision
	}

	// Check drop rate
	if rule.DropRate > 0 && f.rand.Float64() < rule.DropRate {
		decision.Drop = true
		decision.Reason = "random drop"
		return decision
	}

	// Calculate delay
	if rule.DelayMax > 0 {
		if rule.DelayMin >= rule.DelayMax {
			decision.Delay = rule.DelayMin
		} else {
			delayRange := rule.DelayMax - rule.DelayMin
			decision.Delay = rule.DelayMin + time.Duration(f.rand.Int63n(int64(delayRange)))
		}
	}

	// Check duplication
	if rule.DuplicateRate > 0 && f.rand.Float64() < rule.DuplicateRate {
		decision.Duplicate = true
	}

	return decision
}

// IsMessageReplayed checks if a message hash has been seen before.
// Returns true if this is a replay.
func (f *MessageFilter) IsMessageReplayed(messageHash string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, seen := f.messageHistory[messageHash]; seen {
		return true
	}

	f.messageHistory[messageHash] = time.Now()
	return false
}

// CleanupHistory removes old entries from message history.
func (f *MessageFilter) CleanupHistory(ctx context.Context) {
	f.mu.Lock()
	defer f.mu.Unlock()

	cutoff := time.Now().Add(-f.historyRetention)
	for hash, timestamp := range f.messageHistory {
		if timestamp.Before(cutoff) {
			delete(f.messageHistory, hash)
		}
	}
}

// SetHistoryRetention sets how long message hashes are retained.
func (f *MessageFilter) SetHistoryRetention(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.historyRetention = d
}

// MessageHistorySize returns the current size of the message history.
func (f *MessageFilter) MessageHistorySize() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.messageHistory)
}

// InterceptResult represents the result of message interception.
type InterceptResult struct {
	// Allow indicates the message should be delivered.
	Allow bool

	// Delay is an optional delay before delivery.
	Delay time.Duration

	// Copies is the number of copies to deliver (1 for normal, 2+ for duplicates).
	Copies int

	// Error is set if interception failed.
	Error error
}

// Intercept processes a message through the filter and returns the result.
func (f *MessageFilter) Intercept(from, to NodeID, messageHash string) InterceptResult {
	// Check for replay
	if messageHash != "" && f.IsMessageReplayed(messageHash) {
		return InterceptResult{
			Allow:  false,
			Copies: 0,
		}
	}

	decision := f.Filter(from, to)

	if decision.Drop {
		return InterceptResult{
			Allow:  false,
			Copies: 0,
		}
	}

	copies := 1
	if decision.Duplicate {
		copies = 2
	}

	return InterceptResult{
		Allow:  true,
		Delay:  decision.Delay,
		Copies: copies,
	}
}

// NetworkCondition represents network conditions for simulation.
type NetworkCondition struct {
	Name        string
	Description string
	DropRate    float64
	DelayMin    time.Duration
	DelayMax    time.Duration
}

// PredefinedConditions provides common network condition presets.
var PredefinedConditions = map[string]NetworkCondition{
	"healthy": {
		Name:        "healthy",
		Description: "Normal network conditions",
		DropRate:    0,
		DelayMin:    0,
		DelayMax:    0,
	},
	"flaky": {
		Name:        "flaky",
		Description: "Intermittent packet loss",
		DropRate:    0.05,
		DelayMin:    10 * time.Millisecond,
		DelayMax:    100 * time.Millisecond,
	},
	"congested": {
		Name:        "congested",
		Description: "High latency network",
		DropRate:    0.01,
		DelayMin:    100 * time.Millisecond,
		DelayMax:    500 * time.Millisecond,
	},
	"degraded": {
		Name:        "degraded",
		Description: "Significantly degraded network",
		DropRate:    0.10,
		DelayMin:    200 * time.Millisecond,
		DelayMax:    1000 * time.Millisecond,
	},
	"severe": {
		Name:        "severe",
		Description: "Severely degraded network",
		DropRate:    0.25,
		DelayMin:    500 * time.Millisecond,
		DelayMax:    2000 * time.Millisecond,
	},
}

// ApplyCondition applies a network condition to all connections.
func (f *MessageFilter) ApplyCondition(condition NetworkCondition) {
	f.SetGlobalRule(&FilterRule{
		DropRate: condition.DropRate,
		DelayMin: condition.DelayMin,
		DelayMax: condition.DelayMax,
	})
}
