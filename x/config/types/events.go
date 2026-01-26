package types

// Event types for the config module
const (
	EventTypeClientRegistered  = "client_registered"
	EventTypeClientUpdated     = "client_updated"
	EventTypeClientSuspended   = "client_suspended"
	EventTypeClientRevoked     = "client_revoked"
	EventTypeClientReactivated = "client_reactivated"
	EventTypeSignatureVerified = "signature_verified"

	AttributeKeyClientID     = "client_id"
	AttributeKeyClientName   = "client_name"
	AttributeKeyStatus       = "status"
	AttributeKeyUpdatedBy    = "updated_by"
	AttributeKeyReason       = "reason"
	AttributeKeyMinVersion   = "min_version"
	AttributeKeyMaxVersion   = "max_version"
	AttributeKeyAllowedScope = "allowed_scope"
	AttributeKeyKeyType      = "key_type"
	AttributeKeyTimestamp    = "timestamp"
)

// EventClientRegistered is emitted when a new client is registered
type EventClientRegistered struct {
	ClientID     string `json:"client_id"`
	Name         string `json:"name"`
	KeyType      string `json:"key_type"`
	MinVersion   string `json:"min_version"`
	MaxVersion   string `json:"max_version,omitempty"`
	RegisteredBy string `json:"registered_by"`
	RegisteredAt int64  `json:"registered_at"`
}

// EventClientUpdated is emitted when a client is updated
type EventClientUpdated struct {
	ClientID     string            `json:"client_id"`
	UpdatedBy    string            `json:"updated_by"`
	UpdatedAt    int64             `json:"updated_at"`
	Changes      map[string]string `json:"changes,omitempty"`
}

// EventClientSuspended is emitted when a client is suspended
type EventClientSuspended struct {
	ClientID    string `json:"client_id"`
	SuspendedBy string `json:"suspended_by"`
	SuspendedAt int64  `json:"suspended_at"`
	Reason      string `json:"reason"`
}

// EventClientRevoked is emitted when a client is revoked
type EventClientRevoked struct {
	ClientID  string `json:"client_id"`
	RevokedBy string `json:"revoked_by"`
	RevokedAt int64  `json:"revoked_at"`
	Reason    string `json:"reason"`
}

// EventClientReactivated is emitted when a client is reactivated
type EventClientReactivated struct {
	ClientID      string `json:"client_id"`
	ReactivatedBy string `json:"reactivated_by"`
	ReactivatedAt int64  `json:"reactivated_at"`
	Reason        string `json:"reason"`
}

// EventSignatureVerified is emitted when a signature is successfully verified
type EventSignatureVerified struct {
	ClientID       string `json:"client_id"`
	ClientVersion  string `json:"client_version"`
	AccountAddress string `json:"account_address"`
	PayloadHashHex string `json:"payload_hash_hex"`
	VerifiedAt     int64  `json:"verified_at"`
}
