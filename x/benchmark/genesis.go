// Package benchmark implements the Benchmark module for VirtEngine.
package benchmark

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/benchmark/keeper"
	"github.com/virtengine/virtengine/x/benchmark/types"
)

// InitGenesis initializes the benchmark module's state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Set benchmark reports
	for _, report := range data.Reports {
		if err := k.SetBenchmarkReport(ctx, report); err != nil {
			panic(err)
		}
	}

	// Set reliability scores
	for _, score := range data.Scores {
		if err := k.SetReliabilityScore(ctx, score); err != nil {
			panic(err)
		}
	}

	// Set challenges
	for _, challenge := range data.Challenges {
		if err := k.SetChallenge(ctx, challenge); err != nil {
			panic(err)
		}
	}

	// Set anomaly flags
	for _, flag := range data.AnomalyFlags {
		if err := k.SetAnomalyFlag(ctx, flag); err != nil {
			panic(err)
		}
	}

	// Set provider flags
	for _, flag := range data.ProviderFlags {
		if err := k.SetProviderFlag(ctx, flag); err != nil {
			panic(err)
		}
	}

	// Set sequences
	k.SetNextReportSequence(ctx, data.NextReportSequence)
	k.SetNextChallengeSequence(ctx, data.NextChallengeSequence)
	k.SetNextAnomalySequence(ctx, data.NextAnomalySequence)
}

// ExportGenesis exports the benchmark module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	gs := types.DefaultGenesisState()

	// Export parameters
	gs.Params = k.GetParams(ctx)

	// Export benchmark reports
	k.WithBenchmarkReports(ctx, func(report types.BenchmarkReport) bool {
		gs.Reports = append(gs.Reports, report)
		return false
	})

	// Export reliability scores
	k.WithReliabilityScores(ctx, func(score types.ReliabilityScore) bool {
		gs.Scores = append(gs.Scores, score)
		return false
	})

	// Export challenges
	k.WithChallenges(ctx, func(challenge types.BenchmarkChallenge) bool {
		gs.Challenges = append(gs.Challenges, challenge)
		return false
	})

	// Export anomaly flags
	k.WithAnomalyFlags(ctx, func(flag types.AnomalyFlag) bool {
		gs.AnomalyFlags = append(gs.AnomalyFlags, flag)
		return false
	})

	// Export provider flags
	k.WithProviderFlags(ctx, func(flag types.ProviderFlag) bool {
		gs.ProviderFlags = append(gs.ProviderFlags, flag)
		return false
	})

	// Export sequences
	gs.NextReportSequence = k.GetNextReportSequence(ctx)
	gs.NextChallengeSequence = k.GetNextChallengeSequence(ctx)
	gs.NextAnomalySequence = k.GetNextAnomalySequence(ctx)

	return gs
}
