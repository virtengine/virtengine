package slurm_adapter

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// VE-2020: SSH SLURM Client tests

// withTempKnownHosts writes a throwaway known_hosts file containing a real
// ed25519 host key, generated in-process so the known_hosts parser is genuinely
// exercised and the fixture can never go stale or be syntactically invalid.
//
// The client fails closed on a missing known_hosts file, and the default path
// is the operator's real ~/.ssh/known_hosts. Constructing a client in a test
// must not depend on that file existing on whichever machine or CI runner is
// executing the suite, so every construction test points KnownHostsPath at this
// temp file instead. This keeps the real verification path under test.
func withTempKnownHosts(t *testing.T) string {
	t.Helper()
	_, publicKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPublicKey, err := ssh.NewPublicKey(publicKey.Public())
	require.NoError(t, err)
	line := "localhost " + strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPublicKey))) + "\n"

	path := filepath.Join(t.TempDir(), "known_hosts")
	require.NoError(t, os.WriteFile(path, []byte(line), 0o600))
	return path
}

// sshTestConfig returns an SSHConfig that constructs successfully without
// depending on the ambient home directory.
func sshTestConfig(t *testing.T) SSHConfig {
	t.Helper()
	return SSHConfig{
		Host:           "localhost",
		Port:           22,
		User:           "testuser",
		Password:       "testpass",
		KnownHostsPath: withTempKnownHosts(t),
	}
}

// TestNewSSHSLURMClient_MissingKnownHostsFailsClosed pins the fail-closed
// default: with no known_hosts available the constructor must refuse rather
// than silently skip host key verification.
func TestNewSSHSLURMClient_MissingKnownHostsFailsClosed(t *testing.T) {
	config := SSHConfig{
		Host:           "localhost",
		Port:           22,
		User:           "testuser",
		Password:       "testpass",
		KnownHostsPath: filepath.Join(t.TempDir(), "absent_known_hosts"),
	}

	_, err := NewSSHSLURMClient(config, "cluster1", "default")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHostKeyVerification)
}

func TestDefaultSSHConfig(t *testing.T) {
	config := DefaultSSHConfig()

	assert.Equal(t, 22, config.Port)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 30*time.Second, config.KeepAliveInterval)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "known_hosts", config.HostKeyCallback) // Secure default: verify host keys
	assert.Equal(t, 5, config.PoolSize)
	assert.Equal(t, 5*time.Minute, config.PoolIdleTimeout)
}

func TestNewSSHSLURMClient_NoAuth(t *testing.T) {
	config := SSHConfig{
		Host: "localhost",
		Port: 22,
		User: "testuser",
		// No auth method
	}

	_, err := NewSSHSLURMClient(config, "cluster1", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no authentication method")
}

func TestNewSSHSLURMClient_WithPassword(t *testing.T) {
	config := sshTestConfig(t)
	config.Timeout = 5 * time.Second

	client, err := NewSSHSLURMClient(config, "cluster1", "default")
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Equal(t, "cluster1", client.clusterName)
	assert.Equal(t, "default", client.defaultPartition)
	assert.False(t, client.IsConnected())
}

func TestNewSSHSLURMClient_WithInvalidKey(t *testing.T) {
	config := SSHConfig{
		Host:       "localhost",
		Port:       22,
		User:       "testuser",
		PrivateKey: "not-a-valid-key",
	}

	_, err := NewSSHSLURMClient(config, "cluster1", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestNewSSHSLURMClient_WithPoolSize(t *testing.T) {
	config := sshTestConfig(t)
	config.PoolSize = 10
	config.PoolIdleTimeout = 10 * time.Minute

	client, err := NewSSHSLURMClient(config, "cluster1", "default")
	require.NoError(t, err)
	require.NotNil(t, client)

	// Pool should be initialized but empty until Connect
	assert.NotNil(t, client.pool)
	assert.Len(t, client.pool, 0)
}

func TestSSHSLURMClient_GenerateBatchScript_ViaFromJobSpec(t *testing.T) {
	config := sshTestConfig(t)

	client, err := NewSSHSLURMClient(config, "cluster1", "compute")
	require.NoError(t, err)
	_ = client // Client not used directly for script generation

	spec := &SLURMJobSpec{
		JobName:          "test-job",
		Partition:        "gpu",
		Nodes:            2,
		CPUsPerNode:      16,
		MemoryMB:         32768,
		TimeLimit:        3600,
		GPUs:             4,
		GPUType:          "a100",
		WorkingDirectory: "/home/user/workdir",
		OutputDirectory:  "/home/user/output",
		Exclusive:        true,
		Constraints:      []string{"nvme", "ib"},
		Environment: map[string]string{
			"CUDA_VISIBLE_DEVICES": "0,1,2,3",
			"OMP_NUM_THREADS":      "16",
		},
		Command:   "python",
		Arguments: []string{"train.py", "--epochs=100"},
	}

	// Use BatchScriptBuilder
	script := FromJobSpec(spec).Build()

	// Verify SBATCH directives
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "#SBATCH --job-name=test-job")
	assert.Contains(t, script, "#SBATCH --partition=gpu")
	assert.Contains(t, script, "#SBATCH --nodes=2")
	assert.Contains(t, script, "#SBATCH --cpus-per-task=16")
	assert.Contains(t, script, "#SBATCH --mem=32768M")
	assert.Contains(t, script, "#SBATCH --gres=gpu:a100:4")
	assert.Contains(t, script, "#SBATCH --chdir=/home/user/workdir")
	assert.Contains(t, script, "#SBATCH --output=/home/user/output/%j.out")
	assert.Contains(t, script, "#SBATCH --error=/home/user/output/%j.err")
	assert.Contains(t, script, "#SBATCH --exclusive")
	assert.Contains(t, script, "#SBATCH --constraint=nvme")
	assert.Contains(t, script, "#SBATCH --constraint=ib")

	// Verify environment
	assert.Contains(t, script, "export CUDA_VISIBLE_DEVICES=")
	assert.Contains(t, script, "export OMP_NUM_THREADS=")

	// Verify command
	assert.Contains(t, script, "python")
	assert.Contains(t, script, "train.py")
	assert.Contains(t, script, "--epochs=100")
}

func TestSSHSLURMClient_GenerateBatchScript_WithContainer(t *testing.T) {
	spec := &SLURMJobSpec{
		JobName:        "container-job",
		Nodes:          1,
		CPUsPerNode:    8,
		MemoryMB:       16384,
		TimeLimit:      1800,
		ContainerImage: "docker://pytorch/pytorch:latest",
		Command:        "python",
		Arguments:      []string{"inference.py"},
	}

	script := FromJobSpec(spec).Build()

	assert.Contains(t, script, "singularity exec docker://pytorch/pytorch:latest python")
}

func TestSSHSLURMClient_GenerateBatchScript_DefaultPartition(t *testing.T) {
	spec := &SLURMJobSpec{
		JobName:     "test-job",
		Partition:   "", // Empty - should use default
		Nodes:       1,
		CPUsPerNode: 4,
		MemoryMB:    8192,
		TimeLimit:   600,
		Command:     "echo",
		Arguments:   []string{"hello"},
	}

	// When partition is empty, FromJobSpec won't set it
	// The SSH client should use defaultPartition
	builder := FromJobSpec(spec)
	builder.SetPartition("default-partition")
	script := builder.Build()

	assert.Contains(t, script, "#SBATCH --partition=default-partition")
}

func TestMapSLURMState(t *testing.T) {
	tests := []struct {
		input    string
		expected SLURMJobState
	}{
		{"PENDING", SLURMJobStatePending},
		{"PD", SLURMJobStatePending},
		{"RUNNING", SLURMJobStateRunning},
		{"R", SLURMJobStateRunning},
		{"COMPLETED", SLURMJobStateCompleted},
		{"CD", SLURMJobStateCompleted},
		{"FAILED", SLURMJobStateFailed},
		{"F", SLURMJobStateFailed},
		{"CANCELLED", SLURMJobStateCancelled},
		{"CA", SLURMJobStateCancelled},
		{"TIMEOUT", SLURMJobStateTimeout},
		{"TO", SLURMJobStateTimeout},
		{"SUSPENDED", SLURMJobStateSuspended},
		{"S", SLURMJobStateSuspended},
		{"  RUNNING  ", SLURMJobStateRunning}, // With whitespace
		{"unknown", SLURMJobState("UNKNOWN")},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := mapSLURMState(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"00:00:30", 30},
		{"00:05:00", 300},
		{"01:00:00", 3600},
		{"24:00:00", 86400},
		{"1-00:00:00", 86400},
		{"2-12:30:45", 2*86400 + 12*3600 + 30*60 + 45},
		{"05:30", 5*60 + 30},
		{"30", 30},
		{"00:05:30.123", 330}, // With milliseconds
		{"UNLIMITED", 0},
		{"INVALID", 0},
		{"", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseDuration(tc.input)
			assert.Equal(t, tc.expected, result, "input: %s", tc.input)
		})
	}
}

func TestParseMemory(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1024", 1024},
		{"1024K", 1024 * 1024},
		{"1024k", 1024 * 1024},
		{"512M", 512 * 1024 * 1024},
		{"2G", 2 * 1024 * 1024 * 1024},
		{"1T", 1024 * 1024 * 1024 * 1024},
		{"", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseMemory(tc.input)
			assert.Equal(t, tc.expected, result, "input: %s", tc.input)
		})
	}
}

func TestParseGRES(t *testing.T) {
	tests := []struct {
		input    string
		gpuCount int32
		gpuType  string
	}{
		{"gpu:4", 4, ""},
		{"gpu:a100:8", 8, "a100"},
		{"gpu:v100:2", 2, "v100"},
		{"", 0, ""},
		{"cpu:4", 0, ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			count, gpuType := parseGRES(tc.input)
			assert.Equal(t, tc.gpuCount, count, "count mismatch for: %s", tc.input)
			assert.Equal(t, tc.gpuType, gpuType, "type mismatch for: %s", tc.input)
		})
	}
}

func TestParseNodeList(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"node001", []string{"node001"}},
		{"node001,node002,node003", []string{"node001", "node002", "node003"}},
		{"node[001-004]", []string{"node[001-004]"}}, // TODO: Expand ranges
		{"", nil},
		{"(null)", nil},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseNodeList(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNumeric(t *testing.T) {
	assert.True(t, isNumeric("123"))
	assert.True(t, isNumeric("0"))
	assert.True(t, isNumeric("999999"))
	assert.False(t, isNumeric(""))
	assert.False(t, isNumeric("abc"))
	assert.False(t, isNumeric("12.3"))
	assert.False(t, isNumeric("123abc"))
}

func TestSSHSLURMClient_IsConnected_NotConnected(t *testing.T) {
	config := sshTestConfig(t)

	client, err := NewSSHSLURMClient(config, "cluster1", "default")
	require.NoError(t, err)

	assert.False(t, client.IsConnected())
}
