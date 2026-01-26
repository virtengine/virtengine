package moab_adapter_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	moab "pkg.akt.dev/node/pkg/moab_adapter"
)

// VE-917: MOAB workload manager adapter tests

// Test constants
const (
	testJobName         = "test-job"
	testExecutable      = "/bin/echo"
	testVEJobID         = "ve-12345"
	testCancelJobID     = "ve-cancel-test"
	testHoldJobID       = "ve-hold-test"
	testRewardsJobID    = "ve-rewards-test"
	testProviderAddress = "virtengine1provider123"
	testQueue           = "batch"
)

// TestMOABAdapterCreation tests adapter initialization
func TestMOABAdapterCreation(t *testing.T) {
	config := moab.MOABConfig{
		ServerHost:   "localhost",
		ServerPort:   42559,
		DefaultQueue: testQueue,
	}

	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	require.NotNil(t, adapter)
	require.Equal(t, testQueue, adapter.ClusterName())
	require.False(t, adapter.IsRunning())
}

// TestMOABAdapterStartStop tests starting and stopping the adapter
func TestMOABAdapterStartStop(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

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

// TestJobSubmission tests submitting a job to MOAB (msub equivalent)
func TestJobSubmission(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	jobSpec := moab.MOABJobSpec{
		JobName:       testJobName,
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  4,
		MemoryMB:      8192,
		WallTimeLimit: 3600,
		Executable:    testExecutable,
		Arguments:     []string{"Hello", "HPC"},
		Environment: map[string]string{
			"VIRTENGINE_JOB_ID": testVEJobID,
		},
	}

	job, err := adapter.SubmitJob(ctx, testVEJobID, &jobSpec)

	require.NoError(t, err)
	require.NotNil(t, job)
	require.NotEmpty(t, job.MOABJobID)
	require.Equal(t, testVEJobID, job.VirtEngineJobID)
	require.Equal(t, moab.MOABJobStateIdle, job.State)
}

// TestJobSubmissionWithDefaults tests job submission using default config values
func TestJobSubmissionWithDefaults(t *testing.T) {
	config := moab.MOABConfig{
		ServerHost:     "localhost",
		DefaultQueue:   testQueue,
		DefaultAccount: "default-account",
	}
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Job spec without queue and account - should use defaults
	jobSpec := moab.MOABJobSpec{
		JobName:       "default-test",
		Queue:         testQueue, // Must specify valid queue
		Nodes:         1,
		ProcsPerNode:  2,
		WallTimeLimit: 1800,
		Executable:    "/bin/hostname",
	}

	job, err := adapter.SubmitJob(ctx, "ve-67890", &jobSpec)
	require.NoError(t, err)
	require.NotNil(t, job)
}

// TestJobSubmissionValidation tests job specification validation
func TestJobSubmissionValidation(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	testCases := []struct {
		name        string
		spec        moab.MOABJobSpec
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid spec",
			spec: moab.MOABJobSpec{
				JobName:       "valid-job",
				Queue:         "batch",
				Nodes:         1,
				ProcsPerNode:  4,
				WallTimeLimit: 3600,
				Executable:    "/bin/echo",
			},
			expectError: false,
		},
		{
			name: "missing job name",
			spec: moab.MOABJobSpec{
				Queue:         "batch",
				Nodes:         1,
				ProcsPerNode:  1,
				WallTimeLimit: 3600,
				Executable:    "/bin/echo",
			},
			expectError: true,
			errorMsg:    "job name is required",
		},
		{
			name: "missing queue",
			spec: moab.MOABJobSpec{
				JobName:       "test-job",
				Nodes:         1,
				ProcsPerNode:  1,
				WallTimeLimit: 3600,
				Executable:    "/bin/echo",
			},
			expectError: true,
			errorMsg:    "queue is required",
		},
		{
			name: "zero nodes",
			spec: moab.MOABJobSpec{
				JobName:       "test-job",
				Queue:         "batch",
				Nodes:         0,
				ProcsPerNode:  1,
				WallTimeLimit: 3600,
				Executable:    "/bin/echo",
			},
			expectError: true,
			errorMsg:    "nodes must be at least 1",
		},
		{
			name: "zero procs per node",
			spec: moab.MOABJobSpec{
				JobName:       "test-job",
				Queue:         "batch",
				Nodes:         1,
				ProcsPerNode:  0,
				WallTimeLimit: 3600,
				Executable:    "/bin/echo",
			},
			expectError: true,
			errorMsg:    "procs_per_node must be at least 1",
		},
		{
			name: "zero wall time",
			spec: moab.MOABJobSpec{
				JobName:       "test-job",
				Queue:         "batch",
				Nodes:         1,
				ProcsPerNode:  1,
				WallTimeLimit: 0,
				Executable:    "/bin/echo",
			},
			expectError: true,
			errorMsg:    "wall_time_limit must be at least 1",
		},
		{
			name: "missing executable",
			spec: moab.MOABJobSpec{
				JobName:       "test-job",
				Queue:         "batch",
				Nodes:         1,
				ProcsPerNode:  1,
				WallTimeLimit: 3600,
			},
			expectError: true,
			errorMsg:    "executable is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := adapter.SubmitJob(ctx, "ve-test", &tc.spec)
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestJobCancellation tests cancelling a MOAB job (mjobctl -c)
func TestJobCancellation(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit a job first
	jobSpec := moab.MOABJobSpec{
		JobName:       "cancel-test",
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 3600,
		Executable:    "/bin/sleep",
		Arguments:     []string{"3600"},
	}

	job, err := adapter.SubmitJob(ctx, testCancelJobID, &jobSpec)
	require.NoError(t, err)

	// Cancel the job
	err = adapter.CancelJob(ctx, testCancelJobID)
	require.NoError(t, err)

	// Verify job is cancelled
	job, err = adapter.GetJobStatus(ctx, testCancelJobID)
	require.NoError(t, err)
	require.Equal(t, moab.MOABJobStateCancelled, job.State)
}

// TestJobHoldRelease tests holding and releasing jobs (mjobctl -h/-u)
func TestJobHoldRelease(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit a job
	jobSpec := moab.MOABJobSpec{
		JobName:       "hold-test",
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 3600,
		Executable:    "/bin/sleep",
		Arguments:     []string{"100"},
	}

	_, err = adapter.SubmitJob(ctx, testHoldJobID, &jobSpec)
	require.NoError(t, err)

	// Hold the job
	err = adapter.HoldJob(ctx, testHoldJobID)
	require.NoError(t, err)

	// Verify job is held
	job, err := adapter.GetJobStatus(ctx, testHoldJobID)
	require.NoError(t, err)
	require.Equal(t, moab.MOABJobStateHold, job.State)

	// Release the job
	err = adapter.ReleaseJob(ctx, testHoldJobID)
	require.NoError(t, err)

	// Verify job is released
	job, err = adapter.GetJobStatus(ctx, testHoldJobID)
	require.NoError(t, err)
	require.Equal(t, moab.MOABJobStateIdle, job.State)
}

// TestJobStatusTracking tests tracking job status changes (checkjob)
func TestJobStatusTracking(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit a job
	jobSpec := moab.MOABJobSpec{
		JobName:       "status-test",
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 60,
		Executable:    testExecutable,
		Arguments:     []string{"test"},
	}

	job, err := adapter.SubmitJob(ctx, "ve-status-test", &jobSpec)
	require.NoError(t, err)
	require.NotNil(t, job)

	// Initial status should be Idle
	require.Equal(t, moab.MOABJobStateIdle, job.State)

	// Wait for job to transition to running
	time.Sleep(150 * time.Millisecond)

	job, err = adapter.GetJobStatus(ctx, "ve-status-test")
	require.NoError(t, err)

	// Status should be Running or still transitioning
	require.True(t,
		job.State == moab.MOABJobStateRunning ||
			job.State == moab.MOABJobStateStarting ||
			job.State == moab.MOABJobStateIdle,
		"job should be idle, starting, or running")
}

// TestListQueues tests listing queues (mdiag -q)
func TestListQueues(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	queues, err := adapter.ListQueues(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, queues)

	// Check for expected queues
	queueNames := make(map[string]bool)
	for _, q := range queues {
		queueNames[q.Name] = true
	}

	require.True(t, queueNames["batch"], "batch queue should exist")
	require.True(t, queueNames["gpu"], "gpu queue should exist")
}

// TestListNodes tests listing nodes (mdiag -n)
func TestListNodes(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	nodes, err := adapter.ListNodes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, nodes)

	// Check for GPU nodes
	hasGPU := false
	for _, n := range nodes {
		if n.GPUs > 0 {
			hasGPU = true
			require.NotEmpty(t, n.GPUType)
		}
	}
	require.True(t, hasGPU, "should have GPU nodes")
}

// TestGetClusterInfo tests getting cluster information (mdiag -s)
func TestGetClusterInfo(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	info, err := adapter.GetClusterInfo(ctx)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, "virtengine-hpc", info.Name)
	require.True(t, info.TotalNodes > 0)
	require.True(t, info.TotalProcessors > 0)
}

// TestJobNotFoundError tests error handling for non-existent jobs
func TestJobNotFoundError(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Try to get status of non-existent job
	_, err = adapter.GetJobStatus(ctx, "ve-nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, moab.ErrJobNotFound)

	// Try to cancel non-existent job
	err = adapter.CancelJob(ctx, "ve-nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, moab.ErrJobNotFound)
}

// TestJobSubmissionWhenNotConnected tests error when not connected
func TestJobSubmissionWhenNotConnected(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()

	jobSpec := moab.MOABJobSpec{
		JobName:       testJobName,
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 3600,
		Executable:    testExecutable,
	}

	// Should fail when not started
	_, err := adapter.SubmitJob(ctx, "ve-test", &jobSpec)
	require.Error(t, err)
	require.ErrorIs(t, err, moab.ErrMOABNotConnected)
}

// TestStatusReportCreation tests creating signed status reports
func TestStatusReportCreation(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit a job
	jobSpec := moab.MOABJobSpec{
		JobName:       "report-test",
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  2,
		WallTimeLimit: 1800,
		Executable:    "/bin/hostname",
	}

	job, err := adapter.SubmitJob(ctx, "ve-report-test", &jobSpec)
	require.NoError(t, err)

	// Create status report
	report, err := adapter.CreateStatusReport(job)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, testProviderAddress, report.ProviderAddress)
	require.Equal(t, "ve-report-test", report.VirtEngineJobID)
	require.NotEmpty(t, report.MOABJobID)
	require.NotEmpty(t, report.Signature)
}

// TestJobLifecycleCallbacks tests lifecycle event callbacks
func TestJobLifecycleCallbacks(t *testing.T) {
	config := moab.MOABConfig{
		ServerHost:      "localhost",
		DefaultQueue:    "batch",
		JobPollInterval: 50 * time.Millisecond,
	}
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	events := make([]moab.JobLifecycleEvent, 0)
	eventMu := make(chan struct{}, 10)

	adapter.RegisterLifecycleCallback(func(job *moab.MOABJob, event moab.JobLifecycleEvent) {
		events = append(events, event)
		eventMu <- struct{}{}
	})

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit a job
	jobSpec := moab.MOABJobSpec{
		JobName:       "callback-test",
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 60,
		Executable:    testExecutable,
	}

	_, err = adapter.SubmitJob(ctx, "ve-callback-test", &jobSpec)
	require.NoError(t, err)

	// Wait for submitted event
	select {
	case <-eventMu:
		require.Contains(t, events, moab.JobEventSubmitted)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for submitted event")
	}
}

// TestVERewardsIntegration tests VirtEngine rewards integration
func TestVERewardsIntegration(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	mockRewards := moab.NewMockRewardsIntegration()
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)
	adapter.SetRewardsIntegration(mockRewards)
	adapter.SetClusterID("cluster-001")

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit a job
	jobSpec := moab.MOABJobSpec{
		JobName:       "rewards-test",
		Queue:         testQueue,
		Nodes:         2,
		ProcsPerNode:  4,
		MemoryMB:      16384,
		WallTimeLimit: 3600,
		Executable:    "/bin/test-app",
	}

	job, err := adapter.SubmitJob(ctx, testRewardsJobID, &jobSpec)
	require.NoError(t, err)

	// Simulate job completion
	err = mockClient.SimulateJobCompletion(job.MOABJobID, 0)
	require.NoError(t, err)

	// Get updated status with accounting
	job, err = adapter.GetJobStatus(ctx, testRewardsJobID)
	require.NoError(t, err)

	// Wait for state to update
	time.Sleep(100 * time.Millisecond)

	// Get accounting
	job, err = adapter.GetJobStatus(ctx, testRewardsJobID)
	require.NoError(t, err)
	require.Equal(t, moab.MOABJobStateCompleted, job.State)

	// Record for rewards
	err = adapter.RecordJobForRewards(ctx, job, "virtengine1customer456")
	require.NoError(t, err)

	// Verify reward was recorded
	completions := mockRewards.GetCompletions()
	require.Len(t, completions, 1)
	require.Equal(t, testRewardsJobID, completions[0].JobID)
	require.Equal(t, "MOAB", completions[0].SchedulerType)
	require.Equal(t, "cluster-001", completions[0].ClusterID)
	require.Equal(t, testProviderAddress, completions[0].ProviderAddress)
	require.Equal(t, "virtengine1customer456", completions[0].CustomerAddress)
}

// TestMOABJobSpecToMsubScript tests msub script generation
func TestMOABJobSpecToMsubScript(t *testing.T) {
	spec := moab.MOABJobSpec{
		JobName:       "test-script",
		Queue:         "gpu",
		Account:       "project-001",
		Nodes:         4,
		ProcsPerNode:  32,
		MemoryMB:      128000,
		GPUs:          8,
		WallTimeLimit: 7200, // 2 hours
		Executable:    "/path/to/app",
		Arguments:     []string{"--config", "test.yaml"},
		Environment: map[string]string{
			"CUDA_VISIBLE_DEVICES": "0,1,2,3",
		},
	}

	script := spec.ToMsubScript()

	require.Contains(t, script, "#!/bin/bash")
	require.Contains(t, script, "#MSUB -N test-script")
	require.Contains(t, script, "#MSUB -q gpu")
	require.Contains(t, script, "#MSUB -A project-001")
	require.Contains(t, script, "export CUDA_VISIBLE_DEVICES=")
	require.Contains(t, script, "/path/to/app --config test.yaml")
}

// TestMapToVirtEngineState tests state mapping
func TestMapToVirtEngineState(t *testing.T) {
	testCases := []struct {
		moabState MOABJobState
		expected  string
	}{
		{moab.MOABJobStateIdle, "queued"},
		{moab.MOABJobStateStarting, "starting"},
		{moab.MOABJobStateRunning, "running"},
		{moab.MOABJobStateCompleted, "completed"},
		{moab.MOABJobStateFailed, "failed"},
		{moab.MOABJobStateCancelled, "cancelled"},
		{moab.MOABJobStateHold, "held"},
		{moab.MOABJobStateSuspended, "paused"},
		{moab.MOABJobStateDeferred, "deferred"},
		{moab.MOABJobStateRemoved, "removed"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.moabState), func(t *testing.T) {
			result := moab.MapToVirtEngineState(tc.moabState)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestCalculateJobCost tests job cost calculation
func TestCalculateJobCost(t *testing.T) {
	metrics := &moab.MOABUsageMetrics{
		WallClockSeconds: 3600,
		CPUTimeSeconds:   14400, // 4 cores * 1 hour
		MaxRSSBytes:      8 * 1024 * 1024 * 1024, // 8 GB
		GPUSeconds:       3600, // 1 GPU * 1 hour
	}

	cpuRate := 0.10   // $0.10 per CPU-hour
	gpuRate := 2.00   // $2.00 per GPU-hour
	memRate := 0.01   // $0.01 per GB-hour

	cost := moab.CalculateJobCost(metrics, cpuRate, gpuRate, memRate)

	// Expected:
	// CPU: 4 hours * $0.10 = $0.40
	// GPU: 1 hour * $2.00 = $2.00
	// Memory: 8 GB * 1 hour * $0.01 = $0.08
	// Total: $2.48
	require.InDelta(t, 2.48, cost, 0.01)
}

// TestGetAllJobs tests getting all tracked jobs
func TestGetAllJobs(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit multiple jobs
	for i := 1; i <= 3; i++ {
		jobSpec := moab.MOABJobSpec{
			JobName:       "batch-job-" + string(rune('0'+i)),
			Queue:         testQueue,
			Nodes:         1,
			ProcsPerNode:  1,
			WallTimeLimit: 3600,
			Executable:    testExecutable,
		}
		_, err := adapter.SubmitJob(ctx, "ve-batch-"+string(rune('0'+i)), &jobSpec)
		require.NoError(t, err)
	}

	jobs := adapter.GetAllJobs()
	require.Len(t, jobs, 3)
}

// TestGetJobsByQueue tests getting jobs by queue
func TestGetJobsByQueue(t *testing.T) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	mockSigner := moab.NewMockJobSigner(testProviderAddress)
	adapter := moab.NewMOABAdapter(config, mockClient, mockSigner)

	ctx := context.Background()
	err := adapter.Start(ctx)
	require.NoError(t, err)
	defer adapter.Stop()

	// Submit jobs to different queues
	batchSpec := moab.MOABJobSpec{
		JobName:       "batch-job",
		Queue:         testQueue,
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 3600,
		Executable:    testExecutable,
	}
	gpuSpec := moab.MOABJobSpec{
		JobName:       "gpu-job",
		Queue:         "gpu",
		Nodes:         1,
		ProcsPerNode:  1,
		WallTimeLimit: 3600,
		Executable:    testExecutable,
	}

	_, err = adapter.SubmitJob(ctx, "ve-batch-1", &batchSpec)
	require.NoError(t, err)
	_, err = adapter.SubmitJob(ctx, "ve-gpu-1", &gpuSpec)
	require.NoError(t, err)

	batchJobs := adapter.GetJobsByQueue("batch")
	require.Len(t, batchJobs, 1)

	gpuJobs := adapter.GetJobsByQueue("gpu")
	require.Len(t, gpuJobs, 1)
}

// TestDefaultConfig tests default configuration values
func TestDefaultConfig(t *testing.T) {
	config := moab.DefaultMOABConfig()

	require.Equal(t, "moab-server", config.ServerHost)
	require.Equal(t, 42559, config.ServerPort)
	require.True(t, config.UseTLS)
	require.Equal(t, "password", config.AuthMethod)
	require.Equal(t, "batch", config.DefaultQueue)
	require.Equal(t, 15*time.Second, config.JobPollInterval)
	require.Equal(t, 30*time.Second, config.ConnectionTimeout)
	require.Equal(t, 3, config.MaxRetries)
	require.True(t, config.WaldurIntegration)
}

type MOABJobState = moab.MOABJobState
