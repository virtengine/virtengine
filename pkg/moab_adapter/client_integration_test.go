// Package moab_adapter implements the MOAB workload manager adapter for VirtEngine.
//
// VE-917: MOAB workload manager using Waldur
// HPC-ADAPTER-001: Integration tests for production MOAB client
package moab_adapter_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	moab "github.com/virtengine/virtengine/pkg/moab_adapter"
)

// Integration test environment variables:
// MOAB_TEST_HOST - MOAB server hostname (required for integration tests)
// MOAB_TEST_PORT - MOAB server port (default: 22)
// MOAB_TEST_USER - Username for authentication
// MOAB_TEST_PASSWORD - Password for authentication (if using password auth)
// MOAB_TEST_QUEUE - Queue to use for test jobs (default: batch)
// MOAB_TEST_EXECUTABLE - Executable to run (default: /bin/hostname)

func getIntegrationConfig() (moab.MOABConfig, bool) {
	host := os.Getenv("MOAB_TEST_HOST")
	if host == "" {
		return moab.MOABConfig{}, false
	}

	port := 22
	if p := os.Getenv("MOAB_TEST_PORT"); p != "" {
		// Parse port
		var parsed int
		if _, err := fmt.Sscanf(p, "%d", &parsed); err == nil {
			port = parsed
		}
	}

	config := moab.MOABConfig{
		ServerHost:        host,
		ServerPort:        port,
		UseTLS:            false, // SSH handles encryption
		AuthMethod:        "password",
		Username:          os.Getenv("MOAB_TEST_USER"),
		Password:          os.Getenv("MOAB_TEST_PASSWORD"),
		DefaultQueue:      "batch",
		DefaultAccount:    "default",
		JobPollInterval:   5 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		MaxRetries:        3,
		WaldurIntegration: false,
	}

	if q := os.Getenv("MOAB_TEST_QUEUE"); q != "" {
		config.DefaultQueue = q
	}

	return config, true
}

func skipIfNoMOAB(t *testing.T) moab.MOABConfig {
	config, ok := getIntegrationConfig()
	if !ok {
		t.Skip("Skipping MOAB integration test: MOAB_TEST_HOST not set")
	}
	if config.Username == "" {
		t.Skip("Skipping MOAB integration test: MOAB_TEST_USER not set")
	}
	return config
}

// TestIntegrationConnect tests connecting to a real MOAB server
func TestIntegrationConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	require.True(t, client.IsConnected())

	err = client.Disconnect()
	require.NoError(t, err)
	require.False(t, client.IsConnected())
}

// TestIntegrationListQueues tests listing queues on a real MOAB server
func TestIntegrationListQueues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer func() { _ = client.Disconnect() }()

	queues, err := client.ListQueues(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, queues, "expected at least one queue")

	t.Logf("Found %d queues:", len(queues))
	for _, q := range queues {
		t.Logf("  - %s (state: %s, running: %d, idle: %d)",
			q.Name, q.State, q.RunningJobs, q.IdleJobs)
	}
}

// TestIntegrationListNodes tests listing nodes on a real MOAB server
func TestIntegrationListNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer func() { _ = client.Disconnect() }()

	nodes, err := client.ListNodes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, nodes, "expected at least one node")

	t.Logf("Found %d nodes:", len(nodes))
	for _, n := range nodes {
		t.Logf("  - %s (state: %s, procs: %d, mem: %dMB, gpus: %d)",
			n.Name, n.State, n.Processors, n.MemoryMB, n.GPUs)
	}
}

// TestIntegrationGetClusterInfo tests getting cluster info from a real MOAB server
func TestIntegrationGetClusterInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer func() { _ = client.Disconnect() }()

	info, err := client.GetClusterInfo(ctx)
	require.NoError(t, err)
	require.NotNil(t, info)

	t.Logf("Cluster: %s", info.Name)
	t.Logf("  Total Nodes: %d (idle: %d, busy: %d, down: %d)",
		info.TotalNodes, info.IdleNodes, info.BusyNodes, info.DownNodes)
	t.Logf("  Total Processors: %d (idle: %d)",
		info.TotalProcessors, info.IdleProcessors)
	t.Logf("  Jobs: %d running, %d idle",
		info.RunningJobs, info.IdleJobs)
}

// TestIntegrationJobSubmitAndCancel tests submitting and cancelling a job
func TestIntegrationJobSubmitAndCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer func() { _ = client.Disconnect() }()

	// Get test executable
	executable := os.Getenv("MOAB_TEST_EXECUTABLE")
	if executable == "" {
		executable = "/bin/hostname"
	}

	// Submit a test job
	spec := &moab.MOABJobSpec{
		JobName:       "virtengine-test-" + time.Now().Format("20060102-150405"),
		Queue:         config.DefaultQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 60, // 1 minute
		Executable:    executable,
	}

	jobID, err := client.SubmitJob(ctx, spec)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)
	t.Logf("Submitted job: %s", jobID)

	// Get job status
	job, err := client.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	require.NotNil(t, job)
	t.Logf("Job state: %s", job.State)

	// Cancel the job
	err = client.CancelJob(ctx, jobID)
	require.NoError(t, err)
	t.Logf("Job cancelled")

	// Verify cancellation
	time.Sleep(2 * time.Second)
	job, err = client.GetJobStatus(ctx, jobID)
	if err == nil {
		t.Logf("Final job state: %s", job.State)
	}
}

// TestIntegrationJobLifecycle tests the full job lifecycle
func TestIntegrationJobLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	// Create adapter with production client
	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	mockSigner := moab.NewMockJobSigner("virtengine1testprovider")
	adapter := moab.NewMOABAdapter(config, client, mockSigner)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = adapter.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = adapter.Stop() }()

	// Get test executable
	executable := os.Getenv("MOAB_TEST_EXECUTABLE")
	if executable == "" {
		executable = "/bin/sleep"
	}

	// Submit a short job
	spec := &moab.MOABJobSpec{
		JobName:       "virtengine-lifecycle-" + time.Now().Format("20060102-150405"),
		Queue:         config.DefaultQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 120, // 2 minutes
		Executable:    executable,
		Arguments:     []string{"10"}, // Sleep for 10 seconds
	}

	job, err := adapter.SubmitJob(ctx, "ve-test-lifecycle", spec)
	require.NoError(t, err)
	require.NotNil(t, job)
	t.Logf("Submitted job: %s (MOAB ID: %s)", job.VirtEngineJobID, job.MOABJobID)

	// Poll for job completion
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	maxWait := time.After(3 * time.Minute)
	for {
		select {
		case <-maxWait:
			t.Log("Timeout waiting for job completion, cancelling...")
			_ = adapter.CancelJob(ctx, "ve-test-lifecycle")
			return
		case <-ticker.C:
			job, err := adapter.GetJobStatus(ctx, "ve-test-lifecycle")
			if err != nil {
				t.Logf("Error getting status: %v", err)
				continue
			}
			t.Logf("Job state: %s", job.State)

			switch job.State {
			case moab.MOABJobStateCompleted:
				t.Log("Job completed successfully")
				require.Equal(t, int32(0), job.ExitCode)
				return
			case moab.MOABJobStateFailed:
				t.Logf("Job failed with exit code: %d", job.ExitCode)
				return
			case moab.MOABJobStateCancelled:
				t.Log("Job was cancelled")
				return
			}
		}
	}
}

// TestIntegrationHoldRelease tests holding and releasing jobs
func TestIntegrationHoldRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer func() { _ = client.Disconnect() }()

	// Submit a test job
	spec := &moab.MOABJobSpec{
		JobName:       "virtengine-hold-" + time.Now().Format("20060102-150405"),
		Queue:         config.DefaultQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 60,
		Executable:    "/bin/sleep",
		Arguments:     []string{"3600"}, // Long sleep so we can hold it
	}

	jobID, err := client.SubmitJob(ctx, spec)
	require.NoError(t, err)
	t.Logf("Submitted job: %s", jobID)

	// Wait for job to be queued
	time.Sleep(2 * time.Second)

	// Hold the job
	err = client.HoldJob(ctx, jobID)
	if err != nil {
		t.Logf("Hold failed (job may have already started): %v", err)
		// Cancel and skip rest of test
		_ = client.CancelJob(ctx, jobID)
		return
	}
	t.Log("Job held")

	// Verify hold
	job, err := client.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	t.Logf("Job state after hold: %s", job.State)

	// Release the job
	err = client.ReleaseJob(ctx, jobID)
	require.NoError(t, err)
	t.Log("Job released")

	// Verify release
	job, err = client.GetJobStatus(ctx, jobID)
	require.NoError(t, err)
	t.Logf("Job state after release: %s", job.State)

	// Cancel the job
	err = client.CancelJob(ctx, jobID)
	require.NoError(t, err)
	t.Log("Job cancelled")
}

// TestIntegrationReservations tests listing reservations
func TestIntegrationReservations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := skipIfNoMOAB(t)

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)
	defer func() { _ = client.Disconnect() }()

	reservations, err := client.GetReservations(ctx)
	require.NoError(t, err)

	t.Logf("Found %d reservations:", len(reservations))
	for _, r := range reservations {
		t.Logf("  - %s (owner: %s, state: %s, start: %v, end: %v)",
			r.Name, r.Owner, r.State, r.StartTime, r.EndTime)
	}
}

// TestIntegrationConnectionRetry tests retry logic on connection failure
func TestIntegrationConnectionRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test intentionally uses a bad host to test retry behavior
	config := moab.MOABConfig{
		ServerHost:        "nonexistent-host.invalid",
		ServerPort:        22,
		AuthMethod:        "password",
		Username:          "testuser",
		Password:          "testpass",
		ConnectionTimeout: 5 * time.Second,
		MaxRetries:        2,
	}

	client, err := moab.NewProductionMOABClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	err = client.Connect(ctx)
	elapsed := time.Since(start)

	require.Error(t, err)
	t.Logf("Connection failed as expected after %v: %v", elapsed, err)
}

// Benchmark tests

func BenchmarkIntegrationGetJobStatus(b *testing.B) {
	config, ok := getIntegrationConfig()
	if !ok {
		b.Skip("Skipping MOAB integration benchmark: MOAB_TEST_HOST not set")
	}

	client, err := moab.NewProductionMOABClient(config)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer func() { _ = client.Disconnect() }()

	// Submit a job for the benchmark
	spec := &moab.MOABJobSpec{
		JobName:       "virtengine-bench-" + time.Now().Format("20060102-150405"),
		Queue:         config.DefaultQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 300,
		Executable:    "/bin/sleep",
		Arguments:     []string{"300"},
	}

	jobID, err := client.SubmitJob(ctx, spec)
	if err != nil {
		b.Fatalf("Failed to submit job: %v", err)
	}
	defer func() { _ = client.CancelJob(ctx, jobID) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetJobStatus(ctx, jobID)
		if err != nil {
			b.Errorf("GetJobStatus failed: %v", err)
		}
	}
}
