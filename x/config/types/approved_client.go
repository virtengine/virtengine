package types

import (
	"regexp"
	"strings"
	"time"
)

// ClientStatus represents the status of an approved client
type ClientStatus string

const (
	// ClientStatusActive indicates the client is active and can be used
	ClientStatusActive ClientStatus = "active"

	// ClientStatusSuspended indicates the client is temporarily suspended
	ClientStatusSuspended ClientStatus = "suspended"

	// ClientStatusRevoked indicates the client has been permanently revoked
	ClientStatusRevoked ClientStatus = "revoked"
)

// IsValid checks if the status is valid
func (s ClientStatus) IsValid() bool {
	switch s {
	case ClientStatusActive, ClientStatusSuspended, ClientStatusRevoked:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if a status transition is valid
func (s ClientStatus) CanTransitionTo(target ClientStatus) bool {
	switch s {
	case ClientStatusActive:
		return target == ClientStatusSuspended || target == ClientStatusRevoked
	case ClientStatusSuspended:
		return target == ClientStatusActive || target == ClientStatusRevoked
	case ClientStatusRevoked:
		return false // Revoked is permanent
	default:
		return false
	}
}

// KeyType represents the type of cryptographic key
type KeyType string

const (
	// KeyTypeEd25519 indicates an Ed25519 public key
	KeyTypeEd25519 KeyType = "ed25519"

	// KeyTypeSecp256k1 indicates a Secp256k1 public key
	KeyTypeSecp256k1 KeyType = "secp256k1"
)

// IsValid checks if the key type is valid
func (k KeyType) IsValid() bool {
	switch k {
	case KeyTypeEd25519, KeyTypeSecp256k1:
		return true
	default:
		return false
	}
}

// ApprovedClient represents an approved client that can facilitate identity uploads
type ApprovedClient struct {
	// ClientID is a unique identifier for this client
	ClientID string `json:"client_id"`

	// Name is a human-readable name for the client
	Name string `json:"name"`

	// Description is an optional description of the client
	Description string `json:"description,omitempty"`

	// PublicKey is the client's public key for signature verification
	PublicKey []byte `json:"public_key"`

	// KeyType indicates the cryptographic algorithm (ed25519 or secp256k1)
	KeyType KeyType `json:"key_type"`

	// MinVersion is the minimum required client version (semver)
	MinVersion string `json:"min_version"`

	// MaxVersion is the maximum allowed client version (optional, semver)
	MaxVersion string `json:"max_version,omitempty"`

	// AllowedScopes lists the scope types this client can submit
	AllowedScopes []string `json:"allowed_scopes"`

	// Status indicates the current status of the client
	Status ClientStatus `json:"status"`

	// StatusReason explains why the client has its current status
	StatusReason string `json:"status_reason,omitempty"`

	// RegisteredBy is the account address that registered this client
	RegisteredBy string `json:"registered_by"`

	// RegisteredAt is when the client was registered
	RegisteredAt time.Time `json:"registered_at"`

	// LastUpdatedAt is when the client was last updated
	LastUpdatedAt time.Time `json:"last_updated_at"`

	// LastUpdatedBy is the account that last updated this client
	LastUpdatedBy string `json:"last_updated_by,omitempty"`

	// SuspendedAt is when the client was suspended (if applicable)
	SuspendedAt *time.Time `json:"suspended_at,omitempty"`

	// RevokedAt is when the client was revoked (if applicable)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Metadata contains optional additional client information
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewApprovedClient creates a new ApprovedClient
func NewApprovedClient(
	clientID string,
	name string,
	description string,
	publicKey []byte,
	keyType KeyType,
	minVersion string,
	maxVersion string,
	allowedScopes []string,
	registeredBy string,
	registeredAt time.Time,
) *ApprovedClient {
	return &ApprovedClient{
		ClientID:      clientID,
		Name:          name,
		Description:   description,
		PublicKey:     publicKey,
		KeyType:       keyType,
		MinVersion:    minVersion,
		MaxVersion:    maxVersion,
		AllowedScopes: allowedScopes,
		Status:        ClientStatusActive,
		RegisteredBy:  registeredBy,
		RegisteredAt:  registeredAt,
		LastUpdatedAt: registeredAt,
		Metadata:      make(map[string]string),
	}
}

// clientIDRegex validates client ID format
var clientIDRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{2,63}$`)

// Validate validates the ApprovedClient
func (c *ApprovedClient) Validate() error {
	if c.ClientID == "" {
		return ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	if len(c.ClientID) > 64 {
		return ErrInvalidClientID.Wrap("client_id cannot exceed 64 characters")
	}

	if !clientIDRegex.MatchString(c.ClientID) {
		return ErrInvalidClientID.Wrap("client_id must start with a letter and contain only alphanumeric characters, underscores, or hyphens")
	}

	if c.Name == "" {
		return ErrInvalidClientName.Wrap("name cannot be empty")
	}

	if len(c.Name) > 128 {
		return ErrInvalidClientName.Wrap("name cannot exceed 128 characters")
	}

	if len(c.Description) > 512 {
		return ErrInvalidClientDescription.Wrap("description cannot exceed 512 characters")
	}

	if len(c.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public_key cannot be empty")
	}

	if !c.KeyType.IsValid() {
		return ErrInvalidKeyType.Wrapf("invalid key type: %s", c.KeyType)
	}

	// Validate public key length based on key type
	switch c.KeyType {
	case KeyTypeEd25519:
		if len(c.PublicKey) != 32 {
			return ErrInvalidPublicKey.Wrap("ed25519 public key must be 32 bytes")
		}
	case KeyTypeSecp256k1:
		if len(c.PublicKey) != 33 && len(c.PublicKey) != 65 {
			return ErrInvalidPublicKey.Wrap("secp256k1 public key must be 33 (compressed) or 65 (uncompressed) bytes")
		}
	}

	if c.MinVersion == "" {
		return ErrInvalidVersionConstraint.Wrap("min_version cannot be empty")
	}

	if !isValidSemver(c.MinVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid min_version semver: %s", c.MinVersion)
	}

	if c.MaxVersion != "" && !isValidSemver(c.MaxVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid max_version semver: %s", c.MaxVersion)
	}

	if !c.Status.IsValid() {
		return ErrInvalidClientStatus.Wrapf("invalid status: %s", c.Status)
	}

	if c.RegisteredBy == "" {
		return ErrInvalidRegisteredBy.Wrap("registered_by cannot be empty")
	}

	if c.RegisteredAt.IsZero() {
		return ErrInvalidTimestamp.Wrap("registered_at cannot be zero")
	}

	return nil
}

// IsActive returns true if the client is active
func (c *ApprovedClient) IsActive() bool {
	return c.Status == ClientStatusActive
}

// IsSuspended returns true if the client is suspended
func (c *ApprovedClient) IsSuspended() bool {
	return c.Status == ClientStatusSuspended
}

// IsRevoked returns true if the client is revoked
func (c *ApprovedClient) IsRevoked() bool {
	return c.Status == ClientStatusRevoked
}

// CanSubmitScope checks if the client can submit the given scope type
func (c *ApprovedClient) CanSubmitScope(scopeType string) bool {
	if !c.IsActive() {
		return false
	}

	// If no allowed scopes are specified, all scopes are allowed
	if len(c.AllowedScopes) == 0 {
		return true
	}

	for _, allowed := range c.AllowedScopes {
		if allowed == scopeType || allowed == "*" {
			return true
		}
	}

	return false
}

// Suspend suspends the client
func (c *ApprovedClient) Suspend(reason string, updatedBy string, timestamp time.Time) error {
	if !c.Status.CanTransitionTo(ClientStatusSuspended) {
		return ErrInvalidStatusTransition.Wrapf("cannot suspend client in status %s", c.Status)
	}

	c.Status = ClientStatusSuspended
	c.StatusReason = reason
	c.SuspendedAt = &timestamp
	c.LastUpdatedAt = timestamp
	c.LastUpdatedBy = updatedBy
	return nil
}

// Revoke revokes the client
func (c *ApprovedClient) Revoke(reason string, updatedBy string, timestamp time.Time) error {
	if !c.Status.CanTransitionTo(ClientStatusRevoked) {
		return ErrInvalidStatusTransition.Wrapf("cannot revoke client in status %s", c.Status)
	}

	c.Status = ClientStatusRevoked
	c.StatusReason = reason
	c.RevokedAt = &timestamp
	c.LastUpdatedAt = timestamp
	c.LastUpdatedBy = updatedBy
	return nil
}

// Reactivate reactivates a suspended client
func (c *ApprovedClient) Reactivate(reason string, updatedBy string, timestamp time.Time) error {
	if !c.Status.CanTransitionTo(ClientStatusActive) {
		return ErrInvalidStatusTransition.Wrapf("cannot reactivate client in status %s", c.Status)
	}

	c.Status = ClientStatusActive
	c.StatusReason = reason
	c.SuspendedAt = nil
	c.LastUpdatedAt = timestamp
	c.LastUpdatedBy = updatedBy
	return nil
}

// Update updates the client's mutable fields
func (c *ApprovedClient) Update(
	name string,
	description string,
	minVersion string,
	maxVersion string,
	allowedScopes []string,
	updatedBy string,
	timestamp time.Time,
) error {
	if name != "" {
		c.Name = name
	}
	if description != "" {
		c.Description = description
	}
	if minVersion != "" {
		if !isValidSemver(minVersion) {
			return ErrInvalidVersionConstraint.Wrapf("invalid min_version: %s", minVersion)
		}
		c.MinVersion = minVersion
	}
	if maxVersion != "" {
		if !isValidSemver(maxVersion) {
			return ErrInvalidVersionConstraint.Wrapf("invalid max_version: %s", maxVersion)
		}
		c.MaxVersion = maxVersion
	}
	if allowedScopes != nil {
		c.AllowedScopes = allowedScopes
	}

	c.LastUpdatedAt = timestamp
	c.LastUpdatedBy = updatedBy
	return nil
}

// AuditEntry represents an audit log entry for client changes
type AuditEntry struct {
	// ClientID is the client this entry is for
	ClientID string `json:"client_id"`

	// Action is what was done (register, update, suspend, revoke, reactivate)
	Action string `json:"action"`

	// PerformedBy is who performed the action
	PerformedBy string `json:"performed_by"`

	// Timestamp is when the action occurred
	Timestamp time.Time `json:"timestamp"`

	// Reason is why the action was performed
	Reason string `json:"reason,omitempty"`

	// Changes contains field-level changes (for updates)
	Changes map[string]string `json:"changes,omitempty"`
}

// NewAuditEntry creates a new AuditEntry
func NewAuditEntry(clientID string, action string, performedBy string, timestamp time.Time, reason string) *AuditEntry {
	return &AuditEntry{
		ClientID:    clientID,
		Action:      action,
		PerformedBy: performedBy,
		Timestamp:   timestamp,
		Reason:      reason,
		Changes:     make(map[string]string),
	}
}

// isValidSemver checks if a string is a valid semantic version
func isValidSemver(version string) bool {
	// Basic semver validation: X.Y.Z with optional pre-release and build metadata
	// Allows: 1.0.0, 1.0.0-alpha, 1.0.0+build, 1.0.0-alpha+build
	parts := strings.Split(version, "+")
	if len(parts) > 2 {
		return false
	}

	versionCore := parts[0]
	parts = strings.Split(versionCore, "-")
	if len(parts) > 2 {
		return false
	}

	// If there's a pre-release part, it must not be empty
	if len(parts) == 2 && parts[1] == "" {
		return false
	}

	versionNumbers := parts[0]
	numParts := strings.Split(versionNumbers, ".")
	if len(numParts) != 3 {
		return false
	}

	for _, p := range numParts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}

	return true
}
