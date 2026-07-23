package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

func TestFlagDisputePreservesLegacyPathBeforeFinancialCaseActivation(t *testing.T) {
	ctx, hpcKeeper, settlementKeeper, _ := setupHPCKeeperWithSettlement(t)
	customer := sdk.AccAddress([]byte("task84d-customer-addr"))
	provider := sdk.AccAddress([]byte("task84d-provider-addr"))
	job := types.HPCJob{
		JobID: "task84d-legacy-job", OfferingID: "offering", ClusterID: "cluster",
		ProviderAddress: provider.String(), CustomerAddress: customer.String(), State: types.JobStateRunning,
		QueueName: "default", WorkloadSpec: types.JobWorkloadSpec{ContainerImage: "image"},
		Resources: types.JobResources{Nodes: 1}, MaxRuntimeSeconds: 60,
		AgreedPrice: sdk.NewCoins(sdk.NewInt64Coin("uve", 10)), CreatedAt: time.Unix(1_700_000_000, 0),
	}
	require.NoError(t, hpcKeeper.SetJob(ctx, job))
	require.False(t, settlementKeeper.IsFinancialCasesActive(ctx))

	server := keeper.NewMsgServerImpl(hpcKeeper)
	response, err := server.FlagDispute(ctx, &hpcv1.MsgFlagDispute{
		JobId: job.JobID, DisputerAddress: customer.String(), DisputeType: "billing",
		Reason: "legacy compatibility", Evidence: "legacy evidence",
	})
	require.NoError(t, err)
	require.NotEmpty(t, response.DisputeId)
	require.Empty(t, response.FinancialCaseId)
	dispute, found := hpcKeeper.GetDispute(ctx, response.DisputeId)
	require.True(t, found)
	require.Equal(t, "legacy evidence", dispute.Evidence)
	require.Empty(t, dispute.FinancialCaseID)
}

func TestReportJobStatusRollsBackWhenFinancialCaseHoldsReward(t *testing.T) {
	ctx, hpcKeeper, settlementKeeper, _ := setupHPCKeeperWithSettlement(t)
	settlementKeeper.ActivateFinancialCases(ctx)
	customer := sdk.AccAddress([]byte("task84d-customer-84d"))
	provider := sdk.AccAddress([]byte("task84d-provider-84d"))
	escrow := settlementtypes.NewEscrowAccount("hpc-escrow-84d", "hpc-order-84d", customer.String(), sdk.NewCoins(sdk.NewInt64Coin("uve", 10)), ctx.BlockTime().Add(24*time.Hour), nil, ctx.BlockTime(), ctx.BlockHeight())
	require.NoError(t, escrow.Activate(provider.String(), ctx.BlockTime()))
	require.NoError(t, settlementKeeper.SetEscrow(ctx, *escrow))
	job := types.HPCJob{
		JobID: "task84d-reward-held-job", OfferingID: "offering", ClusterID: "cluster",
		ProviderAddress: provider.String(), CustomerAddress: customer.String(), State: types.JobStateRunning,
		QueueName: "default", WorkloadSpec: types.JobWorkloadSpec{ContainerImage: "image"},
		Resources: types.JobResources{Nodes: 1}, MaxRuntimeSeconds: 60,
		AgreedPrice: sdk.NewCoins(sdk.NewInt64Coin("uve", 10)), CreatedAt: time.Unix(1_700_000_000, 0),
		EscrowID: escrow.EscrowID, MarketOrderID: escrow.OrderID,
	}
	require.NoError(t, hpcKeeper.SetJob(ctx, job))
	server := keeper.NewMsgServerImpl(hpcKeeper)
	dispute, err := server.FlagDispute(ctx, &hpcv1.MsgFlagDispute{
		JobId: job.JobID, DisputerAddress: customer.String(), DisputeType: "billing",
		Reason: "hold rewards", Evidence: "private evidence",
	})
	require.NoError(t, err)
	require.NotEmpty(t, dispute.FinancialCaseId)

	_, err = server.ReportJobStatus(ctx, &hpcv1.MsgReportJobStatus{
		ProviderAddress: provider.String(),
		JobId:           job.JobID,
		State:           hpcv1.JobStateCompleted,
		StatusMessage:   "complete",
	})
	require.ErrorIs(t, err, types.ErrInvalidReward)
	stored, found := hpcKeeper.GetJob(ctx, job.JobID)
	require.True(t, found)
	require.Equal(t, types.JobStateRunning, stored.State)
}
