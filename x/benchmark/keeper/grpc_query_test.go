package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/benchmark/types"
)

func TestGRPCQueryBenchmark(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)
	provider := bech32Addr(t)

	report := types.BenchmarkReport{
		ReportID:        "report-1",
		ProviderAddress: provider,
		ClusterID:       "cluster-1",
		SuiteVersion:    "1.0.0",
		SummaryScore:    9000,
		Timestamp:       ctx.BlockTime(),
		BlockHeight:     ctx.BlockHeight(),
	}
	require.NoError(t, keeper.SetBenchmarkReport(ctx, report))

	querier := GRPCQuerier{Keeper: keeper}
	resp, err := querier.Benchmark(ctx, &benchmarkv1.QueryBenchmarkRequest{
		ReportId: report.ReportID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, report.ReportID, resp.Report.ReportId)
}

func TestGRPCQueryBenchmarksByProvider(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)
	provider := bech32Addr(t)

	report1 := types.BenchmarkReport{
		ReportID:        "report-2",
		ProviderAddress: provider,
		ClusterID:       "cluster-1",
		SuiteVersion:    "1.0.0",
		SummaryScore:    8000,
		Timestamp:       ctx.BlockTime(),
		BlockHeight:     ctx.BlockHeight(),
	}
	report2 := types.BenchmarkReport{
		ReportID:        "report-3",
		ProviderAddress: provider,
		ClusterID:       "cluster-2",
		SuiteVersion:    "1.1.0",
		SummaryScore:    8500,
		Timestamp:       ctx.BlockTime(),
		BlockHeight:     ctx.BlockHeight(),
	}
	require.NoError(t, keeper.SetBenchmarkReport(ctx, report1))
	require.NoError(t, keeper.SetBenchmarkReport(ctx, report2))

	querier := GRPCQuerier{Keeper: keeper}
	resp, err := querier.BenchmarksByProvider(ctx, &benchmarkv1.QueryBenchmarksByProviderRequest{
		Provider: provider,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Reports, 2)
}

func bech32Addr(t *testing.T) string {
	t.Helper()
	return sdk.MustBech32ifyAddressBytes(sdkutil.Bech32PrefixAccAddr, testutil.AccAddress(t))
}

func TestGRPCQueryBenchmarkParams(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	customParams := types.Params{
		RetentionCount:                  5,
		DefaultChallengeDeadlineSeconds: 10,
		MinBenchmarkInterval:            20,
		MaxReportsPerSubmission:         2,
		AnomalyThresholdJumpPercent:     15,
		AnomalyThresholdRepeatCount:     4,
	}
	require.NoError(t, keeper.SetParams(ctx, customParams))

	querier := GRPCQuerier{Keeper: keeper}
	resp, err := querier.Params(ctx, &benchmarkv1.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, customParams.RetentionCount, resp.Params.RetentionCount)
	require.Equal(t, customParams.DefaultChallengeDeadlineSeconds, resp.Params.DefaultChallengeDeadlineSeconds)
	require.Equal(t, customParams.MinBenchmarkInterval, resp.Params.MinBenchmarkInterval)
	require.Equal(t, customParams.MaxReportsPerSubmission, resp.Params.MaxReportsPerSubmission)
	require.Equal(t, customParams.AnomalyThresholdJumpPercent, resp.Params.AnomalyThresholdJumpPercent)
	require.Equal(t, customParams.AnomalyThresholdRepeatCount, resp.Params.AnomalyThresholdRepeatCount)
}
