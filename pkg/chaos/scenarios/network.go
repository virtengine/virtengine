// Copyright 2024-2025 VirtEngine Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Network chaos scenarios implement various network failure scenarios including:
//   - Network partitions (split-brain, majority/minority splits)
//   - Latency injection (fixed and intermittent)
//   - Packet loss simulation
//   - Bandwidth throttling
//
// These scenarios are designed to test consensus safety and liveness properties
// under adverse network conditions.
package scenarios

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Network experiment types extend the base ExperimentType.
const (
	// ExperimentTypeNetworkPartition represents network partition experiments
	// that isolate groups of nodes from each other.
	ExperimentTypeNetworkPartition ExperimentType = "network-partition"

	// ExperimentTypeNetworkLatency represents experiments that inject
	// latency into network communications.
	ExperimentTypeNetworkLatency ExperimentType = "network-latency"

	// ExperimentTypePacketLoss represents experiments that simulate
	// packet loss on network interfaces.
	ExperimentTypePacketLoss ExperimentType = "packet-loss"

	// ExperimentTypeBandwidth represents experiments that limit
	// network bandwidth.
	ExperimentTypeBandwidth ExperimentType = "bandwidth-limit"
)

// NetworkPartitionSpec contains the configuration for a network partition experiment.
type NetworkPartitionSpec struct {
	// Groups defines the node groups that will be isolated from each other.
	Groups [][]string `json:"groups"`

	// Asymmetric indicates whether the partition is one-way.
	Asymmetric bool `json:"asymmetric"`

	// GradualHeal indicates whether the partition should heal gradually.
	GradualHeal bool `json:"gradual_heal"`

	// Targets lists all nodes affected by this partition.
	Targets []string `json:"targets"`
}

// NetworkLatencySpec contains the configuration for a network latency experiment.
type NetworkLatencySpec struct {
	// Latency is the base latency to inject.
	Latency time.Duration `json:"latency"`

	// Jitter is the random variation added to the target latency.
	Jitter time.Duration `json:"jitter"`

	// Correlation is the percentage (0-100) indicating delay correlation.
	Correlation float64 `json:"correlation"`

	// Intermittent indicates whether latency is applied intermittently.
	Intermittent bool `json:"intermittent"`

	// Interval specifies the on/off interval for intermittent latency.
	Interval time.Duration `json:"interval,omitempty"`

	// Targets lists the nodes or endpoints affected.
	Targets []string `json:"targets"`
}

// PacketLossSpec contains the configuration for a packet loss experiment.
type PacketLossSpec struct {
	// LossPercent is the percentage of packets to drop (0-100).
	LossPercent float64 `json:"loss_percent"`

	// Correlation is the percentage (0-100) for burst loss patterns.
	Correlation float64 `json:"correlation"`

	// Targets lists the nodes or endpoints affected.
	Targets []string `json:"targets"`
}

// BandwidthSpec contains the configuration for a bandwidth throttling experiment.
type BandwidthSpec struct {
	// Rate specifies the bandwidth limit (e.g., "1mbps", "100kbps").
	Rate string `json:"rate"`

	// Buffer is the maximum bytes that can be queued waiting for tokens.
	Buffer uint32 `json:"buffer,omitempty"`

	// Limit is the maximum bytes that can be queued for transmission.
	Limit uint32 `json:"limit,omitempty"`

	// Targets lists the nodes or endpoints affected.
	Targets []string `json:"targets"`
}

// NetworkScenario defines the interface for all network chaos scenarios.
// Implementations must provide scenario metadata, validation, and the ability
// to build an executable Experiment configuration.
//
// This interface mirrors NodeScenario for consistency across chaos scenarios.
type NetworkScenario interface {
	// Name returns the unique identifier for this scenario.
	Name() string

	// Description returns a human-readable description of the scenario
	// and its expected impact on the system.
	Description() string

	// Type returns the category of chaos experiment this scenario creates.
	Type() ExperimentType

	// Build constructs an Experiment from the scenario configuration.
	// Returns an error if the scenario configuration is invalid.
	Build() (*Experiment, error)

	// Validate checks that the scenario configuration is valid and complete.
	// Returns nil if valid, or an error describing the validation failure.
	Validate() error
}

// PartitionScenario implements network partition chaos experiments.
// It can simulate various partition topologies including split-brain,
// majority/minority splits, and isolated nodes.
type PartitionScenario struct {
	// name is the unique identifier for this partition scenario.
	name string

	// description provides details about the partition's expected impact.
	description string

	// Groups defines the node groups that will be isolated from each other.
	// Each inner slice contains node identifiers that can communicate with
	// each other but not with nodes in other groups.
	Groups [][]string

	// Asymmetric indicates whether the partition is one-way.
	// If true, nodes in Groups[0] cannot reach nodes in Groups[1],
	// but Groups[1] can still reach Groups[0].
	Asymmetric bool

	// Duration specifies how long the partition should be maintained.
	Duration time.Duration

	// GradualHeal indicates whether the partition should heal gradually.
	// If true, connectivity is restored incrementally rather than all at once.
	GradualHeal bool
}

// Name returns the scenario identifier.
func (p *PartitionScenario) Name() string {
	if p.name != "" {
		return p.name
	}
	return "network-partition"
}

// Description returns the scenario description.
func (p *PartitionScenario) Description() string {
	if p.description != "" {
		return p.description
	}
	return "Network partition isolating node groups"
}

// Type returns ExperimentTypeNetworkPartition.
func (p *PartitionScenario) Type() ExperimentType {
	return ExperimentTypeNetworkPartition
}

// Validate checks that the partition scenario is properly configured.
func (p *PartitionScenario) Validate() error {
	if len(p.Groups) < 2 {
		return errors.New("partition scenario requires at least 2 groups")
	}

	for i, group := range p.Groups {
		if len(group) == 0 {
			return fmt.Errorf("group %d is empty", i)
		}
		for _, node := range group {
			if strings.TrimSpace(node) == "" {
				return fmt.Errorf("group %d contains empty node identifier", i)
			}
		}
	}

	if p.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	return nil
}

// Build constructs an Experiment from the partition scenario.
func (p *PartitionScenario) Build() (*Experiment, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Collect all targets from all groups
	var allTargets []string
	for _, group := range p.Groups {
		allTargets = append(allTargets, group...)
	}

	return &Experiment{
		Name:        p.Name(),
		Description: p.Description(),
		Type:        ExperimentTypeNetworkPartition,
		Duration:    p.Duration,
		Spec: NetworkPartitionSpec{
			Groups:      p.Groups,
			Asymmetric:  p.Asymmetric,
			GradualHeal: p.GradualHeal,
			Targets:     allTargets,
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeNetworkPartition),
			"chaos.virtengine.io/category": "network",
		},
	}, nil
}

// NewValidatorPartition creates a partition scenario from predefined validator groups.
// This is useful for testing consensus behavior when validators are split into
// multiple isolated groups.
//
// Example:
//
//	// Split validators into two groups
//	scenario := NewValidatorPartition(
//	    [][]string{
//	        {"validator1", "validator2"},
//	        {"validator3", "validator4"},
//	    },
//	    5*time.Minute,
//	)
func NewValidatorPartition(validatorGroups [][]string, duration time.Duration) *PartitionScenario {
	return &PartitionScenario{
		name:        "validator-partition",
		description: "Network partition isolating validator groups to test consensus safety",
		Groups:      validatorGroups,
		Asymmetric:  false,
		Duration:    duration,
		GradualHeal: false,
	}
}

// NewSplitBrain creates a partition scenario that splits nodes into two equal groups.
// This simulates a classic split-brain scenario where neither side has a majority,
// testing how the system handles ambiguous leadership situations.
//
// If the number of nodes is odd, the first group will have one more node.
func NewSplitBrain(nodes []string) *PartitionScenario {
	if len(nodes) == 0 {
		return &PartitionScenario{
			name:        "split-brain",
			description: "Split-brain partition with 50/50 node split",
			Groups:      [][]string{{}, {}},
			Duration:    5 * time.Minute,
		}
	}

	mid := (len(nodes) + 1) / 2
	group1 := make([]string, mid)
	group2 := make([]string, len(nodes)-mid)
	copy(group1, nodes[:mid])
	copy(group2, nodes[mid:])

	return &PartitionScenario{
		name:        "split-brain",
		description: "Split-brain partition with 50/50 node split - tests consensus under equal partition",
		Groups:      [][]string{group1, group2},
		Asymmetric:  false,
		Duration:    5 * time.Minute,
		GradualHeal: false,
	}
}

// NewMajorityMinority creates a partition where a majority of nodes are
// separated from a minority. This tests whether the majority can continue
// making progress while the minority is isolated.
//
// majorityRatio should be between 0.5 and 1.0 (exclusive).
// For example, 0.67 creates a 2/3 majority split.
func NewMajorityMinority(nodes []string, majorityRatio float64) *PartitionScenario {
	if majorityRatio <= 0.5 || majorityRatio >= 1.0 {
		majorityRatio = 0.67 // Default to 2/3 majority
	}

	if len(nodes) == 0 {
		return &PartitionScenario{
			name:        "majority-minority-partition",
			description: fmt.Sprintf("Majority/minority partition (%.0f%%/%.0f%% split)", majorityRatio*100, (1-majorityRatio)*100),
			Groups:      [][]string{{}, {}},
			Duration:    5 * time.Minute,
		}
	}

	majorityCount := int(float64(len(nodes)) * majorityRatio)
	if majorityCount == len(nodes) {
		majorityCount = len(nodes) - 1
	}
	if majorityCount == 0 {
		majorityCount = 1
	}

	majority := make([]string, majorityCount)
	minority := make([]string, len(nodes)-majorityCount)
	copy(majority, nodes[:majorityCount])
	copy(minority, nodes[majorityCount:])

	return &PartitionScenario{
		name:        "majority-minority-partition",
		description: fmt.Sprintf("Majority/minority partition (%.0f%%/%.0f%% split) - tests consensus liveness", majorityRatio*100, (1-majorityRatio)*100),
		Groups:      [][]string{majority, minority},
		Asymmetric:  false,
		Duration:    5 * time.Minute,
		GradualHeal: false,
	}
}

// NewIsolatedNode creates a partition that isolates a single node from all others.
// This tests how the system handles a single node failure and whether the
// remaining nodes can continue making progress.
//
// isolatedIdx specifies which node (by index) to isolate. If out of range,
// the first node is isolated.
func NewIsolatedNode(nodes []string, isolatedIdx int) *PartitionScenario {
	if len(nodes) == 0 {
		return &PartitionScenario{
			name:        "isolated-node",
			description: "Single node isolation",
			Groups:      [][]string{{}, {}},
			Duration:    5 * time.Minute,
		}
	}

	if isolatedIdx < 0 || isolatedIdx >= len(nodes) {
		isolatedIdx = 0
	}

	isolated := []string{nodes[isolatedIdx]}
	remaining := make([]string, 0, len(nodes)-1)
	for i, node := range nodes {
		if i != isolatedIdx {
			remaining = append(remaining, node)
		}
	}

	return &PartitionScenario{
		name:        "isolated-node",
		description: fmt.Sprintf("Single node (%s) isolated from cluster - tests fault tolerance", nodes[isolatedIdx]),
		Groups:      [][]string{isolated, remaining},
		Asymmetric:  false,
		Duration:    5 * time.Minute,
		GradualHeal: false,
	}
}

// NewAsymmetricPartition creates a one-way partition where 'from' cannot reach 'to',
// but 'to' can still reach 'from'. This simulates scenarios where network issues
// affect traffic in only one direction.
func NewAsymmetricPartition(from, to string) *PartitionScenario {
	return &PartitionScenario{
		name:        "asymmetric-partition",
		description: fmt.Sprintf("Asymmetric partition: %s cannot reach %s (but reverse is allowed)", from, to),
		Groups:      [][]string{{from}, {to}},
		Asymmetric:  true,
		Duration:    5 * time.Minute,
		GradualHeal: false,
	}
}

// LatencyScenario implements network latency injection experiments.
// It can simulate high latency, jitter, and correlated delays.
type LatencyScenario struct {
	// name is the unique identifier for this latency scenario.
	name string

	// description provides details about the latency's expected impact.
	description string

	// TargetLatency is the base latency to inject.
	TargetLatency time.Duration

	// Jitter is the random variation added to the target latency.
	// The actual latency will be TargetLatency ± Jitter.
	Jitter time.Duration

	// Correlation is the percentage (0-100) indicating how much the current
	// delay depends on the previous delay. Higher values create more
	// "bursty" latency patterns.
	Correlation float64

	// Duration specifies how long the latency injection should last.
	Duration time.Duration

	// Targets lists the nodes or endpoints affected by latency injection.
	Targets []string

	// Intermittent indicates whether latency should be applied intermittently.
	Intermittent bool

	// Interval specifies the on/off interval for intermittent latency.
	Interval time.Duration
}

// Name returns the scenario identifier.
func (l *LatencyScenario) Name() string {
	if l.name != "" {
		return l.name
	}
	return "network-latency"
}

// Description returns the scenario description.
func (l *LatencyScenario) Description() string {
	if l.description != "" {
		return l.description
	}
	return "Network latency injection"
}

// Type returns ExperimentTypeNetworkLatency.
func (l *LatencyScenario) Type() ExperimentType {
	return ExperimentTypeNetworkLatency
}

// Validate checks that the latency scenario is properly configured.
func (l *LatencyScenario) Validate() error {
	if l.TargetLatency <= 0 {
		return errors.New("target latency must be positive")
	}

	if l.Jitter < 0 {
		return errors.New("jitter cannot be negative")
	}

	if l.Jitter >= l.TargetLatency {
		return errors.New("jitter should be less than target latency")
	}

	if l.Correlation < 0 || l.Correlation > 100 {
		return errors.New("correlation must be between 0 and 100")
	}

	if l.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	if len(l.Targets) == 0 {
		return errors.New("at least one target is required")
	}

	for _, target := range l.Targets {
		if strings.TrimSpace(target) == "" {
			return errors.New("target cannot be empty")
		}
	}

	if l.Intermittent && l.Interval <= 0 {
		return errors.New("interval must be positive for intermittent latency")
	}

	return nil
}

// Build constructs an Experiment from the latency scenario.
func (l *LatencyScenario) Build() (*Experiment, error) {
	if err := l.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        l.Name(),
		Description: l.Description(),
		Type:        ExperimentTypeNetworkLatency,
		Duration:    l.Duration,
		Spec: NetworkLatencySpec{
			Latency:      l.TargetLatency,
			Jitter:       l.Jitter,
			Correlation:  l.Correlation,
			Intermittent: l.Intermittent,
			Interval:     l.Interval,
			Targets:      l.Targets,
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeNetworkLatency),
			"chaos.virtengine.io/category": "network",
		},
	}, nil
}

// NewHighLatency creates a latency scenario with specified latency, jitter, and duration.
// This simulates high network latency between nodes, useful for testing timeout
// handling and consensus delays.
//
// Example:
//
//	// Add 500ms latency with ±100ms jitter for 5 minutes
//	scenario := NewHighLatency(
//	    []string{"validator1", "validator2"},
//	    500*time.Millisecond,
//	    100*time.Millisecond,
//	    5*time.Minute,
//	)
func NewHighLatency(targets []string, latency, jitter, duration time.Duration) *LatencyScenario {
	return &LatencyScenario{
		name:          "high-latency",
		description:   fmt.Sprintf("High latency injection (%v ± %v) - tests timeout and retry handling", latency, jitter),
		TargetLatency: latency,
		Jitter:        jitter,
		Correlation:   25, // Default moderate correlation
		Duration:      duration,
		Targets:       targets,
		Intermittent:  false,
	}
}

// NewIntermittentLatency creates a latency scenario that applies latency in on/off cycles.
// This simulates unstable network conditions where latency comes and goes.
//
// Example:
//
//	// Apply 200ms latency in 30-second on/off cycles
//	scenario := NewIntermittentLatency(
//	    []string{"validator1"},
//	    200*time.Millisecond,
//	    30*time.Second,
//	)
func NewIntermittentLatency(targets []string, latency time.Duration, interval time.Duration) *LatencyScenario {
	return &LatencyScenario{
		name:          "intermittent-latency",
		description:   fmt.Sprintf("Intermittent latency (%v every %v) - tests recovery from transient delays", latency, interval),
		TargetLatency: latency,
		Jitter:        latency / 10, // 10% jitter by default
		Correlation:   50,           // Higher correlation for bursty behavior
		Duration:      10 * time.Minute,
		Targets:       targets,
		Intermittent:  true,
		Interval:      interval,
	}
}

// PacketLossScenario implements packet loss simulation experiments.
// It can simulate random packet loss with configurable correlation.
type PacketLossScenario struct {
	// name is the unique identifier for this packet loss scenario.
	name string

	// description provides details about the packet loss impact.
	description string

	// LossPercent is the percentage of packets to drop (0-100).
	LossPercent float64

	// Correlation is the percentage (0-100) indicating how much the current
	// packet loss depends on the previous packet. Higher values create
	// burst losses instead of random distribution.
	Correlation float64

	// Duration specifies how long the packet loss should be applied.
	Duration time.Duration

	// Targets lists the nodes or endpoints affected by packet loss.
	Targets []string
}

// Name returns the scenario identifier.
func (p *PacketLossScenario) Name() string {
	if p.name != "" {
		return p.name
	}
	return "packet-loss"
}

// Description returns the scenario description.
func (p *PacketLossScenario) Description() string {
	if p.description != "" {
		return p.description
	}
	return "Packet loss simulation"
}

// Type returns ExperimentTypePacketLoss.
func (p *PacketLossScenario) Type() ExperimentType {
	return ExperimentTypePacketLoss
}

// Validate checks that the packet loss scenario is properly configured.
func (p *PacketLossScenario) Validate() error {
	if p.LossPercent < 0 || p.LossPercent > 100 {
		return errors.New("loss percent must be between 0 and 100")
	}

	if p.Correlation < 0 || p.Correlation > 100 {
		return errors.New("correlation must be between 0 and 100")
	}

	if p.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	if len(p.Targets) == 0 {
		return errors.New("at least one target is required")
	}

	for _, target := range p.Targets {
		if strings.TrimSpace(target) == "" {
			return errors.New("target cannot be empty")
		}
	}

	return nil
}

// Build constructs an Experiment from the packet loss scenario.
func (p *PacketLossScenario) Build() (*Experiment, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        p.Name(),
		Description: p.Description(),
		Type:        ExperimentTypePacketLoss,
		Duration:    p.Duration,
		Spec: PacketLossSpec{
			LossPercent: p.LossPercent,
			Correlation: p.Correlation,
			Targets:     p.Targets,
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypePacketLoss),
			"chaos.virtengine.io/category": "network",
		},
	}, nil
}

// NewPacketLoss creates a packet loss scenario with specified loss rate and duration.
// This simulates unreliable network conditions where packets are randomly dropped.
//
// Example:
//
//	// Drop 10% of packets for 5 minutes
//	scenario := NewPacketLoss(
//	    []string{"validator1", "validator2"},
//	    10.0,
//	    5*time.Minute,
//	)
func NewPacketLoss(targets []string, lossPercent float64, duration time.Duration) *PacketLossScenario {
	return &PacketLossScenario{
		name:        "packet-loss",
		description: fmt.Sprintf("Packet loss simulation (%.1f%% loss) - tests message retry and consensus resilience", lossPercent),
		LossPercent: lossPercent,
		Correlation: 25, // Default moderate correlation for realistic loss patterns
		Duration:    duration,
		Targets:     targets,
	}
}

// BandwidthScenario implements bandwidth throttling experiments.
// It can simulate constrained network conditions with limited throughput.
type BandwidthScenario struct {
	// name is the unique identifier for this bandwidth scenario.
	name string

	// description provides details about the bandwidth limit impact.
	description string

	// Rate specifies the bandwidth limit (e.g., "1mbps", "100kbps", "10mbit").
	Rate string

	// Buffer is the maximum number of bytes that can be queued waiting for tokens.
	// Larger values allow more burst traffic but increase latency under congestion.
	Buffer uint32

	// Limit is the maximum number of bytes that can be queued for transmission.
	// This limits the queue size to prevent memory exhaustion.
	Limit uint32

	// Duration specifies how long the bandwidth limit should be applied.
	Duration time.Duration

	// Targets lists the nodes or endpoints affected by bandwidth limiting.
	Targets []string
}

// Name returns the scenario identifier.
func (b *BandwidthScenario) Name() string {
	if b.name != "" {
		return b.name
	}
	return "bandwidth-limit"
}

// Description returns the scenario description.
func (b *BandwidthScenario) Description() string {
	if b.description != "" {
		return b.description
	}
	return "Bandwidth throttling"
}

// Type returns ExperimentTypeBandwidth.
func (b *BandwidthScenario) Type() ExperimentType {
	return ExperimentTypeBandwidth
}

// ratePattern matches valid rate specifications like "1mbps", "100kbps", "10mbit"
var ratePattern = regexp.MustCompile(`^(\d+)(kbps|mbps|gbps|kbit|mbit|gbit|bps|bit)$`)

// Validate checks that the bandwidth scenario is properly configured.
func (b *BandwidthScenario) Validate() error {
	if b.Rate == "" {
		return errors.New("rate is required")
	}

	rate := strings.ToLower(b.Rate)
	if !ratePattern.MatchString(rate) {
		return fmt.Errorf("invalid rate format: %s (expected format like '1mbps', '100kbps', '10mbit')", b.Rate)
	}

	if b.Duration <= 0 {
		return errors.New("duration must be positive")
	}

	if len(b.Targets) == 0 {
		return errors.New("at least one target is required")
	}

	for _, target := range b.Targets {
		if strings.TrimSpace(target) == "" {
			return errors.New("target cannot be empty")
		}
	}

	return nil
}

// Build constructs an Experiment from the bandwidth scenario.
func (b *BandwidthScenario) Build() (*Experiment, error) {
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &Experiment{
		Name:        b.Name(),
		Description: b.Description(),
		Type:        ExperimentTypeBandwidth,
		Duration:    b.Duration,
		Spec: BandwidthSpec{
			Rate:    b.Rate,
			Buffer:  b.Buffer,
			Limit:   b.Limit,
			Targets: b.Targets,
		},
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeBandwidth),
			"chaos.virtengine.io/category": "network",
		},
	}, nil
}

// NewBandwidthLimit creates a bandwidth throttling scenario with specified rate and duration.
// This simulates constrained network conditions for testing behavior under low bandwidth.
//
// Rate should be specified with a unit suffix:
//   - "1mbps" or "1mbit" for 1 megabit per second
//   - "100kbps" or "100kbit" for 100 kilobits per second
//   - "10gbps" or "10gbit" for 10 gigabits per second
//
// Example:
//
//	// Limit bandwidth to 1 Mbps for 5 minutes
//	scenario := NewBandwidthLimit(
//	    []string{"validator1", "validator2"},
//	    "1mbps",
//	    5*time.Minute,
//	)
func NewBandwidthLimit(targets []string, rate string, duration time.Duration) *BandwidthScenario {
	return &BandwidthScenario{
		name:        "bandwidth-limit",
		description: fmt.Sprintf("Bandwidth throttling (%s) - tests behavior under constrained network", rate),
		Rate:        rate,
		Buffer:      1600,  // Default buffer for typical MTU
		Limit:       10000, // Default limit to prevent queue overflow
		Duration:    duration,
		Targets:     targets,
	}
}

// DefaultNetworkScenarios returns a collection of common network chaos scenarios
// suitable for testing VirtEngine validator clusters. These scenarios cover
// various failure modes that can occur in production environments.
//
// The returned scenarios include:
//   - Split-brain partition (50/50 split)
//   - Majority/minority partition (67/33 split)
//   - Single node isolation
//   - High latency injection
//   - Intermittent latency
//   - Packet loss simulation
//   - Bandwidth throttling
//
// Example:
//
//	scenarios := DefaultNetworkScenarios()
//	for _, scenario := range scenarios {
//	    exp, err := scenario.Build()
//	    if err != nil {
//	        log.Printf("Failed to build %s: %v", scenario.Name(), err)
//	        continue
//	    }
//	    runner.Execute(exp)
//	}
func DefaultNetworkScenarios() []NetworkScenario {
	// Default node set for demonstration/testing
	defaultNodes := []string{
		"validator-0",
		"validator-1",
		"validator-2",
		"validator-3",
	}

	return []NetworkScenario{
		// Partition scenarios
		NewSplitBrain(defaultNodes),
		NewMajorityMinority(defaultNodes, 0.67),
		NewIsolatedNode(defaultNodes, 0),
		NewAsymmetricPartition("validator-0", "validator-1"),

		// Latency scenarios
		NewHighLatency(defaultNodes, 500*time.Millisecond, 100*time.Millisecond, 5*time.Minute),
		NewIntermittentLatency(defaultNodes, 200*time.Millisecond, 30*time.Second),

		// Packet loss scenarios
		NewPacketLoss(defaultNodes, 5.0, 5*time.Minute),
		NewPacketLoss(defaultNodes, 20.0, 2*time.Minute),

		// Bandwidth scenarios
		NewBandwidthLimit(defaultNodes, "1mbps", 5*time.Minute),
		NewBandwidthLimit(defaultNodes, "100kbps", 2*time.Minute),
	}
}

