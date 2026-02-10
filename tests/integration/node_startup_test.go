//go:build e2e.integration

// Package integration contains integration tests for VirtEngine.
// These tests verify node startup and basic operations.
package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// NodeStartupTestSuite tests node startup and basic CLI operations.
// This suite verifies:
//   - Binary exists and is executable
//   - Version command works
//   - Init command creates valid genesis
//
// Task Reference: VE-1009
type NodeStartupTestSuite struct {
	suite.Suite

	// binaryPath is the path to the virtengine binary
	binaryPath string

	// tempDir is a temporary directory for test data
	tempDir string
}

// TestNodeStartup runs the node startup test suite.
func TestNodeStartup(t *testing.T) {
	suite.Run(t, new(NodeStartupTestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *NodeStartupTestSuite) SetupSuite() {
	// Determine binary path based on OS
	binaryName := "virtengine"
	if runtime.GOOS == "windows" {
		binaryName = "virtengine.exe"
	}

	// Look for binary in .cache/bin relative to project root
	projectRoot := findProjectRoot(s.T())
	s.binaryPath = filepath.Join(projectRoot, ".cache", "bin", binaryName)

	// Build the binary if it doesn't exist
	if _, err := os.Stat(s.binaryPath); os.IsNotExist(err) {
		s.T().Logf("Binary not found at %s, building...", s.binaryPath)

		binDir := filepath.Dir(s.binaryPath)
		require.NoError(s.T(), os.MkdirAll(binDir, 0o755), "Failed to create bin directory")

		cmd := exec.Command("go", "build", "-o", s.binaryPath, "./cmd/virtengine")
		cmd.Dir = projectRoot
		out, buildErr := cmd.CombinedOutput()
		require.NoError(s.T(), buildErr, "Failed to build binary: %s", string(out))
	}

	s.T().Logf("Using binary: %s", s.binaryPath)
}

// SetupTest runs before each test.
func (s *NodeStartupTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "virtengine-test-*")
	require.NoError(s.T(), err, "Failed to create temp directory")
}

// TearDownTest runs after each test.
func (s *NodeStartupTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// =============================================================================
// Binary Verification Tests
// =============================================================================

// TestBinaryExists verifies the virtengine binary exists and is executable.
func (s *NodeStartupTestSuite) TestBinaryExists() {
	s.T().Log("=== Test: Binary Exists ===")

	// Check binary exists
	info, err := os.Stat(s.binaryPath)
	require.NoError(s.T(), err, "Binary should exist at %s", s.binaryPath)
	require.False(s.T(), info.IsDir(), "Binary path should not be a directory")

	s.T().Logf("Binary found: %s (size: %d bytes)", s.binaryPath, info.Size())
}

// TestVersionCommand verifies the version command works correctly.
func (s *NodeStartupTestSuite) TestVersionCommand() {
	s.T().Log("=== Test: Version Command ===")

	// Run version command with --long flag (plain version outputs to stderr or is empty)
	cmd := exec.Command(s.binaryPath, "version", "--long", "--output", "json")
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Version command should succeed: %s", string(output))

	outputStr := strings.TrimSpace(string(output))
	require.NotEmpty(s.T(), outputStr, "Version output should not be empty")

	// Verify it's valid JSON
	var versionInfo map[string]interface{}
	err = json.Unmarshal([]byte(outputStr), &versionInfo)
	require.NoError(s.T(), err, "Version output should be valid JSON")

	s.T().Logf("Version info: cosmos_sdk_version=%v", versionInfo["cosmos_sdk_version"])
}

// TestVersionLongCommand verifies the version --long command works.
func (s *NodeStartupTestSuite) TestVersionLongCommand() {
	s.T().Log("=== Test: Version Long Command ===")

	cmd := exec.Command(s.binaryPath, "version", "--long")
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Version --long command should succeed: %s", string(output))

	outputStr := string(output)
	s.T().Logf("Version long output:\n%s", outputStr)

	// Verify it contains expected fields
	require.Contains(s.T(), strings.ToLower(outputStr), "name", "Should contain 'name' field")
}

// =============================================================================
// Init Command Tests
// =============================================================================

// TestInitCommand verifies the genesis init command creates a valid chain configuration.
// The init command in VirtEngine is under the genesis subcommand.
func (s *NodeStartupTestSuite) TestInitCommand() {
	s.T().Log("=== Test: Genesis Init Command ===")

	moniker := "test-node"
	chainID := "virtengine-test-1"
	homeDir := filepath.Join(s.tempDir, ".virtengine")

	// Run genesis init command (this is the Cosmos SDK v0.5x pattern)
	cmd := exec.Command(s.binaryPath, "genesis", "init", moniker,
		"--chain-id", chainID,
		"--home", homeDir,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Genesis init command should succeed: %s", string(output))

	s.T().Logf("Init output:\n%s", string(output))

	// Verify config directory was created
	configDir := filepath.Join(homeDir, "config")
	require.DirExists(s.T(), configDir, "Config directory should exist")

	// Verify data directory was created
	dataDir := filepath.Join(homeDir, "data")
	require.DirExists(s.T(), dataDir, "Data directory should exist")
}

// TestGenesisFileCreated verifies genesis init creates a valid genesis.json file.
func (s *NodeStartupTestSuite) TestGenesisFileCreated() {
	s.T().Log("=== Test: Genesis File Created ===")

	moniker := "genesis-test-node"
	chainID := "virtengine-genesis-test-1"
	homeDir := filepath.Join(s.tempDir, ".virtengine-genesis")

	// Run genesis init command
	cmd := exec.Command(s.binaryPath, "genesis", "init", moniker,
		"--chain-id", chainID,
		"--home", homeDir,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Genesis init command should succeed: %s", string(output))

	// Check genesis file exists
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")
	require.FileExists(s.T(), genesisPath, "Genesis file should exist")

	// Read and parse genesis file
	genesisData, err := os.ReadFile(genesisPath)
	require.NoError(s.T(), err, "Should read genesis file")
	require.NotEmpty(s.T(), genesisData, "Genesis file should not be empty")

	// Verify it's valid JSON
	var genesis map[string]interface{}
	err = json.Unmarshal(genesisData, &genesis)
	require.NoError(s.T(), err, "Genesis file should be valid JSON")

	// Verify chain_id matches
	require.Equal(s.T(), chainID, genesis["chain_id"], "Chain ID should match")

	s.T().Logf("Genesis file created successfully with chain_id: %s", chainID)
}

// TestConfigFilesCreated verifies genesis init creates all required config files.
func (s *NodeStartupTestSuite) TestConfigFilesCreated() {
	s.T().Log("=== Test: Config Files Created ===")

	moniker := "config-test-node"
	chainID := "virtengine-config-test-1"
	homeDir := filepath.Join(s.tempDir, ".virtengine-config")

	// Run genesis init command
	cmd := exec.Command(s.binaryPath, "genesis", "init", moniker,
		"--chain-id", chainID,
		"--home", homeDir,
	)
	_, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Genesis init command should succeed")

	configDir := filepath.Join(homeDir, "config")

	// Check required config files (Cosmos SDK v0.53+ pattern)
	// Note: client.toml may not be created by genesis init in all SDK versions
	requiredFiles := []string{
		"genesis.json",
		"config.toml",
		"app.toml",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(configDir, file)
		require.FileExists(s.T(), filePath, "Config file should exist: %s", file)

		info, err := os.Stat(filePath)
		require.NoError(s.T(), err)
		require.Greater(s.T(), info.Size(), int64(0), "Config file should not be empty: %s", file)

		s.T().Logf("Config file created: %s (size: %d bytes)", file, info.Size())
	}
}

// =============================================================================
// Help Command Tests
// =============================================================================

// TestHelpCommand verifies the help command works.
func (s *NodeStartupTestSuite) TestHelpCommand() {
	s.T().Log("=== Test: Help Command ===")

	cmd := exec.Command(s.binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Help command should succeed: %s", string(output))

	outputStr := string(output)
	require.NotEmpty(s.T(), outputStr, "Help output should not be empty")

	// Verify it contains expected subcommands
	require.Contains(s.T(), outputStr, "genesis", "Help should mention 'genesis' command")
	require.Contains(s.T(), outputStr, "start", "Help should mention 'start' command")
	require.Contains(s.T(), outputStr, "version", "Help should mention 'version' command")

	s.T().Log("Help command shows expected subcommands")
}

// TestQueryHelpCommand verifies query subcommand help works.
func (s *NodeStartupTestSuite) TestQueryHelpCommand() {
	s.T().Log("=== Test: Query Help Command ===")

	cmd := exec.Command(s.binaryPath, "query", "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Query help command should succeed: %s", string(output))

	outputStr := string(output)
	require.NotEmpty(s.T(), outputStr, "Query help output should not be empty")

	s.T().Log("Query help command works")
}

// TestTxHelpCommand verifies tx subcommand help works.
func (s *NodeStartupTestSuite) TestTxHelpCommand() {
	s.T().Log("=== Test: Tx Help Command ===")

	cmd := exec.Command(s.binaryPath, "tx", "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Tx help command should succeed: %s", string(output))

	outputStr := string(output)
	require.NotEmpty(s.T(), outputStr, "Tx help output should not be empty")

	s.T().Log("Tx help command works")
}

// TestKeysHelpCommand verifies keys subcommand help works.
func (s *NodeStartupTestSuite) TestKeysHelpCommand() {
	s.T().Log("=== Test: Keys Help Command ===")

	cmd := exec.Command(s.binaryPath, "keys", "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Keys help command should succeed: %s", string(output))

	outputStr := string(output)
	require.NotEmpty(s.T(), outputStr, "Keys help output should not be empty")
	require.Contains(s.T(), outputStr, "add", "Keys help should mention 'add' command")
	require.Contains(s.T(), outputStr, "list", "Keys help should mention 'list' command")

	s.T().Log("Keys help command works")
}
