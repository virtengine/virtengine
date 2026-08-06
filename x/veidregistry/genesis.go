package veidregistry

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veidregistry/keeper"
	"github.com/virtengine/virtengine/x/veidregistry/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
	for _, verifier := range data.Verifiers {
		if err := k.SetVerifierVersion(ctx, verifier); err != nil {
			panic(err)
		}
	}
	for _, readiness := range data.ValidatorReadiness {
		if err := k.SetValidatorReadiness(ctx, readiness); err != nil {
			panic(err)
		}
	}
	if data.ActiveVerifier != nil {
		if err := k.SetActiveVerifier(ctx, *data.ActiveVerifier); err != nil {
			panic(err)
		}
	}
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	state := types.DefaultGenesisState()
	state.Params = k.GetParams(ctx)
	state.Verifiers = k.ListVerifierVersions(ctx)
	if active, found := k.GetActiveVerifier(ctx); found {
		state.ActiveVerifier = active
	}
	for _, verifier := range state.Verifiers {
		state.ValidatorReadiness = append(state.ValidatorReadiness, k.ListValidatorReadiness(ctx, verifier.VerifierID)...)
	}
	return state
}

func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}
