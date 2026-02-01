package keeper

import (
	"encoding/json"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Embedding Envelope Management (VE-217: Derived Feature Minimization)
// ============================================================================

// embeddingEnvelopeStore is the storage format for embedding envelope references
type embeddingEnvelopeStore struct {
	EnvelopeID      string            `json:"envelope_id"`
	AccountAddress  string            `json:"account_address"`
	EmbeddingType   types.EmbeddingType `json:"embedding_type"`
	Version         uint32            `json:"version"`
	EmbeddingHash   []byte            `json:"embedding_hash"`
	ModelVersion    string            `json:"model_version"`
	ModelHash       string            `json:"model_hash"`
	Dimension       uint32            `json:"dimension"`
	SourceScopeID   string            `json:"source_scope_id"`
	CreatedAt       int64             `json:"created_at"`
	BlockHeight     int64             `json:"block_height"`
	ComputedBy      string            `json:"computed_by"`
	RetentionPolicy *retentionPolicyStore `json:"retention_policy,omitempty"`
	Revoked         bool              `json:"revoked"`
	RevokedAt       *int64            `json:"revoked_at,omitempty"`
	RevokedReason   string            `json:"revoked_reason,omitempty"`
}

// retentionPolicyStore is the storage format for retention policies
type retentionPolicyStore struct {
	Version                   uint32              `json:"version"`
	PolicyID                  string              `json:"policy_id"`
	RetentionType             types.RetentionType `json:"retention_type"`
	DurationSeconds           int64               `json:"duration_seconds,omitempty"`
	BlockCount                int64               `json:"block_count,omitempty"`
	ExpiresAt                 *int64              `json:"expires_at,omitempty"`
	ExpiresAtBlock            *int64              `json:"expires_at_block,omitempty"`
	CreatedAt                 int64               `json:"created_at"`
	CreatedAtBlock            int64               `json:"created_at_block"`
	DeleteOnExpiry            bool                `json:"delete_on_expiry"`
	NotifyBeforeExpirySeconds int64               `json:"notify_before_expiry_seconds,omitempty"`
	ExtensionAllowed          bool                `json:"extension_allowed"`
	MaxExtensions             uint32              `json:"max_extensions"`
	CurrentExtensions         uint32              `json:"current_extensions"`
}

func retentionPolicyToStore(p *types.RetentionPolicy) *retentionPolicyStore {
	if p == nil {
		return nil
	}
	
	store := &retentionPolicyStore{
		Version:                   p.Version,
		PolicyID:                  p.PolicyID,
		RetentionType:             p.RetentionType,
		DurationSeconds:           p.DurationSeconds,
		BlockCount:                p.BlockCount,
		CreatedAt:                 p.CreatedAt.Unix(),
		CreatedAtBlock:            p.CreatedAtBlock,
		DeleteOnExpiry:            p.DeleteOnExpiry,
		NotifyBeforeExpirySeconds: p.NotifyBeforeExpirySeconds,
		ExtensionAllowed:          p.ExtensionAllowed,
		MaxExtensions:             p.MaxExtensions,
		CurrentExtensions:         p.CurrentExtensions,
	}
	
	if p.ExpiresAt != nil {
		ts := p.ExpiresAt.Unix()
		store.ExpiresAt = &ts
	}
	if p.ExpiresAtBlock != nil {
		store.ExpiresAtBlock = p.ExpiresAtBlock
	}
	
	return store
}

func retentionPolicyFromStore(s *retentionPolicyStore) *types.RetentionPolicy {
	if s == nil {
		return nil
	}
	
	p := &types.RetentionPolicy{
		Version:                   s.Version,
		PolicyID:                  s.PolicyID,
		RetentionType:             s.RetentionType,
		DurationSeconds:           s.DurationSeconds,
		BlockCount:                s.BlockCount,
		CreatedAt:                 time.Unix(s.CreatedAt, 0),
		CreatedAtBlock:            s.CreatedAtBlock,
		DeleteOnExpiry:            s.DeleteOnExpiry,
		NotifyBeforeExpirySeconds: s.NotifyBeforeExpirySeconds,
		ExtensionAllowed:          s.ExtensionAllowed,
		MaxExtensions:             s.MaxExtensions,
		CurrentExtensions:         s.CurrentExtensions,
	}
	
	if s.ExpiresAt != nil {
		t := time.Unix(*s.ExpiresAt, 0)
		p.ExpiresAt = &t
	}
	if s.ExpiresAtBlock != nil {
		p.ExpiresAtBlock = s.ExpiresAtBlock
	}
	
	return p
}

// SetEmbeddingEnvelope stores an embedding envelope reference
// SECURITY: Only stores the hash reference, NOT the encrypted payload
func (k Keeper) SetEmbeddingEnvelope(ctx sdk.Context, envelope types.EmbeddingEnvelopeReference) error {
	if err := envelope.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	
	es := embeddingEnvelopeStore{
		EnvelopeID:     envelope.EnvelopeID,
		AccountAddress: envelope.AccountAddress,
		EmbeddingType:  envelope.EmbeddingType,
		Version:        envelope.Version,
		EmbeddingHash:  envelope.EmbeddingHash,
		ModelVersion:   envelope.ModelVersion,
		ModelHash:      envelope.ModelHash,
		Dimension:      envelope.Dimension,
		SourceScopeID:  envelope.SourceScopeID,
		CreatedAt:      envelope.CreatedAt.Unix(),
		BlockHeight:    envelope.BlockHeight,
		ComputedBy:     envelope.ComputedBy,
		RetentionPolicy: retentionPolicyToStore(envelope.RetentionPolicy),
		Revoked:        envelope.Revoked,
		RevokedReason:  envelope.RevokedReason,
	}
	
	if envelope.RevokedAt != nil {
		ts := envelope.RevokedAt.Unix()
		es.RevokedAt = &ts
	}

	bz, err := json.Marshal(&es)
	if err != nil {
		return err
	}

	store.Set(types.EmbeddingEnvelopeKey(envelope.EnvelopeID), bz)
	
	// Also store lookup by account and type
	addr, err := sdk.AccAddressFromBech32(envelope.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}
	
	k.addEnvelopeToAccountIndex(ctx, addr.Bytes(), envelope.EmbeddingType, envelope.EnvelopeID)
	
	return nil
}

// GetEmbeddingEnvelope retrieves an embedding envelope reference by ID
func (k Keeper) GetEmbeddingEnvelope(ctx sdk.Context, envelopeID string) (types.EmbeddingEnvelopeReference, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.EmbeddingEnvelopeKey(envelopeID))
	if bz == nil {
		return types.EmbeddingEnvelopeReference{}, false
	}

	var es embeddingEnvelopeStore
	if err := json.Unmarshal(bz, &es); err != nil {
		return types.EmbeddingEnvelopeReference{}, false
	}

	envelope := types.EmbeddingEnvelopeReference{
		EnvelopeID:      es.EnvelopeID,
		AccountAddress:  es.AccountAddress,
		EmbeddingType:   es.EmbeddingType,
		Version:         es.Version,
		EmbeddingHash:   es.EmbeddingHash,
		ModelVersion:    es.ModelVersion,
		ModelHash:       es.ModelHash,
		Dimension:       es.Dimension,
		SourceScopeID:   es.SourceScopeID,
		CreatedAt:       time.Unix(es.CreatedAt, 0),
		BlockHeight:     es.BlockHeight,
		ComputedBy:      es.ComputedBy,
		RetentionPolicy: retentionPolicyFromStore(es.RetentionPolicy),
		Revoked:         es.Revoked,
		RevokedReason:   es.RevokedReason,
	}
	
	if es.RevokedAt != nil {
		t := time.Unix(*es.RevokedAt, 0)
		envelope.RevokedAt = &t
	}

	return envelope, true
}

// GetEmbeddingEnvelopesByAccount retrieves all envelope references for an account
func (k Keeper) GetEmbeddingEnvelopesByAccount(ctx sdk.Context, address sdk.AccAddress) []types.EmbeddingEnvelopeReference {
	var envelopes []types.EmbeddingEnvelopeReference
	
	ids := k.getEnvelopeIDsForAccount(ctx, address.Bytes())
	for _, id := range ids {
		if envelope, found := k.GetEmbeddingEnvelope(ctx, id); found {
			envelopes = append(envelopes, envelope)
		}
	}
	
	return envelopes
}

// GetEmbeddingEnvelopesByType retrieves envelopes for an account filtered by type
func (k Keeper) GetEmbeddingEnvelopesByType(ctx sdk.Context, address sdk.AccAddress, embeddingType types.EmbeddingType) []types.EmbeddingEnvelopeReference {
	var envelopes []types.EmbeddingEnvelopeReference
	
	allEnvelopes := k.GetEmbeddingEnvelopesByAccount(ctx, address)
	for _, env := range allEnvelopes {
		if env.EmbeddingType == embeddingType {
			envelopes = append(envelopes, env)
		}
	}
	
	return envelopes
}

// RevokeEmbeddingEnvelope revokes an embedding envelope
func (k Keeper) RevokeEmbeddingEnvelope(ctx sdk.Context, envelopeID string, reason string) error {
	envelope, found := k.GetEmbeddingEnvelope(ctx, envelopeID)
	if !found {
		return types.ErrScopeNotFound.Wrapf("envelope not found: %s", envelopeID)
	}
	
	if envelope.Revoked {
		return types.ErrScopeRevoked.Wrapf("envelope already revoked: %s", envelopeID)
	}
	
	revokedAt := ctx.BlockTime()
	envelope.Revoked = true
	envelope.RevokedAt = &revokedAt
	envelope.RevokedReason = reason
	
	return k.SetEmbeddingEnvelope(ctx, envelope)
}

// DeleteEmbeddingEnvelope deletes an embedding envelope reference
func (k Keeper) DeleteEmbeddingEnvelope(ctx sdk.Context, envelopeID string) error {
	envelope, found := k.GetEmbeddingEnvelope(ctx, envelopeID)
	if !found {
		return types.ErrScopeNotFound.Wrapf("envelope not found: %s", envelopeID)
	}
	
	store := ctx.KVStore(k.skey)
	store.Delete(types.EmbeddingEnvelopeKey(envelopeID))
	
	// Remove from account index
	addr, err := sdk.AccAddressFromBech32(envelope.AccountAddress)
	if err == nil {
		k.removeEnvelopeFromAccountIndex(ctx, addr.Bytes(), envelope.EmbeddingType, envelopeID)
	}
	
	return nil
}

// Helper functions for account index management

func (k Keeper) addEnvelopeToAccountIndex(ctx sdk.Context, address []byte, embeddingType types.EmbeddingType, envelopeID string) {
	store := ctx.KVStore(k.skey)
	key := types.EmbeddingEnvelopeByAccountKey(address, embeddingType)
	
	ids := k.getEnvelopeIDsForKey(ctx, key)
	
	// Check if already exists
	for _, id := range ids {
		if id == envelopeID {
			return
		}
	}
	
	ids = append(ids, envelopeID)
	bz, _ := json.Marshal(ids) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, bz)
}

func (k Keeper) removeEnvelopeFromAccountIndex(ctx sdk.Context, address []byte, embeddingType types.EmbeddingType, envelopeID string) {
	store := ctx.KVStore(k.skey)
	key := types.EmbeddingEnvelopeByAccountKey(address, embeddingType)
	
	ids := k.getEnvelopeIDsForKey(ctx, key)
	
	var newIDs []string
	for _, id := range ids {
		if id != envelopeID {
			newIDs = append(newIDs, id)
		}
	}
	
	if len(newIDs) == 0 {
		store.Delete(key)
	} else {
		bz, _ := json.Marshal(newIDs) //nolint:errchkjson // string slice cannot fail to marshal
		store.Set(key, bz)
	}
}

func (k Keeper) getEnvelopeIDsForKey(ctx sdk.Context, key []byte) []string {
	store := ctx.KVStore(k.skey)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}
	
	var ids []string
	if err := json.Unmarshal(bz, &ids); err != nil {
		return nil
	}
	return ids
}

func (k Keeper) getEnvelopeIDsForAccount(ctx sdk.Context, address []byte) []string {
	var allIDs []string
	
	// Get IDs for all embedding types
	for _, embType := range types.AllEmbeddingTypes() {
		key := types.EmbeddingEnvelopeByAccountKey(address, embType)
		ids := k.getEnvelopeIDsForKey(ctx, key)
		allIDs = append(allIDs, ids...)
	}
	
	return allIDs
}

// GetActiveEmbeddingEnvelope returns the most recent active envelope for an account and type
func (k Keeper) GetActiveEmbeddingEnvelope(ctx sdk.Context, address sdk.AccAddress, embeddingType types.EmbeddingType) (types.EmbeddingEnvelopeReference, bool) {
	envelopes := k.GetEmbeddingEnvelopesByType(ctx, address, embeddingType)
	
	var latestActive *types.EmbeddingEnvelopeReference
	for i := range envelopes {
		env := &envelopes[i]
		if !env.Revoked {
			// Check if not expired
			if env.RetentionPolicy != nil && env.RetentionPolicy.IsExpired(ctx.BlockTime()) {
				continue
			}
			if latestActive == nil || env.CreatedAt.After(latestActive.CreatedAt) {
				latestActive = env
			}
		}
	}
	
	if latestActive == nil {
		return types.EmbeddingEnvelopeReference{}, false
	}
	return *latestActive, true
}

// CleanupExpiredEnvelopes removes expired embedding envelopes
func (k Keeper) CleanupExpiredEnvelopes(ctx sdk.Context) int {
	cleaned := 0
	now := ctx.BlockTime()
	
	store := ctx.KVStore(k.skey)
	
	// Iterate over all envelopes and check expiry
	iter := store.Iterator(types.PrefixEmbeddingEnvelope, storetypes.PrefixEndBytes(types.PrefixEmbeddingEnvelope))
	defer iter.Close()
	
	var toDelete []string
	
	for ; iter.Valid(); iter.Next() {
		var es embeddingEnvelopeStore
		if err := json.Unmarshal(iter.Value(), &es); err != nil {
			continue
		}
		
		if es.RetentionPolicy != nil {
			policy := retentionPolicyFromStore(es.RetentionPolicy)
			if policy != nil && policy.IsExpired(now) && policy.DeleteOnExpiry {
				toDelete = append(toDelete, es.EnvelopeID)
			}
		}
	}
	
	// Delete expired envelopes
	for _, id := range toDelete {
		if err := k.DeleteEmbeddingEnvelope(ctx, id); err == nil {
			cleaned++
		}
	}
	
	return cleaned
}
