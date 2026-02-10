package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Evidence Pipeline (Document + Biometric)
// ============================================================================

const (
	evidenceConfidenceThreshold uint32 = 7000
	evidenceLowConfidenceCutoff uint32 = 5000
)

// EvidenceAssessment summarizes evidence confidence and provenance for scoring.
type EvidenceAssessment struct {
	DocumentConfidence  uint32
	BiometricConfidence uint32
	OverallConfidence   uint32
	ProvenanceHash      []byte
}

// ProcessEvidencePipeline extracts evidence confidence and stores evidence records.
func (k Keeper) ProcessEvidencePipeline(
	ctx sdk.Context,
	address sdk.AccAddress,
	requestID string,
	decryptedScopes []DecryptedScope,
	keyProvider ValidatorKeyProvider,
) (*EvidenceAssessment, error) {
	if len(decryptedScopes) == 0 {
		return &EvidenceAssessment{}, nil
	}

	pipeline := NewFeatureExtractionPipeline(DefaultFeatureExtractionConfig())
	features, err := pipeline.ExtractFeatures(decryptedScopes, address.String(), ctx.BlockHeight(), ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	validatorKeyID := ""
	if keyProvider != nil {
		validatorKeyID = keyProvider.GetKeyFingerprint()
	}

	scopeIDs := make([]string, 0, len(decryptedScopes))
	for _, scope := range decryptedScopes {
		scopeIDs = append(scopeIDs, scope.ScopeID)
	}

	assessment := &EvidenceAssessment{
		DocumentConfidence:  computeDocumentEvidenceConfidence(features),
		BiometricConfidence: computeBiometricEvidenceConfidence(features),
		ProvenanceHash:      computeEvidenceProvenanceHash(features.Metadata, scopeIDs, validatorKeyID),
	}

	assessment.OverallConfidence = computeOverallConfidence(assessment.DocumentConfidence, assessment.BiometricConfidence)

	// Persist evidence records for document + biometric scopes.
	if err := k.storeEvidenceRecords(ctx, address, requestID, decryptedScopes, assessment, validatorKeyID); err != nil {
		return assessment, err
	}

	return assessment, nil
}

func (k Keeper) storeEvidenceRecords(
	ctx sdk.Context,
	address sdk.AccAddress,
	requestID string,
	decryptedScopes []DecryptedScope,
	assessment *EvidenceAssessment,
	validatorKeyID string,
) error {
	docScope := findScopeByType(decryptedScopes, types.ScopeTypeIDDocument)
	if docScope != nil {
		if err := k.createEvidenceRecord(ctx, address, requestID, docScope, types.EvidenceTypeDocument, assessment.DocumentConfidence, assessment.ProvenanceHash, validatorKeyID); err != nil {
			return err
		}
	}

	biometricScope := findBiometricScope(decryptedScopes)
	if biometricScope != nil {
		if err := k.createEvidenceRecord(ctx, address, requestID, biometricScope, types.EvidenceTypeBiometric, assessment.BiometricConfidence, assessment.ProvenanceHash, validatorKeyID); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) createEvidenceRecord(
	ctx sdk.Context,
	address sdk.AccAddress,
	requestID string,
	scope *DecryptedScope,
	evidenceType types.EvidenceType,
	confidence uint32,
	provenanceHash []byte,
	validatorKeyID string,
) error {
	storedScope, found := k.GetScope(ctx, address, scope.ScopeID)
	if !found {
		return types.ErrScopeNotFound.Wrapf("scope %s not found", scope.ScopeID)
	}

	envelopeHash := storedScope.EncryptedPayload.Hash()

	if validatorKeyID != "" && !storedScope.EncryptedPayload.IsRecipient(validatorKeyID) {
		return types.ErrUnauthorized.Wrap("validator not authorized for evidence envelope")
	}

	evidenceID := buildEvidenceID(address.String(), requestID, scope.ScopeID, evidenceType, envelopeHash)

	record := types.NewEvidenceRecord(
		evidenceID,
		evidenceType,
		address.String(),
		scope.ScopeID,
		scope.ContentHash,
		envelopeHash,
		storedScope.EncryptedPayload.RecipientKeyIDs,
		storedScope.EncryptedPayload.AlgorithmID,
		confidence,
		provenanceHash,
	)
	record.VerifierKeyID = validatorKeyID

	if confidence >= evidenceConfidenceThreshold {
		record.SetDecision(types.EvidenceStatusVerified, "evidence verified", ctx.BlockTime())
	} else {
		record.SetDecision(types.EvidenceStatusRejected, "evidence confidence below threshold", ctx.BlockTime())
	}

	if err := k.SetEvidenceRecord(ctx, record); err != nil {
		return err
	}

	details := map[string]interface{}{
		"evidence_id":     record.EvidenceID,
		"evidence_type":   record.EvidenceType,
		"scope_id":        record.ScopeID,
		"confidence":      record.Confidence,
		"status":          record.Status,
		"envelope_hash":   hex.EncodeToString(record.EnvelopeHash),
		"provenance_hash": record.ProvenanceHashHex(),
	}

	return k.RecordAuditEvent(ctx, types.AuditEventTypeEvidenceDecision, record.AccountAddress, details)
}

func buildEvidenceID(accountAddress string, requestID string, scopeID string, evidenceType types.EvidenceType, envelopeHash []byte) string {
	h := sha256.New()
	h.Write([]byte(accountAddress))
	h.Write([]byte(requestID))
	h.Write([]byte(scopeID))
	h.Write([]byte(evidenceType))
	h.Write(envelopeHash)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:16])
}

func computeEvidenceProvenanceHash(metadata *types.FeatureExtractionMetadata, scopeIDs []string, validatorKeyID string) []byte {
	h := sha256.New()

	if metadata != nil {
		h.Write([]byte(metadata.SchemaVersion))
		h.Write([]byte(metadata.FeatureHash))
	}

	if validatorKeyID != "" {
		h.Write([]byte(validatorKeyID))
	}

	sort.Strings(scopeIDs)
	for _, id := range scopeIDs {
		h.Write([]byte(id))
	}

	if metadata != nil {
		modelPairs := make([]string, 0, len(metadata.ModelsUsed))
		for _, model := range metadata.ModelsUsed {
			modelPairs = append(modelPairs, fmt.Sprintf("%s|%s|%s", model.ModelType, model.ModelVersion, model.ModelHash))
		}
		sort.Strings(modelPairs)
		for _, entry := range modelPairs {
			h.Write([]byte(entry))
		}
	}

	return h.Sum(nil)
}

func computeDocumentEvidenceConfidence(features *RealExtractedFeatures) uint32 {
	if features == nil {
		return 0
	}

	docQuality := floatToBasisPoints(features.DocQualityScore)
	ocrConfidence := averageConfidenceBP(features.OCRConfidences)

	if docQuality == 0 && ocrConfidence == 0 {
		return 0
	}

	if ocrConfidence == 0 {
		return docQuality
	}

	return (docQuality*60 + ocrConfidence*40) / 100
}

func computeBiometricEvidenceConfidence(features *RealExtractedFeatures) uint32 {
	if features == nil {
		return 0
	}

	faceConfidence := floatToBasisPoints(features.FaceConfidence)
	liveness := floatToBasisPoints(features.LivenessScore)

	if faceConfidence == 0 && liveness == 0 {
		return 0
	}

	if liveness == 0 {
		return faceConfidence
	}

	return (faceConfidence*70 + liveness*30) / 100
}

func computeOverallConfidence(docConfidence uint32, biometricConfidence uint32) uint32 {
	if docConfidence == 0 && biometricConfidence == 0 {
		return 0
	}
	if docConfidence == 0 {
		return biometricConfidence
	}
	if biometricConfidence == 0 {
		return docConfidence
	}
	return (docConfidence + biometricConfidence) / 2
}

func averageConfidenceBP(values map[string]float32) uint32 {
	if len(values) == 0 {
		return 0
	}

	var total float64
	for _, v := range values {
		total += float64(v)
	}

	avg := total / float64(len(values))
	return floatToBasisPoints(float32(avg))
}

func floatToBasisPoints(value float32) uint32 {
	if value <= 0 {
		return 0
	}
	if value >= 1.0 {
		return uint32(types.MaxBasisPoints)
	}
	return uint32(value * float32(types.MaxBasisPoints))
}

func findScopeByType(scopes []DecryptedScope, scopeType types.ScopeType) *DecryptedScope {
	for i := range scopes {
		if scopes[i].ScopeType == scopeType {
			return &scopes[i]
		}
	}
	return nil
}

func findBiometricScope(scopes []DecryptedScope) *DecryptedScope {
	// Prefer selfie, then face video, then generic biometric
	if scope := findScopeByType(scopes, types.ScopeTypeSelfie); scope != nil {
		return scope
	}
	if scope := findScopeByType(scopes, types.ScopeTypeFaceVideo); scope != nil {
		return scope
	}
	if scope := findScopeByType(scopes, types.ScopeTypeBiometric); scope != nil {
		return scope
	}
	return nil
}

// applyEvidenceConfidence adjusts a score based on evidence confidence.
func applyEvidenceConfidence(score uint32, confidence uint32) (uint32, bool) {
	if confidence == 0 {
		return score, false
	}

	adjusted64 := (uint64(score) * uint64(confidence)) / uint64(types.MaxBasisPoints)
	if adjusted64 > math.MaxUint32 {
		return math.MaxUint32, true
	}
	return uint32(adjusted64), true
}

// shouldFlagLowEvidenceConfidence returns true if confidence is below cutoff.
func shouldFlagLowEvidenceConfidence(confidence uint32) bool {
	return confidence > 0 && confidence < evidenceLowConfidenceCutoff
}
