package artifact_store

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"
)

func TestIPFSBackend(t *testing.T) {
	ctx := context.Background()

	t.Run("NewIPFSBackend", func(t *testing.T) {
		config := DefaultIPFSConfig()

		backend, err := NewIPFSBackend(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if backend.Backend() != BackendIPFS {
			t.Errorf("expected backend %s, got %s", BackendIPFS, backend.Backend())
		}
	})

	t.Run("Put_Get_Exists", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		data := []byte("encrypted ipfs artifact data")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519-XSALSA20-POLY1305",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1ipfs",
			ArtifactType: "face_embedding",
		}

		// Put
		resp, err := backend.Put(ctx, req)
		if err != nil {
			t.Fatalf("put error: %v", err)
		}

		if resp.ContentAddress == nil {
			t.Fatal("expected content address")
		}

		if resp.ContentAddress.Backend != BackendIPFS {
			t.Errorf("expected backend %s, got %s", BackendIPFS, resp.ContentAddress.Backend)
		}

		// Check CID format (stubbed)
		if len(resp.ContentAddress.BackendRef) < 2 || resp.ContentAddress.BackendRef[:2] != "Qm" {
			t.Errorf("expected CID to start with Qm, got %s", resp.ContentAddress.BackendRef)
		}

		// Exists
		exists, err := backend.Exists(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("exists error: %v", err)
		}
		if !exists {
			t.Error("expected artifact to exist")
		}

		// Get
		getResp, err := backend.Get(ctx, &GetRequest{
			ContentAddress: resp.ContentAddress,
		})
		if err != nil {
			t.Fatalf("get error: %v", err)
		}

		if string(getResp.Data) != string(data) {
			t.Error("data mismatch")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		data := []byte("to be deleted from ipfs")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1ipfsdel",
			ArtifactType: "face_embedding",
		}

		resp, _ := backend.Put(ctx, req)

		// Delete
		err := backend.Delete(ctx, &DeleteRequest{
			ContentAddress:    resp.ContentAddress,
			RequestingAccount: "cosmos1ipfsdel",
		})
		if err != nil {
			t.Fatalf("delete error: %v", err)
		}

		// Verify deleted
		exists, _ := backend.Exists(ctx, resp.ContentAddress)
		if exists {
			t.Error("expected artifact to be deleted")
		}
	})

	t.Run("PutChunked", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		// Create data that spans multiple chunks
		data := make([]byte, 1024)
		for i := range data {
			data[i] = byte(i % 256)
		}

		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1chunked",
			ArtifactType: "raw_image",
		}

		resp, manifest, err := backend.PutChunked(ctx, data, 256, meta)
		if err != nil {
			t.Fatalf("put chunked error: %v", err)
		}

		if resp.ContentAddress == nil {
			t.Fatal("expected content address")
		}

		if manifest == nil {
			t.Fatal("expected manifest")
		}

		if manifest.ChunkCount != 4 {
			t.Errorf("expected 4 chunks, got %d", manifest.ChunkCount)
		}

		if len(manifest.Chunks) != 4 {
			t.Errorf("expected 4 chunk infos, got %d", len(manifest.Chunks))
		}

		if manifest.TotalSize != 1024 {
			t.Errorf("expected total size 1024, got %d", manifest.TotalSize)
		}
	})

	t.Run("GetChunked", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		// Create data that spans multiple chunks
		data := make([]byte, 512)
		for i := range data {
			data[i] = byte(i % 256)
		}

		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1getc",
			ArtifactType: "raw_image",
		}

		_, manifest, _ := backend.PutChunked(ctx, data, 128, meta)

		// Get chunked
		retrieved, err := backend.GetChunked(ctx, manifest)
		if err != nil {
			t.Fatalf("get chunked error: %v", err)
		}

		if len(retrieved) != len(data) {
			t.Errorf("expected %d bytes, got %d", len(data), len(retrieved))
		}

		for i := range data {
			if data[i] != retrieved[i] {
				t.Errorf("data mismatch at byte %d", i)
				break
			}
		}
	})

	t.Run("GetChunk", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		// Create chunked data
		data := make([]byte, 512)
		for i := range data {
			data[i] = byte(i % 256)
		}

		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1getch",
			ArtifactType: "raw_image",
		}

		resp, manifest, _ := backend.PutChunked(ctx, data, 128, meta)

		// Get individual chunk
		chunk, err := backend.GetChunk(ctx, resp.ContentAddress, 1)
		if err != nil {
			t.Fatalf("get chunk error: %v", err)
		}

		if chunk.Index != 1 {
			t.Errorf("expected chunk index 1, got %d", chunk.Index)
		}

		if len(chunk.Data) != 128 {
			t.Errorf("expected chunk size 128, got %d", len(chunk.Data))
		}

		// Verify hash
		expectedHash := manifest.Chunks[1].Hash
		if string(chunk.Hash) != string(expectedHash) {
			t.Error("chunk hash mismatch")
		}
	})

	t.Run("GetChunk_InvalidIndex", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		data := make([]byte, 256)
		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1invalid",
			ArtifactType: "raw_image",
		}

		resp, _, _ := backend.PutChunked(ctx, data, 128, meta)

		// Try to get invalid chunk index
		_, err := backend.GetChunk(ctx, resp.ContentAddress, 10)
		if err == nil {
			t.Error("expected error for invalid chunk index")
		}
	})

	t.Run("VerifyChunks", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		data := make([]byte, 256)
		for i := range data {
			data[i] = byte(i)
		}

		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1verify",
			ArtifactType: "raw_image",
		}

		_, manifest, _ := backend.PutChunked(ctx, data, 128, meta)

		err := backend.VerifyChunks(ctx, manifest)
		if err != nil {
			t.Fatalf("verify chunks error: %v", err)
		}
	})

	t.Run("ListByOwner", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		owner := "cosmos1ipfslist"

		// Put multiple artifacts
		for i := 0; i < 3; i++ {
			req := &PutRequest{
				Data: []byte{byte(i)},
				EncryptionMetadata: &EncryptionMetadata{
					AlgorithmID:     "X25519",
					RecipientKeyIDs: []string{"key1"},
					EnvelopeHash:    make([]byte, 32),
				},
				Owner:        owner,
				ArtifactType: "face_embedding",
			}
			_, _ = backend.Put(ctx, req)
		}

		// List
		listResp, err := backend.ListByOwner(ctx, owner, nil)
		if err != nil {
			t.Fatalf("list error: %v", err)
		}

		if listResp.Total != 3 {
			t.Errorf("expected 3 artifacts, got %d", listResp.Total)
		}
	})

	t.Run("PurgeExpired_WithChunks", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		// Create chunked data with past expiration
		data := make([]byte, 256)

		tag := NewRetentionTag("policy-1", "cosmos1purgeipfs", true)
		pastTime := time.Now().Add(-1 * time.Hour)
		tag.SetExpiration(pastTime)

		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1purgeipfs",
			ArtifactType: "raw_image",
			RetentionTag: tag,
		}

		resp, _, _ := backend.PutChunked(ctx, data, 128, meta)

		// Verify chunks exist
		metricsBefore, _ := backend.GetMetrics(ctx)
		if metricsBefore.TotalChunks != 2 {
			t.Errorf("expected 2 chunks before purge, got %d", metricsBefore.TotalChunks)
		}

		// Purge expired
		purged, err := backend.PurgeExpired(ctx, 100)
		if err != nil {
			t.Fatalf("purge error: %v", err)
		}

		if purged != 1 {
			t.Errorf("expected 1 purged, got %d", purged)
		}

		// Verify artifact deleted
		exists, _ := backend.Exists(ctx, resp.ContentAddress)
		if exists {
			t.Error("expected artifact to be deleted")
		}

		// Verify chunks cleaned up
		metricsAfter, _ := backend.GetMetrics(ctx)
		if metricsAfter.TotalChunks != 0 {
			t.Errorf("expected 0 chunks after purge, got %d", metricsAfter.TotalChunks)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		// Put some data
		for i := 0; i < 3; i++ {
			req := &PutRequest{
				Data: []byte{byte(i), byte(i)},
				EncryptionMetadata: &EncryptionMetadata{
					AlgorithmID:     "X25519",
					RecipientKeyIDs: []string{"key1"},
					EnvelopeHash:    make([]byte, 32),
				},
				Owner:        "cosmos1ipfsmetrics",
				ArtifactType: "face_embedding",
			}
			_, _ = backend.Put(ctx, req)
		}

		// Put chunked data
		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1ipfsmetrics",
			ArtifactType: "raw_image",
		}
		_, _, _ = backend.PutChunked(ctx, make([]byte, 256), 128, meta)

		metrics, err := backend.GetMetrics(ctx)
		if err != nil {
			t.Fatalf("get metrics error: %v", err)
		}

		if metrics.TotalArtifacts != 4 {
			t.Errorf("expected 4 artifacts, got %d", metrics.TotalArtifacts)
		}

		if metrics.TotalChunks != 2 {
			t.Errorf("expected 2 chunks, got %d", metrics.TotalChunks)
		}

		if metrics.BackendType != BackendIPFS {
			t.Errorf("expected backend %s, got %s", BackendIPFS, metrics.BackendType)
		}
	})

	t.Run("Health", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		err := backend.Health(ctx)
		if err != nil {
			t.Errorf("health check error: %v", err)
		}
	})

	t.Run("HashVerification", func(t *testing.T) {
		backend, _ := NewIPFSBackend(DefaultIPFSConfig())

		data := []byte("test data for hash verification")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1hash",
			ArtifactType: "face_embedding",
		}

		resp, _ := backend.Put(ctx, req)

		// Verify hash
		expectedHash := sha256.Sum256(data)
		if string(resp.ContentAddress.Hash) != string(expectedHash[:]) {
			t.Error("content hash mismatch")
		}
	})
}

func TestIPFSStreamingBackend(t *testing.T) {
	ctx := context.Background()

	t.Run("PutStream_GetStream", func(t *testing.T) {
		backend, _ := NewIPFSStreamingBackend(DefaultIPFSConfig())

		data := []byte("ipfs streaming test data")
		reader := &bytesReadCloser{data: data}

		req := &PutStreamRequest{
			Reader: reader,
			Size:   int64(len(data)),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1ipfsstream",
			ArtifactType: "raw_image",
		}

		resp, err := backend.PutStream(ctx, req)
		if err != nil {
			t.Fatalf("put stream error: %v", err)
		}

		// Get stream
		stream, err := backend.GetStream(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("get stream error: %v", err)
		}
		defer stream.Close()

		// Read all
		buf := make([]byte, 1024)
		n, _ := stream.Read(buf)

		if string(buf[:n]) != string(data) {
			t.Error("streamed data mismatch")
		}
	})
}

func TestIPFSChunkManifestIntegrity(t *testing.T) {
	ctx := context.Background()
	backend, _ := NewIPFSBackend(DefaultIPFSConfig())

	t.Run("ChunkOrderIsPreserved", func(t *testing.T) {
		// Create data with distinct patterns per chunk
		data := make([]byte, 400)
		for i := 0; i < 100; i++ {
			data[i] = 0x11 // Chunk 0
		}
		for i := 100; i < 200; i++ {
			data[i] = 0x22 // Chunk 1
		}
		for i := 200; i < 300; i++ {
			data[i] = 0x33 // Chunk 2
		}
		for i := 300; i < 400; i++ {
			data[i] = 0x44 // Chunk 3
		}

		meta := &PutMetadata{
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1order",
			ArtifactType: "raw_image",
		}

		_, manifest, err := backend.PutChunked(ctx, data, 100, meta)
		if err != nil {
			t.Fatalf("put chunked error: %v", err)
		}

		// Verify chunks are ordered
		for i := uint32(0); i < manifest.ChunkCount; i++ {
			if manifest.Chunks[i].Index != i {
				t.Errorf("chunk %d has wrong index %d", i, manifest.Chunks[i].Index)
			}
		}

		// Get chunked and verify order is preserved
		retrieved, err := backend.GetChunked(ctx, manifest)
		if err != nil {
			t.Fatalf("get chunked error: %v", err)
		}

		// Verify patterns
		for i := 0; i < 100; i++ {
			if retrieved[i] != 0x11 {
				t.Errorf("chunk 0 data corrupted at byte %d", i)
				break
			}
		}
		for i := 100; i < 200; i++ {
			if retrieved[i] != 0x22 {
				t.Errorf("chunk 1 data corrupted at byte %d", i)
				break
			}
		}
		for i := 200; i < 300; i++ {
			if retrieved[i] != 0x33 {
				t.Errorf("chunk 2 data corrupted at byte %d", i)
				break
			}
		}
		for i := 300; i < 400; i++ {
			if retrieved[i] != 0x44 {
				t.Errorf("chunk 3 data corrupted at byte %d", i)
				break
			}
		}
	})
}
