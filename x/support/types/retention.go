package types

import (
	"fmt"
	"time"
)

// RetentionPolicyVersion is the current retention policy version.
const RetentionPolicyVersion uint32 = 1

// RetentionPolicy defines retention and archival rules for support tickets.
type RetentionPolicy struct {
	// Version is the retention policy version
	Version uint32 `json:"version"`

	// ArchiveAfterSeconds is the number of seconds after creation to archive
	ArchiveAfterSeconds int64 `json:"archive_after_seconds,omitempty"`

	// PurgeAfterSeconds is the number of seconds after creation to purge payload
	PurgeAfterSeconds int64 `json:"purge_after_seconds,omitempty"`

	// CreatedAt is the block time when the policy was attached
	CreatedAt time.Time `json:"created_at"`

	// CreatedAtBlock is the block height when the policy was attached
	CreatedAtBlock int64 `json:"created_at_block"`
}

// DefaultRetentionPolicy returns the default retention policy.
func DefaultRetentionPolicy(now time.Time, height int64) *RetentionPolicy {
	return &RetentionPolicy{
		Version:             RetentionPolicyVersion,
		ArchiveAfterSeconds: int64((90 * 24 * time.Hour).Seconds()),
		PurgeAfterSeconds:   int64((365 * 24 * time.Hour).Seconds()),
		CreatedAt:           now.UTC(),
		CreatedAtBlock:      height,
	}
}

// Validate validates the retention policy.
func (p *RetentionPolicy) Validate() error {
	if p == nil {
		return nil
	}

	if p.Version == 0 || p.Version > RetentionPolicyVersion {
		return ErrInvalidRetentionPolicy.Wrapf("unsupported version: %d", p.Version)
	}

	if p.ArchiveAfterSeconds < 0 {
		return ErrInvalidRetentionPolicy.Wrap("archive_after_seconds cannot be negative")
	}
	if p.PurgeAfterSeconds < 0 {
		return ErrInvalidRetentionPolicy.Wrap("purge_after_seconds cannot be negative")
	}
	if p.ArchiveAfterSeconds > 0 && p.PurgeAfterSeconds > 0 && p.PurgeAfterSeconds < p.ArchiveAfterSeconds {
		return ErrInvalidRetentionPolicy.Wrap("purge_after_seconds must be >= archive_after_seconds")
	}
	if p.CreatedAt.IsZero() && p.CreatedAtBlock != 0 {
		return ErrInvalidRetentionPolicy.Wrap("created_at is required when created_at_block is set")
	}

	return nil
}

// ArchiveAt returns the archive time if configured.
func (p *RetentionPolicy) ArchiveAt() (time.Time, bool) {
	if p == nil || p.ArchiveAfterSeconds == 0 || p.CreatedAt.IsZero() {
		return time.Time{}, false
	}
	return p.CreatedAt.Add(time.Duration(p.ArchiveAfterSeconds) * time.Second).UTC(), true
}

// PurgeAt returns the purge time if configured.
func (p *RetentionPolicy) PurgeAt() (time.Time, bool) {
	if p == nil || p.PurgeAfterSeconds == 0 || p.CreatedAt.IsZero() {
		return time.Time{}, false
	}
	return p.CreatedAt.Add(time.Duration(p.PurgeAfterSeconds) * time.Second).UTC(), true
}

// ShouldArchive checks if the policy indicates archival at the given time.
func (p *RetentionPolicy) ShouldArchive(now time.Time) bool {
	at, ok := p.ArchiveAt()
	if !ok {
		return false
	}
	return !now.Before(at)
}

// ShouldPurge checks if the policy indicates purge at the given time.
func (p *RetentionPolicy) ShouldPurge(now time.Time) bool {
	at, ok := p.PurgeAt()
	if !ok {
		return false
	}
	return !now.Before(at)
}

// CopyWithTimestamps returns a policy with ensured timestamps.
func (p *RetentionPolicy) CopyWithTimestamps(now time.Time, height int64) *RetentionPolicy {
	if p == nil {
		return nil
	}

	clone := *p
	if clone.Version == 0 {
		clone.Version = RetentionPolicyVersion
	}
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now.UTC()
	}
	if clone.CreatedAtBlock == 0 {
		clone.CreatedAtBlock = height
	}
	return &clone
}

// String returns the string representation.
func (p *RetentionPolicy) String() string {
	if p == nil {
		return "RetentionPolicy<nil>"
	}
	return fmt.Sprintf("RetentionPolicy{version=%d, archive_after=%ds, purge_after=%ds}",
		p.Version, p.ArchiveAfterSeconds, p.PurgeAfterSeconds)
}
