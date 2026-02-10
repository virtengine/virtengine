package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// =====================================================================
// Social Media Scope Storage
// =====================================================================

func (k Keeper) SetSocialMediaScope(ctx sdk.Context, scope *types.SocialMediaScope) error {
	if scope == nil {
		return types.ErrInvalidScope.Wrap("social media scope cannot be nil")
	}
	if err := scope.Validate(); err != nil {
		return err
	}

	bz, err := json.Marshal(scope)
	if err != nil {
		return fmt.Errorf("marshal social media scope: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(k.socialMediaScopeKey(scope.ScopeID), bz)

	store.Set(k.socialMediaScopeByAccountKey(scope.AccountAddress, scope.Provider, scope.ScopeID), []byte{1})

	return nil
}

func (k Keeper) GetSocialMediaScope(ctx sdk.Context, scopeID string) (*types.SocialMediaScope, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(k.socialMediaScopeKey(scopeID))
	if bz == nil {
		return nil, false
	}

	var record types.SocialMediaScope
	if err := json.Unmarshal(bz, &record); err != nil {
		k.Logger(ctx).Error("failed to unmarshal social media scope", "error", err, "scope_id", scopeID)
		return nil, false
	}

	return &record, true
}

func (k Keeper) DeleteSocialMediaScope(ctx sdk.Context, scopeID string) bool {
	store := ctx.KVStore(k.skey)
	record, found := k.GetSocialMediaScope(ctx, scopeID)
	if !found {
		return false
	}

	store.Delete(k.socialMediaScopeKey(scopeID))
	store.Delete(k.socialMediaScopeByAccountKey(record.AccountAddress, record.Provider, record.ScopeID))
	return true
}

// GetSocialMediaScopesByAccount returns all social media scopes for an account, optionally filtered by provider.
func (k Keeper) GetSocialMediaScopesByAccount(ctx sdk.Context, account string, provider *types.SocialMediaProviderType) []types.SocialMediaScope {
	store := ctx.KVStore(k.skey)
	prefixKey := k.socialMediaScopeAccountPrefix(account, provider)
	prefixStore := prefix.NewStore(store, prefixKey)

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	scopes := make([]types.SocialMediaScope, 0)
	for ; iterator.Valid(); iterator.Next() {
		scopeID := string(iterator.Key())
		if provider == nil {
			if idx := lastIndexByte(scopeID, '/'); idx >= 0 && idx+1 < len(scopeID) {
				scopeID = scopeID[idx+1:]
			}
		}
		scope, found := k.GetSocialMediaScope(ctx, scopeID)
		if !found {
			continue
		}
		scopes = append(scopes, *scope)
	}

	return scopes
}

// =====================================================================
// Key Helpers
// =====================================================================

func (k Keeper) socialMediaScopeKey(scopeID string) []byte {
	key := make([]byte, 0, len(types.PrefixSocialMediaScope)+len(scopeID))
	key = append(key, types.PrefixSocialMediaScope...)
	key = append(key, []byte(scopeID)...)
	return key
}

func (k Keeper) socialMediaScopeByAccountKey(account string, provider types.SocialMediaProviderType, scopeID string) []byte {
	key := make([]byte, 0, len(types.PrefixSocialMediaScopeByAccount)+len(account)+1+len(provider)+1+len(scopeID))
	key = append(key, types.PrefixSocialMediaScopeByAccount...)
	key = append(key, []byte(account)...)
	key = append(key, byte('/'))
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	key = append(key, []byte(scopeID)...)
	return key
}

func (k Keeper) socialMediaScopeAccountPrefix(account string, provider *types.SocialMediaProviderType) []byte {
	key := make([]byte, 0, len(types.PrefixSocialMediaScopeByAccount)+len(account)+1+len(types.SocialMediaProviderGoogle)+1)
	key = append(key, types.PrefixSocialMediaScopeByAccount...)
	key = append(key, []byte(account)...)
	key = append(key, byte('/'))
	if provider != nil && *provider != "" {
		key = append(key, []byte(*provider)...)
		key = append(key, byte('/'))
	}
	return key
}

func lastIndexByte(value string, sep byte) int {
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] == sep {
			return i
		}
	}
	return -1
}
