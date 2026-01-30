// Package moab_adapter implements the MOAB workload manager adapter for VirtEngine.
//
// VE-917: MOAB workload manager using Waldur
package moab_adapter

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// MockMOABClient is a mock MOAB client for testing
type MockMOABClient struct {
	mu          sync.RWMutex
	connected   bool
	jobs        map[string]*MOABJob
	queues      []QueueInfo
	nodes       []NodeInfo
	clusterInfo *ClusterInfo
	nextJobID   int

	// Error injection for testing
	submitError  error
	cancelError  error
	holdError    error
	releaseError error
	statusError  error
}

// NewMockMOABClient creates a new mock MOAB client
func NewMockMOABClient() *MockMOABClient {
	return &MockMOABClient{
		jobs: make(map[string]*MOABJob),
		queues: []QueueInfo{
			{
				Name:         "batch",
				State:        "Active",
				MaxNodes:     100,
				MaxWalltime:  86400,
				DefaultNodes: 1,
				Priority:     1000,
				IdleJobs:     5,
				RunningJobs:  10,
				HeldJobs:     2,
			},
			{
				Name:        "gpu",
				State:       "Active",
				MaxNodes:    20,
				MaxWalltime: 172800,
				Priority:    2000,
				Features:    []string{"gpu", "nvidia"},
				IdleJobs:    3,
				RunningJobs: 8,
			},
			{
				Name:        "debug",
				State:       "Active",
				MaxNodes:    4,
				MaxWalltime: 3600,
				Priority:    5000,
				IdleJobs:    0,
				RunningJobs: 1,
			},
			{
				Name:        "interactive",
				State:       "Active",
				MaxNodes:    10,
				MaxWalltime: 7200,
				Priority:    3000,
				IdleJobs:    2,
				RunningJobs: 3,
			},
		},
		nodes: []NodeInfo{
			{Name: "compute001", State: "Idle", Processors: 64, MemoryMB: 256000, Features: []string{"compute"}, Load: 0.1},
			{Name: "compute002", State: "Idle", Processors: 64, MemoryMB: 256000, Features: []string{"compute"}, Load: 0.2},
			{Name: "compute003", State: "Busy", Processors: 64, MemoryMB: 256000, Features: []string{"compute"}, Load: 0.95, AllocatedCPU: 60, AllocatedMem: 240000},
			{Name: "compute004", State: "Busy", Processors: 64, MemoryMB: 256000, Features: []string{"compute"}, Load: 0.85, AllocatedCPU: 50, AllocatedMem: 200000},
			{Name: "gpu001", State: "Idle", Processors: 32, MemoryMB: 512000, GPUs: 8, GPUType: "nvidia-a100", Features: []string{"gpu", "nvidia"}, Load: 0.1},
			{Name: "gpu002", State: "Busy", Processors: 32, MemoryMB: 512000, GPUs: 8, GPUType: "nvidia-a100", Features: []string{"gpu", "nvidia"}, Load: 0.9, AllocatedCPU: 28, AllocatedMem: 450000},
		},
		clusterInfo: &ClusterInfo{
			Name:               "virtengine-hpc",
			TotalNodes:         6,
			IdleNodes:          3,
			BusyNodes:          3,
			DownNodes:          0,
			TotalProcessors:    320,
			IdleProcessors:     160,
			RunningJobs:        22,
			IdleJobs:           10,
			ActiveReservations: 2,
		},
		nextJobID: 10000,
	}
}

// SetSubmitError sets an error to return on submit
func (c *MockMOABClient) SetSubmitError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.submitError = err
}

// SetCancelError sets an error to return on cancel
func (c *MockMOABClient) SetCancelError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelError = err
}

// SetStatusError sets an error to return on status check
func (c *MockMOABClient) SetStatusError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statusError = err
}

// Connect connects to the mock MOAB server
func (c *MockMOABClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = true
	return nil
}

// Disconnect disconnects from mock MOAB
func (c *MockMOABClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return nil
}

// IsConnected checks if connected
func (c *MockMOABClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SubmitJob submits a job to mock MOAB (msub equivalent)
func (c *MockMOABClient) SubmitJob(ctx context.Context, spec *MOABJobSpec) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", ErrMOABNotConnected
	}

	if c.submitError != nil {
		return "", c.submitError
	}

	// Validate queue exists
	queueExists := false
	for _, q := range c.queues {
		if q.Name == spec.Queue {
			queueExists = true
			break
		}
	}
	if !queueExists {
		return "", ErrQueueNotFound
	}

	jobID := fmt.Sprintf("moab.%d", c.nextJobID)
	c.nextJobID++

	now := time.Now()
	job := &MOABJob{
		MOABJobID:  jobID,
		Spec:       spec,
		State:      MOABJobStateIdle,
		SubmitTime: now,
	}
	c.jobs[jobID] = job

	// Simulate job starting after a short delay
	verrors.SafeGo("", func() {
		defer func() { }() // WG Done if needed
		time.Sleep(50 * time.Millisecond)
		c.mu.Lock()
		if j, exists := c.jobs[jobID]; exists && j.State == MOABJobStateIdle {
			j.State = MOABJobStateStarting
		}
		c.mu.Unlock()

		time.Sleep(50 * time.Millisecond)
		c.mu.Lock()
		if j, exists := c.jobs[jobID]; exists && j.State == MOABJobStateStarting {
			j.State = MOABJobStateRunning
			startTime := time.Now()
			j.StartTime = &startTime
			j.NodeList = []string{"compute001"}
		}
		c.mu.Unlock()
	}()

	return jobID, nil
}

// CancelJob cancels a job (mjobctl -c equivalent)
func (c *MockMOABClient) CancelJob(ctx context.Context, moabJobID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrMOABNotConnected
	}

	if c.cancelError != nil {
		return c.cancelError
	}

	job, exists := c.jobs[moabJobID]
	if !exists {
		return ErrJobNotFound
	}

	if isTerminalState(job.State) {
		return errors.New("cannot cancel completed job")
	}

	job.State = MOABJobStateCancelled
	job.StatusMessage = "Job cancelled by user via mjobctl"
	job.CompletionCode = "CANCELLED"
	now := time.Now()
	job.EndTime = &now

	return nil
}

// HoldJob puts a job on hold (mjobctl -h equivalent)
func (c *MockMOABClient) HoldJob(ctx context.Context, moabJobID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrMOABNotConnected
	}

	if c.holdError != nil {
		return c.holdError
	}

	job, exists := c.jobs[moabJobID]
	if !exists {
		return ErrJobNotFound
	}

	if job.State != MOABJobStateIdle {
		return errors.New("can only hold idle jobs")
	}

	job.State = MOABJobStateHold
	job.StatusMessage = "Job held by user"

	return nil
}

// ReleaseJob releases a held job (mjobctl -u equivalent)
func (c *MockMOABClient) ReleaseJob(ctx context.Context, moabJobID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrMOABNotConnected
	}

	if c.releaseError != nil {
		return c.releaseError
	}

	job, exists := c.jobs[moabJobID]
	if !exists {
		return ErrJobNotFound
	}

	if job.State != MOABJobStateHold {
		return errors.New("job is not held")
	}

	job.State = MOABJobStateIdle
	job.StatusMessage = "Job released from hold"

	return nil
}

// GetJobStatus gets job status (checkjob equivalent)
func (c *MockMOABClient) GetJobStatus(ctx context.Context, moabJobID string) (*MOABJob, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrMOABNotConnected
	}

	if c.statusError != nil {
		return nil, c.statusError
	}

	job, exists := c.jobs[moabJobID]
	if !exists {
		return nil, ErrJobNotFound
	}

	// Make a copy to avoid race conditions
	jobCopy := *job
	if job.Spec != nil {
		specCopy := *job.Spec
		jobCopy.Spec = &specCopy
	}
	return &jobCopy, nil
}

// GetJobAccounting gets job accounting data
func (c *MockMOABClient) GetJobAccounting(ctx context.Context, moabJobID string) (*MOABUsageMetrics, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrMOABNotConnected
	}

	job, exists := c.jobs[moabJobID]
	if !exists {
		return nil, ErrJobNotFound
	}

	if !isTerminalState(job.State) {
		return nil, errors.New("job not completed")
	}

	// Calculate mock metrics
	var wallClock int64
	if job.StartTime != nil && job.EndTime != nil {
		wallClock = int64(job.EndTime.Sub(*job.StartTime).Seconds())
	}

	cpuTime := wallClock * int64(job.Spec.ProcsPerNode) * int64(job.Spec.Nodes)
	nodeHours := float64(wallClock) * float64(job.Spec.Nodes) / 3600.0

	return &MOABUsageMetrics{
		WallClockSeconds: wallClock,
		CPUTimeSeconds:   cpuTime,
		MaxRSSBytes:      int64(job.Spec.MemoryMB) * 1024 * 1024,
		MaxVMSizeBytes:   int64(job.Spec.MemoryMB) * 1024 * 1024 * 2,
		GPUSeconds:       wallClock * int64(job.Spec.GPUs),
		SUSUsed:          nodeHours * 1.5, // SUS = node hours * factor
		NodeHours:        nodeHours,
		EnergyJoules:     wallClock * 500, // Mock energy consumption
	}, nil
}

// ListQueues lists available queues (mdiag -q equivalent)
func (c *MockMOABClient) ListQueues(ctx context.Context) ([]QueueInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrMOABNotConnected
	}

	// Return a copy of queues
	queues := make([]QueueInfo, len(c.queues))
	copy(queues, c.queues)
	return queues, nil
}

// ListNodes lists nodes (mdiag -n equivalent)
func (c *MockMOABClient) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrMOABNotConnected
	}

	// Return a copy of nodes
	nodes := make([]NodeInfo, len(c.nodes))
	copy(nodes, c.nodes)
	return nodes, nil
}

// GetClusterInfo gets cluster information (mdiag -s equivalent)
func (c *MockMOABClient) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrMOABNotConnected
	}

	// Return a copy of cluster info
	info := *c.clusterInfo
	return &info, nil
}

// GetReservations lists reservations (mdiag -r equivalent)
func (c *MockMOABClient) GetReservations(ctx context.Context) ([]ReservationInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrMOABNotConnected
	}

	now := time.Now()
	return []ReservationInfo{
		{
			Name:      "maintenance",
			StartTime: now.Add(24 * time.Hour),
			EndTime:   now.Add(28 * time.Hour),
			Nodes:     []string{"compute001", "compute002"},
			Owner:     "admin",
			State:     "Scheduled",
		},
		{
			Name:      "project-allocation",
			StartTime: now.Add(-12 * time.Hour),
			EndTime:   now.Add(12 * time.Hour),
			Nodes:     []string{"gpu001"},
			Owner:     "project-team",
			State:     "Active",
		},
	}, nil
}

// SimulateJobCompletion simulates job completion (for testing)
func (c *MockMOABClient) SimulateJobCompletion(moabJobID string, exitCode int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	job, exists := c.jobs[moabJobID]
	if !exists {
		return ErrJobNotFound
	}

	now := time.Now()
	// Ensure StartTime is set if not already set
	if job.StartTime == nil {
		startTime := now.Add(-1 * time.Second) // Started 1 second ago
		job.StartTime = &startTime
	}
	job.EndTime = &now
	job.ExitCode = exitCode

	if exitCode == 0 {
		job.State = MOABJobStateCompleted
		job.CompletionCode = "COMPLETED"
		job.StatusMessage = "Job completed successfully"
	} else {
		job.State = MOABJobStateFailed
		job.CompletionCode = fmt.Sprintf("EXIT:%d", exitCode)
		job.StatusMessage = fmt.Sprintf("Job failed with exit code %d", exitCode)
	}

	return nil
}

// AddQueue adds a queue for testing
func (c *MockMOABClient) AddQueue(queue QueueInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queues = append(c.queues, queue)
}

// AddNode adds a node for testing
func (c *MockMOABClient) AddNode(node NodeInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes = append(c.nodes, node)
}

// MockJobSigner is a mock job signer for testing
type MockJobSigner struct {
	providerAddress string
}

// NewMockJobSigner creates a new mock job signer
func NewMockJobSigner(providerAddress string) *MockJobSigner {
	return &MockJobSigner{providerAddress: providerAddress}
}

// Sign signs data
func (s *MockJobSigner) Sign(data []byte) ([]byte, error) {
	// Mock signature - just hash the data
	return data[:min(32, len(data))], nil
}

// Verify verifies a signature
func (s *MockJobSigner) Verify(data []byte, signature []byte) bool {
	return true
}

// GetProviderAddress returns the provider address
func (s *MockJobSigner) GetProviderAddress() string {
	return s.providerAddress
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MockRewardsIntegration is a mock rewards integration for testing
type MockRewardsIntegration struct {
	mu          sync.Mutex
	completions []*VERewardsData
	rewardRates map[string]float64
	recordError error
}

// NewMockRewardsIntegration creates a new mock rewards integration
func NewMockRewardsIntegration() *MockRewardsIntegration {
	return &MockRewardsIntegration{
		completions: make([]*VERewardsData, 0),
		rewardRates: map[string]float64{
			"MOAB":  1.5,
			"SLURM": 1.5,
		},
	}
}

// RecordJobCompletion records job completion for rewards
func (m *MockRewardsIntegration) RecordJobCompletion(ctx context.Context, data *VERewardsData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.recordError != nil {
		return m.recordError
	}

	m.completions = append(m.completions, data)
	return nil
}

// GetRewardRate gets the current reward rate for HPC jobs
func (m *MockRewardsIntegration) GetRewardRate(ctx context.Context, schedulerType string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rate, exists := m.rewardRates[schedulerType]
	if !exists {
		return 1.0, nil
	}
	return rate, nil
}

// GetCompletions returns recorded completions
func (m *MockRewardsIntegration) GetCompletions() []*VERewardsData {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]*VERewardsData, len(m.completions))
	copy(result, m.completions)
	return result
}

// SetRecordError sets an error to return on record
func (m *MockRewardsIntegration) SetRecordError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordError = err
}
