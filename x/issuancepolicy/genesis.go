package issuancepolicy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/issuancepolicy/keeper"
	"github.com/virtengine/virtengine/x/issuancepolicy/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
	for _, policy := range data.Policies {
		if err := k.SetPolicy(ctx, policy); err != nil {
			panic(err)
		}
	}
	if data.ActivePolicyID != "" {
		if err := k.SetActivePolicy(ctx, data.ActivePolicyID); err != nil {
			panic(err)
		}
	}
	if err := k.SetCounters(ctx, data.Counters); err != nil {
		panic(err)
	}
	for _, record := range data.ProofRecords {
		if err := k.SetProofMintRecord(ctx, record); err != nil {
			panic(err)
		}
	}
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	state := types.DefaultGenesisState()
	state.Params = k.GetParams(ctx)
	state.Policies = k.ListPolicies(ctx)
	if active, found := k.GetActivePolicy(ctx); found {
		state.ActivePolicyID = active.PolicyID
	}
	state.Counters = k.GetCounters(ctx)
	state.ProofRecords = k.ListProofMintRecords(ctx)
	return state
}

func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}
