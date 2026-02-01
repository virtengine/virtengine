package artifact_store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FilesystemArchiveBackend implements ArchiveBackend using local filesystem
// This is primarily for development, testing, and small deployments
// Production deployments should use cloud archive storage (S3 Glacier, etc.)
type FilesystemArchiveBackend struct {
	// archiveDir is the root directory for archived artifacts
	archiveDir string

	// metadataDir is the directory for archive metadata
	metadataDir string

	// mu protects concurrent access
	mu sync.RWMutex

	// config contains configuration options
	config *FilesystemArchiveConfig
}

// FilesystemArchiveConfig contains configuration for filesystem archive backend
type FilesystemArchiveConfig struct {
	// ArchiveDir is the root directory for archives
	ArchiveDir string

	// MetadataDir is the directory for metadata (defaults to ArchiveDir/.metadata)
	MetadataDir string

	// RestoreTTLSeconds is how long restored artifacts remain available
	RestoreTTLSeconds int64

	// SimulateRestoreDelay simulates slow restoration for testing
	SimulateRestoreDelay time.Duration

	// MaxArchiveSize is the maximum size of a single archive (0 = unlimited)
	MaxArchiveSize uint64

	// EnableIntegrityChecks enables periodic integrity verification
	EnableIntegrityChecks bool

	// IntegrityCheckInterval is how often to check integrity
	IntegrityCheckInterval time.Duration
}

// DefaultFilesystemArchiveConfig returns default configuration
func DefaultFilesystemArchiveConfig(archiveDir string) *FilesystemArchiveConfig {
	return &FilesystemArchiveConfig{
		ArchiveDir:             archiveDir,
		MetadataDir:            filepath.Join(archiveDir, ".metadata"),
		RestoreTTLSeconds:      24 * 60 * 60, // 24 hours
		SimulateRestoreDelay:   0,
		MaxArchiveSize:         0, // unlimited
		EnableIntegrityChecks:  true,
		IntegrityCheckInterval: 24 * time.Hour,
	}
}

// NewFilesystemArchiveBackend creates a new filesystem archive backend
func NewFilesystemArchiveBackend(config *FilesystemArchiveConfig) (*FilesystemArchiveBackend, error) {
	if config == nil {
		return nil, ErrInvalidInput.Wrap("config cannot be nil")
	}

	if config.ArchiveDir == "" {
		return nil, ErrInvalidInput.Wrap("archive_dir cannot be empty")
	}

	if config.MetadataDir == "" {
		config.MetadataDir = filepath.Join(config.ArchiveDir, ".metadata")
	}

	// Create directories
	if err := os.MkdirAll(config.ArchiveDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create archive directory: %w", err)
	}
	if err := os.MkdirAll(config.MetadataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	return &FilesystemArchiveBackend{
		archiveDir:  config.ArchiveDir,
		metadataDir: config.MetadataDir,
		config:      config,
	}, nil
}

// Archive stores an artifact in filesystem cold storage
func (b *FilesystemArchiveBackend) Archive(ctx context.Context, req *ArchiveRequest) (*ArchiveResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Check max archive size
	if b.config.MaxArchiveSize > 0 && uint64(len(req.Data)) > b.config.MaxArchiveSize {
		return nil, ErrInvalidInput.Wrapf("archive size %d exceeds maximum %d", len(req.Data), b.config.MaxArchiveSize)
	}

	// Generate archive ID
	archiveID := generateArchiveID(req.ContentAddress, req.Owner)

	// Compute integrity checksum
	checksum := sha256.Sum256(req.Data)

	// Create archive metadata
	now := time.Now()
	metadata := &ArchiveMetadata{
		Version:            ArchiveVersion,
		ArchiveID:          archiveID,
		ContentAddress:     req.ContentAddress,
		ArchiveTier:        req.ArchiveTier,
		ArchiveStatus:      ArchiveStatusArchived,
		ArchivedAt:         now,
		ArchivedAtBlock:    0, // Set by caller
		ArchiveBackend:     "filesystem",
		ArchiveBackendRef:  archiveID,
		EncryptionEnvelope: req.EncryptionEnvelope,
		IntegrityChecksum:  checksum[:],
		IntegrityAlgorithm: "sha256",
		OriginalOwner:      req.Owner,
		ArtifactType:       req.ArtifactType,
		RetentionTag:       req.RetentionTag,
		AccessCount:        0,
		ComplianceFlags:    req.ComplianceFlags,
		Notes:              req.Notes,
	}

	// Write archive data
	archivePath := b.getArchivePath(archiveID)
	if err := os.MkdirAll(filepath.Dir(archivePath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create archive subdirectory: %w", err)
	}
	if err := os.WriteFile(archivePath, req.Data, 0600); err != nil {
		return nil, fmt.Errorf("failed to write archive data: %w", err)
	}

	// Write metadata
	if err := b.writeMetadata(archiveID, metadata); err != nil {
		// Cleanup archive file on metadata write failure
		_ = os.Remove(archivePath)
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	return &ArchiveResponse{
		ArchiveMetadata:   metadata,
		ArchiveID:         archiveID,
		ArchiveBackendRef: archiveID,
	}, nil
}

// GetArchiveMetadata retrieves metadata about an archived artifact
func (b *FilesystemArchiveBackend) GetArchiveMetadata(ctx context.Context, archiveID string) (*ArchiveMetadata, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.readMetadata(archiveID)
}

// Restore initiates restoration of an archived artifact
func (b *FilesystemArchiveBackend) Restore(ctx context.Context, req *RestoreRequest) (*RestoreResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Get metadata
	metadata, err := b.readMetadata(req.ArchiveID)
	if err != nil {
		return nil, err
	}

	// Check if already restored
	if metadata.ArchiveStatus == ArchiveStatusRestored && metadata.RestoreExpiry != nil {
		if time.Now().Before(*metadata.RestoreExpiry) {
			// Already restored and not expired
			return &RestoreResponse{
				ArchiveMetadata:      metadata,
				EstimatedRestoreTime: time.Now(),
				RestoreExpiry:        *metadata.RestoreExpiry,
			}, nil
		}
	}

	// Simulate restore delay for testing
	if b.config.SimulateRestoreDelay > 0 {
		time.Sleep(b.config.SimulateRestoreDelay)
	}

	// Update metadata
	now := time.Now()
	expiry := now.Add(time.Duration(req.RestoreDurationSeconds) * time.Second)
	metadata.ArchiveStatus = ArchiveStatusRestored
	metadata.RestoredAt = &now
	metadata.RestoreExpiry = &expiry
	metadata.AccessCount++

	if err := b.writeMetadata(req.ArchiveID, metadata); err != nil {
		return nil, fmt.Errorf("failed to update metadata: %w", err)
	}

	return &RestoreResponse{
		ArchiveMetadata:      metadata,
		EstimatedRestoreTime: now,
		RestoreExpiry:        expiry,
		RestoreJobID:         req.ArchiveID,
	}, nil
}

// GetRestoreStatus checks the status of a restore operation
func (b *FilesystemArchiveBackend) GetRestoreStatus(ctx context.Context, archiveID string) (*RestoreStatus, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	metadata, err := b.readMetadata(archiveID)
	if err != nil {
		return nil, err
	}

	status := &RestoreStatus{
		ArchiveID:    archiveID,
		Status:       metadata.ArchiveStatus,
		Progress:     100, // Filesystem backend restores instantly
		RestoreJobID: archiveID,
	}

	if metadata.ArchiveStatus == ArchiveStatusRestored {
		status.EstimatedCompletionTime = metadata.RestoredAt
	}

	return status, nil
}

// GetArchived retrieves a restored artifact's data
func (b *FilesystemArchiveBackend) GetArchived(ctx context.Context, archiveID string) (*GetArchivedResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Get metadata
	metadata, err := b.readMetadata(archiveID)
	if err != nil {
		return nil, err
	}

	// Check if restored and not expired
	if metadata.ArchiveStatus != ArchiveStatusRestored {
		return nil, ErrInvalidState.Wrapf("archive not restored: status=%s", metadata.ArchiveStatus)
	}
	if metadata.RestoreExpiry != nil && time.Now().After(*metadata.RestoreExpiry) {
		return nil, ErrInvalidState.Wrap("restored copy has expired")
	}

	// Read archive data
	archivePath := b.getArchivePath(archiveID)
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive data: %w", err)
	}

	// Verify integrity
	checksum := sha256.Sum256(data)
	if hex.EncodeToString(checksum[:]) != hex.EncodeToString(metadata.IntegrityChecksum) {
		return nil, ErrHashMismatch.Wrap("integrity check failed")
	}

	// Update last accessed time
	now := time.Now()
	metadata.LastAccessedAt = &now
	metadata.AccessCount++
	if err := b.writeMetadata(archiveID, metadata); err != nil {
		// Log error but don't fail the operation
		// The data was successfully retrieved
		_ = err // intentionally ignored: metadata update is best-effort
	}

	return &GetArchivedResponse{
		Data:            data,
		ArchiveMetadata: metadata,
		RestoredAt:      *metadata.RestoredAt,
		RestoreExpiry:   *metadata.RestoreExpiry,
	}, nil
}

// DeleteArchive permanently deletes an archived artifact
func (b *FilesystemArchiveBackend) DeleteArchive(ctx context.Context, req *DeleteArchiveRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Get metadata
	metadata, err := b.readMetadata(req.ArchiveID)
	if err != nil {
		return err
	}

	// Check if can delete (compliance checks)
	if !req.Force && !metadata.CanDelete(time.Now()) {
		return ErrInvalidState.Wrap("archive cannot be deleted due to retention policy or legal hold")
	}

	// Delete archive file
	archivePath := b.getArchivePath(req.ArchiveID)
	if err := os.Remove(archivePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete archive file: %w", err)
	}

	// Delete metadata
	metadataPath := b.getMetadataPath(req.ArchiveID)
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// UpdateArchiveMetadata updates archive metadata
func (b *FilesystemArchiveBackend) UpdateArchiveMetadata(ctx context.Context, archiveID string, updates *ArchiveMetadataUpdate) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get existing metadata
	metadata, err := b.readMetadata(archiveID)
	if err != nil {
		return err
	}

	// Apply updates
	if updates.RetentionTag != nil {
		metadata.RetentionTag = updates.RetentionTag
	}
	if updates.ComplianceFlags != nil {
		metadata.ComplianceFlags = updates.ComplianceFlags
	}
	if updates.Notes != nil {
		metadata.Notes = *updates.Notes
	}

	// Write updated metadata
	return b.writeMetadata(archiveID, metadata)
}

// ListArchives lists archives by owner
func (b *FilesystemArchiveBackend) ListArchives(ctx context.Context, owner string, pagination *Pagination) (*ListArchivesResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var archives []*ArchiveMetadata

	// Walk metadata directory
	err := filepath.Walk(b.metadataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		// Read metadata
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip on error
		}

		var metadata ArchiveMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil // Skip on error
		}

		// Filter by owner
		if owner != "" && metadata.OriginalOwner != owner {
			return nil
		}

		archives = append(archives, &metadata)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list archives: %w", err)
	}

	// Apply pagination
	total := uint64(len(archives))
	if pagination != nil {
		start := pagination.Offset
		end := pagination.Offset + pagination.Limit
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		archives = archives[start:end]
	}

	hasMore := false
	if pagination != nil {
		hasMore = pagination.Offset+pagination.Limit < total
	}

	return &ListArchivesResponse{
		Archives: archives,
		Total:    total,
		HasMore:  hasMore,
	}, nil
}

// PurgeExpiredArchives removes expired archives
func (b *FilesystemArchiveBackend) PurgeExpiredArchives(ctx context.Context, currentTime int64) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Unix(currentTime, 0)
	purged := 0

	// Walk metadata directory
	err := filepath.Walk(b.metadataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		// Read metadata
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip on error
		}

		var metadata ArchiveMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil // Skip on error
		}

		// Check if can delete
		if metadata.CanDelete(now) {
			// Delete archive file
			archivePath := b.getArchivePath(metadata.ArchiveID)
			_ = os.Remove(archivePath)

			// Delete metadata
			_ = os.Remove(path)

			purged++
		}

		return nil
	})

	if err != nil {
		return purged, fmt.Errorf("failed to purge archives: %w", err)
	}

	return purged, nil
}

// VerifyArchiveIntegrity verifies the integrity of an archive
func (b *FilesystemArchiveBackend) VerifyArchiveIntegrity(ctx context.Context, archiveID string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Get metadata
	metadata, err := b.readMetadata(archiveID)
	if err != nil {
		return err
	}

	// Read archive data
	archivePath := b.getArchivePath(archiveID)
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return fmt.Errorf("failed to read archive data: %w", err)
	}

	// Compute checksum
	checksum := sha256.Sum256(data)

	// Compare checksums
	if hex.EncodeToString(checksum[:]) != hex.EncodeToString(metadata.IntegrityChecksum) {
		return ErrHashMismatch.Wrap("integrity verification failed")
	}

	return nil
}

// GetArchiveMetrics returns archive metrics
func (b *FilesystemArchiveBackend) GetArchiveMetrics(ctx context.Context) (*ArchiveMetrics, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	metrics := &ArchiveMetrics{
		BackendType:      "filesystem",
		ArchivesByTier:   make(map[ArchiveTier]uint64),
		ArchivesByStatus: make(map[ArchiveStatus]uint64),
		BackendStatus:    make(map[string]string),
	}

	// Walk metadata directory
	err := filepath.Walk(b.metadataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		// Read metadata
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip on error
		}

		var metadata ArchiveMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil // Skip on error
		}

		metrics.TotalArchives++
		metrics.ArchivesByTier[metadata.ArchiveTier]++
		metrics.ArchivesByStatus[metadata.ArchiveStatus]++

		if metadata.IsExpired(time.Now()) {
			metrics.ExpiredArchives++
		}
		if metadata.ComplianceFlags != nil && metadata.ComplianceFlags.LegalHold {
			metrics.LegalHoldArchives++
		}

		// Get file size
		archivePath := b.getArchivePath(metadata.ArchiveID)
		if stat, err := os.Stat(archivePath); err == nil {
			metrics.TotalBytes += uint64(stat.Size())
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	metrics.BackendStatus["archive_dir"] = b.archiveDir
	metrics.BackendStatus["metadata_dir"] = b.metadataDir

	return metrics, nil
}

// Health checks if the backend is healthy
func (b *FilesystemArchiveBackend) Health(ctx context.Context) error {
	// Check if directories are accessible
	if _, err := os.Stat(b.archiveDir); err != nil {
		return fmt.Errorf("archive directory not accessible: %w", err)
	}
	if _, err := os.Stat(b.metadataDir); err != nil {
		return fmt.Errorf("metadata directory not accessible: %w", err)
	}
	return nil
}

// Backend returns the backend type
func (b *FilesystemArchiveBackend) Backend() string {
	return "filesystem"
}

// Helper functions

func (b *FilesystemArchiveBackend) getArchivePath(archiveID string) string {
	// Use first 2 chars for subdirectory to avoid too many files in one dir
	subdir := archiveID[:2]
	return filepath.Join(b.archiveDir, subdir, archiveID+".dat")
}

func (b *FilesystemArchiveBackend) getMetadataPath(archiveID string) string {
	subdir := archiveID[:2]
	return filepath.Join(b.metadataDir, subdir, archiveID+".json")
}

func (b *FilesystemArchiveBackend) readMetadata(archiveID string) (*ArchiveMetadata, error) {
	metadataPath := b.getMetadataPath(archiveID)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound.Wrapf("archive not found: %s", archiveID)
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata ArchiveMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

func (b *FilesystemArchiveBackend) writeMetadata(archiveID string, metadata *ArchiveMetadata) error {
	metadataPath := b.getMetadataPath(archiveID)

	// Create subdirectory
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0700); err != nil {
		return fmt.Errorf("failed to create metadata subdirectory: %w", err)
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func generateArchiveID(contentAddr *ContentAddress, owner string) string {
	// Generate deterministic archive ID from content hash and owner
	data := fmt.Sprintf("%s:%s", contentAddr.HashHex(), owner)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
