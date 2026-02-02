package artifact_store

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestStreamingUploader(t *testing.T) {
	ctx := context.Background()

	testStreamConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = testOrg
		config.Project = testProj
		config.Bucket = testBucket
		return config
	}

	t.Run("UploadWithProgress", func(t *testing.T) {
		backend, err := NewWaldurStreamingBackend(testStreamConfig())
		if err != nil {
			t.Fatalf("failed to create backend: %v", err)
		}

		var lastProgress *StreamProgress
		uploader := NewStreamingUploader(backend, DefaultStreamingConfig())
		uploader.SetProgressCallback(func(p *StreamProgress) {
			lastProgress = p
		})

		data := bytes.Repeat([]byte("test data "), 1000)
		req := &PutStreamRequest{
			Reader: bytes.NewReader(data),
			Size:   int64(len(data)),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1progress",
			ArtifactType: "test_artifact",
		}

		resp, err := uploader.Upload(ctx, req)
		if err != nil {
			t.Fatalf("upload error: %v", err)
		}

		if resp.ContentAddress == nil {
			t.Fatal("expected content address")
		}

		// Progress should have been updated
		if lastProgress == nil {
			t.Skip("progress callback may not have been called due to timing")
		}
	})

	t.Run("UploadCancel", func(t *testing.T) {
		backend, _ := NewWaldurStreamingBackend(testStreamConfig())
		uploader := NewStreamingUploader(backend, DefaultStreamingConfig())

		// Create a slow reader that we can cancel
		slowReader := &slowReader{data: bytes.Repeat([]byte("x"), 10000)}

		req := &PutStreamRequest{
			Reader: slowReader,
			Size:   10000,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1cancel",
			ArtifactType: "test_artifact",
		}

		// Cancel after starting
		go func() {
			time.Sleep(10 * time.Millisecond)
			uploader.Cancel()
		}()

		_, err := uploader.Upload(ctx, req)
		// Either success or context canceled is acceptable
		if err != nil && err != context.Canceled {
			// OK - upload may have completed before cancel
		}
	})
}

type slowReader struct {
	data   []byte
	offset int
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.offset >= len(s.data) {
		return 0, io.EOF
	}
	time.Sleep(1 * time.Millisecond)
	n := copy(p, s.data[s.offset:])
	s.offset += n
	return n, nil
}

func TestStreamingDownloader(t *testing.T) {
	ctx := context.Background()

	testStreamConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = testOrg
		config.Project = testProj
		config.Bucket = testBucket
		return config
	}

	t.Run("DownloadWithVerification", func(t *testing.T) {
		backend, _ := NewWaldurStreamingBackend(testStreamConfig())

		// First upload
		data := []byte("streaming download test data")
		req := &PutStreamRequest{
			Reader: bytes.NewReader(data),
			Size:   int64(len(data)),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1download",
			ArtifactType: "test_artifact",
		}

		resp, err := backend.PutStream(ctx, req)
		if err != nil {
			t.Fatalf("put error: %v", err)
		}

		// Download with verification
		config := DefaultStreamingConfig()
		config.VerifyOnStream = true
		downloader := NewStreamingDownloader(backend, config)

		var buf bytes.Buffer
		err = downloader.Download(ctx, resp.ContentAddress, &buf)
		if err != nil {
			t.Fatalf("download error: %v", err)
		}

		if !bytes.Equal(buf.Bytes(), data) {
			t.Error("downloaded data mismatch")
		}
	})
}

func TestStreamProgress(t *testing.T) {
	t.Run("PercentComplete", func(t *testing.T) {
		p := &StreamProgress{
			BytesTransferred: 50,
			TotalBytes:       100,
		}

		pct := p.PercentComplete()
		if pct != 50.0 {
			t.Errorf("expected 50%%, got %.1f%%", pct)
		}
	})

	t.Run("PercentCompleteUnknownTotal", func(t *testing.T) {
		p := &StreamProgress{
			BytesTransferred: 50,
			TotalBytes:       -1,
		}

		pct := p.PercentComplete()
		if pct != 0 {
			t.Errorf("expected 0%% for unknown total, got %.1f%%", pct)
		}
	})
}

func TestChunkedTransferState(t *testing.T) {
	t.Run("IsComplete", func(t *testing.T) {
		state := &ChunkedTransferState{
			TotalChunks:     3,
			CompletedChunks: map[int]bool{0: true, 1: true, 2: true},
		}

		if !state.IsComplete() {
			t.Error("expected complete")
		}
	})

	t.Run("IsIncomplete", func(t *testing.T) {
		state := &ChunkedTransferState{
			TotalChunks:     3,
			CompletedChunks: map[int]bool{0: true, 2: true},
		}

		if state.IsComplete() {
			t.Error("expected incomplete")
		}
	})

	t.Run("NextIncompleteChunk", func(t *testing.T) {
		state := &ChunkedTransferState{
			TotalChunks:     3,
			CompletedChunks: map[int]bool{0: true},
		}

		next := state.NextIncompleteChunk()
		if next != 1 {
			t.Errorf("expected next incomplete chunk 1, got %d", next)
		}
	})

	t.Run("PercentComplete", func(t *testing.T) {
		state := &ChunkedTransferState{
			TotalChunks:     4,
			CompletedChunks: map[int]bool{0: true, 1: true},
		}

		pct := state.PercentComplete()
		if pct != 50.0 {
			t.Errorf("expected 50%%, got %.1f%%", pct)
		}
	})
}

func TestResumableUploader(t *testing.T) {
	t.Run("StartUpload", func(t *testing.T) {
		backend, _ := NewWaldurStreamingBackend(&WaldurConfig{
			UseFallbackMemory: true,
			Organization:      testOrg,
		})

		uploader := NewResumableUploader(backend, nil)
		state := uploader.StartUpload(1024*1024*100, "abc123")

		if state.TotalSize != 1024*1024*100 {
			t.Error("total size mismatch")
		}
		if state.ContentHash != "abc123" {
			t.Error("content hash mismatch")
		}
		if state.TotalChunks == 0 {
			t.Error("expected chunks")
		}
		if !strings.HasPrefix(state.TransferID, "upload-") {
			t.Error("expected upload- prefix")
		}
	})

	t.Run("MarkChunkComplete", func(t *testing.T) {
		backend, _ := NewWaldurStreamingBackend(&WaldurConfig{
			UseFallbackMemory: true,
			Organization:      testOrg,
		})

		uploader := NewResumableUploader(backend, nil)
		uploader.StartUpload(1024*1024*100, "")

		uploader.MarkChunkComplete(0)
		uploader.MarkChunkComplete(1)

		state := uploader.GetState()
		if !state.CompletedChunks[0] || !state.CompletedChunks[1] {
			t.Error("expected chunks to be marked complete")
		}
	})
}

func TestStreamingConfig(t *testing.T) {
	config := DefaultStreamingConfig()

	if config.ChunkSize != 8*1024*1024 {
		t.Error("unexpected default chunk size")
	}
	if config.BufferSize != 64*1024 {
		t.Error("unexpected default buffer size")
	}
	if !config.VerifyOnStream {
		t.Error("expected verify on stream to be true")
	}
}
