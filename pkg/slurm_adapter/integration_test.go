//go:build integration
// +build integration

package slurm_adapter_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slurm "github.com/virtengine/virtengine/pkg/slurm_adapter"
)

// VE-2020: Integration tests for SLURM SSH adapter
// These tests require a real SLURM cluster connection
// Run with: go test -tags=integration -v ./pkg/slurm_adapter/...

// Environment variables for SLURM connection:
// SLURM_TEST_HOST - SLURM login node hostname
// SLURM_TEST_PORT - SSH port (default: 22)
// SLURM_TEST_USER - SSH username
// SLURM_TEST_KEY  - Path to SSH private key
// SLURM_TEST_PARTITION - Partition to use for tests

func getTestConfig(t *testing.T) slurm.SSHConfig {
	t.Helper()

	host := os.Getenv("SLURM_TEST_HOST")
	if host == "" {
		t.Skip("SLURM_TEST_HOST not set, skipping integration test")
	}

	user := os.Getenv("SLURM_TEST_USER")
	if user == "" {
		t.Skip("SLURM_TEST_USER not set, skipping integration test")
	}

	keyPath := os.Getenv("SLURM_TEST_KEY")
	if keyPath == "" {
		t.Skip("SLURM_TEST_KEY not set, skipping integration test")
	}

	config := slurm.DefaultSSHConfig()
	config.Host = host
	config.User = user
	config.PrivateKeyPath = keyPath

	if portStr := os.Getenv("SLURM_TEST_PORT"); portStr != "" {
		// Parse port if provided
		var port int
		_, err := fmt.Sscanf(portStr, "%d", &port)
		if err == nil {
			config.Port = port
		}
	}

	return config
}

func getTestPartition(t *testing.T) string {
	partition := os.Getenv("SLURM_TEST_PARTITION")
	if partition == "" {
		return "debug" // Default test partition
	}
	return partition
}

func TestIntegration_SSHConnection(t *testing.T) {
	config := getTestConfig(t)
	partition := getTestPartition(t)

	client, err := slurm.NewSSHSLURMClient(config, "test-cluster", partition)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test connection
	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect()

	assert.True(t, client.IsConnected())
}

func TestIntegration_ListPartitions(t *testing.T) {
	config := getTestConfig(t)
	partition := getTestPartition(t)

	client, err := slurm.NewSSHSLURMClient(config, "test-cluster", partition)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect()

	// List partitions
	partitions, err := client.ListPartitions(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, partitions)

	t.Logf("Found %d partitions:", len(partitions))
	for _, p := range partitions {
		t.Logf("  - %s: %s, %d nodes", p.Name, p.State, p.Nodes)
	}
}

func TestIntegration_ListNodes(t *testing.T) {
	config := getTestConfig(t)
	partition := getTestPartition(t)

	client, err := slurm.NewSSHSLURMClient(config, "test-cluster", partition)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect()

	// List nodes
	nodes, err := client.ListNodes(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, nodes)

	t.Logf("Found %d nodes:", len(nodes))
	for _, n := range nodes {
		t.Logf("  - %s: %s, %d CPUs, %d MB RAM", n.Name, n.State, n.CPUs, n.MemoryMB)
	}
}

func TestIntegration_SubmitAndCancelJob(t *testing.T) {
	config := getTestConfig(t)
	partition := getTestPartition(t)

	client, err := slurm.NewSSHSLURMClient(config, "test-cluster", partition)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect()

	// Submit a simple job
	spec := &slurm.SLURMJobSpec{
		JobName:     "virtengine-test-job",
		Partition:   partition,
		Nodes:       1,
		CPUsPerNode: 1,
		MemoryMB:    1024,
		TimeLimit:   5, // 5 minutes
		Command:     "sleep",
		Arguments:   []string{"60"},
	}

	jobID, err := client.SubmitJob(ctx, spec)
	require.NoError(t, err)
	assert.NotEmpty(t, jobID)
	t.Logf("Submitted job: %s", jobID)

	// Get job status
	job, err := client.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	assert.NotNil(t, job)
	t.Logf("Job state: %s", job.State)

	// Cancel job
	err = client.CancelJob(ctx, jobID)
	require.NoError(t, err)
	t.Logf("Job cancelled")

	// Verify cancellation
	time.Sleep(2 * time.Second)
	job, err = client.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, slurm.SLURMJobStateCancelled, job.State)
}

func TestIntegration_JobAccounting(t *testing.T) {
	config := getTestConfig(t)
	partition := getTestPartition(t)

	client, err := slurm.NewSSHSLURMClient(config, "test-cluster", partition)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect()

	// Submit a quick job
	spec := &slurm.SLURMJobSpec{
		JobName:     "virtengine-accounting-test",
		Partition:   partition,
		Nodes:       1,
		CPUsPerNode: 1,
		MemoryMB:    512,
		TimeLimit:   2,
		Command:     "echo",
		Arguments:   []string{"Hello from VirtEngine"},
	}

	jobID, err := client.SubmitJob(ctx, spec)
	require.NoError(t, err)
	t.Logf("Submitted job: %s", jobID)

	// Wait for job to complete
	var job *slurm.SLURMJob
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		job, err = client.GetJobStatus(ctx, jobID)
		require.NoError(t, err)

		if job.State == slurm.SLURMJobStateCompleted ||
			job.State == slurm.SLURMJobStateFailed ||
			job.State == slurm.SLURMJobStateCancelled {
			break
		}
		t.Logf("Job state: %s", job.State)
	}

	require.Equal(t, slurm.SLURMJobStateCompleted, job.State)

	// Get accounting data
	metrics, err := client.GetJobAccounting(ctx, jobID)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	t.Logf("Job accounting:")
	t.Logf("  Wall clock: %d seconds", metrics.WallClockSeconds)
	t.Logf("  CPU time: %d seconds", metrics.CPUTimeSeconds)
	t.Logf("  Max RSS: %d bytes", metrics.MaxRSSBytes)
}

func TestIntegration_ConnectionPooling(t *testing.T) {
	config := getTestConfig(t)
	config.PoolSize = 3
	partition := getTestPartition(t)

	client, err := slurm.NewSSHSLURMClient(config, "test-cluster", partition)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect()

	// Run multiple concurrent operations
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			_, err := client.ListPartitions(ctx)
			if err != nil {
				t.Errorf("Concurrent operation %d failed: %v", idx, err)
			}
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	t.Log("All concurrent operations completed successfully")
}

func TestIntegration_BatchScriptBuilder(t *testing.T) {
	config := getTestConfig(t)
	partition := getTestPartition(t)

	client, err := slurm.NewSSHSLURMClient(config, "test-cluster", partition)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect()

	// Build a complex script using BatchScriptBuilder
	script := slurm.NewBatchScriptBuilder().
		SetJobName("virtengine-builder-test").
		SetPartition(partition).
		SetNodes(1).
		SetCPUsPerTask(1).
		SetMemoryMB(512).
		SetTimeLimitMinutes(2).
		AddHeaderComment("VirtEngine Integration Test").
		SetEnv("TEST_VAR", "hello").
		AddSetupCommand("echo 'Starting test'").
		AddCommand("hostname").
		AddCommand("date").
		AddCommand("echo $TEST_VAR").
		AddCleanupCommand("echo 'Test complete'").
		Build()

	t.Logf("Generated script:\n%s", script)

	// Submit the script
	jobID, err := client.SubmitJobFromScript(ctx, script)
	require.NoError(t, err)
	t.Logf("Submitted job: %s", jobID)

	// Cancel to clean up
	defer client.CancelJob(ctx, jobID)

	// Check it was submitted
	job, err := client.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	assert.NotNil(t, job)
}
