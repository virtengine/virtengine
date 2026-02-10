package types

import (
	"time"
)

// ArchivalEvent represents different archival lifecycle events
type ArchivalEvent string

const (
	// ArchivalEventPending indicates archival is scheduled
	ArchivalEventPending ArchivalEvent = "pending"

	// ArchivalEventStarted indicates archival has started
	ArchivalEventStarted ArchivalEvent = "started"

	// ArchivalEventCompleted indicates archival completed successfully
	ArchivalEventCompleted ArchivalEvent = "completed"

	// ArchivalEventFailed indicates archival failed
	ArchivalEventFailed ArchivalEvent = "failed"

	// ArchivalEventRestoreRequested indicates restoration was requested
	ArchivalEventRestoreRequested ArchivalEvent = "restore_requested"

	// ArchivalEventRestoreCompleted indicates restoration completed
	ArchivalEventRestoreCompleted ArchivalEvent = "restore_completed"

	// ArchivalEventDeleted indicates archived artifact was deleted
	ArchivalEventDeleted ArchivalEvent = "deleted"
)

// ArchivalRecord represents an on-chain record of an archived artifact
// This provides a lightweight reference to cold storage artifacts
type ArchivalRecord struct {
	// ArchiveID is the unique identifier for the archive
	ArchiveID string `json:"archive_id"`

	// ContentHash is the hash of the original content
	ContentHash []byte `json:"content_hash"`

	// Owner is the account that owns this artifact
	Owner string `json:"owner"`

	// ArtifactType is the type of artifact
	ArtifactType ArtifactType `json:"artifact_type"`

	// ArchiveTier is the storage tier
	ArchiveTier string `json:"archive_tier"`

	// ArchivedAt is when the artifact was archived
	ArchivedAt time.Time `json:"archived_at"`

	// ArchivedAtBlock is the block height when archived
	ArchivedAtBlock int64 `json:"archived_at_block"`

	// ArchiveBackend is the backend storing the archive
	ArchiveBackend string `json:"archive_backend"`

	// RetentionPolicyID is the retention policy applied
	RetentionPolicyID string `json:"retention_policy_id"`

	// IntegrityChecksum is the checksum for verification
	IntegrityChecksum []byte `json:"integrity_checksum"`

	// IsRestored indicates if currently restored to hot storage
	IsRestored bool `json:"is_restored"`

	// RestoreExpiry is when the restored copy expires
	RestoreExpiry *time.Time `json:"restore_expiry,omitempty"`

	// AccessCount is the number of times accessed
	AccessCount uint64 `json:"access_count"`

	// LastAccessedAt is when last accessed
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`

	// ArchivalComplianceStatus contains compliance-related metadata
	ArchivalComplianceStatus *ArchivalComplianceStatus `json:"compliance_status,omitempty"`
}

// ArchivalComplianceStatus contains compliance-related status information
type ArchivalComplianceStatus struct {
	// LegalHold indicates if under legal hold
	LegalHold bool `json:"legal_hold"`

	// LegalHoldSetAt is when legal hold was set
	LegalHoldSetAt *time.Time `json:"legal_hold_set_at,omitempty"`

	// ComplianceRetentionMinDays is minimum retention required
	ComplianceRetentionMinDays uint32 `json:"compliance_retention_min_days"`

	// RegulationTags are applicable regulations
	RegulationTags []string `json:"regulation_tags,omitempty"`

	// LastComplianceCheck is when last checked for compliance
	LastComplianceCheck *time.Time `json:"last_compliance_check,omitempty"`

	// ComplianceViolations contains any compliance issues
	ComplianceViolations []string `json:"compliance_violations,omitempty"`
}

// Validate validates the archival record
func (r *ArchivalRecord) Validate() error {
	if r.ArchiveID == "" {
		return ErrInvalidParams.Wrap("archive_id cannot be empty")
	}
	if len(r.ContentHash) != 32 {
		return ErrInvalidParams.Wrapf("content_hash must be 32 bytes, got %d", len(r.ContentHash))
	}
	if r.Owner == "" {
		return ErrInvalidParams.Wrap("owner cannot be empty")
	}
	if !IsValidArtifactType(r.ArtifactType) {
		return ErrInvalidParams.Wrapf("invalid artifact_type: %s", r.ArtifactType)
	}
	if r.ArchivedAt.IsZero() {
		return ErrInvalidParams.Wrap("archived_at cannot be zero")
	}
	if r.ArchiveBackend == "" {
		return ErrInvalidParams.Wrap("archive_backend cannot be empty")
	}
	if len(r.IntegrityChecksum) == 0 {
		return ErrInvalidParams.Wrap("integrity_checksum cannot be empty")
	}
	return nil
}

// IsExpired checks if the restored copy has expired
func (r *ArchivalRecord) IsExpired(now time.Time) bool {
	if !r.IsRestored || r.RestoreExpiry == nil {
		return false
	}
	return now.After(*r.RestoreExpiry)
}

// CanDelete checks if the archive can be deleted
func (r *ArchivalRecord) CanDelete() bool {
	if r.ArchivalComplianceStatus != nil && r.ArchivalComplianceStatus.LegalHold {
		return false
	}
	return true
}

// ArchivalIndex tracks archival operations for efficient querying
type ArchivalIndex struct {
	// TotalArchived is the total number of archived artifacts
	TotalArchived uint64 `json:"total_archived"`

	// TotalRestored is the total number currently restored
	TotalRestored uint64 `json:"total_restored"`

	// ByOwner tracks archives per owner
	ByOwner map[string]uint64 `json:"by_owner"`

	// ByArtifactType tracks archives per artifact type
	ByArtifactType map[string]uint64 `json:"by_artifact_type"`

	// ByTier tracks archives per storage tier
	ByTier map[string]uint64 `json:"by_tier"`

	// LastUpdated is when the index was last updated
	LastUpdated time.Time `json:"last_updated"`

	// LastUpdatedBlock is the block height when last updated
	LastUpdatedBlock int64 `json:"last_updated_block"`
}

// NewArchivalIndex creates a new archival index
func NewArchivalIndex() *ArchivalIndex {
	return &ArchivalIndex{
		ByOwner:        make(map[string]uint64),
		ByArtifactType: make(map[string]uint64),
		ByTier:         make(map[string]uint64),
	}
}

// AddArchive updates index when an artifact is archived
func (idx *ArchivalIndex) AddArchive(record *ArchivalRecord) {
	idx.TotalArchived++
	idx.ByOwner[record.Owner]++
	idx.ByArtifactType[string(record.ArtifactType)]++
	idx.ByTier[record.ArchiveTier]++
}

// RemoveArchive updates index when an artifact is removed
func (idx *ArchivalIndex) RemoveArchive(record *ArchivalRecord) {
	if idx.TotalArchived > 0 {
		idx.TotalArchived--
	}
	if idx.ByOwner[record.Owner] > 0 {
		idx.ByOwner[record.Owner]--
	}
	if idx.ByArtifactType[string(record.ArtifactType)] > 0 {
		idx.ByArtifactType[string(record.ArtifactType)]--
	}
	if idx.ByTier[record.ArchiveTier] > 0 {
		idx.ByTier[record.ArchiveTier]--
	}
}

// MarkRestored updates index when an artifact is restored
func (idx *ArchivalIndex) MarkRestored() {
	idx.TotalRestored++
}

// UnmarkRestored updates index when a restored artifact expires
func (idx *ArchivalIndex) UnmarkRestored() {
	if idx.TotalRestored > 0 {
		idx.TotalRestored--
	}
}

// ArchivalConfig contains configuration for archival operations
type ArchivalConfig struct {
	// Enabled indicates if archival is enabled
	Enabled bool `json:"enabled"`

	// AutoArchive indicates if automatic archival is enabled
	AutoArchive bool `json:"auto_archive"`

	// ArchivalCheckInterval is how often to check for eligible artifacts
	ArchivalCheckIntervalBlocks int64 `json:"archival_check_interval_blocks"`

	// DefaultArchiveTier is the default storage tier
	DefaultArchiveTier string `json:"default_archive_tier"`

	// MaxArchivesPerBlock is the maximum number to archive per block
	MaxArchivesPerBlock uint32 `json:"max_archives_per_block"`

	// MinAgeForArchival is the minimum age in blocks before archival
	MinAgeForArchivalBlocks int64 `json:"min_age_for_archival_blocks"`

	// MinAgeForArchivalSeconds is the minimum age in seconds
	MinAgeForArchivalSeconds int64 `json:"min_age_for_archival_seconds"`

	// RestoreTTLSeconds is how long restored artifacts remain available
	RestoreTTLSeconds int64 `json:"restore_ttl_seconds"`

	// EnableIntegrityChecks enables periodic integrity verification
	EnableIntegrityChecks bool `json:"enable_integrity_checks"`

	// IntegrityCheckIntervalBlocks is how often to check integrity
	IntegrityCheckIntervalBlocks int64 `json:"integrity_check_interval_blocks"`
}

// DefaultArchivalConfig returns default archival configuration
func DefaultArchivalConfig() *ArchivalConfig {
	return &ArchivalConfig{
		Enabled:                      true,
		AutoArchive:                  true,
		ArchivalCheckIntervalBlocks:  100,               // Every 100 blocks (~10 minutes)
		DefaultArchiveTier:           "standard",        // Standard tier by default
		MaxArchivesPerBlock:          10,                // Max 10 per block to avoid congestion
		MinAgeForArchivalBlocks:      12960,             // ~90 days (assuming 6s blocks)
		MinAgeForArchivalSeconds:     90 * 24 * 60 * 60, // 90 days
		RestoreTTLSeconds:            24 * 60 * 60,      // 24 hours
		EnableIntegrityChecks:        true,
		IntegrityCheckIntervalBlocks: 14400, // Daily (assuming 6s blocks)
	}
}

// Validate validates the archival configuration
func (c *ArchivalConfig) Validate() error {
	if c.ArchivalCheckIntervalBlocks <= 0 {
		return ErrInvalidParams.Wrap("archival_check_interval_blocks must be positive")
	}
	if c.MaxArchivesPerBlock == 0 {
		return ErrInvalidParams.Wrap("max_archives_per_block must be positive")
	}
	if c.RestoreTTLSeconds <= 0 {
		return ErrInvalidParams.Wrap("restore_ttl_seconds must be positive")
	}
	return nil
}
