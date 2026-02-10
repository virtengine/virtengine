package types

// Event types for the encryption module
const (
	EventTypeKeyRegistered = "key_registered"
	EventTypeKeyRevoked    = "key_revoked"
	EventTypeKeyUpdated    = "key_updated"
	EventTypeKeyRotated    = "key_rotated"
	EventTypeKeyExpired    = "key_expired"
	EventTypeKeyExpiryWarn = "key_expiry_warning"
)

// Event attribute keys
const (
	AttributeKeyAddress      = "address"
	AttributeKeyFingerprint  = "fingerprint"
	AttributeKeyAlgorithm    = "algorithm"
	AttributeKeyLabel        = "label"
	AttributeKeyRegisteredAt = "registered_at"
	AttributeKeyRevokedAt    = "revoked_at"
	AttributeKeyRevokedBy    = "revoked_by"
	AttributeKeyDeprecatedAt = "deprecated_at"
	AttributeKeyExpiresAt    = "expires_at"
	AttributeKeyOldKey       = "old_fingerprint"
	AttributeKeyNewKey       = "new_fingerprint"
	AttributeKeyWarningSecs  = "warning_window_seconds"
)

// EventKeyRegistered is emitted when a recipient key is registered
type EventKeyRegistered struct {
	Address      string `json:"address"`
	Fingerprint  string `json:"fingerprint"`
	Algorithm    string `json:"algorithm"`
	Label        string `json:"label,omitempty"`
	RegisteredAt int64  `json:"registered_at"`
}

// EventKeyRevoked is emitted when a recipient key is revoked
type EventKeyRevoked struct {
	Address     string `json:"address"`
	Fingerprint string `json:"fingerprint"`
	RevokedBy   string `json:"revoked_by"`
	RevokedAt   int64  `json:"revoked_at"`
}

// EventKeyUpdated is emitted when a recipient key is updated (e.g., label changed)
type EventKeyUpdated struct {
	Address     string `json:"address"`
	Fingerprint string `json:"fingerprint"`
	Field       string `json:"field"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
}

// EventKeyRotated is emitted when a recipient key is rotated
type EventKeyRotated struct {
	Address        string `json:"address"`
	OldFingerprint string `json:"old_fingerprint"`
	NewFingerprint string `json:"new_fingerprint"`
	RotatedAt      int64  `json:"rotated_at"`
}

// EventKeyExpiryWarning is emitted when a key is nearing expiry
type EventKeyExpiryWarning struct {
	Address              string `json:"address"`
	Fingerprint          string `json:"fingerprint"`
	ExpiresAt            int64  `json:"expires_at"`
	WarningWindowSeconds uint64 `json:"warning_window_seconds"`
}

// EventKeyExpired is emitted when a key expires
type EventKeyExpired struct {
	Address     string `json:"address"`
	Fingerprint string `json:"fingerprint"`
	ExpiredAt   int64  `json:"expired_at"`
}
