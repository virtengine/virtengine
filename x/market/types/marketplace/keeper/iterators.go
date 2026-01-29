package keeper

import (
	"encoding/json"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// WithBids iterates over all bids.
func (k Keeper) WithBids(ctx sdk.Context, fn func(marketplace.MarketplaceBid) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.BidKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var bid marketplace.MarketplaceBid
		if err := json.Unmarshal(iter.Value(), &bid); err != nil {
			continue
		}
		if stop := fn(bid); stop {
			return
		}
	}
}

// WithProviderSettings iterates over provider identity settings.
func (k Keeper) WithProviderSettings(ctx sdk.Context, fn func(address string, settings marketplace.ProviderIdentitySettings) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.ProviderSettingsKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key()[len(marketplace.ProviderSettingsKeyPrefix):])
		var settings marketplace.ProviderIdentitySettings
		if err := json.Unmarshal(iter.Value(), &settings); err != nil {
			continue
		}
		if stop := fn(key, settings); stop {
			return
		}
	}
}

// WithMFAActionConfigs iterates over MFA action configs.
func (k Keeper) WithMFAActionConfigs(ctx sdk.Context, fn func(config marketplace.MFAActionConfig) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.MFAConfigKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var cfg marketplace.MFAActionConfig
		if err := json.Unmarshal(iter.Value(), &cfg); err != nil {
			continue
		}
		if stop := fn(cfg); stop {
			return
		}
	}
}
