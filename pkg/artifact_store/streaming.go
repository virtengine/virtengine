// Package artifact_store provides encrypted artifact storage with content addressing.
// This file implements streaming support for large artifacts.
package artifact_store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// StreamingConfig configures streaming behavior for large artifacts.
type StreamingConfig struct {
	// ChunkSize is the size of each chunk for streaming (default: 8MB)
	ChunkSize int64

	// BufferSize is the buffer size for I/O operations (default: 64KB)
	BufferSize int

	// MaxConcurrentChunks is the max concurrent chunks for parallel upload (default: 4)
	MaxConcurrentChunks int

	// ProgressInterval is how often to report progress (default: 1s)
	ProgressInterval time.Duration

	// VerifyOnStream enables hash verification during streaming (default: true)
	VerifyOnStream bool

	// RetryAttempts is the number of retry attempts for failed chunks (default: 3)
	RetryAttempts int

	// RetryDelay is the initial delay between retries (default: 1s)
	RetryDelay time.Duration
}

// DefaultStreamingConfig returns default streaming configuration.
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		ChunkSize:           8 * 1024 * 1024, // 8MB
		BufferSize:          64 * 1024,       // 64KB
		MaxConcurrentChunks: 4,
		ProgressInterval:    time.Second,
		VerifyOnStream:      true,
		RetryAttempts:       3,
		RetryDelay:          time.Second,
	}
}

// StreamProgress reports streaming progress.
type StreamProgress struct {
	// BytesTransferred is the total bytes transferred so far
	BytesTransferred int64

	// TotalBytes is the total expected bytes (-1 if unknown)
	TotalBytes int64

	// ChunksCompleted is the number of chunks completed
	ChunksCompleted int

	// TotalChunks is the total number of chunks (-1 if unknown)
	TotalChunks int

	// CurrentRate is the current transfer rate in bytes/second
	CurrentRate float64

	// EstimatedTimeRemaining is the estimated time to completion
	EstimatedTimeRemaining time.Duration

	// StartTime is when the transfer started
	StartTime time.Time

	// LastUpdateTime is when this progress was recorded
	LastUpdateTime time.Time
}

// PercentComplete returns the completion percentage (0-100).
func (p *StreamProgress) PercentComplete() float64 {
	if p.TotalBytes <= 0 {
		return 0
	}
	return float64(p.BytesTransferred) / float64(p.TotalBytes) * 100
}

// ProgressCallback is called during streaming to report progress.
type ProgressCallback func(progress *StreamProgress)

// StreamingUploader handles streaming uploads with progress tracking.
type StreamingUploader struct {
	config   *StreamingConfig
	backend  StreamingArtifactStore
	progress *StreamProgress
	callback ProgressCallback

	mu         sync.RWMutex
	started    time.Time
	lastUpdate time.Time
	cancelled  atomic.Bool
}

// NewStreamingUploader creates a new streaming uploader.
func NewStreamingUploader(backend StreamingArtifactStore, config *StreamingConfig) *StreamingUploader {
	if config == nil {
		config = DefaultStreamingConfig()
	}
	return &StreamingUploader{
		config:  config,
		backend: backend,
		progress: &StreamProgress{
			TotalBytes:  -1,
			TotalChunks: -1,
		},
	}
}

// SetProgressCallback sets the progress callback.
func (u *StreamingUploader) SetProgressCallback(cb ProgressCallback) {
	u.callback = cb
}

// Upload streams data from the reader to the backend.
func (u *StreamingUploader) Upload(ctx context.Context, req *PutStreamRequest) (*PutResponse, error) {
	if req == nil {
		return nil, ErrInvalidInput.Wrap("request cannot be nil")
	}
	if req.Reader == nil {
		return nil, ErrInvalidInput.Wrap("reader cannot be nil")
	}

	u.started = time.Now()
	u.progress.StartTime = u.started
	u.progress.TotalBytes = req.Size

	// Create a progress-tracking reader
	progressReader := &progressReader{
		reader:   req.Reader,
		uploader: u,
	}

	// Create the wrapped request with progress reader
	wrappedReq := &PutStreamRequest{
		Reader:             progressReader,
		Size:               req.Size,
		ContentHash:        req.ContentHash,
		EncryptionMetadata: req.EncryptionMetadata,
		RetentionTag:       req.RetentionTag,
		Owner:              req.Owner,
		ArtifactType:       req.ArtifactType,
		Metadata:           req.Metadata,
	}

	// Perform the streaming upload
	resp, err := u.backend.PutStream(ctx, wrappedReq)
	if err != nil {
		return nil, err
	}

	// Final progress update
	u.updateProgress(u.progress.BytesTransferred)

	return resp, nil
}

// Cancel cancels an in-progress upload.
func (u *StreamingUploader) Cancel() {
	u.cancelled.Store(true)
}

// updateProgress updates and reports progress.
func (u *StreamingUploader) updateProgress(bytesTransferred int64) {
	u.mu.Lock()
	defer u.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(u.started).Seconds()

	u.progress.BytesTransferred = bytesTransferred
	u.progress.LastUpdateTime = now

	// Calculate rate
	if elapsed > 0 {
		u.progress.CurrentRate = float64(bytesTransferred) / elapsed
	}

	// Estimate remaining time
	if u.progress.TotalBytes > 0 && u.progress.CurrentRate > 0 {
		remaining := u.progress.TotalBytes - bytesTransferred
		u.progress.EstimatedTimeRemaining = time.Duration(float64(remaining)/u.progress.CurrentRate) * time.Second
	}

	// Call callback if enough time has passed
	if u.callback != nil && now.Sub(u.lastUpdate) >= u.config.ProgressInterval {
		u.lastUpdate = now
		progressCopy := *u.progress
		go u.callback(&progressCopy)
	}
}

// progressReader wraps a reader to track progress.
type progressReader struct {
	reader   io.Reader
	uploader *StreamingUploader
	read     int64
}

func (p *progressReader) Read(buf []byte) (int, error) {
	if p.uploader.cancelled.Load() {
		return 0, context.Canceled
	}

	n, err := p.reader.Read(buf)
	if n > 0 {
		p.read += int64(n)
		p.uploader.updateProgress(p.read)
	}
	return n, err
}

// StreamingDownloader handles streaming downloads with verification.
type StreamingDownloader struct {
	config   *StreamingConfig
	backend  StreamingArtifactStore
	progress *StreamProgress
	callback ProgressCallback

	mu         sync.RWMutex
	started    time.Time
	lastUpdate time.Time
	cancelled  atomic.Bool
}

// NewStreamingDownloader creates a new streaming downloader.
func NewStreamingDownloader(backend StreamingArtifactStore, config *StreamingConfig) *StreamingDownloader {
	if config == nil {
		config = DefaultStreamingConfig()
	}
	return &StreamingDownloader{
		config:  config,
		backend: backend,
		progress: &StreamProgress{
			TotalBytes:  -1,
			TotalChunks: -1,
		},
	}
}

// SetProgressCallback sets the progress callback.
func (d *StreamingDownloader) SetProgressCallback(cb ProgressCallback) {
	d.callback = cb
}

// Download streams data from the backend to the writer.
func (d *StreamingDownloader) Download(ctx context.Context, address *ContentAddress, writer io.Writer) error {
	if address == nil {
		return ErrInvalidInput.Wrap("address cannot be nil")
	}
	if writer == nil {
		return ErrInvalidInput.Wrap("writer cannot be nil")
	}

	d.started = time.Now()
	d.progress.StartTime = d.started
	d.progress.TotalBytes = int64(address.Size)

	// Get the stream from backend
	stream, err := d.backend.GetStream(ctx, address)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	// Create verifying reader if configured
	var reader io.Reader = stream
	var hasher *hashingWriter
	if d.config.VerifyOnStream && len(address.Hash) > 0 {
		hasher = &hashingWriter{
			hash: sha256.New(),
		}
		reader = io.TeeReader(stream, hasher)
	}

	// Copy with progress tracking
	buf := make([]byte, d.config.BufferSize)
	var written int64

	for {
		if d.cancelled.Load() {
			return context.Canceled
		}

		n, readErr := reader.Read(buf)
		if n > 0 {
			nw, writeErr := writer.Write(buf[:n])
			if writeErr != nil {
				return ErrBackendUnavailable.Wrapf("write failed: %v", writeErr)
			}
			if nw != n {
				return ErrBackendUnavailable.Wrap("short write")
			}
			written += int64(nw)
			d.updateProgress(written)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return ErrBackendUnavailable.Wrapf("read failed: %v", readErr)
		}
	}

	// Verify hash if configured
	if hasher != nil {
		computedHash := hasher.Sum()
		if !bytesEqual(computedHash, address.Hash) {
			return ErrHashMismatch.Wrapf("expected %s, got %s",
				hex.EncodeToString(address.Hash),
				hex.EncodeToString(computedHash))
		}
	}

	// Final progress update
	d.updateProgress(written)

	return nil
}

// Cancel cancels an in-progress download.
func (d *StreamingDownloader) Cancel() {
	d.cancelled.Store(true)
}

// updateProgress updates and reports progress.
func (d *StreamingDownloader) updateProgress(bytesTransferred int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(d.started).Seconds()

	d.progress.BytesTransferred = bytesTransferred
	d.progress.LastUpdateTime = now

	// Calculate rate
	if elapsed > 0 {
		d.progress.CurrentRate = float64(bytesTransferred) / elapsed
	}

	// Estimate remaining time
	if d.progress.TotalBytes > 0 && d.progress.CurrentRate > 0 {
		remaining := d.progress.TotalBytes - bytesTransferred
		d.progress.EstimatedTimeRemaining = time.Duration(float64(remaining)/d.progress.CurrentRate) * time.Second
	}

	// Call callback if enough time has passed
	if d.callback != nil && now.Sub(d.lastUpdate) >= d.config.ProgressInterval {
		d.lastUpdate = now
		progressCopy := *d.progress
		go d.callback(&progressCopy)
	}
}

// hashingWriter computes hash while writing.
type hashingWriter struct {
	hash io.Writer
}

func (h *hashingWriter) Write(p []byte) (int, error) {
	return h.hash.Write(p)
}

func (h *hashingWriter) Sum() []byte {
	if hasher, ok := h.hash.(interface{ Sum([]byte) []byte }); ok {
		return hasher.Sum(nil)
	}
	return nil
}

// ChunkedTransferState represents the state of a chunked transfer.
type ChunkedTransferState struct {
	// TransferID is a unique identifier for this transfer
	TransferID string `json:"transfer_id"`

	// TotalSize is the total artifact size
	TotalSize int64 `json:"total_size"`

	// ChunkSize is the size of each chunk
	ChunkSize int64 `json:"chunk_size"`

	// CompletedChunks tracks which chunks are complete
	CompletedChunks map[int]bool `json:"completed_chunks"`

	// TotalChunks is the total number of chunks
	TotalChunks int `json:"total_chunks"`

	// ContentHash is the expected final hash
	ContentHash string `json:"content_hash,omitempty"`

	// StartedAt is when the transfer started
	StartedAt time.Time `json:"started_at"`

	// LastUpdatedAt is when the state was last updated
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// IsComplete returns true if all chunks are complete.
func (s *ChunkedTransferState) IsComplete() bool {
	if len(s.CompletedChunks) != s.TotalChunks {
		return false
	}
	for i := 0; i < s.TotalChunks; i++ {
		if !s.CompletedChunks[i] {
			return false
		}
	}
	return true
}

// PercentComplete returns the completion percentage.
func (s *ChunkedTransferState) PercentComplete() float64 {
	if s.TotalChunks == 0 {
		return 0
	}
	return float64(len(s.CompletedChunks)) / float64(s.TotalChunks) * 100
}

// NextIncompleteChunk returns the next incomplete chunk index, or -1 if all complete.
func (s *ChunkedTransferState) NextIncompleteChunk() int {
	for i := 0; i < s.TotalChunks; i++ {
		if !s.CompletedChunks[i] {
			return i
		}
	}
	return -1
}

// ResumableUploader supports resumable uploads with state persistence.
type ResumableUploader struct {
	config   *StreamingConfig
	backend  StreamingArtifactStore
	state    *ChunkedTransferState
	stateMu  sync.RWMutex
	callback ProgressCallback
}

// NewResumableUploader creates a new resumable uploader.
func NewResumableUploader(backend StreamingArtifactStore, config *StreamingConfig) *ResumableUploader {
	if config == nil {
		config = DefaultStreamingConfig()
	}
	return &ResumableUploader{
		config:  config,
		backend: backend,
	}
}

// StartUpload initiates a new resumable upload.
func (u *ResumableUploader) StartUpload(totalSize int64, contentHash string) *ChunkedTransferState {
	chunkSize := u.config.ChunkSize
	totalChunks := int((totalSize + chunkSize - 1) / chunkSize)

	u.state = &ChunkedTransferState{
		TransferID:      fmt.Sprintf("upload-%d", time.Now().UnixNano()),
		TotalSize:       totalSize,
		ChunkSize:       chunkSize,
		CompletedChunks: make(map[int]bool),
		TotalChunks:     totalChunks,
		ContentHash:     contentHash,
		StartedAt:       time.Now().UTC(),
		LastUpdatedAt:   time.Now().UTC(),
	}

	return u.state
}

// ResumeUpload resumes an existing upload from the given state.
func (u *ResumableUploader) ResumeUpload(state *ChunkedTransferState) {
	u.stateMu.Lock()
	defer u.stateMu.Unlock()
	u.state = state
}

// GetState returns the current transfer state.
func (u *ResumableUploader) GetState() *ChunkedTransferState {
	u.stateMu.RLock()
	defer u.stateMu.RUnlock()
	if u.state == nil {
		return nil
	}
	// Return a copy
	stateCopy := *u.state
	stateCopy.CompletedChunks = make(map[int]bool)
	for k, v := range u.state.CompletedChunks {
		stateCopy.CompletedChunks[k] = v
	}
	return &stateCopy
}

// MarkChunkComplete marks a chunk as complete.
func (u *ResumableUploader) MarkChunkComplete(chunkIndex int) {
	u.stateMu.Lock()
	defer u.stateMu.Unlock()
	if u.state != nil {
		u.state.CompletedChunks[chunkIndex] = true
		u.state.LastUpdatedAt = time.Now().UTC()
	}
}

// SetProgressCallback sets the progress callback.
func (u *ResumableUploader) SetProgressCallback(cb ProgressCallback) {
	u.callback = cb
}

// MultipartUploadConfig configures multipart uploads.
type MultipartUploadConfig struct {
	// PartSize is the size of each part (minimum 5MB for most backends)
	PartSize int64

	// MaxConcurrentParts is the maximum concurrent part uploads
	MaxConcurrentParts int

	// RetryAttempts is the number of retry attempts per part
	RetryAttempts int
}

// DefaultMultipartUploadConfig returns default multipart configuration.
func DefaultMultipartUploadConfig() *MultipartUploadConfig {
	return &MultipartUploadConfig{
		PartSize:           8 * 1024 * 1024, // 8MB
		MaxConcurrentParts: 4,
		RetryAttempts:      3,
	}
}
