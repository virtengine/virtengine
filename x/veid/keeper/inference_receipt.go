package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	inferenceReceiptSchemaDomain  = "VEID_INFERENCE_FEATURE_SCHEMA_V1"
	inferenceReceiptFeatureDomain = "VEID_INFERENCE_FEATURE_DIGEST_V1"
	inferenceReceiptInputDomain   = "VEID_INFERENCE_INPUT_DIGEST_V1"
	inferenceReceiptLineageDomain = "VEID_INFERENCE_EVIDENCE_LINEAGE_V1"
	inferenceReceiptModelDomain   = "VEID_INFERENCE_MODEL_ENTRIES_V1"
	inferenceReceiptReplayDomain  = "VEID_INFERENCE_RECEIPT_REPLAY_NONCE_V1"

	inferenceReceiptMaxAge = 5 * time.Minute
)

type inferenceReceiptExpectations struct {
	InputDigest           []byte
	FeatureDigest         []byte
	SchemaDigest          []byte
	EvidenceLineageDigest []byte
	ModelManifestDigest   []byte
	ModelDigest           []byte
	RuntimeImageDigest    []byte
	RuntimeDigest         []byte
	ConfigDigest          []byte
	DeterminismProfile    types.InferenceDeterminismProfile
	PipelineVersion       string
	ScopeIDs              []string
}

type inferenceReceiptReplayCheck struct {
	ContextDigest string
	ReceiptDigest string
	NonceDigest   string
	ExactReplay   bool
}

// ProcessVerificationRequestWithReceipt stages a signed production inference
// result for vote-extension consensus. It never finalizes request state.
func (k Keeper) ProcessVerificationRequestWithReceipt(
	ctx sdk.Context,
	request *types.VerificationRequest,
	keyProvider ValidatorKeyProvider,
	receipt types.InferenceReceipt,
) (*types.VerificationResult, error) {
	if ctx.ExecMode() != sdk.ExecModeVoteExtension {
		return nil, types.ErrUnauthorized.Wrap("inference receipts may only be staged during vote extension execution")
	}
	if request == nil {
		return nil, types.ErrInvalidVerificationRequest.Wrap("verification request is required")
	}
	stored, found := k.GetVerificationRequest(ctx, request.RequestID)
	if !found {
		return nil, types.ErrVerificationRequestNotFound.Wrapf("request %s not found", request.RequestID)
	}
	if stored.AccountAddress != request.AccountAddress {
		return nil, types.ErrInvalidVerificationRequest.Wrap("request account mismatch")
	}
	if types.IsFinalRequestStatus(stored.Status) {
		return nil, types.ErrInvalidVerificationRequest.Wrap("verification request is already final")
	}
	if len(stored.ScopeIDs) == 0 {
		return nil, types.ErrInvalidVerificationRequest.Wrap("verification request has no scopes")
	}
	// The vote-extension bundle commits to one current active profile. An old
	// request profile must be migrated/reissued instead of staged under a stale
	// bundle, so fail before decryption or local receipt-buffer mutation.
	if err := k.requireRequestSnapshotMatchesCurrentBundle(ctx, stored); err != nil {
		return nil, err
	}
	addr, err := sdk.AccAddressFromBech32(stored.AccountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	decryptedScopes, scopeResults, err := k.DecryptScopesForVerification(ctx, addr, stored.ScopeIDs, keyProvider)
	if err != nil {
		return nil, err
	}
	validDecrypted := make([]DecryptedScope, 0, len(decryptedScopes))
	for i, ds := range decryptedScopes {
		valid, reason := k.ValidateDecryptedPayload(ctx, ds)
		if !valid {
			if i < len(scopeResults) {
				scopeResults[i].SetFailure(types.ReasonCodeInvalidPayload)
				scopeResults[i].Details = reason
			}
			continue
		}
		if i < len(scopeResults) {
			scopeResults[i].SetSuccess(0)
		}
		validDecrypted = append(validDecrypted, ds)
	}
	if len(validDecrypted) != len(stored.ScopeIDs) {
		return nil, types.ErrInvalidVerificationResult.Wrap("receipt requires every requested scope to decrypt and validate")
	}
	canonicalScopeIDs := types.CanonicalInferenceReceiptScopeIDs(stored.ScopeIDs)
	if len(canonicalScopeIDs) != len(stored.ScopeIDs) {
		return nil, types.ErrInvalidVerificationRequest.Wrap("verification request scope ids must be non-empty and unique")
	}
	sort.Slice(scopeResults, func(i, j int) bool {
		if scopeResults[i].ScopeID == scopeResults[j].ScopeID {
			return scopeResults[i].ScopeType < scopeResults[j].ScopeType
		}
		return scopeResults[i].ScopeID < scopeResults[j].ScopeID
	})

	expectations, err := k.buildInferenceReceiptExpectations(ctx, stored, validDecrypted, scopeResults)
	if err != nil {
		return nil, err
	}
	replay, err := k.verifyInferenceReceipt(ctx, stored, receipt, expectations)
	if err != nil {
		return nil, err
	}
	if err := validateInferenceReceiptScopeResults(receipt, expectations.ScopeIDs, scopeResults); err != nil {
		return nil, err
	}

	result := k.verificationResultFromReceipt(ctx, stored, receipt, scopeResults, replay)
	inserted, err := k.ensureReceiptBuffer().insert(ctx.BlockHeight(), *result, replay)
	if err != nil {
		return nil, err
	}
	replay.ExactReplay = inserted.ExactReplay
	return result, nil
}

func (k Keeper) buildInferenceReceiptExpectations(
	ctx sdk.Context,
	request *types.VerificationRequest,
	scopes []DecryptedScope,
	scopeResults []types.ScopeVerificationResult,
) (inferenceReceiptExpectations, error) {
	snapshot, err := k.validateInferenceProfileSnapshotReference(ctx, request)
	if err != nil {
		return inferenceReceiptExpectations{}, err
	}
	inputDigest := computeInferenceInputDigest(request.AccountAddress, ctx.BlockHeight(), scopes)
	featureDigest := computeInferenceFeatureDigest(scopes)
	return inferenceReceiptExpectations{
		InputDigest:           inputDigest,
		FeatureDigest:         featureDigest,
		SchemaDigest:          append([]byte(nil), snapshot.FeatureSchemaDigest...),
		EvidenceLineageDigest: computeInferenceEvidenceLineageDigest(request, scopeResults, scopes),
		ModelManifestDigest:   append([]byte(nil), snapshot.ModelManifestDigest...),
		ModelDigest:           append([]byte(nil), snapshot.ModelDigest...),
		RuntimeImageDigest:    append([]byte(nil), snapshot.RuntimeImageDigest...),
		RuntimeDigest:         append([]byte(nil), snapshot.RuntimeDigest...),
		ConfigDigest:          append([]byte(nil), snapshot.DeterminismConfigDigest...),
		DeterminismProfile:    types.CanonicalInferenceDeterminismProfile(),
		PipelineVersion:       snapshot.PipelineVersion,
		ScopeIDs:              types.CanonicalInferenceReceiptScopeIDs(request.ScopeIDs),
	}, nil
}

func (k Keeper) activeInferenceProfileSnapshot(ctx sdk.Context) (*types.InferenceProfileSnapshot, error) {
	active, err := k.GetActivePipelineVersion(ctx)
	if err != nil {
		return nil, err
	}
	manifest, err := k.ensurePipelineVersionUsable(ctx, active)
	if err != nil {
		return nil, err
	}
	return k.inferenceProfileSnapshotFromPipeline(active, manifest)
}

func (k Keeper) inferenceProfileSnapshotFromPipeline(
	pipeline *types.PipelineVersion,
	manifest *types.ModelManifest,
) (*types.InferenceProfileSnapshot, error) {
	if pipeline == nil {
		return nil, types.ErrPipelineVersionNotFound.Wrap("pipeline version is nil")
	}
	if manifest == nil {
		return nil, types.ErrModelManifestMismatch.Wrap("model manifest is nil")
	}
	if !types.IsStrictInferencePipelineDeterminismConfig(pipeline.DeterminismConfig) {
		return nil, types.ErrDeterminismViolation.Wrap("pipeline determinism config is not strict production profile")
	}
	runtimeDigest, err := decodeSHA256Commitment(pipeline.ImageHash)
	if err != nil {
		return nil, types.ErrInvalidPipelineVersion.Wrap(err.Error())
	}
	manifestDigest, err := decodeSHA256Commitment(manifest.ManifestHash)
	if err != nil {
		return nil, types.ErrInvalidModelManifest.Wrap(err.Error())
	}
	snapshot := &types.InferenceProfileSnapshot{
		PipelineVersion:         pipeline.Version,
		RuntimeImageDigest:      append([]byte(nil), runtimeDigest...),
		RuntimeDigest:           append([]byte(nil), runtimeDigest...),
		ModelManifestDigest:     append([]byte(nil), manifestDigest...),
		ModelDigest:             computeInferenceModelEntriesDigest(manifest),
		DeterminismConfigDigest: types.InferencePipelineDeterminismConfigDigest(pipeline.DeterminismConfig),
		FeatureSchemaDigest:     computeInferenceSchemaDigest(),
		ActivationHeight:        pipeline.ActivatedAtHeight,
	}
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (k Keeper) validateInferenceProfileSnapshotReference(
	ctx sdk.Context,
	request *types.VerificationRequest,
) (*types.InferenceProfileSnapshot, error) {
	snapshot, err := request.RequireInferenceProfileSnapshot()
	if err != nil {
		return nil, err
	}
	pipeline, found := k.GetPipelineVersion(ctx, snapshot.PipelineVersion)
	if !found {
		return nil, types.ErrPipelineVersionNotFound.Wrapf("snapshot pipeline version %s not found", snapshot.PipelineVersion)
	}
	if err := pipeline.Validate(); err != nil {
		return nil, types.ErrInvalidPipelineVersion.Wrapf("snapshot pipeline validation failed: %v", err)
	}
	switch types.PipelineVersionStatus(pipeline.Status) {
	case types.PipelineVersionStatusActive, types.PipelineVersionStatusDeprecated:
	default:
		return nil, types.ErrPipelineVersionMismatch.Wrapf(
			"snapshot pipeline version %s is %s, not active or deprecated",
			pipeline.Version,
			pipeline.Status,
		)
	}
	if pipeline.ActivatedAt == nil || pipeline.ActivatedAtHeight <= 0 {
		return nil, types.ErrPipelineVersionMismatch.Wrapf("snapshot pipeline version %s has not been activated", pipeline.Version)
	}
	if ctx.BlockHeight() < pipeline.ActivatedAtHeight {
		return nil, types.ErrPipelineVersionMismatch.Wrapf(
			"snapshot pipeline version %s is not active until height %d",
			pipeline.Version,
			pipeline.ActivatedAtHeight,
		)
	}
	if pipeline.ActivatedAtHeight != snapshot.ActivationHeight {
		return nil, types.ErrPipelineVersionMismatch.Wrap("snapshot pipeline activation height mismatch")
	}
	if !types.IsStrictInferencePipelineDeterminismConfig(pipeline.DeterminismConfig) {
		return nil, types.ErrDeterminismViolation.Wrap("snapshot pipeline determinism config is not strict production profile")
	}

	manifestHash := hex.EncodeToString(snapshot.ModelManifestDigest)
	manifest, found := k.GetModelManifest(ctx, manifestHash)
	if !found {
		return nil, types.ErrModelManifestMismatch.Wrapf("snapshot model manifest %s not found", manifestHash)
	}
	if err := manifest.Validate(); err != nil {
		return nil, types.ErrModelManifestMismatch.Wrapf("snapshot model manifest validation failed: %v", err)
	}
	if normalizeHashString(pipeline.ModelManifest.ManifestHash) != manifestHash ||
		normalizeHashString(manifest.ManifestHash) != manifestHash {
		return nil, types.ErrModelManifestMismatch.Wrap("snapshot model manifest binding mismatch")
	}

	expected, err := k.inferenceProfileSnapshotFromPipeline(pipeline, manifest)
	if err != nil {
		return nil, err
	}
	if !inferenceProfileSnapshotsEqual(snapshot, expected) {
		return nil, types.ErrPipelineVersionMismatch.Wrap("stored inference profile snapshot does not match committed pipeline reference")
	}
	return snapshot, nil
}

func (k Keeper) requireRequestSnapshotMatchesCurrentBundle(ctx sdk.Context, request *types.VerificationRequest) error {
	snapshot, err := k.validateInferenceProfileSnapshotReference(ctx, request)
	if err != nil {
		return err
	}
	expected, err := k.VoteExtensionCommitments(ctx)
	if err != nil {
		return err
	}
	if expected.PipelineVersion == noActivePipelineVersion {
		return types.ErrNoPipelineVersionActive
	}
	currentSnapshot, err := k.activeInferenceProfileSnapshot(ctx)
	if err != nil {
		return err
	}
	if snapshot.PipelineVersion != expected.PipelineVersion ||
		!bytes.Equal(snapshot.RuntimeDigest, expected.RuntimeHash) ||
		!bytes.Equal(snapshot.RuntimeImageDigest, expected.RuntimeHash) ||
		!bytes.Equal(snapshot.ModelManifestDigest, expected.ModelHash) ||
		!inferenceProfileSnapshotsEqual(snapshot, currentSnapshot) {
		return types.ErrPipelineVersionMismatch.Wrap("request inference profile no longer matches the active vote-extension bundle")
	}
	return nil
}

func inferenceProfileSnapshotsEqual(a, b *types.InferenceProfileSnapshot) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.PipelineVersion == b.PipelineVersion &&
		a.ActivationHeight == b.ActivationHeight &&
		bytes.Equal(a.RuntimeImageDigest, b.RuntimeImageDigest) &&
		bytes.Equal(a.RuntimeDigest, b.RuntimeDigest) &&
		bytes.Equal(a.ModelManifestDigest, b.ModelManifestDigest) &&
		bytes.Equal(a.ModelDigest, b.ModelDigest) &&
		bytes.Equal(a.DeterminismConfigDigest, b.DeterminismConfigDigest) &&
		bytes.Equal(a.FeatureSchemaDigest, b.FeatureSchemaDigest)
}

func (k Keeper) verifyInferenceReceipt(
	ctx sdk.Context,
	request *types.VerificationRequest,
	receipt types.InferenceReceipt,
	expectations inferenceReceiptExpectations,
) (inferenceReceiptReplayCheck, error) {
	if err := receipt.Validate(); err != nil {
		return inferenceReceiptReplayCheck{}, err
	}
	if receipt.ChainID != ctx.ChainID() || receipt.AccountAddress != request.AccountAddress || receipt.RequestID != request.RequestID {
		return inferenceReceiptReplayCheck{}, types.ErrInvalidVerificationResult.Wrap("inference receipt request binding mismatch")
	}
	if err := validateInferenceReceiptFreshness(ctx, request, receipt); err != nil {
		return inferenceReceiptReplayCheck{}, err
	}
	if !equalStringSlices(receipt.ScopeIDs, expectations.ScopeIDs) ||
		!bytes.Equal(receipt.InputDigest, expectations.InputDigest) ||
		!bytes.Equal(receipt.FeatureDigest, expectations.FeatureDigest) ||
		!bytes.Equal(receipt.SchemaDigest, expectations.SchemaDigest) ||
		!bytes.Equal(receipt.EvidenceLineageDigest, expectations.EvidenceLineageDigest) ||
		receipt.PipelineVersion != expectations.PipelineVersion ||
		!bytes.Equal(receipt.ModelManifestDigest, expectations.ModelManifestDigest) ||
		!bytes.Equal(receipt.ModelDigest, expectations.ModelDigest) ||
		!bytes.Equal(receipt.RuntimeImageDigest, expectations.RuntimeImageDigest) ||
		!bytes.Equal(receipt.RuntimeDigest, expectations.RuntimeDigest) ||
		!bytes.Equal(receipt.ConfigDigest, expectations.ConfigDigest) ||
		receipt.DeterminismProfile != expectations.DeterminismProfile {
		return inferenceReceiptReplayCheck{}, types.ErrInvalidVerificationResult.Wrap("inference receipt commitment mismatch")
	}

	key, err := k.resolveSignerKey(ctx, receipt.SignerKeyID, receipt.SignerFingerprint)
	if err != nil {
		return inferenceReceiptReplayCheck{}, err
	}
	if err := validateInferenceReceiptSigner(ctx, key, receipt); err != nil {
		return inferenceReceiptReplayCheck{}, err
	}
	if err := receipt.VerifySignature(ed25519.PublicKey(key.PublicKey)); err != nil {
		return inferenceReceiptReplayCheck{}, err
	}
	return k.checkInferenceReceiptReplay(ctx, receipt)
}

func validateInferenceReceiptSigner(ctx sdk.Context, key *types.SignerKeyInfo, receipt types.InferenceReceipt) error {
	if key == nil {
		return types.ErrSignerKeyNotFound.Wrap("missing inference signer key")
	}
	if err := key.Validate(); err != nil {
		return err
	}
	if key.Algorithm != types.ProofTypeEd25519 {
		return types.ErrInvalidSignerKey.Wrap("inference receipt signer must use Ed25519")
	}
	if len(key.PublicKey) != ed25519.PublicKeySize {
		return types.ErrInvalidSignerKey.Wrap("inference receipt signer public key must be 32 bytes")
	}
	if key.KeyID != receipt.SignerKeyID || key.Fingerprint != receipt.SignerFingerprint || key.SequenceNumber != receipt.SignerSequence {
		return types.ErrInvalidSignerKey.Wrap("inference signer key binding mismatch")
	}
	if !key.State.CanVerify() {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is not active or rotating")
	}
	if key.ActivatedAt == nil || receipt.IssuedAt.Before(key.ActivatedAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("inference receipt predates signer activation")
	}
	if key.ActivatedAt != nil && ctx.BlockTime().UTC().Before(key.ActivatedAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is not active yet")
	}
	if key.ExpiresAt != nil && !receipt.IssuedAt.Before(key.ExpiresAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is expired")
	}
	if key.ExpiresAt != nil && !ctx.BlockTime().UTC().Before(key.ExpiresAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is expired")
	}
	if key.RevokedAt != nil && !receipt.IssuedAt.Before(key.RevokedAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is revoked")
	}
	if key.RevokedAt != nil && !ctx.BlockTime().UTC().Before(key.RevokedAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is revoked")
	}
	if !signerKeyAllowsInferenceReceipt(key) {
		return types.ErrInvalidSignerKey.Wrap("signer key is not authorized for inference receipts")
	}
	if err := validateInferenceSignerHeightPolicy(ctx, key, receipt); err != nil {
		return err
	}
	return nil
}

func signerKeyAllowsInferenceReceipt(key *types.SignerKeyInfo) bool {
	if key.Metadata == nil {
		return false
	}
	for _, item := range strings.Split(key.Metadata[types.SignerKeyMetadataEvidenceTypes], ",") {
		if strings.TrimSpace(item) == string(types.AttestationTypeInferenceReceipt) {
			return true
		}
	}
	return false
}

func validateInferenceSignerHeightPolicy(ctx sdk.Context, key *types.SignerKeyInfo, receipt types.InferenceReceipt) error {
	activationHeight, hasActivation, err := parseInferenceSignerHeight(key.Metadata[types.SignerKeyMetadataActivationHeight], "activation")
	if err != nil {
		return err
	}
	if hasActivation && receipt.IssuedHeight < activationHeight {
		return types.ErrInvalidSignerKey.Wrap("inference receipt predates signer activation height")
	}
	if hasActivation && ctx.BlockHeight() < activationHeight {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is not active by height")
	}
	expiryHeight, hasExpiry, err := parseInferenceSignerHeight(key.Metadata[types.SignerKeyMetadataExpiryHeight], "expiry")
	if err != nil {
		return err
	}
	if hasExpiry && receipt.IssuedHeight >= expiryHeight {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is expired by height")
	}
	if hasExpiry && ctx.BlockHeight() >= expiryHeight {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is expired by height")
	}
	revokedHeight, hasRevoked, err := parseInferenceSignerHeight(key.Metadata[types.SignerKeyMetadataRevokedHeight], "revoked")
	if err != nil {
		return err
	}
	if hasRevoked && receipt.IssuedHeight >= revokedHeight {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is revoked by height")
	}
	if hasRevoked && ctx.BlockHeight() >= revokedHeight {
		return types.ErrInvalidSignerKey.Wrap("inference signer key is revoked by height")
	}
	if ctx.BlockHeight() < receipt.IssuedHeight {
		return types.ErrInvalidTimestamp.Wrap("inference receipt issued height is from the future")
	}
	return nil
}

func parseInferenceSignerHeight(raw string, name string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}
	height, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || height < 0 {
		return 0, false, types.ErrInvalidSignerKey.Wrapf("invalid inference signer %s height", name)
	}
	return height, true, nil
}

func validateInferenceReceiptFreshness(ctx sdk.Context, request *types.VerificationRequest, receipt types.InferenceReceipt) error {
	now := ctx.BlockTime().UTC()
	if receipt.IssuedHeight != ctx.BlockHeight() {
		return types.ErrInvalidTimestamp.Wrap("inference receipt issued height is not current")
	}
	if receipt.IssuedHeight < request.RequestedBlock {
		return types.ErrInvalidTimestamp.Wrap("inference receipt predates request")
	}
	if receipt.IssuedAt.UTC().After(now) {
		return types.ErrInvalidTimestamp.Wrap("inference receipt issued time is from the future")
	}
	if now.Sub(receipt.IssuedAt.UTC()) > inferenceReceiptMaxAge {
		return types.ErrInvalidTimestamp.Wrap("inference receipt is stale")
	}
	if !receipt.ExpiresAt.UTC().After(now) {
		return types.ErrInvalidTimestamp.Wrap("inference receipt is expired")
	}
	if receipt.ExpiresAt.UTC().Sub(receipt.IssuedAt.UTC()) > types.InferenceReceiptMaxLifetime {
		return types.ErrInvalidTimestamp.Wrap("inference receipt lifetime exceeds maximum")
	}
	if receipt.ExpiresHeight <= ctx.BlockHeight() {
		return types.ErrInvalidTimestamp.Wrap("inference receipt expiry height is expired")
	}
	if receipt.ExpiresHeight-receipt.IssuedHeight > types.InferenceReceiptMaxHeightLifetime {
		return types.ErrInvalidTimestamp.Wrap("inference receipt height lifetime exceeds maximum")
	}
	return nil
}

func validateInferenceReceiptScopeResults(
	receipt types.InferenceReceipt,
	requiredScopeIDs []string,
	scopeResults []types.ScopeVerificationResult,
) error {
	if !equalStringSlices(receipt.ScopeIDs, requiredScopeIDs) {
		return types.ErrInvalidVerificationResult.Wrap("inference receipt scope binding mismatch")
	}
	if len(scopeResults) != len(requiredScopeIDs) {
		return types.ErrInvalidVerificationResult.Wrap("inference receipt scope result count mismatch")
	}
	successes := 0
	for i, result := range scopeResults {
		if result.ScopeID != requiredScopeIDs[i] {
			return types.ErrInvalidVerificationResult.Wrap("inference receipt scope result order mismatch")
		}
		if result.Success {
			successes++
		}
	}
	if successes != len(requiredScopeIDs) {
		return types.ErrInvalidVerificationResult.Wrap("inference receipt requires all scope payloads validated")
	}
	return nil
}

func (k Keeper) checkInferenceReceiptReplay(_ sdk.Context, receipt types.InferenceReceipt) (inferenceReceiptReplayCheck, error) {
	receiptDigest, err := receipt.Digest()
	if err != nil {
		return inferenceReceiptReplayCheck{}, err
	}
	contextDigest, err := receipt.ContextDigest()
	if err != nil {
		return inferenceReceiptReplayCheck{}, err
	}
	replay := inferenceReceiptReplayCheck{
		ContextDigest: hex.EncodeToString(contextDigest),
		ReceiptDigest: hex.EncodeToString(receiptDigest),
		NonceDigest:   computeInferenceReceiptNonceDigest(receipt),
	}
	return replay, nil
}

func inferenceReplayResultsMatch(existing, expected types.VerificationResult) bool {
	return existing.RequestID == expected.RequestID &&
		existing.AccountAddress == expected.AccountAddress &&
		existing.Score == expected.Score &&
		existing.Status == expected.Status &&
		existing.ModelVersion == expected.ModelVersion &&
		existing.BlockHeight == expected.BlockHeight &&
		bytes.Equal(existing.InputHash, expected.InputHash) &&
		equalReasonCodes(existing.ReasonCodes, expected.ReasonCodes) &&
		equalScopeVerificationResults(existing.ScopeResults, expected.ScopeResults) &&
		existing.Metadata[types.VerificationResultMetadataReceiptDigest] == expected.Metadata[types.VerificationResultMetadataReceiptDigest] &&
		existing.Metadata[types.VerificationResultMetadataReceiptContextDigest] == expected.Metadata[types.VerificationResultMetadataReceiptContextDigest]
}

func equalReasonCodes(a, b []types.ReasonCode) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalScopeVerificationResults(a, b []types.ScopeVerificationResult) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ScopeID != b[i].ScopeID ||
			a[i].ScopeType != b[i].ScopeType ||
			a[i].Success != b[i].Success ||
			a[i].Score != b[i].Score ||
			a[i].Weight != b[i].Weight ||
			a[i].Details != b[i].Details ||
			!equalReasonCodes(a[i].ReasonCodes, b[i].ReasonCodes) {
			return false
		}
	}
	return true
}

func (k Keeper) verificationResultFromReceipt(
	ctx sdk.Context,
	request *types.VerificationRequest,
	receipt types.InferenceReceipt,
	scopeResults []types.ScopeVerificationResult,
	replay inferenceReceiptReplayCheck,
) *types.VerificationResult {
	result := types.NewVerificationResult(request.RequestID, request.AccountAddress, ctx.BlockTime(), ctx.BlockHeight())
	result.Score = receipt.Score
	result.Status = receipt.Status
	result.ModelVersion = receipt.PipelineVersion
	result.InputHash = bytes.Clone(receipt.InputDigest)
	result.ReasonCodes = append([]types.ReasonCode(nil), receipt.ReasonCodes...)
	result.ScopeResults = append([]types.ScopeVerificationResult(nil), scopeResults...)
	result.Metadata[types.VerificationResultMetadataReceiptDigest] = replay.ReceiptDigest
	result.Metadata[types.VerificationResultMetadataReceiptContextDigest] = replay.ContextDigest
	result.Metadata[types.VerificationResultMetadataRuntimeDigest] = hex.EncodeToString(receipt.RuntimeDigest)
	result.Metadata[types.VerificationResultMetadataModelDigest] = hex.EncodeToString(receipt.ModelDigest)
	result.Metadata["pipeline_version"] = receipt.PipelineVersion
	result.Metadata["pipeline_manifest_hash"] = hex.EncodeToString(receipt.ModelManifestDigest)
	result.Metadata["pipeline_image_hash"] = hex.EncodeToString(receipt.RuntimeImageDigest)
	result.Metadata["inference_signer_key_id"] = receipt.SignerKeyID
	result.Metadata["inference_signer_fingerprint"] = receipt.SignerFingerprint
	return result
}

func computeInferenceInputDigest(account string, height int64, scopes []DecryptedScope) []byte {
	h := sha256.New()
	writeReceiptField(h, []byte(inferenceReceiptInputDomain))
	writeReceiptField(h, []byte(account))
	writeReceiptInt64(h, height)
	for _, scope := range sortedDecryptedScopes(scopes) {
		writeReceiptField(h, []byte(scope.ScopeID))
		writeReceiptField(h, []byte(scope.ScopeType))
		writeReceiptField(h, scope.ContentHash)
	}
	return h.Sum(nil)
}

func computeInferenceFeatureDigest(scopes []DecryptedScope) []byte {
	h := sha256.New()
	writeReceiptField(h, []byte(inferenceReceiptFeatureDomain))
	for _, scope := range sortedDecryptedScopes(scopes) {
		writeReceiptField(h, []byte(scope.ScopeID))
		writeReceiptField(h, []byte(scope.ScopeType))
		writeReceiptField(h, scope.ContentHash)
	}
	return h.Sum(nil)
}

func computeInferenceSchemaDigest() []byte {
	h := sha256.New()
	writeReceiptField(h, []byte(inferenceReceiptSchemaDomain))
	writeReceiptField(h, []byte(types.MLFeatureSchemaVersion))
	writeReceiptInt64(h, types.MLFeatureSchemaVersionMajor)
	writeReceiptInt64(h, types.MLFeatureSchemaVersionMinor)
	writeReceiptInt64(h, types.MLFeatureSchemaVersionPatch)
	writeReceiptInt64(h, types.FaceEmbeddingDim)
	writeReceiptInt64(h, types.DocQualityDim)
	writeReceiptInt64(h, types.OCRFieldCount)
	writeReceiptInt64(h, types.OCRFeaturesDim)
	writeReceiptInt64(h, types.MetadataFeaturesDim)
	writeReceiptInt64(h, types.PaddingDim)
	writeReceiptInt64(h, types.TotalFeatureDim)
	for _, offset := range types.FeatureOffsets() {
		writeReceiptField(h, []byte(offset.Group))
		writeReceiptInt64(h, int64(offset.StartIndex))
		writeReceiptInt64(h, int64(offset.EndIndex))
		writeReceiptInt64(h, int64(offset.Dimension))
	}
	for _, field := range types.OCRFieldNames() {
		writeReceiptField(h, []byte(field))
	}
	return h.Sum(nil)
}

func computeInferenceModelEntriesDigest(manifest *types.ModelManifest) []byte {
	h := sha256.New()
	writeReceiptField(h, []byte(inferenceReceiptModelDomain))
	writeReceiptField(h, []byte(manifest.Version))
	writeReceiptField(h, []byte(normalizeHashString(manifest.ManifestHash)))
	models := append([]types.ModelInfo(nil), manifest.Models...)
	sort.Slice(models, func(i, j int) bool {
		if models[i].Name == models[j].Name {
			return models[i].Version < models[j].Version
		}
		return models[i].Name < models[j].Name
	})
	for _, model := range models {
		writeReceiptField(h, []byte(model.Name))
		writeReceiptField(h, []byte(model.Version))
		writeReceiptField(h, []byte(normalizeHashString(model.WeightsHash)))
		writeReceiptField(h, []byte(normalizeHashString(model.ConfigHash)))
		writeReceiptField(h, []byte(model.Framework))
		writeReceiptField(h, []byte(model.Purpose))
		for _, dim := range model.InputShape {
			writeReceiptInt64(h, int64(dim))
		}
		writeReceiptField(h, nil)
		for _, dim := range model.OutputShape {
			writeReceiptInt64(h, int64(dim))
		}
		writeReceiptField(h, nil)
	}
	return h.Sum(nil)
}

func computeInferenceEvidenceLineageDigest(
	request *types.VerificationRequest,
	scopeResults []types.ScopeVerificationResult,
	scopes []DecryptedScope,
) []byte {
	h := sha256.New()
	writeReceiptField(h, []byte(inferenceReceiptLineageDomain))
	writeReceiptField(h, []byte(request.RequestID))
	writeReceiptField(h, []byte(request.AccountAddress))
	for _, scopeID := range types.CanonicalInferenceReceiptScopeIDs(request.ScopeIDs) {
		writeReceiptField(h, []byte(scopeID))
	}
	for _, scope := range sortedDecryptedScopes(scopes) {
		writeReceiptField(h, []byte(scope.ScopeID))
		writeReceiptField(h, []byte(scope.ScopeType))
		writeReceiptField(h, scope.ContentHash)
	}
	results := append([]types.ScopeVerificationResult(nil), scopeResults...)
	sort.Slice(results, func(i, j int) bool { return results[i].ScopeID < results[j].ScopeID })
	for _, result := range results {
		writeReceiptField(h, []byte(result.ScopeID))
		writeReceiptField(h, []byte(result.ScopeType))
		if result.Success {
			writeReceiptField(h, []byte{1})
		} else {
			writeReceiptField(h, []byte{0})
		}
		writeReceiptUint32(h, result.Score)
		for _, reason := range types.CanonicalInferenceReceiptReasonCodes(result.ReasonCodes) {
			writeReceiptField(h, []byte(reason))
		}
	}
	return h.Sum(nil)
}

func sortedDecryptedScopes(scopes []DecryptedScope) []DecryptedScope {
	out := append([]DecryptedScope(nil), scopes...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ScopeID == out[j].ScopeID {
			return out[i].ScopeType < out[j].ScopeType
		}
		return out[i].ScopeID < out[j].ScopeID
	})
	return out
}

func computeInferenceReceiptNonceDigest(receipt types.InferenceReceipt) string {
	h := sha256.New()
	writeReceiptField(h, []byte(inferenceReceiptReplayDomain))
	writeReceiptField(h, []byte(receipt.ChainID))
	writeReceiptField(h, []byte(receipt.SignerKeyID))
	writeReceiptUint64(h, receipt.SignerSequence)
	writeReceiptField(h, []byte(receipt.Nonce))
	return hex.EncodeToString(h.Sum(nil))
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func writeReceiptField(h hashWriter, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = h.Write(length[:])
	_, _ = h.Write(value)
}

func writeReceiptInt64(h hashWriter, value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], safeUint64FromInt64(value))
	_, _ = h.Write(encoded[:])
}

func writeReceiptUint32(h hashWriter, value uint32) {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	_, _ = h.Write(encoded[:])
}

func writeReceiptUint64(h hashWriter, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = h.Write(encoded[:])
}

type hashWriter interface {
	Write([]byte) (int, error)
}
