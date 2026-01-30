package v1

import (
	"time"
)

// EnclaveIdentity represents a validator's enclave identity record
type EnclaveIdentity struct {
	// ValidatorAddress is the validator operator address
	ValidatorAddress string `json:"validator_address" yaml:"validator_address"`

	// TEEType is the type of TEE (SGX, SEV-SNP, NITRO)
	TEEType TEEType `json:"tee_type" yaml:"tee_type"`

	// MeasurementHash is the enclave measurement (MRENCLAVE for SGX)
	MeasurementHash []byte `json:"measurement_hash" yaml:"measurement_hash"`

	// SignerHash is the signer measurement (MRSIGNER for SGX)
	SignerHash []byte `json:"signer_hash,omitempty" yaml:"signer_hash"`

	// EncryptionPubKey is the enclave's public key for encryption
	EncryptionPubKey []byte `json:"encryption_pub_key" yaml:"encryption_pub_key"`

	// SigningPubKey is the enclave's public key for signing attestations
	SigningPubKey []byte `json:"signing_pub_key" yaml:"signing_pub_key"`

	// AttestationQuote is the raw attestation quote from the TEE
	AttestationQuote []byte `json:"attestation_quote" yaml:"attestation_quote"`

	// AttestationChain is the certificate chain for attestation verification
	AttestationChain [][]byte `json:"attestation_chain,omitempty" yaml:"attestation_chain"`

	// ISVProdID is the Independent Software Vendor Product ID
	ISVProdID uint16 `json:"isv_prod_id" yaml:"isv_prod_id"`

	// ISVSVN is the Independent Software Vendor Security Version Number
	ISVSVN uint16 `json:"isv_svn" yaml:"isv_svn"`

	// QuoteVersion is the attestation quote format version
	QuoteVersion uint32 `json:"quote_version" yaml:"quote_version"`

	// DebugMode indicates if the enclave is in debug mode
	DebugMode bool `json:"debug_mode" yaml:"debug_mode"`

	// Epoch is the registration epoch
	Epoch uint64 `json:"epoch" yaml:"epoch"`

	// ExpiryHeight is the block height when this identity expires
	ExpiryHeight int64 `json:"expiry_height" yaml:"expiry_height"`

	// RegisteredAt is the timestamp when this identity was registered
	RegisteredAt time.Time `json:"registered_at" yaml:"registered_at"`

	// UpdatedAt is the timestamp when this identity was last updated
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`

	// Status is the current status of the enclave identity
	Status string `json:"status" yaml:"status"`
}

// MeasurementRecord represents an approved enclave measurement in the allowlist
type MeasurementRecord struct {
	// MeasurementHash is the enclave measurement hash
	MeasurementHash []byte `json:"measurement_hash" yaml:"measurement_hash"`

	// TEEType is the TEE type this measurement is for
	TEEType TEEType `json:"tee_type" yaml:"tee_type"`

	// Description is a human-readable description
	Description string `json:"description" yaml:"description"`

	// MinISVSVN is the minimum required security version
	MinISVSVN uint16 `json:"min_isv_svn" yaml:"min_isv_svn"`

	// AddedAt is when this measurement was added
	AddedAt time.Time `json:"added_at" yaml:"added_at"`

	// AddedByProposal is the governance proposal ID that added this measurement
	AddedByProposal uint64 `json:"added_by_proposal" yaml:"added_by_proposal"`

	// ExpiryHeight is when this measurement expires (0 for no expiry)
	ExpiryHeight int64 `json:"expiry_height,omitempty" yaml:"expiry_height"`

	// Revoked indicates if this measurement has been revoked
	Revoked bool `json:"revoked" yaml:"revoked"`

	// RevokedAt is when this measurement was revoked (if applicable)
	RevokedAt *time.Time `json:"revoked_at,omitempty" yaml:"revoked_at"`

	// RevokedByProposal is the governance proposal that revoked this (if applicable)
	RevokedByProposal uint64 `json:"revoked_by_proposal,omitempty" yaml:"revoked_by_proposal"`
}

// KeyRotationRecord represents a key rotation event
type KeyRotationRecord struct {
	// ValidatorAddress is the validator operator address
	ValidatorAddress string `json:"validator_address" yaml:"validator_address"`

	// Epoch is the epoch when rotation was initiated
	Epoch uint64 `json:"epoch" yaml:"epoch"`

	// OldKeyFingerprint is the fingerprint of the old key
	OldKeyFingerprint string `json:"old_key_fingerprint" yaml:"old_key_fingerprint"`

	// NewKeyFingerprint is the fingerprint of the new key
	NewKeyFingerprint string `json:"new_key_fingerprint" yaml:"new_key_fingerprint"`

	// OverlapStartHeight is when both keys become valid
	OverlapStartHeight int64 `json:"overlap_start_height" yaml:"overlap_start_height"`

	// OverlapEndHeight is when the old key becomes invalid
	OverlapEndHeight int64 `json:"overlap_end_height" yaml:"overlap_end_height"`

	// InitiatedAt is when the rotation was initiated
	InitiatedAt time.Time `json:"initiated_at" yaml:"initiated_at"`

	// CompletedAt is when the rotation was completed (old key invalidated)
	CompletedAt *time.Time `json:"completed_at,omitempty" yaml:"completed_at"`

	// Status is the current status of the rotation
	Status string `json:"status" yaml:"status"`
}

// ValidatorKeyInfo contains key information for a validator
type ValidatorKeyInfo struct {
	ValidatorAddress string `json:"validator_address" yaml:"validator_address"`
	EncryptionKeyID  string `json:"encryption_key_id" yaml:"encryption_key_id"`
	EncryptionPubKey []byte `json:"encryption_pub_key" yaml:"encryption_pub_key"`
	MeasurementHash  []byte `json:"measurement_hash" yaml:"measurement_hash"`
	ExpiryHeight     int64  `json:"expiry_height" yaml:"expiry_height"`
	IsInRotation     bool   `json:"is_in_rotation" yaml:"is_in_rotation"`
}

// Params contains the module parameters
type Params struct {
	// AllowedTEETypes specifies which TEE types are allowed
	AllowedTEETypes []string `json:"allowed_tee_types" yaml:"allowed_tee_types"`

	// DefaultExpiryBlocks is the default expiry period in blocks
	DefaultExpiryBlocks int64 `json:"default_expiry_blocks" yaml:"default_expiry_blocks"`

	// MinOverlapBlocks is the minimum overlap period for key rotation
	MinOverlapBlocks int64 `json:"min_overlap_blocks" yaml:"min_overlap_blocks"`

	// MaxOverlapBlocks is the maximum overlap period for key rotation
	MaxOverlapBlocks int64 `json:"max_overlap_blocks" yaml:"max_overlap_blocks"`

	// MinQuoteVersion is the minimum attestation quote version
	MinQuoteVersion uint32 `json:"min_quote_version" yaml:"min_quote_version"`

	// RequireAttestationChain indicates if attestation chain is required
	RequireAttestationChain bool `json:"require_attestation_chain" yaml:"require_attestation_chain"`
}

// AttestedScoringResult represents an attested VEID scoring result
type AttestedScoringResult struct {
	// ScopeID is the identity scope being scored
	ScopeID string `json:"scope_id" yaml:"scope_id"`

	// AccountAddress is the account that owns the scope
	AccountAddress string `json:"account_address" yaml:"account_address"`

	// Score is the computed identity score
	Score uint32 `json:"score" yaml:"score"`

	// Status is the scoring status
	Status string `json:"status" yaml:"status"`

	// BlockHeight is when the scoring was performed
	BlockHeight int64 `json:"block_height" yaml:"block_height"`

	// ValidatorAddress is the validator that performed the scoring
	ValidatorAddress string `json:"validator_address" yaml:"validator_address"`

	// EnclaveMeasurementHash is the measurement of the enclave that scored
	EnclaveMeasurementHash []byte `json:"enclave_measurement_hash" yaml:"enclave_measurement_hash"`

	// EnclaveSignature is the enclave's signature over the result
	EnclaveSignature []byte `json:"enclave_signature" yaml:"enclave_signature"`
}
