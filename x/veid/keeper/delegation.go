package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Identity Delegation Keeper Methods (VE-3024)
// ============================================================================

// delegationStore is the stored format of a delegation record
type delegationStore struct {
	DelegationID     string   `json:"delegation_id"`
	DelegatorAddress string   `json:"delegator_address"`
	DelegateAddress  string   `json:"delegate_address"`
	Permissions      []int    `json:"permissions"`
	ExpiresAt        int64    `json:"expires_at"`
	MaxUses          uint64   `json:"max_uses"`
	UsesRemaining    uint64   `json:"uses_remaining"`
	CreatedAt        int64    `json:"created_at"`
	RevokedAt        *int64   `json:"revoked_at,omitempty"`
	Status           int      `json:"status"`
	RevocationReason string   `json:"revocation_reason,omitempty"`
	Metadata         []string `json:"metadata,omitempty"` // key=value pairs
}

// delegationToStore converts a delegation record to stored format
func delegationToStore(d *types.DelegationRecord) *delegationStore {
	ds := &delegationStore{
		DelegationID:     d.DelegationID,
		DelegatorAddress: d.DelegatorAddress,
		DelegateAddress:  d.DelegateAddress,
		ExpiresAt:        d.ExpiresAt.Unix(),
		MaxUses:          d.MaxUses,
		UsesRemaining:    d.UsesRemaining,
		CreatedAt:        d.CreatedAt.Unix(),
		Status:           int(d.Status),
		RevocationReason: d.RevocationReason,
	}

	// Convert permissions to ints
	ds.Permissions = make([]int, len(d.Permissions))
	for i, p := range d.Permissions {
		ds.Permissions[i] = int(p)
	}

	// Convert revoked time
	if d.RevokedAt != nil {
		revokedUnix := d.RevokedAt.Unix()
		ds.RevokedAt = &revokedUnix
	}

	// Convert metadata to key=value pairs
	if len(d.Metadata) > 0 {
		ds.Metadata = make([]string, 0, len(d.Metadata))
		for k, v := range d.Metadata {
			ds.Metadata = append(ds.Metadata, k+"="+v)
		}
	}

	return ds
}

// delegationFromStore converts stored format back to delegation record
func delegationFromStore(ds *delegationStore) *types.DelegationRecord {
	d := &types.DelegationRecord{
		DelegationID:     ds.DelegationID,
		DelegatorAddress: ds.DelegatorAddress,
		DelegateAddress:  ds.DelegateAddress,
		ExpiresAt:        time.Unix(ds.ExpiresAt, 0).UTC(),
		MaxUses:          ds.MaxUses,
		UsesRemaining:    ds.UsesRemaining,
		CreatedAt:        time.Unix(ds.CreatedAt, 0).UTC(),
		Status:           types.DelegationStatus(ds.Status),
		RevocationReason: ds.RevocationReason,
		Metadata:         make(map[string]string),
	}

	// Convert permissions from ints
	d.Permissions = make([]types.DelegationPermission, len(ds.Permissions))
	for i, p := range ds.Permissions {
		d.Permissions[i] = types.DelegationPermission(p)
	}

	// Convert revoked time
	if ds.RevokedAt != nil {
		revokedTime := time.Unix(*ds.RevokedAt, 0).UTC()
		d.RevokedAt = &revokedTime
	}

	// Convert metadata from key=value pairs
	for _, kv := range ds.Metadata {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				d.Metadata[kv[:i]] = kv[i+1:]
				break
			}
		}
	}

	return d
}

// GenerateDelegationID generates a unique delegation ID based on delegator, delegate, and timestamp
func GenerateDelegationID(delegatorAddress, delegateAddress string, createdAt time.Time) string {
	data := fmt.Sprintf("%s:%s:%d", delegatorAddress, delegateAddress, createdAt.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return "del_" + hex.EncodeToString(hash[:16])
}

// CreateDelegation creates a new delegation from delegator to delegate with specified permissions
func (k Keeper) CreateDelegation(
	ctx sdk.Context,
	delegatorAddress sdk.AccAddress,
	delegateAddress sdk.AccAddress,
	permissions []types.DelegationPermission,
	expiresAt time.Time,
	maxUses uint64,
) (*types.DelegationRecord, error) {
	if len(permissions) == 0 {
		return nil, types.ErrInvalidDelegation.Wrap("at least one permission is required")
	}

	// Validate addresses are different
	if delegatorAddress.Equals(delegateAddress) {
		return nil, types.ErrInvalidDelegation.Wrap("cannot delegate to self")
	}

	now := ctx.BlockTime()

	// Validate expiration
	if !expiresAt.After(now) {
		return nil, types.ErrInvalidDelegation.Wrap("expires_at must be in the future")
	}

	// Generate delegation ID
	delegationID := GenerateDelegationID(delegatorAddress.String(), delegateAddress.String(), now)

	// Check if delegation already exists
	if _, found := k.GetDelegation(ctx, delegationID); found {
		return nil, types.ErrDelegationAlreadyExists.Wrap("delegation with this ID already exists")
	}

	// Create delegation record
	delegation := types.NewDelegationRecord(
		delegationID,
		delegatorAddress.String(),
		delegateAddress.String(),
		permissions,
		expiresAt,
		maxUses,
		now,
	)

	// Validate the delegation
	if err := delegation.Validate(); err != nil {
		return nil, err
	}

	// Store delegation
	if err := k.setDelegation(ctx, delegation); err != nil {
		return nil, err
	}

	// Create indexes
	k.setDelegationIndexes(ctx, delegation)

	k.Logger(ctx).Info("created delegation",
		"delegation_id", delegationID,
		"delegator", delegatorAddress.String(),
		"delegate", delegateAddress.String(),
		"permissions", len(permissions),
		"expires_at", expiresAt.Format(time.RFC3339),
		"max_uses", maxUses,
	)

	return delegation, nil
}

// setDelegation stores a delegation record
func (k Keeper) setDelegation(ctx sdk.Context, delegation *types.DelegationRecord) error {
	store := ctx.KVStore(k.skey)

	ds := delegationToStore(delegation)
	data, err := json.Marshal(ds)
	if err != nil {
		return types.ErrInvalidDelegation.Wrapf("failed to marshal delegation: %v", err)
	}

	key := types.DelegationKey(delegation.DelegationID)
	store.Set(key, data)
	return nil
}

// setDelegationIndexes creates indexes for a delegation
func (k Keeper) setDelegationIndexes(ctx sdk.Context, delegation *types.DelegationRecord) {
	store := ctx.KVStore(k.skey)

	// Index by delegator
	delegatorAddr, _ := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
	delegatorKey := types.DelegationByDelegatorKey(delegatorAddr, delegation.DelegationID)
	store.Set(delegatorKey, []byte{0x01})

	// Index by delegate
	delegateAddr, _ := sdk.AccAddressFromBech32(delegation.DelegateAddress)
	delegateKey := types.DelegationByDelegateKey(delegateAddr, delegation.DelegationID)
	store.Set(delegateKey, []byte{0x01})

	// Index by expiry time
	expiryKey := types.DelegationExpiryKey(delegation.ExpiresAt.Unix(), delegation.DelegationID)
	store.Set(expiryKey, []byte{0x01})
}

// deleteDelegationIndexes removes indexes for a delegation
func (k Keeper) deleteDelegationIndexes(ctx sdk.Context, delegation *types.DelegationRecord) {
	store := ctx.KVStore(k.skey)

	// Remove delegator index
	delegatorAddr, _ := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
	delegatorKey := types.DelegationByDelegatorKey(delegatorAddr, delegation.DelegationID)
	store.Delete(delegatorKey)

	// Remove delegate index
	delegateAddr, _ := sdk.AccAddressFromBech32(delegation.DelegateAddress)
	delegateKey := types.DelegationByDelegateKey(delegateAddr, delegation.DelegationID)
	store.Delete(delegateKey)

	// Remove expiry index
	expiryKey := types.DelegationExpiryKey(delegation.ExpiresAt.Unix(), delegation.DelegationID)
	store.Delete(expiryKey)
}

// GetDelegation retrieves a delegation by ID
func (k Keeper) GetDelegation(ctx sdk.Context, delegationID string) (*types.DelegationRecord, bool) {
	store := ctx.KVStore(k.skey)

	key := types.DelegationKey(delegationID)
	data := store.Get(key)
	if data == nil {
		return nil, false
	}

	var ds delegationStore
	if err := json.Unmarshal(data, &ds); err != nil {
		k.Logger(ctx).Error("failed to unmarshal delegation", "error", err, "delegation_id", delegationID)
		return nil, false
	}

	delegation := delegationFromStore(&ds)

	// Update status based on current time
	delegation.UpdateStatus(ctx.BlockTime())

	return delegation, true
}

// RevokeDelegation revokes a delegation by the delegator
func (k Keeper) RevokeDelegation(
	ctx sdk.Context,
	delegatorAddress sdk.AccAddress,
	delegationID string,
	reason string,
) error {
	delegation, found := k.GetDelegation(ctx, delegationID)
	if !found {
		return types.ErrDelegationNotFound.Wrapf("delegation %s not found", delegationID)
	}

	// Verify the caller is the delegator
	if delegation.DelegatorAddress != delegatorAddress.String() {
		return types.ErrDelegationUnauthorized.Wrap("only the delegator can revoke a delegation")
	}

	// Revoke the delegation
	now := ctx.BlockTime()
	if err := delegation.Revoke(now, reason); err != nil {
		return err
	}

	// Update in store
	if err := k.setDelegation(ctx, delegation); err != nil {
		return err
	}

	k.Logger(ctx).Info("revoked delegation",
		"delegation_id", delegationID,
		"delegator", delegatorAddress.String(),
		"reason", reason,
	)

	return nil
}

// ListDelegationsForDelegator returns all delegations granted by a delegator
func (k Keeper) ListDelegationsForDelegator(
	ctx sdk.Context,
	delegatorAddress sdk.AccAddress,
	activeOnly bool,
) ([]*types.DelegationRecord, error) {
	store := ctx.KVStore(k.skey)
	prefix := types.DelegationByDelegatorPrefixKey(delegatorAddress)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var delegations []*types.DelegationRecord
	now := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		// Extract delegation ID from key
		keyBytes := iterator.Key()
		delegationID := string(keyBytes[len(prefix):])

		delegation, found := k.GetDelegation(ctx, delegationID)
		if !found {
			continue
		}

		if activeOnly && !delegation.IsActive(now) {
			continue
		}

		delegations = append(delegations, delegation)
	}

	return delegations, nil
}

// ListDelegationsForDelegate returns all delegations received by a delegate
func (k Keeper) ListDelegationsForDelegate(
	ctx sdk.Context,
	delegateAddress sdk.AccAddress,
	activeOnly bool,
) ([]*types.DelegationRecord, error) {
	store := ctx.KVStore(k.skey)
	prefix := types.DelegationByDelegatePrefixKey(delegateAddress)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var delegations []*types.DelegationRecord
	now := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		// Extract delegation ID from key
		keyBytes := iterator.Key()
		delegationID := string(keyBytes[len(prefix):])

		delegation, found := k.GetDelegation(ctx, delegationID)
		if !found {
			continue
		}

		if activeOnly && !delegation.IsActive(now) {
			continue
		}

		delegations = append(delegations, delegation)
	}

	return delegations, nil
}

// ValidateDelegation checks if a delegation is valid and active for the given permission
func (k Keeper) ValidateDelegation(
	ctx sdk.Context,
	delegationID string,
	requiredPermission types.DelegationPermission,
) *types.DelegationValidationResult {
	delegation, found := k.GetDelegation(ctx, delegationID)
	if !found {
		return &types.DelegationValidationResult{
			Valid:  false,
			Reason: "delegation not found",
		}
	}

	now := ctx.BlockTime()

	// Check if delegation is active
	if !delegation.IsActive(now) {
		reason := fmt.Sprintf("delegation is %s", delegation.Status.String())
		if delegation.Status == types.DelegationActive && now.After(delegation.ExpiresAt) {
			reason = "delegation has expired"
		}
		return &types.DelegationValidationResult{
			Valid:      false,
			Reason:     reason,
			Delegation: delegation,
		}
	}

	// Check if delegation has required permission
	if !delegation.HasPermission(requiredPermission) {
		return &types.DelegationValidationResult{
			Valid:      false,
			Reason:     fmt.Sprintf("delegation does not grant %s permission", requiredPermission.String()),
			Delegation: delegation,
		}
	}

	return &types.DelegationValidationResult{
		Valid:      true,
		Delegation: delegation,
	}
}

// UseDelegation decrements the uses remaining and validates the permission
// Returns an error if the delegation is invalid or lacks the required permission
func (k Keeper) UseDelegation(
	ctx sdk.Context,
	delegationID string,
	requiredPermission types.DelegationPermission,
) error {
	delegation, found := k.GetDelegation(ctx, delegationID)
	if !found {
		return types.ErrDelegationNotFound.Wrapf("delegation %s not found", delegationID)
	}

	now := ctx.BlockTime()

	// Update status first
	delegation.UpdateStatus(now)

	// Check if delegation is active
	if !delegation.IsActive(now) {
		switch delegation.Status {
		case types.DelegationExpired:
			return types.ErrDelegationExpired.Wrap("delegation has expired")
		case types.DelegationRevoked:
			return types.ErrDelegationRevoked.Wrap("delegation has been revoked")
		case types.DelegationExhausted:
			return types.ErrDelegationExhausted.Wrap("delegation has no uses remaining")
		default:
			return types.ErrInvalidDelegation.Wrapf("delegation is not active: %s", delegation.Status.String())
		}
	}

	// Check permission
	if !delegation.HasPermission(requiredPermission) {
		return types.ErrDelegationPermissionDenied.Wrapf(
			"delegation does not grant %s permission",
			requiredPermission.String(),
		)
	}

	// Use the delegation
	if !delegation.UseOnce(now) {
		return types.ErrDelegationExhausted.Wrap("failed to use delegation")
	}

	// Update in store
	if err := k.setDelegation(ctx, delegation); err != nil {
		return err
	}

	k.Logger(ctx).Debug("used delegation",
		"delegation_id", delegationID,
		"permission", requiredPermission.String(),
		"uses_remaining", delegation.UsesRemaining,
	)

	return nil
}

// GetDelegationForDelegate retrieves an active delegation between delegator and delegate
func (k Keeper) GetDelegationForDelegate(
	ctx sdk.Context,
	delegatorAddress sdk.AccAddress,
	delegateAddress sdk.AccAddress,
	requiredPermission types.DelegationPermission,
) (*types.DelegationRecord, error) {
	// Get all delegations for the delegate
	delegations, err := k.ListDelegationsForDelegate(ctx, delegateAddress, true)
	if err != nil {
		return nil, err
	}

	now := ctx.BlockTime()

	// Find a valid delegation from the delegator
	for _, d := range delegations {
		if d.DelegatorAddress != delegatorAddress.String() {
			continue
		}

		if !d.IsActive(now) {
			continue
		}

		if d.HasPermission(requiredPermission) {
			return d, nil
		}
	}

	return nil, types.ErrDelegationNotFound.Wrap("no valid delegation found with required permission")
}

// WithDelegations iterates over all delegations
func (k Keeper) WithDelegations(ctx sdk.Context, fn func(delegation *types.DelegationRecord) bool) {
	store := ctx.KVStore(k.skey)
	prefix := types.DelegationPrefixKey()
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var ds delegationStore
		if err := json.Unmarshal(iterator.Value(), &ds); err != nil {
			k.Logger(ctx).Error("failed to unmarshal delegation in iterator", "error", err)
			continue
		}

		delegation := delegationFromStore(&ds)
		delegation.UpdateStatus(ctx.BlockTime())

		if fn(delegation) {
			break
		}
	}
}

// ExpireDelegations updates status for all expired delegations
// This should be called in EndBlock to clean up expired delegations
func (k Keeper) ExpireDelegations(ctx sdk.Context) error {
	store := ctx.KVStore(k.skey)
	now := ctx.BlockTime()

	// Iterate over expiry index for all delegations that have expired
	prefix := types.DelegationExpiryPrefixKey()
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var expiredIDs []string

	for ; iterator.Valid(); iterator.Next() {
		keyBytes := iterator.Key()
		// Key format: prefix | expiry_timestamp | '/' | delegation_id
		// We need to parse out the timestamp and check if it's before now

		// Skip prefix
		keyWithoutPrefix := keyBytes[len(prefix):]

		// First 8 bytes are the timestamp
		if len(keyWithoutPrefix) < 9 {
			continue
		}

		// Decode timestamp (big-endian int64)
		var expiryTimestamp int64
		for i := 0; i < 8; i++ {
			expiryTimestamp = (expiryTimestamp << 8) | int64(keyWithoutPrefix[i])
		}

		// If this delegation hasn't expired yet, we can stop
		// (since keys are sorted by expiry time)
		if expiryTimestamp > now.Unix() {
			break
		}

		// Extract delegation ID (after the '/' separator)
		if keyWithoutPrefix[8] != '/' {
			continue
		}
		delegationID := string(keyWithoutPrefix[9:])
		expiredIDs = append(expiredIDs, delegationID)
	}

	// Update expired delegations
	for _, delegationID := range expiredIDs {
		delegation, found := k.GetDelegation(ctx, delegationID)
		if !found {
			continue
		}

		if delegation.Status == types.DelegationActive {
			delegation.Status = types.DelegationExpired
			if err := k.setDelegation(ctx, delegation); err != nil {
				k.Logger(ctx).Error("failed to update expired delegation", "error", err, "delegation_id", delegationID)
				continue
			}
		}
	}

	if len(expiredIDs) > 0 {
		k.Logger(ctx).Info("expired delegations", "count", len(expiredIDs))
	}

	return nil
}

// DeleteDelegation permanently removes a delegation
// This should only be used for cleanup of old delegations
func (k Keeper) DeleteDelegation(ctx sdk.Context, delegationID string) error {
	delegation, found := k.GetDelegation(ctx, delegationID)
	if !found {
		return types.ErrDelegationNotFound.Wrapf("delegation %s not found", delegationID)
	}

	// Only allow deletion of terminal delegations
	if !delegation.Status.IsTerminal() {
		return types.ErrInvalidDelegation.Wrap("cannot delete active delegation")
	}

	store := ctx.KVStore(k.skey)

	// Delete indexes
	k.deleteDelegationIndexes(ctx, delegation)

	// Delete the delegation record
	key := types.DelegationKey(delegationID)
	store.Delete(key)

	k.Logger(ctx).Info("deleted delegation", "delegation_id", delegationID)

	return nil
}
