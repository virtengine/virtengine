package encryption

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/keeper"
	"github.com/virtengine/virtengine/x/encryption/types"
)

// InitGenesis initializes the encryption module's state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Initialize recipient keys
	for _, keyRecord := range data.RecipientKeys {
		addr, err := sdk.AccAddressFromBech32(keyRecord.Address)
		if err != nil {
			panic(err)
		}

		_, err = k.RegisterRecipientKey(
			ctx,
			addr,
			keyRecord.PublicKey,
			keyRecord.AlgorithmID,
			keyRecord.Label,
		)
		if err != nil {
			// Skip if key already exists
			if err != types.ErrKeyAlreadyExists {
				panic(err)
			}
		}
	}
}

// ExportGenesis exports the encryption module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get all recipient keys
	var recipientKeys []types.RecipientKeyRecord
	k.WithRecipientKeys(ctx, func(record types.RecipientKeyRecord) bool {
		recipientKeys = append(recipientKeys, record)
		return false
	})

	return &types.GenesisState{
		RecipientKeys: recipientKeys,
		Params:        params,
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
