package artifact_store

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"
)

// algorithmSHA256 is the SHA256 hash algorithm identifier
const algorithmSHA256 = "sha256"

func TestContentAddress(t *testing.T) {
	t.Run("NewContentAddress", func(t *testing.T) {
		content := []byte("test content for hashing")
		addr := NewContentAddress(content, BackendWaldur, "test/path/123")

		if addr.Version != ContentAddressVersion {
			t.Errorf("expected version %d, got %d", ContentAddressVersion, addr.Version)
		}

		expectedHash := sha256.Sum256(content)
		if !bytes.Equal(addr.Hash, expectedHash[:]) {
			t.Error("hash mismatch")
		}

		if addr.Algorithm != algorithmSHA256 {
			t.Errorf("expected algorithm sha256, got %s", addr.Algorithm)
		}

		if addr.Size != uint64(len(content)) {
			t.Errorf("expected size %d, got %d", len(content), addr.Size)
		}

		if addr.Backend != BackendWaldur {
			t.Errorf("expected backend %s, got %s", BackendWaldur, addr.Backend)
		}

		if addr.BackendRef != "test/path/123" {
			t.Errorf("expected backend_ref test/path/123, got %s", addr.BackendRef)
		}
	})

	t.Run("Validate_Valid", func(t *testing.T) {
		addr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		if err := addr.Validate(); err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}
	})

	t.Run("Validate_InvalidHash", func(t *testing.T) {
		addr := &ContentAddress{
			Version:    1,
			Hash:       []byte("short"),
			Algorithm:  algorithmSHA256,
			Backend:    BackendWaldur,
			BackendRef: "ref",
		}
		if err := addr.Validate(); err == nil {
			t.Error("expected error for invalid hash length")
		}
	})

	t.Run("Validate_EmptyBackendRef", func(t *testing.T) {
		content := []byte("test")
		hash := sha256.Sum256(content)
		addr := &ContentAddress{
			Version:    1,
			Hash:       hash[:],
			Algorithm:  algorithmSHA256,
			Backend:    BackendWaldur,
			BackendRef: "",
		}
		if err := addr.Validate(); err == nil {
			t.Error("expected error for empty backend_ref")
		}
	})

	t.Run("Equals", func(t *testing.T) {
		addr1 := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		addr2 := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		addr3 := NewContentAddress([]byte("different"), BackendWaldur, "ref")

		if !addr1.Equals(addr2) {
			t.Error("expected equal addresses to be equal")
		}

		if addr1.Equals(addr3) {
			t.Error("expected different addresses to not be equal")
		}
	})

	t.Run("HashHex", func(t *testing.T) {
		addr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		hexStr := addr.HashHex()
		if len(hexStr) != 64 { // 32 bytes * 2 hex chars
			t.Errorf("expected hex length 64, got %d", len(hexStr))
		}
	})
}

func TestChunkManifest(t *testing.T) {
	t.Run("NewChunkManifest", func(t *testing.T) {
		manifest := NewChunkManifest(1024, 256)

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
		manifest := NewChunkManifest(512, 256)

		hash1 := sha256.Sum256([]byte("chunk1"))
		err := manifest.AddChunk(ChunkInfo{
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
		err = manifest.AddChunk(ChunkInfo{
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
		manifest := NewChunkManifest(512, 256)

		hash := sha256.Sum256([]byte("chunk"))
		err := manifest.AddChunk(ChunkInfo{
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
		manifest := NewChunkManifest(512, 256)

		hash1 := sha256.Sum256([]byte("chunk1"))
		_ = manifest.AddChunk(ChunkInfo{Index: 0, Hash: hash1[:], Size: 256, Offset: 0, BackendRef: "cid1"})

		hash2 := sha256.Sum256([]byte("chunk2"))
		_ = manifest.AddChunk(ChunkInfo{Index: 1, Hash: hash2[:], Size: 256, Offset: 256, BackendRef: "cid2"})

		rootHash := manifest.ComputeRootHash()
		if len(rootHash) != 32 {
			t.Errorf("expected root_hash length 32, got %d", len(rootHash))
		}

		if !bytes.Equal(rootHash, manifest.RootHash) {
			t.Error("root hash not set correctly")
		}
	})

	t.Run("Validate_Valid", func(t *testing.T) {
		manifest := NewChunkManifest(512, 256)

		hash1 := sha256.Sum256([]byte("chunk1"))
		_ = manifest.AddChunk(ChunkInfo{Index: 0, Hash: hash1[:], Size: 256, Offset: 0, BackendRef: "cid1"})

		hash2 := sha256.Sum256([]byte("chunk2"))
		_ = manifest.AddChunk(ChunkInfo{Index: 1, Hash: hash2[:], Size: 256, Offset: 256, BackendRef: "cid2"})

		manifest.ComputeRootHash()

		if err := manifest.Validate(); err != nil {
			t.Errorf("expected valid manifest, got error: %v", err)
		}
	})

	t.Run("Validate_InvalidChunkCount", func(t *testing.T) {
		manifest := NewChunkManifest(512, 256)
		manifest.RootHash = make([]byte, 32)
		// No chunks added

		if err := manifest.Validate(); err == nil {
			t.Error("expected error for chunk count mismatch")
		}
	})
}

func TestRetentionTag(t *testing.T) {
	t.Run("NewRetentionTag", func(t *testing.T) {
		tag := NewRetentionTag("policy-1", "cosmos1abc", true)

		if tag.PolicyID != "policy-1" {
			t.Errorf("expected policy_id policy-1, got %s", tag.PolicyID)
		}

		if tag.Owner != "cosmos1abc" {
			t.Errorf("expected owner cosmos1abc, got %s", tag.Owner)
		}

		if !tag.DeleteOnExpiry {
			t.Error("expected delete_on_expiry to be true")
		}
	})

	t.Run("SetExpiration", func(t *testing.T) {
		tag := NewRetentionTag("policy-1", "cosmos1abc", true)
		expiresAt := time.Now().Add(24 * time.Hour)
		tag.SetExpiration(expiresAt)

		if tag.ExpiresAt == nil {
			t.Error("expected expires_at to be set")
		}

		if !tag.ExpiresAt.Equal(expiresAt) {
			t.Error("expires_at mismatch")
		}
	})

	t.Run("IsExpired", func(t *testing.T) {
		tag := NewRetentionTag("policy-1", "cosmos1abc", true)

		// Not expired if no expiration set
		if tag.IsExpired(time.Now()) {
			t.Error("expected not expired when no expiration set")
		}

		// Set past expiration
		pastTime := time.Now().Add(-1 * time.Hour)
		tag.SetExpiration(pastTime)

		if !tag.IsExpired(time.Now()) {
			t.Error("expected expired for past expiration")
		}

		// Set future expiration
		futureTime := time.Now().Add(1 * time.Hour)
		tag.SetExpiration(futureTime)

		if tag.IsExpired(time.Now()) {
			t.Error("expected not expired for future expiration")
		}
	})

	t.Run("IsExpiredAtBlock", func(t *testing.T) {
		tag := NewRetentionTag("policy-1", "cosmos1abc", true)

		// Not expired if no block set
		if tag.IsExpiredAtBlock(100) {
			t.Error("expected not expired when no block set")
		}

		// Set expiration block
		tag.SetExpirationBlock(50)

		if !tag.IsExpiredAtBlock(50) {
			t.Error("expected expired at block 50")
		}

		if !tag.IsExpiredAtBlock(100) {
			t.Error("expected expired at block 100")
		}

		if tag.IsExpiredAtBlock(40) {
			t.Error("expected not expired at block 40")
		}
	})
}

func TestArtifactReference(t *testing.T) {
	t.Run("NewArtifactReference", func(t *testing.T) {
		contentAddr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		encMeta := &EncryptionMetadata{
			AlgorithmID:     "X25519-XSALSA20-POLY1305",
			RecipientKeyIDs: []string{"key1"},
			EnvelopeHash:    make([]byte, 32),
			SenderKeyID:     "sender",
		}

		ref := NewArtifactReference(
			"ref-123",
			contentAddr,
			encMeta,
			"cosmos1abc",
			"face_embedding",
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

		if ref.CreatedAtBlock != 100 {
			t.Errorf("expected created_at_block 100, got %d", ref.CreatedAtBlock)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		contentAddr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		encMeta := &EncryptionMetadata{
			AlgorithmID:     "X25519-XSALSA20-POLY1305",
			RecipientKeyIDs: []string{"key1"},
			EnvelopeHash:    make([]byte, 32),
			SenderKeyID:     "sender",
		}

		ref := NewArtifactReference(
			"ref-123",
			contentAddr,
			encMeta,
			"cosmos1abc",
			"face_embedding",
			100,
		)

		if err := ref.Validate(); err != nil {
			t.Errorf("expected valid reference, got error: %v", err)
		}
	})

	t.Run("SetMetadata", func(t *testing.T) {
		contentAddr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		encMeta := &EncryptionMetadata{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{"key1"},
			EnvelopeHash:    make([]byte, 32),
		}

		ref := NewArtifactReference("ref-123", contentAddr, encMeta, "cosmos1abc", "face_embedding", 100)
		ref.SetMetadata("key1", "value1")

		val, ok := ref.GetMetadata("key1")
		if !ok {
			t.Error("expected metadata key to exist")
		}
		if val != "value1" {
			t.Errorf("expected value1, got %s", val)
		}

		_, ok = ref.GetMetadata("nonexistent")
		if ok {
			t.Error("expected nonexistent key to not exist")
		}
	})

	t.Run("IsChunked", func(t *testing.T) {
		contentAddr := NewContentAddress([]byte("test"), BackendIPFS, "cid")
		encMeta := &EncryptionMetadata{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{"key1"},
			EnvelopeHash:    make([]byte, 32),
		}

		ref := NewArtifactReference("ref-123", contentAddr, encMeta, "cosmos1abc", "face_embedding", 100)

		if ref.IsChunked() {
			t.Error("expected not chunked without manifest")
		}

		manifest := NewChunkManifest(512, 256)
		hash1 := sha256.Sum256([]byte("chunk1"))
		_ = manifest.AddChunk(ChunkInfo{Index: 0, Hash: hash1[:], Size: 256, Offset: 0, BackendRef: "cid1"})
		hash2 := sha256.Sum256([]byte("chunk2"))
		_ = manifest.AddChunk(ChunkInfo{Index: 1, Hash: hash2[:], Size: 256, Offset: 256, BackendRef: "cid2"})
		manifest.ComputeRootHash()

		ref.SetChunkManifest(manifest)

		if !ref.IsChunked() {
			t.Error("expected chunked with manifest")
		}
	})
}

func TestBackendType(t *testing.T) {
	t.Run("IsValid", func(t *testing.T) {
		if !BackendWaldur.IsValid() {
			t.Error("expected waldur to be valid")
		}

		if !BackendIPFS.IsValid() {
			t.Error("expected ipfs to be valid")
		}

		invalid := BackendType("invalid")
		if invalid.IsValid() {
			t.Error("expected invalid backend to be invalid")
		}
	})

	t.Run("String", func(t *testing.T) {
		if BackendWaldur.String() != "waldur" {
			t.Errorf("expected waldur, got %s", BackendWaldur.String())
		}

		if BackendIPFS.String() != "ipfs" {
			t.Errorf("expected ipfs, got %s", BackendIPFS.String())
		}
	})
}

func TestPutRequest(t *testing.T) {
	t.Run("Validate_Valid", func(t *testing.T) {
		req := &PutRequest{
			Data: []byte("test data"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1abc",
			ArtifactType: "face_embedding",
		}

		if err := req.Validate(); err != nil {
			t.Errorf("expected valid request, got error: %v", err)
		}
	})

	t.Run("Validate_EmptyData", func(t *testing.T) {
		req := &PutRequest{
			Data: nil,
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "cosmos1abc",
			ArtifactType: "face_embedding",
		}

		if err := req.Validate(); err == nil {
			t.Error("expected error for empty data")
		}
	})

	t.Run("Validate_EmptyOwner", func(t *testing.T) {
		req := &PutRequest{
			Data: []byte("test"),
			EncryptionMetadata: &EncryptionMetadata{
				AlgorithmID:     "X25519",
				RecipientKeyIDs: []string{"key1"},
				EnvelopeHash:    make([]byte, 32),
			},
			Owner:        "",
			ArtifactType: "face_embedding",
		}

		if err := req.Validate(); err == nil {
			t.Error("expected error for empty owner")
		}
	})
}

func TestChunkData(t *testing.T) {
	t.Run("Verify_Valid", func(t *testing.T) {
		data := []byte("chunk data")
		hash := sha256.Sum256(data)

		chunk := &ChunkData{
			Index: 0,
			Data:  data,
			Hash:  hash[:],
		}

		if err := chunk.Verify(); err != nil {
			t.Errorf("expected valid chunk, got error: %v", err)
		}
	})

	t.Run("Verify_Invalid", func(t *testing.T) {
		data := []byte("chunk data")
		wrongHash := sha256.Sum256([]byte("different data"))

		chunk := &ChunkData{
			Index: 0,
			Data:  data,
			Hash:  wrongHash[:],
		}

		if err := chunk.Verify(); err == nil {
			t.Error("expected error for hash mismatch")
		}
	})

	t.Run("Verify_Nil", func(t *testing.T) {
		var chunk *ChunkData
		if err := chunk.Verify(); err == nil {
			t.Error("expected error for nil chunk")
		}
	})
}

func TestGetRequest(t *testing.T) {
	t.Run("Validate_Valid", func(t *testing.T) {
		addr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		req := &GetRequest{
			ContentAddress: addr,
		}

		if err := req.Validate(); err != nil {
			t.Errorf("expected valid request, got error: %v", err)
		}
	})

	t.Run("Validate_NilAddress", func(t *testing.T) {
		req := &GetRequest{
			ContentAddress: nil,
		}

		if err := req.Validate(); err == nil {
			t.Error("expected error for nil address")
		}
	})
}

func TestDeleteRequest(t *testing.T) {
	t.Run("Validate_Valid", func(t *testing.T) {
		addr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		req := &DeleteRequest{
			ContentAddress:    addr,
			RequestingAccount: "cosmos1abc",
		}

		if err := req.Validate(); err != nil {
			t.Errorf("expected valid request, got error: %v", err)
		}
	})

	t.Run("Validate_EmptyRequestingAccount", func(t *testing.T) {
		addr := NewContentAddress([]byte("test"), BackendWaldur, "ref")
		req := &DeleteRequest{
			ContentAddress:    addr,
			RequestingAccount: "",
		}

		if err := req.Validate(); err == nil {
			t.Error("expected error for empty requesting_account")
		}
	})
}
