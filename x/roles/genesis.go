package roles

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/roles/keeper"
	"github.com/virtengine/virtengine/x/roles/types"
)

// InitGenesis initializes the roles module's state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Initialize genesis accounts
	for _, addrStr := range data.GenesisAccounts {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			panic(err)
		}

		if err := k.AddGenesisAccount(ctx, addr); err != nil {
			panic(err)
		}
	}

	// Initialize role assignments
	for _, ra := range data.RoleAssignments {
		addr, err := sdk.AccAddressFromBech32(ra.Address)
		if err != nil {
			panic(err)
		}

		assignedBy, err := sdk.AccAddressFromBech32(ra.AssignedBy)
		if err != nil {
			// If no assigned_by, use the address itself
			assignedBy = addr
		}

		if err := k.AssignRole(ctx, addr, ra.Role, assignedBy); err != nil {
			// Skip if role already assigned (e.g., genesis accounts)
			if err != types.ErrRoleAlreadyAssigned {
				panic(err)
			}
		}
	}

	// Initialize account states
	for _, as := range data.AccountStates {
		addr, err := sdk.AccAddressFromBech32(as.Address)
		if err != nil {
			panic(err)
		}

		modifiedBy, err := sdk.AccAddressFromBech32(as.ModifiedBy)
		if err != nil {
			modifiedBy = addr
		}

		if err := k.SetAccountState(ctx, addr, as.State, as.Reason, modifiedBy); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the roles module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get genesis accounts
	genesisAccounts := k.GetGenesisAccounts(ctx)
	genesisAccountStrs := make([]string, len(genesisAccounts))
	for i, addr := range genesisAccounts {
		genesisAccountStrs[i] = addr.String()
	}

	// Get all role assignments
	allRoles := types.AllRoles()
	roleAssignments := make([]types.RoleAssignment, 0, len(allRoles))
	for _, role := range allRoles {
		members := k.GetRoleMembers(ctx, role)
		roleAssignments = append(roleAssignments, members...)
	}

	// Get all account states by iterating through known accounts
	var accountStates []types.AccountStateRecord
	seenAddresses := make(map[string]bool)

	// Collect addresses from role assignments
	for _, ra := range roleAssignments {
		if !seenAddresses[ra.Address] {
			seenAddresses[ra.Address] = true
			addr, _ := sdk.AccAddressFromBech32(ra.Address)
			if state, found := k.GetAccountState(ctx, addr); found {
				accountStates = append(accountStates, state)
			}
		}
	}

	return &types.GenesisState{
		GenesisAccounts: genesisAccountStrs,
		RoleAssignments: roleAssignments,
		AccountStates:   accountStates,
		Params:          params,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *types.GenesisState {
	return types.DefaultGenesisState()
}
