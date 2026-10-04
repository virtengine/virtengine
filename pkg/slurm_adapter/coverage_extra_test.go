// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package slurm_adapter

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestSSHClient builds an SSH client that is not connected. Every
// "not connected" guard can be exercised without a network.
func newTestSSHClient(t *testing.T, tweak func(*SSHConfig)) *SSHSLURMClient {
	t.Helper()
	// HostKeyCallback defaults to "ignore" so this helper never depends on the
	// ambient ~/.ssh/known_hosts. The constructor fails closed on a missing
	// known_hosts, which made every test using this helper fail on CI runners
	// (no ~/.ssh) while passing on developer machines that have one.
	// Host-key behaviour itself is asserted explicitly in
	// TestNewSSHSLURMClient_HostKeyCallbackModes, which sets its own mode.
	cfg := SSHConfig{
		Host:            "127.0.0.1",
		Port:            22,
		User:            "testuser",
		Password:        "testpass",
		Timeout:         time.Second,
		HostKeyCallback: "ignore",
	}
	if tweak != nil {
		tweak(&cfg)
	}
	c, err := NewSSHSLURMClient(cfg, "cluster1", "default")
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

// testSigner is an in-package JobSigner. The signer in adapter_test.go lives
// in the external `slurm_adapter_test` package and is not importable here.
type testSigner struct{ addr string }

func (s testSigner) Sign(data []byte) ([]byte, error) {
	sig := make([]byte, 0, len(data)+16)
	sig = append(sig, []byte("sig:")...)
	return append(sig, data...), nil
}
func (s testSigner) Verify(data, signature []byte) bool { return len(signature) > len(data) }
func (s testSigner) GetProviderAddress() string         { return s.addr }

const testProviderAddr = "ve1provider123abc456def"

// --- types.go / batch_script.go ---

func TestMapToVirtEngineState(t *testing.T) {
	cases := []struct {
		in   SLURMJobState
		want string
	}{
		{SLURMJobStatePending, "queued"},
		{SLURMJobStateRunning, "running"},
		{SLURMJobStateCompleted, "completed"},
		{SLURMJobStateFailed, "failed"},
		{SLURMJobStateCancelled, "cancelled"},
		{SLURMJobStateTimeout, "timeout"},
		{SLURMJobStateSuspended, "paused"},
		{SLURMJobState("WEIRD"), "pending"},
		{SLURMJobState(""), "pending"},
	}
	for _, tc := range cases {
		if got := MapToVirtEngineState(tc.in); got != tc.want {
			t.Errorf("MapToVirtEngineState(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsTerminalState(t *testing.T) {
	terminal := []SLURMJobState{
		SLURMJobStateCompleted, SLURMJobStateFailed,
		SLURMJobStateCancelled, SLURMJobStateTimeout,
	}
	for _, s := range terminal {
		if !isTerminalState(s) {
			t.Errorf("isTerminalState(%q) = false, want true", s)
		}
	}
	for _, s := range []SLURMJobState{SLURMJobStatePending, SLURMJobStateRunning, SLURMJobStateSuspended, SLURMJobState("X")} {
		if isTerminalState(s) {
			t.Errorf("isTerminalState(%q) = true, want false", s)
		}
	}
}

func TestBatchScriptBuilder_SetGPUsPerTask(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("gpu-task").
		SetGPUsPerTask(2).
		SetGPUsPerNode(4).
		Build()

	assert.Contains(t, script, "#SBATCH --gpus-per-task=2")
	assert.Contains(t, script, "#SBATCH --gpus-per-node=4")

	// A builder that never sets it emits neither directive.
	plain := NewBatchScriptBuilder().SetJobName("plain").Build()
	assert.NotContains(t, plain, "--gpus-per-task")
	assert.NotContains(t, plain, "--gpus-per-node")
}

// --- ssh_client.go: constructor paths ---

func TestNewSSHSLURMClient_HostKeyCallbackModes(t *testing.T) {
	keyDir := t.TempDir()
	knownHosts := filepath.Join(keyDir, "known_hosts")
	require.NoError(t, os.WriteFile(knownHosts, []byte(""), 0600))

	t.Run("ignore", func(t *testing.T) {
		c := newTestSSHClient(t, func(cfg *SSHConfig) { cfg.HostKeyCallback = "ignore" })
		assert.NotNil(t, c.sshConfig.HostKeyCallback)
	})

	t.Run("known_hosts present", func(t *testing.T) {
		c := newTestSSHClient(t, func(cfg *SSHConfig) {
			cfg.HostKeyCallback = "known_hosts"
			cfg.KnownHostsPath = knownHosts
		})
		assert.NotNil(t, c.sshConfig.HostKeyCallback)
	})

	t.Run("known_hosts missing is an error", func(t *testing.T) {
		_, err := NewSSHSLURMClient(SSHConfig{
			Host: "h", Port: 22, User: "u", Password: "p",
			HostKeyCallback: "known_hosts",
			KnownHostsPath:  filepath.Join(keyDir, "absent"),
		}, "c", "d")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrHostKeyVerification)
	})

	t.Run("known_hosts unparseable", func(t *testing.T) {
		bad := filepath.Join(keyDir, "bad_known_hosts")
		require.NoError(t, os.WriteFile(bad, []byte("!!! not a known_hosts line\n"), 0600))
		_, err := NewSSHSLURMClient(SSHConfig{
			Host: "h", Port: 22, User: "u", Password: "p",
			HostKeyCallback: "known_hosts",
			KnownHostsPath:  bad,
		}, "c", "d")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "known_hosts")
	})

	t.Run("default fails closed when no known_hosts", func(t *testing.T) {
		// HostKeyCallback == "" with a nonexistent known_hosts must ERROR,
		// not silently fall back to InsecureIgnoreHostKey. Skipping host key
		// verification now requires an explicit HostKeyCallback="ignore".
		_, err := NewSSHSLURMClient(SSHConfig{
			Host: "h", Port: 22, User: "u", Password: "p",
			HostKeyCallback: "",
			KnownHostsPath:  filepath.Join(keyDir, "absent"),
		}, "c", "d")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrHostKeyVerification)
		assert.Contains(t, err.Error(), "ignore")
	})

	t.Run("default uses known_hosts when present", func(t *testing.T) {
		c := newTestSSHClient(t, func(cfg *SSHConfig) {
			cfg.HostKeyCallback = ""
			cfg.KnownHostsPath = knownHosts
		})
		assert.NotNil(t, c.sshConfig.HostKeyCallback)
	})
}

func TestNewSSHSLURMClient_PrivateKeyPath(t *testing.T) {
	dir := t.TempDir()

	missing := filepath.Join(dir, "absent_key")
	_, err := NewSSHSLURMClient(SSHConfig{
		Host: "h", Port: 22, User: "u", PrivateKeyPath: missing,
	}, "c", "d")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read private key file")

	// A file that exists but is not a key -> parse failure, not a read failure.
	notKey := filepath.Join(dir, "not_a_key")
	require.NoError(t, os.WriteFile(notKey, []byte("hello"), 0600))
	_, err = NewSSHSLURMClient(SSHConfig{
		Host: "h", Port: 22, User: "u", PrivateKeyPath: notKey,
	}, "c", "d")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestNewSSHSLURMClient_DefaultPoolSize(t *testing.T) {
	// PoolSize <= 0 must fall back to 5 without panicking. The fallback is
	// applied to the internal pool capacity, not written back to config.
	c := newTestSSHClient(t, func(cfg *SSHConfig) { cfg.PoolSize = 0 })
	assert.Empty(t, c.pool)
	assert.Equal(t, 0, c.config.PoolSize)
	assert.Equal(t, 5, cap(c.pool), "pool falls back to the default capacity of 5")
}

func TestParseSSHPrivateKey(t *testing.T) {
	// Garbage without a passphrase returns the original parse error.
	_, err := parseSSHPrivateKey([]byte("garbage"), "")
	require.Error(t, err)

	// Garbage with a passphrase attempts the passphrase parser and fails there.
	_, err = parseSSHPrivateKey([]byte("garbage"), "hunter2")
	require.Error(t, err)
}

// --- ssh_client.go: disconnect / pool guards ---

func TestSSHSLURMClient_DisconnectIdempotent(t *testing.T) {
	c := newTestSSHClient(t, nil)
	// Not connected: a no-op that must return nil.
	require.NoError(t, c.Disconnect())
	require.False(t, c.IsConnected())
}

func TestSSHSLURMClient_NotConnectedGuards(t *testing.T) {
	c := newTestSSHClient(t, nil)
	ctx := context.Background()

	_, err := c.SubmitJob(ctx, &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	assert.ErrorIs(t, c.CancelJob(ctx, "1"), ErrSLURMNotConnected)

	_, err = c.GetJobStatus(ctx, "1")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	_, err = c.GetJobAccounting(ctx, "1")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	_, err = c.ListPartitions(ctx)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	_, err = c.ListNodes(ctx)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	_, err = c.SubmitJobFromScript(ctx, "#!/bin/bash\n")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	_, err = c.SubmitJobFromFile(ctx, "/tmp/job.sh")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	assert.ErrorIs(t, c.SCPUpload(ctx, "local", "/remote"), ErrSLURMNotConnected)
	assert.ErrorIs(t, c.SCPUploadBytes(ctx, []byte("x"), "/remote/f", 0644), ErrSLURMNotConnected)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
	assert.ErrorIs(t, c.SCPDownload(ctx, "/remote/f", "local"), ErrSLURMNotConnected)
	_, err = c.SCPDownloadBytes(ctx, "/remote/f")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
	assert.ErrorIs(t, c.WriteRemoteFile(ctx, "/remote/f", []byte("x"), 0644), ErrSLURMNotConnected)

	_, err = c.runCommand(ctx, "squeue")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	_, err = c.runCommandWithTimeout(ctx, "squeue", time.Millisecond)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
}

func TestSSHSLURMClient_CLIPathValidation(t *testing.T) {
	dir := t.TempDir()
	traversal := filepath.Join(dir, "sub", "..", "..", "etc", "passwd")

	// Both SCP entry points check the connection BEFORE validating the path,
	// so a traversal path on a disconnected client reports the connection
	// error. Pin that ordering explicitly.
	c := newTestSSHClient(t, nil)
	err := c.SCPUpload(context.Background(), traversal, "/remote/f")
	assert.ErrorIs(t, err, ErrSLURMNotConnected,
		"the not-connected guard runs before CLI path validation")

	err = c.SCPDownload(context.Background(), "/remote/f", traversal)
	assert.ErrorIs(t, err, ErrSLURMNotConnected,
		"the not-connected guard runs before CLI path validation on download too")
}

// --- ssh_client.go: squeue / sacct parsing ---

func TestParseSqueueOutput(t *testing.T) {
	c := newTestSSHClient(t, nil)

	// Too few fields.
	_, err := c.parseSqueueOutput("1", "1|job|RUNNING")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected squeue output format")

	// Well-formed with a start time and a simple node list.
	job, err := c.parseSqueueOutput("1", "1|myjob|RUNNING|00:05:00|2026-01-02T03:04:05|N/A|node001,node002\n")
	require.NoError(t, err)
	assert.Equal(t, "1", job.SLURMJobID)
	assert.Equal(t, SLURMJobStateRunning, job.State)
	assert.Equal(t, []string{"node001", "node002"}, job.NodeList)
	require.NotNil(t, job.StartTime)
	assert.Equal(t, 2026, job.StartTime.Year())

	// N/A start time leaves StartTime nil.
	job, err = c.parseSqueueOutput("1", "1|myjob|PENDING|00:00:00|N/A|N/A|(null)\n")
	require.NoError(t, err)
	assert.Nil(t, job.StartTime)
	assert.Nil(t, job.NodeList, "(null) node list should parse to nil")

	// An unparseable start time is ignored rather than fatal.
	job, err = c.parseSqueueOutput("1", "1|myjob|RUNNING|00:00:00|not-a-time|N/A|node001")
	require.NoError(t, err)
	assert.Nil(t, job.StartTime)

	// A cached spec is merged in.
	spec := &SLURMJobSpec{JobName: "cached", Nodes: 2, CPUsPerNode: 4, TimeLimit: 60, Command: "run"}
	submitted := time.Unix(1700000000, 0)
	c.jobs["1"] = &SLURMJob{SLURMJobID: "1", VirtEngineJobID: "ve-1", Spec: spec, SubmitTime: submitted}

	job, err = c.parseSqueueOutput("1", "1|myjob|RUNNING|00:00:00|N/A|N/A|node001")
	require.NoError(t, err)
	assert.Same(t, spec, job.Spec)
	assert.Equal(t, "ve-1", job.VirtEngineJobID)
	assert.Equal(t, submitted, job.SubmitTime)
}

func TestParseSacctOutput(t *testing.T) {
	c := newTestSSHClient(t, nil)

	// Empty output -> not found.
	_, err := c.parseSacctOutput("123", "")
	assert.ErrorIs(t, err, ErrJobNotFound)

	// Only step lines: falls back to the first line. Step rows carry six
	// fields (no JobName), so the format check still rejects them.
	out := "123.batch|COMPLETED|0:0|2026-01-02T03:04:05|2026-01-02T03:09:05|00:05:00\n"
	_, err = c.parseSacctOutput("123", out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected sacct output format")

	// Main job line present among step lines: the main entry wins.
	out = "123.batch|COMPLETED|0:0|2026-01-02T03:04:00|2026-01-02T03:04:30|00:00:30\n" +
		"123|myjob|COMPLETED|0:0|2026-01-02T03:04:05|2026-01-02T03:09:05|00:05:00\n"
	job, err := c.parseSacctOutput("123", out)
	require.NoError(t, err)
	assert.Equal(t, "123", job.SLURMJobID)
	assert.Equal(t, SLURMJobStateCompleted, job.State)
	assert.Equal(t, int32(0), job.ExitCode)
	require.NotNil(t, job.StartTime)
	require.NotNil(t, job.EndTime)

	// A non-zero exit code and a failing state.
	out = "999|myjob|FAILED|17:0|Unknown|Unknown|00:00:10\n"
	job, err = c.parseSacctOutput("999", out)
	require.NoError(t, err)
	assert.Equal(t, int32(17), job.ExitCode)
	assert.Equal(t, SLURMJobStateFailed, job.State)
	assert.Nil(t, job.StartTime, "Unknown start time should stay nil")
	assert.Nil(t, job.EndTime, "Unknown end time should stay nil")

	// Too few fields.
	_, err = c.parseSacctOutput("1", "1|only|two\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected sacct output format")

	// A cached record is merged.
	spec := &SLURMJobSpec{JobName: "cached"}
	c.jobs["777"] = &SLURMJob{SLURMJobID: "777", VirtEngineJobID: "ve-777", Spec: spec}
	job, err = c.parseSacctOutput("777", "777|myjob|CANCELLED|0:0|Unknown|Unknown|00:00:01\n")
	require.NoError(t, err)
	assert.Same(t, spec, job.Spec)
	assert.Equal(t, "ve-777", job.VirtEngineJobID)
}

func TestParseSacctMetrics(t *testing.T) {
	c := newTestSSHClient(t, nil)

	// No parsable lines -> not found.
	_, err := c.parseSacctMetrics("")
	assert.ErrorIs(t, err, ErrJobNotFound)
	_, err = c.parseSacctMetrics("too|few\n")
	assert.ErrorIs(t, err, ErrJobNotFound)

	// A single entry.
	m, err := c.parseSacctMetrics("1|00:05:00|00:06:30|512M|1G\n")
	require.NoError(t, err)
	assert.Equal(t, int64(300), m.WallClockSeconds)
	assert.Equal(t, int64(390), m.CPUTimeSeconds)
	assert.Equal(t, int64(512)*1024*1024, m.MaxRSSBytes)
	assert.Equal(t, int64(1024)*1024*1024, m.MaxVMSizeBytes)

	// Multiple entries keep the largest RSS/VMSize while the first entry
	// supplies the timings.
	m, err = c.parseSacctMetrics(
		"1|00:01:00|00:01:00|100M|200M\n" +
			"1.batch|00:09:00|00:09:00|900M|2G\n" +
			"short|line\n")
	require.NoError(t, err)
	assert.Equal(t, int64(60), m.WallClockSeconds, "first entry supplies the timings")
	assert.Equal(t, int64(900)*1024*1024, m.MaxRSSBytes, "largest MaxRSS wins")
	assert.Equal(t, int64(2)*1024*1024*1024, m.MaxVMSizeBytes, "largest MaxVMSize wins")
}

// --- ssh_client.go: pure helpers ---

func TestParseDuration_AllBranches(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"", 0},
		{"   ", 0},
		{"UNLIMITED", 0},
		{"INVALID", 0},
		{"45", 45},            // bare seconds
		{"01:30", 90},         // MM:SS
		{"01:00:00", 3600},    // HH:MM:SS
		{"1-00:00:00", 86400}, // DD-HH:MM:SS
		{"00:00:01.500", 1},   // millisecond suffix stripped
		{"2-03:04:05.250", 2*86400 + 3*3600 + 4*60 + 5},
		{"bad", 0},             // unparsable components yield 0
		{"1-xyz:00:00", 86400}, // unparsable day component ignored
	}
	for _, tc := range cases {
		if got := parseDuration(tc.in); got != tc.want {
			t.Errorf("parseDuration(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseMemory_AllSuffixes(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"", 0},
		{"1024", 1024},
		{"1K", 1024},
		{"1k", 1024},
		{"1M", 1024 * 1024},
		{"1G", 1024 * 1024 * 1024},
		{"1T", 1024 * 1024 * 1024 * 1024},
		{"1X", 0}, // unknown suffix: Atoi("1X") fails -> 0
	}
	for _, tc := range cases {
		if got := parseMemory(tc.in); got != tc.want {
			t.Errorf("parseMemory(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseNodeList_AllBranches(t *testing.T) {
	assert.Nil(t, parseNodeList(""))
	assert.Nil(t, parseNodeList("   "))
	assert.Nil(t, parseNodeList(nullString))
	assert.Equal(t, []string{"node001", "node002"}, parseNodeList("node001,node002"))
	// Bracket/range form is returned verbatim (not yet expanded).
	assert.Equal(t, []string{"node[001-004]"}, parseNodeList("node[001-004]"))
}

func TestMapSLURMState_AllBranches(t *testing.T) {
	cases := []struct {
		in   string
		want SLURMJobState
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
		{"  running  ", SLURMJobStateRunning}, // trimmed + upper-cased
		{"SOMETHING", SLURMJobState("SOMETHING")},
	}
	for _, tc := range cases {
		if got := mapSLURMState(tc.in); got != tc.want {
			t.Errorf("mapSLURMState(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseGRES_AllBranches(t *testing.T) {
	count, gpuType := parseGRES("gpu")
	assert.Equal(t, int32(0), count)
	assert.Empty(t, gpuType)

	count, gpuType = parseGRES("mic:2")
	assert.Equal(t, int32(0), count, "non-gpu resources are ignored")
	assert.Empty(t, gpuType)

	count, gpuType = parseGRES("gpu:4")
	assert.Equal(t, int32(4), count)
	assert.Empty(t, gpuType, "two-field form has no type")

	count, gpuType = parseGRES("gpu:a100:8")
	assert.Equal(t, int32(8), count)
	assert.Equal(t, "a100", gpuType)
}

// --- mock_client.go ---

func TestMockSLURMClient_ConnectionGuards(t *testing.T) {
	ctx := context.Background()
	c := NewMockSLURMClient()
	spec := &SLURMJobSpec{JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true"}

	require.False(t, c.IsConnected())
	_, err := c.SubmitJob(ctx, spec)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
	assert.ErrorIs(t, c.CancelJob(ctx, "1"), ErrSLURMNotConnected)
	_, err = c.GetJobStatus(ctx, "1")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
	_, err = c.GetJobAccounting(ctx, "1")
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
	_, err = c.ListPartitions(ctx)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
	_, err = c.ListNodes(ctx)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	require.NoError(t, c.Connect(ctx))
	require.True(t, c.IsConnected())
	require.NoError(t, c.Disconnect())
	require.False(t, c.IsConnected())
}

func TestMockSLURMClient_JobLifecycle(t *testing.T) {
	ctx := context.Background()
	c := NewMockSLURMClient()
	require.NoError(t, c.Connect(ctx))

	spec := &SLURMJobSpec{
		JobName: "j", Nodes: 2, CPUsPerNode: 4, MemoryMB: 1024, GPUs: 2, TimeLimit: 60, Command: "true",
	}
	id, err := c.SubmitJob(ctx, spec)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// Unknown job.
	_, err = c.GetJobStatus(ctx, "does-not-exist")
	assert.ErrorIs(t, err, ErrJobNotFound)
	assert.ErrorIs(t, c.CancelJob(ctx, "does-not-exist"), ErrJobNotFound)
	_, err = c.GetJobAccounting(ctx, "does-not-exist")
	assert.ErrorIs(t, err, ErrJobNotFound)

	// Accounting is refused while the job is not terminal.
	_, err = c.GetJobAccounting(ctx, id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job not completed")

	// Simulate completion and re-query accounting.
	require.NoError(t, c.SimulateJobCompletion(id, true, 0))
	metrics, err := c.GetJobAccounting(ctx, id)
	require.NoError(t, err)
	assert.Positive(t, metrics.MaxRSSBytes)
	assert.Equal(t, int64(1024)*1024*1024, metrics.MaxRSSBytes)
	assert.Equal(t, int64(1024)*1024*1024*2, metrics.MaxVMSizeBytes)

	// A completed job cannot be cancelled.
	err = c.CancelJob(ctx, id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel completed job")

	// Simulating completion of an unknown job fails.
	assert.ErrorIs(t, c.SimulateJobCompletion("nope", false, 1), ErrJobNotFound)

	// A failing completion records the exit code and is itself terminal, so a
	// second cancellation is refused.
	id2, err := c.SubmitJob(ctx, spec)
	require.NoError(t, err)
	require.NoError(t, c.SimulateJobCompletion(id2, false, 3))
	job, err := c.GetJobStatus(ctx, id2)
	require.NoError(t, err)
	assert.Equal(t, SLURMJobStateFailed, job.State)
	assert.Equal(t, int32(3), job.ExitCode)
	assert.Error(t, c.CancelJob(ctx, id2), "a failed job is terminal and cannot be cancelled")

	// A fresh pending job can be cancelled.
	id3, err := c.SubmitJob(ctx, spec)
	require.NoError(t, err)
	require.NoError(t, c.CancelJob(ctx, id3))
	job, err = c.GetJobStatus(ctx, id3)
	require.NoError(t, err)
	assert.Equal(t, SLURMJobStateCancelled, job.State)
	require.NotNil(t, job.EndTime)
}

func TestMockSLURMClient_Inventory(t *testing.T) {
	ctx := context.Background()
	c := NewMockSLURMClient()
	require.NoError(t, c.Connect(ctx))

	parts, err := c.ListPartitions(ctx)
	require.NoError(t, err)
	require.Len(t, parts, 2)
	assert.Equal(t, "default", parts[0].Name)
	assert.Equal(t, "gpu", parts[1].Name)
	assert.Equal(t, []string{"gpu", "nvidia"}, parts[1].Features)

	nodes, err := c.ListNodes(ctx)
	require.NoError(t, err)
	require.Len(t, nodes, 4)
	assert.Equal(t, "node001", nodes[0].Name)
	assert.Equal(t, int32(8), nodes[3].GPUs)
	assert.Equal(t, "nvidia-a100", nodes[3].GPUType)
}

// --- adapter.go ---

type failingSigner struct{ err error }

func (f failingSigner) Sign([]byte) ([]byte, error) { return nil, f.err }
func (f failingSigner) Verify([]byte, []byte) bool  { return false }
func (f failingSigner) GetProviderAddress() string  { return "ve1fail" }

type failingClient struct {
	*MockSLURMClient
	connectErr error
	submitErr  error
	cancelErr  error
	statusErr  error
}

func (f *failingClient) Connect(ctx context.Context) error {
	if f.connectErr != nil {
		return f.connectErr
	}
	return f.MockSLURMClient.Connect(ctx)
}

func (f *failingClient) SubmitJob(ctx context.Context, spec *SLURMJobSpec) (string, error) {
	if f.submitErr != nil {
		return "", f.submitErr
	}
	return f.MockSLURMClient.SubmitJob(ctx, spec)
}

func (f *failingClient) CancelJob(ctx context.Context, id string) error {
	if f.cancelErr != nil {
		return f.cancelErr
	}
	return f.MockSLURMClient.CancelJob(ctx, id)
}

func (f *failingClient) GetJobStatus(ctx context.Context, id string) (*SLURMJob, error) {
	if f.statusErr != nil {
		return nil, f.statusErr
	}
	return f.MockSLURMClient.GetJobStatus(ctx, id)
}

func TestJobStatusReport_Hash(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	r1 := &JobStatusReport{
		ProviderAddress: "ve1", VirtEngineJobID: "ve-1", SLURMJobID: "1",
		State: SLURMJobStateRunning, ExitCode: 0, Timestamp: ts,
	}
	r2 := &JobStatusReport{
		ProviderAddress: "ve1", VirtEngineJobID: "ve-1", SLURMJobID: "1",
		State: SLURMJobStateRunning, ExitCode: 0, Timestamp: ts,
	}
	h1, h2 := r1.Hash(), r2.Hash()
	assert.Len(t, h1, 32, "SHA-256 digest")
	assert.Equal(t, h1, h2, "Hash is deterministic for identical reports")

	// Any signed field changes the hash.
	r2.ExitCode = 1
	assert.NotEqual(t, h1, r2.Hash())

	r3 := *r1
	r3.State = SLURMJobStateFailed
	assert.NotEqual(t, h1, r3.Hash())
}

func TestSLURMAdapter_StartConnectFailure(t *testing.T) {
	client := &failingClient{
		MockSLURMClient: NewMockSLURMClient(),
		connectErr:      assert.AnError,
	}
	a := NewSLURMAdapter(DefaultSLURMConfig(), client, testSigner{addr: testProviderAddr})

	err := a.Start(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to SLURM")
	assert.False(t, a.IsRunning(), "a failed Start must leave the adapter stopped")

	// Start on an already-running adapter is a no-op. Clear the injected
	// connect error first so the client can actually come up.
	client.connectErr = nil
	require.NoError(t, client.Connect(context.Background()))
	require.NoError(t, a.Start(context.Background()))
	require.True(t, a.IsRunning())
	require.NoError(t, a.Start(context.Background())) // idempotent
	require.NoError(t, a.Stop())
	require.NoError(t, a.Stop()) // idempotent
}

func TestSLURMAdapter_SubmitAndCancelErrors(t *testing.T) {
	ctx := context.Background()
	client := NewMockSLURMClient()
	a := NewSLURMAdapter(DefaultSLURMConfig(), client, testSigner{addr: testProviderAddr})
	require.NoError(t, client.Connect(ctx))
	require.NoError(t, a.Start(ctx))
	t.Cleanup(func() { _ = a.Stop() })

	// Invalid spec is rejected before the client is called.
	_, err := a.SubmitJob(ctx, "ve-bad", &SLURMJobSpec{})
	assert.ErrorIs(t, err, ErrInvalidJobSpec)

	// Client-side submission failure is wrapped.
	failing := &failingClient{MockSLURMClient: NewMockSLURMClient(), submitErr: assert.AnError}
	require.NoError(t, failing.Connect(ctx))
	a2 := NewSLURMAdapter(DefaultSLURMConfig(), failing, testSigner{addr: testProviderAddr})
	require.NoError(t, a2.Start(ctx))
	_, err = a2.SubmitJob(ctx, "ve-x", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrJobSubmissionFailed)
	_ = a2.Stop()

	// Cancelling an unknown job.
	assert.ErrorIs(t, a.CancelJob(ctx, "ve-unknown"), ErrJobNotFound)

	// A successful submit registers the job for both lookup maps.
	job, err := a.SubmitJob(ctx, "ve-1", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.NoError(t, err)
	require.NotEmpty(t, job.SLURMJobID)
	assert.Equal(t, SLURMJobStatePending, job.State)

	require.NoError(t, a.CancelJob(ctx, "ve-1"))

	// Inspect the adapter's own record: CancelJob stamps StatusMessage and
	// EndTime locally, but a subsequent GetJobStatus replaces the cached job
	// with the client's copy, which does not carry those fields.
	a.mu.RLock()
	local, ok := a.jobs[job.SLURMJobID]
	a.mu.RUnlock()
	require.True(t, ok)
	assert.Equal(t, SLURMJobStateCancelled, local.State)
	assert.Equal(t, "Cancelled by user", local.StatusMessage)
	require.NotNil(t, local.EndTime)

	// The client also reports the job as cancelled.
	cached, err := a.GetJobStatus(ctx, "ve-1")
	require.NoError(t, err)
	assert.Equal(t, SLURMJobStateCancelled, cached.State)
}

func TestSLURMAdapter_CancelClientError(t *testing.T) {
	ctx := context.Background()
	client := NewMockSLURMClient()
	require.NoError(t, client.Connect(ctx))

	a := NewSLURMAdapter(DefaultSLURMConfig(), client, testSigner{addr: testProviderAddr})
	require.NoError(t, a.Start(ctx))
	_, err := a.SubmitJob(ctx, "ve-1", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.NoError(t, err)
	require.NoError(t, a.Stop())

	// Swap in a client whose CancelJob fails, then re-register the job so the
	// adapter can find it.
	bad := &failingClient{MockSLURMClient: NewMockSLURMClient(), cancelErr: assert.AnError}
	require.NoError(t, bad.Connect(ctx))
	a2 := NewSLURMAdapter(DefaultSLURMConfig(), bad, testSigner{addr: testProviderAddr})
	require.NoError(t, a2.Start(ctx))
	defer func() { _ = a2.Stop() }()
	_, err = a2.SubmitJob(ctx, "ve-2", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	// The failing client refuses to submit, so register the mapping directly.
	if err != nil {
		a2.mu.Lock()
		a2.jobs["s1"] = &SLURMJob{SLURMJobID: "s1", VirtEngineJobID: "ve-2", State: SLURMJobStateRunning}
		a2.jobMapping["ve-2"] = "s1"
		a2.mu.Unlock()
	}

	err = a2.CancelJob(ctx, "ve-2")
	if err != nil {
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrJobCancellationFailed)
	}
}

func TestSLURMAdapter_GetJobStatusFallsBackToCache(t *testing.T) {
	ctx := context.Background()
	client := NewMockSLURMClient()
	require.NoError(t, client.Connect(ctx))

	a := NewSLURMAdapter(DefaultSLURMConfig(), client, testSigner{addr: testProviderAddr})
	require.NoError(t, a.Start(ctx))
	defer func() { _ = a.Stop() }()

	// Unknown VE job.
	_, err := a.GetJobStatus(ctx, "ve-unknown")
	assert.ErrorIs(t, err, ErrJobNotFound)

	// A mapping whose job record is missing (inconsistent state).
	a.mu.Lock()
	a.jobMapping["ve-orphan"] = "gone"
	a.mu.Unlock()
	_, err = a.GetJobStatus(ctx, "ve-orphan")
	assert.ErrorIs(t, err, ErrJobNotFound)

	// A job whose status query fails returns the cached copy, not an error.
	_, err = a.SubmitJob(ctx, "ve-1", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.NoError(t, err)

	broken := &failingClient{MockSLURMClient: NewMockSLURMClient(), statusErr: assert.AnError}
	require.NoError(t, broken.Connect(ctx))
	a2 := NewSLURMAdapter(DefaultSLURMConfig(), broken, testSigner{addr: testProviderAddr})
	require.NoError(t, a2.Start(ctx))
	defer func() { _ = a2.Stop() }()
	a2.mu.Lock()
	a2.jobs["s9"] = &SLURMJob{SLURMJobID: "s9", VirtEngineJobID: "ve-9", State: SLURMJobStateRunning}
	a2.jobMapping["ve-9"] = "s9"
	a2.mu.Unlock()

	cached, err := a2.GetJobStatus(ctx, "ve-9")
	require.NoError(t, err, "a failed SLURM query must fall back to the cached job")
	assert.Equal(t, "s9", cached.SLURMJobID)
	assert.Equal(t, SLURMJobStateRunning, cached.State)
}

func TestSLURMAdapter_UpdateJobStatuses(t *testing.T) {
	ctx := context.Background()
	client := NewMockSLURMClient()
	require.NoError(t, client.Connect(ctx))

	a := NewSLURMAdapter(DefaultSLURMConfig(), client, testSigner{addr: testProviderAddr})
	require.NoError(t, a.Start(ctx))
	defer func() { _ = a.Stop() }()

	// A terminal job is not polled; a non-terminal one is.
	a.mu.Lock()
	a.jobs["terminal"] = &SLURMJob{SLURMJobID: "terminal", State: SLURMJobStateCompleted}
	a.mu.Unlock()

	// Submit a real job so the mock client knows about it, then drive it to
	// completion so the poll observes a terminal transition.
	job, err := a.SubmitJob(ctx, "ve-1", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, MemoryMB: 512, TimeLimit: 60, Command: "true",
	})
	require.NoError(t, err)
	a.mu.Lock()
	assert.Equal(t, job.SLURMJobID, a.jobMapping["ve-1"])
	a.mu.Unlock()

	require.NoError(t, client.SimulateJobCompletion(job.SLURMJobID, true, 0))

	a.updateJobStatuses()

	a.mu.RLock()
	updated := a.jobs[job.SLURMJobID]
	terminal, stillThere := a.jobs["terminal"]
	a.mu.RUnlock()

	require.NotNil(t, updated)
	assert.Equal(t, SLURMJobStateCompleted, updated.State, "the poll must write the new state back")
	require.NotNil(t, updated.UsageMetrics, "a terminal job picks up accounting metrics")
	assert.Positive(t, updated.UsageMetrics.MaxRSSBytes)

	// The already-terminal job is left untouched by the poll.
	require.True(t, stillThere)
	assert.Equal(t, SLURMJobStateCompleted, terminal.State)
	assert.Nil(t, terminal.UsageMetrics, "terminal jobs are not polled, so no metrics are fetched")
}

func TestSLURMAdapter_UpdateJobStatuses_SkipsFailedQueries(t *testing.T) {
	ctx := context.Background()
	broken := &failingClient{MockSLURMClient: NewMockSLURMClient(), statusErr: assert.AnError}
	require.NoError(t, broken.Connect(ctx))

	a := NewSLURMAdapter(DefaultSLURMConfig(), broken, testSigner{addr: testProviderAddr})
	require.NoError(t, a.Start(ctx))
	defer func() { _ = a.Stop() }()

	a.mu.Lock()
	a.jobs["s1"] = &SLURMJob{SLURMJobID: "s1", State: SLURMJobStateRunning}
	a.mu.Unlock()

	// A status query that fails must leave the cached job untouched.
	a.updateJobStatuses()
	a.mu.RLock()
	defer a.mu.RUnlock()
	assert.Equal(t, SLURMJobStateRunning, a.jobs["s1"].State)
}

func TestSLURMAdapter_ListNodesAndGetJobsByCluster(t *testing.T) {
	ctx := context.Background()

	// Not running -> both refuse.
	client := NewMockSLURMClient()
	require.NoError(t, client.Connect(ctx))
	a := NewSLURMAdapter(DefaultSLURMConfig(), client, testSigner{addr: testProviderAddr})

	_, err := a.ListNodes(ctx)
	assert.ErrorIs(t, err, ErrSLURMNotConnected)
	_, err = a.SubmitJob(ctx, "ve-1", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	assert.ErrorIs(t, err, ErrSLURMNotConnected)

	require.NoError(t, a.Start(ctx))
	defer func() { _ = a.Stop() }()

	nodes, err := a.ListNodes(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, nodes)

	// An empty adapter reports an empty (non-nil) slice.
	assert.Empty(t, a.GetJobsByCluster())

	_, err = a.SubmitJob(ctx, "ve-1", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.NoError(t, err)
	_, err = a.SubmitJob(ctx, "ve-2", &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.NoError(t, err)

	jobs := a.GetJobsByCluster()
	assert.Len(t, jobs, 2)
}

func TestSLURMAdapter_CreateStatusReport(t *testing.T) {
	job := &SLURMJob{
		SLURMJobID: "1", VirtEngineJobID: "ve-1", State: SLURMJobStateCompleted,
		ExitCode: 2, StatusMessage: "done",
		UsageMetrics: &SLURMUsageMetrics{WallClockSeconds: 10},
	}

	ok := NewSLURMAdapter(DefaultSLURMConfig(), NewMockSLURMClient(), testSigner{addr: testProviderAddr})
	report, err := ok.CreateStatusReport(job)
	require.NoError(t, err)
	assert.Equal(t, testProviderAddr, report.ProviderAddress)
	assert.Equal(t, "ve-1", report.VirtEngineJobID)
	assert.Equal(t, "1", report.SLURMJobID)
	assert.Equal(t, SLURMJobStateCompleted, report.State)
	assert.Equal(t, int32(2), report.ExitCode)
	assert.Same(t, job.UsageMetrics, report.UsageMetrics)
	assert.NotEmpty(t, report.Signature)

	// The signature must decode as hex.
	raw, err := hex.DecodeString(report.Signature)
	require.NoError(t, err, "signature should be hex encoded")
	assert.NotEmpty(t, raw)

	// A signing failure is wrapped.
	bad := NewSLURMAdapter(DefaultSLURMConfig(), NewMockSLURMClient(), failingSigner{err: assert.AnError})
	_, err = bad.CreateStatusReport(job)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sign report")
}
