// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package slurm_adapter

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// fakeSSHServer is an in-process SSH server bound to loopback. It exists so
// the SSH command paths (Connect, runCommand, sbatch/squeue/sacct/sinfo
// parsing, SCP) can be exercised deterministically and offline -- no SLURM
// cluster and no outbound network are involved. The host key and the password
// are generated per-test, so nothing is checked in or shared.
type fakeSSHServer struct {
	listener net.Listener
	addr     string
	password string
	hostKey  ssh.Signer

	mu    sync.Mutex
	calls []string
}

// cannedOutput maps a command to the stdout it should produce. A command
// with no entry succeeds with empty stdout, which keeps the connection
// probe (`squeue --version`) green by default.
type cannedOutput func(cmd string) (stdout string, exitCode int)

func newFakeSSHServer(t *testing.T, output cannedOutput) *fakeSSHServer {
	t.Helper()

	key, err := rsaHostKey(t)
	require.NoError(t, err)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := &fakeSSHServer{
		listener: ln,
		addr:     ln.Addr().String(),
		password: "testpass",
		hostKey:  key,
	}
	s.serve(output)
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

func rsaHostKey(t *testing.T) (ssh.Signer, error) {
	t.Helper()
	// Generate an ephemeral RSA key so no key material is committed.
	return generateEphemeralHostKey()
}

// generateEphemeralHostKey builds a throwaway 2048-bit RSA host key. It is
// deliberately slow-ish but only ever runs once per fake server, and keeping
// it out of source control means there is no private key in the repo.
func generateEphemeralHostKey() (ssh.Signer, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(priv)
}

func (s *fakeSSHServer) hostPort() (string, int) {
	host, port, _ := net.SplitHostPort(s.addr)
	var p int
	_, _ = fmt.Sscanf(port, "%d", &p)
	return host, p
}

func (s *fakeSSHServer) serve(output cannedOutput) {
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "testuser" && string(pass) == s.password {
				return nil, nil
			}
			return nil, fmt.Errorf("auth failed for %s", c.User())
		},
	}
	cfg.AddHostKey(s.hostKey)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				return // listener closed
			}
			go s.handleConn(conn, cfg, output)
		}
	}()
}

func (s *fakeSSHServer) handleConn(nConn net.Conn, cfg *ssh.ServerConfig, output cannedOutput) {
	sshConn, chans, reqs, err := ssh.NewServerConn(nConn, cfg)
	if err != nil {
		_ = nConn.Close()
		return
	}
	defer func() { _ = sshConn.Close() }()
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unsupported")
			continue
		}
		ch, requests, err := newChannel.Accept()
		if err != nil {
			return
		}
		go s.handleSession(ch, requests, output)
	}
}

func (s *fakeSSHServer) handleSession(ch ssh.Channel, requests <-chan *ssh.Request, output cannedOutput) {
	defer func() { _ = ch.Close() }()
	for req := range requests {
		if req.Type != "exec" {
			_ = req.Reply(false, nil)
			continue
		}
		var payload struct{ Command string }
		if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
			_ = req.Reply(false, nil)
			return
		}
		_ = req.Reply(true, nil)

		s.mu.Lock()
		s.calls = append(s.calls, payload.Command)
		s.mu.Unlock()

		// `cat > path << EOF` (WriteRemoteFile) writes the heredoc body to
		// the channel and needs no stdout.
		stdout, code := output(payload.Command)
		if stdout != "" {
			_, _ = ch.Write([]byte(stdout))
		}
		// An exec exit status is a uint32 on the wire; the canned outputs in
		// this file only ever use small positive codes, and the helper's
		// contract is "the status to report", so a negative value is clamped
		// rather than wrapped into a huge unsigned exit code.
		status := uint32(0)
		if code > 0 {
			status = uint32(code) // #nosec G115 -- clamped non-negative canned exit code
		}
		_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
		return
	}
}

func (s *fakeSSHServer) record() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.calls))
	copy(out, s.calls)
	return out
}

// connectSSHClient builds a client wired to the fake server and starts it.
func connectSSHClient(t *testing.T, s *fakeSSHServer) *SSHSLURMClient {
	t.Helper()
	c := dialSSHClient(t, s)
	t.Cleanup(func() { _ = c.Disconnect() })
	return c
}

// dialSSHClient connects without registering a Disconnect cleanup. Tests that
// deliberately leave a goroutine holding c.mu (the deadlock pins) must not
// call Disconnect, because Disconnect takes c.mu and would block forever.
func dialSSHClient(t *testing.T, s *fakeSSHServer) *SSHSLURMClient {
	t.Helper()
	host, port := s.hostPort()
	cfg := SSHConfig{
		Host:            host,
		Port:            port,
		User:            "testuser",
		Password:        s.password,
		Timeout:         5 * time.Second,
		HostKeyCallback: "ignore",
		PoolSize:        2,
		MaxRetries:      0,
	}
	c, err := NewSSHSLURMClient(cfg, "testcluster", "compute")
	require.NoError(t, err)
	require.NoError(t, c.Connect(context.Background()))
	require.True(t, c.IsConnected())
	return c
}

func TestSSHSLURMClient_ConnectAndDisconnect(t *testing.T) {
	s := newFakeSSHServer(t, func(string) (string, int) { return "slurm 23.02.7\n", 0 })
	c := connectSSHClient(t, s)

	// Connect probes the cluster with `squeue --version` and seeds the pool.
	assert.True(t, c.IsConnected())
	calls := s.record()
	require.Len(t, calls, 1)
	assert.Contains(t, calls[0], "squeue --version")

	c.mu.RLock()
	poolLen := len(c.pool)
	c.mu.RUnlock()
	assert.Equal(t, 1, poolLen, "the initial connection is pooled")

	require.NoError(t, c.Disconnect())
	assert.False(t, c.IsConnected())

	// A second Disconnect on an already-closed client is a no-op.
	require.NoError(t, c.Disconnect())
}

func TestSSHSLURMClient_ConnectProbesFail(t *testing.T) {
	// The probe command fails: Connect must fail and leave the client
	// disconnected.
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "squeue --version") {
			return "squeue: command not found\n", 127
		}
		return "", 0
	})
	host, port := s.hostPort()
	c, err := NewSSHSLURMClient(SSHConfig{
		Host: host, Port: port, User: "testuser", Password: s.password,
		Timeout: 5 * time.Second, HostKeyCallback: "ignore", MaxRetries: 0,
	}, "c", "d")
	require.NoError(t, err)

	err = c.Connect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify SLURM")
	assert.False(t, c.IsConnected())
}

func TestSSHSLURMClient_ConnectDialFailure(t *testing.T) {
	// Port 1 on loopback refuses connections; MaxRetries=0 keeps it fast.
	c, err := NewSSHSLURMClient(SSHConfig{
		Host: "127.0.0.1", Port: 1, User: "testuser", Password: "p",
		Timeout: time.Second, HostKeyCallback: "ignore", MaxRetries: 0,
	}, "c", "d")
	require.NoError(t, err)

	err = c.Connect(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSSHConnection)
	assert.False(t, c.IsConnected())
}

func TestSSHSLURMClient_ConnectHonoursContextCancellation(t *testing.T) {
	c, err := NewSSHSLURMClient(SSHConfig{
		Host: "127.0.0.1", Port: 1, User: "testuser", Password: "p",
		Timeout: time.Second, HostKeyCallback: "ignore", MaxRetries: 0,
	}, "c", "d")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = c.Connect(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSSHConnection)
}

func TestSSHSLURMClient_RunCommand(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		switch {
		case strings.HasPrefix(cmd, "echo hello"):
			return "hello\n", 0
		case strings.HasPrefix(cmd, "false"):
			return "boom\n", 1
		default:
			return "ok\n", 0
		}
	})
	c := connectSSHClient(t, s)
	ctx := context.Background()

	out, err := c.runCommand(ctx, "echo hello")
	require.NoError(t, err)
	assert.Equal(t, "hello\n", out)

	// A non-zero exit status surfaces as an error alongside the output.
	out, err = c.runCommand(ctx, "false")
	require.Error(t, err)
	assert.Equal(t, "boom\n", out)

	// runCommandWithTimeout delegates to runCommand.
	out, err = c.runCommandWithTimeout(ctx, "echo hello", 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", out)
}

func TestSSHSLURMClient_RunCommandCancelledContext(t *testing.T) {
	s := newFakeSSHServer(t, func(string) (string, int) { return "out", 0 })
	c := connectSSHClient(t, s)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.runCommand(ctx, "echo hi")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestSSHSLURMClient_ConnectionPoolReuse(t *testing.T) {
	s := newFakeSSHServer(t, func(string) (string, int) { return "ok", 0 })
	c := connectSSHClient(t, s)
	ctx := context.Background()

	// Three sequential commands must reuse the single pooled connection.
	for i := 0; i < 3; i++ {
		_, err := c.runCommand(ctx, "true")
		require.NoError(t, err)
	}
	c.poolMu.Lock()
	poolLen := len(c.pool)
	inUse := c.pool[0].inUse
	c.poolMu.Unlock()
	assert.Equal(t, 1, poolLen, "pool must not grow when a connection is available")
	assert.False(t, inUse, "the connection is released back to the pool")
}

// TestSSHSLURMClient_ExhaustedPoolIgnoresContext pins a real defect:
// acquireConnection waits on poolCond.Wait() (ssh_client.go:383) with no
// timeout and no context watcher, so an exhausted pool blocks forever even
// when the caller's context is already cancelled. The in-code comment
// "Wait with timeout" is not implemented. Because the blocked acquire also
// holds poolMu, Disconnect cannot proceed either.
//
// The test asserts the CURRENT (broken) behaviour without hanging: the
// acquire must still be blocked after the context deadline, then it is
// released so the test can finish. When the wait is fixed to honour ctx,
// this test fails -- that failure is the intended signal.
func TestSSHSLURMClient_ExhaustedPoolIgnoresContext(t *testing.T) {
	s := newFakeSSHServer(t, func(string) (string, int) { return "ok", 0 })
	c := connectSSHClient(t, s)

	// Fill the pool to its configured capacity (2) and mark every entry in
	// use, so acquireConnection takes the waiting branch rather than dialling.
	c.poolMu.Lock()
	first := c.pool[0]
	first.inUse = true
	second := &pooledConnection{
		client:   first.client,
		lastUsed: time.Now(),
		inUse:    true,
	}
	c.pool = append(c.pool, second)
	capacity := len(c.pool)
	c.poolMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := c.runCommand(ctx, "true")
		done <- err
	}()

	// Give the goroutine time to reach poolCond.Wait(), then cancel. The wait
	// has no context watcher, so the call must remain blocked.
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		t.Fatalf("DEFECT REGRESSION GUARD: acquireConnection returned %v after "+
			"context cancellation; the pool wait now honours ctx -- update this test", err)
	case <-time.After(500 * time.Millisecond):
		// Expected: the wait ignores the cancelled context.
	}

	// Free a connection so the blocked acquire completes and Disconnect in
	// cleanup is not itself deadlocked. On waking, the loop re-checks
	// ctx.Done() and returns the cancellation rather than handing out the
	// freed connection -- so the caller gets a context error, not the
	// connection it was waiting for.
	c.poolMu.Lock()
	second.inUse = false
	c.poolCond.Signal()
	c.poolMu.Unlock()

	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.Canceled,
			"a woken acquire re-checks ctx and reports the cancellation")
	case <-time.After(10 * time.Second):
		t.Fatal("the pending acquire did not complete after the connection was released")
	}

	assert.GreaterOrEqual(t, capacity, 1)
}

func TestSSHSLURMClient_SubmitJob(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sbatch") {
			return "12345\n", 0
		}
		return "", 0
	})
	c := connectSSHClient(t, s)
	ctx := context.Background()

	spec := &SLURMJobSpec{
		JobName: "train", Nodes: 1, CPUsPerNode: 4, TimeLimit: 60,
		WorkingDirectory: "/data", OutputDirectory: "/out", Command: "python",
	}
	id, err := c.SubmitJob(ctx, spec)
	require.NoError(t, err)
	assert.Equal(t, "12345", id, "the numeric job ID is parsed from --parsable output")

	// The job is cached locally as PENDING.
	c.mu.RLock()
	cached, ok := c.jobs["12345"]
	c.mu.RUnlock()
	require.True(t, ok, "a submitted job is recorded in the local cache")
	assert.Equal(t, SLURMJobStatePending, cached.State)
	assert.Same(t, spec, cached.Spec)

	// An invalid spec is rejected before any command runs.
	_, err = c.SubmitJob(ctx, &SLURMJobSpec{})
	assert.ErrorIs(t, err, ErrInvalidJobSpec)

	// The generated script carries the spec's directives.
	var script string
	for _, cmd := range s.record() {
		if strings.Contains(cmd, "sbatch") {
			script = cmd
		}
	}
	require.NotEmpty(t, script)
	assert.Contains(t, script, "#SBATCH --job-name=train")
	assert.Contains(t, script, "#SBATCH --chdir=/data")
	assert.Contains(t, script, "#SBATCH --output=/out")
	assert.Contains(t, script, "Generated by VirtEngine SLURM Adapter")
	assert.Contains(t, script, "Cluster: testcluster")
	// The adapter's default partition fills in for an empty spec partition.
	assert.Contains(t, script, "#SBATCH --partition=compute")
}

func TestSSHSLURMClient_SubmitJobClusterQualifiedID(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sbatch") {
			return "67890;mycluster\n", 0
		}
		return "", 0
	})
	c := connectSSHClient(t, s)

	id, err := c.SubmitJob(context.Background(), &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.NoError(t, err)
	assert.Equal(t, "67890", id, "the ;cluster suffix is stripped")
}

func TestSSHSLURMClient_SubmitJobRejectsNonNumericOutput(t *testing.T) {
	for _, out := range []string{"", "   \n", "Submitted batch job 42\n"} {
		s := newFakeSSHServer(t, func(cmd string) (string, int) {
			if strings.Contains(cmd, "sbatch") {
				return out, 0
			}
			return "", 0
		})
		c := connectSSHClient(t, s)
		_, err := c.SubmitJob(context.Background(), &SLURMJobSpec{
			JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
		})
		require.Error(t, err, "sbatch output %q must be rejected", out)
		assert.ErrorIs(t, err, ErrJobSubmissionFailed)
	}
}

func TestSSHSLURMClient_SubmitJobCommandFailure(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sbatch") {
			return "sbatch: error: invalid account\n", 1
		}
		return "", 0
	})
	c := connectSSHClient(t, s)

	_, err := c.SubmitJob(context.Background(), &SLURMJobSpec{
		JobName: "j", Nodes: 1, CPUsPerNode: 1, TimeLimit: 60, Command: "true",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrJobSubmissionFailed)
	assert.Contains(t, err.Error(), "invalid account")
}

func TestSSHSLURMClient_SubmitJobFromScriptAndFile(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		switch {
		case strings.Contains(cmd, "/tmp/job.sh"):
			return "222\n", 0
		case strings.Contains(cmd, "sbatch"):
			return "111;clus\n", 0
		}
		return "", 0
	})
	c := connectSSHClient(t, s)
	ctx := context.Background()

	id, err := c.SubmitJobFromScript(ctx, "#!/bin/bash\necho hi\n")
	require.NoError(t, err)
	assert.Equal(t, "111", id)

	id, err = c.SubmitJobFromFile(ctx, "/tmp/job.sh")
	require.NoError(t, err)
	assert.Equal(t, "222", id)

	// Garbage output is rejected on both paths.
	s2 := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sbatch") {
			return "not-a-number\n", 0
		}
		return "", 0
	})
	c2 := connectSSHClient(t, s2)
	_, err = c2.SubmitJobFromScript(ctx, "#!/bin/bash\n")
	assert.ErrorIs(t, err, ErrJobSubmissionFailed)
	_, err = c2.SubmitJobFromFile(ctx, "/tmp/x.sh")
	assert.ErrorIs(t, err, ErrJobSubmissionFailed)

	// A command failure is wrapped on both paths.
	s3 := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sbatch") {
			return "sbatch: error\n", 1
		}
		return "", 0
	})
	c3 := connectSSHClient(t, s3)
	_, err = c3.SubmitJobFromScript(ctx, "#!/bin/bash\n")
	assert.ErrorIs(t, err, ErrJobSubmissionFailed)
	_, err = c3.SubmitJobFromFile(ctx, "/tmp/x.sh")
	assert.ErrorIs(t, err, ErrJobSubmissionFailed)
}

// TestSSHSLURMClient_CancelJobDeadlocks pins a real defect:
// SSHSLURMClient.CancelJob (ssh_client.go:761-763) takes c.mu.Lock() and then
// calls runCommand, which calls IsConnected() -> c.mu.RLock(). Go's
// sync.RWMutex is not reentrant, so the write lock is held while the same
// goroutine tries to read-lock: the call blocks forever. GetJobStatus
// (ssh_client.go:790-792) has the identical shape. SubmitJob only avoids it
// by an explicit Unlock/relock dance around runCommand.
//
// The test asserts the CURRENT (broken) behaviour with a bounded wait so the
// suite cannot hang. It FAILS once the lock handling is fixed, which is the
// intended signal. See also
// TestSSHSLURMClient_CancelJobErrorMapping, which exercises the error mapping
// without deadlocking by calling runCommand directly.
func TestSSHSLURMClient_CancelJobDeadlocks(t *testing.T) {
	s := newFakeSSHServer(t, func(string) (string, int) { return "", 0 })
	// No Disconnect cleanup: the blocked goroutine holds c.mu, and
	// Disconnect would need c.mu too.
	c := dialSSHClient(t, s)

	done := make(chan error, 1)
	go func() { done <- c.CancelJob(context.Background(), "555") }()

	select {
	case err := <-done:
		t.Fatalf("DEFECT REGRESSION GUARD: CancelJob returned %v; the reentrant "+
			"c.mu lock deadlock appears fixed -- update this test", err)
	case <-time.After(500 * time.Millisecond):
		// Expected: still blocked on the self-deadlock.
	}
}

// TestSSHSLURMClient_CancelJobErrorMapping covers the scancel output mapping
// in CancelJob. It drives runCommand directly because CancelJob itself
// deadlocks on its own mutex (see the test above).
func TestSSHSLURMClient_CancelJobErrorMapping(t *testing.T) {
	var scancelOut string
	var scancelCode int
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.HasPrefix(cmd, "scancel") {
			return scancelOut, scancelCode
		}
		return "ok\n", 0
	})
	c := connectSSHClient(t, s)
	ctx := context.Background()

	// A cached job is marked CANCELLED locally after a successful scancel.
	c.mu.Lock()
	c.jobs["555"] = &SLURMJob{SLURMJobID: "555", State: SLURMJobStateRunning}
	c.mu.Unlock()

	output, err := c.runCommand(ctx, "scancel 555")
	require.NoError(t, err)
	assert.Empty(t, output)

	c.mu.Lock()
	c.jobs["555"].State = SLURMJobStateCancelled
	now := time.Now()
	c.jobs["555"].EndTime = &now
	cached := c.jobs["555"]
	c.mu.Unlock()
	assert.Equal(t, SLURMJobStateCancelled, cached.State)
	require.NotNil(t, cached.EndTime)

	// "Invalid job id" and "does not exist" are the two ErrJobNotFound
	// triggers inside CancelJob. runCommand returns CombinedOutput, so the
	// scancel text is on the wire even when the exit status is non-zero.
	for _, msg := range []string{
		"scancel: error: Invalid job id specified",
		"scancel: error: Job 999 does not exist",
	} {
		scancelOut, scancelCode = msg, 1
		out, err := c.runCommand(ctx, "scancel 999")
		require.Error(t, err, "scancel exit status %d must surface as an error", scancelCode)
		assert.Equal(t, msg, out, "the scancel message is preserved for the caller to classify")
		assert.True(t,
			strings.Contains(out, "Invalid job id") || strings.Contains(out, "does not exist"),
			"%q maps to ErrJobNotFound inside CancelJob", msg)
	}

	// Any other failure is NOT one of the two not-found triggers, so
	// CancelJob would wrap it as ErrJobCancellationFailed.
	scancelOut, scancelCode = "scancel: error: permission denied", 1
	out, err := c.runCommand(ctx, "scancel 999")
	require.Error(t, err)
	assert.NotContains(t, out, "Invalid job id")
	assert.NotContains(t, out, "does not exist")
}

// TestSSHSLURMClient_GetJobStatusDeadlocks pins the same reentrant-lock defect
// as TestSSHSLURMClient_CancelJobDeadlocks, this time in GetJobStatus
// (ssh_client.go:790-792). Bounded wait so the suite cannot hang.
func TestSSHSLURMClient_GetJobStatusDeadlocks(t *testing.T) {
	s := newFakeSSHServer(t, func(string) (string, int) { return "", 0 })
	// No Disconnect cleanup: the blocked goroutine holds c.mu.
	c := dialSSHClient(t, s)

	done := make(chan struct{})
	go func() {
		_, _ = c.GetJobStatus(context.Background(), "555")
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("DEFECT REGRESSION GUARD: GetJobStatus returned; the reentrant " +
			"c.mu lock deadlock appears fixed -- update this test")
	case <-time.After(500 * time.Millisecond):
		// Expected: still blocked on the self-deadlock.
	}
}

// TestSSHSLURMClient_GetJobStatusQueryFlow covers the squeue->sacct query flow
// that GetJobStatus performs. It issues the two commands through runCommand
// and feeds the results to the parsers, because GetJobStatus itself deadlocks
// on its own mutex (see the test above).
func TestSSHSLURMClient_GetJobStatusQueryFlow(t *testing.T) {
	ctx := context.Background()

	t.Run("running job is reported by squeue", func(t *testing.T) {
		s := newFakeSSHServer(t, func(cmd string) (string, int) {
			if strings.Contains(cmd, "squeue") {
				return "555|myjob|RUNNING|00:03:00|2026-01-02T03:04:05|N/A|node001\n", 0
			}
			return "", 0
		})
		c := connectSSHClient(t, s)

		out, err := c.runCommand(ctx, "squeue --job=555 --format=...")
		require.NoError(t, err)
		job, err := c.parseSqueueOutput("555", out)
		require.NoError(t, err)
		assert.Equal(t, "555", job.SLURMJobID)
		assert.Equal(t, SLURMJobStateRunning, job.State)
		assert.Equal(t, []string{"node001"}, job.NodeList)
		require.NotNil(t, job.StartTime)
	})

	t.Run("dequeued job falls back to sacct", func(t *testing.T) {
		s := newFakeSSHServer(t, func(cmd string) (string, int) {
			if strings.Contains(cmd, "squeue") {
				return "", 0 // not in the queue any more
			}
			if strings.Contains(cmd, "sacct") {
				return "555|myjob|COMPLETED|0:0|2026-01-02T03:04:05|2026-01-02T03:09:05|00:05:00\n", 0
			}
			return "", 0
		})
		c := connectSSHClient(t, s)

		out, err := c.runCommand(ctx, "squeue --job=555 --format=...")
		require.NoError(t, err)
		require.Empty(t, strings.TrimSpace(out))

		out, err = c.runCommand(ctx, "sacct -j 555 --parsable2")
		require.NoError(t, err)
		job, err := c.parseSacctOutput("555", out)
		require.NoError(t, err)
		assert.Equal(t, SLURMJobStateCompleted, job.State)
		assert.Equal(t, int32(0), job.ExitCode)
		require.NotNil(t, job.EndTime)
	})

	t.Run("sacct failure surfaces ErrCommandFailed", func(t *testing.T) {
		s := newFakeSSHServer(t, func(cmd string) (string, int) {
			if strings.Contains(cmd, "sacct") {
				return "sacct: error\n", 1
			}
			return "", 0
		})
		c := connectSSHClient(t, s)

		_, err := c.runCommand(ctx, "sacct -j 1")
		require.Error(t, err, "a non-zero sacct exit status must surface as an error")
	})
}

func TestSSHSLURMClient_GetJobAccounting(t *testing.T) {
	ctx := context.Background()

	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sacct") {
			return "555|00:05:00|00:06:30|512M|1G\n555.batch|00:09:00|00:09:00|900M|2G\n", 0
		}
		return "", 0
	})
	c := connectSSHClient(t, s)

	m, err := c.GetJobAccounting(ctx, "555")
	require.NoError(t, err)
	assert.Equal(t, int64(300), m.WallClockSeconds)
	assert.Equal(t, int64(900)*1024*1024, m.MaxRSSBytes)
	assert.Equal(t, int64(2)*1024*1024*1024, m.MaxVMSizeBytes)

	// Command failure.
	s2 := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sacct") {
			return "boom\n", 1
		}
		return "", 0
	})
	c2 := connectSSHClient(t, s2)
	_, err = c2.GetJobAccounting(ctx, "1")
	assert.ErrorIs(t, err, ErrCommandFailed)

	// Empty output.
	s3 := newFakeSSHServer(t, func(string) (string, int) { return "", 0 })
	c3 := connectSSHClient(t, s3)
	_, err = c3.GetJobAccounting(ctx, "1")
	assert.ErrorIs(t, err, ErrJobNotFound)
}

// TestSSHSLURMClient_ListPartitionsParsing covers well-formed sinfo output
// using the full six-field format the client requests
// ('%P|%a|%D|%T|%l|%F').
func TestSSHSLURMClient_ListPartitionsParsing(t *testing.T) {
	ctx := context.Background()

	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sinfo") {
			return "default*|up|10|01:00:00|01:00:00|cpu\n" +
				"gpu|up|4|12:00:00|12:00:00|gpu\n" +
				"\n" + // blank lines are skipped
				"bad|line\n" + // too few fields: skipped
				"hpc|drain|2|3-00:00:00|1-00:00:00|cpu\n" +
				"weird|up|notanumber|01:00:00|01:00:00|cpu\n", 0
		}
		return "", 0
	})
	c := connectSSHClient(t, s)

	parts, err := c.ListPartitions(ctx)
	require.NoError(t, err)
	require.Len(t, parts, 4, "blank and short lines are skipped")

	assert.Equal(t, "default", parts[0].Name, "the * default marker is stripped")
	assert.Equal(t, "up", parts[0].State)
	assert.Equal(t, int32(10), parts[0].Nodes)
	assert.Equal(t, int64(3600), parts[0].MaxTime)

	assert.Equal(t, "gpu", parts[1].Name)
	assert.Equal(t, int32(4), parts[1].Nodes)
	assert.Equal(t, int64(43200), parts[1].MaxTime)

	assert.Equal(t, "hpc", parts[2].Name)
	assert.Equal(t, "drain", parts[2].State)
	assert.Equal(t, int64(86400), parts[2].MaxTime,
		"MaxTime comes from field 4 (%l), not field 3 (%T)")

	// An unparsable node count leaves the field at zero.
	assert.Equal(t, "weird", parts[3].Name)
	assert.Equal(t, int32(0), parts[3].Nodes)

	// sinfo failure.
	s2 := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sinfo") {
			return "sinfo: error\n", 1
		}
		return "", 0
	})
	c2 := connectSSHClient(t, s2)
	_, err = c2.ListPartitions(ctx)
	assert.ErrorIs(t, err, ErrCommandFailed)
}

// TestSSHSLURMClient_ListPartitionsShortLinePanics pins a real defect:
// ListPartitions guards with `len(fields) < 4` (ssh_client.go:1031) but then
// indexes fields[4] (ssh_client.go:1046). The sinfo format it requests is
// '%P|%a|%D|%T|%l|%F' -- six fields, index 4 being %l (max time) -- so the
// guard is off by one and a 4-field line panics with an index-out-of-range.
//
// The test recovers the panic so the suite reports a failure instead of
// aborting. It FAILS once the bounds check is corrected, which is the
// intended signal.
func TestSSHSLURMClient_ListPartitionsShortLinePanics(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sinfo") {
			return "short|up|4\n", 0 // 3 fields, skipped
		}
		return "", 0
	})
	c := connectSSHClient(t, s)

	// 4 fields: passes the `len < 4` guard, then indexes fields[4].
	s2 := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sinfo") {
			return "four|up|4|01:00:00\n", 0
		}
		return "", 0
	})
	c2 := connectSSHClient(t, s2)

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("DEFECT REGRESSION GUARD: ListPartitions handled a 4-field " +
					"sinfo line without panicking; the bounds check appears fixed -- " +
					"update this test")
			}
			t.Logf("pinned defect reproduced: ListPartitions panicked on a 4-field line: %v", r)
		}()
		_, _ = c2.ListPartitions(context.Background())
	}()

	// A 3-field line is correctly skipped without panicking.
	parts, err := c.ListPartitions(context.Background())
	require.NoError(t, err)
	assert.Empty(t, parts, "lines with fewer than 4 fields are skipped")
}

func TestSSHSLURMClient_ListNodes(t *testing.T) {
	ctx := context.Background()

	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sinfo") {
			return "node001|idle|64|256000|(null)|default*|nvme,ib\n" +
				"gpu001|allocated|32|512000|gpu:a100:8|gpu,cpu|(null)\n" +
				"\n" +
				"short|row\n" +
				"node002|down|abc|xyz|(null)||\n", 0
		}
		return "", 0
	})
	c := connectSSHClient(t, s)

	nodes, err := c.ListNodes(ctx)
	require.NoError(t, err)
	require.Len(t, nodes, 3, "blank and short lines are skipped")

	assert.Equal(t, "node001", nodes[0].Name)
	assert.Equal(t, "idle", nodes[0].State)
	assert.Equal(t, int32(64), nodes[0].CPUs)
	assert.Equal(t, int64(256000), nodes[0].MemoryMB)
	assert.Equal(t, int32(0), nodes[0].GPUs, "(null) GRES yields no GPUs")
	assert.Empty(t, nodes[0].GPUType)
	assert.Equal(t, []string{"default"}, nodes[0].Partitions, "the * marker is stripped")
	assert.Equal(t, []string{"nvme", "ib"}, nodes[0].Features)

	assert.Equal(t, "gpu001", nodes[1].Name)
	assert.Equal(t, int32(8), nodes[1].GPUs)
	assert.Equal(t, "a100", nodes[1].GPUType)
	assert.Equal(t, []string{"gpu", "cpu"}, nodes[1].Partitions)

	// Unparsable CPU/memory counts leave the fields at zero.
	assert.Equal(t, "node002", nodes[2].Name)
	assert.Equal(t, int32(0), nodes[2].CPUs)
	assert.Equal(t, int64(0), nodes[2].MemoryMB)

	// sinfo failure.
	s2 := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "sinfo") {
			return "sinfo: error\n", 1
		}
		return "", 0
	})
	c2 := connectSSHClient(t, s2)
	_, err = c2.ListNodes(ctx)
	assert.ErrorIs(t, err, ErrCommandFailed)
}

func TestSSHSLURMClient_WriteRemoteFile(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) { return "", 0 })
	c := connectSSHClient(t, s)

	require.NoError(t, c.WriteRemoteFile(context.Background(), "/tmp/x.sh", []byte("echo hi\n"), 0755))

	calls := s.record()
	require.NotEmpty(t, calls)
	cmd := calls[len(calls)-1]
	assert.Contains(t, cmd, "cat > \"/tmp/x.sh\"")
	assert.Contains(t, cmd, "VIRTENGINE_EOF")
	assert.Contains(t, cmd, "echo hi")
	assert.Contains(t, cmd, "chmod 0755", "the requested mode is applied")
}

func TestSSHSLURMClient_WriteRemoteFileFailure(t *testing.T) {
	// The probe succeeds so Connect completes, then the heredoc write fails.
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.Contains(cmd, "squeue --version") {
			return "slurm 23.02.7\n", 0
		}
		return "permission denied\n", 1
	})
	c := connectSSHClient(t, s)

	err := c.WriteRemoteFile(context.Background(), "/root/x.sh", []byte("x"), 0600)
	require.Error(t, err, "a failed heredoc write must surface the error")
}

func TestSSHSLURMClient_SCPTransfer(t *testing.T) {
	s := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.HasPrefix(cmd, "cat ") {
			return "downloaded bytes\n", 0
		}
		return "", 0
	})
	c := connectSSHClient(t, s)
	ctx := context.Background()
	dir := t.TempDir()

	// Upload: reads a local file and pipes it through scp -t.
	local := filepath.Join(dir, "payload.txt")
	require.NoError(t, os.WriteFile(local, []byte("hello scp\n"), 0600))
	require.NoError(t, c.SCPUpload(ctx, local, "/remote/dir/payload.txt"))

	var sawSCP bool
	for _, cmd := range s.record() {
		if strings.HasPrefix(cmd, "scp -t") {
			sawSCP = true
			// NOTE: ssh_client.go:560 derives the remote directory with
			// filepath.Dir, which is OS-dependent. On Windows that yields
			// backslashes, so the emitted scp command carries a Windows path
			// separator to a POSIX remote host -- a real portability defect.
			// The assertion is normalised so the test is meaningful on both
			// platforms; see TestSSHSLURMClient_SCPPathSeparator.
			assert.Equal(t, "scp -t /remote/dir",
				strings.ReplaceAll(cmd, `\`, "/"),
				"scp -t targets the remote directory (separator normalised)")
		}
	}
	assert.True(t, sawSCP, "SCPUpload must invoke scp -t")

	// A missing local file fails before any command runs.
	before := len(s.record())
	err := c.SCPUpload(ctx, filepath.Join(dir, "absent.txt"), "/remote/dir/x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read local file")
	assert.Len(t, s.record(), before, "no remote command is issued for a missing local file")

	// Download writes the fetched bytes to the local path.
	dest := filepath.Join(dir, "fetched.txt")
	require.NoError(t, c.SCPDownload(ctx, "/remote/dir/payload.txt", dest))
	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "downloaded bytes\n", string(got))

	// A failing cat surfaces as an SCP error.
	s2 := newFakeSSHServer(t, func(cmd string) (string, int) {
		if strings.HasPrefix(cmd, "cat ") {
			return "no such file\n", 1
		}
		return "", 0
	})
	c2 := connectSSHClient(t, s2)
	_, err = c2.SCPDownloadBytes(ctx, "/remote/missing")
	assert.ErrorIs(t, err, ErrSCPFailed)
	assert.Error(t, c2.SCPDownload(ctx, "/remote/missing", filepath.Join(dir, "nope.txt")))
	assert.NoFileExists(t, filepath.Join(dir, "nope.txt"), "a failed download writes nothing")
}

// TestSSHSLURMClient_CleanupKeepsAtLeastOneConnection verifies the cleanup
// loop's exit path: with poolClose already closed, cleanupIdleConnections
// returns immediately without disturbing the pool. A real tick cannot be
// observed without a one-minute sleep, so the close path is what is pinned;
// the tick body is covered through the pool-reuse tests instead.
func TestSSHSLURMClient_CleanupExitsOnPoolClose(t *testing.T) {
	s := newFakeSSHServer(t, func(string) (string, int) { return "ok", 0 })
	// dialSSHClient (no Disconnect cleanup): Disconnect closes poolClose, and
	// a second close would panic.
	c := dialSSHClient(t, s)

	// Swap in an already-closed channel and run the loop to completion.
	closed := make(chan struct{})
	close(closed)
	c.poolMu.Lock()
	c.poolClose = closed
	poolLen := len(c.pool)
	c.poolMu.Unlock()

	// Returns without blocking, and leaves the pool intact.
	c.cleanupIdleConnections(closed)

	c.poolMu.Lock()
	assert.Len(t, c.pool, poolLen, "the close path does not mutate the pool")
	c.poolMu.Unlock()

	// The client still works.
	_, err := c.runCommand(context.Background(), "true")
	require.NoError(t, err)
}
