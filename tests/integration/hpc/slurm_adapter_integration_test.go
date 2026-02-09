package hpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	slurm "github.com/virtengine/virtengine/pkg/slurm_adapter"
)

type integrationJobSigner struct {
	providerAddress string
}

func (s *integrationJobSigner) Sign(data []byte) ([]byte, error) {
	return append([]byte("sig-"), data...), nil
}

func (s *integrationJobSigner) Verify(_ []byte, _ []byte) bool {
	return true
}

func (s *integrationJobSigner) GetProviderAddress() string {
	return s.providerAddress
}

func TestJobWithSLURMAdapterIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := slurm.NewMockSLURMClient()
	signer := &integrationJobSigner{providerAddress: "ve1provider-integration"}
	adapter := slurm.NewSLURMAdapter(slurm.DefaultSLURMConfig(), client, signer)

	require.NoError(t, adapter.Start(ctx))
	defer func() { _ = adapter.Stop() }()

	spec := &slurm.SLURMJobSpec{
		JobName:     "integration-job",
		Partition:   "default",
		Nodes:       1,
		CPUsPerNode: 2,
		MemoryMB:    1024,
		TimeLimit:   30,
		Command:     "/bin/echo",
		Environment: map[string]string{
			"VIRTENGINE_JOB_ID": "ve-job-integration",
		},
	}

	job, err := adapter.SubmitJob(ctx, "ve-job-integration", spec)
	require.NoError(t, err)
	require.NotEmpty(t, job.SLURMJobID)

	time.Sleep(200 * time.Millisecond)

	status, err := adapter.GetJobStatus(ctx, "ve-job-integration")
	require.NoError(t, err)
	require.NotNil(t, status)
	require.NotEqual(t, slurm.SLURMJobStatePending, status.State)

	require.NoError(t, client.SimulateJobCompletion(job.SLURMJobID, true, 0))

	status, err = adapter.GetJobStatus(ctx, "ve-job-integration")
	require.NoError(t, err)
	require.Equal(t, slurm.SLURMJobStateCompleted, status.State)

	metrics, err := client.GetJobAccounting(ctx, job.SLURMJobID)
	require.NoError(t, err)
	require.NotNil(t, metrics)
}
