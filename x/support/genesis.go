package support

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/support/keeper"
	"github.com/virtengine/virtengine/x/support/types"
)

// InitGenesis initializes the support module's state from a genesis state.
// This simplified module only manages external ticket references.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Initialize external refs
	for _, ref := range data.ExternalRefs {
		refCopy := ref // Create copy to avoid pointer issues
		if err := k.RegisterExternalRef(ctx, &refCopy); err != nil {
			// Skip if ref already exists (may happen during re-init)
			if err != types.ErrRefAlreadyExists {
				panic(err)
			}
		}
	}
}

// ExportGenesis exports the support module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get all external refs
	var refs []types.ExternalTicketRef
	k.WithExternalRefs(ctx, func(ref types.ExternalTicketRef) bool {
		refs = append(refs, ref)
		return false
	})

	return &types.GenesisState{
		ExternalRefs: refs,
		Params:       params,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}
