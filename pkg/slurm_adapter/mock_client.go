// Package slurm_adapter implements the SLURM orchestration adapter for VirtEngine.
//
// VE-501: SLURM orchestration adapter in Provider Daemon (v1)
package slurm_adapter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// MockSLURMClient is a mock SLURM client for testing
type MockSLURMClient struct {
	mu         sync.RWMutex
	connected  bool
	jobs       map[string]*SLURMJob
	partitions []PartitionInfo
	nodes      []NodeInfo
	nextJobID  int
}

// NewMockSLURMClient creates a new mock SLURM client
func NewMockSLURMClient() *MockSLURMClient {
	return &MockSLURMClient{
		jobs: make(map[string]*SLURMJob),
		partitions: []PartitionInfo{
			{
				Name:        "default",
				Nodes:       10,
				State:       "UP",
				MaxTime:     86400,
				DefaultTime: 3600,
				MaxNodes:    10,
			},
			{
				Name:     "gpu",
				Nodes:    4,
				State:    "UP",
				MaxTime:  172800,
				MaxNodes: 4,
				Features: []string{"gpu", "nvidia"},
			},
		},
		nodes: []NodeInfo{
			{Name: "node001", State: "IDLE", CPUs: 64, MemoryMB: 256000, Partitions: []string{"default"}},
			{Name: "node002", State: "IDLE", CPUs: 64, MemoryMB: 256000, Partitions: []string{"default"}},
			{Name: "node003", State: "ALLOCATED", CPUs: 64, MemoryMB: 256000, Partitions: []string{"default"}},
			{Name: "gpu001", State: "IDLE", CPUs: 32, MemoryMB: 512000, GPUs: 8, GPUType: "nvidia-a100", Partitions: []string{"gpu"}},
		},
		nextJobID: 1000,
	}
}

// Connect connects to the mock SLURM controller
func (c *MockSLURMClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = true
	return nil
}

// Disconnect disconnects from mock SLURM
func (c *MockSLURMClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return nil
}

// IsConnected checks if connected
func (c *MockSLURMClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SubmitJob submits a job to mock SLURM
func (c *MockSLURMClient) SubmitJob(ctx context.Context, spec *SLURMJobSpec) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", ErrSLURMNotConnected
	}

	jobID := fmt.Sprintf("%d", c.nextJobID)
	c.nextJobID++

	now := time.Now()
	job := &SLURMJob{
		SLURMJobID: jobID,
		Spec:       spec,
		State:      SLURMJobStatePending,
		SubmitTime: now,
	}
	c.jobs[jobID] = job

	// Simulate job starting after a short delay
	verrors.SafeGo("", func() {
		defer func() {}() // WG Done if needed
		time.Sleep(100 * time.Millisecond)
		c.mu.Lock()
		if j, exists := c.jobs[jobID]; exists && j.State == SLURMJobStatePending {
			j.State = SLURMJobStateRunning
			startTime := time.Now()
			j.StartTime = &startTime
			j.NodeList = []string{"node001"}
		}
		c.mu.Unlock()
	})

	return jobID, nil
}

// CancelJob cancels a job
func (c *MockSLURMClient) CancelJob(ctx context.Context, slurmJobID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrSLURMNotConnected
	}

	job, exists := c.jobs[slurmJobID]
	if !exists {
		return ErrJobNotFound
	}

	if isTerminalState(job.State) {
		return errors.New("cannot cancel completed job")
	}

	job.State = SLURMJobStateCancelled
	now := time.Now()
	job.EndTime = &now

	return nil
}

// GetJobStatus gets job status
func (c *MockSLURMClient) GetJobStatus(ctx context.Context, slurmJobID string) (*SLURMJob, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrSLURMNotConnected
	}

	job, exists := c.jobs[slurmJobID]
	if !exists {
		return nil, ErrJobNotFound
	}

	// Make a copy to avoid race conditions
	jobCopy := *job
	return &jobCopy, nil
}

// GetJobAccounting gets job accounting data
func (c *MockSLURMClient) GetJobAccounting(ctx context.Context, slurmJobID string) (*SLURMUsageMetrics, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrSLURMNotConnected
	}

	job, exists := c.jobs[slurmJobID]
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

	return &SLURMUsageMetrics{
		WallClockSeconds: wallClock,
		CPUTimeSeconds:   wallClock * int64(job.Spec.CPUsPerNode) * int64(job.Spec.Nodes),
		MaxRSSBytes:      int64(job.Spec.MemoryMB) * 1024 * 1024,
		MaxVMSizeBytes:   int64(job.Spec.MemoryMB) * 1024 * 1024 * 2,
		GPUSeconds:       wallClock * int64(job.Spec.GPUs),
	}, nil
}

// ListPartitions lists available partitions
func (c *MockSLURMClient) ListPartitions(ctx context.Context) ([]PartitionInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrSLURMNotConnected
	}

	return c.partitions, nil
}

// ListNodes lists nodes in the cluster
func (c *MockSLURMClient) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, ErrSLURMNotConnected
	}

	return c.nodes, nil
}

// SimulateJobCompletion simulates a job completing (for testing)
func (c *MockSLURMClient) SimulateJobCompletion(slurmJobID string, success bool, exitCode int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	job, exists := c.jobs[slurmJobID]
	if !exists {
		return ErrJobNotFound
	}

	if success {
		job.State = SLURMJobStateCompleted
	} else {
		job.State = SLURMJobStateFailed
	}
	job.ExitCode = exitCode
	now := time.Now()
	job.EndTime = &now

	return nil
}

