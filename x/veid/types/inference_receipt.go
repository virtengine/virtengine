package types

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const (
	InferenceReceiptDomain  = "VEID_INFERENCE_RECEIPT"
	InferenceReceiptVersion = uint32(1)

	InferenceReceiptSignDomain    = "VEID_INFERENCE_RECEIPT_SIGN_V1"
	InferenceReceiptDigestDomain  = "VEID_INFERENCE_RECEIPT_DIGEST_V1"
	InferenceReceiptContextDomain = "VEID_INFERENCE_RECEIPT_CONTEXT_V1"
	InferenceReceiptConfigDomain  = "VEID_INFERENCE_DETERMINISM_CONFIG_V1"

	InferenceReceiptMaxString                = 128
	InferenceReceiptMaxReasonCodes           = 16
	InferenceReceiptMaxScopes                = 32
	InferenceReceiptMaxConfidencePPM         = 1_000_000
	InferenceReceiptRequiredRandomSeed int64 = 42
	InferenceReceiptMaxLifetime               = 10 * time.Minute
	InferenceReceiptMaxHeightLifetime   int64 = 2
)

// InferenceDeterminismProfile is the bounded deterministic runtime profile
// bound into inference receipts.
type InferenceDeterminismProfile struct {
	ForceCPU         bool  `json:"force_cpu"`
	RandomSeed       int64 `json:"random_seed"`
	DeterministicOps bool  `json:"deterministic_ops"`
	InterOpThreads   int32 `json:"inter_op_threads"`
	IntraOpThreads   int32 `json:"intra_op_threads"`
	DisableGPU       bool  `json:"disable_gpu"`
}

// StrictInferencePipelineDeterminismConfig returns the only committed pipeline
// determinism config accepted for production inference receipts.
func StrictInferencePipelineDeterminismConfig() PipelineDeterminismConfig {
	return PipelineDeterminismConfig{
		RandomSeed:              InferenceReceiptRequiredRandomSeed,
		ForceCPU:                true,
		SingleThread:            true,
		FloatPrecision:          6,
		TensorFlowDeterministic: true,
		DisableCUDNN:            true,
		ONNXDeterministic:       true,
	}
}

// IsStrictInferencePipelineDeterminismConfig reports whether the committed
// pipeline config exactly matches the strict production receipt profile.
func IsStrictInferencePipelineDeterminismConfig(config PipelineDeterminismConfig) bool {
	return config == StrictInferencePipelineDeterminismConfig()
}

// InferenceDeterminismProfileFromPipelineConfig maps the committed pipeline
// determinism config into the bounded receipt profile carried by signers.
func InferenceDeterminismProfileFromPipelineConfig(config PipelineDeterminismConfig) InferenceDeterminismProfile {
	threads := int32(0)
	if config.SingleThread {
		threads = 1
	}
	return InferenceDeterminismProfile{
		ForceCPU:         config.ForceCPU,
		RandomSeed:       config.RandomSeed,
		DeterministicOps: config.TensorFlowDeterministic && config.ONNXDeterministic,
		InterOpThreads:   threads,
		IntraOpThreads:   threads,
		DisableGPU:       config.ForceCPU && config.DisableCUDNN,
	}
}

// CanonicalInferenceDeterminismProfile returns the only production profile
// currently eligible for consensus receipt validation.
func CanonicalInferenceDeterminismProfile() InferenceDeterminismProfile {
	return InferenceDeterminismProfileFromPipelineConfig(StrictInferencePipelineDeterminismConfig())
}

// CanonicalInferenceDeterminismConfigDigest returns the SHA-256 digest of the
// strict committed PipelineVersion determinism config.
func CanonicalInferenceDeterminismConfigDigest() []byte {
	return InferencePipelineDeterminismConfigDigest(StrictInferencePipelineDeterminismConfig())
}

// InferencePipelineDeterminismConfigDigest returns the domain-separated digest
// of the committed PipelineVersion determinism config, not a hard-coded local
// runtime profile.
func InferencePipelineDeterminismConfigDigest(config PipelineDeterminismConfig) []byte {
	env := struct {
		Domain                  string `json:"domain"`
		RandomSeed              int64  `json:"random_seed"`
		ForceCPU                bool   `json:"force_cpu"`
		SingleThread            bool   `json:"single_thread"`
		FloatPrecision          int32  `json:"float_precision"`
		TensorFlowDeterministic bool   `json:"tensorflow_deterministic"`
		DisableCUDNN            bool   `json:"disable_cudnn"`
		ONNXDeterministic       bool   `json:"onnx_deterministic"`
	}{
		Domain:                  InferenceReceiptConfigDomain,
		RandomSeed:              config.RandomSeed,
		ForceCPU:                config.ForceCPU,
		SingleThread:            config.SingleThread,
		FloatPrecision:          config.FloatPrecision,
		TensorFlowDeterministic: config.TensorFlowDeterministic,
		DisableCUDNN:            config.DisableCUDNN,
		ONNXDeterministic:       config.ONNXDeterministic,
	}
	bz, err := json.Marshal(env)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(bz)
	return sum[:]
}

// Digest returns the domain-separated digest of the deterministic profile.
func (p InferenceDeterminismProfile) Digest() []byte {
	env := struct {
		Domain           string `json:"domain"`
		ForceCPU         bool   `json:"force_cpu"`
		RandomSeed       int64  `json:"random_seed"`
		DeterministicOps bool   `json:"deterministic_ops"`
		InterOpThreads   int32  `json:"inter_op_threads"`
		IntraOpThreads   int32  `json:"intra_op_threads"`
		DisableGPU       bool   `json:"disable_gpu"`
	}{
		Domain:           InferenceReceiptConfigDomain,
		ForceCPU:         p.ForceCPU,
		RandomSeed:       p.RandomSeed,
		DeterministicOps: p.DeterministicOps,
		InterOpThreads:   p.InterOpThreads,
		IntraOpThreads:   p.IntraOpThreads,
		DisableGPU:       p.DisableGPU,
	}
	bz, err := json.Marshal(env)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(bz)
	return sum[:]
}

// IsCanonical returns true only for the strict production profile.
func (p InferenceDeterminismProfile) IsCanonical() bool {
	return p == CanonicalInferenceDeterminismProfile()
}

// InferenceReceipt is the bounded signed evidence produced by validator-local
// production inference before consensus.
type InferenceReceipt struct {
	Domain  string `json:"domain"`
	Version uint32 `json:"version"`

	ChainID        string   `json:"chain_id"`
	AccountAddress string   `json:"account_address"`
	RequestID      string   `json:"request_id"`
	ScopeIDs       []string `json:"scope_ids"`
	Nonce          string   `json:"nonce"`

	InputDigest           []byte `json:"input_digest"`
	FeatureDigest         []byte `json:"feature_digest"`
	SchemaDigest          []byte `json:"schema_digest"`
	EvidenceLineageDigest []byte `json:"evidence_lineage_digest"`

	PipelineVersion     string                      `json:"pipeline_version"`
	ModelManifestDigest []byte                      `json:"model_manifest_digest"`
	ModelDigest         []byte                      `json:"model_digest"`
	RuntimeImageDigest  []byte                      `json:"runtime_image_digest"`
	RuntimeDigest       []byte                      `json:"runtime_digest"`
	ConfigDigest        []byte                      `json:"config_digest"`
	DeterminismProfile  InferenceDeterminismProfile `json:"determinism_profile"`

	Score                uint32                   `json:"score"`
	Status               VerificationResultStatus `json:"status"`
	ConfidenceMillionths uint32                   `json:"confidence_millionths"`
	ReasonCodes          []ReasonCode             `json:"reason_codes"`
	IssuedHeight         int64                    `json:"issued_height"`
	IssuedAt             time.Time                `json:"issued_at"`
	ExpiresHeight        int64                    `json:"expires_height"`
	ExpiresAt            time.Time                `json:"expires_at"`
	SignerKeyID          string                   `json:"signer_key_id"`
	SignerFingerprint    string                   `json:"signer_fingerprint"`
	SignerSequence       uint64                   `json:"signer_sequence"`
	Signature            []byte                   `json:"signature"`
}

type inferenceReceiptSignEnvelope struct {
	Domain                string                      `json:"domain"`
	Version               uint32                      `json:"version"`
	ChainID               string                      `json:"chain_id"`
	AccountAddress        string                      `json:"account_address"`
	RequestID             string                      `json:"request_id"`
	ScopeIDs              []string                    `json:"scope_ids"`
	Nonce                 string                      `json:"nonce"`
	InputDigest           []byte                      `json:"input_digest"`
	FeatureDigest         []byte                      `json:"feature_digest"`
	SchemaDigest          []byte                      `json:"schema_digest"`
	EvidenceLineageDigest []byte                      `json:"evidence_lineage_digest"`
	PipelineVersion       string                      `json:"pipeline_version"`
	ModelManifestDigest   []byte                      `json:"model_manifest_digest"`
	ModelDigest           []byte                      `json:"model_digest"`
	RuntimeImageDigest    []byte                      `json:"runtime_image_digest"`
	RuntimeDigest         []byte                      `json:"runtime_digest"`
	ConfigDigest          []byte                      `json:"config_digest"`
	DeterminismProfile    InferenceDeterminismProfile `json:"determinism_profile"`
	Score                 uint32                      `json:"score"`
	Status                VerificationResultStatus    `json:"status"`
	ConfidenceMillionths  uint32                      `json:"confidence_millionths"`
	ReasonCodes           []ReasonCode                `json:"reason_codes"`
	IssuedHeight          int64                       `json:"issued_height"`
	IssuedAtUnix          int64                       `json:"issued_at_unix"`
	ExpiresHeight         int64                       `json:"expires_height"`
	ExpiresAtUnix         int64                       `json:"expires_at_unix"`
	SignerKeyID           string                      `json:"signer_key_id"`
	SignerFingerprint     string                      `json:"signer_fingerprint"`
	SignerSequence        uint64                      `json:"signer_sequence"`
}

type inferenceReceiptContextEnvelope struct {
	Domain                string                      `json:"domain"`
	Version               uint32                      `json:"version"`
	ChainID               string                      `json:"chain_id"`
	AccountAddress        string                      `json:"account_address"`
	RequestID             string                      `json:"request_id"`
	ScopeIDs              []string                    `json:"scope_ids"`
	Nonce                 string                      `json:"nonce"`
	InputDigest           []byte                      `json:"input_digest"`
	FeatureDigest         []byte                      `json:"feature_digest"`
	SchemaDigest          []byte                      `json:"schema_digest"`
	EvidenceLineageDigest []byte                      `json:"evidence_lineage_digest"`
	PipelineVersion       string                      `json:"pipeline_version"`
	ModelManifestDigest   []byte                      `json:"model_manifest_digest"`
	ModelDigest           []byte                      `json:"model_digest"`
	RuntimeImageDigest    []byte                      `json:"runtime_image_digest"`
	RuntimeDigest         []byte                      `json:"runtime_digest"`
	ConfigDigest          []byte                      `json:"config_digest"`
	DeterminismProfile    InferenceDeterminismProfile `json:"determinism_profile"`
	IssuedHeight          int64                       `json:"issued_height"`
	IssuedAtUnix          int64                       `json:"issued_at_unix"`
	ExpiresHeight         int64                       `json:"expires_height"`
	ExpiresAtUnix         int64                       `json:"expires_at_unix"`
	SignerKeyID           string                      `json:"signer_key_id"`
	SignerFingerprint     string                      `json:"signer_fingerprint"`
	SignerSequence        uint64                      `json:"signer_sequence"`
}

// CanonicalInferenceReceiptScopeIDs returns sorted unique scope IDs.
func CanonicalInferenceReceiptScopeIDs(scopeIDs []string) []string {
	out := append([]string(nil), scopeIDs...)
	sort.Strings(out)
	write := 0
	for _, scopeID := range out {
		if scopeID == "" {
			continue
		}
		if write > 0 && out[write-1] == scopeID {
			continue
		}
		out[write] = scopeID
		write++
	}
	return out[:write]
}

// CanonicalInferenceReceiptReasonCodes returns sorted unique reason codes.
func CanonicalInferenceReceiptReasonCodes(codes []ReasonCode) []ReasonCode {
	out := append([]ReasonCode(nil), codes...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	write := 0
	for _, code := range out {
		if code == "" {
			continue
		}
		if write > 0 && out[write-1] == code {
			continue
		}
		out[write] = code
		write++
	}
	return out[:write]
}

// SignBytes returns deterministic domain-separated bytes covered by the
// inference signer key.
func (r InferenceReceipt) SignBytes() ([]byte, error) {
	if err := r.validate(false); err != nil {
		return nil, err
	}
	env := inferenceReceiptSignEnvelope{
		Domain:                InferenceReceiptSignDomain,
		Version:               r.Version,
		ChainID:               r.ChainID,
		AccountAddress:        r.AccountAddress,
		RequestID:             r.RequestID,
		ScopeIDs:              append([]string(nil), r.ScopeIDs...),
		Nonce:                 r.Nonce,
		InputDigest:           append([]byte(nil), r.InputDigest...),
		FeatureDigest:         append([]byte(nil), r.FeatureDigest...),
		SchemaDigest:          append([]byte(nil), r.SchemaDigest...),
		EvidenceLineageDigest: append([]byte(nil), r.EvidenceLineageDigest...),
		PipelineVersion:       r.PipelineVersion,
		ModelManifestDigest:   append([]byte(nil), r.ModelManifestDigest...),
		ModelDigest:           append([]byte(nil), r.ModelDigest...),
		RuntimeImageDigest:    append([]byte(nil), r.RuntimeImageDigest...),
		RuntimeDigest:         append([]byte(nil), r.RuntimeDigest...),
		ConfigDigest:          append([]byte(nil), r.ConfigDigest...),
		DeterminismProfile:    r.DeterminismProfile,
		Score:                 r.Score,
		Status:                r.Status,
		ConfidenceMillionths:  r.ConfidenceMillionths,
		ReasonCodes:           append([]ReasonCode(nil), r.ReasonCodes...),
		IssuedHeight:          r.IssuedHeight,
		IssuedAtUnix:          r.IssuedAt.UTC().Unix(),
		ExpiresHeight:         r.ExpiresHeight,
		ExpiresAtUnix:         r.ExpiresAt.UTC().Unix(),
		SignerKeyID:           r.SignerKeyID,
		SignerFingerprint:     strings.ToLower(r.SignerFingerprint),
		SignerSequence:        r.SignerSequence,
	}
	return json.Marshal(env)
}

// Digest returns the SHA-256 digest of SignBytes.
func (r InferenceReceipt) Digest() ([]byte, error) {
	signBytes, err := r.SignBytes()
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, _ = h.Write([]byte(InferenceReceiptDigestDomain))
	_, _ = h.Write(signBytes)
	return h.Sum(nil), nil
}

// DigestHex returns the hex-encoded receipt digest.
func (r InferenceReceipt) DigestHex() (string, error) {
	digest, err := r.Digest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}

// ContextDigest returns the replay context digest over immutable execution
// context. It intentionally excludes output score/status/confidence/reasons
// and signature so same-context changed-output attempts are replay conflicts.
func (r InferenceReceipt) ContextDigest() ([]byte, error) {
	if err := r.validate(false); err != nil {
		return nil, err
	}
	env := inferenceReceiptContextEnvelope{
		Domain:                InferenceReceiptContextDomain,
		Version:               r.Version,
		ChainID:               r.ChainID,
		AccountAddress:        r.AccountAddress,
		RequestID:             r.RequestID,
		ScopeIDs:              append([]string(nil), r.ScopeIDs...),
		Nonce:                 r.Nonce,
		InputDigest:           append([]byte(nil), r.InputDigest...),
		FeatureDigest:         append([]byte(nil), r.FeatureDigest...),
		SchemaDigest:          append([]byte(nil), r.SchemaDigest...),
		EvidenceLineageDigest: append([]byte(nil), r.EvidenceLineageDigest...),
		PipelineVersion:       r.PipelineVersion,
		ModelManifestDigest:   append([]byte(nil), r.ModelManifestDigest...),
		ModelDigest:           append([]byte(nil), r.ModelDigest...),
		RuntimeImageDigest:    append([]byte(nil), r.RuntimeImageDigest...),
		RuntimeDigest:         append([]byte(nil), r.RuntimeDigest...),
		ConfigDigest:          append([]byte(nil), r.ConfigDigest...),
		DeterminismProfile:    r.DeterminismProfile,
		IssuedHeight:          r.IssuedHeight,
		IssuedAtUnix:          r.IssuedAt.UTC().Unix(),
		ExpiresHeight:         r.ExpiresHeight,
		ExpiresAtUnix:         r.ExpiresAt.UTC().Unix(),
		SignerKeyID:           r.SignerKeyID,
		SignerFingerprint:     strings.ToLower(r.SignerFingerprint),
		SignerSequence:        r.SignerSequence,
	}
	bz, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, _ = h.Write([]byte(InferenceReceiptContextDomain))
	_, _ = h.Write(bz)
	return h.Sum(nil), nil
}

// Sign signs the canonical receipt bytes with an Ed25519 private key.
func (r *InferenceReceipt) Sign(privateKey ed25519.PrivateKey) error {
	if len(privateKey) != ed25519.PrivateKeySize {
		return ErrInvalidSignerKey.Wrap("ed25519 private key must be 64 bytes")
	}
	signBytes, err := r.SignBytes()
	if err != nil {
		return err
	}
	r.Signature = ed25519.Sign(privateKey, signBytes)
	return nil
}

// VerifySignature verifies the receipt signature against an Ed25519 public key.
func (r InferenceReceipt) VerifySignature(publicKey ed25519.PublicKey) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return ErrInvalidSignerKey.Wrap("ed25519 public key must be 32 bytes")
	}
	if err := r.Validate(); err != nil {
		return err
	}
	signBytes, err := r.SignBytes()
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, signBytes, r.Signature) {
		return ErrInvalidSignerKey.Wrap("invalid inference receipt signature")
	}
	return nil
}

// Validate validates a signed receipt.
func (r InferenceReceipt) Validate() error {
	return r.validate(true)
}

func (r InferenceReceipt) validate(requireSignature bool) error {
	if r.Domain != InferenceReceiptDomain {
		return ErrInvalidVerificationResult.Wrap("invalid inference receipt domain")
	}
	if r.Version != InferenceReceiptVersion {
		return ErrInvalidVerificationResult.Wrap("unsupported inference receipt version")
	}
	for name, value := range map[string]string{
		"chain_id":           r.ChainID,
		"account_address":    r.AccountAddress,
		"request_id":         r.RequestID,
		"nonce":              r.Nonce,
		"pipeline_version":   r.PipelineVersion,
		"signer_key_id":      r.SignerKeyID,
		"signer_fingerprint": r.SignerFingerprint,
	} {
		if strings.TrimSpace(value) == "" {
			return ErrInvalidVerificationResult.Wrapf("inference receipt %s is required", name)
		}
		if len(value) > InferenceReceiptMaxString {
			return ErrInvalidVerificationResult.Wrapf("inference receipt %s exceeds %d bytes", name, InferenceReceiptMaxString)
		}
	}
	if _, err := hex.DecodeString(r.SignerFingerprint); err != nil || len(r.SignerFingerprint) != sha256.Size*2 {
		return ErrInvalidSignerKey.Wrap("signer fingerprint must be a SHA-256 hex digest")
	}
	if len(r.ScopeIDs) == 0 || len(r.ScopeIDs) > InferenceReceiptMaxScopes {
		return ErrInvalidVerificationResult.Wrap("invalid inference receipt scope count")
	}
	for i, scopeID := range r.ScopeIDs {
		if scopeID == "" || len(scopeID) > InferenceReceiptMaxString {
			return ErrInvalidVerificationResult.Wrap("invalid inference receipt scope id")
		}
		if i > 0 && scopeID <= r.ScopeIDs[i-1] {
			return ErrInvalidVerificationResult.Wrap("inference receipt scope ids must be strictly sorted")
		}
	}
	for name, digest := range map[string][]byte{
		"input_digest":            r.InputDigest,
		"feature_digest":          r.FeatureDigest,
		"schema_digest":           r.SchemaDigest,
		"evidence_lineage_digest": r.EvidenceLineageDigest,
		"model_manifest_digest":   r.ModelManifestDigest,
		"model_digest":            r.ModelDigest,
		"runtime_image_digest":    r.RuntimeImageDigest,
		"runtime_digest":          r.RuntimeDigest,
		"config_digest":           r.ConfigDigest,
	} {
		if len(digest) != sha256.Size {
			return ErrInvalidVerificationResult.Wrapf("%s must be SHA-256", name)
		}
	}
	if !r.DeterminismProfile.IsCanonical() {
		return ErrDeterminismViolation.Wrap("inference receipt deterministic profile is not canonical")
	}
	if !bytes.Equal(r.ConfigDigest, CanonicalInferenceDeterminismConfigDigest()) {
		return ErrDeterminismViolation.Wrap("inference receipt config digest mismatch")
	}
	if r.Score > MaxScore {
		return ErrInvalidVerificationResult.Wrap("inference receipt score exceeds maximum")
	}
	if !IsValidVerificationResultStatus(r.Status) {
		return ErrInvalidVerificationResult.Wrap("invalid inference receipt status")
	}
	if r.ConfidenceMillionths > InferenceReceiptMaxConfidencePPM {
		return ErrInvalidVerificationResult.Wrap("inference receipt confidence exceeds maximum")
	}
	if len(r.ReasonCodes) == 0 || len(r.ReasonCodes) > InferenceReceiptMaxReasonCodes {
		return ErrInvalidVerificationResult.Wrap("invalid inference receipt reason code count")
	}
	for i, code := range r.ReasonCodes {
		if code == "" || len(code) > 64 {
			return ErrInvalidVerificationResult.Wrap("invalid inference receipt reason code")
		}
		if !IsCanonicalInferenceReceiptReasonCode(code) {
			return ErrInvalidVerificationResult.Wrap("non-canonical inference receipt reason code")
		}
		if i > 0 && code <= r.ReasonCodes[i-1] {
			return ErrInvalidVerificationResult.Wrap("inference receipt reason codes must be strictly sorted")
		}
	}
	if err := ValidateInferenceReceiptResultSemantics(r.Status, r.Score, r.ReasonCodes); err != nil {
		return err
	}
	if r.IssuedHeight <= 0 || r.ExpiresHeight <= r.IssuedHeight {
		return ErrInvalidTimestamp.Wrap("invalid inference receipt height bounds")
	}
	if r.ExpiresHeight-r.IssuedHeight > InferenceReceiptMaxHeightLifetime {
		return ErrInvalidTimestamp.Wrap("inference receipt height lifetime exceeds maximum")
	}
	if r.IssuedAt.IsZero() || r.ExpiresAt.IsZero() || !r.ExpiresAt.After(r.IssuedAt) {
		return ErrInvalidTimestamp.Wrap("invalid inference receipt time bounds")
	}
	if r.IssuedAt.Location() != time.UTC || r.ExpiresAt.Location() != time.UTC {
		return ErrInvalidTimestamp.Wrap("inference receipt times must use UTC")
	}
	if r.IssuedAt.Nanosecond() != 0 || r.ExpiresAt.Nanosecond() != 0 {
		return ErrInvalidTimestamp.Wrap("inference receipt times must be second-aligned")
	}
	if r.ExpiresAt.Sub(r.IssuedAt) > InferenceReceiptMaxLifetime {
		return ErrInvalidTimestamp.Wrap("inference receipt lifetime exceeds maximum")
	}
	if r.SignerSequence == 0 {
		return ErrInvalidSignerKey.Wrap("signer sequence is required")
	}
	if requireSignature && len(r.Signature) != ed25519.SignatureSize {
		return ErrInvalidSignerKey.Wrap("inference receipt signature must be 64 bytes")
	}
	return nil
}

// IsCanonicalInferenceReceiptReasonCode restricts receipt outcomes to stable
// module reason codes that can be safely replayed and voted on.
func IsCanonicalInferenceReceiptReasonCode(code ReasonCode) bool {
	switch code {
	case ReasonCodeSuccess,
		ReasonCodeDecryptError,
		ReasonCodeInvalidScope,
		ReasonCodeScopeNotFound,
		ReasonCodeScopeRevoked,
		ReasonCodeScopeExpired,
		ReasonCodeMLInferenceError,
		ReasonCodeTimeout,
		ReasonCodeMaxRetriesExceeded,
		ReasonCodeInvalidPayload,
		ReasonCodeKeyNotFound,
		ReasonCodeInsufficientScopes,
		ReasonCodeFaceMismatch,
		ReasonCodeDocumentInvalid,
		ReasonCodeLivenessCheckFailed,
		ReasonCodeLowConfidence,
		ReasonCodeLowDocQuality,
		ReasonCodeLowOCRConfidence,
		ReasonCodeStaleArtifactState,
		ReasonCodeUnauthorizedArtifactState:
		return true
	default:
		return false
	}
}

// ValidateInferenceReceiptResultSemantics enforces coherent receipt status,
// score, and reason codes before a result can be staged or voted on.
func ValidateInferenceReceiptResultSemantics(status VerificationResultStatus, score uint32, reasons []ReasonCode) error {
	if !IsValidVerificationResultStatus(status) {
		return ErrInvalidVerificationResult.Wrap("invalid inference receipt status")
	}
	hasSuccess := false
	for _, reason := range reasons {
		if reason == ReasonCodeSuccess {
			hasSuccess = true
			break
		}
	}
	switch status {
	case VerificationResultStatusSuccess:
		if len(reasons) != 1 || !hasSuccess {
			return ErrInvalidVerificationResult.Wrap("successful inference receipt must carry only SUCCESS")
		}
	case VerificationResultStatusPartial:
		if score > MaxScore {
			return ErrInvalidVerificationResult.Wrap("partial inference receipt score exceeds maximum")
		}
		if hasSuccess {
			return ErrInvalidVerificationResult.Wrap("partial inference receipt cannot carry SUCCESS")
		}
	case VerificationResultStatusFailed, VerificationResultStatusError:
		if score != 0 {
			return ErrInvalidVerificationResult.Wrap("failed inference receipt score must be zero")
		}
		if hasSuccess {
			return ErrInvalidVerificationResult.Wrap("failed inference receipt cannot carry SUCCESS")
		}
	default:
		return ErrInvalidVerificationResult.Wrap("invalid inference receipt status")
	}
	return nil
}
