package artifact_store

import (
	"context"
	"time"
)

// ArchiveBackend defines the interface for cold storage archival operations.
// Implementations must support long-term storage with encryption and compliance features.
//
// Supported backends:
//   - S3 Glacier/Deep Archive
//   - Azure Archive Storage
//   - Local filesystem archival
//   - Google Cloud Archive Storage
//
// All archived artifacts must be encrypted with archive-specific encryption keys
// to support key rotation and compliance requirements.
type ArchiveBackend interface {
	// Archive stores an artifact in cold storage
	// The artifact must already be encrypted using ArchiveEncryptionEnvelope
	Archive(ctx context.Context, req *ArchiveRequest) (*ArchiveResponse, error)

	// GetArchiveMetadata retrieves metadata about an archived artifact
	// Does not retrieve the actual artifact data
	GetArchiveMetadata(ctx context.Context, archiveID string) (*ArchiveMetadata, error)

	// Restore initiates restoration of an archived artifact to hot storage
	// Restoration may be asynchronous and take hours depending on storage tier
	Restore(ctx context.Context, req *RestoreRequest) (*RestoreResponse, error)

	// GetRestoreStatus checks the status of a restore operation
	GetRestoreStatus(ctx context.Context, archiveID string) (*RestoreStatus, error)

	// GetArchived retrieves a restored artifact's data
	// Only works after Restore() has completed successfully
	GetArchived(ctx context.Context, archiveID string) (*GetArchivedResponse, error)

	// DeleteArchive permanently deletes an archived artifact
	// This is a privileged operation with compliance checks
	DeleteArchive(ctx context.Context, req *DeleteArchiveRequest) error

	// UpdateArchiveMetadata updates archive metadata (e.g., retention tags, legal hold)
	UpdateArchiveMetadata(ctx context.Context, archiveID string, updates *ArchiveMetadataUpdate) error

	// ListArchives lists archives by owner with pagination
	ListArchives(ctx context.Context, owner string, pagination *Pagination) (*ListArchivesResponse, error)

	// PurgeExpiredArchives removes expired archives based on retention policies
	// Returns the number of archives purged
	PurgeExpiredArchives(ctx context.Context, currentTime int64) (int, error)

	// VerifyArchiveIntegrity verifies the integrity of an archived artifact
	// Checks that checksums match and data is not corrupted
	VerifyArchiveIntegrity(ctx context.Context, archiveID string) error

	// GetArchiveMetrics returns metrics about archived storage
	GetArchiveMetrics(ctx context.Context) (*ArchiveMetrics, error)

	// Health checks if the archive backend is healthy
	Health(ctx context.Context) error

	// Backend returns the backend type
	Backend() string
}

// ArchiveRequest contains parameters for archiving an artifact
type ArchiveRequest struct {
	// ContentAddress is the original content address from hot storage
	ContentAddress *ContentAddress

	// Data is the encrypted artifact data
	Data []byte

	// ArchiveTier specifies the storage tier
	ArchiveTier ArchiveTier

	// EncryptionEnvelope contains the archive encryption metadata
	EncryptionEnvelope *ArchiveEncryptionEnvelope

	// RetentionTag specifies the retention policy
	RetentionTag *RetentionTag

	// Owner is the account that owns this artifact
	Owner string

	// ArtifactType is the type of artifact
	ArtifactType string

	// ComplianceFlags contains compliance metadata
	ComplianceFlags *ArchiveComplianceFlags

	// Metadata contains optional additional metadata
	Metadata map[string]string

	// Notes contains optional human-readable notes
	Notes string
}

// Validate validates the archive request
func (r *ArchiveRequest) Validate() error {
	if r.ContentAddress == nil {
		return ErrInvalidInput.Wrap("content_address cannot be nil")
	}
	if err := r.ContentAddress.Validate(); err != nil {
		return err
	}
	if len(r.Data) == 0 {
		return ErrInvalidInput.Wrap("data cannot be empty")
	}
	if !r.ArchiveTier.IsValid() {
		return ErrInvalidInput.Wrapf("invalid archive_tier: %s", r.ArchiveTier)
	}
	if r.EncryptionEnvelope == nil {
		return ErrInvalidInput.Wrap("encryption_envelope cannot be nil")
	}
	if err := r.EncryptionEnvelope.Validate(); err != nil {
		return err
	}
	if r.Owner == "" {
		return ErrInvalidInput.Wrap("owner cannot be empty")
	}
	if r.ArtifactType == "" {
		return ErrInvalidInput.Wrap("artifact_type cannot be empty")
	}
	return nil
}

// ArchiveResponse contains the result of an archive operation
type ArchiveResponse struct {
	// ArchiveMetadata contains the archive metadata
	ArchiveMetadata *ArchiveMetadata

	// ArchiveID is the unique identifier for this archive
	ArchiveID string

	// ArchiveBackendRef is the backend-specific reference
	ArchiveBackendRef string
}

// RestoreStatus contains the status of a restore operation
type RestoreStatus struct {
	// ArchiveID is the archive identifier
	ArchiveID string `json:"archive_id"`

	// Status is the current restore status
	Status ArchiveStatus `json:"status"`

	// Progress is the restore progress percentage (0-100)
	Progress uint8 `json:"progress"`

	// EstimatedCompletionTime is when the restore is expected to complete
	EstimatedCompletionTime *time.Time `json:"estimated_completion_time,omitempty"`

	// RestoreJobID is the backend-specific restore job identifier
	RestoreJobID string `json:"restore_job_id,omitempty"`

	// Error contains error message if restore failed
	Error string `json:"error,omitempty"`
}

// GetArchivedResponse contains the retrieved archived artifact data
type GetArchivedResponse struct {
	// Data is the encrypted artifact data
	Data []byte

	// ArchiveMetadata contains the archive metadata
	ArchiveMetadata *ArchiveMetadata

	// RestoredAt is when this artifact was restored
	RestoredAt time.Time

	// RestoreExpiry is when the restored copy will expire
	RestoreExpiry time.Time
}

// DeleteArchiveRequest contains parameters for deleting an archived artifact
type DeleteArchiveRequest struct {
	// ArchiveID is the archive identifier
	ArchiveID string

	// RequestingAccount is the account requesting deletion
	RequestingAccount string

	// AuthToken is an optional authentication token
	AuthToken string

	// Force indicates if deletion should bypass compliance checks (dangerous!)
	Force bool

	// Reason is the reason for deletion (for audit trail)
	Reason string
}

// Validate validates the delete archive request
func (r *DeleteArchiveRequest) Validate() error {
	if r.ArchiveID == "" {
		return ErrInvalidInput.Wrap("archive_id cannot be empty")
	}
	if r.RequestingAccount == "" {
		return ErrInvalidInput.Wrap("requesting_account cannot be empty")
	}
	if r.Reason == "" {
		return ErrInvalidInput.Wrap("reason cannot be empty (required for audit trail)")
	}
	return nil
}

// ArchiveMetadataUpdate contains fields that can be updated on archive metadata
type ArchiveMetadataUpdate struct {
	// RetentionTag updates the retention policy tag
	RetentionTag *RetentionTag `json:"retention_tag,omitempty"`

	// ComplianceFlags updates compliance flags
	ComplianceFlags *ArchiveComplianceFlags `json:"compliance_flags,omitempty"`

	// Notes updates notes field
	Notes *string `json:"notes,omitempty"`
}

// ListArchivesResponse contains a list of archive metadata
type ListArchivesResponse struct {
	// Archives are the matching archive metadata entries
	Archives []*ArchiveMetadata

	// Total is the total count (ignoring pagination)
	Total uint64

	// HasMore indicates if more results are available
	HasMore bool
}

// ArchiveMetrics contains metrics about archived storage
type ArchiveMetrics struct {
	// TotalArchives is the total number of archives
	TotalArchives uint64 `json:"total_archives"`

	// TotalBytes is the total storage used in bytes
	TotalBytes uint64 `json:"total_bytes"`

	// ArchivesByTier contains counts by storage tier
	ArchivesByTier map[ArchiveTier]uint64 `json:"archives_by_tier"`

	// ArchivesByStatus contains counts by status
	ArchivesByStatus map[ArchiveStatus]uint64 `json:"archives_by_status"`

	// ExpiredArchives is the number of expired archives awaiting cleanup
	ExpiredArchives uint64 `json:"expired_archives"`

	// LegalHoldArchives is the number of archives under legal hold
	LegalHoldArchives uint64 `json:"legal_hold_archives"`

	// BackendType is the backend type
	BackendType string `json:"backend_type"`

	// BackendStatus contains backend-specific status information
	BackendStatus map[string]string `json:"backend_status,omitempty"`

	// LastPurgeTime is when the last purge was performed
	LastPurgeTime *time.Time `json:"last_purge_time,omitempty"`

	// LastIntegrityCheckTime is when the last integrity check was performed
	LastIntegrityCheckTime *time.Time `json:"last_integrity_check_time,omitempty"`
}

// ArchivalManager coordinates archival operations across backends
// This is the high-level interface used by the keeper
type ArchivalManager interface {
	// DetermineEligibleArtifacts identifies artifacts eligible for archival
	DetermineEligibleArtifacts(ctx context.Context, criteria *ArchivalEligibilityCriteria) ([]*ContentAddress, error)

	// ArchiveArtifact archives a single artifact from hot storage
	ArchiveArtifact(ctx context.Context, contentAddress *ContentAddress, tier ArchiveTier) (*ArchiveOperationResult, error)

	// RestoreArtifact restores an archived artifact to hot storage
	RestoreArtifact(ctx context.Context, archiveID string, restoreTier string) (*RestoreResponse, error)

	// CleanupExpiredArtifacts removes artifacts from hot storage after archival
	CleanupExpiredArtifacts(ctx context.Context) (int, error)

	// GetArchivalStatus returns the status of archival operations
	GetArchivalStatus(ctx context.Context) (*ArchivalStatus, error)
}

// ArchivalStatus contains status information about archival operations
type ArchivalStatus struct {
	// TotalArchived is the total number of artifacts archived
	TotalArchived uint64 `json:"total_archived"`

	// TotalRestored is the total number of artifacts restored
	TotalRestored uint64 `json:"total_restored"`

	// PendingArchival is the number of artifacts pending archival
	PendingArchival uint64 `json:"pending_archival"`

	// PendingRestoration is the number of artifacts pending restoration
	PendingRestoration uint64 `json:"pending_restoration"`

	// FailedArchival is the number of failed archival operations
	FailedArchival uint64 `json:"failed_archival"`

	// FailedRestoration is the number of failed restoration operations
	FailedRestoration uint64 `json:"failed_restoration"`

	// LastArchivalTime is when the last successful archival was performed
	LastArchivalTime *time.Time `json:"last_archival_time,omitempty"`

	// LastRestorationTime is when the last successful restoration was performed
	LastRestorationTime *time.Time `json:"last_restoration_time,omitempty"`
}
