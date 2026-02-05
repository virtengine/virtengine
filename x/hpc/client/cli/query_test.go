package cli

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	query "github.com/cosmos/cosmos-sdk/types/query"

	hpctypes "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

func TestParseJobStateFilter(t *testing.T) {
	state, ok := parseJobStateFilter("queued")
	require.True(t, ok)
	require.Equal(t, hpctypes.JobStateQueued, state)

	state, ok = parseJobStateFilter("JOB_STATE_COMPLETED")
	require.True(t, ok)
	require.Equal(t, hpctypes.JobStateCompleted, state)

	_, ok = parseJobStateFilter("not-a-state")
	require.False(t, ok)
}

func TestResolveOwnerFilter(t *testing.T) {
	owner, err := resolveOwnerFilter("", "addr1")
	require.NoError(t, err)
	require.Equal(t, "addr1", owner)

	owner, err = resolveOwnerFilter("addr2", "")
	require.NoError(t, err)
	require.Equal(t, "addr2", owner)

	_, err = resolveOwnerFilter("addr1", "addr2")
	require.Error(t, err)
}

func TestReadActiveFilter(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool(flagActive, false, "")
	cmd.Flags().Bool(flagInactive, false, "")

	value, err := readActiveFilter(cmd)
	require.NoError(t, err)
	require.Nil(t, value)

	require.NoError(t, cmd.Flags().Set(flagActive, "true"))
	value, err = readActiveFilter(cmd)
	require.NoError(t, err)
	require.NotNil(t, value)
	require.True(t, *value)

	require.NoError(t, cmd.Flags().Set(flagInactive, "true"))
	_, err = readActiveFilter(cmd)
	require.Error(t, err)
}

func TestPaginateSlice(t *testing.T) {
	items := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	pageReq := &query.PageRequest{
		Offset:     2,
		Limit:      3,
		CountTotal: true,
	}

	pageItems, pageResp := paginateSlice(items, pageReq)
	require.Equal(t, []int{2, 3, 4}, pageItems)
	require.NotNil(t, pageResp)
	require.Equal(t, uint64(10), pageResp.Total)
}

func TestParseOutputLocation(t *testing.T) {
	location := parseOutputLocation("done output=s3://bucket/key")
	require.Equal(t, "s3://bucket/key", location)

	location = parseOutputLocation("no output here")
	require.Equal(t, "", location)
}

func TestQueuePositionForJob(t *testing.T) {
	t1 := time.Now().Add(-3 * time.Hour)
	t2 := time.Now().Add(-2 * time.Hour)
	t3 := time.Now().Add(-1 * time.Hour)

	job1 := hpctypes.HPCJob{JobId: "job-1", ClusterId: "cluster", QueueName: "main", State: hpctypes.JobStateQueued, CreatedAt: t1, QueuedAt: &t1}
	job2 := hpctypes.HPCJob{JobId: "job-2", ClusterId: "cluster", QueueName: "main", State: hpctypes.JobStateQueued, CreatedAt: t2, QueuedAt: &t2}
	job3 := hpctypes.HPCJob{JobId: "job-3", ClusterId: "cluster", QueueName: "main", State: hpctypes.JobStateQueued, CreatedAt: t3, QueuedAt: &t3}
	job4 := hpctypes.HPCJob{JobId: "job-4", ClusterId: "cluster", QueueName: "gpu", State: hpctypes.JobStateQueued, CreatedAt: t2, QueuedAt: &t2}

	resp, err := queuePositionForJob(job2, []hpctypes.HPCJob{job4, job3, job2, job1})
	require.NoError(t, err)
	require.Equal(t, int64(2), resp.Position)
	require.Equal(t, int64(1), resp.Ahead)
	require.Equal(t, int64(3), resp.TotalInQueue)
	require.Equal(t, "main", resp.QueueName)
}

func TestFilterJobsByFields(t *testing.T) {
	jobs := []hpctypes.HPCJob{
		{JobId: "a", ProviderAddress: "prov1", ClusterId: "c1", QueueName: "main", CustomerAddress: "cust1", State: hpctypes.JobStateQueued},
		{JobId: "b", ProviderAddress: "prov2", ClusterId: "c2", QueueName: "gpu", CustomerAddress: "cust2", State: hpctypes.JobStateRunning},
		{JobId: "c", ProviderAddress: "prov1", ClusterId: "c1", QueueName: "main", CustomerAddress: "cust1", State: hpctypes.JobStateCompleted},
	}

	state := hpctypes.JobStateQueued
	filtered := filterJobsByFields(jobs, jobFilterOptions{
		state:     &state,
		provider:  "prov1",
		clusterID: "c1",
		queueName: "main",
		owner:     "cust1",
	})

	require.Len(t, filtered, 1)
	require.Equal(t, "a", filtered[0].JobId)
}
