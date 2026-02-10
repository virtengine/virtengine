package types

import (
	"crypto/sha256"
	"testing"
	"time"
)

func TestContentAddressReference(t *testing.T) {
	t.Run("NewContentAddressReference", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))

		ref := NewContentAddressReference(
			origHash[:],
			encHash[:],
			1024,
			1100,
			StorageBackendWaldur,
			"waldur/path/123",
		)

		if ref.Version != ArtifactReferenceVersion {
			t.Errorf("expected version %d, got %d", ArtifactReferenceVersion, ref.Version)
		}

		if ref.Size != 1024 {
			t.Errorf("expected size 1024, got %d", ref.Size)
		}

		if ref.EncryptedSize != 1100 {
			t.Errorf("expected encrypted_size 1100, got %d", ref.EncryptedSize)
		}

		if ref.Backend != StorageBackendWaldur {
			t.Errorf("expected backend %s, got %s", StorageBackendWaldur, ref.Backend)
		}

		if ref.BackendRef != "waldur/path/123" {
			t.Errorf("expected backend_ref waldur/path/123, got %s", ref.BackendRef)
		}
	})

	t.Run("Validate_Valid", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))

		ref := NewContentAddressReference(
			origHash[:],
			encHash[:],
			1024,
			1100,
			StorageBackendIPFS,
			"QmXYZ123",
		)

		if err := ref.Validate(); err != nil {
			t.Errorf("expected valid reference, got error: %v", err)
		}
	})

	t.Run("Validate_InvalidHash", func(t *testing.T) {
		encHash := sha256.Sum256([]byte("encrypted"))

		ref := &ContentAddressReference{
			Version:       1,
			Hash:          []byte("short"),
			EncryptedHash: encHash[:],
			Algorithm:     "sha256",
			Size:          1024,
			EncryptedSize: 1100,
			Backend:       StorageBackendWaldur,
			BackendRef:    "ref",
		}

		if err := ref.Validate(); err == nil {
			t.Error("expected error for invalid hash length")
		}
	})

	t.Run("Validate_InvalidEncryptedHash", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))

		ref := &ContentAddressReference{
			Version:       1,
			Hash:          origHash[:],
			EncryptedHash: []byte("short"),
			Algorithm:     "sha256",
			Size:          1024,
			EncryptedSize: 1100,
			Backend:       StorageBackendWaldur,
			BackendRef:    "ref",
		}

		if err := ref.Validate(); err == nil {
			t.Error("expected error for invalid encrypted_hash length")
		}
	})

	t.Run("Validate_InvalidBackend", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))

		ref := &ContentAddressReference{
			Version:       1,
			Hash:          origHash[:],
			EncryptedHash: encHash[:],
			Algorithm:     "sha256",
			Size:          1024,
			EncryptedSize: 1100,
			Backend:       StorageBackend("invalid"),
			BackendRef:    "ref",
		}

		if err := ref.Validate(); err == nil {
			t.Error("expected error for invalid backend")
		}
	})

	t.Run("HashHex", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("test"))
		encHash := sha256.Sum256([]byte("encrypted"))

		ref := NewContentAddressReference(origHash[:], encHash[:], 100, 110, StorageBackendWaldur, "ref")

		hexStr := ref.HashHex()
		if len(hexStr) != 64 {
			t.Errorf("expected hex length 64, got %d", len(hexStr))
		}

		encHexStr := ref.EncryptedHashHex()
		if len(encHexStr) != 64 {
			t.Errorf("expected encrypted hex length 64, got %d", len(encHexStr))
		}
	})
}

func TestChunkManifestReference(t *testing.T) {
	t.Run("NewChunkManifestReference", func(t *testing.T) {
		manifest := NewChunkManifestReference(1024, 256)

		if manifest.Version != ChunkManifestVersion {
			t.Errorf("expected version %d, got %d", ChunkManifestVersion, manifest.Version)
		}

		if manifest.TotalSize != 1024 {
			t.Errorf("expected total_size 1024, got %d", manifest.TotalSize)
		}

		if manifest.ChunkSize != 256 {
			t.Errorf("expected chunk_size 256, got %d", manifest.ChunkSize)
		}

		if manifest.ChunkCount != 4 {
			t.Errorf("expected chunk_count 4, got %d", manifest.ChunkCount)
		}
	})

	t.Run("AddChunk", func(t *testing.T) {
		manifest := NewChunkManifestReference(512, 256)

		hash1 := sha256.Sum256([]byte("chunk1"))
		err := manifest.AddChunk(ChunkReference{
			Index:      0,
			Hash:       hash1[:],
			Size:       256,
			Offset:     0,
			BackendRef: "cid1",
		})
		if err != nil {
			t.Errorf("unexpected error adding chunk: %v", err)
		}

		hash2 := sha256.Sum256([]byte("chunk2"))
		err = manifest.AddChunk(ChunkReference{
			Index:      1,
			Hash:       hash2[:],
			Size:       256,
			Offset:     256,
			BackendRef: "cid2",
		})
		if err != nil {
			t.Errorf("unexpected error adding chunk: %v", err)
		}

		if len(manifest.Chunks) != 2 {
			t.Errorf("expected 2 chunks, got %d", len(manifest.Chunks))
		}
	})

	t.Run("AddChunk_WrongIndex", func(t *testing.T) {
		manifest := NewChunkManifestReference(512, 256)

		hash := sha256.Sum256([]byte("chunk"))
		err := manifest.AddChunk(ChunkReference{
			Index:      1, // Should be 0
			Hash:       hash[:],
			Size:       256,
			Offset:     0,
			BackendRef: "cid",
		})
		if err == nil {
			t.Error("expected error for wrong chunk index")
		}
	})

	t.Run("ComputeRootHash", func(t *testing.T) {
		manifest := NewChunkManifestReference(512, 256)

		hash1 := sha256.Sum256([]byte("chunk1"))
		_ = manifest.AddChunk(ChunkReference{Index: 0, Hash: hash1[:], Size: 256, Offset: 0, BackendRef: "cid1"})

		hash2 := sha256.Sum256([]byte("chunk2"))
		_ = manifest.AddChunk(ChunkReference{Index: 1, Hash: hash2[:], Size: 256, Offset: 256, BackendRef: "cid2"})

		rootHash := manifest.ComputeRootHash()
		if len(rootHash) != 32 {
			t.Errorf("expected root_hash length 32, got %d", len(rootHash))
		}
	})

	t.Run("Validate_Valid", func(t *testing.T) {
		manifest := NewChunkManifestReference(512, 256)

		hash1 := sha256.Sum256([]byte("chunk1"))
		_ = manifest.AddChunk(ChunkReference{Index: 0, Hash: hash1[:], Size: 256, Offset: 0, BackendRef: "cid1"})

		hash2 := sha256.Sum256([]byte("chunk2"))
		_ = manifest.AddChunk(ChunkReference{Index: 1, Hash: hash2[:], Size: 256, Offset: 256, BackendRef: "cid2"})

		manifest.ComputeRootHash()

		if err := manifest.Validate(); err != nil {
			t.Errorf("expected valid manifest, got error: %v", err)
		}
	})

	t.Run("Validate_ChunkSizeMismatch", func(t *testing.T) {
		manifest := NewChunkManifestReference(512, 256)

		hash1 := sha256.Sum256([]byte("chunk1"))
		_ = manifest.AddChunk(ChunkReference{Index: 0, Hash: hash1[:], Size: 256, Offset: 0, BackendRef: "cid1"})

		hash2 := sha256.Sum256([]byte("chunk2"))
		_ = manifest.AddChunk(ChunkReference{Index: 1, Hash: hash2[:], Size: 200, Offset: 256, BackendRef: "cid2"}) // Wrong size

		manifest.ComputeRootHash()

		if err := manifest.Validate(); err == nil {
			t.Error("expected error for chunk size mismatch")
		}
	})
}

func TestEncryptionEnvelopeMetadata(t *testing.T) {
	t.Run("NewEncryptionEnvelopeMetadata", func(t *testing.T) {
		hash := sha256.Sum256([]byte("envelope"))
		meta := NewEncryptionEnvelopeMetadata(
			"X25519-XSALSA20-POLY1305",
			[]string{"key1", "key2"},
			hash[:],
			"sender-key",
		)

		if meta.AlgorithmID != "X25519-XSALSA20-POLY1305" {
			t.Errorf("expected algorithm X25519-XSALSA20-POLY1305, got %s", meta.AlgorithmID)
		}

		if len(meta.RecipientKeyIDs) != 2 {
			t.Errorf("expected 2 recipient keys, got %d", len(meta.RecipientKeyIDs))
		}

		if meta.SenderKeyID != "sender-key" {
			t.Errorf("expected sender_key_id sender-key, got %s", meta.SenderKeyID)
		}
	})

	t.Run("Validate_Valid", func(t *testing.T) {
		hash := sha256.Sum256([]byte("envelope"))
		meta := NewEncryptionEnvelopeMetadata(
			"X25519",
			[]string{"key1"},
			hash[:],
			"sender",
		)

		if err := meta.Validate(); err != nil {
			t.Errorf("expected valid metadata, got error: %v", err)
		}
	})

	t.Run("Validate_EmptyAlgorithm", func(t *testing.T) {
		hash := sha256.Sum256([]byte("envelope"))
		meta := &EncryptionEnvelopeMetadata{
			AlgorithmID:     "",
			RecipientKeyIDs: []string{"key1"},
			EnvelopeHash:    hash[:],
		}

		if err := meta.Validate(); err == nil {
			t.Error("expected error for empty algorithm")
		}
	})

	t.Run("Validate_EmptyRecipients", func(t *testing.T) {
		hash := sha256.Sum256([]byte("envelope"))
		meta := &EncryptionEnvelopeMetadata{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{},
			EnvelopeHash:    hash[:],
		}

		if err := meta.Validate(); err == nil {
			t.Error("expected error for empty recipients")
		}
	})

	t.Run("Validate_InvalidEnvelopeHash", func(t *testing.T) {
		meta := &EncryptionEnvelopeMetadata{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{"key1"},
			EnvelopeHash:    []byte("short"),
		}

		if err := meta.Validate(); err == nil {
			t.Error("expected error for invalid envelope hash")
		}
	})
}

func TestIdentityArtifactReference(t *testing.T) {
	t.Run("NewIdentityArtifactReference", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendIPFS, "QmXYZ")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		now := time.Now().UTC()
		ref := NewIdentityArtifactReference(
			"ref-123",
			"cosmos1abc",
			ArtifactTypeFaceEmbedding,
			contentAddr,
			encMeta,
			now,
			100,
		)

		if ref.Version != ArtifactReferenceVersion {
			t.Errorf("expected version %d, got %d", ArtifactReferenceVersion, ref.Version)
		}

		if ref.ReferenceID != "ref-123" {
			t.Errorf("expected reference_id ref-123, got %s", ref.ReferenceID)
		}

		if ref.AccountAddress != "cosmos1abc" {
			t.Errorf("expected account_address cosmos1abc, got %s", ref.AccountAddress)
		}

		if ref.ArtifactType != ArtifactTypeFaceEmbedding {
			t.Errorf("expected artifact_type %s, got %s", ArtifactTypeFaceEmbedding, ref.ArtifactType)
		}

		if ref.CreatedAtBlock != 100 {
			t.Errorf("expected created_at_block 100, got %d", ref.CreatedAtBlock)
		}

		if ref.Revoked {
			t.Error("expected not revoked")
		}
	})

	t.Run("Validate_Valid", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendWaldur, "path")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		ref := NewIdentityArtifactReference(
			"ref-123",
			"cosmos1abc",
			ArtifactTypeRawImage,
			contentAddr,
			encMeta,
			time.Now(),
			100,
		)

		if err := ref.Validate(); err != nil {
			t.Errorf("expected valid reference, got error: %v", err)
		}
	})

	t.Run("Validate_EmptyReferenceID", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendWaldur, "path")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		ref := NewIdentityArtifactReference(
			"",
			"cosmos1abc",
			ArtifactTypeRawImage,
			contentAddr,
			encMeta,
			time.Now(),
			100,
		)

		if err := ref.Validate(); err == nil {
			t.Error("expected error for empty reference_id")
		}
	})

	t.Run("Validate_InvalidArtifactType", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendWaldur, "path")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		ref := &IdentityArtifactReference{
			Version:            1,
			ReferenceID:        "ref-123",
			AccountAddress:     "cosmos1abc",
			ArtifactType:       ArtifactType("invalid"),
			ContentAddress:     contentAddr,
			EncryptionMetadata: encMeta,
		}

		if err := ref.Validate(); err == nil {
			t.Error("expected error for invalid artifact_type")
		}
	})

	t.Run("SetMetadata_GetMetadata", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendIPFS, "cid")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		ref := NewIdentityArtifactReference("ref-123", "cosmos1abc", ArtifactTypeFaceEmbedding, contentAddr, encMeta, time.Now(), 100)

		ref.SetMetadata("key1", "value1")
		ref.SetMetadata("key2", "value2")

		val1, ok1 := ref.GetMetadata("key1")
		if !ok1 || val1 != "value1" {
			t.Errorf("expected value1, got %s (found=%v)", val1, ok1)
		}

		val2, ok2 := ref.GetMetadata("key2")
		if !ok2 || val2 != "value2" {
			t.Errorf("expected value2, got %s (found=%v)", val2, ok2)
		}

		_, ok3 := ref.GetMetadata("nonexistent")
		if ok3 {
			t.Error("expected nonexistent key to not be found")
		}
	})

	t.Run("Revoke", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendIPFS, "cid")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		ref := NewIdentityArtifactReference("ref-123", "cosmos1abc", ArtifactTypeFaceEmbedding, contentAddr, encMeta, time.Now(), 100)

		if ref.IsRevoked() {
			t.Error("expected not revoked initially")
		}

		revokeTime := time.Now().UTC()
		ref.Revoke("user requested revocation", revokeTime)

		if !ref.IsRevoked() {
			t.Error("expected revoked after Revoke()")
		}

		if ref.RevokedAt == nil {
			t.Error("expected revoked_at to be set")
		}

		if ref.RevokedReason != "user requested revocation" {
			t.Errorf("expected revoked_reason 'user requested revocation', got %s", ref.RevokedReason)
		}
	})

	t.Run("IsChunked", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendIPFS, "cid")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		ref := NewIdentityArtifactReference("ref-123", "cosmos1abc", ArtifactTypeFaceEmbedding, contentAddr, encMeta, time.Now(), 100)

		if ref.IsChunked() {
			t.Error("expected not chunked without manifest")
		}

		manifest := NewChunkManifestReference(512, 256)
		hash1 := sha256.Sum256([]byte("chunk1"))
		_ = manifest.AddChunk(ChunkReference{Index: 0, Hash: hash1[:], Size: 256, Offset: 0, BackendRef: "cid1"})
		hash2 := sha256.Sum256([]byte("chunk2"))
		_ = manifest.AddChunk(ChunkReference{Index: 1, Hash: hash2[:], Size: 256, Offset: 256, BackendRef: "cid2"})
		manifest.ComputeRootHash()

		ref.SetChunkManifest(manifest)

		if !ref.IsChunked() {
			t.Error("expected chunked with manifest")
		}
	})

	t.Run("SetRetentionPolicy", func(t *testing.T) {
		origHash := sha256.Sum256([]byte("original"))
		encHash := sha256.Sum256([]byte("encrypted"))
		contentAddr := NewContentAddressReference(origHash[:], encHash[:], 1024, 1100, StorageBackendWaldur, "path")

		envHash := sha256.Sum256([]byte("envelope"))
		encMeta := NewEncryptionEnvelopeMetadata("X25519", []string{"key1"}, envHash[:], "sender")

		ref := NewIdentityArtifactReference("ref-123", "cosmos1abc", ArtifactTypeFaceEmbedding, contentAddr, encMeta, time.Now(), 100)

		ref.SetRetentionPolicy("policy-30days")

		if ref.RetentionPolicyID != "policy-30days" {
			t.Errorf("expected retention_policy_id policy-30days, got %s", ref.RetentionPolicyID)
		}
	})
}

func TestStorageBackend(t *testing.T) {
	t.Run("IsValidStorageBackend", func(t *testing.T) {
		if !IsValidStorageBackend(StorageBackendWaldur) {
			t.Error("expected waldur to be valid")
		}

		if !IsValidStorageBackend(StorageBackendIPFS) {
			t.Error("expected ipfs to be valid")
		}

		if IsValidStorageBackend(StorageBackend("invalid")) {
			t.Error("expected invalid backend to be invalid")
		}
	})

	t.Run("AllStorageBackends", func(t *testing.T) {
		backends := AllStorageBackends()
		if len(backends) != 2 {
			t.Errorf("expected 2 backends, got %d", len(backends))
		}
	})
}

func TestChunkReference(t *testing.T) {
	t.Run("HashHex", func(t *testing.T) {
		hash := sha256.Sum256([]byte("chunk"))
		chunk := ChunkReference{
			Index:      0,
			Hash:       hash[:],
			Size:       256,
			Offset:     0,
			BackendRef: "cid",
		}

		hexStr := chunk.HashHex()
		if len(hexStr) != 64 {
			t.Errorf("expected hex length 64, got %d", len(hexStr))
		}
	})
}
