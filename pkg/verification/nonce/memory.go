// Package nonce provides replay protection storage for verification attestations.
package nonce

import (
	"context"
	"crypto/rand"
	"sync"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// MemoryStore implements NonceStore using in-memory storage.
type MemoryStore struct {
	mu             sync.RWMutex
	nonces         map[string]*veidtypes.NonceRecord
	noncesByIssuer map[string][]string
	config         StoreConfig
	lastCleanup    time.Time
	cleanedUpCount int64
	closed         bool
}

// NewMemoryStore creates a new in-memory nonce store.
func NewMemoryStore(config StoreConfig) (*MemoryStore, error) {
	if config.Memory == nil {
		config.Memory = &MemoryConfig{
			MaxNonces:        100000,
			CleanupBatchSize: veidtypes.NonceCleanupBatchSize,
		}
	}

	store := &MemoryStore{
		nonces:         make(map[string]*veidtypes.NonceRecord),
		noncesByIssuer: make(map[string][]string),
		config:         config,
		lastCleanup:    time.Now(),
	}

	return store, nil
}

// CreateNonce creates and stores a new nonce.
func (m *MemoryStore) CreateNonce(ctx context.Context, req CreateNonceRequest) (*veidtypes.NonceRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrStoreClosed
	}

	// Check limits
	if len(m.nonces) >= m.config.Memory.MaxNonces {
		return nil, ErrStoreFull
	}

	if issuerNonces := m.noncesByIssuer[req.IssuerFingerprint]; len(issuerNonces) >= m.config.MaxNoncesPerIssuer {
		return nil, ErrTooManyNonces.Wrapf("issuer: %s...", req.IssuerFingerprint[:16])
	}

	// Generate random nonce
	nonce := make([]byte, veidtypes.NonceDefaultLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, ErrNonceGeneration.Wrapf("failed to generate random nonce: %v", err)
	}

	// Determine window
	windowSeconds := m.config.Policy.NonceWindowSeconds
	if req.WindowSeconds > 0 {
		windowSeconds = req.WindowSeconds
	}

	// Create record
	now := time.Now()
	record := veidtypes.NewNonceRecord(
		nonce,
		req.IssuerFingerprint,
		req.AttestationType,
		now,
		windowSeconds,
	)
	record.SubjectAddress = req.SubjectAddress

	// Store
	m.nonces[record.NonceHash] = record
	m.noncesByIssuer[req.IssuerFingerprint] = append(
		m.noncesByIssuer[req.IssuerFingerprint],
		record.NonceHash,
	)

	return record, nil
}

// ValidateAndUse validates a nonce and marks it as used atomically.
func (m *MemoryStore) ValidateAndUse(ctx context.Context, req ValidateNonceRequest) (*ValidateNonceResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrStoreClosed
	}

	// Validate nonce format
	if err := veidtypes.ValidateNonce(req.Nonce); err != nil {
		return &ValidateNonceResult{
			Valid:     false,
			Error:     err.Error(),
			ErrorCode: "invalid_format",
		}, nil
	}

	// Compute hash
	nonceHash := veidtypes.ComputeNonceHash(req.Nonce)

	result := &ValidateNonceResult{
		NonceHash: nonceHash,
	}

	// Look up record
	record, exists := m.nonces[nonceHash]
	if !exists {
		// Nonce not found - could be new or unknown
		// For verification, we typically expect the nonce to be pre-registered
		// But we can also accept new nonces and create records
		result.Valid = false
		result.Error = "nonce not found"
		result.ErrorCode = "not_found"
		return result, nil
	}

	result.Record = record

	// Check if already used
	if record.IsUsed() {
		result.Valid = false
		result.Error = "nonce already used"
		result.ErrorCode = "already_used"
		return result, nil
	}

	// Check if expired
	now := time.Now()
	if record.IsExpired(now) {
		result.Valid = false
		result.Error = "nonce expired"
		result.ErrorCode = "expired"
		return result, nil
	}

	// Check issuer binding
	if m.config.Policy.RequireIssuerBinding && record.IssuerFingerprint != req.IssuerFingerprint {
		result.Valid = false
		result.Error = "issuer mismatch"
		result.ErrorCode = "issuer_mismatch"
		return result, nil
	}

	// Check subject binding (if required)
	if m.config.Policy.RequireSubjectBinding && record.SubjectAddress != "" && record.SubjectAddress != req.SubjectAddress {
		result.Valid = false
		result.Error = "subject mismatch"
		result.ErrorCode = "subject_mismatch"
		return result, nil
	}

	// Validate timestamp
	if err := veidtypes.ValidateTimestamp(req.Timestamp, now, m.config.Policy); err != nil {
		result.Valid = false
		result.Error = err.Error()
		result.ErrorCode = "invalid_timestamp"
		return result, nil
	}

	// Mark as used if requested
	if req.MarkAsUsed {
		if err := record.MarkUsed(now, req.AttestationID, req.BlockHeight); err != nil {
			result.Valid = false
			result.Error = err.Error()
			result.ErrorCode = "mark_used_failed"
			return result, nil
		}
	}

	result.Valid = true
	return result, nil
}

// GetNonce retrieves a nonce record by hash.
func (m *MemoryStore) GetNonce(ctx context.Context, nonceHash string) (*veidtypes.NonceRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStoreClosed
	}

	record, exists := m.nonces[nonceHash]
	if !exists {
		return nil, ErrNonceNotFound.Wrapf("hash: %s...", nonceHash[:16])
	}

	// Return a copy
	recordCopy := *record
	return &recordCopy, nil
}

// MarkUsed marks a nonce as used.
func (m *MemoryStore) MarkUsed(ctx context.Context, nonceHash string, attestationID string, blockHeight int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStoreClosed
	}

	record, exists := m.nonces[nonceHash]
	if !exists {
		return ErrNonceNotFound.Wrapf("hash: %s...", nonceHash[:16])
	}

	return record.MarkUsed(time.Now(), attestationID, blockHeight)
}

// MarkExpired marks a nonce as expired.
func (m *MemoryStore) MarkExpired(ctx context.Context, nonceHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrStoreClosed
	}

	record, exists := m.nonces[nonceHash]
	if !exists {
		return ErrNonceNotFound.Wrapf("hash: %s...", nonceHash[:16])
	}

	record.MarkExpired()
	return nil
}

// CleanupExpired removes expired nonce records.
func (m *MemoryStore) CleanupExpired(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, ErrStoreClosed
	}

	now := time.Now()
	var cleaned int64
	toDelete := make([]string, 0)

	// Find expired nonces
	for hash, record := range m.nonces {
		if record.IsExpired(now) && !m.config.Policy.TrackNonceHistory {
			toDelete = append(toDelete, hash)
			if len(toDelete) >= m.config.Memory.CleanupBatchSize {
				break
			}
		}
	}

	// Delete expired nonces
	for _, hash := range toDelete {
		record := m.nonces[hash]
		delete(m.nonces, hash)

		// Remove from issuer index
		issuerNonces := m.noncesByIssuer[record.IssuerFingerprint]
		for i, h := range issuerNonces {
			if h == hash {
				m.noncesByIssuer[record.IssuerFingerprint] = append(
					issuerNonces[:i],
					issuerNonces[i+1:]...,
				)
				break
			}
		}

		cleaned++
	}

	m.lastCleanup = now
	m.cleanedUpCount = cleaned

	return cleaned, nil
}

// GetStats returns nonce store statistics.
func (m *MemoryStore) GetStats(ctx context.Context) (*NonceStoreStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, ErrStoreClosed
	}

	stats := &NonceStoreStats{
		TotalNonces:    int64(len(m.nonces)),
		NoncesByIssuer: make(map[string]int64),
		LastCleanupAt:  &m.lastCleanup,
		CleanedUpCount: m.cleanedUpCount,
	}

	now := time.Now()
	for _, record := range m.nonces {
		if record.IsUsed() {
			stats.UsedNonces++
		} else if record.IsExpired(now) {
			stats.ExpiredNonces++
		} else {
			stats.PendingNonces++
		}
	}

	for issuer, hashes := range m.noncesByIssuer {
		// Truncate issuer fingerprint for display
		displayIssuer := issuer
		if len(displayIssuer) > 16 {
			displayIssuer = displayIssuer[:16] + "..."
		}
		stats.NoncesByIssuer[displayIssuer] = int64(len(hashes))
	}

	return stats, nil
}

// HealthCheck verifies the store is accessible.
func (m *MemoryStore) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrStoreClosed
	}

	return nil
}

// Close closes the nonce store.
func (m *MemoryStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	m.nonces = nil
	m.noncesByIssuer = nil

	return nil
}

// Ensure MemoryStore implements NonceStore
var _ NonceStore = (*MemoryStore)(nil)
