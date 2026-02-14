// Package moab_adapter implements the MOAB workload manager adapter for VirtEngine.
//
// VE-917: MOAB workload manager using Waldur
// HPC-ADAPTER-001: Production MOAB RPC client implementation
package moab_adapter

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// ProductionMOABClient implements the MOABClient interface for real MOAB clusters.
// It uses SSH to execute MOAB commands (msub, mjobctl, checkjob, mdiag) on the
// MOAB server and parses the XML/text output.
type ProductionMOABClient struct {
	config     MOABConfig
	mu         sync.RWMutex
	pool       *sshConnectionPool
	connected  bool
	authMethod ssh.AuthMethod
}

// sshConnectionPool manages a pool of SSH connections
type sshConnectionPool struct {
	mu              sync.Mutex
	config          MOABConfig
	authMethod      ssh.AuthMethod
	hostKeyCallback ssh.HostKeyCallback
	connections     chan *sshConnection
	maxSize         int
	idleTimeout     time.Duration
	closed          bool
}

// sshConnection represents a pooled SSH connection
type sshConnection struct {
	client    *ssh.Client
	createdAt time.Time
	lastUsed  time.Time
}

// NewProductionMOABClient creates a new production MOAB client
func NewProductionMOABClient(config MOABConfig) (*ProductionMOABClient, error) {
	client := &ProductionMOABClient{
		config: config,
	}

	// Setup authentication method
	authMethod, err := client.setupAuth()
	if err != nil {
		return nil, fmt.Errorf("failed to setup authentication: %w", err)
	}
	client.authMethod = authMethod

	return client, nil
}

// setupAuth configures the SSH authentication method
func (c *ProductionMOABClient) setupAuth() (ssh.AuthMethod, error) {
	switch c.config.AuthMethod {
	case "password":
		if c.config.Password == "" {
			return nil, errors.New("password authentication requires password")
		}
		return ssh.Password(c.config.Password), nil

	case "key":
		keyPath := os.Getenv("MOAB_SSH_KEY")
		if keyPath == "" {
			keyPath = os.ExpandEnv("$HOME/.ssh/id_rsa")
		}
		keyPath = filepath.Clean(keyPath)
		// #nosec G304,G703 -- key path is configured via environment and cleaned.
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// Try with passphrase
			passphrase := os.Getenv("MOAB_SSH_PASSPHRASE")
			if passphrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
				if err != nil {
					return nil, fmt.Errorf("failed to parse SSH key with passphrase: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to parse SSH key: %w", err)
			}
		}
		return ssh.PublicKeys(signer), nil

	case "kerberos":
		// Kerberos uses GSSAPI-with-MIC authentication
		// This requires the user to have a valid Kerberos ticket (kinit)
		return nil, errors.New("kerberos authentication not yet implemented - use password or key")

	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", c.config.AuthMethod)
	}
}

// setupHostKeyCallback creates an SSH host key callback based on configuration.
// SECURITY: This is critical for MITM protection. The callback mode determines
// how the server's host key is verified during SSH handshake.
func (c *ProductionMOABClient) setupHostKeyCallback() (ssh.HostKeyCallback, error) {
	switch c.config.SSHHostKeyCallback {
	case SSHHostKeyInsecure:
		// SECURITY WARNING: This disables host key verification and is vulnerable to MITM attacks.
		// Only use for testing or when other security measures (e.g., VPN, private network) are in place.
		// #nosec G106 -- InsecureIgnoreHostKey intentional for SSHHostKeyInsecure mode.
		return ssh.InsecureIgnoreHostKey(), nil

	case SSHHostKeyPinned:
		// Use a pinned public key for verification
		if c.config.SSHHostPublicKey == "" {
			return nil, errors.New("SSH host key verification requires SSHHostPublicKey when using 'pinned' mode")
		}
		return createPinnedKeyCallback(c.config.SSHHostPublicKey)

	case SSHHostKeyKnownHosts, "":
		// Default: Use known_hosts file
		knownHostsPath := c.config.SSHKnownHostsPath
		if knownHostsPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory for known_hosts: %w", err)
			}
			knownHostsPath = filepath.Join(homeDir, ".ssh", "known_hosts")
		}

		// Check if known_hosts file exists
		if _, err := os.Stat(knownHostsPath); err != nil {
			return nil, fmt.Errorf("SSH host key verification requires known_hosts file at %s (set SSHHostKeyCallback to 'insecure' to disable, NOT RECOMMENDED): %w", knownHostsPath, err)
		}

		hostKeyCallback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create host key callback from known_hosts: %w", err)
		}
		return hostKeyCallback, nil

	default:
		return nil, fmt.Errorf("unsupported SSH host key callback mode: %s (valid options: known_hosts, pinned, insecure)", c.config.SSHHostKeyCallback)
	}
}

// createPinnedKeyCallback creates a host key callback that verifies against a pinned public key.
// The key should be in the standard SSH public key format: "ssh-rsa AAAA..." or "ssh-ed25519 AAAA..."
func createPinnedKeyCallback(pinnedKey string) (ssh.HostKeyCallback, error) {
	// Parse the pinned key - it should be in authorized_keys format
	parts := strings.Fields(pinnedKey)
	if len(parts) < 2 {
		return nil, errors.New("invalid pinned SSH host key format: expected 'key-type base64-key [comment]'")
	}

	keyData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode pinned SSH host key: %w", err)
	}

	publicKey, err := ssh.ParsePublicKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pinned SSH host key: %w", err)
	}

	// Return a callback that verifies the server's key matches the pinned key
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if key.Type() != publicKey.Type() {
			return fmt.Errorf("SSH host key type mismatch: expected %s, got %s", publicKey.Type(), key.Type())
		}
		if !bytes.Equal(key.Marshal(), publicKey.Marshal()) {
			return fmt.Errorf("SSH host key mismatch for %s: possible MITM attack", hostname)
		}
		return nil
	}, nil
}

// Connect connects to the MOAB server
func (c *ProductionMOABClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Setup host key callback for MITM protection
	hostKeyCallback, err := c.setupHostKeyCallback()
	if err != nil {
		return fmt.Errorf("failed to setup SSH host key verification: %w", err)
	}

	// Create connection pool
	poolSize := 5
	if c.config.MaxRetries > 0 {
		poolSize = c.config.MaxRetries + 2
	}

	pool := &sshConnectionPool{
		config:          c.config,
		authMethod:      c.authMethod,
		hostKeyCallback: hostKeyCallback,
		connections:     make(chan *sshConnection, poolSize),
		maxSize:         poolSize,
		idleTimeout:     5 * time.Minute,
	}

	// Create initial connection to verify connectivity
	conn, err := pool.createConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to MOAB server: %w", err)
	}

	// Return connection to pool
	pool.connections <- conn
	c.pool = pool
	c.connected = true

	return nil
}

// Disconnect disconnects from MOAB
func (c *ProductionMOABClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.pool == nil {
		return nil
	}

	c.pool.close()
	c.connected = false
	c.pool = nil

	return nil
}

// IsConnected checks if connected
func (c *ProductionMOABClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SubmitJob submits a job using msub
func (c *ProductionMOABClient) SubmitJob(ctx context.Context, spec *MOABJobSpec) (string, error) {
	if !c.IsConnected() {
		return "", ErrMOABNotConnected
	}

	// Generate msub script
	script := spec.ToMsubScript()

	// Execute msub with script via stdin
	output, err := c.executeWithRetry(ctx, "msub", script)
	if err != nil {
		return "", fmt.Errorf("msub failed: %w", err)
	}

	// Parse job ID from output (format: "moab.12345" or just "12345")
	jobID := strings.TrimSpace(output)
	if jobID == "" {
		return "", errors.New("msub returned empty job ID")
	}

	// Validate job ID format
	if !isValidJobID(jobID) {
		return "", fmt.Errorf("invalid job ID from msub: %s", jobID)
	}

	return jobID, nil
}

// CancelJob cancels a job using mjobctl -c
func (c *ProductionMOABClient) CancelJob(ctx context.Context, moabJobID string) error {
	if !c.IsConnected() {
		return ErrMOABNotConnected
	}

	_, err := c.executeWithRetry(ctx, fmt.Sprintf("mjobctl -c %s", moabJobID), "")
	if err != nil {
		if strings.Contains(err.Error(), "job not found") ||
			strings.Contains(err.Error(), "invalid job") {
			return ErrJobNotFound
		}
		return fmt.Errorf("mjobctl cancel failed: %w", err)
	}

	return nil
}

// HoldJob puts a job on hold using mjobctl -h
func (c *ProductionMOABClient) HoldJob(ctx context.Context, moabJobID string) error {
	if !c.IsConnected() {
		return ErrMOABNotConnected
	}

	_, err := c.executeWithRetry(ctx, fmt.Sprintf("mjobctl -h %s", moabJobID), "")
	if err != nil {
		if strings.Contains(err.Error(), "job not found") {
			return ErrJobNotFound
		}
		return fmt.Errorf("mjobctl hold failed: %w", err)
	}

	return nil
}

// ReleaseJob releases a held job using mjobctl -u
func (c *ProductionMOABClient) ReleaseJob(ctx context.Context, moabJobID string) error {
	if !c.IsConnected() {
		return ErrMOABNotConnected
	}

	_, err := c.executeWithRetry(ctx, fmt.Sprintf("mjobctl -u %s", moabJobID), "")
	if err != nil {
		if strings.Contains(err.Error(), "job not found") {
			return ErrJobNotFound
		}
		return fmt.Errorf("mjobctl release failed: %w", err)
	}

	return nil
}

// GetJobStatus gets job status using checkjob
func (c *ProductionMOABClient) GetJobStatus(ctx context.Context, moabJobID string) (*MOABJob, error) {
	if !c.IsConnected() {
		return nil, ErrMOABNotConnected
	}

	// Use checkjob with XML output for easier parsing
	output, err := c.executeWithRetry(ctx, fmt.Sprintf("checkjob --xml %s", moabJobID), "")
	if err != nil {
		if strings.Contains(err.Error(), "job not found") ||
			strings.Contains(err.Error(), "invalid job") ||
			strings.Contains(err.Error(), "ERROR:  invalid job specified") {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("checkjob failed: %w", err)
	}

	// Parse XML output
	job, err := parseCheckjobXML(output, moabJobID)
	if err != nil {
		// Fallback to text parsing if XML fails
		job, err = parseCheckjobText(output, moabJobID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse checkjob output: %w", err)
		}
	}

	return job, nil
}

// GetJobAccounting gets job accounting data
func (c *ProductionMOABClient) GetJobAccounting(ctx context.Context, moabJobID string) (*MOABUsageMetrics, error) {
	if !c.IsConnected() {
		return nil, ErrMOABNotConnected
	}

	// Use showstats or mdiag -j for job accounting
	// First try checkjob which includes some accounting data
	output, err := c.executeWithRetry(ctx, fmt.Sprintf("checkjob --xml %s", moabJobID), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get job accounting: %w", err)
	}

	metrics, err := parseJobAccountingXML(output)
	if err != nil {
		// Fallback to mdiag -j
		output, err = c.executeWithRetry(ctx, fmt.Sprintf("mdiag -j %s", moabJobID), "")
		if err != nil {
			return nil, fmt.Errorf("failed to get job accounting: %w", err)
		}
		metrics, err = parseJobAccountingText(output)
		if err != nil {
			return nil, fmt.Errorf("failed to parse job accounting: %w", err)
		}
	}

	return metrics, nil
}

// ListQueues lists available queues using mdiag -q
func (c *ProductionMOABClient) ListQueues(ctx context.Context) ([]QueueInfo, error) {
	if !c.IsConnected() {
		return nil, ErrMOABNotConnected
	}

	// Use mdiag -q for queue information
	output, err := c.executeWithRetry(ctx, "mdiag -q --xml", "")
	if err != nil {
		// Try without XML if not supported
		output, err = c.executeWithRetry(ctx, "mdiag -q", "")
		if err != nil {
			return nil, fmt.Errorf("mdiag -q failed: %w", err)
		}
		return parseQueuesText(output)
	}

	return parseQueuesXML(output)
}

// ListNodes lists nodes using mdiag -n
func (c *ProductionMOABClient) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	if !c.IsConnected() {
		return nil, ErrMOABNotConnected
	}

	output, err := c.executeWithRetry(ctx, "mdiag -n --xml", "")
	if err != nil {
		// Try without XML if not supported
		output, err = c.executeWithRetry(ctx, "mdiag -n", "")
		if err != nil {
			return nil, fmt.Errorf("mdiag -n failed: %w", err)
		}
		return parseNodesText(output)
	}

	return parseNodesXML(output)
}

// GetClusterInfo gets cluster information using mdiag -s
func (c *ProductionMOABClient) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	if !c.IsConnected() {
		return nil, ErrMOABNotConnected
	}

	output, err := c.executeWithRetry(ctx, "mdiag -s --xml", "")
	if err != nil {
		// Try without XML if not supported
		output, err = c.executeWithRetry(ctx, "mdiag -s", "")
		if err != nil {
			return nil, fmt.Errorf("mdiag -s failed: %w", err)
		}
		return parseClusterInfoText(output)
	}

	return parseClusterInfoXML(output)
}

// GetReservations lists reservations using mdiag -r
func (c *ProductionMOABClient) GetReservations(ctx context.Context) ([]ReservationInfo, error) {
	if !c.IsConnected() {
		return nil, ErrMOABNotConnected
	}

	output, err := c.executeWithRetry(ctx, "mdiag -r --xml", "")
	if err != nil {
		// Try without XML if not supported
		output, err = c.executeWithRetry(ctx, "mdiag -r", "")
		if err != nil {
			return nil, fmt.Errorf("mdiag -r failed: %w", err)
		}
		return parseReservationsText(output)
	}

	return parseReservationsXML(output)
}

// executeWithRetry executes a command with retry logic
func (c *ProductionMOABClient) executeWithRetry(ctx context.Context, command string, stdin string) (string, error) {
	var lastErr error
	maxRetries := c.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			backoff := calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		output, err := c.execute(ctx, command, stdin)
		if err == nil {
			return output, nil
		}

		lastErr = err

		// Don't retry on permanent errors
		if isPermanentError(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

// execute executes a command via SSH
func (c *ProductionMOABClient) execute(ctx context.Context, command string, stdin string) (string, error) {
	c.mu.RLock()
	pool := c.pool
	c.mu.RUnlock()

	if pool == nil {
		return "", ErrMOABNotConnected
	}

	conn, err := pool.getConnection(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get connection: %w", err)
	}
	defer pool.returnConnection(conn)

	session, err := conn.client.NewSession()
	if err != nil {
		// Mark connection as bad
		closeSSHClient(conn.client)
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set up stdin if provided
	if stdin != "" {
		stdinPipe, err := session.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		go func() {
			defer stdinPipe.Close()
			_, _ = io.WriteString(stdinPipe, stdin)
		}()
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Run command with context timeout
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			// Include stderr in error message
			errMsg := stderr.String()
			if errMsg != "" {
				return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(errMsg))
			}
			return "", err
		}
	}

	return stdout.String(), nil
}

// Connection pool methods

func (p *sshConnectionPool) createConnection(ctx context.Context) (*sshConnection, error) {
	addr := fmt.Sprintf("%s:%d", p.config.ServerHost, p.config.ServerPort)

	sshConfig := &ssh.ClientConfig{
		User:            p.config.Username,
		Auth:            []ssh.AuthMethod{p.authMethod},
		HostKeyCallback: p.hostKeyCallback,
		Timeout:         p.config.ConnectionTimeout,
	}

	// Use context for connection timeout
	var conn net.Conn
	var err error

	dialer := &net.Dialer{Timeout: p.config.ConnectionTimeout}
	conn, err = dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	// SSH handshake
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, sshConfig)
	if err != nil {
		closeNetConn(conn)
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)

	now := time.Now()
	return &sshConnection{
		client:    client,
		createdAt: now,
		lastUsed:  now,
	}, nil
}

func (p *sshConnectionPool) getConnection(ctx context.Context) (*sshConnection, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("pool is closed")
	}
	p.mu.Unlock()

	// Try to get an existing connection
	select {
	case conn := <-p.connections:
		// Check if connection is still valid
		if time.Since(conn.lastUsed) > p.idleTimeout {
			closeSSHClient(conn.client)
			return p.createConnection(ctx)
		}
		conn.lastUsed = time.Now()
		return conn, nil
	default:
		// No available connections, create a new one
		return p.createConnection(ctx)
	}
}

func (p *sshConnectionPool) returnConnection(conn *sshConnection) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		closeSSHClient(conn.client)
		return
	}
	p.mu.Unlock()

	conn.lastUsed = time.Now()

	select {
	case p.connections <- conn:
		// Connection returned to pool
	default:
		// Pool is full, close connection
		closeSSHClient(conn.client)
	}
}

func (p *sshConnectionPool) close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()

	close(p.connections)
	for conn := range p.connections {
		closeSSHClient(conn.client)
	}
}

// Helper functions

func closeSSHClient(client *ssh.Client) {
	if client == nil {
		return
	}
	if err := client.Close(); err != nil {
		// best-effort cleanup
	}
}

func closeNetConn(conn net.Conn) {
	if conn == nil {
		return
	}
	if err := conn.Close(); err != nil {
		// best-effort cleanup
	}
}

func calculateBackoff(attempt int) time.Duration {
	// Base backoff: 100ms, 200ms, 400ms, 800ms...
	base := time.Duration(100) * time.Millisecond
	if attempt < 0 {
		attempt = 0
	}
	if attempt > 30 {
		attempt = 30
	}
	// #nosec G115 -- attempt is bounded to a safe range for shift.
	backoff := base * (1 << uint(attempt))

	// Cap at 30 seconds
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}

	// Add jitter (0-25% of backoff)
	jitterMax := int64(backoff / 4)
	if jitterMax > 0 {
		jitterBig, _ := rand.Int(rand.Reader, big.NewInt(jitterMax))
		backoff += time.Duration(jitterBig.Int64())
	}

	return backoff
}

func isPermanentError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "job not found") ||
		strings.Contains(errStr, "invalid job") ||
		strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "authentication failed") ||
		strings.Contains(errStr, "queue not found") ||
		errors.Is(err, ErrJobNotFound) ||
		errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrQueueNotFound)
}

func isValidJobID(jobID string) bool {
	// MOAB job IDs are typically "moab.NNNNNN" or just numeric
	if strings.HasPrefix(jobID, "moab.") {
		return true
	}
	// Check if it's numeric
	_, err := strconv.ParseInt(jobID, 10, 64)
	return err == nil
}

// XML parsing structures and functions

func decodeMOABXML(output string, v interface{}) error {
	decoder := xml.NewDecoder(strings.NewReader(output))
	decoder.Entity = map[string]string{}
	return decoder.Decode(v)
}

type checkjobXMLResponse struct {
	XMLName xml.Name     `xml:"Data"`
	Jobs    []jobXMLData `xml:"job"`
}

type jobXMLData struct {
	JobID          string `xml:"JobID,attr"`
	State          string `xml:"State,attr"`
	CompletionCode string `xml:"CompletionCode,attr"`
	StartTime      int64  `xml:"StartTime,attr"`
	EndTime        int64  `xml:"CompletionTime,attr"`
	SubmitTime     int64  `xml:"SubmitTime,attr"`
	NodeList       string `xml:"AllocNodeList,attr"`
	ExitCode       int32  `xml:"EState,attr"`
	Queue          string `xml:"Class,attr"`
	ReqNodes       int32  `xml:"ReqNodes,attr"`
	ReqProcs       int32  `xml:"ReqProcs,attr"`
	Message        string `xml:"Message,attr"`
	AWallTime      int64  `xml:"AWallTime,attr"`
	SysSTime       int64  `xml:"SysSTime,attr"`
	SysUTime       int64  `xml:"SysUTime,attr"`
}

//nolint:unparam // moabJobID kept for parse context and diagnostics
func parseCheckjobXML(output string, _ string) (*MOABJob, error) {
	var response checkjobXMLResponse
	if err := decodeMOABXML(output, &response); err != nil {
		return nil, err
	}

	if len(response.Jobs) == 0 {
		return nil, ErrJobNotFound
	}

	jobData := response.Jobs[0]

	job := &MOABJob{
		MOABJobID:      jobData.JobID,
		State:          mapMOABState(jobData.State),
		StatusMessage:  jobData.Message,
		ExitCode:       jobData.ExitCode,
		CompletionCode: jobData.CompletionCode,
		SubmitTime:     time.Unix(jobData.SubmitTime, 0),
	}

	if jobData.StartTime > 0 {
		startTime := time.Unix(jobData.StartTime, 0)
		job.StartTime = &startTime
	}
	if jobData.EndTime > 0 {
		endTime := time.Unix(jobData.EndTime, 0)
		job.EndTime = &endTime
	}
	if jobData.NodeList != "" {
		job.NodeList = strings.Split(jobData.NodeList, ",")
	}

	// Include basic metrics if available
	if jobData.AWallTime > 0 {
		job.UsageMetrics = &MOABUsageMetrics{
			WallClockSeconds: jobData.AWallTime,
			CPUTimeSeconds:   jobData.SysSTime + jobData.SysUTime,
		}
	}

	return job, nil
}

//nolint:unparam // result 1 (error) reserved for future parse failures
func parseCheckjobText(output string, moabJobID string) (*MOABJob, error) {
	job := &MOABJob{
		MOABJobID: moabJobID,
		State:     MOABJobStateIdle,
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse key: value pairs
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "State":
				job.State = mapMOABState(value)
			case "Completion Code":
				job.CompletionCode = value
			case "Start Time":
				if t, err := parseTime(value); err == nil {
					job.StartTime = &t
				}
			case "Completion Time", "End Time":
				if t, err := parseTime(value); err == nil {
					job.EndTime = &t
				}
			case "Submit Time":
				if t, err := parseTime(value); err == nil {
					job.SubmitTime = t
				}
			case "Allocated Nodes", "NodeList":
				job.NodeList = strings.Split(value, ",")
			case "Exit Code":
				if code, err := strconv.ParseInt(value, 10, 32); err == nil {
					job.ExitCode = int32(code)
				}
			case "Message":
				job.StatusMessage = value
			}
		}
	}

	return job, nil
}

func parseJobAccountingXML(output string) (*MOABUsageMetrics, error) {
	var response checkjobXMLResponse
	if err := decodeMOABXML(output, &response); err != nil {
		return nil, err
	}

	if len(response.Jobs) == 0 {
		return nil, ErrJobNotFound
	}

	jobData := response.Jobs[0]
	return &MOABUsageMetrics{
		WallClockSeconds: jobData.AWallTime,
		CPUTimeSeconds:   jobData.SysSTime + jobData.SysUTime,
		SUSUsed:          float64(jobData.AWallTime) * float64(jobData.ReqNodes) / 3600.0,
		NodeHours:        float64(jobData.AWallTime) * float64(jobData.ReqNodes) / 3600.0,
	}, nil
}

//nolint:unparam // result 1 (error) reserved for future parse failures
func parseJobAccountingText(output string) (*MOABUsageMetrics, error) {
	metrics := &MOABUsageMetrics{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "WallTime", "Wallclock":
				if secs, err := parseDurationSeconds(value); err == nil {
					metrics.WallClockSeconds = secs
				}
			case "CPUTime", "CPU Time":
				if secs, err := parseDurationSeconds(value); err == nil {
					metrics.CPUTimeSeconds = secs
				}
			case "MaxRSS", "Max RSS":
				if bytes, err := parseBytes(value); err == nil {
					metrics.MaxRSSBytes = bytes
				}
			case "MaxVMSize", "Max VM Size":
				if bytes, err := parseBytes(value); err == nil {
					metrics.MaxVMSizeBytes = bytes
				}
			case "GPUTime", "GPU Time":
				if secs, err := parseDurationSeconds(value); err == nil {
					metrics.GPUSeconds = secs
				}
			case "SUS", "Service Units":
				if sus, err := strconv.ParseFloat(value, 64); err == nil {
					metrics.SUSUsed = sus
				}
			case "NodeHours", "Node Hours":
				if nh, err := strconv.ParseFloat(value, 64); err == nil {
					metrics.NodeHours = nh
				}
			}
		}
	}

	return metrics, nil
}

type queuesXMLResponse struct {
	XMLName xml.Name       `xml:"Data"`
	Classes []queueXMLData `xml:"class"`
}

type queueXMLData struct {
	Name         string `xml:"Name,attr"`
	State        string `xml:"State,attr"`
	MaxNode      int32  `xml:"MAXNODE,attr"`
	MaxWalltime  int64  `xml:"MAXWCLIMIT,attr"`
	DefaultNodes int32  `xml:"DEFNODES,attr"`
	Priority     int32  `xml:"PRIORITY,attr"`
	IdleJobs     int32  `xml:"IDLEJOBS,attr"`
	RunningJobs  int32  `xml:"ACTIVEJOBS,attr"`
	HeldJobs     int32  `xml:"HELDJOBS,attr"`
}

func parseQueuesXML(output string) ([]QueueInfo, error) {
	var response queuesXMLResponse
	if err := decodeMOABXML(output, &response); err != nil {
		return nil, err
	}

	queues := make([]QueueInfo, len(response.Classes))
	for i, c := range response.Classes {
		queues[i] = QueueInfo{
			Name:         c.Name,
			State:        c.State,
			MaxNodes:     c.MaxNode,
			MaxWalltime:  c.MaxWalltime,
			DefaultNodes: c.DefaultNodes,
			Priority:     c.Priority,
			IdleJobs:     c.IdleJobs,
			RunningJobs:  c.RunningJobs,
			HeldJobs:     c.HeldJobs,
		}
	}

	return queues, nil
}

func parseQueuesText(output string) ([]QueueInfo, error) {
	var queues []QueueInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	// Skip header lines
	headerSkipped := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Skip header lines (usually start with "Name" or contain dashes)
		if !headerSkipped {
			if strings.HasPrefix(strings.ToLower(line), "name") || strings.HasPrefix(line, "-") {
				headerSkipped = true
				continue
			}
		}

		// Parse queue line
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			queue := QueueInfo{
				Name:  fields[0],
				State: "Active",
			}
			if len(fields) >= 3 {
				if maxNodes, err := strconv.ParseInt(fields[2], 10, 32); err == nil {
					queue.MaxNodes = int32(maxNodes)
				}
			}
			queues = append(queues, queue)
		}
	}

	return queues, nil
}

type nodesXMLResponse struct {
	XMLName xml.Name      `xml:"Data"`
	Nodes   []nodeXMLData `xml:"node"`
}

type nodeXMLData struct {
	Name       string  `xml:"NodeID,attr"`
	State      string  `xml:"State,attr"`
	Processors int32   `xml:"ConfiguredProcessors,attr"`
	MemoryMB   int64   `xml:"ConfiguredMemory,attr"`
	GPUs       int32   `xml:"GRES,attr"`
	Load       float64 `xml:"Load,attr"`
	AllocCPU   int32   `xml:"AllocatedProcessors,attr"`
	AllocMem   int64   `xml:"AllocatedMemory,attr"`
	Features   string  `xml:"Features,attr"`
}

func parseNodesXML(output string) ([]NodeInfo, error) {
	var response nodesXMLResponse
	if err := decodeMOABXML(output, &response); err != nil {
		return nil, err
	}

	nodes := make([]NodeInfo, len(response.Nodes))
	for i, n := range response.Nodes {
		nodes[i] = NodeInfo{
			Name:         n.Name,
			State:        n.State,
			Processors:   n.Processors,
			MemoryMB:     n.MemoryMB,
			GPUs:         n.GPUs,
			Load:         n.Load,
			AllocatedCPU: n.AllocCPU,
			AllocatedMem: n.AllocMem,
		}
		if n.Features != "" {
			nodes[i].Features = strings.Split(n.Features, ",")
		}
	}

	return nodes, nil
}

func parseNodesText(output string) ([]NodeInfo, error) {
	var nodes []NodeInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	headerSkipped := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if !headerSkipped {
			if strings.HasPrefix(strings.ToLower(line), "node") || strings.HasPrefix(line, "-") {
				headerSkipped = true
				continue
			}
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			node := NodeInfo{
				Name:  fields[0],
				State: fields[1],
			}
			if procs, err := strconv.ParseInt(fields[2], 10, 32); err == nil {
				node.Processors = int32(procs)
			}
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

type clusterXMLResponse struct {
	XMLName xml.Name       `xml:"Data"`
	Cluster clusterXMLData `xml:"cluster"`
}

type clusterXMLData struct {
	Name         string `xml:"Name,attr"`
	TotalNodes   int32  `xml:"TotalNodes,attr"`
	IdleNodes    int32  `xml:"IdleNodes,attr"`
	BusyNodes    int32  `xml:"BusyNodes,attr"`
	DownNodes    int32  `xml:"DownNodes,attr"`
	TotalProcs   int32  `xml:"TotalProcs,attr"`
	IdleProcs    int32  `xml:"IdleProcs,attr"`
	RunningJobs  int32  `xml:"ActiveJobs,attr"`
	IdleJobs     int32  `xml:"IdleJobs,attr"`
	Reservations int32  `xml:"ActiveRsvs,attr"`
}

func parseClusterInfoXML(output string) (*ClusterInfo, error) {
	var response clusterXMLResponse
	if err := decodeMOABXML(output, &response); err != nil {
		return nil, err
	}

	c := response.Cluster
	return &ClusterInfo{
		Name:               c.Name,
		TotalNodes:         c.TotalNodes,
		IdleNodes:          c.IdleNodes,
		BusyNodes:          c.BusyNodes,
		DownNodes:          c.DownNodes,
		TotalProcessors:    c.TotalProcs,
		IdleProcessors:     c.IdleProcs,
		RunningJobs:        c.RunningJobs,
		IdleJobs:           c.IdleJobs,
		ActiveReservations: c.Reservations,
	}, nil
}

func parseClusterInfoText(output string) (*ClusterInfo, error) {
	info := &ClusterInfo{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Name", "Cluster Name":
				info.Name = value
			case "Total Nodes":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.TotalNodes = int32(n)
				}
			case "Idle Nodes":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.IdleNodes = int32(n)
				}
			case "Busy Nodes":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.BusyNodes = int32(n)
				}
			case "Down Nodes":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.DownNodes = int32(n)
				}
			case "Total Processors", "Total Procs":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.TotalProcessors = int32(n)
				}
			case "Idle Processors", "Idle Procs":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.IdleProcessors = int32(n)
				}
			case "Running Jobs", "Active Jobs":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.RunningJobs = int32(n)
				}
			case "Idle Jobs", "Queued Jobs":
				if n, err := strconv.ParseInt(value, 10, 32); err == nil {
					info.IdleJobs = int32(n)
				}
			}
		}
	}

	return info, nil
}

type reservationsXMLResponse struct {
	XMLName      xml.Name             `xml:"Data"`
	Reservations []reservationXMLData `xml:"rsv"`
}

type reservationXMLData struct {
	Name      string `xml:"Name,attr"`
	StartTime int64  `xml:"StartTime,attr"`
	EndTime   int64  `xml:"EndTime,attr"`
	Nodes     string `xml:"AllocNodeList,attr"`
	Owner     string `xml:"Owner,attr"`
	State     string `xml:"State,attr"`
}

func parseReservationsXML(output string) ([]ReservationInfo, error) {
	var response reservationsXMLResponse
	if err := decodeMOABXML(output, &response); err != nil {
		return nil, err
	}

	reservations := make([]ReservationInfo, len(response.Reservations))
	for i, r := range response.Reservations {
		reservations[i] = ReservationInfo{
			Name:      r.Name,
			StartTime: time.Unix(r.StartTime, 0),
			EndTime:   time.Unix(r.EndTime, 0),
			Owner:     r.Owner,
			State:     r.State,
		}
		if r.Nodes != "" {
			reservations[i].Nodes = strings.Split(r.Nodes, ",")
		}
	}

	return reservations, nil
}

func parseReservationsText(output string) ([]ReservationInfo, error) {
	var reservations []ReservationInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	headerSkipped := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if !headerSkipped {
			if strings.HasPrefix(strings.ToLower(line), "name") || strings.HasPrefix(line, "-") {
				headerSkipped = true
				continue
			}
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			rsv := ReservationInfo{
				Name:  fields[0],
				State: "Active",
			}
			reservations = append(reservations, rsv)
		}
	}

	return reservations, nil
}

func mapMOABState(state string) MOABJobState {
	state = strings.ToLower(strings.TrimSpace(state))
	switch state {
	case "idle", "pending", "eligible":
		return MOABJobStateIdle
	case "starting":
		return MOABJobStateStarting
	case "running", "active":
		return MOABJobStateRunning
	case "completed", "complete":
		return MOABJobStateCompleted
	case "removed":
		return MOABJobStateRemoved
	case "hold", "held", "userhold", "systemhold", "batchhold":
		return MOABJobStateHold
	case "suspended":
		return MOABJobStateSuspended
	case "vacated":
		return MOABJobStateVacated
	case "cancelled", "canceled":
		return MOABJobStateCancelled
	case "deferred":
		return MOABJobStateDeferred
	case "failed":
		return MOABJobStateFailed
	default:
		return MOABJobStateIdle
	}
}

func parseTime(s string) (time.Time, error) {
	// Try various time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"Mon Jan 2 15:04:05 2006",
		"Jan 2 15:04:05 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Try Unix timestamp
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

func parseDurationSeconds(s string) (int64, error) {
	// Try HH:MM:SS format
	re := regexp.MustCompile(`(\d+):(\d+):(\d+)`)
	if matches := re.FindStringSubmatch(s); len(matches) == 4 {
		hours, _ := strconv.ParseInt(matches[1], 10, 64)
		minutes, _ := strconv.ParseInt(matches[2], 10, 64)
		seconds, _ := strconv.ParseInt(matches[3], 10, 64)
		return hours*3600 + minutes*60 + seconds, nil
	}

	// Try seconds only
	return strconv.ParseInt(strings.TrimSuffix(s, "s"), 10, 64)
}

func parseBytes(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	multipliers := map[string]int64{
		"B":  1,
		"K":  1024,
		"KB": 1024,
		"M":  1024 * 1024,
		"MB": 1024 * 1024,
		"G":  1024 * 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"T":  1024 * 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(num * float64(mult)), nil
		}
	}

	return strconv.ParseInt(s, 10, 64)
}
