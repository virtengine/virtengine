package provider

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	"github.com/virtengine/virtengine/x/provider/keeper"
)

// ValidateGenesis does validation check of the Genesis and returns error in case of failure

func ValidateGenesis(data *types.GenesisState) error {
	for _, record := range data.Providers {
		msg := &types.MsgCreateProvider{
			Owner:      record.Owner,
			HostURI:    record.HostURI,
			Attributes: record.Attributes,
			Info:       record.Info,
		}

		if err := msg.ValidateBasic(); err != nil {
			return err
		}

	}
	currentByOwner := make(map[string]bool)
	seenEpoch := make(map[string]bool)
	for _, item := range data.SigningKeys {
		owner, err := sdk.AccAddressFromBech32(item.Owner)
		if err != nil {
			return err
		}
		record := providerKeyRecordFromProto(item.Key)
		if err := record.Validate(); err != nil {
			return err
		}
		key := owner.String() + "/" + fmt.Sprint(record.Epoch)
		if seenEpoch[key] {
			return fmt.Errorf("duplicate provider signing key epoch %s", key)
		}
		seenEpoch[key] = true
		if item.Current {
			if currentByOwner[owner.String()] {
				return fmt.Errorf("multiple current signing keys for %s", owner)
			}
			currentByOwner[owner.String()] = true
		}
	}

	return nil
}

// InitGenesis initiate genesis state and return updated validator details
func InitGenesis(ctx sdk.Context, kpr keeper.IKeeper, data *types.GenesisState) {
	store := ctx.KVStore(kpr.StoreKey())
	cdc := kpr.Codec()

	for _, record := range data.Providers {
		owner, err := sdk.AccAddressFromBech32(record.Owner)
		if err != nil {
			panic(fmt.Sprintf("provider genesis init: %s", err.Error()))
		}

		key := keeper.ProviderKey(owner)

		if store.Has(key) {
			panic(fmt.Sprintf("provider genesis init: %s", types.ErrProviderExists.Error()))
		}

		store.Set(key, cdc.MustMarshal(&record))
	}
	for _, item := range data.SigningKeys {
		owner, err := sdk.AccAddressFromBech32(item.Owner)
		if err != nil {
			panic(err)
		}
		if err := kpr.ImportProviderSigningKeyEpoch(ctx, owner, providerKeyRecordFromProto(item.Key), item.Current); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns genesis state as raw bytes for the provider module
func ExportGenesis(ctx sdk.Context, k keeper.IKeeper) *types.GenesisState {
	var providers []types.Provider
	var signingKeys []types.ProviderSigningKeyGenesisRecord

	k.WithProviders(ctx, func(provider types.Provider) bool {
		providers = append(providers, provider)
		return false
	})
	k.WithProviderSigningKeys(ctx, func(owner sdk.AccAddress, record types.ProviderPublicKeyRecord) bool {
		current, found := k.GetProviderPublicKeyRecord(ctx, owner)
		signingKeys = append(signingKeys, types.ProviderSigningKeyGenesisRecord{
			Owner:   owner.String(),
			Key:     providerKeyRecordToProto(record),
			Current: found && record.KeyID == current.KeyID && record.Epoch == current.Epoch,
		})
		return false
	})

	return &types.GenesisState{
		Providers:   providers,
		SigningKeys: signingKeys,
	}
}

func providerKeyRecordFromProto(record types.ProviderSigningKeyRecord) types.ProviderPublicKeyRecord {
	return types.ProviderPublicKeyRecord{
		PublicKey:         append([]byte(nil), record.PublicKey...),
		KeyType:           record.KeyType,
		KeyID:             record.KeyId,
		Epoch:             record.Epoch,
		ActivatedAtHeight: record.ActivatedAtHeight,
		ActivatedAtUnix:   record.ActivatedAtUnix,
		ExpiresAtHeight:   record.ExpiresAtHeight,
		ExpiresAtUnix:     record.ExpiresAtUnix,
		RetiredAtHeight:   record.RetiredAtHeight,
		RetiredAtUnix:     record.RetiredAtUnix,
		RevokedAtHeight:   record.RevokedAtHeight,
		RevokedAtUnix:     record.RevokedAtUnix,
		PreviousKeyID:     record.PreviousKeyId,
		RotationCount:     record.RotationCount,
		UpdatedAt:         record.ActivatedAtHeight,
	}
}

func providerKeyRecordToProto(record types.ProviderPublicKeyRecord) types.ProviderSigningKeyRecord {
	return types.ProviderSigningKeyRecord{
		PublicKey:         append([]byte(nil), record.PublicKey...),
		KeyType:           record.KeyType,
		KeyId:             record.KeyID,
		Epoch:             record.Epoch,
		ActivatedAtHeight: record.ActivatedAtHeight,
		ActivatedAtUnix:   record.ActivatedAtUnix,
		ExpiresAtHeight:   record.ExpiresAtHeight,
		ExpiresAtUnix:     record.ExpiresAtUnix,
		RetiredAtHeight:   record.RetiredAtHeight,
		RetiredAtUnix:     record.RetiredAtUnix,
		RevokedAtHeight:   record.RevokedAtHeight,
		RevokedAtUnix:     record.RevokedAtUnix,
		PreviousKeyId:     record.PreviousKeyID,
		RotationCount:     record.RotationCount,
	}
}

// DefaultGenesisState returns default genesis state as raw bytes for the provider
// module.
func DefaultGenesisState() *types.GenesisState {
	return &types.GenesisState{}
}

// GetGenesisStateFromAppState returns x/provider GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *types.GenesisState {
	var genesisState types.GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}
