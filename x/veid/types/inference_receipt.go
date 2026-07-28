package types

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
	"unicode/utf8"
)

const (
	InferenceReceiptDomain  = "VEID_INFERENCE_RECEIPT"
	InferenceReceiptVersion = uint32(1)

	InferenceReceiptSignDomain    = "VEID_INFERENCE_RECEIPT_SIGN_V1"
	InferenceReceiptDigestDomain  = "VEID_INFERENCE_RECEIPT_DIGEST_V1"
	InferenceReceiptContextDomain = "VEID_INFERENCE_RECEIPT_CONTEXT_V1"
	InferenceReceiptConfigDomain  = "VEID_INFERENCE_DETERMINISM_CONFIG_V1"
	InferenceReceiptProfileDomain = "VEID_INFERENCE_DETERMINISM_PROFILE_V1"

	InferenceReceiptMaxString                = 128
	InferenceReceiptMaxReasonCodes           = 16
	InferenceReceiptMaxScopes                = 32
	InferenceReceiptMaxConfidencePPM         = 1_000_000
	InferenceReceiptRequiredRandomSeed int64 = 42
	InferenceReceiptMaxSignedBytes           = 8 * 1024
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
	enc := newInferenceReceiptBinaryEncoder(InferenceReceiptConfigDomain)
	enc.writeInt64(config.RandomSeed)
	enc.writeBool(config.ForceCPU)
	enc.writeBool(config.SingleThread)
	enc.writeInt32(config.FloatPrecision)
	enc.writeBool(config.TensorFlowDeterministic)
	enc.writeBool(config.DisableCUDNN)
	enc.writeBool(config.ONNXDeterministic)
	sum := sha256.Sum256(enc.bytes())
	return sum[:]
}

// Digest returns the domain-separated digest of the deterministic profile.
func (p InferenceDeterminismProfile) Digest() []byte {
	enc := newInferenceReceiptBinaryEncoder(InferenceReceiptProfileDomain)
	writeInferenceDeterminismProfile(enc, p)
	sum := sha256.Sum256(enc.bytes())
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
	enc := newInferenceReceiptBinaryEncoder(InferenceReceiptSignDomain)
	r.writeBinaryPayload(enc, true, false)
	return enc.bytes(), nil
}

// CanonicalSignedBytes returns the bounded canonical receipt envelope carried
// in vote extensions. It includes the signature and is byte-stable.
func (r InferenceReceipt) CanonicalSignedBytes() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	enc := newInferenceReceiptBinaryEncoder(InferenceReceiptDomain)
	r.writeBinaryPayload(enc, true, true)
	bz := enc.bytes()
	if len(bz) == 0 || len(bz) > InferenceReceiptMaxSignedBytes {
		return nil, ErrInvalidVerificationResult.Wrapf("canonical inference receipt exceeds %d bytes", InferenceReceiptMaxSignedBytes)
	}
	return bz, nil
}

// CanonicalBytes is an alias for the signed receipt bytes carried by consensus.
func (r InferenceReceipt) CanonicalBytes() ([]byte, error) {
	return r.CanonicalSignedBytes()
}

// DecodeCanonicalSignedInferenceReceipt parses and validates a byte-exact
// canonical signed receipt envelope.
func DecodeCanonicalSignedInferenceReceipt(bz []byte) (InferenceReceipt, error) {
	if len(bz) == 0 {
		return InferenceReceipt{}, ErrInvalidVerificationResult.Wrap("canonical inference receipt is required")
	}
	if len(bz) > InferenceReceiptMaxSignedBytes {
		return InferenceReceipt{}, ErrInvalidVerificationResult.Wrapf("canonical inference receipt exceeds %d bytes", InferenceReceiptMaxSignedBytes)
	}
	dec := newInferenceReceiptBinaryDecoder(bz)
	domain, err := dec.readString(InferenceReceiptMaxString)
	if err != nil {
		return InferenceReceipt{}, err
	}
	if domain != InferenceReceiptDomain {
		return InferenceReceipt{}, ErrInvalidVerificationResult.Wrap("invalid inference receipt domain")
	}
	receipt, err := dec.readReceipt(domain, true, true)
	if err != nil {
		return InferenceReceipt{}, err
	}
	if dec.remaining() != 0 {
		return InferenceReceipt{}, ErrInvalidVerificationResult.Wrap("canonical inference receipt has trailing data")
	}
	if err := receipt.Validate(); err != nil {
		return InferenceReceipt{}, err
	}
	canonical, err := receipt.CanonicalSignedBytes()
	if err != nil {
		return InferenceReceipt{}, err
	}
	if !bytes.Equal(canonical, bz) {
		return InferenceReceipt{}, ErrInvalidVerificationResult.Wrap("inference receipt encoding is not canonical")
	}
	return receipt, nil
}

// Digest returns the domain-separated SHA-256 digest of the canonical signed
// receipt bytes, including the signature carried by consensus.
func (r InferenceReceipt) Digest() ([]byte, error) {
	receiptBytes, err := r.CanonicalSignedBytes()
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, _ = h.Write([]byte(InferenceReceiptDigestDomain))
	_, _ = h.Write(receiptBytes)
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
	enc := newInferenceReceiptBinaryEncoder(InferenceReceiptContextDomain)
	r.writeBinaryPayload(enc, false, false)
	h := sha256.New()
	_, _ = h.Write([]byte(InferenceReceiptContextDomain))
	_, _ = h.Write(enc.bytes())
	return h.Sum(nil), nil
}

type inferenceReceiptBinaryEncoder struct {
	buf bytes.Buffer
}

func newInferenceReceiptBinaryEncoder(domain string) *inferenceReceiptBinaryEncoder {
	enc := &inferenceReceiptBinaryEncoder{}
	enc.writeString(domain)
	return enc
}

func (e *inferenceReceiptBinaryEncoder) bytes() []byte {
	return append([]byte(nil), e.buf.Bytes()...)
}

func (e *inferenceReceiptBinaryEncoder) writeBool(value bool) {
	if value {
		e.buf.WriteByte(1)
		return
	}
	e.buf.WriteByte(0)
}

func (e *inferenceReceiptBinaryEncoder) writeUint32(value uint32) {
	var out [4]byte
	binary.BigEndian.PutUint32(out[:], value)
	e.buf.Write(out[:])
}

func (e *inferenceReceiptBinaryEncoder) writeInt32(value int32) {
	if err := binary.Write(&e.buf, binary.BigEndian, value); err != nil {
		panic("failed to encode canonical inference receipt int32: " + err.Error())
	}
}

func (e *inferenceReceiptBinaryEncoder) writeUint64(value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	e.buf.Write(out[:])
}

func (e *inferenceReceiptBinaryEncoder) writeInt64(value int64) {
	if err := binary.Write(&e.buf, binary.BigEndian, value); err != nil {
		panic("failed to encode canonical inference receipt int64: " + err.Error())
	}
}

func (e *inferenceReceiptBinaryEncoder) writeBytes(value []byte) {
	e.writeUint64(uint64(len(value)))
	e.buf.Write(value)
}

func (e *inferenceReceiptBinaryEncoder) writeString(value string) {
	e.writeBytes([]byte(value))
}

func (r InferenceReceipt) writeBinaryPayload(enc *inferenceReceiptBinaryEncoder, includeOutput bool, includeSignature bool) {
	enc.writeUint32(r.Version)
	enc.writeString(r.ChainID)
	enc.writeString(r.AccountAddress)
	enc.writeString(r.RequestID)
	enc.writeUint64(uint64(len(r.ScopeIDs)))
	for _, scopeID := range r.ScopeIDs {
		enc.writeString(scopeID)
	}
	enc.writeString(r.Nonce)
	enc.writeBytes(r.InputDigest)
	enc.writeBytes(r.FeatureDigest)
	enc.writeBytes(r.SchemaDigest)
	enc.writeBytes(r.EvidenceLineageDigest)
	enc.writeString(r.PipelineVersion)
	enc.writeBytes(r.ModelManifestDigest)
	enc.writeBytes(r.ModelDigest)
	enc.writeBytes(r.RuntimeImageDigest)
	enc.writeBytes(r.RuntimeDigest)
	enc.writeBytes(r.ConfigDigest)
	writeInferenceDeterminismProfile(enc, r.DeterminismProfile)
	if includeOutput {
		enc.writeUint32(r.Score)
		enc.writeString(string(r.Status))
		enc.writeUint32(r.ConfidenceMillionths)
		enc.writeUint64(uint64(len(r.ReasonCodes)))
		for _, reason := range r.ReasonCodes {
			enc.writeString(string(reason))
		}
	}
	enc.writeInt64(r.IssuedHeight)
	enc.writeInt64(r.IssuedAt.UTC().Unix())
	enc.writeInt64(r.ExpiresHeight)
	enc.writeInt64(r.ExpiresAt.UTC().Unix())
	enc.writeString(r.SignerKeyID)
	enc.writeString(strings.ToLower(r.SignerFingerprint))
	enc.writeUint64(r.SignerSequence)
	if includeSignature {
		enc.writeBytes(r.Signature)
	}
}

func writeInferenceDeterminismProfile(enc *inferenceReceiptBinaryEncoder, profile InferenceDeterminismProfile) {
	enc.writeBool(profile.ForceCPU)
	enc.writeInt64(profile.RandomSeed)
	enc.writeBool(profile.DeterministicOps)
	enc.writeInt32(profile.InterOpThreads)
	enc.writeInt32(profile.IntraOpThreads)
	enc.writeBool(profile.DisableGPU)
}

type inferenceReceiptBinaryDecoder struct {
	bz  []byte
	pos int
}

func newInferenceReceiptBinaryDecoder(bz []byte) *inferenceReceiptBinaryDecoder {
	return &inferenceReceiptBinaryDecoder{bz: bz}
}

func (d *inferenceReceiptBinaryDecoder) remaining() int {
	return len(d.bz) - d.pos
}

func (d *inferenceReceiptBinaryDecoder) readBool() (bool, error) {
	if d.remaining() < 1 {
		return false, ErrInvalidVerificationResult.Wrap("truncated canonical inference receipt")
	}
	value := d.bz[d.pos]
	d.pos++
	switch value {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, ErrInvalidVerificationResult.Wrap("invalid canonical inference receipt boolean")
	}
}

func (d *inferenceReceiptBinaryDecoder) readUint32() (uint32, error) {
	if d.remaining() < 4 {
		return 0, ErrInvalidVerificationResult.Wrap("truncated canonical inference receipt")
	}
	value := binary.BigEndian.Uint32(d.bz[d.pos : d.pos+4])
	d.pos += 4
	return value, nil
}

func (d *inferenceReceiptBinaryDecoder) readInt32() (int32, error) {
	if d.remaining() < 4 {
		return 0, ErrInvalidVerificationResult.Wrap("truncated canonical inference receipt")
	}
	var value int32
	if err := binary.Read(bytes.NewReader(d.bz[d.pos:d.pos+4]), binary.BigEndian, &value); err != nil {
		return 0, ErrInvalidVerificationResult.Wrap("invalid canonical inference receipt int32")
	}
	d.pos += 4
	return value, nil
}

func (d *inferenceReceiptBinaryDecoder) readUint64() (uint64, error) {
	if d.remaining() < 8 {
		return 0, ErrInvalidVerificationResult.Wrap("truncated canonical inference receipt")
	}
	value := binary.BigEndian.Uint64(d.bz[d.pos : d.pos+8])
	d.pos += 8
	return value, nil
}

func (d *inferenceReceiptBinaryDecoder) readInt64() (int64, error) {
	if d.remaining() < 8 {
		return 0, ErrInvalidVerificationResult.Wrap("truncated canonical inference receipt")
	}
	var value int64
	if err := binary.Read(bytes.NewReader(d.bz[d.pos:d.pos+8]), binary.BigEndian, &value); err != nil {
		return 0, ErrInvalidVerificationResult.Wrap("invalid canonical inference receipt int64")
	}
	d.pos += 8
	return value, nil
}

func (d *inferenceReceiptBinaryDecoder) readBytes(max int) ([]byte, error) {
	length, err := d.readUint64()
	if err != nil {
		return nil, err
	}
	maxUint, err := canonicalIntToUint64(max)
	if err != nil {
		return nil, err
	}
	if length > maxUint {
		return nil, ErrInvalidVerificationResult.Wrap("canonical inference receipt field exceeds limit")
	}
	lengthInt, err := canonicalUint64ToInt(length)
	if err != nil {
		return nil, err
	}
	if lengthInt > d.remaining() {
		return nil, ErrInvalidVerificationResult.Wrap("truncated canonical inference receipt")
	}
	value := append([]byte(nil), d.bz[d.pos:d.pos+lengthInt]...)
	d.pos += lengthInt
	return value, nil
}

func (d *inferenceReceiptBinaryDecoder) readFixedBytes(size int) ([]byte, error) {
	value, err := d.readBytes(size)
	if err != nil {
		return nil, err
	}
	if len(value) != size {
		return nil, ErrInvalidVerificationResult.Wrap("canonical inference receipt fixed field has invalid length")
	}
	return value, nil
}

func (d *inferenceReceiptBinaryDecoder) readString(max int) (string, error) {
	value, err := d.readBytes(max)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(value) {
		return "", ErrInvalidVerificationResult.Wrap("canonical inference receipt string is not valid UTF-8")
	}
	return string(value), nil
}

func (d *inferenceReceiptBinaryDecoder) readReceipt(domain string, includeOutput bool, includeSignature bool) (InferenceReceipt, error) {
	version, err := d.readUint32()
	if err != nil {
		return InferenceReceipt{}, err
	}
	receipt := InferenceReceipt{Domain: domain, Version: version}
	if receipt.ChainID, err = d.readString(InferenceReceiptMaxString); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.AccountAddress, err = d.readString(InferenceReceiptMaxString); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.RequestID, err = d.readString(InferenceReceiptMaxString); err != nil {
		return InferenceReceipt{}, err
	}
	scopeCount, err := d.readBoundedCount(InferenceReceiptMaxScopes, "scope count")
	if err != nil {
		return InferenceReceipt{}, err
	}
	receipt.ScopeIDs = make([]string, scopeCount)
	for i := range receipt.ScopeIDs {
		if receipt.ScopeIDs[i], err = d.readString(InferenceReceiptMaxString); err != nil {
			return InferenceReceipt{}, err
		}
	}
	if receipt.Nonce, err = d.readString(InferenceReceiptMaxString); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.InputDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.FeatureDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.SchemaDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.EvidenceLineageDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.PipelineVersion, err = d.readString(InferenceReceiptMaxString); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.ModelManifestDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.ModelDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.RuntimeImageDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.RuntimeDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.ConfigDigest, err = d.readFixedBytes(sha256.Size); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.DeterminismProfile, err = d.readInferenceDeterminismProfile(); err != nil {
		return InferenceReceipt{}, err
	}
	if includeOutput {
		if receipt.Score, err = d.readUint32(); err != nil {
			return InferenceReceipt{}, err
		}
		status, err := d.readString(InferenceReceiptMaxString)
		if err != nil {
			return InferenceReceipt{}, err
		}
		receipt.Status = VerificationResultStatus(status)
		if receipt.ConfidenceMillionths, err = d.readUint32(); err != nil {
			return InferenceReceipt{}, err
		}
		reasonCount, err := d.readBoundedCount(InferenceReceiptMaxReasonCodes, "reason code count")
		if err != nil {
			return InferenceReceipt{}, err
		}
		receipt.ReasonCodes = make([]ReasonCode, reasonCount)
		for i := range receipt.ReasonCodes {
			reason, err := d.readString(64)
			if err != nil {
				return InferenceReceipt{}, err
			}
			receipt.ReasonCodes[i] = ReasonCode(reason)
		}
	}
	if receipt.IssuedHeight, err = d.readInt64(); err != nil {
		return InferenceReceipt{}, err
	}
	issuedAt, err := d.readInt64()
	if err != nil {
		return InferenceReceipt{}, err
	}
	receipt.IssuedAt = time.Unix(issuedAt, 0).UTC()
	if receipt.ExpiresHeight, err = d.readInt64(); err != nil {
		return InferenceReceipt{}, err
	}
	expiresAt, err := d.readInt64()
	if err != nil {
		return InferenceReceipt{}, err
	}
	receipt.ExpiresAt = time.Unix(expiresAt, 0).UTC()
	if receipt.SignerKeyID, err = d.readString(InferenceReceiptMaxString); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.SignerFingerprint, err = d.readString(InferenceReceiptMaxString); err != nil {
		return InferenceReceipt{}, err
	}
	if receipt.SignerSequence, err = d.readUint64(); err != nil {
		return InferenceReceipt{}, err
	}
	if includeSignature {
		if receipt.Signature, err = d.readFixedBytes(ed25519.SignatureSize); err != nil {
			return InferenceReceipt{}, err
		}
	}
	return receipt, nil
}

func (d *inferenceReceiptBinaryDecoder) readBoundedCount(max int, name string) (int, error) {
	count, err := d.readUint64()
	if err != nil {
		return 0, err
	}
	maxUint, err := canonicalIntToUint64(max)
	if err != nil {
		return 0, err
	}
	if count > maxUint {
		return 0, ErrInvalidVerificationResult.Wrapf("canonical inference receipt %s exceeds limit", name)
	}
	countInt, err := canonicalUint64ToInt(count)
	if err != nil {
		return 0, err
	}
	return countInt, nil
}

func canonicalIntToUint64(value int) (uint64, error) {
	if value < 0 {
		return 0, ErrInvalidVerificationResult.Wrap("canonical inference receipt limit is negative")
	}
	converted, err := strconv.ParseUint(strconv.Itoa(value), 10, 64)
	if err != nil {
		return 0, ErrInvalidVerificationResult.Wrap("canonical inference receipt limit exceeds uint64")
	}
	return converted, nil
}

func canonicalUint64ToInt(value uint64) (int, error) {
	converted, err := strconv.Atoi(strconv.FormatUint(value, 10))
	if err != nil {
		return 0, ErrInvalidVerificationResult.Wrap("canonical inference receipt length exceeds platform int")
	}
	return converted, nil
}

func (d *inferenceReceiptBinaryDecoder) readInferenceDeterminismProfile() (InferenceDeterminismProfile, error) {
	forceCPU, err := d.readBool()
	if err != nil {
		return InferenceDeterminismProfile{}, err
	}
	randomSeed, err := d.readInt64()
	if err != nil {
		return InferenceDeterminismProfile{}, err
	}
	deterministicOps, err := d.readBool()
	if err != nil {
		return InferenceDeterminismProfile{}, err
	}
	interOpThreads, err := d.readInt32()
	if err != nil {
		return InferenceDeterminismProfile{}, err
	}
	intraOpThreads, err := d.readInt32()
	if err != nil {
		return InferenceDeterminismProfile{}, err
	}
	disableGPU, err := d.readBool()
	if err != nil {
		return InferenceDeterminismProfile{}, err
	}
	return InferenceDeterminismProfile{
		ForceCPU:         forceCPU,
		RandomSeed:       randomSeed,
		DeterministicOps: deterministicOps,
		InterOpThreads:   interOpThreads,
		IntraOpThreads:   intraOpThreads,
		DisableGPU:       disableGPU,
	}, nil
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
		if !utf8.ValidString(value) {
			return ErrInvalidVerificationResult.Wrapf("inference receipt %s is not valid UTF-8", name)
		}
	}
	if _, err := hex.DecodeString(r.SignerFingerprint); err != nil || len(r.SignerFingerprint) != sha256.Size*2 {
		return ErrInvalidSignerKey.Wrap("signer fingerprint must be a SHA-256 hex digest")
	}
	if r.SignerFingerprint != strings.ToLower(r.SignerFingerprint) {
		return ErrInvalidSignerKey.Wrap("signer fingerprint must be lowercase canonical hex")
	}
	if len(r.ScopeIDs) == 0 || len(r.ScopeIDs) > InferenceReceiptMaxScopes {
		return ErrInvalidVerificationResult.Wrap("invalid inference receipt scope count")
	}
	for i, scopeID := range r.ScopeIDs {
		if scopeID == "" || len(scopeID) > InferenceReceiptMaxString {
			return ErrInvalidVerificationResult.Wrap("invalid inference receipt scope id")
		}
		if !utf8.ValidString(scopeID) {
			return ErrInvalidVerificationResult.Wrap("inference receipt scope id is not valid UTF-8")
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
		if !utf8.ValidString(string(code)) {
			return ErrInvalidVerificationResult.Wrap("inference receipt reason code is not valid UTF-8")
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
	if r.IssuedAt.IsZero() || r.ExpiresAt.IsZero() || !r.ExpiresAt.After(r.IssuedAt) {
		return ErrInvalidTimestamp.Wrap("invalid inference receipt time bounds")
	}
	if r.IssuedAt.Nanosecond() != 0 || r.ExpiresAt.Nanosecond() != 0 {
		return ErrInvalidTimestamp.Wrap("inference receipt timestamps must use whole UTC seconds")
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
