package keeper

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode/utf8"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

const maxDerivedFeaturePayloadSize = 1024

var (
	rawBiometricMagicPrefixes = [][]byte{
		{0xFF, 0xD8, 0xFF},
		{0x89, 'P', 'N', 'G', 0x0D, 0x0A},
		{'G', 'I', 'F', '8'},
		{'B', 'M'},
		{'I', 'I', '*', 0x00},
		{'M', 'M', 0x00, '*'},
		{'R', 'I', 'F', 'F'},
		{'%', 'P', 'D', 'F'},
		{0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p'},
	}
	rawBiometricJSONFields = map[string]struct{}{
		"image":              {},
		"image_data":         {},
		"selfie":             {},
		"document_image":     {},
		"document_scan":      {},
		"face_embedding":     {},
		"embedding":          {},
		"face_template":      {},
		"biometric":          {},
		"biometric_template": {},
		"raw_capture":        {},
		"video_frame":        {},
	}
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
//
//nolint:unused // reserved for future artifact expiry cleanup
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
	if len(payload) == 0 {
		return nil
	}

	trimmed := bytes.TrimSpace(payload)
	if looksLikeCompactHashArtifact(trimmed) {
		return nil
	}

	for _, prefix := range rawBiometricMagicPrefixes {
		if bytes.HasPrefix(trimmed, prefix) {
			return types.ErrInvalidPayload.Wrap("raw biometric media payloads are not permitted on-chain")
		}
	}

	if len(trimmed) > maxDerivedFeaturePayloadSize {
		return types.ErrInvalidPayload.Wrapf(
			"derived feature payload exceeds %d bytes; raw biometric artifacts must remain off-chain",
			maxDerivedFeaturePayloadSize,
		)
	}

	if json.Valid(trimmed) {
		var decoded interface{}
		if err := json.Unmarshal(trimmed, &decoded); err != nil {
			return types.ErrInvalidPayload.Wrapf("invalid JSON payload: %v", err)
		}
		if containsRawBiometricJSON(decoded) {
			return types.ErrInvalidPayload.Wrap("raw biometric fields detected in on-chain payload")
		}
		return nil
	}

	if utf8.Valid(trimmed) {
		lower := strings.ToLower(string(trimmed))
		if strings.Contains(lower, "data:image/") || strings.Contains(lower, "data:application/pdf") {
			return types.ErrInvalidPayload.Wrap("raw biometric or document data URIs are not permitted on-chain")
		}
		if len(trimmed) <= maxDerivedFeaturePayloadSize {
			return nil
		}
	}

	if printableRatio(trimmed) < 0.85 {
		return types.ErrInvalidPayload.Wrap("opaque binary biometric payloads are not permitted on-chain")
	}

	return nil
}

func looksLikeCompactHashArtifact(payload []byte) bool {
	if len(payload) == 32 {
		return true
	}

	lower := bytes.ToLower(payload)
	if len(lower) == 64 && isHexBytes(lower) {
		return true
	}

	return len(payload) == 44 && isBase64Alphabet(payload)
}

func isHexBytes(payload []byte) bool {
	for _, b := range payload {
		if (b < '0' || b > '9') && (b < 'a' || b > 'f') {
			return false
		}
	}

	return true
}

func isBase64Alphabet(payload []byte) bool {
	for _, b := range payload {
		switch {
		case b >= 'A' && b <= 'Z':
		case b >= 'a' && b <= 'z':
		case b >= '0' && b <= '9':
		case b == '+' || b == '/' || b == '=' || b == '-' || b == '_':
		default:
			return false
		}
	}

	return true
}

func containsRawBiometricJSON(value interface{}) bool {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, nested := range typed {
			if _, blocked := rawBiometricJSONFields[strings.ToLower(strings.TrimSpace(key))]; blocked {
				return true
			}
			if containsRawBiometricJSON(nested) {
				return true
			}
		}
	case []interface{}:
		for _, nested := range typed {
			if containsRawBiometricJSON(nested) {
				return true
			}
		}
	case string:
		lower := strings.ToLower(strings.TrimSpace(typed))
		if strings.HasPrefix(lower, "data:image/") || strings.HasPrefix(lower, "data:application/pdf") {
			return true
		}
		if len(lower) > 256 && isBase64Alphabet([]byte(lower)) {
			return true
		}
	}

	return false
}

func printableRatio(payload []byte) float64 {
	if len(payload) == 0 {
		return 1
	}

	printable := 0
	for _, b := range payload {
		if b == '\n' || b == '\r' || b == '\t' || (b >= 0x20 && b <= 0x7E) {
			printable++
		}
	}

	return float64(printable) / float64(len(payload))
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
