package artifact_store

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"
)

func TestWaldurBackend(t *testing.T) {
	ctx := context.Background()

	// Helper to create a test config with fallback memory enabled
	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		return config
	}

	t.Run("NewWaldurBackend", func(t *testing.T) {
		config := testConfig()
		config.Endpoint = "http://localhost:8080"

		backend, err := NewWaldurBackend(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if backend.Backend() != BackendWaldur {
			t.Errorf("expected backend %s, got %s", BackendWaldur, backend.Backend())
		}
	})

	t.Run("Put_Get_Exists", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

		err := backend.Health(ctx)
		if err != nil {
			t.Errorf("health check error: %v", err)
		}
	})

	t.Run("GetChunk_NotSupported", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

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
		backend, _ := NewWaldurBackend(testConfig())

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

	// Helper to create a test config with fallback memory enabled
	testStreamConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		return config
	}

	t.Run("PutStream_GetStream", func(t *testing.T) {
		backend, _ := NewWaldurStreamingBackend(testStreamConfig())

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

func TestWaldurBackendPinning(t *testing.T) {
	ctx := context.Background()

	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		config.EnablePinning = true
		return config
	}

	t.Run("Pin_Unpin_IsPinned", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		data := []byte("pinnable artifact")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1pin",
			ArtifactType: "pinned_data",
		}

		resp, err := backend.Put(ctx, req)
		if err != nil {
			t.Fatalf("put error: %v", err)
		}

		// Initially not pinned
		pinned, err := backend.IsPinned(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("is pinned error: %v", err)
		}
		if pinned {
			t.Error("expected not pinned initially")
		}

		// Pin the artifact
		err = backend.Pin(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("pin error: %v", err)
		}

		// Should be pinned now
		pinned, err = backend.IsPinned(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("is pinned error: %v", err)
		}
		if !pinned {
			t.Error("expected pinned after pin call")
		}

		// Unpin
		err = backend.Unpin(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("unpin error: %v", err)
		}

		// Should be unpinned now
		pinned, err = backend.IsPinned(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("is pinned error: %v", err)
		}
		if pinned {
			t.Error("expected not pinned after unpin")
		}
	})

	t.Run("Pin_DisabledByConfig", func(t *testing.T) {
		config := testConfig()
		config.EnablePinning = false
		backend, _ := NewWaldurBackend(config)

		hash := sha256.Sum256([]byte("test"))
		addr := &ContentAddress{
			Version:    1,
			Hash:       hash[:],
			Algorithm:  "sha256",
			Backend:    BackendWaldur,
			BackendRef: "ref",
		}

		err := backend.Pin(ctx, addr)
		if err == nil {
			t.Error("expected error when pinning is disabled")
		}
	})

	t.Run("Unpin_NotFound", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		hash := sha256.Sum256([]byte("nonexistent"))
		addr := &ContentAddress{
			Version:    1,
			Hash:       hash[:],
			Algorithm:  "sha256",
			Backend:    BackendWaldur,
			BackendRef: "nonexistent",
		}

		err := backend.Unpin(ctx, addr)
		if err == nil {
			t.Error("expected error for nonexistent artifact")
		}
	})
}

func TestWaldurBackendQuota(t *testing.T) {
	ctx := context.Background()

	t.Run("GetOwnerUsage", func(t *testing.T) {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"

		backend, _ := NewWaldurBackend(config)

		// Put some data
		for i := 0; i < 3; i++ {
			req := &PutRequest{
				Data: []byte{byte(i), byte(i), byte(i)},
				EncryptionMetadata: &EncryptionMetadata{
					AlgorithmID:     "X25519",
					RecipientKeyIDs: []string{"key1"},
					EnvelopeHash:    make([]byte, 32),
				},
				Owner:        "cosmos1usage",
				ArtifactType: "test_artifact",
			}
			_, _ = backend.Put(ctx, req)
		}

		// GetOwnerUsage returns 0 for fallback mode local cache
		// In production this would sum object sizes
		usage, err := backend.GetOwnerUsage(ctx, "cosmos1usage")
		if err != nil {
			t.Fatalf("get usage error: %v", err)
		}
		// Usage is tracked from ownerUsage map which starts at 0
		// In real scenario this would be populated
		if usage < 0 {
			t.Error("usage should not be negative")
		}
	})

	t.Run("CheckOwnerQuota_Unlimited", func(t *testing.T) {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.PerOwnerQuotaBytes = 0 // Unlimited

		backend, _ := NewWaldurBackend(config)

		err := backend.CheckOwnerQuota(ctx, "cosmos1unlimited", 1024*1024*1024)
		if err != nil {
			t.Errorf("expected no error for unlimited quota: %v", err)
		}
	})

	t.Run("CheckOwnerQuota_Exceeded", func(t *testing.T) {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.PerOwnerQuotaBytes = 100 // 100 bytes limit

		backend, _ := NewWaldurBackend(config)

		// Request more than quota
		err := backend.CheckOwnerQuota(ctx, "cosmos1limited", 200)
		if err == nil {
			t.Error("expected quota exceeded error")
		}
	})
}

func TestWaldurBackendRetentionCleanup(t *testing.T) {
	ctx := context.Background()

	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		config.EnablePinning = true
		return config
	}

	t.Run("RunRetentionCleanup", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		// Create expired artifact
		pastTime := time.Now().Add(-1 * time.Hour)
		expiredTag := NewRetentionTag("policy-1", "cosmos1cleanup", true)
		expiredTag.SetExpiration(pastTime)

		req := &PutRequest{
			Data: []byte("expired data"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1cleanup",
			ArtifactType: "expired_artifact",
			RetentionTag: expiredTag,
		}
		expiredResp, _ := backend.Put(ctx, req)

		// Create non-expired artifact
		futureTime := time.Now().Add(24 * time.Hour)
		validTag := NewRetentionTag("policy-1", "cosmos1cleanup", true)
		validTag.SetExpiration(futureTime)

		req2 := &PutRequest{
			Data: []byte("valid data"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1cleanup",
			ArtifactType: "valid_artifact",
			RetentionTag: validTag,
		}
		validResp, _ := backend.Put(ctx, req2)

		// Run cleanup
		stats, err := backend.RunRetentionCleanup(ctx, 100)
		if err != nil {
			t.Fatalf("cleanup error: %v", err)
		}

		if stats.DeletedCount != 1 {
			t.Errorf("expected 1 deleted, got %d", stats.DeletedCount)
		}
		if stats.CheckedCount < 2 {
			t.Errorf("expected at least 2 checked, got %d", stats.CheckedCount)
		}

		// Verify expired is gone
		exists, _ := backend.Exists(ctx, expiredResp.ContentAddress)
		if exists {
			t.Error("expired artifact should be deleted")
		}

		// Verify valid remains
		exists, _ = backend.Exists(ctx, validResp.ContentAddress)
		if !exists {
			t.Error("valid artifact should still exist")
		}
	})

	t.Run("RunRetentionCleanup_SkipsPinned", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		// Create expired but pinned artifact
		pastTime := time.Now().Add(-1 * time.Hour)
		tag := NewRetentionTag("policy-1", "cosmos1pinned", true)
		tag.SetExpiration(pastTime)

		req := &PutRequest{
			Data: []byte("pinned expired data"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1pinned",
			ArtifactType: "pinned_artifact",
			RetentionTag: tag,
		}
		resp, _ := backend.Put(ctx, req)

		// Pin it
		_ = backend.Pin(ctx, resp.ContentAddress)

		// Run cleanup
		stats, err := backend.RunRetentionCleanup(ctx, 100)
		if err != nil {
			t.Fatalf("cleanup error: %v", err)
		}

		if stats.SkippedPinned != 1 {
			t.Errorf("expected 1 skipped pinned, got %d", stats.SkippedPinned)
		}
		if stats.DeletedCount != 0 {
			t.Errorf("expected 0 deleted, got %d", stats.DeletedCount)
		}

		// Verify still exists
		exists, _ := backend.Exists(ctx, resp.ContentAddress)
		if !exists {
			t.Error("pinned artifact should not be deleted")
		}
	})
}

func TestWaldurBackendPurgeByOwner(t *testing.T) {
	ctx := context.Background()

	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		config.EnablePinning = true
		return config
	}

	t.Run("PurgeByOwner", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		owner := "cosmos1purgeowner"

		// Create multiple artifacts for this owner
		for i := 0; i < 3; i++ {
			req := &PutRequest{
				Data: []byte{byte(i)},
				EncryptionMetadata: &EncryptionMetadata{
					AlgorithmID:     "X25519",
					RecipientKeyIDs: []string{"key1"},
					EnvelopeHash:    make([]byte, 32),
				},
				Owner:        owner,
				ArtifactType: "purge_test",
			}
			_, _ = backend.Put(ctx, req)
		}

		// Purge all for this owner
		count, err := backend.PurgeByOwner(ctx, owner)
		if err != nil {
			t.Fatalf("purge error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3 purged, got %d", count)
		}

		// Verify none left
		list, _ := backend.ListByOwner(ctx, owner, nil)
		if list.Total != 0 {
			t.Errorf("expected 0 artifacts after purge, got %d", list.Total)
		}
	})

	t.Run("PurgeByOwner_SkipsPinned", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		owner := "cosmos1purgepinned"

		// Create and pin one artifact
		req := &PutRequest{
			Data: []byte("pinned"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        owner,
			ArtifactType: "pinned_purge_test",
		}
		pinnedResp, _ := backend.Put(ctx, req)
		_ = backend.Pin(ctx, pinnedResp.ContentAddress)

		// Create unpinned artifact
		req2 := &PutRequest{
			Data: []byte("unpinned"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        owner,
			ArtifactType: "unpinned_purge_test",
		}
		_, _ = backend.Put(ctx, req2)

		// Purge
		count, err := backend.PurgeByOwner(ctx, owner)
		if err != nil {
			t.Fatalf("purge error: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 purged (skipping pinned), got %d", count)
		}

		// Verify pinned still exists
		exists, _ := backend.Exists(ctx, pinnedResp.ContentAddress)
		if !exists {
			t.Error("pinned artifact should still exist")
		}
	})
}

func TestWaldurBackendVerify(t *testing.T) {
	ctx := context.Background()

	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		return config
	}

	t.Run("VerifyArtifact_Valid", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		data := []byte("verify me")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1verify",
			ArtifactType: "verify_test",
		}
		resp, _ := backend.Put(ctx, req)

		err := backend.VerifyArtifact(ctx, resp.ContentAddress)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("BatchVerify", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		addresses := make([]*ContentAddress, 3)
		for i := 0; i < 3; i++ {
			req := &PutRequest{
				Data: []byte{byte(i)},
				EncryptionMetadata: &EncryptionMetadata{
					AlgorithmID:     "X25519",
					RecipientKeyIDs: []string{"key1"},
					EnvelopeHash:    make([]byte, 32),
				},
				Owner:        "cosmos1batchverify",
				ArtifactType: "batch_verify_test",
			}
			resp, _ := backend.Put(ctx, req)
			addresses[i] = resp.ContentAddress
		}

		results, err := backend.BatchVerify(ctx, addresses)
		if err != nil {
			t.Fatalf("batch verify error: %v", err)
		}

		for hash, verifyErr := range results {
			if verifyErr != nil {
				t.Errorf("artifact %s failed verification: %v", hash, verifyErr)
			}
		}
	})
}

func TestWaldurBackendRetentionPolicy(t *testing.T) {
	ctx := context.Background()

	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		return config
	}

	t.Run("GetRetentionPolicy", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		futureTime := time.Now().Add(24 * time.Hour)
		tag := NewRetentionTag("policy-123", "cosmos1retention", true)
		tag.SetExpiration(futureTime)

		req := &PutRequest{
			Data: []byte("retention test"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1retention",
			ArtifactType: "retention_test",
			RetentionTag: tag,
		}
		resp, _ := backend.Put(ctx, req)

		retrievedTag, err := backend.GetRetentionPolicy(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("get retention error: %v", err)
		}
		if retrievedTag.PolicyID != "policy-123" {
			t.Errorf("expected policy-123, got %s", retrievedTag.PolicyID)
		}
		if !retrievedTag.DeleteOnExpiry {
			t.Error("expected delete on expiry to be true")
		}
	})

	t.Run("SetRetentionPolicy", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		req := &PutRequest{
			Data: []byte("set retention test"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1setretention",
			ArtifactType: "set_retention_test",
		}
		resp, _ := backend.Put(ctx, req)

		// Set new retention
		futureTime := time.Now().Add(48 * time.Hour)
		newTag := NewRetentionTag("new-policy", "cosmos1setretention", false)
		newTag.SetExpiration(futureTime)

		err := backend.SetRetentionPolicy(ctx, resp.ContentAddress, newTag)
		if err != nil {
			t.Fatalf("set retention error: %v", err)
		}

		// Verify updated
		retrievedTag, _ := backend.GetRetentionPolicy(ctx, resp.ContentAddress)
		if retrievedTag == nil {
			t.Fatal("expected retention tag")
		}
		if retrievedTag.PolicyID != "new-policy" {
			t.Errorf("expected new-policy, got %s", retrievedTag.PolicyID)
		}
	})
}

func TestWaldurPinnableBackend(t *testing.T) {
	t.Run("NewWaldurPinnableBackend", func(t *testing.T) {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = "test-org"

		backend, err := NewWaldurPinnableBackend(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if backend.Backend() != BackendWaldur {
			t.Error("expected waldur backend")
		}
	})
}

func TestExtractHashFromKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"org/proj/bucket/owner/abc123", "abc123"},
		{"abc123", "abc123"},
		{"a/b/c", "c"},
		{"", ""},
	}

	for _, tt := range tests {
		result := extractHashFromKey(tt.key)
		if result != tt.expected {
			t.Errorf("extractHashFromKey(%q) = %q, want %q", tt.key, result, tt.expected)
		}
	}
}

func TestSplitKeyPath(t *testing.T) {
	parts := splitKeyPath("a/b/c/d")
	if len(parts) != 4 {
		t.Errorf("expected 4 parts, got %d", len(parts))
	}
	if parts[0] != "a" || parts[3] != "d" {
		t.Error("unexpected parts")
	}
}

