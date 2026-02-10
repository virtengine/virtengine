package encryption

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
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
	addresses := make(map[string]sdk.AccAddress)
	for _, keyRecord := range data.RecipientKeys {
		addr, err := sdk.AccAddressFromBech32(keyRecord.Address)
		if err != nil {
			panic(err)
		}

		record := types.RecipientKeyRecord{
			Address:        keyRecord.Address,
			PublicKey:      keyRecord.PublicKey,
			KeyFingerprint: keyRecord.KeyFingerprint,
			KeyVersion:     keyRecord.KeyVersion,
			AlgorithmID:    keyRecord.AlgorithmId,
			RegisteredAt:   keyRecord.RegisteredAt,
			RevokedAt:      keyRecord.RevokedAt,
			DeprecatedAt:   keyRecord.DeprecatedAt,
			ExpiresAt:      keyRecord.ExpiresAt,
			PurgeAt:        keyRecord.PurgeAt,
			Label:          keyRecord.Label,
		}

		if err := k.ImportRecipientKeyRecord(ctx, record); err != nil {
			panic(err)
		}
		addresses[addr.String()] = addr
	}

	for _, addr := range addresses {
		k.RefreshActiveRecipientKey(ctx, addr)
	}
}

// ExportGenesis exports the encryption module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get all recipient keys and convert to proto type
	var recipientKeys []encryptionv1.RecipientKeyRecord
	k.WithRecipientKeys(ctx, func(record types.RecipientKeyRecord) bool {
		recipientKeys = append(recipientKeys, encryptionv1.RecipientKeyRecord{
			Address:        record.Address,
			PublicKey:      record.PublicKey,
			KeyFingerprint: record.KeyFingerprint,
			KeyVersion:     record.KeyVersion,
			AlgorithmId:    record.AlgorithmID,
			RegisteredAt:   record.RegisteredAt,
			RevokedAt:      record.RevokedAt,
			DeprecatedAt:   record.DeprecatedAt,
			ExpiresAt:      record.ExpiresAt,
			PurgeAt:        record.PurgeAt,
			Label:          record.Label,
		})
		return false
	})

	return &types.GenesisState{
		RecipientKeys: recipientKeys,
		Params:        params,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return types.ValidateGenesis(data)
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *types.GenesisState {
	return types.DefaultGenesisState()
}
