package data_vault

import (
	"context"
	"io"
)

// VaultService defines the main interface for the encrypted data vault
type VaultService interface {
	// Upload encrypts and stores a new blob
	Upload(ctx context.Context, req *UploadRequest) (*EncryptedBlob, error)

	// Retrieve retrieves and decrypts a blob
	// Performs access control checks and logs audit event
	Retrieve(ctx context.Context, req *RetrieveRequest) ([]byte, *BlobMetadata, error)

	// RetrieveStream retrieves and decrypts a blob as a stream
	// Useful for large blobs (>10MB)
	RetrieveStream(ctx context.Context, req *RetrieveRequest) (io.ReadCloser, *BlobMetadata, error)

	// GetMetadata retrieves blob metadata without decrypting content
	// Access control is enforced based on requester identity.
	GetMetadata(ctx context.Context, id BlobID, requester string, orgID string) (*BlobMetadata, error)

	// Delete marks a blob for deletion
	// Actual deletion may be delayed based on retention policy
	Delete(ctx context.Context, id BlobID, requester string) error

	// RotateKeys initiates key rotation for a scope
	// Old keys remain valid for decryption during transition
	RotateKeys(ctx context.Context, scope Scope) error

	// GetAuditEvents retrieves audit events for a scope or blob
	GetAuditEvents(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error)

	// Close closes the vault service and releases resources
	Close() error
}

// AuditFilter filters audit events
type AuditFilter struct {
	// BlobID filters by blob ID
	BlobID BlobID

	// Scope filters by scope
	Scope Scope

	// Requester filters by requester address
	Requester string

	// OrgID filters by organization
	OrgID string

	// StartTime filters events after this time
	StartTime *int64

	// EndTime filters events before this time
	EndTime *int64

	// Limit limits the number of results
	Limit int
}
