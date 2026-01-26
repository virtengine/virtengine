package types

import (
	"bytes"
	"crypto/sha256"
	"time"
)

// ============================================================================
// Derived Features Types
// ============================================================================

// DerivedFeatures contains hashes of derived features from identity scopes
// These are used for verification matching without revealing the underlying data
type DerivedFeatures struct {
	// FaceEmbeddingHash is the SHA-256 hash of the face embedding vector
	// The actual embedding is stored encrypted; this hash allows matching
	FaceEmbeddingHash []byte `json:"face_embedding_hash,omitempty"`

	// DocFieldHashes contains hashes of extracted document fields
	// Keys are field names: "name_hash", "dob_hash", "doc_number_hash", etc.
	DocFieldHashes map[string][]byte `json:"doc_field_hashes,omitempty"`

	// BiometricHash is the hash of biometric data (fingerprint, voice, etc.)
	BiometricHash []byte `json:"biometric_hash,omitempty"`

	// LivenessProofHash is the hash of liveness detection proof
	LivenessProofHash []byte `json:"liveness_proof_hash,omitempty"`

	// LastComputedAt is when these features were last computed
	LastComputedAt time.Time `json:"last_computed_at"`

	// ModelVersion is the ML model version used to compute these features
	ModelVersion string `json:"model_version"`

	// ComputedBy is the validator address that computed these features
	ComputedBy string `json:"computed_by,omitempty"`

	// BlockHeight is the block height when features were computed
	BlockHeight int64 `json:"block_height"`

	// FeatureVersion tracks the derived features schema version
	FeatureVersion uint32 `json:"feature_version"`
}

// CurrentDerivedFeaturesVersion is the current schema version for derived features
const CurrentDerivedFeaturesVersion uint32 = 1

// Well-known document field hash keys
const (
	// DocFieldNameHash is the hash of the full name from document
	DocFieldNameHash = "name_hash"

	// DocFieldDOBHash is the hash of the date of birth
	DocFieldDOBHash = "dob_hash"

	// DocFieldDocNumberHash is the hash of the document number
	DocFieldDocNumberHash = "doc_number_hash"

	// DocFieldNationalityHash is the hash of the nationality
	DocFieldNationalityHash = "nationality_hash"

	// DocFieldGenderHash is the hash of the gender
	DocFieldGenderHash = "gender_hash"

	// DocFieldExpiryHash is the hash of the document expiry date
	DocFieldExpiryHash = "expiry_hash"

	// DocFieldAddressHash is the hash of the address
	DocFieldAddressHash = "address_hash"

	// DocFieldEmailHash is the hash of the email
	DocFieldEmailHash = "email_hash"

	// DocFieldPhoneHash is the hash of the phone number
	DocFieldPhoneHash = "phone_hash"
)

// AllDocFieldKeys returns all well-known document field hash keys
func AllDocFieldKeys() []string {
	return []string{
		DocFieldNameHash,
		DocFieldDOBHash,
		DocFieldDocNumberHash,
		DocFieldNationalityHash,
		DocFieldGenderHash,
		DocFieldExpiryHash,
		DocFieldAddressHash,
		DocFieldEmailHash,
		DocFieldPhoneHash,
	}
}

// NewDerivedFeatures creates a new empty derived features struct
func NewDerivedFeatures() DerivedFeatures {
	return DerivedFeatures{
		DocFieldHashes: make(map[string][]byte),
		FeatureVersion: CurrentDerivedFeaturesVersion,
	}
}

// Validate validates the derived features
func (df *DerivedFeatures) Validate() error {
	// Face embedding hash should be SHA-256 (32 bytes) or empty
	if len(df.FaceEmbeddingHash) > 0 && len(df.FaceEmbeddingHash) != 32 {
		return ErrInvalidWallet.Wrap("face_embedding_hash must be 32 bytes (SHA-256)")
	}

	// Biometric hash should be SHA-256 (32 bytes) or empty
	if len(df.BiometricHash) > 0 && len(df.BiometricHash) != 32 {
		return ErrInvalidWallet.Wrap("biometric_hash must be 32 bytes (SHA-256)")
	}

	// Liveness proof hash should be SHA-256 (32 bytes) or empty
	if len(df.LivenessProofHash) > 0 && len(df.LivenessProofHash) != 32 {
		return ErrInvalidWallet.Wrap("liveness_proof_hash must be 32 bytes (SHA-256)")
	}

	// All doc field hashes should be SHA-256 (32 bytes)
	for key, hash := range df.DocFieldHashes {
		if len(hash) != 32 {
			return ErrInvalidWallet.Wrapf("doc_field_hash[%s] must be 32 bytes (SHA-256)", key)
		}
	}

	return nil
}

// IsEmpty checks if no derived features have been computed
func (df *DerivedFeatures) IsEmpty() bool {
	return len(df.FaceEmbeddingHash) == 0 &&
		len(df.DocFieldHashes) == 0 &&
		len(df.BiometricHash) == 0 &&
		len(df.LivenessProofHash) == 0
}

// SetFaceEmbeddingHash sets the face embedding hash from raw embedding data
func (df *DerivedFeatures) SetFaceEmbeddingHash(embedding []byte) {
	hash := sha256.Sum256(embedding)
	df.FaceEmbeddingHash = hash[:]
}

// SetFaceEmbeddingHashDirect sets the face embedding hash directly
func (df *DerivedFeatures) SetFaceEmbeddingHashDirect(hash []byte) {
	df.FaceEmbeddingHash = hash
}

// SetBiometricHash sets the biometric hash from raw biometric data
func (df *DerivedFeatures) SetBiometricHash(biometricData []byte) {
	hash := sha256.Sum256(biometricData)
	df.BiometricHash = hash[:]
}

// SetBiometricHashDirect sets the biometric hash directly
func (df *DerivedFeatures) SetBiometricHashDirect(hash []byte) {
	df.BiometricHash = hash
}

// SetLivenessProofHash sets the liveness proof hash
func (df *DerivedFeatures) SetLivenessProofHash(proofData []byte) {
	hash := sha256.Sum256(proofData)
	df.LivenessProofHash = hash[:]
}

// SetLivenessProofHashDirect sets the liveness proof hash directly
func (df *DerivedFeatures) SetLivenessProofHashDirect(hash []byte) {
	df.LivenessProofHash = hash
}

// SetDocFieldHash sets a document field hash from raw field data
func (df *DerivedFeatures) SetDocFieldHash(fieldKey string, fieldData []byte) {
	hash := sha256.Sum256(fieldData)
	df.DocFieldHashes[fieldKey] = hash[:]
}

// SetDocFieldHashDirect sets a document field hash directly
func (df *DerivedFeatures) SetDocFieldHashDirect(fieldKey string, hash []byte) {
	df.DocFieldHashes[fieldKey] = hash
}

// GetDocFieldHash returns a document field hash by key
func (df *DerivedFeatures) GetDocFieldHash(fieldKey string) ([]byte, bool) {
	hash, found := df.DocFieldHashes[fieldKey]
	return hash, found
}

// HasDocField checks if a document field hash exists
func (df *DerivedFeatures) HasDocField(fieldKey string) bool {
	_, found := df.DocFieldHashes[fieldKey]
	return found
}

// RemoveDocFieldHash removes a document field hash
func (df *DerivedFeatures) RemoveDocFieldHash(fieldKey string) {
	delete(df.DocFieldHashes, fieldKey)
}

// MatchesFaceEmbedding checks if a face embedding matches the stored hash
func (df *DerivedFeatures) MatchesFaceEmbedding(embedding []byte) bool {
	if len(df.FaceEmbeddingHash) == 0 {
		return false
	}
	hash := sha256.Sum256(embedding)
	return bytes.Equal(df.FaceEmbeddingHash, hash[:])
}

// MatchesBiometric checks if biometric data matches the stored hash
func (df *DerivedFeatures) MatchesBiometric(biometricData []byte) bool {
	if len(df.BiometricHash) == 0 {
		return false
	}
	hash := sha256.Sum256(biometricData)
	return bytes.Equal(df.BiometricHash, hash[:])
}

// MatchesDocField checks if a document field value matches the stored hash
func (df *DerivedFeatures) MatchesDocField(fieldKey string, fieldData []byte) bool {
	storedHash, found := df.DocFieldHashes[fieldKey]
	if !found {
		return false
	}
	hash := sha256.Sum256(fieldData)
	return bytes.Equal(storedHash, hash[:])
}

// UpdateMetadata updates the computation metadata
func (df *DerivedFeatures) UpdateMetadata(modelVersion string, computedBy string, blockHeight int64, computedAt time.Time) {
	df.ModelVersion = modelVersion
	df.ComputedBy = computedBy
	df.BlockHeight = blockHeight
	df.LastComputedAt = computedAt
}

// GetFieldCount returns the number of document field hashes
func (df *DerivedFeatures) GetFieldCount() int {
	return len(df.DocFieldHashes)
}

// GetFieldKeys returns all document field hash keys
func (df *DerivedFeatures) GetFieldKeys() []string {
	keys := make([]string, 0, len(df.DocFieldHashes))
	for k := range df.DocFieldHashes {
		keys = append(keys, k)
	}
	return keys
}

// DerivedFeaturesUpdate represents an update to derived features
// Only validators can submit these updates
type DerivedFeaturesUpdate struct {
	// AccountAddress is the address to update features for
	AccountAddress string `json:"account_address"`

	// FaceEmbeddingHash is the new face embedding hash (nil to keep existing)
	FaceEmbeddingHash []byte `json:"face_embedding_hash,omitempty"`

	// DocFieldHashes are new document field hashes (merged with existing)
	DocFieldHashes map[string][]byte `json:"doc_field_hashes,omitempty"`

	// BiometricHash is the new biometric hash (nil to keep existing)
	BiometricHash []byte `json:"biometric_hash,omitempty"`

	// LivenessProofHash is the new liveness proof hash (nil to keep existing)
	LivenessProofHash []byte `json:"liveness_proof_hash,omitempty"`

	// ModelVersion is the ML model version used
	ModelVersion string `json:"model_version"`

	// ValidatorAddress is the validator submitting this update
	ValidatorAddress string `json:"validator_address"`
}

// Apply applies the update to derived features
func (dfu *DerivedFeaturesUpdate) Apply(df *DerivedFeatures, blockHeight int64, computedAt time.Time) {
	if len(dfu.FaceEmbeddingHash) > 0 {
		df.FaceEmbeddingHash = dfu.FaceEmbeddingHash
	}

	if len(dfu.BiometricHash) > 0 {
		df.BiometricHash = dfu.BiometricHash
	}

	if len(dfu.LivenessProofHash) > 0 {
		df.LivenessProofHash = dfu.LivenessProofHash
	}

	// Merge document field hashes
	for key, hash := range dfu.DocFieldHashes {
		df.DocFieldHashes[key] = hash
	}

	df.UpdateMetadata(dfu.ModelVersion, dfu.ValidatorAddress, blockHeight, computedAt)
}

// Validate validates the derived features update
func (dfu *DerivedFeaturesUpdate) Validate() error {
	if dfu.AccountAddress == "" {
		return ErrInvalidAddress.Wrap(errMsgAccountAddrEmpty)
	}

	if dfu.ModelVersion == "" {
		return ErrInvalidWallet.Wrap("model_version cannot be empty")
	}

	if dfu.ValidatorAddress == "" {
		return ErrValidatorOnly.Wrap("validator_address cannot be empty")
	}

	// Validate hash lengths
	if len(dfu.FaceEmbeddingHash) > 0 && len(dfu.FaceEmbeddingHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("face_embedding_hash must be 32 bytes")
	}

	if len(dfu.BiometricHash) > 0 && len(dfu.BiometricHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("biometric_hash must be 32 bytes")
	}

	if len(dfu.LivenessProofHash) > 0 && len(dfu.LivenessProofHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("liveness_proof_hash must be 32 bytes")
	}

	for key, hash := range dfu.DocFieldHashes {
		if len(hash) != 32 {
			return ErrInvalidPayloadHash.Wrapf("doc_field_hash[%s] must be 32 bytes", key)
		}
	}

	return nil
}

// PublicDerivedFeaturesInfo represents non-sensitive information about derived features
// This is returned by query endpoints
type PublicDerivedFeaturesInfo struct {
	// HasFaceEmbedding indicates if face embedding hash exists
	HasFaceEmbedding bool `json:"has_face_embedding"`

	// HasBiometric indicates if biometric hash exists
	HasBiometric bool `json:"has_biometric"`

	// HasLivenessProof indicates if liveness proof hash exists
	HasLivenessProof bool `json:"has_liveness_proof"`

	// DocFieldKeys lists which document fields have hashes
	DocFieldKeys []string `json:"doc_field_keys"`

	// LastComputedAt is when features were last computed
	LastComputedAt time.Time `json:"last_computed_at"`

	// ModelVersion is the model version used
	ModelVersion string `json:"model_version"`

	// FeatureVersion is the schema version
	FeatureVersion uint32 `json:"feature_version"`
}

// ToPublicInfo converts derived features to public info
func (df *DerivedFeatures) ToPublicInfo() PublicDerivedFeaturesInfo {
	return PublicDerivedFeaturesInfo{
		HasFaceEmbedding: len(df.FaceEmbeddingHash) > 0,
		HasBiometric:     len(df.BiometricHash) > 0,
		HasLivenessProof: len(df.LivenessProofHash) > 0,
		DocFieldKeys:     df.GetFieldKeys(),
		LastComputedAt:   df.LastComputedAt,
		ModelVersion:     df.ModelVersion,
		FeatureVersion:   df.FeatureVersion,
	}
}
