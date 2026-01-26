package types

// Event types for the encryption module
const (
	EventTypeKeyRegistered = "key_registered"
	EventTypeKeyRevoked    = "key_revoked"
	EventTypeKeyUpdated    = "key_updated"
)

// Event attribute keys
const (
	AttributeKeyAddress        = "address"
	AttributeKeyFingerprint    = "fingerprint"
	AttributeKeyAlgorithm      = "algorithm"
	AttributeKeyLabel          = "label"
	AttributeKeyRegisteredAt   = "registered_at"
	AttributeKeyRevokedAt      = "revoked_at"
	AttributeKeyRevokedBy      = "revoked_by"
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
