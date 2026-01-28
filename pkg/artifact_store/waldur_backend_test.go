package artifact_store

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"
)

func TestWaldurBackend(t *testing.T) {
	ctx := context.Background()

	t.Run("NewWaldurBackend", func(t *testing.T) {
		config := &WaldurConfig{
			Endpoint:     "http://localhost:8080",
			Organization: "org1",
			Project:      "proj1",
			Bucket:       "bucket1",
		}

		backend, err := NewWaldurBackend(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if backend.Backend() != BackendWaldur {
			t.Errorf("expected backend %s, got %s", BackendWaldur, backend.Backend())
		}
	})

	t.Run("Put_Get_Exists", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		data := []byte("encrypted artifact data")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519-XSALSA20-POLY1305",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1abc",
			ArtifactType: "face_embedding",
			Metadata:     map[string]string{"key": "value"},
		}

		// Put
		resp, err := backend.Put(ctx, req)
		if err != nil {
			t.Fatalf("put error: %v", err)
		}

		if resp.ContentAddress == nil {
			t.Fatal("expected content address")
		}

		if resp.ContentAddress.Backend != BackendWaldur {
			t.Errorf("expected backend %s, got %s", BackendWaldur, resp.ContentAddress.Backend)
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
			ContentAddress:    resp.ContentAddress,
			RequestingAccount: "cosmos1abc",
		})
		if err != nil {
			t.Fatalf("get error: %v", err)
		}

		if string(getResp.Data) != string(data) {
			t.Error("data mismatch")
		}
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		hash := sha256.Sum256([]byte("nonexistent"))
		addr := &ContentAddress{
			Version:    1,
			Hash:       hash[:],
			Algorithm:  "sha256",
			Backend:    BackendWaldur,
			BackendRef: "nonexistent",
		}

		_, err := backend.Get(ctx, &GetRequest{ContentAddress: addr})
		if err == nil {
			t.Error("expected error for nonexistent artifact")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		data := []byte("to be deleted")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1abc",
			ArtifactType: "face_embedding",
		}

		resp, _ := backend.Put(ctx, req)

		// Delete
		err := backend.Delete(ctx, &DeleteRequest{
			ContentAddress:    resp.ContentAddress,
			RequestingAccount: "cosmos1abc",
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

	t.Run("Delete_Unauthorized", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		data := []byte("protected data")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1abc",
			ArtifactType: "face_embedding",
		}

		resp, _ := backend.Put(ctx, req)

		// Try delete as different account
		err := backend.Delete(ctx, &DeleteRequest{
			ContentAddress:    resp.ContentAddress,
			RequestingAccount: "cosmos1different",
		})
		if err == nil {
			t.Error("expected error for unauthorized delete")
		}
	})

	t.Run("ListByOwner", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		owner := "cosmos1listowner"

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

		if len(listResp.References) != 3 {
			t.Errorf("expected 3 references, got %d", len(listResp.References))
		}
	})

	t.Run("UpdateRetention", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		data := []byte("data with retention")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1abc",
			ArtifactType: "face_embedding",
		}

		resp, _ := backend.Put(ctx, req)

		// Update retention
		tag := NewRetentionTag("policy-1", "cosmos1abc", true)
		tag.SetExpiration(time.Now().Add(24 * time.Hour))

		err := backend.UpdateRetention(ctx, resp.ContentAddress, tag)
		if err != nil {
			t.Fatalf("update retention error: %v", err)
		}
	})

	t.Run("PurgeExpired", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		// Put artifact with past expiration
		tag := NewRetentionTag("policy-1", "cosmos1purge", true)
		pastTime := time.Now().Add(-1 * time.Hour)
		tag.SetExpiration(pastTime)

		req := &PutRequest{
			Data: []byte("expired data"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1purge",
			ArtifactType: "face_embedding",
			RetentionTag: tag,
		}

		resp, _ := backend.Put(ctx, req)

		// Purge expired
		purged, err := backend.PurgeExpired(ctx, 100)
		if err != nil {
			t.Fatalf("purge error: %v", err)
		}

		if purged != 1 {
			t.Errorf("expected 1 purged, got %d", purged)
		}

		// Verify deleted
		exists, _ := backend.Exists(ctx, resp.ContentAddress)
		if exists {
			t.Error("expected expired artifact to be deleted")
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		// Put some data
		for i := 0; i < 5; i++ {
			req := &PutRequest{
				Data: []byte{byte(i), byte(i), byte(i)},
				EncryptionMetadata: &EncryptionMetadata{
					AlgorithmID:     "X25519",
					RecipientKeyIDs: []string{"key1"},
					EnvelopeHash:    make([]byte, 32),
				},
				Owner:        "cosmos1metrics",
				ArtifactType: "face_embedding",
			}
			_, _ = backend.Put(ctx, req)
		}

		metrics, err := backend.GetMetrics(ctx)
		if err != nil {
			t.Fatalf("get metrics error: %v", err)
		}

		if metrics.TotalArtifacts != 5 {
			t.Errorf("expected 5 artifacts, got %d", metrics.TotalArtifacts)
		}

		if metrics.TotalBytes != 15 { // 5 * 3 bytes each
			t.Errorf("expected 15 bytes, got %d", metrics.TotalBytes)
		}

		if metrics.BackendType != BackendWaldur {
			t.Errorf("expected backend %s, got %s", BackendWaldur, metrics.BackendType)
		}
	})

	t.Run("Health", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		err := backend.Health(ctx)
		if err != nil {
			t.Errorf("health check error: %v", err)
		}
	})

	t.Run("GetChunk_NotSupported", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		hash := sha256.Sum256([]byte("test"))
		addr := &ContentAddress{
			Version:    1,
			Hash:       hash[:],
			Algorithm:  "sha256",
			Backend:    BackendWaldur,
			BackendRef: "ref",
		}

		_, err := backend.GetChunk(ctx, addr, 0)
		if err == nil {
			t.Error("expected error for unsupported chunked storage")
		}
	})

	t.Run("RetentionExpired_GetBlocked", func(t *testing.T) {
		backend, _ := NewWaldurBackend(DefaultWaldurConfig())

		// Put artifact with past expiration
		tag := NewRetentionTag("policy-1", "cosmos1expired", false) // Not auto-delete
		pastTime := time.Now().Add(-1 * time.Hour)
		tag.SetExpiration(pastTime)

		req := &PutRequest{
			Data: []byte("expired data"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1expired",
			ArtifactType: "face_embedding",
			RetentionTag: tag,
		}

		resp, _ := backend.Put(ctx, req)

		// Try to get - should fail due to expired retention
		_, err := backend.Get(ctx, &GetRequest{
			ContentAddress:    resp.ContentAddress,
			RequestingAccount: "cosmos1expired",
		})
		if err == nil {
			t.Error("expected error for expired retention")
		}
	})
}

func TestWaldurStreamingBackend(t *testing.T) {
	ctx := context.Background()

	t.Run("PutStream_GetStream", func(t *testing.T) {
		backend, _ := NewWaldurStreamingBackend(DefaultWaldurConfig())

		data := []byte("streaming test data")
		reader := &bytesReadCloser{data: data}

		req := &PutStreamRequest{
			Reader: reader,
			Size:   int64(len(data)),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1stream",
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
		defer func() { _ = stream.Close() }()

		// Read all
		buf := make([]byte, 1024)
		n, _ := stream.Read(buf)

		if string(buf[:n]) != string(data) {
			t.Error("streamed data mismatch")
		}
	})
}
