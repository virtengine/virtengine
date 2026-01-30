package artifact_store

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient defines the interface for IPFS operations
// This abstraction allows for both real IPFS and stub implementations
type IPFSClient interface {
	// Add adds data to IPFS and returns the CID
	Add(ctx context.Context, data []byte) (string, error)

	// Cat retrieves data by CID from IPFS
	Cat(ctx context.Context, cid string) ([]byte, error)

	// CatStream retrieves data as a stream by CID from IPFS
	CatStream(ctx context.Context, cid string) (io.ReadCloser, error)

	// Pin pins a CID to prevent garbage collection
	Pin(ctx context.Context, cid string) error

	// Unpin removes a pin, allowing garbage collection
	Unpin(ctx context.Context, cid string) error

	// IsPinned checks if a CID is pinned
	IsPinned(ctx context.Context, cid string) (bool, error)

	// Version returns the IPFS version info for health checks
	Version(ctx context.Context) (string, error)

	// IsHealthy checks if the IPFS connection is healthy
	IsHealthy(ctx context.Context) error
}

// RealIPFSClient implements IPFSClient using the go-ipfs-api shell
type RealIPFSClient struct {
	sh         *shell.Shell
	endpoint   string
	timeout    time.Duration
	maxRetries int
	mu         sync.RWMutex
}

// RealIPFSClientConfig contains configuration for the real IPFS client
type RealIPFSClientConfig struct {
	// Endpoint is the IPFS API endpoint (e.g., "localhost:5001")
	Endpoint string

	// Timeout is the request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int
}

// DefaultRealIPFSClientConfig returns default configuration
func DefaultRealIPFSClientConfig() *RealIPFSClientConfig {
	return &RealIPFSClientConfig{
		Endpoint:   "localhost:5001",
		Timeout:    60 * time.Second,
		MaxRetries: 3,
	}
}

// NewRealIPFSClient creates a new real IPFS client
func NewRealIPFSClient(config *RealIPFSClientConfig) (*RealIPFSClient, error) {
	if config == nil {
		config = DefaultRealIPFSClientConfig()
	}

	if config.Endpoint == "" {
		return nil, ErrInvalidInput.Wrap("ipfs endpoint cannot be empty")
	}

	sh := shell.NewShell(config.Endpoint)
	if config.Timeout > 0 {
		sh.SetTimeout(config.Timeout)
	}

	client := &RealIPFSClient{
		sh:         sh,
		endpoint:   config.Endpoint,
		timeout:    config.Timeout,
		maxRetries: config.MaxRetries,
	}

	// Test connectivity
	if err := client.IsHealthy(context.Background()); err != nil {
		return nil, fmt.Errorf("ipfs connection failed: %w", err)
	}

	return client, nil
}

// Add adds data to IPFS and returns the CID
func (c *RealIPFSClient) Add(ctx context.Context, data []byte) (string, error) {
	var cid string
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		cid, err = c.sh.Add(bytes.NewReader(data))
		if err == nil {
			return cid, nil
		}

		// Check if context is cancelled
		if ctx.Err() != nil {
			return "", fmt.Errorf("context cancelled: %w", ctx.Err())
		}
	}

	return "", fmt.Errorf("ipfs add failed after %d attempts: %w", c.maxRetries+1, err)
}

// Cat retrieves data by CID from IPFS
func (c *RealIPFSClient) Cat(ctx context.Context, cid string) ([]byte, error) {
	reader, err := c.CatStream(ctx, cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// CatStream retrieves data as a stream by CID from IPFS
func (c *RealIPFSClient) CatStream(ctx context.Context, cid string) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		reader, err = c.sh.Cat(cid)
		if err == nil {
			return reader, nil
		}

		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("ipfs cat failed after %d attempts: %w", c.maxRetries+1, err)
}

// Pin pins a CID to prevent garbage collection
func (c *RealIPFSClient) Pin(ctx context.Context, cid string) error {
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		err = c.sh.Pin(cid)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("ipfs pin failed after %d attempts: %w", c.maxRetries+1, err)
}

// Unpin removes a pin, allowing garbage collection
func (c *RealIPFSClient) Unpin(ctx context.Context, cid string) error {
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		err = c.sh.Unpin(cid)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("ipfs unpin failed after %d attempts: %w", c.maxRetries+1, err)
}

// IsPinned checks if a CID is pinned
func (c *RealIPFSClient) IsPinned(ctx context.Context, cid string) (bool, error) {
	pins, err := c.sh.Pins()
	if err != nil {
		return false, fmt.Errorf("failed to list pins: %w", err)
	}

	_, exists := pins[cid]
	return exists, nil
}

// Version returns the IPFS version info for health checks
func (c *RealIPFSClient) Version(ctx context.Context) (string, error) {
	ver, _, err := c.sh.Version()
	if err != nil {
		return "", fmt.Errorf("failed to get ipfs version: %w", err)
	}
	return ver, nil
}

// IsHealthy checks if the IPFS connection is healthy
func (c *RealIPFSClient) IsHealthy(ctx context.Context) error {
	_, err := c.Version(ctx)
	return err
}

// Ensure RealIPFSClient implements IPFSClient
var _ IPFSClient = (*RealIPFSClient)(nil)

// StubIPFSClient implements IPFSClient using in-memory storage.
// This is used for testing without a real IPFS node.
//
// WARNING: This client generates FAKE CIDs that are NOT valid IPFS CIDs.
// The stub CIDs use a "Qm" prefix followed by hex-encoded hash bytes,
// which differs from real CIDv0 format (base58-encoded multihash).
//
// NEVER use StubIPFSClient in production:
//   - Data is stored in-memory and lost on restart
//   - Generated CIDs are not valid IPFS content identifiers
//   - CID validation will reject stub CIDs in production mode
//
// For production, use RealIPFSClient with a real IPFS node.
type StubIPFSClient struct {
	mu         sync.RWMutex
	storage    map[string][]byte
	pins       map[string]bool
	cidCounter uint64
}

// NewStubIPFSClient creates a new stub IPFS client for testing.
// WARNING: This is for testing only. See StubIPFSClient documentation.
func NewStubIPFSClient() *StubIPFSClient {
	return &StubIPFSClient{
		storage: make(map[string][]byte),
		pins:    make(map[string]bool),
	}
}

// generateCID generates a deterministic FAKE CID for the stub implementation.
// This is NOT a valid IPFS CID - it uses hex encoding instead of base58.
// The format is: "Qm" + 32 hex characters (16 bytes of SHA256 hash).
// Real CIDv0 uses base58-encoded multihash which includes non-hex characters.
func (c *StubIPFSClient) generateCID(data []byte) string {
	c.cidCounter++
	hash := sha256.Sum256(data)
	// Generate a FAKE CID: Qm + hex encoding (NOT real base58)
	// This is intentionally distinguishable from real CIDs
	return "Qm" + hex.EncodeToString(hash[:16])
}

// Add adds data to in-memory storage and returns a FAKE CID.
// WARNING: The CID returned is NOT a valid IPFS CID.
func (c *StubIPFSClient) Add(_ context.Context, data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cid := c.generateCID(data)
	c.storage[cid] = make([]byte, len(data))
	copy(c.storage[cid], data)

	return cid, nil
}

// Cat retrieves data by CID from in-memory storage
func (c *StubIPFSClient) Cat(_ context.Context, cid string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, exists := c.storage[cid]
	if !exists {
		return nil, ErrArtifactNotFound.Wrapf("cid not found: %s", cid)
	}

	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// CatStream retrieves data as a stream by CID from in-memory storage
func (c *StubIPFSClient) CatStream(ctx context.Context, cid string) (io.ReadCloser, error) {
	data, err := c.Cat(ctx, cid)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

// Pin marks a CID as pinned in-memory
func (c *StubIPFSClient) Pin(_ context.Context, cid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.storage[cid]; !exists {
		return ErrArtifactNotFound.Wrapf("cid not found: %s", cid)
	}

	c.pins[cid] = true
	return nil
}

// Unpin removes a pin from in-memory
func (c *StubIPFSClient) Unpin(_ context.Context, cid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.pins, cid)
	return nil
}

// IsPinned checks if a CID is pinned in-memory
func (c *StubIPFSClient) IsPinned(_ context.Context, cid string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.pins[cid]
	return exists, nil
}

// Version returns a stub version string
func (c *StubIPFSClient) Version(_ context.Context) (string, error) {
	return "stub-0.0.0", nil
}

// IsHealthy always returns nil for stub client
func (c *StubIPFSClient) IsHealthy(_ context.Context) error {
	return nil
}

// Delete removes data from in-memory storage (stub-only method)
func (c *StubIPFSClient) Delete(cid string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.storage, cid)
	delete(c.pins, cid)
}

// Ensure StubIPFSClient implements IPFSClient
var _ IPFSClient = (*StubIPFSClient)(nil)
