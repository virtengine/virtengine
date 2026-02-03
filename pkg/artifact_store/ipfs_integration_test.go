package artifact_store

import (
	"context"
	"os"
	"testing"
	"time"
)

// Integration tests for real IPFS backend
// These tests require a running IPFS node
// Set IPFS_TEST_ENDPOINT environment variable to run (e.g., "localhost:5001")

func TestIPFSIntegration_StoreRetrieve(t *testing.T) {
	endpoint := os.Getenv("IPFS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("IPFS_TEST_ENDPOINT not set - skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := ProductionIPFSConfig(endpoint)
	backend, err := NewIPFSBackend(config)
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}

	// Test data
	testData := []byte("encrypted artifact data for IPFS integration test")

	req := &PutRequest{
		Data: testData,
		EncryptionMetadata: &EncryptionMetadata{
			AlgorithmID:     "X25519-XSALSA20-POLY1305",
			RecipientKeyIDs: []string{"test-key-1"},
			EnvelopeHash:    make([]byte, 32),
		},
		Owner:        "cosmos1integrationtest",
		ArtifactType: "face_embedding",
	}

	// Put
	resp, err := backend.Put(ctx, req)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if resp.ContentAddress == nil {
		t.Fatal("expected content address")
	}

	t.Logf("Stored artifact with CID: %s", resp.ContentAddress.BackendRef)

	// Verify it exists
	exists, err := backend.Exists(ctx, resp.ContentAddress)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Error("artifact should exist")
	}

	// Get
	getResp, err := backend.Get(ctx, &GetRequest{
		ContentAddress: resp.ContentAddress,
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(getResp.Data) != string(testData) {
		t.Errorf("data mismatch: got %q, want %q", getResp.Data, testData)
	}

	// Cleanup - Delete (unpin)
	err = backend.Delete(ctx, &DeleteRequest{
		ContentAddress:    resp.ContentAddress,
		RequestingAccount: "cosmos1integrationtest",
	})
	if err != nil {
		t.Logf("Delete (unpin) warning: %v", err)
	}
}

func TestIPFSIntegration_ChunkedStorage(t *testing.T) {
	endpoint := os.Getenv("IPFS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("IPFS_TEST_ENDPOINT not set - skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	config := ProductionIPFSConfig(endpoint)
	backend, err := NewIPFSBackend(config)
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}

	// Create test data spanning multiple chunks
	testData := make([]byte, 1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	meta := &PutMetadata{
		EncryptionMetadata: &EncryptionMetadata{
			AlgorithmID:     "X25519-XSALSA20-POLY1305",
			RecipientKeyIDs: []string{"test-key-1"},
			EnvelopeHash:    make([]byte, 32),
		},
		Owner:        "cosmos1chunkedtest",
		ArtifactType: "raw_image",
	}

	// Store chunked (256 byte chunks = 4 chunks)
	resp, manifest, err := backend.PutChunked(ctx, testData, 256, meta)
	if err != nil {
		t.Fatalf("PutChunked failed: %v", err)
	}

	if manifest == nil {
		t.Fatal("expected manifest")
	}

	if manifest.ChunkCount != 4 {
		t.Errorf("expected 4 chunks, got %d", manifest.ChunkCount)
	}

	t.Logf("Stored %d chunks with manifest CID: %s", manifest.ChunkCount, resp.ContentAddress.BackendRef)

	// Retrieve chunked
	retrieved, err := backend.GetChunked(ctx, manifest)
	if err != nil {
		t.Fatalf("GetChunked failed: %v", err)
	}

	if len(retrieved) != len(testData) {
		t.Errorf("retrieved size mismatch: got %d, want %d", len(retrieved), len(testData))
	}

	for i := range testData {
		if retrieved[i] != testData[i] {
			t.Errorf("data mismatch at byte %d: got %d, want %d", i, retrieved[i], testData[i])
			break
		}
	}

	// Verify chunks
	err = backend.VerifyChunks(ctx, manifest)
	if err != nil {
		t.Errorf("VerifyChunks failed: %v", err)
	}

	// Cleanup
	err = backend.Delete(ctx, &DeleteRequest{
		ContentAddress:    resp.ContentAddress,
		RequestingAccount: "cosmos1chunkedtest",
	})
	if err != nil {
		t.Logf("Delete warning: %v", err)
	}
}

func TestIPFSIntegration_Health(t *testing.T) {
	endpoint := os.Getenv("IPFS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("IPFS_TEST_ENDPOINT not set - skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config := ProductionIPFSConfig(endpoint)
	backend, err := NewIPFSBackend(config)
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}

	err = backend.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	metrics, err := backend.GetMetrics(ctx)
	if err != nil {
		t.Errorf("GetMetrics failed: %v", err)
	}

	if metrics.BackendStatus["ipfs_version"] == "" {
		t.Error("expected IPFS version in metrics")
	}

	t.Logf("IPFS version: %s", metrics.BackendStatus["ipfs_version"])
}

func TestIPFSIntegration_Pin_Unpin(t *testing.T) {
	endpoint := os.Getenv("IPFS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("IPFS_TEST_ENDPOINT not set - skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create real IPFS client directly for pin testing
	clientConfig := &RealIPFSClientConfig{
		Endpoint:   endpoint,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	client, err := NewRealIPFSClient(clientConfig)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	testData := []byte("test data for pin/unpin")

	// Add data
	cid, err := client.Add(ctx, testData)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	t.Logf("Added with CID: %s", cid)

	// Pin
	err = client.Pin(ctx, cid)
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	// Check if pinned
	pinned, err := client.IsPinned(ctx, cid)
	if err != nil {
		t.Fatalf("IsPinned failed: %v", err)
	}
	if !pinned {
		t.Error("expected CID to be pinned")
	}

	// Unpin
	err = client.Unpin(ctx, cid)
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	// Check no longer pinned
	pinned, err = client.IsPinned(ctx, cid)
	if err != nil {
		t.Fatalf("IsPinned failed after unpin: %v", err)
	}
	if pinned {
		t.Error("expected CID to not be pinned after unpin")
	}
}

func TestIPFSIntegration_StreamingBackend(t *testing.T) {
	endpoint := os.Getenv("IPFS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("IPFS_TEST_ENDPOINT not set - skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := ProductionIPFSConfig(endpoint)
	backend, err := NewIPFSStreamingBackend(config)
	if err != nil {
		t.Fatalf("failed to create streaming backend: %v", err)
	}

	// Store data
	testData := []byte("streaming test data for IPFS")
	req := &PutRequest{
		Data: testData,
		EncryptionMetadata: &EncryptionMetadata{
			AlgorithmID:     "X25519-XSALSA20-POLY1305",
			RecipientKeyIDs: []string{"test-key-1"},
			EnvelopeHash:    make([]byte, 32),
		},
		Owner:        "cosmos1streamtest",
		ArtifactType: "document",
	}

	resp, err := backend.Put(ctx, req)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get as stream
	stream, err := backend.GetStream(ctx, resp.ContentAddress)
	if err != nil {
		t.Fatalf("GetStream failed: %v", err)
	}
	defer stream.Close()

	// Read from stream
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Read failed: %v", err)
	}

	if string(buf[:n]) != string(testData) {
		t.Errorf("stream data mismatch: got %q, want %q", buf[:n], testData)
	}

	// Cleanup
	_ = backend.Delete(ctx, &DeleteRequest{
		ContentAddress:    resp.ContentAddress,
		RequestingAccount: "cosmos1streamtest",
	})
}
