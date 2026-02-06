// Package keeper provides the VEID module keeper.
//
// This file implements rate limiting for VEID message handlers to prevent DoS attacks.
// It provides per-account and per-block rate limiting for sensitive operations.
//
// Task Reference: 22A - Pre-mainnet security hardening
package keeper

import (
	"encoding/binary"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Rate Limit Constants
// ============================================================================

const (
	// MaxUploadsPerAccountPerBlock limits scope uploads per account per block
	MaxUploadsPerAccountPerBlock uint32 = 3

	// MaxUploadsPerBlock limits total scope uploads per block
	MaxUploadsPerBlock uint32 = 50

	// MaxVerificationRequestsPerAccountPerBlock limits verification requests
	MaxVerificationRequestsPerAccountPerBlock uint32 = 5

	// MaxScoreUpdatesPerBlock limits total score updates from all validators per block
	MaxScoreUpdatesPerBlock uint32 = 100

	// AccountCooldownBlocks is the minimum blocks between operations for an account
	AccountCooldownBlocks int64 = 2
)

// ============================================================================
// Rate Limit Store Keys
// ============================================================================

// Rate limit key helpers using dedicated prefixes from types/keys.go
func accountUploadCountKey(address sdk.AccAddress, height int64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, safeUint64FromIntBiometricToUint64(height))
	key := make([]byte, 0, len(types.PrefixMsgRateLimit)+len(address)+8+1)
	key = append(key, types.PrefixMsgRateLimit...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, bz...)
	return key
}

func blockUploadCountKey(height int64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, safeUint64FromIntBiometricToUint64(height))
	key := make([]byte, 0, len(types.PrefixMsgRateLimitBlock)+8)
	key = append(key, types.PrefixMsgRateLimitBlock...)
	key = append(key, bz...)
	return key
}

func accountLastOpKey(address sdk.AccAddress) []byte {
	key := make([]byte, 0, len(types.PrefixMsgRateLimitCooldown)+len(address))
	key = append(key, types.PrefixMsgRateLimitCooldown...)
	key = append(key, address...)
	return key
}

// safeUint64FromIntBiometricToUint64 safely converts int64 to uint64
func safeUint64FromIntBiometricToUint64(v int64) uint64 {
	if v < 0 {
		return 0
	}
	//nolint:gosec // range checked above
	return uint64(v)
}

// ============================================================================
// Rate Limit Checking
// ============================================================================

// CheckUploadRateLimit verifies that upload rate limits are not exceeded
func (k Keeper) CheckUploadRateLimit(ctx sdk.Context, address sdk.AccAddress) error {
	currentHeight := ctx.BlockHeight()

	// Check per-account per-block limit
	accountCount := k.getAccountUploadCount(ctx, address, currentHeight)
	if accountCount >= MaxUploadsPerAccountPerBlock {
		return types.ErrRateLimitExceeded.Wrapf(
			"account %s has reached the upload limit of %d per block",
			address.String(), MaxUploadsPerAccountPerBlock,
		)
	}

	// Check per-block global limit
	blockCount := k.getBlockUploadCount(ctx, currentHeight)
	if blockCount >= MaxUploadsPerBlock {
		return types.ErrRateLimitExceeded.Wrapf(
			"block %d has reached the upload limit of %d",
			currentHeight, MaxUploadsPerBlock,
		)
	}

	// Check account cooldown
	if err := k.checkAccountCooldown(ctx, address); err != nil {
		return err
	}

	return nil
}

// CheckVerificationRequestRateLimit verifies that verification request limits are not exceeded
func (k Keeper) CheckVerificationRequestRateLimit(ctx sdk.Context, address sdk.AccAddress) error {
	currentHeight := ctx.BlockHeight()

	accountCount := k.getAccountUploadCount(ctx, address, currentHeight)
	if accountCount >= MaxVerificationRequestsPerAccountPerBlock {
		return types.ErrRateLimitExceeded.Wrapf(
			"account %s has reached the verification request limit of %d per block",
			address.String(), MaxVerificationRequestsPerAccountPerBlock,
		)
	}

	return nil
}

// CheckScoreUpdateRateLimit verifies that score update limits are not exceeded
func (k Keeper) CheckScoreUpdateRateLimit(ctx sdk.Context) error {
	currentHeight := ctx.BlockHeight()

	blockCount := k.getBlockUploadCount(ctx, currentHeight)
	if blockCount >= MaxScoreUpdatesPerBlock {
		return types.ErrRateLimitExceeded.Wrapf(
			"block %d has reached the score update limit of %d",
			currentHeight, MaxScoreUpdatesPerBlock,
		)
	}

	return nil
}

// RecordUpload records an upload for rate limiting
func (k Keeper) RecordUpload(ctx sdk.Context, address sdk.AccAddress) {
	currentHeight := ctx.BlockHeight()
	k.incrementAccountUploadCount(ctx, address, currentHeight)
	k.incrementBlockUploadCount(ctx, currentHeight)
	k.recordAccountLastOp(ctx, address, currentHeight)
}

// ============================================================================
// Rate Limit Store Operations
// ============================================================================

func (k Keeper) getAccountUploadCount(ctx sdk.Context, address sdk.AccAddress, height int64) uint32 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(accountUploadCountKey(address, height))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint32(bz)
}

func (k Keeper) incrementAccountUploadCount(ctx sdk.Context, address sdk.AccAddress, height int64) {
	count := k.getAccountUploadCount(ctx, address, height) + 1
	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, count)
	store := ctx.KVStore(k.skey)
	store.Set(accountUploadCountKey(address, height), bz)
}

func (k Keeper) getBlockUploadCount(ctx sdk.Context, height int64) uint32 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(blockUploadCountKey(height))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint32(bz)
}

func (k Keeper) incrementBlockUploadCount(ctx sdk.Context, height int64) {
	count := k.getBlockUploadCount(ctx, height) + 1
	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, count)
	store := ctx.KVStore(k.skey)
	store.Set(blockUploadCountKey(height), bz)
}

func (k Keeper) checkAccountCooldown(ctx sdk.Context, address sdk.AccAddress) error {
	store := ctx.KVStore(k.skey)
	bz := store.Get(accountLastOpKey(address))
	if bz == nil {
		return nil
	}

	rawHeight := binary.BigEndian.Uint64(bz)
	//nolint:gosec // block heights are always positive and bounded by int64 max
	lastHeight := int64(rawHeight)
	currentHeight := ctx.BlockHeight()
	blocksSinceLastOp := currentHeight - lastHeight

	if blocksSinceLastOp < AccountCooldownBlocks {
		return types.ErrRateLimitExceeded.Wrapf(
			"account must wait %d more blocks before next operation",
			AccountCooldownBlocks-blocksSinceLastOp,
		)
	}

	return nil
}

func (k Keeper) recordAccountLastOp(ctx sdk.Context, address sdk.AccAddress, height int64) {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, safeUint64FromIntBiometricToUint64(height))
	store := ctx.KVStore(k.skey)
	store.Set(accountLastOpKey(address), bz)
}

// CleanupOldRateLimitData removes stale rate limit entries to prevent unbounded growth
func (k Keeper) CleanupOldRateLimitData(ctx sdk.Context) {
	currentHeight := ctx.BlockHeight()
	cutoffHeight := currentHeight - 1000
	if cutoffHeight <= 0 {
		return
	}

	store := ctx.KVStore(k.skey)

	// Clean up block-level rate limit data
	iter := store.Iterator(types.PrefixMsgRateLimitBlock, blockUploadCountKey(cutoffHeight))
	defer iter.Close()

	keysToDelete := [][]byte{}
	for ; iter.Valid(); iter.Next() {
		keysToDelete = append(keysToDelete, iter.Key())
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// ============================================================================
// Msg Handler Input Validation
// ============================================================================

// MsgInputLimits defines maximum input sizes for message fields
type MsgInputLimits struct {
	MaxScopeIDLength         int `json:"max_scope_id_length"`
	MaxReasonLength          int `json:"max_reason_length"`
	MaxClientIDLength        int `json:"max_client_id_length"`
	MaxDeviceFingerprintSize int `json:"max_device_fingerprint_size"`
	MaxSaltSize              int `json:"max_salt_size"`
	MaxSignatureSize         int `json:"max_signature_size"`
	MaxPayloadHashSize       int `json:"max_payload_hash_size"`
	MaxGeoHintLength         int `json:"max_geo_hint_length"`
	MaxPurposeLength         int `json:"max_purpose_length"`
}

// DefaultMsgInputLimits returns the default input limits
func DefaultMsgInputLimits() MsgInputLimits {
	return MsgInputLimits{
		MaxScopeIDLength:         128,
		MaxReasonLength:          512,
		MaxClientIDLength:        64,
		MaxDeviceFingerprintSize: 256,
		MaxSaltSize:              128,
		MaxSignatureSize:         512,
		MaxPayloadHashSize:       64,
		MaxGeoHintLength:         128,
		MaxPurposeLength:         256,
	}
}

// ValidateMsgInputLimits validates message field sizes against limits
func ValidateMsgInputLimits(limits MsgInputLimits, fields map[string]int) error {
	checks := map[string]struct {
		actual int
		max    int
	}{
		"scope_id":           {fields["scope_id"], limits.MaxScopeIDLength},
		"reason":             {fields["reason"], limits.MaxReasonLength},
		"client_id":          {fields["client_id"], limits.MaxClientIDLength},
		"device_fingerprint": {fields["device_fingerprint"], limits.MaxDeviceFingerprintSize},
		"salt":               {fields["salt"], limits.MaxSaltSize},
		"signature":          {fields["signature"], limits.MaxSignatureSize},
		"payload_hash":       {fields["payload_hash"], limits.MaxPayloadHashSize},
		"geo_hint":           {fields["geo_hint"], limits.MaxGeoHintLength},
		"purpose":            {fields["purpose"], limits.MaxPurposeLength},
	}

	for name, check := range checks {
		if check.actual > 0 && check.max > 0 && check.actual > check.max {
			return types.ErrInputTooLarge.Wrapf(
				"field %s exceeds maximum size: %d > %d",
				name, check.actual, check.max,
			)
		}
	}

	return nil
}

// ============================================================================
// Privilege Escalation Prevention
// ============================================================================

// PrivilegeAuditRecord records a privilege check for audit purposes
type PrivilegeAuditRecord struct {
	// Operation is the operation being attempted
	Operation string `json:"operation"`

	// Account is the account attempting the operation
	Account string `json:"account"`

	// RequiredPrivilege is the minimum privilege required
	RequiredPrivilege string `json:"required_privilege"`

	// Granted indicates if access was granted
	Granted bool `json:"granted"`

	// BlockHeight is when the check occurred
	BlockHeight int64 `json:"block_height"`

	// Reason is the denial reason if not granted
	Reason string `json:"reason,omitempty"`
}

// ValidatePrivilegedOperation performs a comprehensive privilege check
// and records the result for audit purposes
func (k Keeper) ValidatePrivilegedOperation(
	ctx sdk.Context,
	sender sdk.AccAddress,
	operation string,
	requireValidator bool,
	requireAuthority bool,
) error {
	if requireAuthority {
		if sender.String() != k.authority {
			k.recordPrivilegeAudit(ctx, PrivilegeAuditRecord{
				Operation:         operation,
				Account:           sender.String(),
				RequiredPrivilege: "governance_authority",
				Granted:           false,
				BlockHeight:       ctx.BlockHeight(),
				Reason:            "sender is not governance authority",
			})
			return types.ErrUnauthorized.Wrapf(
				"operation %s requires governance authority", operation,
			)
		}
	}

	if requireValidator {
		if !k.IsValidator(ctx, sender) {
			k.recordPrivilegeAudit(ctx, PrivilegeAuditRecord{
				Operation:         operation,
				Account:           sender.String(),
				RequiredPrivilege: "bonded_validator",
				Granted:           false,
				BlockHeight:       ctx.BlockHeight(),
				Reason:            "sender is not a bonded validator",
			})
			return types.ErrUnauthorized.Wrapf(
				"operation %s requires bonded validator status", operation,
			)
		}
	}

	k.recordPrivilegeAudit(ctx, PrivilegeAuditRecord{
		Operation:         operation,
		Account:           sender.String(),
		RequiredPrivilege: "none",
		Granted:           true,
		BlockHeight:       ctx.BlockHeight(),
	})

	return nil
}

func (k Keeper) recordPrivilegeAudit(ctx sdk.Context, record PrivilegeAuditRecord) {
	bz, err := json.Marshal(record)
	if err != nil {
		return
	}

	store := ctx.KVStore(k.skey)
	key := make([]byte, 0, len(types.PrefixPrivilegeAudit)+8+len(record.Operation))
	key = append(key, types.PrefixPrivilegeAudit...)
	heightBz := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBz, safeUint64FromIntBiometricToUint64(ctx.BlockHeight()))
	key = append(key, heightBz...)
	key = append(key, []byte(record.Operation)...)
	store.Set(key, bz)
}
