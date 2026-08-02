package inference

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ReceiptSigner signs receipt bytes without exposing private key material.
// Implementations must return immutable metadata for their lifetime.
type ReceiptSigner interface {
	KeyID() string
	Fingerprint() string
	Sequence() uint64
	Sign([]byte) ([]byte, error)
}

// ReceiptProductionRequest contains caller-supplied, pre-committed inference
// context. Production is pre-consensus and performs no chain, clock, or network
// lookups; consensus remains responsible only for receipt verification. The
// trusted caller must derive committed digest fields from the applicable
// RuntimePolicy and Profile. ExpectedScorerModelHash authenticates only scorer
// implementation metadata and is not the keeper-authoritative ModelDigest.
type ReceiptProductionRequest struct {
	ScoreInputs *ScoreInputs

	ChainID        string
	AccountAddress string
	RequestID      string
	ScopeIDs       []string
	Nonce          string

	InputDigest           []byte
	FeatureDigest         []byte
	SchemaDigest          []byte
	EvidenceLineageDigest []byte
	PipelineVersion       string
	ModelManifestDigest   []byte
	ModelDigest           []byte
	RuntimeImageDigest    []byte
	RuntimeDigest         []byte
	ConfigDigest          []byte

	IssuedHeight  int64
	IssuedAt      time.Time
	ExpiresHeight int64
	ExpiresAt     time.Time

	ExpectedScorerModelVersion string
	ExpectedScorerModelHash    string
}

// ReceiptProducer runs an injected scorer in-process and emits a signed,
// deterministic receipt. Its interfaces are intentionally network-free.
// Expected scorer metadata identifies the scorer implementation only; receipt
// commitments remain caller-supplied and keeper verification is authoritative.
type ReceiptProducer struct {
	scorer Scorer
	signer ReceiptSigner
}

// NewReceiptProducer creates a producer and rejects typed-nil dependencies.
func NewReceiptProducer(scorer Scorer, signer ReceiptSigner) (*ReceiptProducer, error) {
	if isNilInterface(scorer) {
		return nil, fmt.Errorf("receipt scorer is required")
	}
	if isNilInterface(signer) {
		return nil, fmt.Errorf("receipt signer is required")
	}
	return &ReceiptProducer{scorer: scorer, signer: signer}, nil
}

// Produce scores a detached input copy exactly once and signs the resulting
// receipt. Scope IDs are canonicalized from a private copy before use.
func (p *ReceiptProducer) Produce(request *ReceiptProductionRequest) (*veidtypes.InferenceReceipt, error) {
	if p == nil || isNilInterface(p.scorer) || isNilInterface(p.signer) {
		return nil, fmt.Errorf("receipt producer dependencies are required")
	}
	if request == nil {
		return nil, fmt.Errorf("receipt production request is required")
	}
	request = cloneReceiptProductionRequest(request)

	input := request.ScoreInputs
	if input == nil {
		return nil, fmt.Errorf("score inputs are required")
	}
	scopeIDs := veidtypes.CanonicalInferenceReceiptScopeIDs(request.ScopeIDs)
	keyID, fingerprint, sequence := p.signer.KeyID(), p.signer.Fingerprint(), p.signer.Sequence()
	if err := validateReceiptProductionRequest(request, scopeIDs, keyID, fingerprint, sequence); err != nil {
		return nil, err
	}
	if !p.scorer.IsHealthy() {
		return nil, fmt.Errorf("receipt scorer is unhealthy")
	}
	if p.scorer.GetModelVersion() != request.ExpectedScorerModelVersion {
		return nil, fmt.Errorf("scorer model version does not match expectation")
	}
	if p.scorer.GetModelHash() != request.ExpectedScorerModelHash {
		return nil, fmt.Errorf("scorer model hash does not match expectation")
	}

	result, err := p.scorer.ComputeScore(input)
	if err != nil {
		return nil, fmt.Errorf("compute inference score: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("scorer returned nil result")
	}
	if result.ModelVersion != request.ExpectedScorerModelVersion || result.ModelHash != request.ExpectedScorerModelHash {
		return nil, fmt.Errorf("scorer result model metadata does not match expectation")
	}
	if result.Score > veidtypes.MaxScore {
		return nil, fmt.Errorf("scorer result exceeds maximum score")
	}
	confidence, err := confidenceMillionths(result.Confidence)
	if err != nil {
		return nil, err
	}
	status, reasons, err := normalizeReceiptReasons(append([]string(nil), result.ReasonCodes...))
	if err != nil {
		return nil, err
	}

	receipt := veidtypes.InferenceReceipt{
		Domain:                veidtypes.InferenceReceiptDomain,
		Version:               veidtypes.InferenceReceiptVersion,
		ChainID:               request.ChainID,
		AccountAddress:        request.AccountAddress,
		RequestID:             request.RequestID,
		ScopeIDs:              scopeIDs,
		Nonce:                 request.Nonce,
		InputDigest:           cloneBytes(request.InputDigest),
		FeatureDigest:         cloneBytes(request.FeatureDigest),
		SchemaDigest:          cloneBytes(request.SchemaDigest),
		EvidenceLineageDigest: cloneBytes(request.EvidenceLineageDigest),
		PipelineVersion:       request.PipelineVersion,
		ModelManifestDigest:   cloneBytes(request.ModelManifestDigest),
		ModelDigest:           cloneBytes(request.ModelDigest),
		RuntimeImageDigest:    cloneBytes(request.RuntimeImageDigest),
		RuntimeDigest:         cloneBytes(request.RuntimeDigest),
		ConfigDigest:          cloneBytes(request.ConfigDigest),
		DeterminismProfile:    veidtypes.CanonicalInferenceDeterminismProfile(),
		Score:                 result.Score,
		Status:                status,
		ConfidenceMillionths:  confidence,
		ReasonCodes:           reasons,
		IssuedHeight:          request.IssuedHeight,
		IssuedAt:              request.IssuedAt,
		ExpiresHeight:         request.ExpiresHeight,
		ExpiresAt:             request.ExpiresAt,
		SignerKeyID:           keyID,
		SignerFingerprint:     fingerprint,
		SignerSequence:        sequence,
	}
	signBytes, err := receipt.SignBytes()
	if err != nil {
		return nil, fmt.Errorf("build inference receipt sign bytes: %w", err)
	}
	signature, err := p.signer.Sign(cloneBytes(signBytes))
	if err != nil {
		return nil, fmt.Errorf("sign inference receipt: %w", err)
	}
	if len(signature) != ed25519.SignatureSize {
		return nil, fmt.Errorf("inference receipt signature must be %d bytes", ed25519.SignatureSize)
	}
	receipt.Signature = cloneBytes(signature)
	if err := receipt.Validate(); err != nil {
		return nil, fmt.Errorf("validate produced inference receipt: %w", err)
	}
	return cloneInferenceReceipt(&receipt), nil
}

func validateReceiptProductionRequest(request *ReceiptProductionRequest, scopeIDs []string, keyID, fingerprint string, sequence uint64) error {
	for name, value := range map[string]string{
		"chain ID": request.ChainID, "account address": request.AccountAddress,
		"request ID": request.RequestID, "nonce": request.Nonce,
		"pipeline version": request.PipelineVersion, "signer key ID": keyID,
		"expected scorer model version": request.ExpectedScorerModelVersion,
	} {
		if strings.TrimSpace(value) == "" || len(value) > veidtypes.InferenceReceiptMaxString {
			return fmt.Errorf("%s is required and must not exceed %d bytes", name, veidtypes.InferenceReceiptMaxString)
		}
	}
	if len(scopeIDs) == 0 || len(scopeIDs) > veidtypes.InferenceReceiptMaxScopes {
		return fmt.Errorf("scope IDs must contain between 1 and %d unique values", veidtypes.InferenceReceiptMaxScopes)
	}
	for _, scopeID := range scopeIDs {
		if len(scopeID) > veidtypes.InferenceReceiptMaxString {
			return fmt.Errorf("scope ID exceeds %d bytes", veidtypes.InferenceReceiptMaxString)
		}
	}
	for name, digest := range map[string][]byte{
		"input digest": request.InputDigest, "feature digest": request.FeatureDigest,
		"schema digest": request.SchemaDigest, "evidence lineage digest": request.EvidenceLineageDigest,
		"model manifest digest": request.ModelManifestDigest, "model digest": request.ModelDigest,
		"runtime image digest": request.RuntimeImageDigest, "runtime digest": request.RuntimeDigest,
		"config digest": request.ConfigDigest,
	} {
		if len(digest) != sha256.Size {
			return fmt.Errorf("%s must be SHA-256", name)
		}
	}
	if !bytes.Equal(request.ConfigDigest, veidtypes.CanonicalInferenceDeterminismConfigDigest()) {
		return fmt.Errorf("config digest is not canonical")
	}
	if request.IssuedHeight <= 0 || request.ExpiresHeight <= request.IssuedHeight {
		return fmt.Errorf("receipt height bounds are invalid")
	}
	if request.ExpiresHeight-request.IssuedHeight > veidtypes.InferenceReceiptMaxHeightLifetime {
		return fmt.Errorf("receipt height lifetime exceeds maximum")
	}
	if request.IssuedAt.IsZero() || request.ExpiresAt.IsZero() || !request.ExpiresAt.After(request.IssuedAt) {
		return fmt.Errorf("receipt time bounds are invalid")
	}
	if request.IssuedAt.Location() != time.UTC || request.ExpiresAt.Location() != time.UTC {
		return fmt.Errorf("receipt times must use UTC")
	}
	if request.IssuedAt.Nanosecond() != 0 || request.ExpiresAt.Nanosecond() != 0 {
		return fmt.Errorf("receipt times must be second-aligned")
	}
	if request.ExpiresAt.Sub(request.IssuedAt) > veidtypes.InferenceReceiptMaxLifetime {
		return fmt.Errorf("receipt lifetime exceeds maximum")
	}
	if keyID == "" || sequence == 0 {
		return fmt.Errorf("signer metadata is invalid")
	}
	if fingerprint != strings.ToLower(fingerprint) || !isLowerSHA256Hex(fingerprint) {
		return fmt.Errorf("signer fingerprint must be lowercase SHA-256 hex")
	}
	if !isLowerSHA256Hex(request.ExpectedScorerModelHash) {
		return fmt.Errorf("expected scorer model hash must be lowercase SHA-256 hex without a prefix")
	}
	return nil
}

func normalizeReceiptReasons(reasonStrings []string) (veidtypes.VerificationResultStatus, []veidtypes.ReasonCode, error) {
	reasons := make([]veidtypes.ReasonCode, 0, len(reasonStrings))
	for _, reasonString := range reasonStrings {
		var reason veidtypes.ReasonCode
		switch reasonString {
		case ReasonCodeSuccess, ReasonCodeHighConfidence:
			continue
		case ReasonCodeLowConfidence:
			reason = veidtypes.ReasonCodeLowConfidence
		case ReasonCodeLowDocQuality:
			reason = veidtypes.ReasonCodeLowDocQuality
		case ReasonCodeLowOCRConfidence:
			reason = veidtypes.ReasonCodeLowOCRConfidence
		case ReasonCodeInsufficientScopes:
			reason = veidtypes.ReasonCodeInsufficientScopes
		case ReasonCodeFaceMismatch, ReasonCodeMissingFace:
			reason = veidtypes.ReasonCodeFaceMismatch
		case ReasonCodeMissingDocument:
			reason = veidtypes.ReasonCodeDocumentInvalid
		case ReasonCodeInferenceError, ReasonCodeModelLoadError, ReasonCodeMemoryLimit:
			reason = veidtypes.ReasonCodeMLInferenceError
		case ReasonCodeTimeout:
			reason = veidtypes.ReasonCodeTimeout
		default:
			return "", nil, fmt.Errorf("unsupported scorer reason code %q", reasonString)
		}
		reasons = append(reasons, reason)
	}
	if len(reasons) == 0 {
		return veidtypes.VerificationResultStatusSuccess, []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess}, nil
	}
	return veidtypes.VerificationResultStatusPartial, veidtypes.CanonicalInferenceReceiptReasonCodes(reasons), nil
}

func confidenceMillionths(confidence float32) (uint32, error) {
	value := float64(confidence)
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
		return 0, fmt.Errorf("scorer confidence must be finite and between zero and one")
	}
	return uint32(math.Floor(value*veidtypes.InferenceReceiptMaxConfidencePPM + 0.5)), nil
}

func isLowerSHA256Hex(value string) bool {
	if len(value) != sha256.Size*2 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func isNilInterface(value interface{}) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func cloneScoreInputs(inputs *ScoreInputs) *ScoreInputs {
	if inputs == nil {
		return nil
	}
	clone := *inputs
	clone.FaceEmbedding = append([]float32(nil), inputs.FaceEmbedding...)
	clone.OCRConfidences = make(map[string]float32, len(inputs.OCRConfidences))
	for key, value := range inputs.OCRConfidences {
		clone.OCRConfidences[key] = value
	}
	clone.OCRFieldValidation = make(map[string]bool, len(inputs.OCRFieldValidation))
	for key, value := range inputs.OCRFieldValidation {
		clone.OCRFieldValidation[key] = value
	}
	clone.ScopeTypes = append([]string(nil), inputs.ScopeTypes...)
	return &clone
}

func cloneReceiptProductionRequest(request *ReceiptProductionRequest) *ReceiptProductionRequest {
	clone := *request
	clone.ScoreInputs = cloneScoreInputs(request.ScoreInputs)
	clone.ScopeIDs = append([]string(nil), request.ScopeIDs...)
	clone.InputDigest = cloneBytes(request.InputDigest)
	clone.FeatureDigest = cloneBytes(request.FeatureDigest)
	clone.SchemaDigest = cloneBytes(request.SchemaDigest)
	clone.EvidenceLineageDigest = cloneBytes(request.EvidenceLineageDigest)
	clone.ModelManifestDigest = cloneBytes(request.ModelManifestDigest)
	clone.ModelDigest = cloneBytes(request.ModelDigest)
	clone.RuntimeImageDigest = cloneBytes(request.RuntimeImageDigest)
	clone.RuntimeDigest = cloneBytes(request.RuntimeDigest)
	clone.ConfigDigest = cloneBytes(request.ConfigDigest)
	return &clone
}

func cloneInferenceReceipt(receipt *veidtypes.InferenceReceipt) *veidtypes.InferenceReceipt {
	clone := *receipt
	clone.ScopeIDs = append([]string(nil), receipt.ScopeIDs...)
	clone.InputDigest = cloneBytes(receipt.InputDigest)
	clone.FeatureDigest = cloneBytes(receipt.FeatureDigest)
	clone.SchemaDigest = cloneBytes(receipt.SchemaDigest)
	clone.EvidenceLineageDigest = cloneBytes(receipt.EvidenceLineageDigest)
	clone.ModelManifestDigest = cloneBytes(receipt.ModelManifestDigest)
	clone.ModelDigest = cloneBytes(receipt.ModelDigest)
	clone.RuntimeImageDigest = cloneBytes(receipt.RuntimeImageDigest)
	clone.RuntimeDigest = cloneBytes(receipt.RuntimeDigest)
	clone.ConfigDigest = cloneBytes(receipt.ConfigDigest)
	clone.ReasonCodes = append([]veidtypes.ReasonCode(nil), receipt.ReasonCodes...)
	clone.Signature = cloneBytes(receipt.Signature)
	return &clone
}

func cloneBytes(value []byte) []byte {
	return append([]byte(nil), value...)
}
