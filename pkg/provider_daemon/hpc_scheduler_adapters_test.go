// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: HPC scheduler adapter wrapper tests
package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	moab "github.com/virtengine/virtengine/pkg/moab_adapter"
	ood "github.com/virtengine/virtengine/pkg/ood_adapter"
	slurm "github.com/virtengine/virtengine/pkg/slurm_adapter"
)

const (
	schedulerTestProviderAddress = "ve1provider123abc456def"
	testJupyterCommand           = "jupyter"
	testMOABQueue                = "batch"
)

type testSigner struct {
	address string
}

func (s *testSigner) Sign(data []byte) ([]byte, error) {
	return append([]byte("sig-"), data...), nil
}

func (s *testSigner) Verify(data []byte, signature []byte) bool {
	return true
}

func (s *testSigner) GetProviderAddress() string {
	return s.address
}

var (
	_ HPCScheduler = (*SLURMSchedulerWrapper)(nil)
	_ HPCScheduler = (*MOABSchedulerWrapper)(nil)
	_ HPCScheduler = (*OODSchedulerWrapper)(nil)
)

func newSLURMWrapper() (*SLURMSchedulerWrapper, *slurm.SLURMAdapter, *slurm.MockSLURMClient) {
	config := slurm.DefaultSLURMConfig()
	mockClient := slurm.NewMockSLURMClient()
	signer := &testSigner{address: schedulerTestProviderAddress}
	adapter := slurm.NewSLURMAdapter(config, mockClient, signer)
	wrapper := NewSLURMSchedulerWrapper(adapter, signer, testClusterID)
	return wrapper, adapter, mockClient
}

func newMOABWrapper() (*MOABSchedulerWrapper, *moab.MOABAdapter, *moab.MockMOABClient) {
	config := moab.DefaultMOABConfig()
	mockClient := moab.NewMockMOABClient()
	signer := &testSigner{address: schedulerTestProviderAddress}
	adapter := moab.NewMOABAdapter(config, mockClient, signer)
	wrapper := NewMOABSchedulerWrapper(adapter, signer, testClusterID)
	return wrapper, adapter, mockClient
}

func newOODWrapper() (*OODSchedulerWrapper, *ood.MockOODClient) {
	config := ood.DefaultOODConfig()
	mockClient := ood.NewMockOODClient()
	mockAuth := ood.NewMockVEIDAuthProvider()
	mockSigner := ood.NewMockSessionSigner(schedulerTestProviderAddress)
	adapter := ood.NewOODAdapter(config, mockClient, mockAuth, mockSigner)
	wrapper := NewOODSchedulerWrapper(adapter, &testSigner{address: schedulerTestProviderAddress}, testClusterID)
	return wrapper, mockClient
}

func TestSLURMSchedulerWrapper_SubmitStatusAndSpec(t *testing.T) {
	wrapper, adapter, client := newSLURMWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	job := createTestJob("slurm-job-1")
	hpcJob, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)
	require.Equal(t, HPCSchedulerTypeSLURM, hpcJob.SchedulerType)
	require.Equal(t, HPCJobStateQueued, hpcJob.State)
	require.NotEmpty(t, hpcJob.SchedulerJobID)

	slurmJob, err := adapter.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("/scratch/virtengine/%s/%s", testClusterID, job.JobID), slurmJob.Spec.OutputDirectory)
	require.Equal(t, job.JobID, slurmJob.Spec.Environment["VIRTENGINE_JOB_ID"])

	require.NoError(t, client.SimulateJobCompletion(hpcJob.SchedulerJobID, true, 0))
	status, err := wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateCompleted, status.State)
	require.Equal(t, int32(0), status.ExitCode)
}

func TestSLURMSchedulerWrapper_CancelAndErrors(t *testing.T) {
	wrapper, _, _ := newSLURMWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	job := createTestJob("slurm-cancel-1")
	_, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)

	require.NoError(t, wrapper.CancelJob(ctx, job.JobID))
	status, err := wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateCancelled, status.State)
	require.NotNil(t, status.EndTime)

	err = wrapper.CancelJob(ctx, "missing-job")
	require.Error(t, err)
	var schedErr *HPCSchedulerError
	require.ErrorAs(t, err, &schedErr)
	require.Equal(t, HPCErrorCodeJobCancellationFailed, schedErr.Code)
	require.ErrorIs(t, err, slurm.ErrJobNotFound)

	notRunningWrapper, _, _ := newSLURMWrapper()
	_, err = notRunningWrapper.SubmitJob(ctx, createTestJob("slurm-not-running"))
	require.Error(t, err)
	require.ErrorAs(t, err, &schedErr)
	require.Equal(t, HPCErrorCodeJobSubmissionFailed, schedErr.Code)
	require.ErrorIs(t, err, slurm.ErrSLURMNotConnected)
}

func TestSLURMSchedulerWrapper_LifecycleEvents(t *testing.T) {
	wrapper, _, client := newSLURMWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	var events []HPCJobLifecycleEvent
	wrapper.RegisterLifecycleCallback(func(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
		events = append(events, event)
	})

	job := createTestJob("slurm-events-1")
	hpcJob, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)
	_, err = wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)

	require.NoError(t, client.SimulateJobCompletion(hpcJob.SchedulerJobID, true, 0))
	_, err = wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)

	require.Equal(t, []HPCJobLifecycleEvent{
		HPCJobEventSubmitted,
		HPCJobEventStarted,
		HPCJobEventCompleted,
	}, events)
}

func TestSLURMSchedulerWrapper_ConcurrentSubmissions(t *testing.T) {
	wrapper, _, _ := newSLURMWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	errCh := make(chan error, 5)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			job := createTestJob(fmt.Sprintf("slurm-parallel-%d", idx))
			_, err := wrapper.SubmitJob(ctx, job)
			errCh <- err
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	active, err := wrapper.ListActiveJobs(ctx)
	require.NoError(t, err)
	require.Len(t, active, 5)
}

func TestOODSchedulerWrapper_LifecycleAndAccounting(t *testing.T) {
	wrapper, client := newOODWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	job := createTestJob("ood-job-1")
	job.WorkloadSpec.Command = testJupyterCommand
	hpcJob, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)
	require.Equal(t, HPCJobStatePending, hpcJob.State)

	client.SetSessionState(hpcJob.SchedulerJobID, ood.SessionStateRunning)
	status, err := wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateRunning, status.State)
	require.NotNil(t, status.StartTime)

	time.Sleep(2100 * time.Millisecond)
	client.SetSessionState(hpcJob.SchedulerJobID, ood.SessionStateCompleted)
	status, err = wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateCompleted, status.State)
	require.NotNil(t, status.EndTime)

	metrics, err := wrapper.GetJobAccounting(ctx, job.JobID)
	require.NoError(t, err)
	require.NotNil(t, metrics)
	require.GreaterOrEqual(t, metrics.WallClockSeconds, int64(1))
}

func TestOODSchedulerWrapper_CancelAndErrors(t *testing.T) {
	wrapper, client := newOODWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	job := createTestJob("ood-cancel-1")
	job.WorkloadSpec.Command = testJupyterCommand
	_, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)

	require.NoError(t, wrapper.CancelJob(ctx, job.JobID))
	status, err := wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateCancelled, status.State)

	err = wrapper.CancelJob(ctx, "missing-session")
	require.Error(t, err)
	var schedErr *HPCSchedulerError
	require.ErrorAs(t, err, &schedErr)
	require.Equal(t, HPCErrorCodeJobNotFound, schedErr.Code)

	client.SetFailLaunch(true)
	failJob := createTestJob("ood-launch-fail")
	failJob.WorkloadSpec.Command = testJupyterCommand
	_, err = wrapper.SubmitJob(ctx, failJob)
	require.Error(t, err)
	require.ErrorAs(t, err, &schedErr)
	require.Equal(t, HPCErrorCodeJobSubmissionFailed, schedErr.Code)
	require.ErrorIs(t, err, ood.ErrSessionCreationFailed)
	client.SetFailLaunch(false)
}

func TestOODSchedulerWrapper_StatusOnAdapterError(t *testing.T) {
	wrapper, _ := newOODWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))

	job := createTestJob("ood-status-error")
	job.WorkloadSpec.Command = testJupyterCommand
	statusJob, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)

	require.NoError(t, wrapper.Stop())
	status, err := wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, statusJob.State, status.State)
}

func TestMOABSchedulerWrapper_SubmitStatusAndMetrics(t *testing.T) {
	wrapper, adapter, client := newMOABWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	job := createTestJob("moab-job-1")
	job.QueueName = testMOABQueue
	hpcJob, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateQueued, hpcJob.State)
	require.NotEmpty(t, hpcJob.SchedulerJobID)

	moabJob, err := adapter.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	expectedOutput := fmt.Sprintf("/scratch/virtengine/%s/%s/%s.stdout", testClusterID, job.JobID, job.JobID)
	expectedError := fmt.Sprintf("/scratch/virtengine/%s/%s/%s.stderr", testClusterID, job.JobID, job.JobID)
	require.Equal(t, expectedOutput, moabJob.Spec.OutputFile)
	require.Equal(t, expectedError, moabJob.Spec.ErrorFile)

	require.NoError(t, client.SimulateJobCompletion(hpcJob.SchedulerJobID, 0))
	status, err := wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateCompleted, status.State)
	require.NotNil(t, status.Metrics)

	metrics, err := wrapper.GetJobAccounting(ctx, job.JobID)
	require.NoError(t, err)
	require.NotNil(t, metrics)
}

func TestMOABSchedulerWrapper_CancelAndErrors(t *testing.T) {
	wrapper, _, client := newMOABWrapper()
	ctx := context.Background()

	require.NoError(t, wrapper.Start(ctx))
	defer func() { _ = wrapper.Stop() }()

	job := createTestJob("moab-cancel-1")
	job.QueueName = testMOABQueue
	_, err := wrapper.SubmitJob(ctx, job)
	require.NoError(t, err)

	require.NoError(t, wrapper.CancelJob(ctx, job.JobID))
	status, err := wrapper.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, HPCJobStateCancelled, status.State)

	client.SetCancelError(errors.New("permission denied"))
	err = wrapper.CancelJob(ctx, job.JobID)
	require.Error(t, err)
	var schedErr *HPCSchedulerError
	require.ErrorAs(t, err, &schedErr)
	require.Equal(t, HPCErrorCodeJobCancellationFailed, schedErr.Code)

	err = wrapper.CancelJob(ctx, "missing-job")
	require.Error(t, err)
	require.ErrorAs(t, err, &schedErr)
	require.Equal(t, HPCErrorCodeJobCancellationFailed, schedErr.Code)
	require.ErrorIs(t, err, moab.ErrJobNotFound)

	notRunningWrapper, _, _ := newMOABWrapper()
	job.QueueName = testMOABQueue
	_, err = notRunningWrapper.SubmitJob(ctx, job)
	require.Error(t, err)
	require.ErrorAs(t, err, &schedErr)
	require.Equal(t, HPCErrorCodeJobSubmissionFailed, schedErr.Code)
	require.ErrorIs(t, err, moab.ErrMOABNotConnected)
}
