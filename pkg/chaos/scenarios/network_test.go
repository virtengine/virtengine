// Behaviour tests for the network chaos scenario builders.
//
// These four builders (partition, latency, packet loss, bandwidth) were at 0%
// coverage. PartitionScenario.Build carries the only real logic among them --
// it FLATTENS Groups into a single Targets list, which is what the engine
// actually faults. A regression there would emit an experiment whose Spec
// advertises two groups but an empty target set, i.e. a partition that
// partitions nothing.
package scenarios

import (
	"strings"
	"testing"
	"time"
)

// TestPartitionBuildFlattensGroupTargets pins the flattening in
// PartitionScenario.Build. This is the load-bearing assertion in this file:
// Targets must be every node in every group, in group order.
func TestPartitionBuildFlattensGroupTargets(t *testing.T) {
	s := NewValidatorPartition(
		[][]string{
			{"val-0", "val-1"},
			{"val-2"},
			{"val-3", "val-4"},
		},
		time.Minute,
	)

	exp, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	spec, ok := exp.Spec.(NetworkPartitionSpec)
	if !ok {
		t.Fatalf("Spec is %T, want NetworkPartitionSpec", exp.Spec)
	}

	want := []string{"val-0", "val-1", "val-2", "val-3", "val-4"}
	if len(spec.Targets) != len(want) {
		t.Fatalf("Targets = %v, want %v (every node in every group must be targeted)",
			spec.Targets, want)
	}
	for i := range want {
		if spec.Targets[i] != want[i] {
			t.Errorf("Targets[%d] = %q, want %q (group order must be preserved)",
				i, spec.Targets[i], want[i])
		}
	}

	// Groups must survive intact too -- the engine needs them to know which
	// nodes can still reach each other.
	if len(spec.Groups) != 3 {
		t.Errorf("Groups = %v, want 3 groups", spec.Groups)
	}

	if exp.Type != ExperimentTypeNetworkPartition {
		t.Errorf("Type = %q, want %q", exp.Type, ExperimentTypeNetworkPartition)
	}
	if got := exp.Labels[ChaosLabelCategory]; got != ChaosCategoryNetwork {
		t.Errorf("category label = %q, want %q", got, ChaosCategoryNetwork)
	}
}

// TestPartitionValidateRequiresTwoNonEmptyGroups covers every rejection in
// PartitionScenario.Validate. A partition needs at least two groups to be a
// partition at all.
func TestPartitionValidateRequiresTwoNonEmptyGroups(t *testing.T) {
	t.Run("no groups", func(t *testing.T) {
		s := NewValidatorPartition(nil, time.Minute)
		if err := s.Validate(); err == nil {
			t.Error("Validate with no groups = nil error; want a failure")
		}
	})

	t.Run("single group is not a partition", func(t *testing.T) {
		s := NewValidatorPartition([][]string{{"val-0", "val-1"}}, time.Minute)
		if err := s.Validate(); err == nil {
			t.Error("Validate with one group = nil error; a single group partitions nothing")
		}
	})

	t.Run("empty inner group", func(t *testing.T) {
		s := NewValidatorPartition([][]string{{"val-0"}, {}}, time.Minute)
		if err := s.Validate(); err == nil {
			t.Error("Validate with an empty inner group = nil error; want a failure")
		}
	})

	t.Run("whitespace-only node identifier", func(t *testing.T) {
		s := NewValidatorPartition([][]string{{"val-0"}, {"  "}}, time.Minute)
		if err := s.Validate(); err == nil {
			t.Error("Validate with a whitespace-only node id = nil error; want a failure")
		}
	})

	t.Run("non-positive duration", func(t *testing.T) {
		s := NewValidatorPartition([][]string{{"val-0"}, {"val-1"}}, 0)
		if err := s.Validate(); err == nil {
			t.Error("Validate with a zero duration = nil error; want a failure")
		}
	})
}

// TestPartitionConstructorsProduceValidScenarios drives the convenience
// constructors. Each one encodes a partition topology an operator picks by
// name, so each must produce something that actually validates.
func TestPartitionConstructorsProduceValidScenarios(t *testing.T) {
	nodes := []string{"n0", "n1", "n2", "n3", "n4"}

	cases := []struct {
		name     string
		scenario *PartitionScenario
	}{
		{"split-brain", NewSplitBrain(nodes)},
		{"majority-minority", NewMajorityMinority(nodes, 0.6)},
		{"isolated-node", NewIsolatedNode(nodes, 0)},
		{"asymmetric", NewAsymmetricPartition("n0", "n1")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.scenario.Validate(); err != nil {
				t.Fatalf("Validate: %v (constructor %q produced an invalid scenario)",
					err, tc.name)
			}
			exp, err := tc.scenario.Build()
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			spec, ok := exp.Spec.(NetworkPartitionSpec)
			if !ok {
				t.Fatalf("Spec is %T, want NetworkPartitionSpec", exp.Spec)
			}
			if len(spec.Targets) == 0 {
				t.Error("Targets is empty; the partition would target nothing")
			}
		})
	}
}

// TestSplitBrainIsBalanced pins the split-brain topology: it must split the
// nodes in half, not produce a lopsided or empty partition. Getting this wrong
// would let a majority still reach a minority, i.e. no partition at all.
func TestSplitBrainIsBalanced(t *testing.T) {
	exp, err := NewSplitBrain([]string{"n0", "n1", "n2", "n3"}).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	spec, ok := exp.Spec.(NetworkPartitionSpec)
	if !ok {
		t.Fatalf("Spec is %T, want NetworkPartitionSpec", exp.Spec)
	}
	if len(spec.Groups) != 2 {
		t.Fatalf("Groups = %d, want 2 (a split brain is exactly two halves)", len(spec.Groups))
	}
	if len(spec.Groups[0]) != len(spec.Groups[1]) {
		t.Errorf("split brain is uneven: %v", spec.Groups)
	}
}

// TestIsolatedNodeIsolatesExactlyOne pins the isolated-node topology: the
// chosen index must be alone in its group and everyone else together.
func TestIsolatedNodeIsolatesExactlyOne(t *testing.T) {
	exp, err := NewIsolatedNode([]string{"n0", "n1", "n2"}, 1).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	spec, ok := exp.Spec.(NetworkPartitionSpec)
	if !ok {
		t.Fatalf("Spec is %T, want NetworkPartitionSpec", exp.Spec)
	}
	if len(spec.Groups) != 2 {
		t.Fatalf("Groups = %v, want exactly 2", spec.Groups)
	}
	sizes := []int{len(spec.Groups[0]), len(spec.Groups[1])}
	// The acceptable shapes are (1,2) and (2,1), so the condition is stated
	// positively and negated as a whole: staticcheck QF1001 rewrites any
	// negated `&&`/`||` compound (`!(a && b)`, `!(x || y)`) into De Morgan form,
	// and this lint runs `--new-from-rev=origin/main`, so one such finding in a
	// test file fails the whole lint job. Negating a named boolean is the same
	// assertion with nothing for QF1001 to rewrite.
	isolatedPlusPair := (sizes[0] == 1 && sizes[1] == 2) || (sizes[0] == 2 && sizes[1] == 1)
	if !isolatedPlusPair {
		t.Errorf("group sizes = %v, want one isolated node and one group of 2", sizes)
	}
}

// TestLatencyBuildAndValidate covers the latency builder and its seven
// validation branches. The jitter/latency ordering check is the subtle one:
// jitter >= latency would inject negative effective latency.
func TestLatencyBuildAndValidate(t *testing.T) {
	t.Run("valid build carries the spec", func(t *testing.T) {
		exp, err := NewHighLatency([]string{"val-0"}, 500*time.Millisecond,
			100*time.Millisecond, 5*time.Minute).Build()
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		spec, ok := exp.Spec.(NetworkLatencySpec)
		if !ok {
			t.Fatalf("Spec is %T, want NetworkLatencySpec", exp.Spec)
		}
		if spec.Latency != 500*time.Millisecond {
			t.Errorf("Latency = %v, want 500ms", spec.Latency)
		}
		if spec.Jitter != 100*time.Millisecond {
			t.Errorf("Jitter = %v, want 100ms", spec.Jitter)
		}
		if len(spec.Targets) != 1 || spec.Targets[0] != "val-0" {
			t.Errorf("Targets = %v, want [val-0]", spec.Targets)
		}
		if exp.Type != ExperimentTypeNetworkLatency {
			t.Errorf("Type = %q, want %q", exp.Type, ExperimentTypeNetworkLatency)
		}
	})

	cases := []struct {
		name     string
		mutate   func(*LatencyScenario)
		wantFail bool
	}{
		{"zero latency", func(l *LatencyScenario) { l.TargetLatency = 0 }, true},
		{"negative jitter", func(l *LatencyScenario) { l.Jitter = -time.Second }, true},
		{"jitter >= latency", func(l *LatencyScenario) {
			l.Jitter = l.TargetLatency
		}, true},
		{"correlation above 100", func(l *LatencyScenario) { l.Correlation = 101 }, true},
		{"negative correlation", func(l *LatencyScenario) { l.Correlation = -1 }, true},
		{"zero duration", func(l *LatencyScenario) { l.Duration = 0 }, true},
		{"no targets", func(l *LatencyScenario) { l.Targets = nil }, true},
		{"empty target", func(l *LatencyScenario) { l.Targets = []string{"  "} }, true},
		{"intermittent with no interval", func(l *LatencyScenario) {
			l.Intermittent = true
			l.Interval = 0
		}, true},
		{"intermittent with an interval is valid", func(l *LatencyScenario) {
			l.Intermittent = true
			l.Interval = time.Second
		}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewHighLatency([]string{"val-0"}, 500*time.Millisecond,
				100*time.Millisecond, 5*time.Minute)
			tc.mutate(s)

			err := s.Validate()
			if tc.wantFail && err == nil {
				t.Errorf("Validate = nil error; want a failure")
			}
			if !tc.wantFail && err != nil {
				t.Errorf("Validate = %v; want nil", err)
			}
		})
	}
}

// TestPacketLossBuildAndValidate covers the packet-loss builder. The 0-100
// bound is the invariant: outside it, the injection either never fires or
// fires for packets that do not exist.
func TestPacketLossBuildAndValidate(t *testing.T) {
	t.Run("valid build carries the spec", func(t *testing.T) {
		exp, err := NewPacketLoss([]string{"val-0"}, 25, time.Minute).Build()
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		spec, ok := exp.Spec.(PacketLossSpec)
		if !ok {
			t.Fatalf("Spec is %T, want PacketLossSpec", exp.Spec)
		}
		if spec.LossPercent != 25 {
			t.Errorf("LossPercent = %v, want 25", spec.LossPercent)
		}
		if exp.Type != ExperimentTypePacketLoss {
			t.Errorf("Type = %q, want %q", exp.Type, ExperimentTypePacketLoss)
		}
	})

	cases := []struct {
		name   string
		mutate func(*PacketLossScenario)
		fails  bool
	}{
		{"loss above 100", func(p *PacketLossScenario) { p.LossPercent = 101 }, true},
		{"negative loss", func(p *PacketLossScenario) { p.LossPercent = -1 }, true},
		{"correlation above 100", func(p *PacketLossScenario) { p.Correlation = 101 }, true},
		{"zero duration", func(p *PacketLossScenario) { p.Duration = 0 }, true},
		{"no targets", func(p *PacketLossScenario) { p.Targets = nil }, true},
		{"empty target", func(p *PacketLossScenario) { p.Targets = []string{""} }, true},
		{"0% and 100% are valid", func(p *PacketLossScenario) { p.LossPercent = 100 }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewPacketLoss([]string{"val-0"}, 25, time.Minute)
			tc.mutate(s)
			err := s.Validate()
			if tc.fails && err == nil {
				t.Errorf("Validate = nil error; want a failure")
			}
			if !tc.fails && err != nil {
				t.Errorf("Validate = %v; want nil", err)
			}
		})
	}
}

// TestBandwidthBuildAndValidate covers the bandwidth builder, whose Validate
// gates a rate FORMAT via ratePattern. The accepted spellings are a real
// contract with the chaos backend, so they are pinned rather than assumed.
func TestBandwidthBuildAndValidate(t *testing.T) {
	t.Run("valid build carries the spec", func(t *testing.T) {
		exp, err := NewBandwidthLimit([]string{"val-0"}, "10mbps", time.Minute).Build()
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		spec, ok := exp.Spec.(BandwidthSpec)
		if !ok {
			t.Fatalf("Spec is %T, want BandwidthSpec", exp.Spec)
		}
		if spec.Rate != "10mbps" {
			t.Errorf("Rate = %q, want %q", spec.Rate, "10mbps")
		}
		if exp.Type != ExperimentTypeBandwidth {
			t.Errorf("Type = %q, want %q", exp.Type, ExperimentTypeBandwidth)
		}
	})

	t.Run("rate formats", func(t *testing.T) {
		// The accepted set is exactly ratePattern's unit alternation. Upper
		// case is normalised by strings.ToLower before matching, so "1MBPS"
		// is valid too.
		for _, rate := range []string{
			"1mbps", "100kbps", "10mbit", "1gbps", "50kbit", "5gbit",
			"1bps", "8bit", "1MBPS",
		} {
			s := NewBandwidthLimit([]string{"val-0"}, rate, time.Minute)
			if err := s.Validate(); err != nil {
				t.Errorf("Validate(%q) = %v; want nil (this unit is in the pattern)",
					rate, err)
			}
		}
	})

	t.Run("rate rejections", func(t *testing.T) {
		// A bare number with no unit is the interesting rejection: "10" and
		// "1M" look plausible to an operator but have no unit, and silently
		// accepting them would mean the chaos backend gets an unparseable rate.
		for _, rate := range []string{
			"", "fast", "10", "mbps", "10 gigabits", "1M", "500K", "-1mbps",
			"1.5mbps", "1 mbps",
		} {
			s := NewBandwidthLimit([]string{"val-0"}, rate, time.Minute)
			if err := s.Validate(); err == nil {
				t.Errorf("Validate(%q) = nil error; want a failure (no valid unit)", rate)
			}
		}
	})

	t.Run("other validation branches", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(*BandwidthScenario)
			fails  bool
		}{
			{"zero duration", func(b *BandwidthScenario) { b.Duration = 0 }, true},
			{"no targets", func(b *BandwidthScenario) { b.Targets = nil }, true},
			{"whitespace target", func(b *BandwidthScenario) {
				b.Targets = []string{" "}
			}, true},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				s := NewBandwidthLimit([]string{"val-0"}, "10mbps", time.Minute)
				tc.mutate(s)
				err := s.Validate()
				if tc.fails && err == nil {
					t.Error("Validate = nil error; want a failure")
				}
				if !tc.fails && err != nil {
					t.Errorf("Validate = %v; want nil", err)
				}
			})
		}
	})
}

// TestScenarioNameFallbacks pins the name/description fallback contract: a
// zero-value scenario still reports a usable identity, because the engine
// writes these into experiment records.
func TestScenarioNameFallbacks(t *testing.T) {
	t.Run("partition falls back", func(t *testing.T) {
		s := &PartitionScenario{}
		if got := s.Name(); got != "network-partition" {
			t.Errorf("Name() = %q, want %q", got, "network-partition")
		}
		if strings.TrimSpace(s.Description()) == "" {
			t.Error("Description() is empty")
		}
	})

	t.Run("explicit name wins", func(t *testing.T) {
		s := &PartitionScenario{name: "custom-partition", description: "custom desc"}
		if got := s.Name(); got != "custom-partition" {
			t.Errorf("Name() = %q, want %q", got, "custom-partition")
		}
		if got := s.Description(); got != "custom desc" {
			t.Errorf("Description() = %q, want %q", got, "custom desc")
		}
	})
}

// TestTypeConstants pins the Type() accessors. If one of these returned the
// wrong constant, the chaos engine would dispatch a partition as a packet-loss
// experiment -- a silent, dangerous mismatch with no error anywhere.
func TestTypeConstants(t *testing.T) {
	cases := []struct {
		name string
		got  ExperimentType
		want ExperimentType
	}{
		{"double-signing", NewDoubleSigningScenario([]string{"v"}, time.Minute).Type(),
			ExperimentTypeByzantineDoubleSigning},
		{"equivocation", NewEquivocationScenario([]string{"v"}, "prevote", time.Minute).Type(),
			ExperimentTypeByzantineEquivocation},
		{"invalid-block", NewInvalidBlockScenario([]string{"v"}, "malformed", time.Minute).Type(),
			ExperimentTypeByzantineInvalidBlock},
		{"tampering", NewMessageTamperingScenario([]string{"v"}, "replay", time.Minute).Type(),
			ExperimentTypeByzantineMessageTampering},
		{"selective-forwarding", NewSelectiveForwardingScenario([]string{"v"}, nil, time.Minute).Type(),
			ExperimentTypeByzantineSelectiveForwarding},
		{"partition", NewValidatorPartition([][]string{{"a"}, {"b"}}, time.Minute).Type(),
			ExperimentTypeNetworkPartition},
		{"latency", NewHighLatency([]string{"v"}, time.Second, 0, time.Minute).Type(),
			ExperimentTypeNetworkLatency},
		{"packet-loss", NewPacketLoss([]string{"v"}, 10, time.Minute).Type(),
			ExperimentTypePacketLoss},
		{"bandwidth", NewBandwidthLimit([]string{"v"}, "1mbps", time.Minute).Type(),
			ExperimentTypeBandwidth},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("Type() = %q, want %q", tc.got, tc.want)
			}
		})
	}
}
