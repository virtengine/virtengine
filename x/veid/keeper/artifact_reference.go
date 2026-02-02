package keeper

import (
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Artifact Reference Management (VE-218)
// ============================================================================

// SetArtifactReference stores an identity artifact reference
func (k Keeper) SetArtifactReference(ctx sdk.Context, ref *types.IdentityArtifactReference) error {
	if ref == nil {
		return types.ErrInvalidPayload.Wrap("artifact reference cannot be nil")
	}

	if err := ref.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	// Serialize the reference
	bz, err := json.Marshal(ref)
	if err != nil {
		return types.ErrInvalidPayload.Wrapf("failed to marshal artifact reference: %v", err)
	}

	// Store by reference ID
	store.Set(types.ArtifactReferenceKey(ref.ReferenceID), bz)

	// Store index by account and type
	addrBytes := []byte(ref.AccountAddress)
	indexKey := types.ArtifactReferenceByAccountKey(addrBytes, ref.ArtifactType)
	k.appendToIndex(ctx, indexKey, ref.ReferenceID)

	// Store index by content hash
	contentHashKey := types.ArtifactReferenceByContentHashKey(ref.ContentAddress.Hash)
	store.Set(contentHashKey, []byte(ref.ReferenceID))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"artifact_reference_stored",
			sdk.NewAttribute("reference_id", ref.ReferenceID),
			sdk.NewAttribute("account_address", ref.AccountAddress),
			sdk.NewAttribute("artifact_type", string(ref.ArtifactType)),
			sdk.NewAttribute("backend", string(ref.ContentAddress.Backend)),
			sdk.NewAttribute("content_hash", ref.ContentAddress.HashHex()),
		),
	)

	k.Logger(ctx).Debug("stored artifact reference",
		"reference_id", ref.ReferenceID,
		"account", ref.AccountAddress,
		"artifact_type", ref.ArtifactType,
		"backend", ref.ContentAddress.Backend,
	)

	return nil
}

// GetArtifactReference retrieves an artifact reference by ID
func (k Keeper) GetArtifactReference(ctx sdk.Context, referenceID string) (*types.IdentityArtifactReference, bool) {
	store := ctx.KVStore(k.skey)

	bz := store.Get(types.ArtifactReferenceKey(referenceID))
	if bz == nil {
		return nil, false
	}

	var ref types.IdentityArtifactReference
	if err := json.Unmarshal(bz, &ref); err != nil {
		k.Logger(ctx).Error("failed to unmarshal artifact reference", "reference_id", referenceID, "error", err)
		return nil, false
	}

	return &ref, true
}

// GetArtifactReferenceByContentHash retrieves an artifact reference by content hash
func (k Keeper) GetArtifactReferenceByContentHash(ctx sdk.Context, contentHash []byte) (*types.IdentityArtifactReference, bool) {
	store := ctx.KVStore(k.skey)

	referenceIDBytes := store.Get(types.ArtifactReferenceByContentHashKey(contentHash))
	if referenceIDBytes == nil {
		return nil, false
	}

	return k.GetArtifactReference(ctx, string(referenceIDBytes))
}

// GetArtifactReferencesByAccount retrieves all artifact references for an account
func (k Keeper) GetArtifactReferencesByAccount(ctx sdk.Context, address sdk.AccAddress) []*types.IdentityArtifactReference {
	store := ctx.KVStore(k.skey)
	prefix := types.ArtifactReferenceByAccountPrefixKey(address.Bytes())

	refs := make([]*types.IdentityArtifactReference, 0)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// The value is a list of reference IDs
		var refIDs []string
		if err := json.Unmarshal(iterator.Value(), &refIDs); err != nil {
			continue
		}

		for _, refID := range refIDs {
			if ref, found := k.GetArtifactReference(ctx, refID); found {
				refs = append(refs, ref)
			}
		}
	}

	return refs
}

// GetArtifactReferencesByAccountAndType retrieves artifact references for an account and type
func (k Keeper) GetArtifactReferencesByAccountAndType(ctx sdk.Context, address sdk.AccAddress, artifactType types.ArtifactType) []*types.IdentityArtifactReference {
	store := ctx.KVStore(k.skey)
	indexKey := types.ArtifactReferenceByAccountKey(address.Bytes(), artifactType)

	bz := store.Get(indexKey)
	if bz == nil {
		return make([]*types.IdentityArtifactReference, 0)
	}

	var refIDs []string
	if err := json.Unmarshal(bz, &refIDs); err != nil {
		return make([]*types.IdentityArtifactReference, 0)
	}

	refs := make([]*types.IdentityArtifactReference, 0, len(refIDs))
	for _, refID := range refIDs {
		if ref, found := k.GetArtifactReference(ctx, refID); found {
			refs = append(refs, ref)
		}
	}

	return refs
}

// DeleteArtifactReference removes an artifact reference
func (k Keeper) DeleteArtifactReference(ctx sdk.Context, referenceID string) error {
	ref, found := k.GetArtifactReference(ctx, referenceID)
	if !found {
		return types.ErrInvalidPayload.Wrapf("artifact reference not found: %s", referenceID)
	}

	store := ctx.KVStore(k.skey)

	// Remove from main store
	store.Delete(types.ArtifactReferenceKey(referenceID))

	// Remove from content hash index
	store.Delete(types.ArtifactReferenceByContentHashKey(ref.ContentAddress.Hash))

	// Remove from account index
	addrBytes := []byte(ref.AccountAddress)
	indexKey := types.ArtifactReferenceByAccountKey(addrBytes, ref.ArtifactType)
	k.removeFromIndex(ctx, indexKey, referenceID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"artifact_reference_deleted",
			sdk.NewAttribute("reference_id", referenceID),
			sdk.NewAttribute("account_address", ref.AccountAddress),
			sdk.NewAttribute("artifact_type", string(ref.ArtifactType)),
		),
	)

	return nil
}

// RevokeArtifactReference marks an artifact reference as revoked
func (k Keeper) RevokeArtifactReference(ctx sdk.Context, referenceID string, reason string) error {
	ref, found := k.GetArtifactReference(ctx, referenceID)
	if !found {
		return types.ErrInvalidPayload.Wrapf("artifact reference not found: %s", referenceID)
	}

	if ref.IsRevoked() {
		return types.ErrInvalidPayload.Wrap("artifact reference already revoked")
	}

	// Revoke the reference
	ref.Revoke(reason, ctx.BlockTime())

	// Update in store
	return k.SetArtifactReference(ctx, ref)
}

// HasArtifactReference checks if an artifact reference exists
func (k Keeper) HasArtifactReference(ctx sdk.Context, referenceID string) bool {
	store := ctx.KVStore(k.skey)
	return store.Has(types.ArtifactReferenceKey(referenceID))
}

// HasArtifactByContentHash checks if an artifact with the given content hash exists
func (k Keeper) HasArtifactByContentHash(ctx sdk.Context, contentHash []byte) bool {
	store := ctx.KVStore(k.skey)
	return store.Has(types.ArtifactReferenceByContentHashKey(contentHash))
}

// IterateArtifactReferences iterates over all artifact references
func (k Keeper) IterateArtifactReferences(ctx sdk.Context, fn func(ref *types.IdentityArtifactReference) bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PrefixArtifactReference)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var ref types.IdentityArtifactReference
		if err := json.Unmarshal(iterator.Value(), &ref); err != nil {
			continue
		}
		if fn(&ref) {
			break
		}
	}
}

// ============================================================================
// Chunk Manifest Management
// ============================================================================

// SetChunkManifest stores a chunk manifest
func (k Keeper) SetChunkManifest(ctx sdk.Context, manifestID string, manifest *types.ChunkManifestReference) error {
	if manifest == nil {
		return types.ErrInvalidPayload.Wrap("chunk manifest cannot be nil")
	}

	if err := manifest.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(manifest)
	if err != nil {
		return types.ErrInvalidPayload.Wrapf("failed to marshal chunk manifest: %v", err)
	}

	store.Set(types.ChunkManifestKey(manifestID), bz)
	return nil
}

// GetChunkManifest retrieves a chunk manifest by ID
func (k Keeper) GetChunkManifest(ctx sdk.Context, manifestID string) (*types.ChunkManifestReference, bool) {
	store := ctx.KVStore(k.skey)

	bz := store.Get(types.ChunkManifestKey(manifestID))
	if bz == nil {
		return nil, false
	}

	var manifest types.ChunkManifestReference
	if err := json.Unmarshal(bz, &manifest); err != nil {
		k.Logger(ctx).Error("failed to unmarshal chunk manifest", "manifest_id", manifestID, "error", err)
		return nil, false
	}

	return &manifest, true
}

// DeleteChunkManifest removes a chunk manifest
func (k Keeper) DeleteChunkManifest(ctx sdk.Context, manifestID string) {
	store := ctx.KVStore(k.skey)
	store.Delete(types.ChunkManifestKey(manifestID))
}

// ============================================================================
// Pending Artifact Retrieval Management
// ============================================================================

// PendingArtifactRetrieval represents a pending artifact retrieval request
type PendingArtifactRetrieval struct {
	RequestID         string `json:"request_id"`
	ReferenceID       string `json:"reference_id"`
	RequestingAccount string `json:"requesting_account"`
	Purpose           string `json:"purpose"`
	RequestedAt       int64  `json:"requested_at"`
	ExpiresAt         int64  `json:"expires_at"`
	Status            string `json:"status"`
}

// SetPendingArtifactRetrieval stores a pending retrieval request
func (k Keeper) SetPendingArtifactRetrieval(ctx sdk.Context, retrieval *PendingArtifactRetrieval) error {
	if retrieval == nil {
		return types.ErrInvalidPayload.Wrap("pending retrieval cannot be nil")
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(retrieval)
	if err != nil {
		return types.ErrInvalidPayload.Wrapf("failed to marshal pending retrieval: %v", err)
	}

	store.Set(types.PendingArtifactRetrievalKey(retrieval.RequestID), bz)
	return nil
}

// GetPendingArtifactRetrieval retrieves a pending retrieval request
func (k Keeper) GetPendingArtifactRetrieval(ctx sdk.Context, requestID string) (*PendingArtifactRetrieval, bool) {
	store := ctx.KVStore(k.skey)

	bz := store.Get(types.PendingArtifactRetrievalKey(requestID))
	if bz == nil {
		return nil, false
	}

	var retrieval PendingArtifactRetrieval
	if err := json.Unmarshal(bz, &retrieval); err != nil {
		return nil, false
	}

	return &retrieval, true
}

// DeletePendingArtifactRetrieval removes a pending retrieval request
func (k Keeper) DeletePendingArtifactRetrieval(ctx sdk.Context, requestID string) {
	store := ctx.KVStore(k.skey)
	store.Delete(types.PendingArtifactRetrievalKey(requestID))
}

// IteratePendingArtifactRetrievals iterates over all pending retrieval requests
func (k Keeper) IteratePendingArtifactRetrievals(ctx sdk.Context, fn func(retrieval *PendingArtifactRetrieval) bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PendingArtifactRetrievalPrefixKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var retrieval PendingArtifactRetrieval
		if err := json.Unmarshal(iterator.Value(), &retrieval); err != nil {
			continue
		}
		if fn(&retrieval) {
			break
		}
	}
}

// ============================================================================
// Index Management Helpers
// ============================================================================

// appendToIndex adds an ID to an index stored as a JSON array
func (k Keeper) appendToIndex(ctx sdk.Context, key []byte, id string) {
	store := ctx.KVStore(k.skey)

	var ids []string
	bz := store.Get(key)
	if bz != nil {
		_ = json.Unmarshal(bz, &ids)
	}

	// Check if already exists
	for _, existingID := range ids {
		if existingID == id {
			return
		}
	}

	ids = append(ids, id)

	newBz, _ := json.Marshal(ids) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, newBz)
}

// removeFromIndex removes an ID from an index stored as a JSON array
func (k Keeper) removeFromIndex(ctx sdk.Context, key []byte, id string) {
	store := ctx.KVStore(k.skey)

	var ids []string
	bz := store.Get(key)
	if bz == nil {
		return
	}

	if err := json.Unmarshal(bz, &ids); err != nil {
		return
	}

	newIDs := make([]string, 0, len(ids))
	for _, existingID := range ids {
		if existingID != id {
			newIDs = append(newIDs, existingID)
		}
	}

	if len(newIDs) == 0 {
		store.Delete(key)
	} else {
		newBz, _ := json.Marshal(newIDs) //nolint:errchkjson // string slice cannot fail to marshal
		store.Set(key, newBz)
	}
}
