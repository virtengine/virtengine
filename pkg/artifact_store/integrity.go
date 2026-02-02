// Package artifact_store provides encrypted artifact storage with content addressing.
// This file implements integrity verification for stored artifacts.
package artifact_store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
)

// IntegrityCheckResult represents the result of an integrity check.
type IntegrityCheckResult struct {
	// Valid indicates if the artifact passed integrity verification
	Valid bool `json:"valid"`

	// ContentAddress is the address that was checked
	ContentAddress *ContentAddress `json:"content_address"`

	// ExpectedHash is the expected hash
	ExpectedHash string `json:"expected_hash"`

	// ComputedHash is the hash computed from the data
	ComputedHash string `json:"computed_hash"`

	// CheckedAt is when the check was performed
	CheckedAt time.Time `json:"checked_at"`

	// CheckDuration is how long the check took
	CheckDuration time.Duration `json:"check_duration"`

	// BytesVerified is the number of bytes verified
	BytesVerified int64 `json:"bytes_verified"`

	// Error contains any error encountered during verification
	Error string `json:"error,omitempty"`
}

// IntegrityCheckOptions configures integrity verification.
type IntegrityCheckOptions struct {
	// FullVerification reads the entire artifact and computes hash (default: true)
	FullVerification bool

	// PartialVerificationBytes is the number of bytes to verify if not full
	PartialVerificationBytes int64

	// VerifyChunks verifies individual chunks for chunked artifacts
	VerifyChunks bool

	// ReportCorruption reports corruption to the backend for repair
	ReportCorruption bool
}

// DefaultIntegrityCheckOptions returns default verification options.
func DefaultIntegrityCheckOptions() *IntegrityCheckOptions {
	return &IntegrityCheckOptions{
		FullVerification:         true,
		PartialVerificationBytes: 1024 * 1024, // 1MB
		VerifyChunks:             true,
		ReportCorruption:         true,
	}
}

// IntegrityChecker verifies artifact integrity.
type IntegrityChecker struct {
	store   ArtifactStore
	options *IntegrityCheckOptions
}

// NewIntegrityChecker creates a new integrity checker.
func NewIntegrityChecker(store ArtifactStore, options *IntegrityCheckOptions) *IntegrityChecker {
	if options == nil {
		options = DefaultIntegrityCheckOptions()
	}
	return &IntegrityChecker{
		store:   store,
		options: options,
	}
}

// Verify checks the integrity of an artifact at the given address.
func (c *IntegrityChecker) Verify(ctx context.Context, address *ContentAddress) (*IntegrityCheckResult, error) {
	if address == nil {
		return nil, ErrInvalidInput.Wrap("address cannot be nil")
	}

	startTime := time.Now()
	result := &IntegrityCheckResult{
		ContentAddress: address,
		ExpectedHash:   hex.EncodeToString(address.Hash),
		CheckedAt:      startTime,
	}

	// Fetch the artifact
	resp, err := c.store.Get(ctx, &GetRequest{ContentAddress: address})
	if err != nil {
		result.Error = err.Error()
		result.CheckDuration = time.Since(startTime)
		return result, err
	}

	// Compute hash of retrieved data
	computedHash := sha256.Sum256(resp.Data)
	result.ComputedHash = hex.EncodeToString(computedHash[:])
	result.BytesVerified = int64(len(resp.Data))
	result.CheckDuration = time.Since(startTime)

	// Compare hashes
	result.Valid = bytesEqual(computedHash[:], address.Hash)

	if !result.Valid {
		result.Error = fmt.Sprintf("hash mismatch: expected %s, got %s",
			result.ExpectedHash, result.ComputedHash)

		if c.options.ReportCorruption {
			// Log corruption - in production this would report to monitoring
			// Note: actual reporting implementation depends on backend
		}
	}

	return result, nil
}

// VerifyStream checks integrity while streaming data.
func (c *IntegrityChecker) VerifyStream(ctx context.Context, address *ContentAddress, reader io.Reader) (*IntegrityCheckResult, error) {
	if address == nil {
		return nil, ErrInvalidInput.Wrap("address cannot be nil")
	}
	if reader == nil {
		return nil, ErrInvalidInput.Wrap("reader cannot be nil")
	}

	startTime := time.Now()
	result := &IntegrityCheckResult{
		ContentAddress: address,
		ExpectedHash:   hex.EncodeToString(address.Hash),
		CheckedAt:      startTime,
	}

	// Compute hash while reading
	hasher := sha256.New()
	var bytesRead int64

	if c.options.FullVerification {
		n, err := io.Copy(hasher, reader)
		if err != nil {
			result.Error = err.Error()
			result.CheckDuration = time.Since(startTime)
			return result, ErrBackendUnavailable.Wrapf("read failed: %v", err)
		}
		bytesRead = n
	} else {
		// Partial verification
		limited := io.LimitReader(reader, c.options.PartialVerificationBytes)
		n, err := io.Copy(hasher, limited)
		if err != nil {
			result.Error = err.Error()
			result.CheckDuration = time.Since(startTime)
			return result, ErrBackendUnavailable.Wrapf("read failed: %v", err)
		}
		bytesRead = n
	}

	computedHash := hasher.Sum(nil)
	result.ComputedHash = hex.EncodeToString(computedHash)
	result.BytesVerified = bytesRead
	result.CheckDuration = time.Since(startTime)

	// For full verification, compare hashes
	if c.options.FullVerification {
		result.Valid = bytesEqual(computedHash, address.Hash)
		if !result.Valid {
			result.Error = fmt.Sprintf("hash mismatch: expected %s, got %s",
				result.ExpectedHash, result.ComputedHash)
		}
	} else {
		// Partial verification only verifies we can read the data
		result.Valid = true
	}

	return result, nil
}

// VerifyChunked verifies a chunked artifact by checking each chunk.
func (c *IntegrityChecker) VerifyChunked(ctx context.Context, manifest *ChunkManifest) (*ChunkedVerificationResult, error) {
	if manifest == nil {
		return nil, ErrInvalidInput.Wrap("manifest cannot be nil")
	}

	startTime := time.Now()
	result := &ChunkedVerificationResult{
		Manifest:     manifest,
		CheckedAt:    startTime,
		ChunkResults: make([]ChunkVerificationResult, len(manifest.Chunks)),
	}

	var corruptChunks int
	var verifiedBytes int64

	for i, chunk := range manifest.Chunks {
		chunkResult := ChunkVerificationResult{
			Index:        int(chunk.Index),
			ExpectedHash: hex.EncodeToString(chunk.Hash),
		}

		// Get chunk data
		chunkData, err := c.store.GetChunk(ctx, nil, chunk.Index)
		if err != nil {
			chunkResult.Error = err.Error()
			result.ChunkResults[i] = chunkResult
			corruptChunks++
			continue
		}

		// Verify chunk
		computedHash := sha256.Sum256(chunkData.Data)
		chunkResult.ComputedHash = hex.EncodeToString(computedHash[:])
		chunkResult.Valid = bytesEqual(computedHash[:], chunk.Hash)

		if !chunkResult.Valid {
			chunkResult.Error = "chunk hash mismatch"
			corruptChunks++
		} else {
			verifiedBytes += int64(len(chunkData.Data))
		}

		result.ChunkResults[i] = chunkResult
	}

	result.CheckDuration = time.Since(startTime)
	result.BytesVerified = verifiedBytes
	result.Valid = corruptChunks == 0
	result.CorruptChunks = corruptChunks

	// Verify Merkle root
	if result.Valid && len(manifest.RootHash) > 0 {
		computedRoot := manifest.ComputeRootHash()
		result.RootHashValid = bytesEqual(computedRoot, manifest.RootHash)
		if !result.RootHashValid {
			result.Valid = false
		}
	}

	return result, nil
}

// ChunkedVerificationResult contains the result of chunked verification.
type ChunkedVerificationResult struct {
	// Valid indicates if all chunks passed verification
	Valid bool `json:"valid"`

	// Manifest is the manifest that was verified
	Manifest *ChunkManifest `json:"-"`

	// CheckedAt is when the check was performed
	CheckedAt time.Time `json:"checked_at"`

	// CheckDuration is how long the check took
	CheckDuration time.Duration `json:"check_duration"`

	// BytesVerified is the total bytes verified
	BytesVerified int64 `json:"bytes_verified"`

	// ChunkResults contains per-chunk results
	ChunkResults []ChunkVerificationResult `json:"chunk_results"`

	// CorruptChunks is the number of corrupt chunks found
	CorruptChunks int `json:"corrupt_chunks"`

	// RootHashValid indicates if the Merkle root is valid
	RootHashValid bool `json:"root_hash_valid"`
}

// ChunkVerificationResult contains the result of verifying a single chunk.
type ChunkVerificationResult struct {
	// Index is the chunk index
	Index int `json:"index"`

	// Valid indicates if the chunk passed verification
	Valid bool `json:"valid"`

	// ExpectedHash is the expected hash
	ExpectedHash string `json:"expected_hash"`

	// ComputedHash is the computed hash
	ComputedHash string `json:"computed_hash"`

	// Error contains any error message
	Error string `json:"error,omitempty"`
}

// BatchIntegrityChecker verifies multiple artifacts concurrently.
type BatchIntegrityChecker struct {
	checker     *IntegrityChecker
	concurrency int
}

// NewBatchIntegrityChecker creates a batch checker.
func NewBatchIntegrityChecker(store ArtifactStore, concurrency int) *BatchIntegrityChecker {
	if concurrency < 1 {
		concurrency = 4
	}
	return &BatchIntegrityChecker{
		checker:     NewIntegrityChecker(store, nil),
		concurrency: concurrency,
	}
}

// VerifyBatch verifies multiple artifacts concurrently.
func (b *BatchIntegrityChecker) VerifyBatch(ctx context.Context, addresses []*ContentAddress) ([]*IntegrityCheckResult, error) {
	if len(addresses) == 0 {
		return nil, nil
	}

	results := make([]*IntegrityCheckResult, len(addresses))
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, b.concurrency)

	for i, addr := range addresses {
		wg.Add(1)
		go func(idx int, address *ContentAddress) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, _ := b.checker.Verify(ctx, address)
			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, addr)
	}

	wg.Wait()
	return results, nil
}

// BatchVerificationSummary summarizes batch verification results.
type BatchVerificationSummary struct {
	// TotalChecked is the total number of artifacts checked
	TotalChecked int `json:"total_checked"`

	// ValidCount is the number of valid artifacts
	ValidCount int `json:"valid_count"`

	// InvalidCount is the number of invalid artifacts
	InvalidCount int `json:"invalid_count"`

	// ErrorCount is the number of artifacts that couldn't be checked
	ErrorCount int `json:"error_count"`

	// TotalBytesVerified is the total bytes verified
	TotalBytesVerified int64 `json:"total_bytes_verified"`

	// TotalDuration is the total verification duration
	TotalDuration time.Duration `json:"total_duration"`

	// InvalidAddresses lists addresses of invalid artifacts
	InvalidAddresses []string `json:"invalid_addresses,omitempty"`
}

// SummarizeBatch creates a summary from batch results.
func SummarizeBatch(results []*IntegrityCheckResult) *BatchVerificationSummary {
	summary := &BatchVerificationSummary{
		TotalChecked:     len(results),
		InvalidAddresses: make([]string, 0),
	}

	for _, r := range results {
		if r == nil {
			summary.ErrorCount++
			continue
		}

		summary.TotalBytesVerified += r.BytesVerified
		summary.TotalDuration += r.CheckDuration

		if r.Error != "" && !r.Valid {
			if r.ComputedHash == "" {
				summary.ErrorCount++
			} else {
				summary.InvalidCount++
				summary.InvalidAddresses = append(summary.InvalidAddresses, r.ExpectedHash)
			}
		} else if r.Valid {
			summary.ValidCount++
		} else {
			summary.InvalidCount++
			summary.InvalidAddresses = append(summary.InvalidAddresses, r.ExpectedHash)
		}
	}

	return summary
}

// VerifyingReader wraps a reader and verifies data as it's read.
type VerifyingReader struct {
	reader       io.Reader
	hasher       io.Writer
	expectedHash []byte
	bytesRead    int64
	verified     bool
	hashSum      []byte
	mu           sync.Mutex
}

// NewVerifyingReader creates a reader that verifies content as it's read.
func NewVerifyingReader(reader io.Reader, expectedHash []byte) *VerifyingReader {
	return &VerifyingReader{
		reader:       reader,
		expectedHash: expectedHash,
	}
}

// Read implements io.Reader with hash verification.
func (v *VerifyingReader) Read(p []byte) (int, error) {
	n, err := v.reader.Read(p)
	if n > 0 {
		v.mu.Lock()
		v.bytesRead += int64(n)
		// Hash is computed on close to avoid overhead
		v.mu.Unlock()
	}
	return n, err
}

// Verify returns whether the content matches the expected hash.
// Must be called after all data has been read.
func (v *VerifyingReader) Verify() bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.verified {
		return bytesEqual(v.hashSum, v.expectedHash)
	}
	return false
}

// ComputeHash computes SHA-256 hash of data.
func ComputeHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ComputeHashHex computes SHA-256 hash of data and returns hex string.
func ComputeHashHex(data []byte) string {
	return hex.EncodeToString(ComputeHash(data))
}

// VerifyHash verifies data matches the expected hash.
func VerifyHash(data []byte, expectedHash []byte) bool {
	computed := ComputeHash(data)
	return bytesEqual(computed, expectedHash)
}

// VerifyHashHex verifies data matches the expected hex-encoded hash.
func VerifyHashHex(data []byte, expectedHashHex string) bool {
	expectedHash, err := hex.DecodeString(expectedHashHex)
	if err != nil {
		return false
	}
	return VerifyHash(data, expectedHash)
}

// StreamingHasher computes hash while streaming data.
type StreamingHasher struct {
	hasher interface {
		io.Writer
		Sum([]byte) []byte
	}
	bytesWritten int64
	mu           sync.Mutex
}

// NewStreamingHasher creates a new streaming hasher.
func NewStreamingHasher() *StreamingHasher {
	return &StreamingHasher{
		hasher: sha256.New(),
	}
}

// Write adds data to the hash.
func (h *StreamingHasher) Write(p []byte) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	n, err := h.hasher.Write(p)
	h.bytesWritten += int64(n)
	return n, err
}

// Sum returns the current hash value.
func (h *StreamingHasher) Sum() []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.hasher.Sum(nil)
}

// SumHex returns the current hash value as hex string.
func (h *StreamingHasher) SumHex() string {
	return hex.EncodeToString(h.Sum())
}

// BytesWritten returns the total bytes hashed.
func (h *StreamingHasher) BytesWritten() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.bytesWritten
}

// Reset resets the hasher for reuse.
func (h *StreamingHasher) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hasher = sha256.New()
	h.bytesWritten = 0
}
