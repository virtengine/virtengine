package keeper

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// RegisterSignerKey stores a governed verification signer key and its lookup indexes.
func (k Keeper) RegisterSignerKey(ctx sdk.Context, authority string, key *types.SignerKeyInfo) error {
	if authority != k.GetAuthority() {
		return types.ErrUnauthorized.Wrapf("invalid authority: expected %s, got %s", k.GetAuthority(), authority)
	}
	return k.registerSignerKey(ctx, key)
}

func (k Keeper) registerSignerKey(ctx sdk.Context, key *types.SignerKeyInfo) error {
	if key == nil {
		return types.ErrInvalidSignerKey.Wrap("signer key cannot be nil")
	}
	if err := key.Validate(); err != nil {
		return err
	}
	if err := validateSignerKeyRegistration(ctx, key); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	if existing, found := k.GetSignerKey(ctx, key.KeyID); found {
		if err := validateSignerKeyUpdate(store, existing, key); err != nil {
			return err
		}
	} else {
		if err := k.validateNewSignerKeySequence(ctx, key); err != nil {
			return err
		}
	}
	if err := validateSignerKeyIndexes(store, key); err != nil {
		return err
	}

	bz, err := json.Marshal(key)
	if err != nil {
		return err
	}
	store.Set(signerKeyStoreKey(key.KeyID), bz)
	store.Set(signerKeyFingerprintStoreKey(key.Fingerprint), []byte(key.KeyID))
	store.Set(signerKeyBySignerStoreKey(key.SignerID, key.KeyID), []byte{1})
	return nil
}

func validateSignerKeyRegistration(ctx sdk.Context, key *types.SignerKeyInfo) error {
	if key.SequenceNumber == 0 {
		return types.ErrInvalidSignerKey.Wrap("signer key sequence number is required")
	}
	if key.State == types.SignerKeyStateActive || key.State == types.SignerKeyStateRotating {
		if key.ActivatedAt == nil {
			return types.ErrInvalidSignerKey.Wrap("active signer key activation timestamp is required")
		}
	}
	if key.State == types.SignerKeyStateRevoked && key.RevokedAt == nil {
		return types.ErrInvalidSignerKey.Wrap("revoked signer key revocation timestamp is required")
	}
	if key.State == types.SignerKeyStateExpired && key.ExpiresAt == nil {
		return types.ErrInvalidSignerKey.Wrap("expired signer key expiry timestamp is required")
	}
	if key.ActivatedAt != nil && key.ExpiresAt != nil && !key.ExpiresAt.After(*key.ActivatedAt) {
		return types.ErrInvalidSignerKey.Wrap("signer key expiry must be after activation")
	}
	if key.RevokedAt != nil {
		if key.ActivatedAt != nil && key.RevokedAt.Before(*key.ActivatedAt) {
			return types.ErrInvalidSignerKey.Wrap("signer key revocation cannot predate activation")
		}
	}
	if err := validateSignerKeyRegistrationLifecycle(ctx, key); err != nil {
		return err
	}
	if err := validateSignerKeyRegistrationEvidencePolicy(key); err != nil {
		return err
	}
	if err := validateSignerKeyRegistrationHeightPolicy(key); err != nil {
		return err
	}
	return nil
}

func validateSignerKeyRegistrationLifecycle(ctx sdk.Context, key *types.SignerKeyInfo) error {
	blockTime := ctx.BlockTime().UTC()
	switch key.State {
	case types.SignerKeyStateActive:
		if key.ActivatedAt != nil && key.ActivatedAt.UTC().After(blockTime) {
			return types.ErrInvalidSignerKey.Wrap("active signer key activation timestamp cannot be in the future")
		}
	case types.SignerKeyStateRotating:
		if key.ActivatedAt != nil && key.ActivatedAt.UTC().After(blockTime) {
			return types.ErrInvalidSignerKey.Wrap("rotating signer key activation timestamp cannot be in the future")
		}
		if strings.TrimSpace(key.SuccessorKeyID) == "" {
			return types.ErrInvalidSignerKey.Wrap("rotating signer key successor key id is required")
		}
	case types.SignerKeyStateRevoked:
		if key.RevokedAt != nil && key.RevokedAt.UTC().After(blockTime) {
			return types.ErrInvalidSignerKey.Wrap("revoked signer key revocation timestamp cannot be in the future")
		}
		if key.RevocationReason == "" {
			return types.ErrInvalidSignerKey.Wrap("revoked signer key revocation reason is required")
		}
	case types.SignerKeyStateExpired:
		if key.ExpiresAt != nil && key.ExpiresAt.UTC().After(blockTime) {
			return types.ErrInvalidSignerKey.Wrap("expired signer key expiry timestamp cannot be in the future")
		}
	}
	if key.RevocationReason != "" {
		if key.State != types.SignerKeyStateRevoked {
			return types.ErrInvalidSignerKey.Wrap("signer key revocation reason is only valid for revoked keys")
		}
		if !types.IsValidRevocationReason(key.RevocationReason) {
			return types.ErrInvalidSignerKey.Wrapf("invalid signer key revocation reason: %s", key.RevocationReason)
		}
	}
	return nil
}

func validateSignerKeyUpdate(store storetypes.KVStore, existing *types.SignerKeyInfo, next *types.SignerKeyInfo) error {
	if err := validateSignerKeyIndexConsistency(store, existing); err != nil {
		return err
	}
	if existing.Fingerprint != next.Fingerprint ||
		!bytes.Equal(existing.PublicKey, next.PublicKey) ||
		existing.Algorithm != next.Algorithm ||
		existing.SignerID != next.SignerID ||
		existing.SequenceNumber != next.SequenceNumber ||
		!existing.CreatedAt.Equal(next.CreatedAt) {
		return types.ErrInvalidSignerKey.Wrap("signer key immutable identity fields cannot change")
	}
	if !isAllowedSignerKeyTransition(existing.State, next.State) {
		return types.ErrInvalidSignerKey.Wrapf("invalid signer key state transition: %s -> %s", existing.State, next.State)
	}
	if err := validateSignerKeyMutableUpdate(existing, next); err != nil {
		return err
	}
	if err := validateSignerKeyTimePolicyUpdate(existing, next); err != nil {
		return err
	}
	return validateSignerKeyHeightPolicyUpdate(existing, next)
}

func validateSignerKeyIndexConsistency(store storetypes.KVStore, existing *types.SignerKeyInfo) error {
	if existing == nil {
		return types.ErrInvalidSignerKey.Wrap("existing signer key is required")
	}
	if indexedKeyID := string(store.Get(signerKeyFingerprintStoreKey(existing.Fingerprint))); indexedKeyID != existing.KeyID {
		return types.ErrInvalidSignerKey.Wrap("signer key fingerprint index is inconsistent")
	}
	if !store.Has(signerKeyBySignerStoreKey(existing.SignerID, existing.KeyID)) {
		return types.ErrInvalidSignerKey.Wrap("signer key signer index is inconsistent")
	}
	return nil
}

func validateSignerKeyIndexes(store storetypes.KVStore, key *types.SignerKeyInfo) error {
	if indexedKeyID := string(store.Get(signerKeyFingerprintStoreKey(key.Fingerprint))); indexedKeyID != "" {
		if indexedKeyID != key.KeyID {
			return types.ErrInvalidSignerKey.Wrap("signer key fingerprint already registered to another key")
		}
		if store.Get(signerKeyStoreKey(indexedKeyID)) == nil {
			return types.ErrInvalidSignerKey.Wrap("signer key fingerprint index points to missing key")
		}
	}
	if store.Has(signerKeyBySignerStoreKey(key.SignerID, key.KeyID)) && store.Get(signerKeyStoreKey(key.KeyID)) == nil {
		return types.ErrInvalidSignerKey.Wrap("signer key signer index points to missing key")
	}
	return nil
}

func isAllowedSignerKeyTransition(from types.SignerKeyState, to types.SignerKeyState) bool {
	switch from {
	case types.SignerKeyStatePending:
		return to == types.SignerKeyStatePending ||
			to == types.SignerKeyStateActive
	case types.SignerKeyStateActive:
		return to == types.SignerKeyStateActive ||
			to == types.SignerKeyStateRotating ||
			to == types.SignerKeyStateRevoked ||
			to == types.SignerKeyStateExpired
	case types.SignerKeyStateRotating:
		return to == types.SignerKeyStateRotating ||
			to == types.SignerKeyStateRevoked ||
			to == types.SignerKeyStateExpired
	case types.SignerKeyStateRevoked:
		return to == types.SignerKeyStateRevoked
	case types.SignerKeyStateExpired:
		return to == types.SignerKeyStateExpired
	default:
		return false
	}
}

func validateSignerKeyMutableUpdate(existing *types.SignerKeyInfo, next *types.SignerKeyInfo) error {
	if existing.SuccessorKeyID != "" && next.SuccessorKeyID != existing.SuccessorKeyID {
		return types.ErrInvalidSignerKey.Wrap("signer key successor key id cannot be retargeted")
	}
	if existing.SuccessorKeyID == "" && next.SuccessorKeyID != "" && next.State != types.SignerKeyStateRotating {
		return types.ErrInvalidSignerKey.Wrap("signer key successor key id can only be introduced when rotating")
	}
	if existing.PredecessorKeyID != "" && next.PredecessorKeyID != existing.PredecessorKeyID {
		return types.ErrInvalidSignerKey.Wrap("signer key predecessor key id cannot be retargeted")
	}
	if existing.RevocationReason != "" && next.RevocationReason != existing.RevocationReason {
		return types.ErrInvalidSignerKey.Wrap("signer key revocation reason cannot change")
	}
	if err := validateSignerKeyEvidencePolicyUpdate(existing, next); err != nil {
		return err
	}
	return validateSignerKeyServiceMetadataPolicyUpdate(existing, next)
}

func validateSignerKeyEvidencePolicyUpdate(existing *types.SignerKeyInfo, next *types.SignerKeyInfo) error {
	existingTypes, err := signerKeyEvidencePolicySet(existing.Metadata[types.SignerKeyMetadataEvidenceTypes])
	if err != nil {
		return err
	}
	nextTypes, err := signerKeyEvidencePolicySet(next.Metadata[types.SignerKeyMetadataEvidenceTypes])
	if err != nil {
		return err
	}
	for evidenceType := range nextTypes {
		if _, ok := existingTypes[evidenceType]; !ok {
			return types.ErrInvalidSignerKey.Wrap("signer key evidence type policy cannot be broadened")
		}
	}
	return nil
}

func signerKeyEvidencePolicySet(raw string) (map[string]struct{}, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, types.ErrInvalidSignerKey.Wrap("signer key evidence type policy is required")
	}
	out := make(map[string]struct{})
	for _, item := range strings.Split(raw, ",") {
		evidenceType := strings.TrimSpace(item)
		if evidenceType == "" {
			return nil, types.ErrInvalidSignerKey.Wrap("signer key evidence type policy contains an empty type")
		}
		if !types.IsValidAttestationType(types.AttestationType(evidenceType)) {
			return nil, types.ErrInvalidSignerKey.Wrapf("signer key evidence type policy contains invalid type: %s", evidenceType)
		}
		out[evidenceType] = struct{}{}
	}
	return out, nil
}

func validateSignerKeyServiceMetadataPolicyUpdate(existing *types.SignerKeyInfo, next *types.SignerKeyInfo) error {
	existingHash := existing.Metadata[types.SignerKeyMetadataServiceMetadataHash]
	nextHash := next.Metadata[types.SignerKeyMetadataServiceMetadataHash]
	if existingHash != "" && nextHash != existingHash {
		return types.ErrInvalidSignerKey.Wrap("signer key service metadata policy cannot change")
	}
	return nil
}

func validateSignerKeyTimePolicyUpdate(existing *types.SignerKeyInfo, next *types.SignerKeyInfo) error {
	if err := validateSignerKeyActivationTimeUpdate(existing.ActivatedAt, next.ActivatedAt); err != nil {
		return err
	}
	if err := validateSignerKeyTerminalTimeUpdate(existing.ExpiresAt, next.ExpiresAt, "expiry"); err != nil {
		return err
	}
	return validateSignerKeyTerminalTimeUpdate(existing.RevokedAt, next.RevokedAt, "revocation")
}

func validateSignerKeyActivationTimeUpdate(existing *time.Time, next *time.Time) error {
	if existing == nil {
		return nil
	}
	if next == nil {
		return types.ErrInvalidSignerKey.Wrap("signer key activation timestamp cannot be cleared")
	}
	if !next.Equal(*existing) {
		return types.ErrInvalidSignerKey.Wrap("signer key activation timestamp cannot change")
	}
	return nil
}

func validateSignerKeyTerminalTimeUpdate(existing *time.Time, next *time.Time, name string) error {
	if existing == nil {
		return nil
	}
	if next == nil {
		return types.ErrInvalidSignerKey.Wrapf("signer key %s timestamp cannot be cleared", name)
	}
	if next.After(*existing) {
		return types.ErrInvalidSignerKey.Wrapf("signer key %s timestamp cannot move later", name)
	}
	return nil
}

func validateSignerKeyHeightPolicyUpdate(existing *types.SignerKeyInfo, next *types.SignerKeyInfo) error {
	if err := validateSignerKeyHeightUpdate(existing, next, types.SignerKeyMetadataActivationHeight, false); err != nil {
		return err
	}
	if err := validateSignerKeyHeightUpdate(existing, next, types.SignerKeyMetadataExpiryHeight, true); err != nil {
		return err
	}
	return validateSignerKeyHeightUpdate(existing, next, types.SignerKeyMetadataRevokedHeight, true)
}

func validateSignerKeyHeightUpdate(existing *types.SignerKeyInfo, next *types.SignerKeyInfo, field string, terminal bool) error {
	oldHeight, oldOK, err := parseSignerKeyRegistrationHeight(existing.Metadata[field], field)
	if err != nil {
		return err
	}
	newHeight, newOK, err := parseSignerKeyRegistrationHeight(next.Metadata[field], field)
	if err != nil {
		return err
	}
	if oldOK && !newOK {
		return types.ErrInvalidSignerKey.Wrapf("signer key %s height cannot be cleared", field)
	}
	if !oldOK {
		return nil
	}
	if terminal {
		if newHeight > oldHeight {
			return types.ErrInvalidSignerKey.Wrapf("signer key %s height cannot move later", field)
		}
		return nil
	}
	if newHeight != oldHeight {
		return types.ErrInvalidSignerKey.Wrapf("signer key %s height cannot change", field)
	}
	return nil
}

func (k Keeper) validateNewSignerKeySequence(ctx sdk.Context, key *types.SignerKeyInfo) error {
	store := ctx.KVStore(k.skey)
	prefix := signerKeyBySignerPrefixKey(key.SignerID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var maxSequence uint64
	for ; iterator.Valid(); iterator.Next() {
		indexedKeyID := string(iterator.Key()[len(prefix):])
		existing, found := k.GetSignerKey(ctx, indexedKeyID)
		if !found {
			return types.ErrInvalidSignerKey.Wrap("signer key signer index points to missing key")
		}
		if existing.SignerID != key.SignerID {
			return types.ErrInvalidSignerKey.Wrap("signer key signer index collision")
		}
		if existing.SequenceNumber == key.SequenceNumber {
			return types.ErrInvalidSignerKey.Wrap("signer key sequence already registered")
		}
		if existing.SequenceNumber > maxSequence {
			maxSequence = existing.SequenceNumber
		}
	}
	if maxSequence > 0 && key.SequenceNumber <= maxSequence {
		return types.ErrInvalidSignerKey.Wrap("signer key sequence must increase for signer")
	}
	return nil
}

func validateSignerKeyRegistrationEvidencePolicy(key *types.SignerKeyInfo) error {
	if key.Metadata == nil {
		return types.ErrInvalidSignerKey.Wrap("signer key evidence type policy is required")
	}
	raw := key.Metadata[types.SignerKeyMetadataEvidenceTypes]
	if strings.TrimSpace(raw) == "" {
		return types.ErrInvalidSignerKey.Wrap("signer key evidence type policy is required")
	}
	for _, item := range strings.Split(raw, ",") {
		evidenceType := strings.TrimSpace(item)
		if evidenceType == "" {
			return types.ErrInvalidSignerKey.Wrap("signer key evidence type policy contains an empty type")
		}
		if !types.IsValidAttestationType(types.AttestationType(evidenceType)) {
			return types.ErrInvalidSignerKey.Wrapf("signer key evidence type policy contains invalid type: %s", evidenceType)
		}
	}
	return nil
}

func validateSignerKeyRegistrationHeightPolicy(key *types.SignerKeyInfo) error {
	if key.Metadata == nil {
		return nil
	}
	activationHeight, hasActivationHeight, err := parseSignerKeyRegistrationHeight(key.Metadata[types.SignerKeyMetadataActivationHeight], "activation")
	if err != nil {
		return err
	}
	expiryHeight, hasExpiryHeight, err := parseSignerKeyRegistrationHeight(key.Metadata[types.SignerKeyMetadataExpiryHeight], "expiry")
	if err != nil {
		return err
	}
	revokedHeight, hasRevokedHeight, err := parseSignerKeyRegistrationHeight(key.Metadata[types.SignerKeyMetadataRevokedHeight], "revoked")
	if err != nil {
		return err
	}
	if hasActivationHeight && hasExpiryHeight && expiryHeight <= activationHeight {
		return types.ErrInvalidSignerKey.Wrap("signer key expiry height must be after activation height")
	}
	if hasActivationHeight && hasRevokedHeight && revokedHeight < activationHeight {
		return types.ErrInvalidSignerKey.Wrap("signer key revoked height cannot predate activation height")
	}
	return nil
}

func parseSignerKeyRegistrationHeight(raw string, name string) (int64, bool, error) {
	if raw == "" {
		return 0, false, nil
	}
	height, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || height < 0 {
		return 0, false, types.ErrInvalidSignerKey.Wrapf("invalid signer %s height", name)
	}
	return height, true, nil
}

// GetSignerKey returns a signer key by governed key ID.
func (k Keeper) GetSignerKey(ctx sdk.Context, keyID string) (*types.SignerKeyInfo, bool) {
	if keyID == "" {
		return nil, false
	}
	bz := ctx.KVStore(k.skey).Get(signerKeyStoreKey(keyID))
	if bz == nil {
		return nil, false
	}
	var key types.SignerKeyInfo
	if err := json.Unmarshal(bz, &key); err != nil {
		return nil, false
	}
	return &key, true
}

// GetSignerKeyByFingerprint returns a signer key through the fingerprint index.
func (k Keeper) GetSignerKeyByFingerprint(ctx sdk.Context, fingerprint string) (*types.SignerKeyInfo, bool) {
	if fingerprint == "" {
		return nil, false
	}
	keyID := string(ctx.KVStore(k.skey).Get(signerKeyFingerprintStoreKey(fingerprint)))
	if keyID == "" {
		return nil, false
	}
	return k.GetSignerKey(ctx, keyID)
}

func (k Keeper) resolveSignerKey(ctx sdk.Context, keyID string, fingerprint string) (*types.SignerKeyInfo, error) {
	if keyID == "" || fingerprint == "" {
		return nil, types.ErrSignerKeyNotFound.Wrap("signer key id and fingerprint are required")
	}
	key, found := k.GetSignerKey(ctx, keyID)
	if !found {
		return nil, types.ErrSignerKeyNotFound.Wrapf("signer key not found: %s", keyID)
	}
	indexed, found := k.GetSignerKeyByFingerprint(ctx, fingerprint)
	if !found {
		return nil, types.ErrSignerKeyNotFound.Wrapf("signer key fingerprint not found: %s", fingerprint)
	}
	if indexed.KeyID != key.KeyID || key.Fingerprint != fingerprint {
		return nil, types.ErrInvalidSignerKey.Wrap("signer key id/fingerprint mismatch")
	}
	return key, nil
}

func signerKeyStoreKey(keyID string) []byte {
	key := make([]byte, 0, len(types.PrefixSignerKey)+len(keyID))
	key = append(key, types.PrefixSignerKey...)
	key = append(key, []byte(keyID)...)
	return key
}

func signerKeyFingerprintStoreKey(fingerprint string) []byte {
	key := make([]byte, 0, len(types.PrefixSignerKeyByFingerprint)+len(fingerprint))
	key = append(key, types.PrefixSignerKeyByFingerprint...)
	key = append(key, []byte(fingerprint)...)
	return key
}

func signerKeyBySignerStoreKey(signerID string, keyID string) []byte {
	key := make([]byte, 0, len(types.PrefixSignerKeyBySigner)+len(signerID)+1+len(keyID))
	key = append(key, types.PrefixSignerKeyBySigner...)
	key = append(key, []byte(signerID)...)
	key = append(key, byte('/'))
	key = append(key, []byte(keyID)...)
	return key
}

func signerKeyBySignerPrefixKey(signerID string) []byte {
	key := make([]byte, 0, len(types.PrefixSignerKeyBySigner)+len(signerID)+1)
	key = append(key, types.PrefixSignerKeyBySigner...)
	key = append(key, []byte(signerID)...)
	key = append(key, byte('/'))
	return key
}
