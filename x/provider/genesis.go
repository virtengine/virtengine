package provider

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	"github.com/virtengine/virtengine/x/provider/keeper"
)

// ValidateGenesis does validation check of the Genesis and returns error in case of failure

func ValidateGenesis(data *types.GenesisState) error {
	if data == nil {
		return fmt.Errorf("provider genesis state is nil")
	}
	providerOwners := make(map[string]struct{}, len(data.Providers))
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
		if _, exists := providerOwners[record.Owner]; exists {
			return fmt.Errorf("duplicate provider owner %s", record.Owner)
		}
		providerOwners[record.Owner] = struct{}{}
	}
	currentByOwner := make(map[string]uint64)
	seenEpoch := make(map[string]bool)
	epochsByOwner := make(map[string][]types.ProviderPublicKeyRecord)
	seenKeyIDByOwner := make(map[string]map[string]struct{})
	for _, item := range data.SigningKeys {
		owner, err := sdk.AccAddressFromBech32(item.Owner)
		if err != nil {
			return err
		}
		if _, exists := providerOwners[owner.String()]; !exists {
			return fmt.Errorf("provider signing key owner %s has no provider record", owner)
		}
		record := ProviderKeyRecordFromProto(item.Key)
		if err := record.Validate(); err != nil {
			return err
		}
		if record.Epoch == 0 || record.KeyID == "" {
			return fmt.Errorf("provider signing key for %s is missing epoch metadata", owner)
		}
		key := owner.String() + "/" + fmt.Sprint(record.Epoch)
		if seenEpoch[key] {
			return fmt.Errorf("duplicate provider signing key epoch %s", key)
		}
		seenEpoch[key] = true
		if seenKeyIDByOwner[owner.String()] == nil {
			seenKeyIDByOwner[owner.String()] = make(map[string]struct{})
		}
		if _, exists := seenKeyIDByOwner[owner.String()][record.KeyID]; exists {
			return fmt.Errorf("duplicate provider signing key ID %s for %s", record.KeyID, owner)
		}
		seenKeyIDByOwner[owner.String()][record.KeyID] = struct{}{}
		epochsByOwner[owner.String()] = append(epochsByOwner[owner.String()], record)
		if item.Current {
			if currentByOwner[owner.String()] != 0 {
				return fmt.Errorf("multiple current signing keys for %s", owner)
			}
			currentByOwner[owner.String()] = record.Epoch
		}
	}
	for owner, epochs := range epochsByOwner {
		sort.Slice(epochs, func(i, j int) bool { return epochs[i].Epoch < epochs[j].Epoch })
		if currentByOwner[owner] == 0 {
			return fmt.Errorf("provider %s signing-key history has no current epoch", owner)
		}
		for index, record := range epochs {
			expectedEpoch := uint64(index + 1) //nolint:gosec // index is bounded by the in-memory genesis slice length
			if record.Epoch != expectedEpoch {
				return fmt.Errorf("provider %s signing-key epochs are not contiguous: expected %d, got %d", owner, expectedEpoch, record.Epoch)
			}
			if index == 0 {
				if record.PreviousKeyID != "" || record.RotationCount != 0 {
					return fmt.Errorf("provider %s initial signing-key epoch has invalid lineage", owner)
				}
				continue
			}
			previous := epochs[index-1]
			if record.PreviousKeyID != previous.KeyID || record.RotationCount != uint32(index) { //nolint:gosec // index is bounded by the in-memory genesis slice length
				return fmt.Errorf("provider %s signing-key epoch %d has invalid predecessor lineage", owner, record.Epoch)
			}
			if record.ActivatedAtHeight < previous.ActivatedAtHeight ||
				(record.ActivatedAtUnix > 0 && previous.ActivatedAtUnix > 0 && record.ActivatedAtUnix < previous.ActivatedAtUnix) {
				return fmt.Errorf("provider %s signing-key activation order regresses", owner)
			}
		}
		if currentByOwner[owner] != epochs[len(epochs)-1].Epoch {
			return fmt.Errorf("provider %s current signing key is not the highest epoch", owner)
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
		if err := kpr.ImportProviderSigningKeyEpoch(ctx, owner, ProviderKeyRecordFromProto(item.Key), item.Current); err != nil {
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
			Key:     ProviderKeyRecordToProto(record),
			Current: found && record.KeyID == current.KeyID && record.Epoch == current.Epoch,
		})
		return false
	})

	return &types.GenesisState{
		Providers:   providers,
		SigningKeys: signingKeys,
	}
}

func ProviderKeyRecordFromProto(record types.ProviderSigningKeyRecord) types.ProviderPublicKeyRecord {
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

func ProviderKeyRecordToProto(record types.ProviderPublicKeyRecord) types.ProviderSigningKeyRecord {
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
