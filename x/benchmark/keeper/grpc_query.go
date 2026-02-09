package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
	"github.com/virtengine/virtengine/x/benchmark/types"
)

// GRPCQuerier implements the gRPC query interface for the benchmark module.
type GRPCQuerier struct {
	Keeper
}

var _ benchmarkv1.QueryServer = GRPCQuerier{}

// Benchmark returns a benchmark report by ID.
func (q GRPCQuerier) Benchmark(c context.Context, req *benchmarkv1.QueryBenchmarkRequest) (*benchmarkv1.QueryBenchmarkResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	report, found := q.GetBenchmarkReport(ctx, req.ReportId)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrReportNotFound.Error())
	}

	return &benchmarkv1.QueryBenchmarkResponse{
		Report: toProtoBenchmarkReport(report),
	}, nil
}

// BenchmarksByProvider returns benchmark reports for a provider.
func (q GRPCQuerier) BenchmarksByProvider(c context.Context, req *benchmarkv1.QueryBenchmarksByProviderRequest) (*benchmarkv1.QueryBenchmarksByProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if _, err := sdk.AccAddressFromBech32(req.Provider); err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	reports := q.GetBenchmarksByProvider(ctx, req.Provider)

	resp := make([]benchmarkv1.BenchmarkReport, 0, len(reports))
	for _, report := range reports {
		resp = append(resp, toProtoBenchmarkReport(report))
	}

	return &benchmarkv1.QueryBenchmarksByProviderResponse{
		Reports: resp,
	}, nil
}

// Params returns the module parameters.
func (q GRPCQuerier) Params(c context.Context, req *benchmarkv1.QueryParamsRequest) (*benchmarkv1.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)

	return &benchmarkv1.QueryParamsResponse{
		Params: benchmarkv1.BenchmarkParams{
			RetentionCount:                  params.RetentionCount,
			DefaultChallengeDeadlineSeconds: params.DefaultChallengeDeadlineSeconds,
			MinBenchmarkInterval:            params.MinBenchmarkInterval,
			MaxReportsPerSubmission:         params.MaxReportsPerSubmission,
			AnomalyThresholdJumpPercent:     params.AnomalyThresholdJumpPercent,
			AnomalyThresholdRepeatCount:     params.AnomalyThresholdRepeatCount,
		},
	}, nil
}

func toProtoBenchmarkReport(report types.BenchmarkReport) benchmarkv1.BenchmarkReport {
	return benchmarkv1.BenchmarkReport{
		ReportId:      report.ReportID,
		Provider:      report.ProviderAddress,
		ClusterId:     report.ClusterID,
		BenchmarkType: report.SuiteVersion,
		Score:         fmt.Sprintf("%d", report.SummaryScore),
		Timestamp:     report.Timestamp.Unix(),
		BlockHeight:   report.BlockHeight,
		ChallengeId:   report.ChallengeID,
	}
}
