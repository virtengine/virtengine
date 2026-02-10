package veid

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// InitGenesis initializes the veid module's state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Initialize approved clients
	for _, client := range data.ApprovedClients {
		if err := k.SetApprovedClient(ctx, client); err != nil {
			panic(err)
		}
	}

	// Initialize identity records
	for _, record := range data.IdentityRecords {
		if err := k.SetIdentityRecord(ctx, record); err != nil {
			panic(err)
		}
	}

	// Initialize scopes
	for _, scope := range data.Scopes {
		// Find the address from identity records
		for _, record := range data.IdentityRecords {
			for _, ref := range record.ScopeRefs {
				if ref.ScopeID == scope.ScopeID {
					addr, err := sdk.AccAddressFromBech32(record.AccountAddress)
					if err != nil {
						panic(err)
					}
					// Use internal method to bypass validation since this is genesis
					if err := k.UploadScope(ctx, addr, &scope); err != nil {
						// Skip if scope already exists (may happen during re-init)
						if err != types.ErrScopeAlreadyExists {
							panic(err)
						}
					}
					break
				}
			}
		}
	}
}

// ExportGenesis exports the veid module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get all identity records
	var identityRecords []types.IdentityRecord
	k.WithIdentityRecords(ctx, func(record types.IdentityRecord) bool {
		identityRecords = append(identityRecords, record)
		return false
	})

	// Get all scopes
	var scopes []types.IdentityScope
	for _, record := range identityRecords {
		addr, err := sdk.AccAddressFromBech32(record.AccountAddress)
		if err != nil {
			continue
		}
		k.WithScopes(ctx, addr, func(scope types.IdentityScope) bool {
			scopes = append(scopes, scope)
			return false
		})
	}

	// Get all approved clients
	var approvedClients []types.ApprovedClient
	k.WithApprovedClients(ctx, func(client types.ApprovedClient) bool {
		approvedClients = append(approvedClients, client)
		return false
	})

	return &types.GenesisState{
		IdentityRecords: identityRecords,
		Scopes:          scopes,
		ApprovedClients: approvedClients,
		Params:          params,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}

// NOTE: DefaultGenesisState is defined in alias.go as an alias to types.DefaultGenesisState
