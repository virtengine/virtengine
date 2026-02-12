package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	"github.com/virtengine/virtengine/testutil"
	"github.com/virtengine/virtengine/x/audit/keeper"
)

func TestAuditLog_AppendAndGet(t *testing.T) {
	ctx, k := setupAuditLogKeeper(t)

	actor := testutil.AccAddress(t).String()
	module := "veid"
	action := "scope_uploaded"
	resourceID := "scope_123"
	metadata := map[string]interface{}{
		"scope_type": "facial",
		"status":     "verified",
	}

	// Append log
	err := k.AppendLog(ctx, actor, module, action, resourceID, metadata)
	require.NoError(t, err)

	// Query logs by module
	filter := types.ExportFilter{
		Module: module,
	}
	logs, err := k.QueryLogs(ctx, filter, 10)
	require.NoError(t, err)
	require.Len(t, logs, 1)

	log := logs[0]
	require.Equal(t, actor, log.Actor)
	require.Equal(t, module, log.Module)
	require.Equal(t, action, log.Action)
	require.Equal(t, resourceID, log.ResourceId)
	require.False(t, log.Exported)

	// Get by ID
	fetchedLog, found := k.GetLogEntry(ctx, log.Id)
	require.True(t, found)
	require.Equal(t, log.Id, fetchedLog.Id)
	require.Equal(t, actor, fetchedLog.Actor)
}

func TestAuditLog_QueryByActor(t *testing.T) {
	ctx, k := setupAuditLogKeeper(t)

	// Use distinct actor addresses to avoid prefix collisions
	actor1 := "ve1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	actor2 := "ve1bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	// Append logs for actor1 (with timestamp separation to avoid ID collisions)
	err := k.AppendLog(ctx, actor1, "veid", "action1", "res1", nil)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Microsecond))
	err = k.AppendLog(ctx, actor1, "market", "action2", "res2", nil)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Microsecond))
	err = k.AppendLog(ctx, actor2, "veid", "action3", "res3", nil)
	require.NoError(t, err)

	// Query logs by actor1
	filter := types.ExportFilter{
		Actor: actor1,
	}
	logs, err := k.QueryLogs(ctx, filter, 10)
	require.NoError(t, err)
	require.Len(t, logs, 2)

	// All logs should be for actor1
	for _, log := range logs {
		require.Equal(t, actor1, log.Actor)
	}
}

func TestAuditLog_ExportJob(t *testing.T) {
	ctx, k := setupAuditLogKeeper(t)

	requester := testutil.AccAddress(t).String()
	filter := types.ExportFilter{
		Module: "veid",
	}
	format := "json"

	// Create export job
	jobID, err := k.CreateExportJob(ctx, requester, filter, format)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)

	// Get export job
	job, found := k.GetExportJob(ctx, jobID)
	require.True(t, found)
	require.Equal(t, jobID, job.Id)
	require.Equal(t, requester, job.Requester)
	require.Equal(t, format, job.Format)
	require.Equal(t, types.ExportStatusUnspecified, job.Status)

	// Update export job status
	job.Status = types.ExportStatusCompleted
	job.EntryCount = 10
	job.FilePath = "/exports/test.json"
	err = k.UpdateExportJob(ctx, job)
	require.NoError(t, err)

	// Verify update
	updatedJob, found := k.GetExportJob(ctx, jobID)
	require.True(t, found)
	require.Equal(t, types.ExportStatusCompleted, updatedJob.Status)
	require.Equal(t, int64(10), updatedJob.EntryCount)
	require.Equal(t, "/exports/test.json", updatedJob.FilePath)
}

func TestAuditLog_Params(t *testing.T) {
	ctx, k := setupAuditLogKeeper(t)

	// Get default params
	params := k.GetAuditLogParams(ctx)
	require.True(t, params.Enabled)
	require.Equal(t, int64(0), params.RetentionBlocks)
	require.Equal(t, int64(10000), params.MaxExportBatchSize)
	require.False(t, params.PruneExported)

	// Set custom params
	newParams := types.AuditLogParams{
		Enabled:            true,
		RetentionBlocks:    1000,
		MaxExportBatchSize: 5000,
		PruneExported:      true,
	}
	err := k.SetAuditLogParams(ctx, newParams)
	require.NoError(t, err)

	// Verify updated params
	updatedParams := k.GetAuditLogParams(ctx)
	require.Equal(t, newParams.Enabled, updatedParams.Enabled)
	require.Equal(t, newParams.RetentionBlocks, updatedParams.RetentionBlocks)
	require.Equal(t, newParams.MaxExportBatchSize, updatedParams.MaxExportBatchSize)
	require.Equal(t, newParams.PruneExported, updatedParams.PruneExported)
}

func TestAuditLog_PruneOldLogs(t *testing.T) {
	ctx, k := setupAuditLogKeeper(t)

	// Set retention policy
	params := types.AuditLogParams{
		Enabled:            true,
		RetentionBlocks:    100,
		MaxExportBatchSize: 10000,
		PruneExported:      true,
	}
	err := k.SetAuditLogParams(ctx, params)
	require.NoError(t, err)

	actor := testutil.AccAddress(t).String()

	// Append log at current height
	err = k.AppendLog(ctx, actor, "veid", "action1", "res1", nil)
	require.NoError(t, err)

	// Move to a much later height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 200)

	// Append another log at new height
	err = k.AppendLog(ctx, actor, "veid", "action2", "res2", nil)
	require.NoError(t, err)

	// Before pruning - should have 2 logs
	filter := types.ExportFilter{
		Actor: actor,
	}
	logs, err := k.QueryLogs(ctx, filter, 10)
	require.NoError(t, err)
	require.Len(t, logs, 2)

	// Prune old logs
	err = k.PruneOldLogs(ctx)
	require.NoError(t, err)

	// After pruning - should have 1 log (the recent one)
	logs, err = k.QueryLogs(ctx, filter, 10)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "action2", logs[0].Action)
}

// Helper function to setup keeper for audit log tests
func setupAuditLogKeeper(t *testing.T) (sdk.Context, keeper.Keeper) {
	t.Helper()
	return setupKeeper(t)
}
