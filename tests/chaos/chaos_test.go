//go:build e2e_integration

// Package chaos_test contains integration tests for the chaos engineering framework.
package chaos_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/chaos"
	"github.com/virtengine/virtengine/pkg/chaos/scenarios"
)

// =============================================================================
// Type Tests - Verify types are properly defined
// =============================================================================

func TestExperimentTypes(t *testing.T) {
	// Verify experiment types are defined
	types := []chaos.ExperimentType{
		chaos.ExperimentTypeNetworkPartition,
		chaos.ExperimentTypeNetworkLatency,
		chaos.ExperimentTypePacketLoss,
		chaos.ExperimentTypeBandwidth,
		chaos.ExperimentTypePodFailure,
		chaos.ExperimentTypeNodeFailure,
		chaos.ExperimentTypeContainerFailure,
		chaos.ExperimentTypeCPUStress,
		chaos.ExperimentTypeMemoryStress,
		chaos.ExperimentTypeDiskStress,
		chaos.ExperimentTypeClockSkew,
		chaos.ExperimentTypeByzantineDoubleSigning,
		chaos.ExperimentTypeByzantineEquivocation,
		chaos.ExperimentTypeByzantineInvalidBlock,
	}

	for _, expType := range types {
		require.NotEmpty(t, expType.String())
	}
}

func TestExperimentStates(t *testing.T) {
	// Verify experiment states
	states := []chaos.ExperimentState{
		chaos.ExperimentStatePending,
		chaos.ExperimentStateRunning,
		chaos.ExperimentStatePaused,
		chaos.ExperimentStateCompleted,
		chaos.ExperimentStateFailed,
		chaos.ExperimentStateAborted,
	}

	for _, state := range states {
		require.True(t, state.Valid())
		require.NotEmpty(t, state.String())
	}

	// Check terminal states
	require.True(t, chaos.ExperimentStateCompleted.IsTerminal())
	require.True(t, chaos.ExperimentStateFailed.IsTerminal())
	require.True(t, chaos.ExperimentStateAborted.IsTerminal())
	require.False(t, chaos.ExperimentStateRunning.IsTerminal())
	require.False(t, chaos.ExperimentStatePending.IsTerminal())
}

func TestExperimentSpec(t *testing.T) {
	spec := &chaos.ExperimentSpec{
		ID:       "test-001",
		Name:     "Test Experiment",
		Type:     chaos.ExperimentTypeNetworkPartition,
		State:    chaos.ExperimentStatePending,
		Duration: 5 * time.Minute,
		Targets: []chaos.Target{
			{
				Type: chaos.TargetTypePod,
				Name: "test-pod",
			},
		},
	}

	require.Equal(t, "test-001", spec.ID)
	require.Equal(t, chaos.ExperimentTypeNetworkPartition, spec.Type)
	require.Equal(t, chaos.ExperimentStatePending, spec.State)
}

// =============================================================================
// Scenario Tests - Verify scenario builders work
// =============================================================================

func TestNetworkPartitionScenario(t *testing.T) {
	nodes := []string{"validator-0", "validator-1", "validator-2", "validator-3"}

	scenario := scenarios.NewSplitBrain(nodes)
	require.NotNil(t, scenario)
	require.Equal(t, "split-brain", scenario.Name())

	err := scenario.Validate()
	require.NoError(t, err)

	exp, err := scenario.Build()
	require.NoError(t, err)
	require.NotNil(t, exp)
	require.Equal(t, scenarios.ExperimentTypeNetworkPartition, exp.Type)
}

func TestMajorityMinorityPartition(t *testing.T) {
	nodes := []string{"validator-0", "validator-1", "validator-2", "validator-3"}

	scenario := scenarios.NewMajorityMinority(nodes, 0.67)
	require.NotNil(t, scenario)
	require.Contains(t, scenario.Name(), "majority")

	err := scenario.Validate()
	require.NoError(t, err)

	exp, err := scenario.Build()
	require.NoError(t, err)
	require.NotNil(t, exp)

	// Verify groups are correctly split
	spec, ok := exp.Spec.(scenarios.NetworkPartitionSpec)
	require.True(t, ok)
	require.Len(t, spec.Groups, 2)
	require.GreaterOrEqual(t, len(spec.Groups[0]), 2) // Majority should have 2+ nodes
}

func TestLatencyScenario(t *testing.T) {
	targets := []string{"validator-0", "validator-1"}

	scenario := scenarios.NewHighLatency(targets, 500*time.Millisecond, 100*time.Millisecond, 5*time.Minute)
	require.NotNil(t, scenario)
	require.Equal(t, "high-latency", scenario.Name())

	err := scenario.Validate()
	require.NoError(t, err)

	exp, err := scenario.Build()
	require.NoError(t, err)
	require.NotNil(t, exp)
}

func TestPacketLossScenario(t *testing.T) {
	targets := []string{"validator-0"}

	scenario := scenarios.NewPacketLoss(targets, 10.0, 5*time.Minute)
	require.NotNil(t, scenario)
	require.Equal(t, "packet-loss", scenario.Name())

	err := scenario.Validate()
	require.NoError(t, err)
}

func TestBandwidthScenario(t *testing.T) {
	targets := []string{"validator-0"}

	scenario := scenarios.NewBandwidthLimit(targets, "1mbps", 5*time.Minute)
	require.NotNil(t, scenario)
	require.Equal(t, "bandwidth-limit", scenario.Name())

	err := scenario.Validate()
	require.NoError(t, err)
}

func TestNodeFailureScenario(t *testing.T) {
	scenario := scenarios.NewValidatorCrash("validator-0", 5*time.Minute)
	require.NotNil(t, scenario)
	require.Contains(t, scenario.Name(), "validator-crash")

	err := scenario.Validate()
	require.NoError(t, err)

	exp, err := scenario.Build()
	require.NoError(t, err)
	require.NotNil(t, exp)
}

func TestProviderCrashScenario(t *testing.T) {
	scenario := scenarios.NewProviderDaemonCrash("provider-0", 2*time.Minute)
	require.NotNil(t, scenario)
	require.Contains(t, scenario.Name(), "provider")

	err := scenario.Validate()
	require.NoError(t, err)
}

func TestResourceScenarios(t *testing.T) {
	// CPU stress
	cpuScenario := scenarios.NewCPUSaturation([]string{"validator-0"}, 80, 5*time.Minute)
	require.NotNil(t, cpuScenario)
	require.NoError(t, cpuScenario.Validate())

	// Memory stress
	memScenario := scenarios.NewMemoryPressure([]string{"validator-0"}, 512, 5*time.Minute)
	require.NotNil(t, memScenario)
	require.NoError(t, memScenario.Validate())

	// Disk stress
	diskScenario := scenarios.NewDiskFill([]string{"validator-0"}, 80, 5*time.Minute)
	require.NotNil(t, diskScenario)
	require.NoError(t, diskScenario.Validate())
}

func TestByzantineScenarios(t *testing.T) {
	validators := []string{"validator-0"}

	// Double signing
	dsScenario := scenarios.NewDoubleSigningScenario(validators, 5*time.Minute)
	require.NotNil(t, dsScenario)
	require.NoError(t, dsScenario.Validate())

	exp, err := dsScenario.Build()
	require.NoError(t, err)
	require.NotNil(t, exp)

	// Equivocation
	eqScenario := scenarios.NewEquivocationScenario(validators, "prevote", 5*time.Minute)
	require.NotNil(t, eqScenario)
	require.NoError(t, eqScenario.Validate())

	// Invalid block
	ibScenario := scenarios.NewInvalidBlockScenario(validators, "malformed", 5*time.Minute)
	require.NotNil(t, ibScenario)
	require.NoError(t, ibScenario.Validate())

	// Message tampering
	mtScenario := scenarios.NewMessageTamperingScenario(validators, "corrupt_signature", 5*time.Minute)
	require.NotNil(t, mtScenario)
	require.NoError(t, mtScenario.Validate())
}

// =============================================================================
// Default Scenarios Tests
// =============================================================================

func TestDefaultNetworkScenarios(t *testing.T) {
	networkScenarios := scenarios.DefaultNetworkScenarios()
	require.NotEmpty(t, networkScenarios)

	for _, scenario := range networkScenarios {
		t.Logf("Testing scenario: %s", scenario.Name())
		require.NoError(t, scenario.Validate())

		exp, err := scenario.Build()
		require.NoError(t, err)
		require.NotNil(t, exp)
	}
}

func TestDefaultNodeScenarios(t *testing.T) {
	nodeScenarios := scenarios.DefaultNodeScenarios()
	require.NotEmpty(t, nodeScenarios)

	for _, scenario := range nodeScenarios {
		t.Logf("Testing scenario: %s", scenario.Name())
		require.NoError(t, scenario.Validate())
	}
}

func TestDefaultResourceScenarios(t *testing.T) {
	resourceScenarios := scenarios.DefaultResourceScenarios()
	require.NotEmpty(t, resourceScenarios)

	for _, scenario := range resourceScenarios {
		t.Logf("Testing scenario: %s", scenario.Name())
		require.NoError(t, scenario.Validate())
	}
}

func TestDefaultByzantineScenarios(t *testing.T) {
	byzantineScenarios := scenarios.DefaultByzantineScenarios()
	require.NotEmpty(t, byzantineScenarios)

	for _, scenario := range byzantineScenarios {
		t.Logf("Testing scenario: %s", scenario.Name())
		require.NoError(t, scenario.Validate())
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestInvalidPartitionScenario(t *testing.T) {
	// Empty groups
	scenario := &scenarios.PartitionScenario{
		Groups:   [][]string{},
		Duration: 5 * time.Minute,
	}
	require.Error(t, scenario.Validate())

	// Only one group
	scenario = &scenarios.PartitionScenario{
		Groups:   [][]string{{"node1"}},
		Duration: 5 * time.Minute,
	}
	require.Error(t, scenario.Validate())

	// Zero duration
	scenario = &scenarios.PartitionScenario{
		Groups:   [][]string{{"node1"}, {"node2"}},
		Duration: 0,
	}
	require.Error(t, scenario.Validate())
}

func TestInvalidLatencyScenario(t *testing.T) {
	// Zero latency
	scenario := &scenarios.LatencyScenario{
		TargetLatency: 0,
		Duration:      5 * time.Minute,
		Targets:       []string{"node1"},
	}
	require.Error(t, scenario.Validate())

	// Jitter > latency
	scenario = &scenarios.LatencyScenario{
		TargetLatency: 100 * time.Millisecond,
		Jitter:        200 * time.Millisecond,
		Duration:      5 * time.Minute,
		Targets:       []string{"node1"},
	}
	require.Error(t, scenario.Validate())

	// No targets
	scenario = &scenarios.LatencyScenario{
		TargetLatency: 100 * time.Millisecond,
		Duration:      5 * time.Minute,
		Targets:       []string{},
	}
	require.Error(t, scenario.Validate())
}

func TestInvalidBandwidthScenario(t *testing.T) {
	// Invalid rate format
	scenario := &scenarios.BandwidthScenario{
		Rate:     "invalid",
		Duration: 5 * time.Minute,
		Targets:  []string{"node1"},
	}
	require.Error(t, scenario.Validate())

	// Empty rate
	scenario = &scenarios.BandwidthScenario{
		Rate:     "",
		Duration: 5 * time.Minute,
		Targets:  []string{"node1"},
	}
	require.Error(t, scenario.Validate())
}
