// Package slurm_adapter implements the SLURM orchestration adapter for VirtEngine.
//
// VE-2020: Real SLURM adapter implementation using SSH for remote SLURM execution
package slurm_adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// ErrSSHConnection is returned when SSH connection fails
var ErrSSHConnection = errors.New("SSH connection failed")

// ErrCommandFailed is returned when a SLURM command fails
var ErrCommandFailed = errors.New("SLURM command failed")

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
	HostKeyCallback string `json:"host_key_callback"`

	// KnownHostsFile is the path to known_hosts file
	KnownHostsFile string `json:"known_hosts_file"`

	// Timeout is the SSH connection timeout
	Timeout time.Duration `json:"timeout"`

	// KeepAliveInterval is the interval for keepalive packets
	KeepAliveInterval time.Duration `json:"keepalive_interval"`

	// MaxRetries is the maximum number of reconnection attempts
	MaxRetries int `json:"max_retries"`
}

// DefaultSSHConfig returns default SSH configuration
func DefaultSSHConfig() SSHConfig {
	return SSHConfig{
		Port:              22,
		Timeout:          30 * time.Second,
		KeepAliveInterval: 30 * time.Second,
		MaxRetries:        3,
		HostKeyCallback:   "ignore", // TODO: Change to "known_hosts" for production
	}
}

// SSHSLURMClient implements SLURMClient using SSH to execute SLURM commands
type SSHSLURMClient struct {
	config     SSHConfig
	sshConfig  *ssh.ClientConfig
	client     *ssh.Client
	mu         sync.RWMutex
	connected  bool
	jobs       map[string]*SLURMJob // Local cache for job tracking
	clusterName string
	defaultPartition string
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
	default:
		// TODO: Implement known_hosts verification
		clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	return &SSHSLURMClient{
		config:          sshConfig,
		sshConfig:       clientConfig,
		jobs:            make(map[string]*SLURMJob),
		clusterName:     clusterName,
		defaultPartition: defaultPartition,
	}, nil
}

// Connect establishes SSH connection to the SLURM login node
func (c *SSHSLURMClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected && c.client != nil {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	
	var client *ssh.Client
	var err error
	
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		client, err = ssh.Dial("tcp", addr, c.sshConfig)
		if err == nil {
			break
		}
		
		if attempt < c.config.MaxRetries {
			time.Sleep(time.Second * time.Duration(attempt+1))
		}
	}

	if err != nil {
		return fmt.Errorf("%w: %v", ErrSSHConnection, err)
	}

	c.client = client
	c.connected = true

	// Test connection with a simple SLURM command
	output, err := c.runCommand(ctx, "squeue --version")
	if err != nil {
		c.Disconnect()
		return fmt.Errorf("failed to verify SLURM: %w", err)
	}

	// Log SLURM version (for debugging)
	_ = output

	return nil
}

// Disconnect closes the SSH connection
func (c *SSHSLURMClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		c.connected = false
		return err
	}
	return nil
}

// IsConnected checks if connected to SLURM
func (c *SSHSLURMClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.client != nil
}

// runCommand executes a command via SSH and returns the output
func (c *SSHSLURMClient) runCommand(ctx context.Context, cmd string) (string, error) {
	if c.client == nil {
		return "", ErrSLURMNotConnected
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Handle context cancellation
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			session.Signal(ssh.SIGTERM)
			session.Close()
		case <-done:
		}
	}()

	output, err := session.CombinedOutput(cmd)
	close(done)

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	return string(output), err
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

	// Generate SLURM batch script
	script := c.generateBatchScript(spec)

	// Create temporary script file and submit
	// Using heredoc approach to avoid file creation
	submitCmd := fmt.Sprintf(`cat << 'SLURM_SCRIPT_EOF' | sbatch --parsable
%s
SLURM_SCRIPT_EOF`, script)

	output, err := c.runCommand(ctx, submitCmd)
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

// generateBatchScript creates a SLURM batch script from the job spec
func (c *SSHSLURMClient) generateBatchScript(spec *SLURMJobSpec) string {
	var sb strings.Builder

	sb.WriteString("#!/bin/bash\n")
	sb.WriteString(fmt.Sprintf("#SBATCH --job-name=%s\n", spec.JobName))

	partition := spec.Partition
	if partition == "" {
		partition = c.defaultPartition
	}
	sb.WriteString(fmt.Sprintf("#SBATCH --partition=%s\n", partition))

	sb.WriteString(fmt.Sprintf("#SBATCH --nodes=%d\n", spec.Nodes))
	sb.WriteString(fmt.Sprintf("#SBATCH --cpus-per-task=%d\n", spec.CPUsPerNode))
	sb.WriteString(fmt.Sprintf("#SBATCH --mem=%dM\n", spec.MemoryMB))
	sb.WriteString(fmt.Sprintf("#SBATCH --time=%d\n", spec.TimeLimit))

	if spec.GPUs > 0 {
		if spec.GPUType != "" {
			sb.WriteString(fmt.Sprintf("#SBATCH --gres=gpu:%s:%d\n", spec.GPUType, spec.GPUs))
		} else {
			sb.WriteString(fmt.Sprintf("#SBATCH --gres=gpu:%d\n", spec.GPUs))
		}
	}

	if spec.WorkingDirectory != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --chdir=%s\n", spec.WorkingDirectory))
	}

	if spec.OutputDirectory != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --output=%s/%%j.out\n", spec.OutputDirectory))
		sb.WriteString(fmt.Sprintf("#SBATCH --error=%s/%%j.err\n", spec.OutputDirectory))
	}

	if spec.Exclusive {
		sb.WriteString("#SBATCH --exclusive\n")
	}

	for _, constraint := range spec.Constraints {
		sb.WriteString(fmt.Sprintf("#SBATCH --constraint=%s\n", constraint))
	}

	sb.WriteString("\n# Environment variables\n")
	for key, value := range spec.Environment {
		sb.WriteString(fmt.Sprintf("export %s=%q\n", key, value))
	}

	sb.WriteString("\n# Job execution\n")
	if spec.ContainerImage != "" {
		// Use Singularity/Apptainer for containerized workloads
		sb.WriteString(fmt.Sprintf("singularity exec %s ", spec.ContainerImage))
	}

	if spec.Command != "" {
		sb.WriteString(spec.Command)
		for _, arg := range spec.Arguments {
			sb.WriteString(fmt.Sprintf(" %q", arg))
		}
		sb.WriteString("\n")
	}

	return sb.String()
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

// Silence unused import warning
var _ = regexp.MustCompile
var _ io.Reader = nil
var _ = json.Marshal
