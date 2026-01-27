// Package review implements the Review module for VirtEngine.
//
// VE-911: Provider public reviews - Genesis init/export
package review

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/review/keeper"
	"github.com/virtengine/virtengine/x/review/types"
)

// InitGenesis initializes the review module's state from a provided genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	// Set next review sequence
	k.SetNextReviewSequence(ctx, gs.NextReviewSequence)

	// Import reviews
	for _, review := range gs.Reviews {
		if err := k.SetReview(ctx, review); err != nil {
			panic(err)
		}
	}

	// Import aggregations
	for _, agg := range gs.Aggregations {
		if err := k.SetProviderAggregation(ctx, agg); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the review module's state to a genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var reviews []types.Review
	k.WithReviews(ctx, func(r types.Review) bool {
		reviews = append(reviews, r)
		return false
	})

	var aggregations []types.ProviderAggregation
	k.WithProviderAggregations(ctx, func(a types.ProviderAggregation) bool {
		aggregations = append(aggregations, a)
		return false
	})

	return &types.GenesisState{
		Params:             k.GetParams(ctx),
		Reviews:            reviews,
		Aggregations:       aggregations,
		NextReviewSequence: k.GetNextReviewSequence(ctx),
	}
}
