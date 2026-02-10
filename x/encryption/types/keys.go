package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "encryption"

	// StoreKey is the store key string for encryption module
	StoreKey = ModuleName

	// RouterKey is the message route for encryption module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for encryption module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixRecipientKey is the prefix for recipient public key storage
	// Key: PrefixRecipientKey | address | fingerprint -> RecipientKeyRecord
	PrefixRecipientKey = []byte{0x01}

	// PrefixKeyByFingerprint is the prefix for key lookup by fingerprint
	// Key: PrefixKeyByFingerprint | fingerprint -> address
	PrefixKeyByFingerprint = []byte{0x02}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x03}

	// PrefixEnvelopeLog is the prefix for envelope validation log (optional audit)
	// Key: PrefixEnvelopeLog | hash -> EnvelopeLogEntry
	PrefixEnvelopeLog = []byte{0x04}

	// PrefixActiveKey is the prefix for active key lookup per address
	// Key: PrefixActiveKey | address -> fingerprint
	PrefixActiveKey = []byte{0x05}

	// PrefixRecipientKeyVersion is the prefix for key lookup by version per address
	// Key: PrefixRecipientKeyVersion | address | version -> fingerprint
	PrefixRecipientKeyVersion = []byte{0x06}

	// PrefixEnvelopeRecord is the prefix for envelope storage
	// Key: PrefixEnvelopeRecord | envelope_hash -> EnvelopeRecord
	PrefixEnvelopeRecord = []byte{0x07}

	// PrefixReencryptionJob is the prefix for reencryption job queue entries
	// Key: PrefixReencryptionJob | job_id -> ReencryptionJob
	PrefixReencryptionJob = []byte{0x08}

	// PrefixKeyRotationRecord is the prefix for key rotation records
	// Key: PrefixKeyRotationRecord | rotation_id -> KeyRotationRecord
	PrefixKeyRotationRecord = []byte{0x09}

	// PrefixEphemeralKey is the prefix for ephemeral session keys
	// Key: PrefixEphemeralKey | session_id -> EphemeralKeyRecord
	PrefixEphemeralKey = []byte{0x0A}

	// PrefixExpiryWarning is the prefix for emitted expiry warnings
	// Key: PrefixExpiryWarning | fingerprint | warning_window_seconds -> 1
	PrefixExpiryWarning = []byte{0x0B}
)

// RecipientKeyKey returns the store key for a recipient's public key
func RecipientKeyKey(address []byte, fingerprint []byte) []byte {
	key := make([]byte, 0, len(PrefixRecipientKey)+len(address)+len(fingerprint))
	key = append(key, PrefixRecipientKey...)
	key = append(key, address...)
	key = append(key, fingerprint...)
	return key
}

// RecipientKeyPrefix returns the prefix for all keys owned by an address.
func RecipientKeyPrefix(address []byte) []byte {
	key := make([]byte, 0, len(PrefixRecipientKey)+len(address))
	key = append(key, PrefixRecipientKey...)
	key = append(key, address...)
	return key
}

// ActiveKeyKey returns the store key for an address's active key fingerprint.
func ActiveKeyKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixActiveKey)+len(address))
	key = append(key, PrefixActiveKey...)
	key = append(key, address...)
	return key
}

// KeyByFingerprintKey returns the store key for looking up address by key fingerprint
func KeyByFingerprintKey(fingerprint []byte) []byte {
	key := make([]byte, 0, len(PrefixKeyByFingerprint)+len(fingerprint))
	key = append(key, PrefixKeyByFingerprint...)
	key = append(key, fingerprint...)
	return key
}

// RecipientKeyVersionKey returns the store key for a recipient's key version mapping.
func RecipientKeyVersionKey(address []byte, version uint32) []byte {
	key := make([]byte, 0, len(PrefixRecipientKeyVersion)+len(address)+4)
	key = append(key, PrefixRecipientKeyVersion...)
	key = append(key, address...)
	key = append(key, byte(version>>24), byte(version>>16), byte(version>>8), byte(version))
	return key
}

// EnvelopeRecordKey returns the store key for an envelope record.
func EnvelopeRecordKey(envelopeHash []byte) []byte {
	key := make([]byte, 0, len(PrefixEnvelopeRecord)+len(envelopeHash))
	key = append(key, PrefixEnvelopeRecord...)
	key = append(key, envelopeHash...)
	return key
}

// ReencryptionJobKey returns the store key for a reencryption job.
func ReencryptionJobKey(jobID []byte) []byte {
	key := make([]byte, 0, len(PrefixReencryptionJob)+len(jobID))
	key = append(key, PrefixReencryptionJob...)
	key = append(key, jobID...)
	return key
}

// KeyRotationRecordKey returns the store key for a key rotation record.
func KeyRotationRecordKey(rotationID []byte) []byte {
	key := make([]byte, 0, len(PrefixKeyRotationRecord)+len(rotationID))
	key = append(key, PrefixKeyRotationRecord...)
	key = append(key, rotationID...)
	return key
}

// EphemeralKeyRecordKey returns the store key for an ephemeral key record.
func EphemeralKeyRecordKey(sessionID []byte) []byte {
	key := make([]byte, 0, len(PrefixEphemeralKey)+len(sessionID))
	key = append(key, PrefixEphemeralKey...)
	key = append(key, sessionID...)
	return key
}

// ExpiryWarningKey returns the store key for an expiry warning marker.
func ExpiryWarningKey(fingerprint []byte, warningWindowSeconds uint64) []byte {
	key := make([]byte, 0, len(PrefixExpiryWarning)+len(fingerprint)+8)
	key = append(key, PrefixExpiryWarning...)
	key = append(key, fingerprint...)
	key = append(key,
		byte(warningWindowSeconds>>56),
		byte(warningWindowSeconds>>48),
		byte(warningWindowSeconds>>40),
		byte(warningWindowSeconds>>32),
		byte(warningWindowSeconds>>24),
		byte(warningWindowSeconds>>16),
		byte(warningWindowSeconds>>8),
		byte(warningWindowSeconds),
	)
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// EnvelopeLogKey returns the store key for an envelope log entry
func EnvelopeLogKey(envelopeHash []byte) []byte {
	key := make([]byte, 0, len(PrefixEnvelopeLog)+len(envelopeHash))
	key = append(key, PrefixEnvelopeLog...)
	key = append(key, envelopeHash...)
	return key
}
