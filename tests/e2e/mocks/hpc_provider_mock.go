//go:build e2e.integration

// Package mocks provides mock implementations for E2E testing.
//
// VE-HPC-E2E: Mock provider daemon with queueing, scheduling, and lifecycle tracking.
package mocks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// LifecyclePhase represents high-level lifecycle phases for tests.
type LifecyclePhase string

const (
	LifecycleSubmitted  LifecyclePhase = "submitted"
	LifecycleQueued     LifecyclePhase = "queued"
	LifecycleScheduled  LifecyclePhase = "scheduled"
	LifecycleRunning    LifecyclePhase = "running"
	LifecycleCompleted  LifecyclePhase = "completed"
	LifecycleFailed     LifecyclePhase = "failed"
	LifecycleCancelled  LifecyclePhase = "cancelled"
	LifecycleTimeout    LifecyclePhase = "timeout"
	LifecycleSettled    LifecyclePhase = "settled"
	LifecycleUnassigned LifecyclePhase = "unassigned"
)

// JobQueueOptions configures queue behavior for a job.
type JobQueueOptions struct {
	Priority       int
	CustomerTier   int32
	RequiredTier   int32
	RequiredRegion string
	AllowedRegions []string
}

// QueuedJob represents a job in the provider queue.
type QueuedJob struct {
	Job         *hpctypes.HPCJob
	Options     JobQueueOptions
	SubmittedAt time.Time
}

// SchedulingDecision captures scheduling details.
type SchedulingDecision struct {
	DecisionID        string
	JobID             string
	SelectedClusterID string
	SelectedProvider  string
	Reason            string
	Scores            map[string]float64
	CreatedAt         time.Time
}

// ProviderCluster represents available cluster resources.
type ProviderCluster struct {
	ClusterID        string
	ProviderID       string
	Region           string
	AvailableCPU     int32
	AvailableMemory  int64
	AvailableGPUs    int32
	GPUType          string
	LatencyScore     float64
	PriceScore       float64
	IdentityTier     int32
	SupportsGPUTypes []string
}

// MockHPCProviderDaemon simulates provider queueing/scheduling/execution.
type MockHPCProviderDaemon struct {
	mu sync.RWMutex

	scheduler pd.HPCScheduler

	queue          []*QueuedJob
	queueDepth     int
	fairShare      bool
	runningCounts  map[string]int
	assignedJobs   map[string]*SchedulingDecision
	jobs           map[string]*hpctypes.HPCJob
	lifecycle      map[string][]LifecyclePhase
	usageSnapshots map[string][]*pd.HPCSchedulerMetrics
	clusters       map[string]*ProviderCluster
}

// NewMockHPCProviderDaemon creates a new provider mock.
func NewMockHPCProviderDaemon(scheduler pd.HPCScheduler) *MockHPCProviderDaemon {
	return &MockHPCProviderDaemon{
		scheduler:      scheduler,
		queue:          make([]*QueuedJob, 0),
		queueDepth:     100,
		fairShare:      true,
		runningCounts:  make(map[string]int),
		assignedJobs:   make(map[string]*SchedulingDecision),
		jobs:           make(map[string]*hpctypes.HPCJob),
		lifecycle:      make(map[string][]LifecyclePhase),
		usageSnapshots: make(map[string][]*pd.HPCSchedulerMetrics),
		clusters:       make(map[string]*ProviderCluster),
	}
}

// SetQueueDepthLimit sets the maximum queue depth.
func (m *MockHPCProviderDaemon) SetQueueDepthLimit(limit int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queueDepth = limit
}

// EnableFairShare enables or disables fair-share scheduling.
func (m *MockHPCProviderDaemon) EnableFairShare(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fairShare = enabled
}

// AddCluster registers a provider cluster.
func (m *MockHPCProviderDaemon) AddCluster(cluster ProviderCluster) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copyCluster := cluster
	m.clusters[cluster.ClusterID] = &copyCluster
}

// GetCluster returns a cluster by ID.
func (m *MockHPCProviderDaemon) GetCluster(clusterID string) (*ProviderCluster, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cluster, ok := m.clusters[clusterID]
	return cluster, ok
}

// EnqueueJob adds a job to the queue.
func (m *MockHPCProviderDaemon) EnqueueJob(job *hpctypes.HPCJob, opts JobQueueOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.queue) >= m.queueDepth {
		return fmt.Errorf("queue depth limit exceeded")
	}
	m.jobs[job.JobID] = job
	m.queue = append(m.queue, &QueuedJob{
		Job:         job,
		Options:     opts,
		SubmittedAt: time.Now(),
	})
	m.appendLifecycle(job.JobID, LifecycleSubmitted)
	m.appendLifecycle(job.JobID, LifecycleQueued)
	return nil
}

// PeekQueue returns job IDs in current queue order.
func (m *MockHPCProviderDaemon) PeekQueue() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.queue))
	for _, q := range m.queue {
		ids = append(ids, q.Job.JobID)
	}
	return ids
}

// ScheduleNext schedules the next job in the queue.
func (m *MockHPCProviderDaemon) ScheduleNext(ctx context.Context) (*SchedulingDecision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) == 0 {
		return nil, fmt.Errorf("queue empty")
	}

	m.sortQueueLocked()
	candidate := m.queue[0]
	m.queue = m.queue[1:]

	if candidate.Options.RequiredTier > 0 && candidate.Options.CustomerTier < candidate.Options.RequiredTier {
		m.appendLifecycle(candidate.Job.JobID, LifecycleFailed)
		return nil, fmt.Errorf("customer tier below requirement")
	}

	cluster, scores, reason, err := m.selectClusterLocked(candidate)
	if err != nil {
		m.appendLifecycle(candidate.Job.JobID, LifecycleUnassigned)
		return nil, err
	}

	decision := &SchedulingDecision{
		DecisionID:        fmt.Sprintf("sched-%s-%d", candidate.Job.JobID, time.Now().UnixNano()),
		JobID:             candidate.Job.JobID,
		SelectedClusterID: cluster.ClusterID,
		SelectedProvider:  cluster.ProviderID,
		Reason:            reason,
		Scores:            scores,
		CreatedAt:         time.Now(),
	}

	m.assignedJobs[candidate.Job.JobID] = decision
	candidate.Job.ClusterID = cluster.ClusterID
	m.appendLifecycle(candidate.Job.JobID, LifecycleScheduled)

	return decision, nil
}

// StartJob submits the job to the scheduler and marks running.
func (m *MockHPCProviderDaemon) StartJob(ctx context.Context, jobID string) (*pd.HPCSchedulerJob, error) {
	m.mu.Lock()
	decision, ok := m.assignedJobs[jobID]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("job not scheduled")
	}

	cluster, ok := m.GetCluster(decision.SelectedClusterID)
	if !ok {
		return nil, fmt.Errorf("cluster not found")
	}

	job, err := m.findJobByID(jobID)
	if err != nil {
		return nil, err
	}

	if job.Resources.CPUCoresPerNode > cluster.AvailableCPU || int64(job.Resources.MemoryGBPerNode) > cluster.AvailableMemory || job.Resources.GPUsPerNode > cluster.AvailableGPUs {
		m.appendLifecycle(jobID, LifecycleFailed)
		return nil, fmt.Errorf("insufficient cluster capacity")
	}

	schedulerJob, err := m.scheduler.SubmitJob(ctx, job)
	if err != nil {
		m.appendLifecycle(jobID, LifecycleFailed)
		return nil, err
	}

	m.mu.Lock()
	m.runningCounts[job.CustomerAddress]++
	m.appendLifecycle(jobID, LifecycleRunning)
	m.mu.Unlock()

	return schedulerJob, nil
}

// MarkCompleted marks the job completed with metrics.
func (m *MockHPCProviderDaemon) MarkCompleted(jobID string, metrics *pd.HPCSchedulerMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appendLifecycle(jobID, LifecycleCompleted)
	m.usageSnapshots[jobID] = append(m.usageSnapshots[jobID], metrics)
	m.runningCounts[m.jobCustomer(jobID)]--
}

// MarkFailed marks the job failed with metrics.
func (m *MockHPCProviderDaemon) MarkFailed(jobID string, metrics *pd.HPCSchedulerMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appendLifecycle(jobID, LifecycleFailed)
	if metrics != nil {
		m.usageSnapshots[jobID] = append(m.usageSnapshots[jobID], metrics)
	}
	m.runningCounts[m.jobCustomer(jobID)]--
}

// MarkCancelled marks the job cancelled.
func (m *MockHPCProviderDaemon) MarkCancelled(jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appendLifecycle(jobID, LifecycleCancelled)
	m.runningCounts[m.jobCustomer(jobID)]--
}

// MarkTimeout marks the job timed out.
func (m *MockHPCProviderDaemon) MarkTimeout(jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appendLifecycle(jobID, LifecycleTimeout)
	m.runningCounts[m.jobCustomer(jobID)]--
}

// MarkSettled marks the job settled.
func (m *MockHPCProviderDaemon) MarkSettled(jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appendLifecycle(jobID, LifecycleSettled)
}

// GetLifecycle returns lifecycle phases for a job.
func (m *MockHPCProviderDaemon) GetLifecycle(jobID string) []LifecyclePhase {
	m.mu.RLock()
	defer m.mu.RUnlock()
	phases := m.lifecycle[jobID]
	copyPhases := make([]LifecyclePhase, len(phases))
	copy(copyPhases, phases)
	return copyPhases
}

// GetUsageSnapshots returns stored metrics snapshots.
func (m *MockHPCProviderDaemon) GetUsageSnapshots(jobID string) []*pd.HPCSchedulerMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snapshots := m.usageSnapshots[jobID]
	copySnapshots := make([]*pd.HPCSchedulerMetrics, len(snapshots))
	copy(copySnapshots, snapshots)
	return copySnapshots
}

func (m *MockHPCProviderDaemon) appendLifecycle(jobID string, phase LifecyclePhase) {
	m.lifecycle[jobID] = append(m.lifecycle[jobID], phase)
}

func (m *MockHPCProviderDaemon) sortQueueLocked() {
	sort.SliceStable(m.queue, func(i, j int) bool {
		if m.queue[i].Options.Priority != m.queue[j].Options.Priority {
			return m.queue[i].Options.Priority > m.queue[j].Options.Priority
		}
		if m.fairShare {
			ci := m.runningCounts[m.queue[i].Job.CustomerAddress]
			cj := m.runningCounts[m.queue[j].Job.CustomerAddress]
			if ci != cj {
				return ci < cj
			}
		}
		return m.queue[i].SubmittedAt.Before(m.queue[j].SubmittedAt)
	})
}

func (m *MockHPCProviderDaemon) selectClusterLocked(job *QueuedJob) (*ProviderCluster, map[string]float64, string, error) {
	var best *ProviderCluster
	bestScore := -1.0
	scores := make(map[string]float64)

	for _, cluster := range m.clusters {
		if job.Options.RequiredRegion != "" && cluster.Region != job.Options.RequiredRegion {
			if !regionAllowed(cluster.Region, job.Options.AllowedRegions) {
				continue
			}
		}
		if job.Job.Resources.GPUsPerNode > 0 && !supportsGPUType(cluster, job.Job.Resources.GPUType) {
			continue
		}
		if job.Job.Resources.CPUCoresPerNode > cluster.AvailableCPU || int64(job.Job.Resources.MemoryGBPerNode) > cluster.AvailableMemory || job.Job.Resources.GPUsPerNode > cluster.AvailableGPUs {
			continue
		}
		resourceScore := resourceFitScore(job.Job, cluster)
		regionScore := regionScore(job.Options.RequiredRegion, cluster.Region)
		latencyScore := cluster.LatencyScore
		priceScore := cluster.PriceScore
		combined := 0.4*resourceScore + 0.3*latencyScore + 0.2*priceScore + 0.1*regionScore
		scores[cluster.ClusterID] = combined
		if combined > bestScore {
			bestScore = combined
			best = cluster
		}
	}

	if best == nil {
		return nil, scores, "no eligible clusters", fmt.Errorf("no eligible clusters")
	}

	reason := fmt.Sprintf("selected cluster %s with score %.3f", best.ClusterID, bestScore)
	return best, scores, reason, nil
}

func resourceFitScore(job *hpctypes.HPCJob, cluster *ProviderCluster) float64 {
	cpuRatio := float64(job.Resources.CPUCoresPerNode) / float64(maxInt32(1, cluster.AvailableCPU))
	memRatio := float64(job.Resources.MemoryGBPerNode) / float64(maxInt64(1, cluster.AvailableMemory))
	gpuRatio := 0.0
	if job.Resources.GPUsPerNode > 0 {
		gpuRatio = float64(job.Resources.GPUsPerNode) / float64(maxInt32(1, cluster.AvailableGPUs))
	}
	score := 1.0 - (cpuRatio+memRatio+gpuRatio)/3.0
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score
}

func regionAllowed(region string, allowed []string) bool {
	for _, r := range allowed {
		if r == region {
			return true
		}
	}
	return false
}

func regionScore(required, candidate string) float64 {
	if required == "" {
		return 1.0
	}
	if required == candidate {
		return 1.0
	}
	return 0.2
}

func supportsGPUType(cluster *ProviderCluster, gpuType string) bool {
	if gpuType == "" {
		return true
	}
	if cluster.GPUType == gpuType {
		return true
	}
	for _, t := range cluster.SupportsGPUTypes {
		if t == gpuType {
			return true
		}
	}
	return false
}

func (m *MockHPCProviderDaemon) findJobByID(jobID string) (*hpctypes.HPCJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if job, ok := m.jobs[jobID]; ok {
		return job, nil
	}
	return nil, fmt.Errorf("job not found")
}

func (m *MockHPCProviderDaemon) jobCustomer(jobID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if job, ok := m.jobs[jobID]; ok {
		return job.CustomerAddress
	}
	return ""
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
