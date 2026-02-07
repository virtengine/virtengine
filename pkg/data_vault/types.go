package data_vault

import (
	"time"

	enctypes "github.com/virtengine/virtengine/x/encryption/types"
)

// Scope defines the data classification for vault blobs
type Scope string

const (
	// ScopeVEID is for VEID identity documents and attestations
	ScopeVEID Scope = "veid"

	// ScopeSupport is for support ticket attachments
	ScopeSupport Scope = "support"

	// ScopeMarket is for marketplace deployment artifacts
	ScopeMarket Scope = "market"

	// ScopeAudit is for audit logs and compliance artifacts
	ScopeAudit Scope = "audit"
)

// BlobID is a unique identifier for a vault blob
type BlobID string

// BlobMetadata contains metadata for an encrypted blob
type BlobMetadata struct {
	// ID is the unique blob identifier
	ID BlobID `json:"id"`

	// Scope defines the data classification
	Scope Scope `json:"scope"`

	// KeyID identifies which DEK encrypted this blob
	KeyID string `json:"key_id"`

	// KeyVersion is the version of the DEK
	KeyVersion uint32 `json:"key_version"`

	// ContentHash is the SHA-256 hash of the plaintext content
	ContentHash []byte `json:"content_hash"`

	// Size is the plaintext size in bytes
	Size int64 `json:"size"`

	// EncryptedSize is the ciphertext size in bytes
	EncryptedSize int64 `json:"encrypted_size"`

	// Owner is the wallet address that created this blob
	Owner string `json:"owner"`

	// OrgID is the organization ID for multi-tenancy
	OrgID string `json:"org_id,omitempty"`

	// CreatedAt is when the blob was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is optional expiration time
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// RetentionPolicy defines retention requirements
	RetentionPolicy string `json:"retention_policy,omitempty"`

	// Tags are optional public metadata tags
	Tags map[string]string `json:"tags,omitempty"`
}

// EncryptedBlob represents a stored encrypted blob with metadata
type EncryptedBlob struct {
	// Metadata contains the blob metadata
	Metadata BlobMetadata `json:"metadata"`

	// Envelope is the encrypted payload envelope
	Envelope *enctypes.EncryptedPayloadEnvelope `json:"envelope"`

	// BackendPath is the storage backend path (filesystem, IPFS, etc.)
	BackendPath string `json:"backend_path"`
}

// UploadRequest is a request to upload a new blob
type UploadRequest struct {
	// Scope defines the data classification
	Scope Scope

	// Plaintext is the data to encrypt
	Plaintext []byte

	// Owner is the wallet address creating this blob
	Owner string

	// OrgID is optional organization ID
	OrgID string

	// RetentionPolicy defines retention requirements
	RetentionPolicy string

	// ExpiresAt is optional expiration time
	ExpiresAt *time.Time

	// Tags are optional public metadata
	Tags map[string]string
}

// RetrieveRequest is a request to retrieve and decrypt a blob
type RetrieveRequest struct {
	// ID is the blob identifier
	ID BlobID

	// Requester is the wallet address requesting access
	Requester string

	// OrgID is the requester's organization ID
	OrgID string
}

// AuditEvent records a vault access event
type AuditEvent struct {
	// ID is the unique event identifier
	ID string `json:"id"`

	// EventType is the type of event (decrypt, upload, rotate, etc.)
	EventType string `json:"event_type"`

	// BlobID is the blob being accessed
	BlobID BlobID `json:"blob_id"`

	// Scope is the blob scope
	Scope Scope `json:"scope"`

	// Requester is the wallet address that performed the action
	Requester string `json:"requester"`

	// OrgID is the requester's organization
	OrgID string `json:"org_id,omitempty"`

	// Success indicates if the operation succeeded
	Success bool `json:"success"`

	// Error is the error message if failed
	Error string `json:"error,omitempty"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}
