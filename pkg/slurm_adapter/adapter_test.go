package slurm_adapter_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	slurm "github.com/virtengine/virtengine/pkg/slurm_adapter"
)

// VE-501: SLURM orchestration adapter tests

// Test constants
const (
	testProviderAddress = "ve1provider123abc456def"
	testVEJobID         = "ve-job-12345"
)

// mockJobSigner implements JobSigner for testing
type mockJobSigner struct {
	providerAddress string
}

//nolint:unparam // addr kept for future multi-address test scenarios
func newMockJobSigner(_ string) *mockJobSigner {
	return &mockJobSigner{providerAddress: testProviderAddress}
}

func (m *mockJobSigner) Sign(data []byte) ([]byte, error) {
	// Return a dummy signature for testing
	return []byte("test-signature-" + string(data[:min(8, len(data))])), nil
}

func (m *mockJobSigner) Verify(data []byte, signature []byte) bool {
	return true
}

func (m *mockJobSigner) GetProviderAddress() string {
	return m.providerAddress
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestSLURMAdapterCreation tests adapter initialization
func TestSLURMAdapterCreation(t *testing.T) {
	config := slurm.DefaultSLURMConfig()

	mockClient := slurm.NewMockSLURMClient()
	mockSigner := newMockJobSigner(testProviderAddress)
	adapter := slurm.NewSLURMAdapter(config, mockClient, mockSigner)

	require.NotNil(t, adapter)
	require.False(t, adapter.IsRunning())
}

// TestSLURMAdapterStartStop tests starting and stopping the adapter
func TestSLURMAdapterStartStop(t *testing.T) {
	config := slurm.DefaultSLURMConfig()
	mockClient := slurm.NewMockSLURMClient()
	mockSigner := newMockJobSigner(testProviderAddress)
	adapter := slurm.NewSLURMAdapter(config, mockClient, mockSigner)

	ctx := context.Background()

	// Start
	err := adapter.Start(ctx)
	require.NoError(t, err)
	require.True(t, adapter.IsRunning())

	// Start again (should be no-op)
	err = adapter.Start(ctx)
	require.NoError(t, err)

	// Stop
	err = adapter.Stop()
	require.NoError(t, err)
	require.False(t, adapter.IsRunning())

	// Stop again (should be no-op)
	err = adapter.Stop()
	require.NoError(t, err)
}

// TestJobSubmission tests submitting a job to SLURM
func TestJobSubmission(t *testing.T) {
	config := slurm.DefaultSLURMConfig()
	mockClient := slurm.NewMockSLURMClient()
	mockSigner := newMockJobSigner(testProviderAddress)
	adapter := slurm.NewSLURMAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	jobSpec := &slurm.SLURMJobSpec{
		JobName:     "test-job",
		Partition:   "default",
		Nodes:       1,
		CPUsPerNode: 4,
		MemoryMB:    8192,
		TimeLimit:   60, // 60 minutes
		Command:     "echo 'Hello HPC'",
		Environment: map[string]string{
			"VIRTENGINE_JOB_ID": testVEJobID,
		},
	}

	job, err := adapter.SubmitJob(ctx, testVEJobID, jobSpec)

	require.NoError(t, err)
	require.NotNil(t, job)
	require.NotEmpty(t, job.SLURMJobID)
}

// TestJobCancellation tests cancelling a SLURM job
func TestJobCancellation(t *testing.T) {
	config := slurm.DefaultSLURMConfig()
	mockClient := slurm.NewMockSLURMClient()
	mockSigner := newMockJobSigner(testProviderAddress)
	adapter := slurm.NewSLURMAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Submit a job first
	jobSpec := &slurm.SLURMJobSpec{
		JobName:     "cancel-test",
		Partition:   "default",
		Nodes:       1,
		CPUsPerNode: 1,
		TimeLimit:   60,
		Command:     "sleep 3600",
	}

	_, err = adapter.SubmitJob(ctx, "ve-cancel-test", jobSpec)
	require.NoError(t, err)

	// Cancel the job
	err = adapter.CancelJob(ctx, "ve-cancel-test")
	require.NoError(t, err)
}

// TestJobStatusTracking tests tracking job status changes
func TestJobStatusTracking(t *testing.T) {
	config := slurm.DefaultSLURMConfig()
	mockClient := slurm.NewMockSLURMClient()
	mockSigner := newMockJobSigner(testProviderAddress)
	adapter := slurm.NewSLURMAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Submit a job
	jobSpec := &slurm.SLURMJobSpec{
		JobName:     "status-test",
		Partition:   "default",
		Nodes:       1,
		CPUsPerNode: 1,
		TimeLimit:   30,
		Command:     "echo test",
	}

	job, err := adapter.SubmitJob(ctx, "ve-status-test", jobSpec)
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
	}{
		{
			name: "valid spec",
			spec: slurm.SLURMJobSpec{
				JobName:     "valid-job",
				Partition:   "default",
				Nodes:       1,
				CPUsPerNode: 4,
				MemoryMB:    4096,
				TimeLimit:   60,
				Command:     "echo hello",
			},
			expectError: false,
		},
		{
			name: "missing name",
			spec: slurm.SLURMJobSpec{
				JobName:     "",
				Partition:   "default",
				Nodes:       1,
				CPUsPerNode: 1,
				TimeLimit:   30,
				Command:     "echo hello",
			},
			expectError: true,
		},
		{
			name: "zero nodes",
			spec: slurm.SLURMJobSpec{
				JobName:     "test-job",
				Partition:   "default",
				Nodes:       0,
				CPUsPerNode: 1,
				TimeLimit:   30,
				Command:     "echo hello",
			},
			expectError: true,
		},
		{
			name: "zero CPUs",
			spec: slurm.SLURMJobSpec{
				JobName:     "test-job",
				Partition:   "default",
				Nodes:       1,
				CPUsPerNode: 0,
				TimeLimit:   30,
				Command:     "echo hello",
			},
			expectError: true,
		},
		{
			name: "zero time limit",
			spec: slurm.SLURMJobSpec{
				JobName:     "test-job",
				Partition:   "default",
				Nodes:       1,
				CPUsPerNode: 1,
				TimeLimit:   0,
				Command:     "echo hello",
			},
			expectError: true,
		},
		{
			name: "missing command and container",
			spec: slurm.SLURMJobSpec{
				JobName:     "test-job",
				Partition:   "default",
				Nodes:       1,
				CPUsPerNode: 1,
				TimeLimit:   30,
				Command:     "",
			},
			expectError: true,
		},
		{
			name: "valid with container image instead of command",
			spec: slurm.SLURMJobSpec{
				JobName:        "container-job",
				Partition:      "default",
				Nodes:          1,
				CPUsPerNode:    2,
				TimeLimit:      60,
				ContainerImage: "nvidia/cuda:12.0",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDefaultSLURMConfig tests default configuration values
func TestDefaultSLURMConfig(t *testing.T) {
	config := slurm.DefaultSLURMConfig()

	require.Equal(t, "virtengine-hpc", config.ClusterName)
	require.Equal(t, "slurmctld", config.ControllerHost)
	require.Equal(t, 6817, config.ControllerPort)
	require.Equal(t, "default", config.DefaultPartition)
	require.Equal(t, time.Second*10, config.JobPollInterval)
	require.Equal(t, 3, config.MaxRetries)
}

// TestSubmitJobNotRunning tests that job submission fails when adapter is not running
func TestSubmitJobNotRunning(t *testing.T) {
	config := slurm.DefaultSLURMConfig()
	mockClient := slurm.NewMockSLURMClient()
	mockSigner := newMockJobSigner(testProviderAddress)
	adapter := slurm.NewSLURMAdapter(config, mockClient, mockSigner)

	// Don't start the adapter

	jobSpec := &slurm.SLURMJobSpec{
		JobName:     "test-job",
		Partition:   "default",
		Nodes:       1,
		CPUsPerNode: 1,
		TimeLimit:   30,
		Command:     "echo hello",
	}

	ctx := context.Background()
	_, err := adapter.SubmitJob(ctx, "ve-test", jobSpec)
	require.Error(t, err)
	require.ErrorIs(t, err, slurm.ErrSLURMNotConnected)
}

// TestCancelJobNotFound tests cancelling a non-existent job
func TestCancelJobNotFound(t *testing.T) {
	config := slurm.DefaultSLURMConfig()
	mockClient := slurm.NewMockSLURMClient()
	mockSigner := newMockJobSigner(testProviderAddress)
	adapter := slurm.NewSLURMAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Try to cancel a job that doesn't exist
	err = adapter.CancelJob(ctx, "nonexistent-job-id")
	require.Error(t, err)
	require.ErrorIs(t, err, slurm.ErrJobNotFound)
}

