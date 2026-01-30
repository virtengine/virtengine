package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// SSO Linkage Methods
// ============================================================================

// CreateSSOLinkage creates a new SSO linkage record.
func (k Keeper) CreateSSOLinkage(ctx sdk.Context, msg *types.MsgCreateSSOLinkage) (*types.MsgCreateSSOLinkageResponse, error) {
	// Validate the nonce hasn't been used
	nonceHash := hashNonce(msg.Nonce)
	if k.IsSSONonceUsed(ctx, nonceHash) {
		return nil, types.ErrNonceAlreadyUsed.Wrap("SSO nonce has already been used")
	}

	// Check if linkage already exists for this account and provider
	existingLinkageID := k.GetSSOLinkageByAccountAndProvider(ctx, msg.AccountAddress, msg.Provider)
	if existingLinkageID != "" {
		return nil, types.ErrDuplicateLinkage.Wrapf("SSO linkage already exists: %s", existingLinkageID)
	}

	// Create linkage metadata
	now := ctx.BlockTime()
	linkage := &types.SSOLinkageMetadata{
		Version:          types.SSOVerificationVersion,
		LinkageID:        msg.LinkageID,
		Provider:         msg.Provider,
		Issuer:           msg.Issuer,
		SubjectHash:      msg.SubjectHash,
		Nonce:            msg.Nonce,
		VerifiedAt:       now,
		Status:           types.SSOStatusVerified,
		AccountSignature: msg.AccountSignature,
		EmailDomainHash:  msg.EmailDomainHash,
		OrgIDHash:        msg.TenantIDHash,
	}

	if msg.ExpiresAt != nil {
		linkage.ExpiresAt = msg.ExpiresAt
	}

	// Validate the linkage
	if err := linkage.Validate(); err != nil {
		return nil, err
	}

	// Store the linkage
	if err := k.SetSSOLinkage(ctx, linkage); err != nil {
		return nil, err
	}

	// Index by account and provider
	k.SetSSOLinkageByAccountAndProvider(ctx, msg.AccountAddress, msg.Provider, msg.LinkageID)

	// Mark nonce as used
	nonceRecord := types.NewSSONonceRecord(
		nonceHash,
		msg.AccountAddress,
		msg.Provider,
		msg.Issuer,
		msg.LinkageID,
		now,
		ctx.BlockHeight(),
		24*time.Hour*365, // Keep nonce records for 1 year
	)
	k.SetSSONonceRecord(ctx, nonceRecord)

	// Get score weight for this provider
	scoreContribution := types.GetSSOScoringWeight(msg.Provider)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSSOLinkageCreated,
			sdk.NewAttribute(types.AttributeKeyAccount, msg.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyLinkageID, msg.LinkageID),
			sdk.NewAttribute(types.AttributeKeyProvider, string(msg.Provider)),
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.Issuer),
		),
	)

	k.Logger(ctx).Info("SSO linkage created",
		"account", msg.AccountAddress,
		"linkage_id", msg.LinkageID,
		"provider", msg.Provider,
		"issuer", msg.Issuer,
	)

	return &types.MsgCreateSSOLinkageResponse{
		LinkageID:         msg.LinkageID,
		Status:            types.SSOStatusVerified,
		ScoreContribution: scoreContribution,
		VerifiedAt:        now,
	}, nil
}

// RevokeSSOLinkage revokes an existing SSO linkage.
func (k Keeper) RevokeSSOLinkage(ctx sdk.Context, msg *types.MsgRevokeSSOLinkage) (*types.MsgRevokeSSOLinkageResponse, error) {
	// Get existing linkage
	linkage, found := k.GetSSOLinkage(ctx, msg.LinkageID)
	if !found {
		return nil, types.ErrLinkageNotFound.Wrapf("linkage ID: %s", msg.LinkageID)
	}

	// Verify account owns this linkage
	// We need to check the index since the linkage itself doesn't store the account
	existingLinkageID := k.GetSSOLinkageByAccountAndProvider(ctx, msg.AccountAddress, linkage.Provider)
	if existingLinkageID != msg.LinkageID {
		return nil, types.ErrUnauthorized.Wrap("account does not own this linkage")
	}

	// Update status
	now := ctx.BlockTime()
	linkage.Status = types.SSOStatusRevoked
	linkage.ExpiresAt = &now

	// Store updated linkage
	if err := k.SetSSOLinkage(ctx, linkage); err != nil {
		return nil, err
	}

	// Remove from account index
	k.DeleteSSOLinkageByAccountAndProvider(ctx, msg.AccountAddress, linkage.Provider)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSSOLinkageRevoked,
			sdk.NewAttribute(types.AttributeKeyAccount, msg.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyLinkageID, msg.LinkageID),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	k.Logger(ctx).Info("SSO linkage revoked",
		"account", msg.AccountAddress,
		"linkage_id", msg.LinkageID,
		"reason", msg.Reason,
	)

	return &types.MsgRevokeSSOLinkageResponse{
		LinkageID: msg.LinkageID,
		Status:    types.SSOStatusRevoked,
		RevokedAt: now,
	}, nil
}

// ============================================================================
// SSO Linkage Storage
// ============================================================================

// SetSSOLinkage stores an SSO linkage.
func (k Keeper) SetSSOLinkage(ctx sdk.Context, linkage *types.SSOLinkageMetadata) error {
	store := ctx.KVStore(k.skey)
	key := k.ssoLinkageKey(linkage.LinkageID)

	bz, err := json.Marshal(linkage)
	if err != nil {
		return fmt.Errorf("failed to marshal SSO linkage: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// GetSSOLinkage retrieves an SSO linkage by ID.
func (k Keeper) GetSSOLinkage(ctx sdk.Context, linkageID string) (*types.SSOLinkageMetadata, bool) {
	store := ctx.KVStore(k.skey)
	key := k.ssoLinkageKey(linkageID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var linkage types.SSOLinkageMetadata
	if err := json.Unmarshal(bz, &linkage); err != nil {
		k.Logger(ctx).Error("failed to unmarshal SSO linkage", "error", err)
		return nil, false
	}

	return &linkage, true
}

// DeleteSSOLinkage deletes an SSO linkage.
func (k Keeper) DeleteSSOLinkage(ctx sdk.Context, linkageID string) {
	store := ctx.KVStore(k.skey)
	key := k.ssoLinkageKey(linkageID)
	store.Delete(key)
}

// SetSSOLinkageByAccountAndProvider sets the linkage ID for an account and provider.
func (k Keeper) SetSSOLinkageByAccountAndProvider(ctx sdk.Context, account string, provider types.SSOProviderType, linkageID string) {
	store := ctx.KVStore(k.skey)
	key := k.ssoLinkageByAccountKey(account, provider)
	store.Set(key, []byte(linkageID))
}

// GetSSOLinkageByAccountAndProvider gets the linkage ID for an account and provider.
func (k Keeper) GetSSOLinkageByAccountAndProvider(ctx sdk.Context, account string, provider types.SSOProviderType) string {
	store := ctx.KVStore(k.skey)
	key := k.ssoLinkageByAccountKey(account, provider)

	bz := store.Get(key)
	if bz == nil {
		return ""
	}

	return string(bz)
}

// DeleteSSOLinkageByAccountAndProvider removes the linkage index for an account and provider.
func (k Keeper) DeleteSSOLinkageByAccountAndProvider(ctx sdk.Context, account string, provider types.SSOProviderType) {
	store := ctx.KVStore(k.skey)
	key := k.ssoLinkageByAccountKey(account, provider)
	store.Delete(key)
}

// GetSSOLinkagesForAccount returns all SSO linkages for an account.
func (k Keeper) GetSSOLinkagesForAccount(ctx sdk.Context, account string) []*types.SSOLinkageMetadata {
	linkages := make([]*types.SSOLinkageMetadata, 0)

	for _, provider := range types.AllSSOProviderTypes() {
		linkageID := k.GetSSOLinkageByAccountAndProvider(ctx, account, provider)
		if linkageID != "" {
			if linkage, found := k.GetSSOLinkage(ctx, linkageID); found {
				linkages = append(linkages, linkage)
			}
		}
	}

	return linkages
}

// IterateSSOLinkages iterates over all SSO linkages.
func (k Keeper) IterateSSOLinkages(ctx sdk.Context, fn func(linkage *types.SSOLinkageMetadata) bool) {
	store := ctx.KVStore(k.skey)
	prefixStore := prefix.NewStore(store, types.PrefixSSOLinkage)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var linkage types.SSOLinkageMetadata
		if err := json.Unmarshal(iter.Value(), &linkage); err != nil {
			continue
		}
		if fn(&linkage) {
			break
		}
	}
}

// ============================================================================
// SSO Nonce Tracking
// ============================================================================

// SetSSONonceRecord stores an SSO nonce record.
func (k Keeper) SetSSONonceRecord(ctx sdk.Context, record *types.SSONonceRecord) {
	store := ctx.KVStore(k.skey)
	key := k.ssoNonceKey(record.NonceHash)

	bz, err := json.Marshal(record)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal SSO nonce record", "error", err)
		return
	}

	store.Set(key, bz)

	// Set indexes
	k.setSSONonceByAccount(ctx, record.AccountAddress, record.NonceHash)
	k.setSSONonceByProvider(ctx, record.Provider, record.NonceHash)
	k.setSSONonceExpiry(ctx, record.ExpiresAt, record.NonceHash)
}

// GetSSONonceRecord retrieves an SSO nonce record.
func (k Keeper) GetSSONonceRecord(ctx sdk.Context, nonceHash string) (*types.SSONonceRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := k.ssoNonceKey(nonceHash)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var record types.SSONonceRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}

	return &record, true
}

// IsSSONonceUsed checks if an SSO nonce has been used.
func (k Keeper) IsSSONonceUsed(ctx sdk.Context, nonceHash string) bool {
	store := ctx.KVStore(k.skey)
	key := k.ssoNonceKey(nonceHash)
	return store.Has(key)
}

// setSSONonceByAccount sets the nonce index by account.
func (k Keeper) setSSONonceByAccount(ctx sdk.Context, account, nonceHash string) {
	store := ctx.KVStore(k.skey)
	key := append(append(types.PrefixSSONonceByAccount, []byte(account)...), []byte(nonceHash)...)
	store.Set(key, []byte{1})
}

// setSSONonceByProvider sets the nonce index by provider.
func (k Keeper) setSSONonceByProvider(ctx sdk.Context, provider types.SSOProviderType, nonceHash string) {
	store := ctx.KVStore(k.skey)
	key := append(append(types.PrefixSSONonceByProvider, []byte(provider)...), []byte(nonceHash)...)
	store.Set(key, []byte{1})
}

// setSSONonceExpiry sets the nonce expiry index.
func (k Keeper) setSSONonceExpiry(ctx sdk.Context, expiresAt time.Time, nonceHash string) {
	store := ctx.KVStore(k.skey)
	expiryBytes := sdk.FormatTimeBytes(expiresAt)
	key := append(append(types.PrefixSSONonceExpiry, expiryBytes...), []byte(nonceHash)...)
	store.Set(key, []byte{1})
}

// PruneExpiredSSONonces removes expired SSO nonce records.
func (k Keeper) PruneExpiredSSONonces(ctx sdk.Context) int {
	store := ctx.KVStore(k.skey)
	now := ctx.BlockTime()
	cutoffBytes := sdk.FormatTimeBytes(now)

	prefixStore := prefix.NewStore(store, types.PrefixSSONonceExpiry)
	iter := prefixStore.Iterator(nil, cutoffBytes)
	defer iter.Close()

	toDelete := make([][]byte, 0)
	for ; iter.Valid(); iter.Next() {
		toDelete = append(toDelete, iter.Key())
	}

	for _, key := range toDelete {
		// Extract nonce hash from key (after the timestamp prefix)
		if len(key) > 14 { // Timestamp is 14 bytes
			nonceHash := string(key[14:])

			// Get the full record to clean up indexes
			if record, found := k.GetSSONonceRecord(ctx, nonceHash); found {
				k.deleteSSONonceRecord(ctx, record)
			}
		}
	}

	return len(toDelete)
}

// deleteSSONonceRecord removes an SSO nonce record and its indexes.
func (k Keeper) deleteSSONonceRecord(ctx sdk.Context, record *types.SSONonceRecord) {
	store := ctx.KVStore(k.skey)

	// Delete main record
	store.Delete(k.ssoNonceKey(record.NonceHash))

	// Delete indexes
	accountKey := append(append(types.PrefixSSONonceByAccount, []byte(record.AccountAddress)...), []byte(record.NonceHash)...)
	store.Delete(accountKey)

	providerKey := append(append(types.PrefixSSONonceByProvider, []byte(record.Provider)...), []byte(record.NonceHash)...)
	store.Delete(providerKey)

	expiryBytes := sdk.FormatTimeBytes(record.ExpiresAt)
	expiryKey := append(append(types.PrefixSSONonceExpiry, expiryBytes...), []byte(record.NonceHash)...)
	store.Delete(expiryKey)
}

// ============================================================================
// Key Functions
// ============================================================================

func (k Keeper) ssoLinkageKey(linkageID string) []byte {
	return append(types.PrefixSSOLinkage, []byte(linkageID)...)
}

func (k Keeper) ssoLinkageByAccountKey(account string, provider types.SSOProviderType) []byte {
	return append(append(types.PrefixSSOLinkageByAccount, []byte(account)...), []byte(provider)...)
}

func (k Keeper) ssoNonceKey(nonceHash string) []byte {
	return append(types.PrefixSSONonce, []byte(nonceHash)...)
}

// hashNonce creates a SHA256 hash of a nonce.
func hashNonce(nonce string) string {
	hash := sha256.Sum256([]byte(nonce))
	return hex.EncodeToString(hash[:])
}

// Suppress unused import warning for storetypes
var _ storetypes.StoreKey
