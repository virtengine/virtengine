// Package slurm_adapter implements the SLURM orchestration adapter for VirtEngine.
//
// VE-2020: Real SLURM adapter implementation using SSH for remote SLURM execution
package slurm_adapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// ErrSSHConnection is returned when SSH connection fails
var ErrSSHConnection = errors.New("SSH connection failed")

// ErrCommandFailed is returned when a SLURM command fails
var ErrCommandFailed = errors.New("SLURM command failed")

// ErrSCPFailed is returned when SCP file transfer fails
var ErrSCPFailed = errors.New("SCP file transfer failed")

// ErrConnectionPoolExhausted is returned when connection pool is exhausted
var ErrConnectionPoolExhausted = errors.New("SSH connection pool exhausted")

// ErrHostKeyVerification is returned when host key verification fails
var ErrHostKeyVerification = errors.New("SSH host key verification failed")

// SSHConfig contains SSH connection configuration
type SSHConfig struct {
	// Host is the SSH host (SLURM login node)
	Host string `json:"host"`

	// Port is the SSH port (default: 22)
	Port int `json:"port"`

	// User is the SSH username
	User string `json:"user"`

	// PrivateKeyPath is the path to the private key file
	PrivateKeyPath string `json:"private_key_path"`

	// PrivateKey is the private key content (alternative to path)
	PrivateKey string `json:"-"` // Never log

	// Password is the SSH password (if not using key auth)
	Password string `json:"-"` // Never log

	// HostKeyCallback controls host key verification
	// "ignore" to skip verification (not recommended for production)
	// "known_hosts" to use known_hosts file verification
	HostKeyCallback string `json:"host_key_callback"`

	// KnownHostsPath is the path to known_hosts file (default: ~/.ssh/known_hosts)
	// Used when HostKeyCallback is set to "known_hosts"
	KnownHostsPath string `json:"known_hosts_path"`

	// Timeout is the SSH connection timeout
	Timeout time.Duration `json:"timeout"`

	// KeepAliveInterval is the interval for keepalive packets
	KeepAliveInterval time.Duration `json:"keepalive_interval"`

	// MaxRetries is the maximum number of reconnection attempts
	MaxRetries int `json:"max_retries"`

	// PoolSize is the number of connections in the pool (default: 5)
	PoolSize int `json:"pool_size"`

	// PoolIdleTimeout is how long idle connections are kept (default: 5m)
	PoolIdleTimeout time.Duration `json:"pool_idle_timeout"`
}

// DefaultSSHConfig returns default SSH configuration
func DefaultSSHConfig() SSHConfig {
	return SSHConfig{
		Port:              22,
		Timeout:           30 * time.Second,
		KeepAliveInterval: 30 * time.Second,
		MaxRetries:        3,
		HostKeyCallback:   "known_hosts", // Use known_hosts verification by default
		PoolSize:          5,
		PoolIdleTimeout:   5 * time.Minute,
	}
}

// pooledConnection represents a connection in the pool
type pooledConnection struct {
	client   *ssh.Client
	lastUsed time.Time
	inUse    bool
}

// SSHSLURMClient implements SLURMClient using SSH to execute SLURM commands
type SSHSLURMClient struct {
	config     SSHConfig
	sshConfig  *ssh.ClientConfig
	mu         sync.RWMutex
	connected  bool
	jobs       map[string]*SLURMJob // Local cache for job tracking
	clusterName string
	defaultPartition string

	// Connection pool
	pool      []*pooledConnection
	poolMu    sync.Mutex
	poolCond  *sync.Cond
	poolClose chan struct{}
}

// NewSSHSLURMClient creates a new SSH-based SLURM client
func NewSSHSLURMClient(sshConfig SSHConfig, clusterName, defaultPartition string) (*SSHSLURMClient, error) {
	// Build SSH client config
	clientConfig := &ssh.ClientConfig{
		User:    sshConfig.User,
		Timeout: sshConfig.Timeout,
	}

	// Configure authentication
	if sshConfig.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(sshConfig.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if sshConfig.PrivateKeyPath != "" {
		key, err := os.ReadFile(sshConfig.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if sshConfig.Password != "" {
		clientConfig.Auth = []ssh.AuthMethod{ssh.Password(sshConfig.Password)}
	} else {
		return nil, errors.New("no authentication method provided")
	}

	// Configure host key callback
	switch sshConfig.HostKeyCallback {
	case "ignore":
		clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	case "known_hosts":
		knownHostsPath := sshConfig.KnownHostsPath
		if knownHostsPath == "" {
			// Default to ~/.ssh/known_hosts
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %w", err)
			}
			knownHostsPath = filepath.Join(homeDir, ".ssh", "known_hosts")
		}

		// Check if known_hosts file exists
		if _, err := os.Stat(knownHostsPath); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: known_hosts file not found at %s", ErrHostKeyVerification, knownHostsPath)
			}
			return nil, fmt.Errorf("failed to access known_hosts file: %w", err)
		}

		// Create host key callback from known_hosts file
		hostKeyCallback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create host key callback: %w", err)
		}
		clientConfig.HostKeyCallback = hostKeyCallback
	default:
		// Default to known_hosts for security
		knownHostsPath := sshConfig.KnownHostsPath
		if knownHostsPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %w", err)
			}
			knownHostsPath = filepath.Join(homeDir, ".ssh", "known_hosts")
		}

		// If known_hosts exists, use it; otherwise fall back to insecure
		if _, err := os.Stat(knownHostsPath); err == nil {
			hostKeyCallback, err := knownhosts.New(knownHostsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create host key callback: %w", err)
			}
			clientConfig.HostKeyCallback = hostKeyCallback
		} else {
			// Fall back to insecure if no known_hosts file exists
			clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		}
	}

	// Set defaults for pool
	poolSize := sshConfig.PoolSize
	if poolSize <= 0 {
		poolSize = 5
	}

	client := &SSHSLURMClient{
		config:          sshConfig,
		sshConfig:       clientConfig,
		jobs:            make(map[string]*SLURMJob),
		clusterName:     clusterName,
		defaultPartition: defaultPartition,
		pool:            make([]*pooledConnection, 0, poolSize),
		poolClose:       make(chan struct{}),
	}
	client.poolCond = sync.NewCond(&client.poolMu)

	return client, nil
}

// Connect establishes SSH connection pool to the SLURM login node
func (c *SSHSLURMClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Initialize connection pool
	// Note: pool size is configured but actual pooling is handled by dialWithRetry
	_ = c.config.PoolSize // poolSize used in future connection pool expansion

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// Create initial connection to verify connectivity
	client, err := c.dialWithRetry(ctx, addr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSSHConnection, err)
	}

	// Test connection with a simple SLURM command
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	_, err = session.CombinedOutput("squeue --version")
	session.Close()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to verify SLURM: %w", err)
	}

	// Add first connection to pool
	c.poolMu.Lock()
	c.pool = append(c.pool, &pooledConnection{
		client:   client,
		lastUsed: time.Now(),
		inUse:    false,
	})
	c.poolMu.Unlock()

	c.connected = true

	// Start idle connection cleanup goroutine
	go c.cleanupIdleConnections()

	return nil
}

// dialWithRetry attempts to dial with retries
func (c *SSHSLURMClient) dialWithRetry(ctx context.Context, addr string) (*ssh.Client, error) {
	var client *ssh.Client
	var err error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		client, err = ssh.Dial("tcp", addr, c.sshConfig)
		if err == nil {
			return client, nil
		}

		if attempt < c.config.MaxRetries {
			time.Sleep(time.Second * time.Duration(attempt+1))
		}
	}

	return nil, err
}

// acquireConnection gets a connection from the pool or creates a new one
func (c *SSHSLURMClient) acquireConnection(ctx context.Context) (*ssh.Client, error) {
	c.poolMu.Lock()
	defer c.poolMu.Unlock()

	poolSize := c.config.PoolSize
	if poolSize <= 0 {
		poolSize = 5
	}

	// Try to find an available connection
	for _, pc := range c.pool {
		if !pc.inUse && pc.client != nil {
			pc.inUse = true
			pc.lastUsed = time.Now()
			return pc.client, nil
		}
	}

	// Create new connection if pool not full
	if len(c.pool) < poolSize {
		addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
		client, err := c.dialWithRetry(ctx, addr)
		if err != nil {
			return nil, err
		}
		c.pool = append(c.pool, &pooledConnection{
			client:   client,
			lastUsed: time.Now(),
			inUse:    true,
		})
		return client, nil
	}

	// Wait for a connection to become available
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		for _, pc := range c.pool {
			if !pc.inUse && pc.client != nil {
				pc.inUse = true
				pc.lastUsed = time.Now()
				return pc.client, nil
			}
		}

		// Wait with timeout
		c.poolCond.Wait()
	}
}

// releaseConnection returns a connection to the pool
func (c *SSHSLURMClient) releaseConnection(client *ssh.Client) {
	c.poolMu.Lock()
	defer c.poolMu.Unlock()

	for _, pc := range c.pool {
		if pc.client == client {
			pc.inUse = false
			pc.lastUsed = time.Now()
			c.poolCond.Signal()
			return
		}
	}
}

// cleanupIdleConnections removes idle connections from the pool
func (c *SSHSLURMClient) cleanupIdleConnections() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	idleTimeout := c.config.PoolIdleTimeout
	if idleTimeout <= 0 {
		idleTimeout = 5 * time.Minute
	}

	for {
		select {
		case <-c.poolClose:
			return
		case <-ticker.C:
			c.poolMu.Lock()
			now := time.Now()
			keepPool := make([]*pooledConnection, 0, len(c.pool))
			
			for _, pc := range c.pool {
				// Keep at least one connection
				if len(keepPool) == 0 || pc.inUse || now.Sub(pc.lastUsed) < idleTimeout {
					keepPool = append(keepPool, pc)
				} else {
					// Close idle connection
					if pc.client != nil {
						pc.client.Close()
					}
				}
			}
			c.pool = keepPool
			c.poolMu.Unlock()
		}
	}
}

// Disconnect closes all SSH connections in the pool
func (c *SSHSLURMClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Signal cleanup goroutine to stop
	close(c.poolClose)

	// Close all connections in pool
	c.poolMu.Lock()
	for _, pc := range c.pool {
		if pc.client != nil {
			pc.client.Close()
		}
	}
	c.pool = nil
	c.poolMu.Unlock()

	c.connected = false
	c.poolClose = make(chan struct{}) // Reset for potential reconnect
	
	return nil
}

// IsConnected checks if connected to SLURM
func (c *SSHSLURMClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// runCommand executes a command via SSH and returns the output
func (c *SSHSLURMClient) runCommand(ctx context.Context, cmd string) (string, error) {
	if !c.IsConnected() {
		return "", ErrSLURMNotConnected
	}

	client, err := c.acquireConnection(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer c.releaseConnection(client)

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Handle context cancellation
	done := make(chan struct{})
	verrors.SafeGo("", func() {
		defer func() {}() // WG Done if needed
		select {
		case <-ctx.Done():
			session.Signal(ssh.SIGTERM)
			session.Close()
		case <-done:
		}
	})

	output, err := session.CombinedOutput(cmd)
	close(done)

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	return string(output), err
}

// runCommandWithTimeout executes a command with a specific timeout
func (c *SSHSLURMClient) runCommandWithTimeout(ctx context.Context, cmd string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.runCommand(ctx, cmd)
}

// SCPUpload uploads a file to the remote host via SCP
func (c *SSHSLURMClient) SCPUpload(ctx context.Context, localPath, remotePath string) error {
	if !c.IsConnected() {
		return ErrSLURMNotConnected
	}

	// Read local file
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	return c.SCPUploadBytes(ctx, content, remotePath, 0644)
}

// SCPUploadBytes uploads bytes to a remote file via SCP
func (c *SSHSLURMClient) SCPUploadBytes(ctx context.Context, content []byte, remotePath string, mode os.FileMode) error {
	if !c.IsConnected() {
		return ErrSLURMNotConnected
	}

	client, err := c.acquireConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer c.releaseConnection(client)

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Get just the filename for SCP protocol
	filename := filepath.Base(remotePath)
	dir := filepath.Dir(remotePath)

	// Prepare content with SCP protocol
	verrors.SafeGo("", func() {
		defer func() {}() // WG Done if needed
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", mode, len(content), filename)
		w.Write(content)
		fmt.Fprint(w, "\x00")
	})

	// Run scp command
	cmd := fmt.Sprintf("scp -t %s", dir)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("%w: %v", ErrSCPFailed, err)
	}

	return nil
}

// SCPDownload downloads a file from the remote host via SCP
func (c *SSHSLURMClient) SCPDownload(ctx context.Context, remotePath, localPath string) error {
	if !c.IsConnected() {
		return ErrSLURMNotConnected
	}

	content, err := c.SCPDownloadBytes(ctx, remotePath)
	if err != nil {
		return err
	}

	return os.WriteFile(localPath, content, 0644)
}

// SCPDownloadBytes downloads a file from the remote host as bytes
func (c *SSHSLURMClient) SCPDownloadBytes(ctx context.Context, remotePath string) ([]byte, error) {
	if !c.IsConnected() {
		return nil, ErrSLURMNotConnected
	}

	client, err := c.acquireConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer c.releaseConnection(client)

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Use cat for simple file download
	cmd := fmt.Sprintf("cat %q", remotePath)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSCPFailed, err)
	}

	return output, nil
}

// WriteRemoteFile writes content to a remote file using heredoc
func (c *SSHSLURMClient) WriteRemoteFile(ctx context.Context, remotePath string, content []byte, mode os.FileMode) error {
	if !c.IsConnected() {
		return ErrSLURMNotConnected
	}

	// Use heredoc approach which is more reliable than SCP
	cmd := fmt.Sprintf(`cat > %q << 'VIRTENGINE_EOF'
%s
VIRTENGINE_EOF
chmod %04o %q`, remotePath, string(content), mode, remotePath)

	_, err := c.runCommand(ctx, cmd)
	return err
}

// SubmitJob submits a job to SLURM using sbatch
func (c *SSHSLURMClient) SubmitJob(ctx context.Context, spec *SLURMJobSpec) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", ErrSLURMNotConnected
	}

	if err := spec.Validate(); err != nil {
		return "", err
	}

	// Generate SLURM batch script using BatchScriptBuilder
	builder := FromJobSpec(spec)
	if spec.Partition == "" && c.defaultPartition != "" {
		builder.SetPartition(c.defaultPartition)
	}
	builder.AddHeaderComment(fmt.Sprintf("Generated by VirtEngine SLURM Adapter"))
	builder.AddHeaderComment(fmt.Sprintf("Cluster: %s", c.clusterName))
	script := builder.Build()

	// Create temporary script file and submit
	// Using heredoc approach to avoid file creation
	submitCmd := fmt.Sprintf(`cat << 'SLURM_SCRIPT_EOF' | sbatch --parsable
%s
SLURM_SCRIPT_EOF`, script)

	c.mu.Unlock() // Unlock before running command
	output, err := c.runCommand(ctx, submitCmd)
	c.mu.Lock() // Re-lock after command
	
	if err != nil {
		return "", fmt.Errorf("%w: %v - %s", ErrJobSubmissionFailed, err, output)
	}

	// Parse job ID from sbatch output (--parsable returns just the job ID)
	jobID := strings.TrimSpace(output)
	
	// Handle cluster-specific output: "job_id;cluster_name"
	if strings.Contains(jobID, ";") {
		parts := strings.Split(jobID, ";")
		jobID = parts[0]
	}

	if jobID == "" || !isNumeric(jobID) {
		return "", fmt.Errorf("%w: unexpected sbatch output: %s", ErrJobSubmissionFailed, output)
	}

	// Store job in local cache
	now := time.Now()
	job := &SLURMJob{
		SLURMJobID: jobID,
		Spec:       spec,
		State:      SLURMJobStatePending,
		SubmitTime: now,
	}
	c.jobs[jobID] = job

	return jobID, nil
}

// SubmitJobFromScript submits a pre-built batch script to SLURM
func (c *SSHSLURMClient) SubmitJobFromScript(ctx context.Context, script string) (string, error) {
	if !c.IsConnected() {
		return "", ErrSLURMNotConnected
	}

	submitCmd := fmt.Sprintf(`cat << 'SLURM_SCRIPT_EOF' | sbatch --parsable
%s
SLURM_SCRIPT_EOF`, script)

	output, err := c.runCommand(ctx, submitCmd)
	if err != nil {
		return "", fmt.Errorf("%w: %v - %s", ErrJobSubmissionFailed, err, output)
	}

	jobID := strings.TrimSpace(output)
	if strings.Contains(jobID, ";") {
		parts := strings.Split(jobID, ";")
		jobID = parts[0]
	}

	if jobID == "" || !isNumeric(jobID) {
		return "", fmt.Errorf("%w: unexpected sbatch output: %s", ErrJobSubmissionFailed, output)
	}

	return jobID, nil
}

// SubmitJobFromFile submits a job script file that exists on the remote system
func (c *SSHSLURMClient) SubmitJobFromFile(ctx context.Context, scriptPath string) (string, error) {
	if !c.IsConnected() {
		return "", ErrSLURMNotConnected
	}

	cmd := fmt.Sprintf("sbatch --parsable %q", scriptPath)
	output, err := c.runCommand(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("%w: %v - %s", ErrJobSubmissionFailed, err, output)
	}

	jobID := strings.TrimSpace(output)
	if strings.Contains(jobID, ";") {
		parts := strings.Split(jobID, ";")
		jobID = parts[0]
	}

	if jobID == "" || !isNumeric(jobID) {
		return "", fmt.Errorf("%w: unexpected sbatch output: %s", ErrJobSubmissionFailed, output)
	}

	return jobID, nil
}

// CancelJob cancels a SLURM job using scancel
func (c *SSHSLURMClient) CancelJob(ctx context.Context, slurmJobID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrSLURMNotConnected
	}

	cmd := fmt.Sprintf("scancel %s", slurmJobID)
	output, err := c.runCommand(ctx, cmd)
	if err != nil {
		// Check if job doesn't exist (already finished)
		if strings.Contains(output, "Invalid job id") || strings.Contains(output, "does not exist") {
			return ErrJobNotFound
		}
		return fmt.Errorf("%w: %v - %s", ErrJobCancellationFailed, err, output)
	}

	// Update local cache
	if job, exists := c.jobs[slurmJobID]; exists {
		job.State = SLURMJobStateCancelled
		now := time.Now()
		job.EndTime = &now
	}

	return nil
}

// GetJobStatus gets job status using squeue (for running jobs) or sacct (for completed)
func (c *SSHSLURMClient) GetJobStatus(ctx context.Context, slurmJobID string) (*SLURMJob, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, ErrSLURMNotConnected
	}

	// First try squeue for running/pending jobs
	cmd := fmt.Sprintf("squeue --job=%s --format='%%i|%%j|%%T|%%M|%%S|%%e|%%N' --noheader", slurmJobID)
	output, err := c.runCommand(ctx, cmd)
	
	if err == nil && strings.TrimSpace(output) != "" {
		job, err := c.parseSqueueOutput(slurmJobID, output)
		if err == nil {
			c.jobs[slurmJobID] = job
			return job, nil
		}
	}

	// If not in squeue, try sacct for completed jobs
	cmd = fmt.Sprintf("sacct -j %s --format=JobID,JobName,State,ExitCode,Start,End,Elapsed --parsable2 --noheader", slurmJobID)
	output, err = c.runCommand(ctx, cmd)
	
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}

	if strings.TrimSpace(output) == "" {
		return nil, ErrJobNotFound
	}

	job, err := c.parseSacctOutput(slurmJobID, output)
	if err != nil {
		return nil, err
	}

	c.jobs[slurmJobID] = job
	return job, nil
}

// parseSqueueOutput parses squeue output into a SLURMJob
func (c *SSHSLURMClient) parseSqueueOutput(jobID, output string) (*SLURMJob, error) {
	// Format: %i|%j|%T|%M|%S|%e|%N
	// JobID|Name|State|TimeUsed|StartTime|EndTime|NodeList
	line := strings.TrimSpace(output)
	fields := strings.Split(line, "|")
	
	if len(fields) < 7 {
		return nil, fmt.Errorf("unexpected squeue output format: %s", output)
	}

	job := &SLURMJob{
		SLURMJobID: fields[0],
		State:      mapSLURMState(fields[2]),
		NodeList:   parseNodeList(fields[6]),
	}

	// Parse start time if available
	if fields[4] != "N/A" && fields[4] != "" {
		if t, err := time.Parse("2006-01-02T15:04:05", fields[4]); err == nil {
			job.StartTime = &t
		}
	}

	// Merge with cached spec if available
	if cached, exists := c.jobs[jobID]; exists && cached.Spec != nil {
		job.Spec = cached.Spec
		job.VirtEngineJobID = cached.VirtEngineJobID
		job.SubmitTime = cached.SubmitTime
	}

	return job, nil
}

// parseSacctOutput parses sacct output into a SLURMJob
func (c *SSHSLURMClient) parseSacctOutput(jobID, output string) (*SLURMJob, error) {
	// Format: JobID|JobName|State|ExitCode|Start|End|Elapsed
	scanner := bufio.NewScanner(strings.NewReader(output))
	var mainJobLine string
	
	for scanner.Scan() {
		line := scanner.Text()
		// Skip step entries (e.g., "123.0", "123.batch"), only use main job entry
		if strings.Contains(line, jobID) && !strings.Contains(line, ".") {
			mainJobLine = line
			break
		}
	}

	if mainJobLine == "" {
		// Fall back to first line if no main job line found
		scanner = bufio.NewScanner(strings.NewReader(output))
		if scanner.Scan() {
			mainJobLine = scanner.Text()
		}
	}

	if mainJobLine == "" {
		return nil, ErrJobNotFound
	}

	fields := strings.Split(mainJobLine, "|")
	if len(fields) < 7 {
		return nil, fmt.Errorf("unexpected sacct output format: %s", output)
	}

	job := &SLURMJob{
		SLURMJobID: fields[0],
		State:      mapSLURMState(fields[2]),
	}

	// Parse exit code
	if exitParts := strings.Split(fields[3], ":"); len(exitParts) > 0 {
		if code, err := strconv.ParseInt(exitParts[0], 10, 32); err == nil {
			job.ExitCode = int32(code)
		}
	}

	// Parse start time
	if fields[4] != "Unknown" && fields[4] != "" {
		if t, err := time.Parse("2006-01-02T15:04:05", fields[4]); err == nil {
			job.StartTime = &t
		}
	}

	// Parse end time
	if fields[5] != "Unknown" && fields[5] != "" {
		if t, err := time.Parse("2006-01-02T15:04:05", fields[5]); err == nil {
			job.EndTime = &t
		}
	}

	// Merge with cached spec if available
	if cached, exists := c.jobs[jobID]; exists {
		job.Spec = cached.Spec
		job.VirtEngineJobID = cached.VirtEngineJobID
		job.SubmitTime = cached.SubmitTime
	}

	return job, nil
}

// GetJobAccounting gets detailed job accounting using sacct
func (c *SSHSLURMClient) GetJobAccounting(ctx context.Context, slurmJobID string) (*SLURMUsageMetrics, error) {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return nil, ErrSLURMNotConnected
	}

	// Get accounting data with more fields
	cmd := fmt.Sprintf("sacct -j %s --format=JobID,Elapsed,TotalCPU,MaxRSS,MaxVMSize --parsable2 --noheader", slurmJobID)
	output, err := c.runCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}

	if strings.TrimSpace(output) == "" {
		return nil, ErrJobNotFound
	}

	return c.parseSacctMetrics(output)
}

// parseSacctMetrics parses sacct metrics output
func (c *SSHSLURMClient) parseSacctMetrics(output string) (*SLURMUsageMetrics, error) {
	// Find the main job entry (not steps)
	scanner := bufio.NewScanner(strings.NewReader(output))
	var metrics *SLURMUsageMetrics
	
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "|")
		if len(fields) < 5 {
			continue
		}

		// Skip step entries for aggregated metrics (use .batch or main)
		m := &SLURMUsageMetrics{}

		// Parse elapsed time (format: [DD-]HH:MM:SS or just seconds)
		m.WallClockSeconds = parseDuration(fields[1])

		// Parse total CPU time (format: [DD-]HH:MM:SS.ms)
		m.CPUTimeSeconds = parseDuration(fields[2])

		// Parse MaxRSS (format: NNK or NNM or NNG)
		m.MaxRSSBytes = parseMemory(fields[3])

		// Parse MaxVMSize
		m.MaxVMSizeBytes = parseMemory(fields[4])

		// Keep the largest values across all entries
		if metrics == nil {
			metrics = m
		} else {
			if m.MaxRSSBytes > metrics.MaxRSSBytes {
				metrics.MaxRSSBytes = m.MaxRSSBytes
			}
			if m.MaxVMSizeBytes > metrics.MaxVMSizeBytes {
				metrics.MaxVMSizeBytes = m.MaxVMSizeBytes
			}
		}
	}

	if metrics == nil {
		return nil, ErrJobNotFound
	}

	return metrics, nil
}

// ListPartitions lists available SLURM partitions using sinfo
func (c *SSHSLURMClient) ListPartitions(ctx context.Context) ([]PartitionInfo, error) {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return nil, ErrSLURMNotConnected
	}

	cmd := "sinfo --format='%P|%a|%D|%T|%l|%F' --noheader"
	output, err := c.runCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}

	var partitions []PartitionInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 4 {
			continue
		}

		name := strings.TrimSuffix(fields[0], "*") // Remove default marker
		partition := PartitionInfo{
			Name:  name,
			State: fields[1],
		}

		if nodes, err := strconv.ParseInt(fields[2], 10, 32); err == nil {
			partition.Nodes = int32(nodes)
		}

		// Parse time limit
		partition.MaxTime = parseDuration(fields[4])

		partitions = append(partitions, partition)
	}

	return partitions, nil
}

// ListNodes lists nodes in the SLURM cluster using sinfo
func (c *SSHSLURMClient) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return nil, ErrSLURMNotConnected
	}

	cmd := "sinfo --Node --format='%N|%T|%c|%m|%G|%P|%f' --noheader"
	output, err := c.runCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}

	var nodes []NodeInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 5 {
			continue
		}

		node := NodeInfo{
			Name:  fields[0],
			State: fields[1],
		}

		if cpus, err := strconv.ParseInt(fields[2], 10, 32); err == nil {
			node.CPUs = int32(cpus)
		}

		if mem, err := strconv.ParseInt(fields[3], 10, 64); err == nil {
			node.MemoryMB = mem
		}

		// Parse GPU info (format: gpu:type:count or (null))
		if fields[4] != "(null)" && fields[4] != "" {
			node.GPUs, node.GPUType = parseGRES(fields[4])
		}

		// Parse partitions
		if len(fields) > 5 {
			node.Partitions = strings.Split(strings.TrimSuffix(fields[5], "*"), ",")
		}

		// Parse features
		if len(fields) > 6 && fields[6] != "(null)" {
			node.Features = strings.Split(fields[6], ",")
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// Helper functions

func mapSLURMState(state string) SLURMJobState {
	state = strings.ToUpper(strings.TrimSpace(state))
	switch state {
	case "PENDING", "PD":
		return SLURMJobStatePending
	case "RUNNING", "R":
		return SLURMJobStateRunning
	case "COMPLETED", "CD":
		return SLURMJobStateCompleted
	case "FAILED", "F":
		return SLURMJobStateFailed
	case "CANCELLED", "CA":
		return SLURMJobStateCancelled
	case "TIMEOUT", "TO":
		return SLURMJobStateTimeout
	case "SUSPENDED", "S":
		return SLURMJobStateSuspended
	default:
		return SLURMJobState(state)
	}
}

func parseNodeList(nodeSpec string) []string {
	// Handle compressed node lists like "node[001-004]"
	nodeSpec = strings.TrimSpace(nodeSpec)
	if nodeSpec == "" || nodeSpec == "(null)" {
		return nil
	}

	// Simple case: comma-separated list
	if !strings.Contains(nodeSpec, "[") {
		return strings.Split(nodeSpec, ",")
	}

	// TODO: Implement range expansion for node[001-004] format
	return []string{nodeSpec}
}

func parseDuration(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "UNLIMITED" || s == "INVALID" {
		return 0
	}

	// Format: [DD-]HH:MM:SS[.ms] or MM:SS
	var days, hours, minutes, seconds int64

	// Check for days prefix
	if idx := strings.Index(s, "-"); idx != -1 {
		if d, err := strconv.ParseInt(s[:idx], 10, 64); err == nil {
			days = d
		}
		s = s[idx+1:]
	}

	// Remove milliseconds suffix
	if idx := strings.Index(s, "."); idx != -1 {
		s = s[:idx]
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 3:
		hours, _ = strconv.ParseInt(parts[0], 10, 64)
		minutes, _ = strconv.ParseInt(parts[1], 10, 64)
		seconds, _ = strconv.ParseInt(parts[2], 10, 64)
	case 2:
		minutes, _ = strconv.ParseInt(parts[0], 10, 64)
		seconds, _ = strconv.ParseInt(parts[1], 10, 64)
	case 1:
		seconds, _ = strconv.ParseInt(parts[0], 10, 64)
	}

	return days*86400 + hours*3600 + minutes*60 + seconds
}

func parseMemory(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Handle suffixes: K, M, G, T
	var multiplier int64 = 1
	if len(s) > 0 {
		suffix := s[len(s)-1:]
		switch strings.ToUpper(suffix) {
		case "K":
			multiplier = 1024
			s = s[:len(s)-1]
		case "M":
			multiplier = 1024 * 1024
			s = s[:len(s)-1]
		case "G":
			multiplier = 1024 * 1024 * 1024
			s = s[:len(s)-1]
		case "T":
			multiplier = 1024 * 1024 * 1024 * 1024
			s = s[:len(s)-1]
		}
	}

	val, _ := strconv.ParseInt(s, 10, 64)
	return val * multiplier
}

func parseGRES(gres string) (int32, string) {
	// Format: gpu:type:count or gpu:count
	parts := strings.Split(gres, ":")
	if len(parts) < 2 {
		return 0, ""
	}

	if parts[0] != "gpu" {
		return 0, ""
	}

	if len(parts) == 2 {
		count, _ := strconv.ParseInt(parts[1], 10, 32)
		return int32(count), ""
	}

	if len(parts) >= 3 {
		count, _ := strconv.ParseInt(parts[2], 10, 32)
		return int32(count), parts[1]
	}

	return 0, ""
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// Compile-time check that SSHSLURMClient implements SLURMClient
var _ SLURMClient = (*SSHSLURMClient)(nil)

// Silence unused import warnings (used in SCP operations and JSON parsing)
var _ = regexp.MustCompile
var _ io.Reader = nil
var _ = json.Marshal
var _ = bytes.NewBuffer
