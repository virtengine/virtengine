package artifact_store

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ArchiveVersion is the current archive format version
const ArchiveVersion uint32 = 1

// ArchiveTier represents different storage tiers for archived data
type ArchiveTier string

const (
	// ArchiveTierStandard is standard archival storage (retrieval in minutes)
	ArchiveTierStandard ArchiveTier = "standard"

	// ArchiveTierGlacier is cold storage with slower retrieval (retrieval in hours)
	ArchiveTierGlacier ArchiveTier = "glacier"

	// ArchiveTierDeepArchive is deep cold storage (retrieval in 12-48 hours)
	ArchiveTierDeepArchive ArchiveTier = "deep_archive"

	// ArchiveTierLocal is local filesystem archival
	ArchiveTierLocal ArchiveTier = "local"
)

// IsValid checks if the archive tier is valid
func (t ArchiveTier) IsValid() bool {
	switch t {
	case ArchiveTierStandard, ArchiveTierGlacier, ArchiveTierDeepArchive, ArchiveTierLocal:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (t ArchiveTier) String() string {
	return string(t)
}

// ArchiveStatus represents the status of an archived artifact
type ArchiveStatus string

const (
	// ArchiveStatusPending indicates archival is pending
	ArchiveStatusPending ArchiveStatus = "pending"

	// ArchiveStatusArchiving indicates archival is in progress
	ArchiveStatusArchiving ArchiveStatus = "archiving"

	// ArchiveStatusArchived indicates artifact is archived
	ArchiveStatusArchived ArchiveStatus = "archived"

	// ArchiveStatusRestoring indicates restoration is in progress
	ArchiveStatusRestoring ArchiveStatus = "restoring"

	// ArchiveStatusRestored indicates artifact has been temporarily restored
	ArchiveStatusRestored ArchiveStatus = "restored"

	// ArchiveStatusFailed indicates archival failed
	ArchiveStatusFailed ArchiveStatus = "failed"
)

// IsValid checks if the archive status is valid
func (s ArchiveStatus) IsValid() bool {
	switch s {
	case ArchiveStatusPending, ArchiveStatusArchiving, ArchiveStatusArchived,
		ArchiveStatusRestoring, ArchiveStatusRestored, ArchiveStatusFailed:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (s ArchiveStatus) String() string {
	return string(s)
}

// ArchiveEncryptionEnvelope contains encryption metadata for archived artifacts
// Uses a separate encryption layer from hot storage for key rotation and compliance
type ArchiveEncryptionEnvelope struct {
	// Version is the envelope format version
	Version uint32 `json:"version"`

	// EncryptionAlgorithm is the encryption algorithm (e.g., "AES-256-GCM")
	EncryptionAlgorithm string `json:"encryption_algorithm"`

	// ArchiveKeyID is the identifier for the archive encryption key
	// Supports key rotation for long-term storage
	ArchiveKeyID string `json:"archive_key_id"`

	// Ciphertext is the encrypted data
	Ciphertext []byte `json:"ciphertext"`

	// Nonce is the encryption nonce
	Nonce []byte `json:"nonce"`

	// AuthTag is the authentication tag for GCM mode
	AuthTag []byte `json:"auth_tag"`

	// Timestamp is when the artifact was encrypted
	Timestamp time.Time `json:"timestamp"`

	// OriginalHash is the hash of the unencrypted data for integrity verification
	OriginalHash []byte `json:"original_hash"`

	// OriginalSize is the size of the unencrypted data
	OriginalSize uint64 `json:"original_size"`
}

// Validate validates the archive encryption envelope
func (e *ArchiveEncryptionEnvelope) Validate() error {
	if e.Version == 0 {
		return ErrInvalidInput.Wrap("version cannot be zero")
	}
	if e.EncryptionAlgorithm == "" {
		return ErrInvalidInput.Wrap("encryption_algorithm cannot be empty")
	}
	if e.ArchiveKeyID == "" {
		return ErrInvalidInput.Wrap("archive_key_id cannot be empty")
	}
	if len(e.Ciphertext) == 0 {
		return ErrInvalidInput.Wrap("ciphertext cannot be empty")
	}
	if len(e.Nonce) == 0 {
		return ErrInvalidInput.Wrap("nonce cannot be empty")
	}
	if len(e.OriginalHash) != 32 {
		return ErrInvalidInput.Wrapf("original_hash must be 32 bytes, got %d", len(e.OriginalHash))
	}
	return nil
}

// ArchiveMetadata contains metadata about an archived artifact
type ArchiveMetadata struct {
	// Version is the metadata format version
	Version uint32 `json:"version"`

	// ArchiveID is the unique identifier for this archive
	ArchiveID string `json:"archive_id"`

	// ContentAddress is the original content address from hot storage
	ContentAddress *ContentAddress `json:"content_address"`

	// ArchiveTier specifies the storage tier
	ArchiveTier ArchiveTier `json:"archive_tier"`

	// ArchiveStatus is the current status
	ArchiveStatus ArchiveStatus `json:"archive_status"`

	// ArchivedAt is when the artifact was archived
	ArchivedAt time.Time `json:"archived_at"`

	// ArchivedAtBlock is the block height when archived
	ArchivedAtBlock int64 `json:"archived_at_block"`

	// ArchiveBackend is the backend storing the archive
	ArchiveBackend string `json:"archive_backend"`

	// ArchiveBackendRef is the backend-specific reference
	ArchiveBackendRef string `json:"archive_backend_ref"`

	// EncryptionEnvelope contains the archive encryption metadata
	EncryptionEnvelope *ArchiveEncryptionEnvelope `json:"encryption_envelope"`

	// IntegrityChecksum is a checksum for integrity verification
	IntegrityChecksum []byte `json:"integrity_checksum"`

	// IntegrityAlgorithm is the checksum algorithm (e.g., "sha256")
	IntegrityAlgorithm string `json:"integrity_algorithm"`

	// OriginalOwner is the account that originally owned the artifact
	OriginalOwner string `json:"original_owner"`

	// ArtifactType is the type of artifact
	ArtifactType string `json:"artifact_type"`

	// RetentionTag is the retention policy tag
	RetentionTag *RetentionTag `json:"retention_tag"`

	// LastAccessedAt is when the archive was last accessed
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`

	// AccessCount is the number of times this archive has been accessed
	AccessCount uint64 `json:"access_count"`

	// RestoredAt is when the artifact was last restored (if applicable)
	RestoredAt *time.Time `json:"restored_at,omitempty"`

	// RestoreExpiry is when the restored copy will expire
	RestoreExpiry *time.Time `json:"restore_expiry,omitempty"`

	// ComplianceFlags contains compliance-related flags
	ComplianceFlags *ArchiveComplianceFlags `json:"compliance_flags"`

	// Notes contains optional human-readable notes
	Notes string `json:"notes,omitempty"`
}

// ArchiveComplianceFlags contains compliance-related metadata
type ArchiveComplianceFlags struct {
	// LegalHold indicates if the archive is under legal hold (cannot be deleted)
	LegalHold bool `json:"legal_hold"`

	// LegalHoldReason is the reason for legal hold
	LegalHoldReason string `json:"legal_hold_reason,omitempty"`

	// LegalHoldSetAt is when the legal hold was set
	LegalHoldSetAt *time.Time `json:"legal_hold_set_at,omitempty"`

	// ComplianceRetentionMinDays is the minimum retention required by regulation
	ComplianceRetentionMinDays uint32 `json:"compliance_retention_min_days"`

	// DataClassification is the data classification level
	DataClassification string `json:"data_classification,omitempty"`

	// RegulationTags are regulation identifiers (e.g., "GDPR", "HIPAA", "CCPA")
	RegulationTags []string `json:"regulation_tags,omitempty"`
}

// Validate validates the archive metadata
func (m *ArchiveMetadata) Validate() error {
	if m.Version == 0 {
		return ErrInvalidInput.Wrap("version cannot be zero")
	}
	if m.ArchiveID == "" {
		return ErrInvalidInput.Wrap("archive_id cannot be empty")
	}
	if m.ContentAddress == nil {
		return ErrInvalidInput.Wrap("content_address cannot be nil")
	}
	if err := m.ContentAddress.Validate(); err != nil {
		return err
	}
	if !m.ArchiveTier.IsValid() {
		return ErrInvalidInput.Wrapf("invalid archive_tier: %s", m.ArchiveTier)
	}
	if !m.ArchiveStatus.IsValid() {
		return ErrInvalidInput.Wrapf("invalid archive_status: %s", m.ArchiveStatus)
	}
	if m.ArchivedAt.IsZero() {
		return ErrInvalidInput.Wrap("archived_at cannot be zero")
	}
	if m.ArchiveBackend == "" {
		return ErrInvalidInput.Wrap("archive_backend cannot be empty")
	}
	if m.ArchiveBackendRef == "" {
		return ErrInvalidInput.Wrap("archive_backend_ref cannot be empty")
	}
	if m.EncryptionEnvelope != nil {
		if err := m.EncryptionEnvelope.Validate(); err != nil {
			return err
		}
	}
	if len(m.IntegrityChecksum) == 0 {
		return ErrInvalidInput.Wrap("integrity_checksum cannot be empty")
	}
	if m.IntegrityAlgorithm == "" {
		return ErrInvalidInput.Wrap("integrity_algorithm cannot be empty")
	}
	if m.OriginalOwner == "" {
		return ErrInvalidInput.Wrap("original_owner cannot be empty")
	}
	if m.ArtifactType == "" {
		return ErrInvalidInput.Wrap("artifact_type cannot be empty")
	}
	return nil
}

// IsExpired checks if the archive has expired based on retention policy
func (m *ArchiveMetadata) IsExpired(now time.Time) bool {
	if m.RetentionTag == nil {
		return false
	}
	return m.RetentionTag.IsExpired(now)
}

// CanDelete checks if the archive can be deleted
func (m *ArchiveMetadata) CanDelete(now time.Time) bool {
	// Cannot delete if under legal hold
	if m.ComplianceFlags != nil && m.ComplianceFlags.LegalHold {
		return false
	}

	// Cannot delete if still within compliance retention period
	if m.ComplianceFlags != nil && m.ComplianceFlags.ComplianceRetentionMinDays > 0 {
		minRetentionDuration := time.Duration(m.ComplianceFlags.ComplianceRetentionMinDays) * 24 * time.Hour
		if now.Before(m.ArchivedAt.Add(minRetentionDuration)) {
			return false
		}
	}

	// Can delete if expired based on retention policy
	return m.IsExpired(now)
}

// ComputeIntegrityChecksum computes the integrity checksum
func (m *ArchiveMetadata) ComputeIntegrityChecksum() []byte {
	if m.EncryptionEnvelope == nil || len(m.EncryptionEnvelope.Ciphertext) == 0 {
		return nil
	}

	hash := sha256.Sum256(m.EncryptionEnvelope.Ciphertext)
	m.IntegrityChecksum = hash[:]
	m.IntegrityAlgorithm = "sha256"
	return m.IntegrityChecksum
}

// IntegrityChecksumHex returns the integrity checksum as hex string
func (m *ArchiveMetadata) IntegrityChecksumHex() string {
	return hex.EncodeToString(m.IntegrityChecksum)
}

// ArchivalEligibilityCriteria defines criteria for archival eligibility
type ArchivalEligibilityCriteria struct {
	// MinAgeSeconds is the minimum age in seconds before archival
	MinAgeSeconds int64 `json:"min_age_seconds"`

	// MinAgeBlocks is the minimum age in blocks before archival
	MinAgeBlocks int64 `json:"min_age_blocks"`

	// MaxAccessCount is the max access count to be eligible (low access = eligible)
	MaxAccessCount uint64 `json:"max_access_count"`

	// LastAccessThresholdSeconds is how long since last access to be eligible
	LastAccessThresholdSeconds int64 `json:"last_access_threshold_seconds"`

	// RequireRetentionPolicy indicates if retention policy is required
	RequireRetentionPolicy bool `json:"require_retention_policy"`

	// ExcludeArtifactTypes are artifact types to exclude from archival
	ExcludeArtifactTypes []string `json:"exclude_artifact_types"`
}

// DefaultArchivalEligibilityCriteria returns default archival criteria
// Data is eligible for archival after 90 days with no recent access
func DefaultArchivalEligibilityCriteria() *ArchivalEligibilityCriteria {
	return &ArchivalEligibilityCriteria{
		MinAgeSeconds:              90 * 24 * 60 * 60, // 90 days
		MinAgeBlocks:               0,                 // Not used by default
		MaxAccessCount:             5,                 // Low access count
		LastAccessThresholdSeconds: 30 * 24 * 60 * 60, // 30 days since last access
		RequireRetentionPolicy:     true,
		ExcludeArtifactTypes: []string{
			"verification_record", // Keep verification records in hot storage
		},
	}
}

// IsEligibleForArchival checks if an artifact is eligible for archival
func (c *ArchivalEligibilityCriteria) IsEligibleForArchival(
	createdAt time.Time,
	createdAtBlock int64,
	lastAccessedAt *time.Time,
	accessCount uint64,
	artifactType string,
	hasRetentionPolicy bool,
	now time.Time,
	currentBlock int64,
) bool {
	// Check if artifact type is excluded
	for _, excluded := range c.ExcludeArtifactTypes {
		if artifactType == excluded {
			return false
		}
	}

	// Check if retention policy is required
	if c.RequireRetentionPolicy && !hasRetentionPolicy {
		return false
	}

	// Check minimum age in seconds
	if c.MinAgeSeconds > 0 {
		age := now.Sub(createdAt).Seconds()
		if age < float64(c.MinAgeSeconds) {
			return false
		}
	}

	// Check minimum age in blocks
	if c.MinAgeBlocks > 0 {
		ageBlocks := currentBlock - createdAtBlock
		if ageBlocks < c.MinAgeBlocks {
			return false
		}
	}

	// Check access count
	if c.MaxAccessCount > 0 && accessCount > c.MaxAccessCount {
		return false // Too frequently accessed
	}

	// Check last access threshold
	if c.LastAccessThresholdSeconds > 0 && lastAccessedAt != nil {
		timeSinceLastAccess := now.Sub(*lastAccessedAt).Seconds()
		if timeSinceLastAccess < float64(c.LastAccessThresholdSeconds) {
			return false // Recently accessed
		}
	}

	return true
}

// ArchiveOperationResult contains the result of an archive operation
type ArchiveOperationResult struct {
	// Success indicates if the operation was successful
	Success bool `json:"success"`

	// ArchiveMetadata contains the archive metadata (on success)
	ArchiveMetadata *ArchiveMetadata `json:"archive_metadata,omitempty"`

	// Error contains error message (on failure)
	Error string `json:"error,omitempty"`

	// Duration is how long the operation took
	Duration time.Duration `json:"duration"`

	// BytesProcessed is the number of bytes processed
	BytesProcessed uint64 `json:"bytes_processed"`
}

// RestoreRequest contains parameters for restoring an archived artifact
type RestoreRequest struct {
	// ArchiveID is the unique identifier of the archive
	ArchiveID string `json:"archive_id"`

	// RequestingAccount is the account requesting restoration
	RequestingAccount string `json:"requesting_account"`

	// RestoreTier specifies the restore speed tier
	// "expedited" = fastest (minutes), "standard" = normal (hours), "bulk" = slowest (12-48h)
	RestoreTier string `json:"restore_tier"`

	// RestoreDurationSeconds is how long to keep the restored copy available
	RestoreDurationSeconds int64 `json:"restore_duration_seconds"`

	// AuthToken is an optional authentication token
	AuthToken string `json:"auth_token,omitempty"`
}

// Validate validates the restore request
func (r *RestoreRequest) Validate() error {
	if r.ArchiveID == "" {
		return ErrInvalidInput.Wrap("archive_id cannot be empty")
	}
	if r.RequestingAccount == "" {
		return ErrInvalidInput.Wrap("requesting_account cannot be empty")
	}
	if r.RestoreTier == "" {
		r.RestoreTier = "standard" // Default to standard
	}
	if r.RestoreDurationSeconds <= 0 {
		r.RestoreDurationSeconds = 24 * 60 * 60 // Default 24 hours
	}
	return nil
}

// RestoreResponse contains the result of a restore operation
type RestoreResponse struct {
	// ArchiveMetadata contains the updated archive metadata
	ArchiveMetadata *ArchiveMetadata `json:"archive_metadata"`

	// EstimatedRestoreTime is when the restore is expected to complete
	EstimatedRestoreTime time.Time `json:"estimated_restore_time"`

	// RestoreExpiry is when the restored copy will expire
	RestoreExpiry time.Time `json:"restore_expiry"`

	// RestoreJobID is the backend-specific restore job identifier
	RestoreJobID string `json:"restore_job_id,omitempty"`
}
