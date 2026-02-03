package artifact_store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestIntegrityChecker(t *testing.T) {
	ctx := context.Background()

	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = testOrg
		config.Project = testProj
		config.Bucket = testBucket
		return config
	}

	t.Run("VerifyValid", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		data := []byte("test data for verification")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1verify",
			ArtifactType: "test_artifact",
		}

		resp, err := backend.Put(ctx, req)
		if err != nil {
			t.Fatalf("put error: %v", err)
		}

		checker := NewIntegrityChecker(backend, nil)
		result, err := checker.Verify(ctx, resp.ContentAddress)
		if err != nil {
			t.Fatalf("verify error: %v", err)
		}

		if !result.Valid {
			t.Errorf("expected valid, got invalid: %s", result.Error)
		}
		if result.BytesVerified != int64(len(data)) {
			t.Errorf("expected %d bytes verified, got %d", len(data), result.BytesVerified)
		}
	})

	t.Run("VerifyInvalid", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		data := []byte("test data")
		req := &PutRequest{
			Data: data,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1invalid",
			ArtifactType: "test_artifact",
		}

		resp, _ := backend.Put(ctx, req)

		// Corrupt the expected hash
		corruptedAddr := &ContentAddress{
			Version:    resp.ContentAddress.Version,
			Hash:       bytes.Repeat([]byte{0xff}, 32), // Wrong hash
			Algorithm:  resp.ContentAddress.Algorithm,
			Size:       resp.ContentAddress.Size,
			Backend:    resp.ContentAddress.Backend,
			BackendRef: resp.ContentAddress.BackendRef,
		}

		checker := NewIntegrityChecker(backend, nil)
		result, _ := checker.Verify(ctx, corruptedAddr)

		if result.Valid {
			t.Error("expected invalid result for corrupted hash")
		}
	})

	t.Run("VerifyNotFound", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		hash := sha256.Sum256([]byte("nonexistent"))
		addr := &ContentAddress{
			Version:    1,
			Hash:       hash[:],
			Algorithm:  "sha256",
			Backend:    BackendWaldur,
			BackendRef: "nonexistent",
		}

		checker := NewIntegrityChecker(backend, nil)
		result, err := checker.Verify(ctx, addr)

		if err == nil {
			t.Error("expected error for nonexistent artifact")
		}
		if result.Error == "" {
			t.Error("expected error message in result")
		}
	})
}

func TestVerifyStream(t *testing.T) {
	t.Run("FullVerification", func(t *testing.T) {
		config := func() *WaldurConfig {
			c := DefaultWaldurConfig()
			c.UseFallbackMemory = true
			c.Organization = testOrg
			return c
		}

		backend, _ := NewWaldurBackend(config())

		data := []byte("streaming verification data")
		hash := sha256.Sum256(data)

		addr := &ContentAddress{
			Version:   1,
			Hash:      hash[:],
			Algorithm: "sha256",
		}

		checker := NewIntegrityChecker(backend, DefaultIntegrityCheckOptions())
		reader := bytes.NewReader(data)
		result, err := checker.VerifyStream(context.Background(), addr, reader)

		if err != nil {
			t.Fatalf("verify stream error: %v", err)
		}
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("PartialVerification", func(t *testing.T) {
		config := func() *WaldurConfig {
			c := DefaultWaldurConfig()
			c.UseFallbackMemory = true
			c.Organization = testOrg
			return c
		}

		backend, _ := NewWaldurBackend(config())

		data := bytes.Repeat([]byte("x"), 10*1024*1024) // 10MB
		hash := sha256.Sum256(data)

		addr := &ContentAddress{
			Version:   1,
			Hash:      hash[:],
			Algorithm: "sha256",
		}

		options := DefaultIntegrityCheckOptions()
		options.FullVerification = false
		options.PartialVerificationBytes = 1024 * 1024 // 1MB

		checker := NewIntegrityChecker(backend, options)
		reader := bytes.NewReader(data)
		result, err := checker.VerifyStream(context.Background(), addr, reader)

		if err != nil {
			t.Fatalf("verify stream error: %v", err)
		}
		// Partial verification doesn't compare hashes
		if !result.Valid {
			t.Error("expected valid result for partial verification")
		}
		if result.BytesVerified != 1024*1024 {
			t.Errorf("expected 1MB verified, got %d", result.BytesVerified)
		}
	})
}

func TestBatchIntegrityChecker(t *testing.T) {
	ctx := context.Background()

	testConfig := func() *WaldurConfig {
		config := DefaultWaldurConfig()
		config.UseFallbackMemory = true
		config.Organization = testOrg
		config.Project = "test-proj"
		config.Bucket = "test-bucket"
		return config
	}

	t.Run("VerifyBatch", func(t *testing.T) {
		backend, _ := NewWaldurBackend(testConfig())

		// Create multiple artifacts
		addresses := make([]*ContentAddress, 3)
		for i := 0; i < 3; i++ {
			data := []byte{byte(i), byte(i + 1), byte(i + 2)}
			req := &PutRequest{
				Data: data,
				EncryptionMetadata: &EncryptionMetadata{
					AlgorithmID:     "X25519",
					RecipientKeyIDs: []string{"key1"},
					EnvelopeHash:    make([]byte, 32),
				},
				Owner:        "cosmos1batch",
				ArtifactType: "test_artifact",
			}
			resp, _ := backend.Put(ctx, req)
			addresses[i] = resp.ContentAddress
		}

		batchChecker := NewBatchIntegrityChecker(backend, 2)
		results, err := batchChecker.VerifyBatch(ctx, addresses)

		if err != nil {
			t.Fatalf("batch verify error: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		for i, result := range results {
			if !result.Valid {
				t.Errorf("result %d should be valid", i)
			}
		}
	})
}

func TestSummarizeBatch(t *testing.T) {
	results := []*IntegrityCheckResult{
		{Valid: true, BytesVerified: 100},
		{Valid: true, BytesVerified: 200},
		{Valid: false, Error: "hash mismatch", ExpectedHash: "abc", ComputedHash: "def"},
		nil,
	}

	summary := SummarizeBatch(results)

	if summary.TotalChecked != 4 {
		t.Errorf("expected 4 checked, got %d", summary.TotalChecked)
	}
	if summary.ValidCount != 2 {
		t.Errorf("expected 2 valid, got %d", summary.ValidCount)
	}
	if summary.InvalidCount != 1 {
		t.Errorf("expected 1 invalid, got %d", summary.InvalidCount)
	}
	if summary.ErrorCount != 1 {
		t.Errorf("expected 1 error (nil result), got %d", summary.ErrorCount)
	}
	if summary.TotalBytesVerified != 300 {
		t.Errorf("expected 300 bytes, got %d", summary.TotalBytesVerified)
	}
}

func TestHashFunctions(t *testing.T) {
	t.Run("ComputeHash", func(t *testing.T) {
		data := []byte("test data")
		expected := sha256.Sum256(data)

		result := ComputeHash(data)
		if !bytes.Equal(result, expected[:]) {
			t.Error("hash mismatch")
		}
	})

	t.Run("ComputeHashHex", func(t *testing.T) {
		data := []byte("test data")
		expected := sha256.Sum256(data)
		expectedHex := hex.EncodeToString(expected[:])

		result := ComputeHashHex(data)
		if result != expectedHex {
			t.Errorf("expected %s, got %s", expectedHex, result)
		}
	})

	t.Run("VerifyHash", func(t *testing.T) {
		data := []byte("test data")
		hash := sha256.Sum256(data)

		if !VerifyHash(data, hash[:]) {
			t.Error("expected hash verification to pass")
		}

		wrongHash := make([]byte, 32)
		if VerifyHash(data, wrongHash) {
			t.Error("expected hash verification to fail")
		}
	})

	t.Run("VerifyHashHex", func(t *testing.T) {
		data := []byte("test data")
		hash := sha256.Sum256(data)
		hashHex := hex.EncodeToString(hash[:])

		if !VerifyHashHex(data, hashHex) {
			t.Error("expected hash verification to pass")
		}

		if VerifyHashHex(data, "invalid hex") {
			t.Error("expected invalid hex to fail")
		}
	})
}

func TestStreamingHasher(t *testing.T) {
	t.Run("BasicUsage", func(t *testing.T) {
		hasher := NewStreamingHasher()

		data1 := []byte("first part ")
		data2 := []byte("second part")

		_, _ = hasher.Write(data1)
		_, _ = hasher.Write(data2)

		fullData := make([]byte, 0, len(data1)+len(data2))
		fullData = append(fullData, data1...)
		fullData = append(fullData, data2...)
		expectedHash := sha256.Sum256(fullData)

		result := hasher.Sum()
		if !bytes.Equal(result, expectedHash[:]) {
			t.Error("streaming hash mismatch")
		}
	})

	t.Run("BytesWritten", func(t *testing.T) {
		hasher := NewStreamingHasher()

		_, _ = hasher.Write([]byte("12345"))
		_, _ = hasher.Write([]byte("67890"))

		if hasher.BytesWritten() != 10 {
			t.Errorf("expected 10 bytes, got %d", hasher.BytesWritten())
		}
	})

	t.Run("Reset", func(t *testing.T) {
		hasher := NewStreamingHasher()

		_, _ = hasher.Write([]byte("data"))
		hasher.Reset()

		if hasher.BytesWritten() != 0 {
			t.Error("expected 0 bytes after reset")
		}
	})

	t.Run("SumHex", func(t *testing.T) {
		hasher := NewStreamingHasher()
		_, _ = hasher.Write([]byte("test"))

		result := hasher.SumHex()
		if len(result) != 64 {
			t.Errorf("expected 64 hex chars, got %d", len(result))
		}
	})
}

func TestIntegrityCheckResult(t *testing.T) {
	result := &IntegrityCheckResult{
		Valid:         true,
		ExpectedHash:  "abc123",
		ComputedHash:  "abc123",
		BytesVerified: 1024,
	}

	if !result.Valid {
		t.Error("expected valid")
	}
}

func TestDefaultIntegrityCheckOptions(t *testing.T) {
	options := DefaultIntegrityCheckOptions()

	if !options.FullVerification {
		t.Error("expected full verification by default")
	}
	if options.PartialVerificationBytes != 1024*1024 {
		t.Error("unexpected partial verification size")
	}
	if !options.VerifyChunks {
		t.Error("expected verify chunks by default")
	}
	if !options.ReportCorruption {
		t.Error("expected report corruption by default")
	}
}
