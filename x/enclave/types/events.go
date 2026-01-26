package types

// Event types for enclave module
const (
	EventTypeEnclaveIdentityRegistered = "enclaVIRTENGINE_identity_registered"
	EventTypeEnclaveIdentityUpdated    = "enclaVIRTENGINE_identity_updated"
	EventTypeEnclaveIdentityRevoked    = "enclaVIRTENGINE_identity_revoked"
	EventTypeEnclaveKeyRotated         = "enclaVIRTENGINE_key_rotated"
	EventTypeKeyRotationCompleted      = "key_rotation_completed"
	EventTypeMeasurementAdded          = "measurement_added"
	EventTypeMeasurementRevoked        = "measurement_revoked"
	EventTypeVEIDScoreComputedAttested = "veid_score_computed_attested"
	EventTypeVEIDScoreRejectedAttestation = "veid_score_rejected_attestation"
	EventTypeConsensusVerificationFailed  = "consensus_verification_failed"
)

// Event attribute keys
const (
	AttributeKeyValidator           = "validator"
	AttributeKeyTEEType             = "tee_type"
	AttributeKeyMeasurementHash     = "measurement_hash"
	AttributeKeyEncryptionKeyID     = "encryption_key_id"
	AttributeKeySigningKeyID        = "signing_key_id"
	AttributeKeyEpoch               = "epoch"
	AttributeKeyExpiryHeight        = "expiry_height"
	AttributeKeyOldKeyFingerprint   = "old_key_fingerprint"
	AttributeKeyNewKeyFingerprint   = "new_key_fingerprint"
	AttributeKeyOverlapStartHeight  = "overlap_start_height"
	AttributeKeyOverlapEndHeight    = "overlap_end_height"
	AttributeKeyDescription         = "description"
	AttributeKeyMinISVSVN           = "min_isv_svn"
	AttributeKeyReason              = "reason"
	AttributeKeyScopeID             = "scope_id"
	AttributeKeyAccountAddress      = "account_address"
	AttributeKeyScore               = "score"
	AttributeKeyStatus              = "status"
	AttributeKeyBlockHeight         = "block_height"
	AttributeKeyProposedScore       = "proposed_score"
	AttributeKeyComputedScore       = "computed_score"
	AttributeKeyScoreDifference     = "score_difference"
)
