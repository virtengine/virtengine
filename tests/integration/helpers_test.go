//go:build e2e.integration

package integration

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/virtengine/virtengine/app"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// =============================================================================
// Project Root Detection
// =============================================================================

// findProjectRoot walks up the directory tree to find the project root.
// It looks for markers like go.mod, Makefile, or .git directory.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Walk up the directory tree looking for project root markers
	dir := cwd
	for {
		// Check for go.mod (primary marker)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Verify it's the VirtEngine project by checking for Makefile
			if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
				return dir
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// Fallback: check if VIRTENGINE_ROOT is set
	if root := os.Getenv("VIRTENGINE_ROOT"); root != "" {
		return root
	}

	// Last resort: return cwd and let the tests fail with a clear message
	t.Logf("Warning: Could not find project root, using: %s", cwd)
	return cwd
}

// =============================================================================
// Environment Helpers
// =============================================================================

// getEnvWithDefault returns the environment variable value or a default.
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// =============================================================================
// VEID Integration Helpers
// =============================================================================

type veidTestClient struct {
	ClientID   string
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

func newVEIDTestClient() veidTestClient {
	seed := sha256.Sum256([]byte("virtengine-veid-approved-client"))
	privKey := ed25519.NewKeyFromSeed(seed[:])
	pubKey := privKey.Public().(ed25519.PublicKey)

	return veidTestClient{
		ClientID:   "test-capture-client",
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
}

func genesisWithVEIDApprovedClient(t testing.TB, cdc codec.Codec, client veidTestClient) app.GenesisState {
	t.Helper()

	genesis := app.GenesisStateWithValSet(cdc)

	var veidGenesis veidtypes.GenesisState
	require.NoError(t, json.Unmarshal(genesis[veidtypes.ModuleName], &veidGenesis))

	veidGenesis.ApprovedClients = append(veidGenesis.ApprovedClients, veidtypes.ApprovedClient{
		ClientID:     client.ClientID,
		Name:         "Integration Test Capture Client",
		PublicKey:    client.PublicKey,
		Algorithm:    "ed25519",
		Active:       true,
		RegisteredAt: time.Unix(0, 0).Unix(),
		Metadata: map[string]string{
			"purpose": "integration-tests",
		},
	})

	veidGenesis.Params.RequireClientSignature = true
	veidGenesis.Params.RequireUserSignature = true

	updated, err := json.Marshal(&veidGenesis)
	require.NoError(t, err)
	genesis[veidtypes.ModuleName] = updated

	return genesis
}

// =============================================================================
// Test Skip Helpers
// =============================================================================

// skipIfNoBinary skips the test if the binary is not available.
func skipIfNoBinary(t *testing.T, binaryPath string) {
	t.Helper()

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("Binary not found at %s - run 'make' first", binaryPath)
	}
}

// skipIfShort skips the test in short mode.
func skipIfShort(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// =============================================================================
// File Verification Helpers
// =============================================================================

// FileInfo contains information about a file for verification.
type FileInfo struct {
	Path      string
	MinSize   int64 // Minimum expected size in bytes
	MustExist bool
}

// verifyFilesExist checks that all specified files exist and meet size requirements.
func verifyFilesExist(t *testing.T, files []FileInfo) {
	t.Helper()

	for _, f := range files {
		info, err := os.Stat(f.Path)
		if f.MustExist {
			if err != nil {
				t.Errorf("Required file missing: %s", f.Path)
				continue
			}
		} else if os.IsNotExist(err) {
			continue
		}

		if err == nil && f.MinSize > 0 && info.Size() < f.MinSize {
			t.Errorf("File too small: %s (got %d bytes, want >= %d)", f.Path, info.Size(), f.MinSize)
		}
	}
}

// =============================================================================
// Cleanup Helpers
// =============================================================================

// TempDirWithCleanup creates a temporary directory and registers cleanup.
func TempDirWithCleanup(t *testing.T, prefix string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}
