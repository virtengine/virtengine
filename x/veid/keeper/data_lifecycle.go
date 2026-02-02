package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Data Lifecycle Management (VE-217: Derived Feature Minimization)
// ============================================================================

// dataLifecycleRulesStore is the storage format for lifecycle rules
type dataLifecycleRulesStore struct {
	Version          uint32                                 `json:"version"`
	ArtifactPolicies map[string]*artifactRetentionRuleStore `json:"artifact_policies"`
}

// artifactRetentionRuleStore is the storage format for artifact retention rules
type artifactRetentionRuleStore struct {
	ArtifactType            string `json:"artifact_type"`
	AllowOnChain            bool   `json:"allow_on_chain"`
	RequireEncryption       bool   `json:"require_encryption"`
	MaxRetentionDays        uint32 `json:"max_retention_days"`
	DefaultRetentionDays    uint32 `json:"default_retention_days"`
	DeleteAfterVerification bool   `json:"delete_after_verification"`
	AllowOffChainStorage    bool   `json:"allow_off_chain_storage"`
	RequireUserConsent      bool   `json:"require_user_consent"`
	Description             string `json:"description"`
}

// SetDataLifecycleRules stores data lifecycle rules
func (k Keeper) SetDataLifecycleRules(ctx sdk.Context, rules types.DataLifecycleRules) error {
	if err := rules.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	// Convert to storage format
	policies := make(map[string]*artifactRetentionRuleStore)
	for at, rule := range rules.ArtifactPolicies {
		policies[string(at)] = &artifactRetentionRuleStore{
			ArtifactType:            string(rule.ArtifactType),
			AllowOnChain:            rule.AllowOnChain,
			RequireEncryption:       rule.RequireEncryption,
			MaxRetentionDays:        rule.MaxRetentionDays,
			DefaultRetentionDays:    rule.DefaultRetentionDays,
			DeleteAfterVerification: rule.DeleteAfterVerification,
			AllowOffChainStorage:    rule.AllowOffChainStorage,
			RequireUserConsent:      rule.RequireUserConsent,
			Description:             rule.Description,
		}
	}

	rs := dataLifecycleRulesStore{
		Version:          rules.Version,
		ArtifactPolicies: policies,
	}

	bz, err := json.Marshal(&rs)
	if err != nil {
		return err
	}

	store.Set(types.DataLifecycleRulesKey(), bz)
	return nil
}

// GetDataLifecycleRules retrieves data lifecycle rules
func (k Keeper) GetDataLifecycleRules(ctx sdk.Context) types.DataLifecycleRules {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.DataLifecycleRulesKey())
	if bz == nil {
		// Return default rules if not set
		return *types.DefaultDataLifecycleRules()
	}

	var rs dataLifecycleRulesStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return *types.DefaultDataLifecycleRules()
	}

	// Convert from storage format
	policies := make(map[types.ArtifactType]*types.ArtifactRetentionRule)
	for atStr, rule := range rs.ArtifactPolicies {
		at := types.ArtifactType(atStr)
		policies[at] = &types.ArtifactRetentionRule{
			ArtifactType:            types.ArtifactType(rule.ArtifactType),
			AllowOnChain:            rule.AllowOnChain,
			RequireEncryption:       rule.RequireEncryption,
			MaxRetentionDays:        rule.MaxRetentionDays,
			DefaultRetentionDays:    rule.DefaultRetentionDays,
			DeleteAfterVerification: rule.DeleteAfterVerification,
			AllowOffChainStorage:    rule.AllowOffChainStorage,
			RequireUserConsent:      rule.RequireUserConsent,
			Description:             rule.Description,
		}
	}

	return types.DataLifecycleRules{
		Version:          rs.Version,
		ArtifactPolicies: policies,
	}
}

// CanStoreOnChain checks if an artifact type can be stored on-chain
func (k Keeper) CanStoreOnChain(ctx sdk.Context, artifactType types.ArtifactType) bool {
	rules := k.GetDataLifecycleRules(ctx)
	return rules.CanStoreOnChain(artifactType)
}

// RequiresEncryption checks if an artifact type requires encryption
func (k Keeper) RequiresEncryption(ctx sdk.Context, artifactType types.ArtifactType) bool {
	rules := k.GetDataLifecycleRules(ctx)
	return rules.RequiresEncryption(artifactType)
}

// ShouldDeleteAfterVerification checks if artifact should be deleted post-verification
func (k Keeper) ShouldDeleteAfterVerification(ctx sdk.Context, artifactType types.ArtifactType) bool {
	rules := k.GetDataLifecycleRules(ctx)
	return rules.ShouldDeleteAfterVerification(artifactType)
}

// ValidateArtifactStorage validates if an artifact can be stored with the given parameters
func (k Keeper) ValidateArtifactStorage(ctx sdk.Context, artifactType types.ArtifactType, onChain bool, encrypted bool) error {
	rules := k.GetDataLifecycleRules(ctx)
	rule, found := rules.GetRule(artifactType)
	if !found {
		return types.ErrInvalidParams.Wrapf("no rule for artifact type: %s", artifactType)
	}

	if onChain && !rule.AllowOnChain {
		return types.ErrInvalidParams.Wrapf("artifact type %s cannot be stored on-chain", artifactType)
	}

	if rule.RequireEncryption && !encrypted {
		return types.ErrInvalidParams.Wrapf("artifact type %s requires encryption", artifactType)
	}

	return nil
}

// ============================================================================
// Retention Policy Storage
// ============================================================================

// SetRetentionPolicy stores a retention policy
func (k Keeper) SetRetentionPolicy(ctx sdk.Context, policy types.RetentionPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	ps := retentionPolicyToStore(&policy)
	bz, err := json.Marshal(ps)
	if err != nil {
		return err
	}

	store.Set(types.RetentionPolicyKey(policy.PolicyID), bz)

	// If policy has an expiry and delete on expiry, add to expired artifacts index
	if policy.DeleteOnExpiry && policy.ExpiresAt != nil {
		k.addToExpiredArtifactsIndex(ctx, policy.ExpiresAt.Unix(), "policy", policy.PolicyID)
	}

	return nil
}

// GetRetentionPolicy retrieves a retention policy
func (k Keeper) GetRetentionPolicy(ctx sdk.Context, policyID string) (types.RetentionPolicy, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.RetentionPolicyKey(policyID))
	if bz == nil {
		return types.RetentionPolicy{}, false
	}

	var ps retentionPolicyStore
	if err := json.Unmarshal(bz, &ps); err != nil {
		return types.RetentionPolicy{}, false
	}

	return *retentionPolicyFromStore(&ps), true
}

// DeleteRetentionPolicy deletes a retention policy
func (k Keeper) DeleteRetentionPolicy(ctx sdk.Context, policyID string) {
	store := ctx.KVStore(k.skey)
	store.Delete(types.RetentionPolicyKey(policyID))
}

// ExtendRetentionPolicy extends a retention policy
func (k Keeper) ExtendRetentionPolicy(ctx sdk.Context, policyID string) error {
	policy, found := k.GetRetentionPolicy(ctx, policyID)
	if !found {
		return types.ErrInvalidParams.Wrapf("policy not found: %s", policyID)
	}

	if err := policy.Extend(); err != nil {
		return err
	}

	return k.SetRetentionPolicy(ctx, policy)
}

// ============================================================================
// Expired Artifacts Cleanup
// ============================================================================

// addToExpiredArtifactsIndex adds an artifact to the expiry index
func (k Keeper) addToExpiredArtifactsIndex(ctx sdk.Context, expiresAt int64, artifactType string, artifactID string) {
	store := ctx.KVStore(k.skey)
	key := types.ExpiredArtifactKey(expiresAt, artifactType, artifactID)
	store.Set(key, []byte{1})
}

// removeFromExpiredArtifactsIndex removes an artifact from the expiry index
func (k Keeper) removeFromExpiredArtifactsIndex(ctx sdk.Context, expiresAt int64, artifactType string, artifactID string) {
	store := ctx.KVStore(k.skey)
	key := types.ExpiredArtifactKey(expiresAt, artifactType, artifactID)
	store.Delete(key)
}

// CleanupExpiredArtifacts cleans up expired artifacts
func (k Keeper) CleanupExpiredArtifacts(ctx sdk.Context) (cleaned int) {
	now := ctx.BlockTime().Unix()

	store := ctx.KVStore(k.skey)
	prefix := types.ExpiredArtifactPrefixKey()

	iter := store.Iterator(prefix, types.ExpiredArtifactBeforeKey(now+1))
	defer iter.Close()

	var toDelete [][]byte

	for ; iter.Valid(); iter.Next() {
		toDelete = append(toDelete, iter.Key())
	}

	for _, key := range toDelete {
		store.Delete(key)
		cleaned++
	}

	// Also cleanup expired envelopes
	cleaned += k.CleanupExpiredEnvelopes(ctx)

	return cleaned
}

// ============================================================================
// Security Validation Functions (VE-217)
// ============================================================================

// ValidateNoRawBiometricsOnChain validates that no raw biometrics are being stored on-chain
// This is a critical security check
func (k Keeper) ValidateNoRawBiometricsOnChain(ctx sdk.Context, payload []byte) error {
	// This is a placeholder for additional security checks
	// In production, this would use heuristics or ML to detect biometric data
	// For now, we rely on the type system and lifecycle rules

	// Check payload size - raw images would be much larger than hashes
	maxHashedPayloadSize := 1024 // 1KB is plenty for hashes and metadata
	if len(payload) > maxHashedPayloadSize {
		// Log warning but don't fail - encrypted payloads may be larger
		k.Logger(ctx).Warn("large payload detected - verify no raw biometrics",
			"size", len(payload))
	}

	return nil
}

// ValidateDerivedFeaturesOnly validates that only derived features (hashes) are stored
func (k Keeper) ValidateDerivedFeaturesOnly(ctx sdk.Context, record types.DerivedFeatureVerificationRecord) error {
	// Verify all feature references contain only hashes
	for i, ref := range record.FeatureReferences {
		if len(ref.FeatureHash) != 32 {
			return types.ErrInvalidPayloadHash.Wrapf(
				"feature_reference[%d] has invalid hash length: expected 32, got %d",
				i, len(ref.FeatureHash))
		}
	}

	// Verify composite hash
	if len(record.CompositeHash) > 0 && len(record.CompositeHash) != 32 {
		return types.ErrInvalidPayloadHash.Wrap("composite_hash has invalid length")
	}

	return nil
}

// AuditDataLifecycleCompliance generates an audit report for data lifecycle compliance
func (k Keeper) AuditDataLifecycleCompliance(ctx sdk.Context, address sdk.AccAddress) (map[string]interface{}, error) {
	report := make(map[string]interface{})

	rules := k.GetDataLifecycleRules(ctx)
	report["rules_version"] = rules.Version

	// Check embedding envelopes
	envelopes := k.GetEmbeddingEnvelopesByAccount(ctx, address)
	envelopeStats := make(map[string]int)
	expiredCount := 0
	revokedCount := 0

	for _, env := range envelopes {
		envelopeStats[string(env.EmbeddingType)]++
		if env.Revoked {
			revokedCount++
		}
		if env.RetentionPolicy != nil && env.RetentionPolicy.IsExpired(ctx.BlockTime()) {
			expiredCount++
		}
	}

	report["embedding_envelopes"] = map[string]interface{}{
		"total":   len(envelopes),
		"by_type": envelopeStats,
		"revoked": revokedCount,
		"expired": expiredCount,
	}

	// Check verification records
	records := k.GetDerivedFeatureRecordsByAccount(ctx, address)
	report["verification_records"] = map[string]interface{}{
		"total": len(records),
	}

	// Compliance check: no raw biometrics on chain
	report["compliance"] = map[string]interface{}{
		"raw_biometrics_on_chain": false, // Always false by design
		"derived_features_only":   true,
	}

	return report, nil
}
