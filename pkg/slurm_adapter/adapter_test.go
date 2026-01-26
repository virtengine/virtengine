package slurm_adapter_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	slurm "pkg.akt.dev/node/pkg/slurm_adapter"
)

// VE-501: SLURM orchestration adapter tests

// TestSLURMAdapterCreation tests adapter initialization
func TestSLURMAdapterCreation(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	// Use mock client for testing
	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	require.NotNil(t, adapter)
	require.Equal(t, "test-cluster", adapter.ClusterName())
}

// TestJobSubmission tests submitting a job to SLURM
func TestJobSubmission(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	jobSpec := slurm.SLURMJobSpec{
		Name:        "test-job",
		Partition:   "default",
		Nodes:       1,
		Tasks:       4,
		CPUsPerTask: 2,
		MemoryMB:    8192,
		TimeLimit:   "01:00:00",
		Command:     "echo 'Hello HPC'",
		Environment: map[string]string{
			"VIRTENGINE_JOB_ID": "12345",
		},
	}

	ctx := context.Background()
	jobID, err := adapter.SubmitJob(ctx, jobSpec)

	require.NoError(t, err)
	require.NotEmpty(t, jobID)
}

// TestJobCancellation tests cancelling a SLURM job
func TestJobCancellation(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	// Submit a job first
	jobSpec := slurm.SLURMJobSpec{
		Name:      "cancel-test",
		Partition: "default",
		Nodes:     1,
		Tasks:     1,
		Command:   "sleep 3600",
	}

	jobID, err := adapter.SubmitJob(ctx, jobSpec)
	require.NoError(t, err)

	// Cancel the job
	err = adapter.CancelJob(ctx, jobID)
	require.NoError(t, err)

	// Verify job is cancelled
	job, err := adapter.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	require.Equal(t, slurm.SLURMJobStateCancelled, job.State)
}

// TestJobStatusTracking tests tracking job status changes
func TestJobStatusTracking(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	// Submit a job
	jobSpec := slurm.SLURMJobSpec{
		Name:      "status-test",
		Partition: "default",
		Nodes:     1,
		Tasks:     1,
		Command:   "echo test",
	}

	jobID, err := adapter.SubmitJob(ctx, jobSpec)
	require.NoError(t, err)

	// Get initial status
	job, err := adapter.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	require.NotNil(t, job)

	// Status should be pending or running
	require.True(t,
		job.State == slurm.SLURMJobStatePending ||
			job.State == slurm.SLURMJobStateRunning,
		"job should be pending or running")
}

// TestSLURMJobSpecValidation tests job specification validation
func TestSLURMJobSpecValidation(t *testing.T) {
	testCases := []struct {
		name        string
		spec        slurm.SLURMJobSpec
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid spec",
			spec: slurm.SLURMJobSpec{
				Name:        "valid-job",
				Partition:   "default",
				Nodes:       1,
				Tasks:       4,
				CPUsPerTask: 1,
				MemoryMB:    4096,
				TimeLimit:   "01:00:00",
				Command:     "echo hello",
			},
			expectError: false,
		},
		{
			name: "missing name",
			spec: slurm.SLURMJobSpec{
				Name:      "",
				Partition: "default",
				Command:   "echo hello",
			},
			expectError: true,
			errorMsg:    "job name is required",
		},
		{
			name: "missing partition",
			spec: slurm.SLURMJobSpec{
				Name:      "test-job",
				Partition: "",
				Command:   "echo hello",
			},
			expectError: true,
			errorMsg:    "partition is required",
		},
		{
			name: "missing command",
			spec: slurm.SLURMJobSpec{
				Name:      "test-job",
				Partition: "default",
				Command:   "",
			},
			expectError: true,
			errorMsg:    "command is required",
		},
		{
			name: "zero nodes",
			spec: slurm.SLURMJobSpec{
				Name:      "test-job",
				Partition: "default",
				Nodes:     0,
				Command:   "echo hello",
			},
			expectError: true,
			errorMsg:    "nodes must be positive",
		},
		{
			name: "invalid time format",
			spec: slurm.SLURMJobSpec{
				Name:      "test-job",
				Partition: "default",
				Nodes:     1,
				Tasks:     1,
				TimeLimit: "invalid",
				Command:   "echo hello",
			},
			expectError: true,
			errorMsg:    "invalid time limit format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestUsageMetricsCollection tests collecting job usage metrics
func TestUsageMetricsCollection(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	// Submit and complete a job
	jobSpec := slurm.SLURMJobSpec{
		Name:        "metrics-test",
		Partition:   "default",
		Nodes:       2,
		Tasks:       8,
		CPUsPerTask: 4,
		MemoryMB:    16384,
		TimeLimit:   "00:30:00",
		Command:     "echo test",
	}

	jobID, err := adapter.SubmitJob(ctx, jobSpec)
	require.NoError(t, err)

	// Simulate job completion in mock
	mockClient.CompleteJob(jobID)

	// Get usage metrics
	metrics, err := adapter.GetJobMetrics(ctx, jobID)
	require.NoError(t, err)
	require.NotNil(t, metrics)

	// Verify metrics are populated
	require.Greater(t, metrics.CPUSeconds, uint64(0))
	require.Greater(t, metrics.MemoryMBHours, uint64(0))
	require.Greater(t, metrics.WallTimeSeconds, uint64(0))
}

// TestSignedStatusReport tests creating signed status reports
func TestSignedStatusReport(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
		ProviderKey:    "test-provider-key",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	// Submit a job
	jobSpec := slurm.SLURMJobSpec{
		Name:      "signed-report-test",
		Partition: "default",
		Nodes:     1,
		Tasks:     1,
		Command:   "echo test",
	}

	jobID, err := adapter.SubmitJob(ctx, jobSpec)
	require.NoError(t, err)

	// Create signed status report
	report, err := adapter.CreateStatusReport(ctx, jobID)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify report fields
	require.Equal(t, jobID, report.SLURMJobID)
	require.NotEmpty(t, report.Timestamp)
	require.NotEmpty(t, report.Signature)
}

// TestPartitionInfoRetrieval tests getting partition information
func TestPartitionInfoRetrieval(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	partitions, err := adapter.GetPartitions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, partitions)

	// Verify partition structure
	for _, p := range partitions {
		require.NotEmpty(t, p.Name)
		require.Greater(t, p.TotalNodes, uint32(0))
	}
}

// TestNodeInfoRetrieval tests getting node information
func TestNodeInfoRetrieval(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	nodes, err := adapter.GetNodes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, nodes)

	// Verify node structure
	for _, n := range nodes {
		require.NotEmpty(t, n.Name)
		require.Greater(t, n.CPUs, uint32(0))
		require.Greater(t, n.MemoryMB, uint64(0))
	}
}

// TestConfigValidation tests SLURM config validation
func TestConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      slurm.SLURMConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: slurm.SLURMConfig{
				ControllerHost: "localhost",
				ControllerPort: 6817,
				ClusterName:    "test-cluster",
				PartitionName:  "default",
			},
			expectError: false,
		},
		{
			name: "missing controller host",
			config: slurm.SLURMConfig{
				ControllerHost: "",
				ControllerPort: 6817,
				ClusterName:    "test-cluster",
			},
			expectError: true,
			errorMsg:    "controller host is required",
		},
		{
			name: "invalid port",
			config: slurm.SLURMConfig{
				ControllerHost: "localhost",
				ControllerPort: 0,
				ClusterName:    "test-cluster",
			},
			expectError: true,
			errorMsg:    "controller port must be valid",
		},
		{
			name: "missing cluster name",
			config: slurm.SLURMConfig{
				ControllerHost: "localhost",
				ControllerPort: 6817,
				ClusterName:    "",
			},
			expectError: true,
			errorMsg:    "cluster name is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestJobRetryBehavior tests job submission retry on failure
func TestJobRetryBehavior(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
	}

	// Create mock that fails first 2 attempts
	mockClient := slurm.NewMockSLURMClient()
	mockClient.SetFailCount(2)

	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	jobSpec := slurm.SLURMJobSpec{
		Name:      "retry-test",
		Partition: "default",
		Nodes:     1,
		Tasks:     1,
		Command:   "echo hello",
	}

	// Should succeed after retries
	jobID, err := adapter.SubmitJob(ctx, jobSpec)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)

	// Verify retry count
	require.Equal(t, 2, mockClient.FailedAttempts())
}

// TestJobTerminationMidExecution tests handling mid-job termination
func TestJobTerminationMidExecution(t *testing.T) {
	config := slurm.SLURMConfig{
		ControllerHost: "localhost",
		ControllerPort: 6817,
		ClusterName:    "test-cluster",
		PartitionName:  "default",
	}

	mockClient := slurm.NewMockSLURMClient()
	adapter := slurm.NewSLURMAdapter(config, mockClient)

	ctx := context.Background()

	// Submit a long-running job
	jobSpec := slurm.SLURMJobSpec{
		Name:      "termination-test",
		Partition: "default",
		Nodes:     1,
		Tasks:     1,
		TimeLimit: "24:00:00",
		Command:   "sleep 86400",
	}

	jobID, err := adapter.SubmitJob(ctx, jobSpec)
	require.NoError(t, err)

	// Simulate job starting
	mockClient.StartJob(jobID)

	// Get partial metrics before cancellation
	job, err := adapter.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	require.Equal(t, slurm.SLURMJobStateRunning, job.State)

	// Cancel mid-execution
	err = adapter.CancelJob(ctx, jobID)
	require.NoError(t, err)

	// Get final metrics - should have partial usage
	metrics, err := adapter.GetJobMetrics(ctx, jobID)
	require.NoError(t, err)
	require.Greater(t, metrics.WallTimeSeconds, uint64(0), "should have partial wall time")
}
