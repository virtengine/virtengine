package types

import (
	"crypto/sha256"
	"math"
	"time"

	encryptiontypes "pkg.akt.dev/node/x/encryption/types"
)

// ============================================================================
// Embedding Envelope Types
// ============================================================================

// EmbeddingEnvelopeVersion is the current version of the embedding envelope format
const EmbeddingEnvelopeVersion uint32 = 1

// EmbeddingType represents the type of embedding stored
type EmbeddingType string

const (
	// EmbeddingTypeFace represents a face embedding vector
	EmbeddingTypeFace EmbeddingType = "face"

	// EmbeddingTypeDocumentFace represents a face extracted from a document
	EmbeddingTypeDocumentFace EmbeddingType = "document_face"

	// EmbeddingTypeVoice represents a voice print embedding
	EmbeddingTypeVoice EmbeddingType = "voice"

	// EmbeddingTypeFingerprint represents a fingerprint embedding
	EmbeddingTypeFingerprint EmbeddingType = "fingerprint"
)

// AllEmbeddingTypes returns all valid embedding types
func AllEmbeddingTypes() []EmbeddingType {
	return []EmbeddingType{
		EmbeddingTypeFace,
		EmbeddingTypeDocumentFace,
		EmbeddingTypeVoice,
		EmbeddingTypeFingerprint,
	}
}

// IsValidEmbeddingType checks if an embedding type is valid
func IsValidEmbeddingType(t EmbeddingType) bool {
	for _, valid := range AllEmbeddingTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// EmbeddingEnvelope represents an encrypted embedding payload with on-chain hash reference
// SECURITY: Raw embedding vectors are NEVER stored on-chain in plaintext
// Only the hash of the embedding is stored on-chain; the encrypted embedding
// is stored off-chain or sent directly to authorized recipients
type EmbeddingEnvelope struct {
	// EnvelopeID is the unique identifier for this envelope
	EnvelopeID string `json:"envelope_id"`

	// AccountAddress is the account this embedding belongs to
	AccountAddress string `json:"account_address"`

	// EmbeddingType identifies what kind of embedding this is
	EmbeddingType EmbeddingType `json:"embedding_type"`

	// Version is the envelope format version
	Version uint32 `json:"version"`

	// EmbeddingHash is the SHA-256 hash of the raw embedding vector
	// This is the ONLY on-chain reference to the embedding
	// Stored as 32 bytes (256 bits)
	EmbeddingHash []byte `json:"embedding_hash"`

	// EncryptedPayload contains the encrypted embedding vector
	// Encrypted to the intended recipients (validators, user, etc.)
	// This is stored OFF-CHAIN; only the hash is on-chain
	EncryptedPayload *encryptiontypes.EncryptedPayloadEnvelope `json:"encrypted_payload,omitempty"`

	// ModelVersion is the ML model version used to generate this embedding
	ModelVersion string `json:"model_version"`

	// ModelHash is the SHA-256 hash of the model weights for reproducibility
	ModelHash string `json:"model_hash"`

	// Dimension is the embedding vector dimension
	Dimension uint32 `json:"dimension"`

	// SourceScopeID references the scope this embedding was extracted from
	SourceScopeID string `json:"source_scope_id"`

	// CreatedAt is when this envelope was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is the block at which this was created
	BlockHeight int64 `json:"block_height"`

	// ComputedBy is the validator address that computed this embedding
	ComputedBy string `json:"computed_by"`

	// RetentionPolicy defines when this envelope should be deleted
	RetentionPolicy *RetentionPolicy `json:"retention_policy,omitempty"`

	// Revoked indicates if this embedding has been revoked
	Revoked bool `json:"revoked"`

	// RevokedAt is when the embedding was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevokedReason is the reason for revocation
	RevokedReason string `json:"revoked_reason,omitempty"`
}

// NewEmbeddingEnvelope creates a new embedding envelope
func NewEmbeddingEnvelope(
	envelopeID string,
	accountAddress string,
	embeddingType EmbeddingType,
	embeddingHash []byte,
	modelVersion string,
	modelHash string,
	dimension uint32,
	sourceScopeID string,
	createdAt time.Time,
	blockHeight int64,
	computedBy string,
) *EmbeddingEnvelope {
	return &EmbeddingEnvelope{
		EnvelopeID:     envelopeID,
		AccountAddress: accountAddress,
		EmbeddingType:  embeddingType,
		Version:        EmbeddingEnvelopeVersion,
		EmbeddingHash:  embeddingHash,
		ModelVersion:   modelVersion,
		ModelHash:      modelHash,
		Dimension:      dimension,
		SourceScopeID:  sourceScopeID,
		CreatedAt:      createdAt,
		BlockHeight:    blockHeight,
		ComputedBy:     computedBy,
		Revoked:        false,
	}
}

// Validate validates the embedding envelope
func (e *EmbeddingEnvelope) Validate() error {
	if e.EnvelopeID == "" {
		return ErrInvalidPayload.Wrap("envelope_id cannot be empty")
	}

	if e.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if !IsValidEmbeddingType(e.EmbeddingType) {
		return ErrInvalidPayload.Wrapf("invalid embedding_type: %s", e.EmbeddingType)
	}

	if e.Version == 0 || e.Version > EmbeddingEnvelopeVersion {
		return ErrInvalidPayload.Wrapf("unsupported version: %d", e.Version)
	}

	// Embedding hash must be 32 bytes (SHA-256)
	if len(e.EmbeddingHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("embedding_hash must be 32 bytes (SHA-256)")
	}

	if e.ModelVersion == "" {
		return ErrInvalidPayload.Wrap("model_version cannot be empty")
	}

	if e.Dimension == 0 {
		return ErrInvalidPayload.Wrap("dimension cannot be zero")
	}

	if e.SourceScopeID == "" {
		return ErrInvalidPayload.Wrap("source_scope_id cannot be empty")
	}

	if e.CreatedAt.IsZero() {
		return ErrInvalidPayload.Wrap("created_at cannot be zero")
	}

	if e.BlockHeight < 0 {
		return ErrInvalidPayload.Wrap("block_height cannot be negative")
	}

	// Validate retention policy if present
	if e.RetentionPolicy != nil {
		if err := e.RetentionPolicy.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// SetEncryptedPayload sets the encrypted payload (stored off-chain)
func (e *EmbeddingEnvelope) SetEncryptedPayload(payload *encryptiontypes.EncryptedPayloadEnvelope) {
	e.EncryptedPayload = payload
}

// SetRetentionPolicy sets the retention policy for this envelope
func (e *EmbeddingEnvelope) SetRetentionPolicy(policy *RetentionPolicy) {
	e.RetentionPolicy = policy
}

// Revoke revokes this embedding envelope
func (e *EmbeddingEnvelope) Revoke(reason string, revokedAt time.Time) {
	e.Revoked = true
	e.RevokedAt = &revokedAt
	e.RevokedReason = reason
}

// IsActive checks if the envelope is active (not revoked and not expired)
func (e *EmbeddingEnvelope) IsActive(now time.Time) bool {
	if e.Revoked {
		return false
	}

	if e.RetentionPolicy != nil && e.RetentionPolicy.IsExpired(now) {
		return false
	}

	return true
}

// MatchesEmbedding checks if a raw embedding matches this envelope's hash
func (e *EmbeddingEnvelope) MatchesEmbedding(embedding []byte) bool {
	if len(e.EmbeddingHash) == 0 {
		return false
	}
	hash := sha256.Sum256(embedding)
	return len(hash) == len(e.EmbeddingHash) && compareHashBytes(hash[:], e.EmbeddingHash)
}

// compareHashBytes performs a constant-time comparison of two byte slices
func compareHashBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// ToOnChainReference returns the on-chain portion of the envelope (no encrypted data)
// This is what gets stored on the blockchain
func (e *EmbeddingEnvelope) ToOnChainReference() EmbeddingEnvelopeReference {
	return EmbeddingEnvelopeReference{
		EnvelopeID:      e.EnvelopeID,
		AccountAddress:  e.AccountAddress,
		EmbeddingType:   e.EmbeddingType,
		Version:         e.Version,
		EmbeddingHash:   e.EmbeddingHash,
		ModelVersion:    e.ModelVersion,
		ModelHash:       e.ModelHash,
		Dimension:       e.Dimension,
		SourceScopeID:   e.SourceScopeID,
		CreatedAt:       e.CreatedAt,
		BlockHeight:     e.BlockHeight,
		ComputedBy:      e.ComputedBy,
		RetentionPolicy: e.RetentionPolicy,
		Revoked:         e.Revoked,
		RevokedAt:       e.RevokedAt,
		RevokedReason:   e.RevokedReason,
	}
}

// EmbeddingEnvelopeReference is the on-chain portion of an embedding envelope
// Contains only the hash and metadata - NO encrypted payload or raw embedding
type EmbeddingEnvelopeReference struct {
	// EnvelopeID is the unique identifier for this envelope
	EnvelopeID string `json:"envelope_id"`

	// AccountAddress is the account this embedding belongs to
	AccountAddress string `json:"account_address"`

	// EmbeddingType identifies what kind of embedding this is
	EmbeddingType EmbeddingType `json:"embedding_type"`

	// Version is the envelope format version
	Version uint32 `json:"version"`

	// EmbeddingHash is the SHA-256 hash of the raw embedding vector
	EmbeddingHash []byte `json:"embedding_hash"`

	// ModelVersion is the ML model version used
	ModelVersion string `json:"model_version"`

	// ModelHash is the model weights hash
	ModelHash string `json:"model_hash"`

	// Dimension is the embedding vector dimension
	Dimension uint32 `json:"dimension"`

	// SourceScopeID references the source scope
	SourceScopeID string `json:"source_scope_id"`

	// CreatedAt is when this was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is the creation block
	BlockHeight int64 `json:"block_height"`

	// ComputedBy is the validator that computed this
	ComputedBy string `json:"computed_by"`

	// RetentionPolicy defines retention rules
	RetentionPolicy *RetentionPolicy `json:"retention_policy,omitempty"`

	// Revoked indicates if revoked
	Revoked bool `json:"revoked"`

	// RevokedAt is when revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevokedReason is the revocation reason
	RevokedReason string `json:"revoked_reason,omitempty"`
}

// Validate validates the embedding envelope reference
func (r *EmbeddingEnvelopeReference) Validate() error {
	if r.EnvelopeID == "" {
		return ErrInvalidPayload.Wrap("envelope_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if len(r.EmbeddingHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("embedding_hash must be 32 bytes")
	}

	return nil
}

// ComputeEmbeddingHash computes the SHA-256 hash of a raw embedding vector
func ComputeEmbeddingHash(embedding []byte) []byte {
	hash := sha256.Sum256(embedding)
	return hash[:]
}

// ComputeEmbeddingHashFromFloat32 computes hash from float32 slice
// Uses IEEE 754 binary representation for deterministic cross-validator consistency
func ComputeEmbeddingHashFromFloat32(embedding []float32) []byte {
	// Convert float32 to bytes - use consistent big-endian byte order
	bytes := make([]byte, len(embedding)*4)
	for i, v := range embedding {
		bits := math.Float32bits(v)
		bytes[i*4] = byte(bits >> 24)
		bytes[i*4+1] = byte(bits >> 16)
		bytes[i*4+2] = byte(bits >> 8)
		bytes[i*4+3] = byte(bits)
	}
	return ComputeEmbeddingHash(bytes)
}
