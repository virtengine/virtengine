// Package nonce provides replay protection storage for verification attestations.
//
// This package implements NonceStore for tracking and validating attestation nonces
// to prevent replay attacks.
//
// Task Reference: VE-2B - Verification Shared Infrastructure
package nonce

import (
	"context"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Nonce Store Interface
// ============================================================================

// NonceStore defines the interface for nonce storage and validation.
type NonceStore interface {
	// CreateNonce creates and stores a new nonce.
	CreateNonce(ctx context.Context, req CreateNonceRequest) (*veidtypes.NonceRecord, error)

	// ValidateAndUse validates a nonce and marks it as used atomically.
	ValidateAndUse(ctx context.Context, req ValidateNonceRequest) (*ValidateNonceResult, error)

	// GetNonce retrieves a nonce record by hash.
	GetNonce(ctx context.Context, nonceHash string) (*veidtypes.NonceRecord, error)

	// MarkUsed marks a nonce as used.
	MarkUsed(ctx context.Context, nonceHash string, attestationID string, blockHeight int64) error

	// MarkExpired marks a nonce as expired.
	MarkExpired(ctx context.Context, nonceHash string) error

	// CleanupExpired removes expired nonce records.
	CleanupExpired(ctx context.Context) (int64, error)

	// GetStats returns nonce store statistics.
	GetStats(ctx context.Context) (*NonceStoreStats, error)

	// HealthCheck verifies the store is accessible.
	HealthCheck(ctx context.Context) error

	// Close closes the nonce store.
	Close() error
}

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateNonceRequest contains parameters for creating a nonce.
type CreateNonceRequest struct {
	// IssuerFingerprint is the issuer key fingerprint
	IssuerFingerprint string `json:"issuer_fingerprint"`

	// SubjectAddress is the subject address (optional)
	SubjectAddress string `json:"subject_address,omitempty"`

	// AttestationType is the attestation type
	AttestationType veidtypes.AttestationType `json:"attestation_type"`

	// WindowSeconds is the validity window (uses default if 0)
	WindowSeconds int64 `json:"window_seconds,omitempty"`
}

// ValidateNonceRequest contains parameters for validating a nonce.
type ValidateNonceRequest struct {
	// Nonce is the raw nonce bytes
	Nonce []byte `json:"nonce"`

	// IssuerFingerprint is the expected issuer fingerprint
	IssuerFingerprint string `json:"issuer_fingerprint"`

	// SubjectAddress is the expected subject address (optional)
	SubjectAddress string `json:"subject_address,omitempty"`

	// AttestationType is the expected attestation type
	AttestationType veidtypes.AttestationType `json:"attestation_type"`

	// Timestamp is the attestation timestamp
	Timestamp time.Time `json:"timestamp"`

	// AttestationID is the attestation ID (for marking as used)
	AttestationID string `json:"attestation_id"`

	// BlockHeight is the block height (for marking as used)
	BlockHeight int64 `json:"block_height,omitempty"`

	// MarkAsUsed indicates if the nonce should be marked as used
	MarkAsUsed bool `json:"mark_as_used"`
}

// ValidateNonceResult contains the result of nonce validation.
type ValidateNonceResult struct {
	// Valid indicates if the nonce is valid for use
	Valid bool `json:"valid"`

	// NonceHash is the computed nonce hash
	NonceHash string `json:"nonce_hash"`

	// Error contains the error if validation failed
	Error string `json:"error,omitempty"`

	// ErrorCode is the error code if validation failed
	ErrorCode string `json:"error_code,omitempty"`

	// Record is the nonce record (if found)
	Record *veidtypes.NonceRecord `json:"record,omitempty"`
}

// NonceStoreStats contains nonce store statistics.
type NonceStoreStats struct {
	// TotalNonces is the total number of tracked nonces
	TotalNonces int64 `json:"total_nonces"`

	// UsedNonces is the number of used nonces
	UsedNonces int64 `json:"used_nonces"`

	// ExpiredNonces is the number of expired nonces
	ExpiredNonces int64 `json:"expired_nonces"`

	// PendingNonces is the number of pending (unused) nonces
	PendingNonces int64 `json:"pending_nonces"`

	// NoncesByIssuer is the count by issuer
	NoncesByIssuer map[string]int64 `json:"nonces_by_issuer,omitempty"`

	// LastCleanupAt is when the last cleanup occurred
	LastCleanupAt *time.Time `json:"last_cleanup_at,omitempty"`

	// CleanedUpCount is the number of nonces cleaned up in last cleanup
	CleanedUpCount int64 `json:"cleaned_up_count,omitempty"`
}

// ============================================================================
// Configuration
// ============================================================================

// StoreConfig contains configuration for the nonce store.
type StoreConfig struct {
	// Backend specifies the storage backend
	Backend StoreBackend `json:"backend"`

	// Policy is the replay protection policy
	Policy veidtypes.ReplayProtectionPolicy `json:"policy"`

	// CleanupInterval is how often to run cleanup
	CleanupInterval time.Duration `json:"cleanup_interval"`

	// MaxNoncesPerIssuer is the maximum nonces to track per issuer
	MaxNoncesPerIssuer int `json:"max_nonces_per_issuer"`

	// Redis contains Redis backend configuration
	Redis *RedisConfig `json:"redis,omitempty"`

	// Memory contains memory backend configuration
	Memory *MemoryConfig `json:"memory,omitempty"`
}

// StoreBackend identifies the nonce store backend.
type StoreBackend string

const (
	BackendMemory StoreBackend = "memory"
	BackendRedis  StoreBackend = "redis"
)

// RedisConfig contains Redis backend configuration.
type RedisConfig struct {
	// URL is the Redis connection URL
	URL string `json:"url"`

	// Prefix is the key prefix
	Prefix string `json:"prefix"`

	// PoolSize is the connection pool size
	PoolSize int `json:"pool_size"`
}

// MemoryConfig contains memory backend configuration.
type MemoryConfig struct {
	// MaxNonces is the maximum number of nonces to store
	MaxNonces int `json:"max_nonces"`

	// CleanupBatchSize is the batch size for cleanup
	CleanupBatchSize int `json:"cleanup_batch_size"`
}

// DefaultStoreConfig returns the default nonce store configuration.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		Backend:            BackendMemory,
		Policy:             veidtypes.DefaultReplayProtectionPolicy(),
		CleanupInterval:    5 * time.Minute,
		MaxNoncesPerIssuer: veidtypes.DefaultMaxNoncesPerIssuer,
		Memory: &MemoryConfig{
			MaxNonces:        100000,
			CleanupBatchSize: veidtypes.NonceCleanupBatchSize,
		},
	}
}

// DefaultRedisConfig returns the default Redis configuration.
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		URL:      "redis://localhost:6379/0",
		Prefix:   "virtengine:nonce",
		PoolSize: 10,
	}
}

