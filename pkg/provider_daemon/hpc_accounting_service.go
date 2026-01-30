// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-5A: HPC Accounting Service - extracts, normalizes, and submits usage accounting
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCAccountingConfig configures the accounting service
type HPCAccountingConfig struct {
	// SnapshotInterval is how often to capture usage snapshots
	SnapshotInterval time.Duration `json:"snapshot_interval"`

	// SubmissionBatchSize is how many records to submit at once
	SubmissionBatchSize int `json:"submission_batch_size"`

	// RetryAttempts is how many times to retry failed submissions
	RetryAttempts int `json:"retry_attempts"`

	// RetryDelay is the delay between retries
	RetryDelay time.Duration `json:"retry_delay"`

	// ReconciliationInterval is how often to run reconciliation
	ReconciliationInterval time.Duration `json:"reconciliation_interval"`

	// EnableAutoReconciliation enables automatic reconciliation
	EnableAutoReconciliation bool `json:"enable_auto_reconciliation"`
}

// DefaultHPCAccountingConfig returns default accounting configuration
func DefaultHPCAccountingConfig() HPCAccountingConfig {
	return HPCAccountingConfig{
		SnapshotInterval:         5 * time.Minute,
		SubmissionBatchSize:      50,
		RetryAttempts:            3,
		RetryDelay:               10 * time.Second,
		ReconciliationInterval:   1 * time.Hour,
		EnableAutoReconciliation: true,
	}
}

// HPCAccountingService manages usage accounting for HPC jobs
type HPCAccountingService struct {
	config        HPCAccountingConfig
	clusterID     string
	schedulerType HPCSchedulerType
	scheduler     HPCScheduler
	signer        HPCSchedulerSigner
	submitter     HPCAccountingSubmitter

	mu              sync.RWMutex
	running         bool
	stopCh          chan struct{}
	wg              sync.WaitGroup
	jobAggregators  map[string]*HPCUsageAggregator // jobID -> aggregator
	snapshotCounter map[string]uint32              // jobID -> snapshot sequence
	pendingRecords  []*hpctypes.HPCAccountingRecord
}

// HPCAccountingSubmitter submits accounting records to the chain
type HPCAccountingSubmitter interface {
	SubmitAccountingRecord(ctx context.Context, record *hpctypes.HPCAccountingRecord) error
	SubmitUsageSnapshot(ctx context.Context, snapshot *hpctypes.HPCUsageSnapshot) error
	GetBillingRules(ctx context.Context, providerAddr string) (*hpctypes.HPCBillingRules, error)
}

// NewHPCAccountingService creates a new accounting service
func NewHPCAccountingService(
	config HPCAccountingConfig,
	clusterID string,
	schedulerType HPCSchedulerType,
	scheduler HPCScheduler,
	signer HPCSchedulerSigner,
	submitter HPCAccountingSubmitter,
) *HPCAccountingService {
	return &HPCAccountingService{
		config:          config,
		clusterID:       clusterID,
		schedulerType:   schedulerType,
		scheduler:       scheduler,
		signer:          signer,
		submitter:       submitter,
		stopCh:          make(chan struct{}),
		jobAggregators:  make(map[string]*HPCUsageAggregator),
		snapshotCounter: make(map[string]uint32),
		pendingRecords:  make([]*hpctypes.HPCAccountingRecord, 0),
	}
}

// Start starts the accounting service
func (s *HPCAccountingService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	// Start snapshot loop
	s.wg.Add(1)
	go s.snapshotLoop(ctx)

	// Start submission loop
	s.wg.Add(1)
	go s.submissionLoop(ctx)

	// Register lifecycle callback for job events
	s.scheduler.RegisterLifecycleCallback(s.onJobLifecycleEvent)

	return nil
}

// Stop stops the accounting service
func (s *HPCAccountingService) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
	return nil
}

// snapshotLoop periodically captures usage snapshots
func (s *HPCAccountingService) snapshotLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.SnapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.captureSnapshots(ctx); err != nil {
				// Log error but continue
				fmt.Printf("failed to capture snapshots: %v\n", err)
			}
		}
	}
}

// submissionLoop periodically submits pending records
func (s *HPCAccountingService) submissionLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.submitPendingRecords(ctx); err != nil {
				fmt.Printf("failed to submit records: %v\n", err)
			}
		}
	}
}

// captureSnapshots captures usage snapshots for all active jobs
func (s *HPCAccountingService) captureSnapshots(ctx context.Context) error {
	activeJobs, err := s.scheduler.ListActiveJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active jobs: %w", err)
	}

	for _, job := range activeJobs {
		if err := s.captureJobSnapshot(ctx, job); err != nil {
			fmt.Printf("failed to capture snapshot for job %s: %v\n", job.VirtEngineJobID, err)
		}
	}

	return nil
}

// captureJobSnapshot captures a usage snapshot for a single job
func (s *HPCAccountingService) captureJobSnapshot(ctx context.Context, job *HPCSchedulerJob) error {
	// Get current accounting data from scheduler
	metrics, err := s.scheduler.GetJobAccounting(ctx, job.VirtEngineJobID)
	if err != nil {
		return err
	}

	if metrics == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get or create aggregator
	aggregator, exists := s.jobAggregators[job.VirtEngineJobID]
	if !exists {
		customerAddr := ""
		if job.OriginalJob != nil {
			customerAddr = job.OriginalJob.CustomerAddress
		}
		aggregator = NewHPCUsageAggregator(
			job.VirtEngineJobID,
			s.clusterID,
			customerAddr,
			s.signer.GetProviderAddress(),
		)
		s.jobAggregators[job.VirtEngineJobID] = aggregator
	}

	// Get previous cumulative metrics
	prevMetrics := aggregator.GetAggregatedMetrics()

	// Add new metrics
	aggregator.AddMetrics(metrics)

	// Increment snapshot counter
	s.snapshotCounter[job.VirtEngineJobID]++
	seqNum := s.snapshotCounter[job.VirtEngineJobID]

	// Create detailed metrics
	detailedMetrics := s.convertToDetailedMetrics(metrics, job)

	// Calculate delta
	deltaMetrics := s.calculateDeltaMetrics(prevMetrics, metrics)

	// Create snapshot
	snapshot := &hpctypes.HPCUsageSnapshot{
		JobID:             job.VirtEngineJobID,
		ClusterID:         s.clusterID,
		SchedulerType:     string(s.schedulerType),
		SchedulerJobID:    job.SchedulerJobID,
		SnapshotType:      hpctypes.SnapshotTypeInterim,
		SequenceNumber:    seqNum,
		ProviderAddress:   s.signer.GetProviderAddress(),
		CustomerAddress:   aggregator.CustomerAddress,
		Metrics:           detailedMetrics,
		CumulativeMetrics: s.convertToDetailedMetrics(aggregator.GetAggregatedMetrics(), job),
		DeltaMetrics:      s.convertToDetailedMetrics(&deltaMetrics, job),
		JobState:          s.convertJobState(job.State),
		SnapshotTime:      time.Now(),
	}

	// Set previous snapshot ID
	if seqNum > 1 {
		snapshot.PreviousSnapshotID = fmt.Sprintf("snap-%s-%d", job.VirtEngineJobID, seqNum-1)
	}

	// Sign the snapshot
	hash := s.hashSnapshot(snapshot)
	sig, err := s.signer.Sign(hash)
	if err != nil {
		return fmt.Errorf("failed to sign snapshot: %w", err)
	}
	snapshot.ProviderSignature = hex.EncodeToString(sig)
	snapshot.ContentHash = hex.EncodeToString(hash)

	// Submit snapshot
	if err := s.submitter.SubmitUsageSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("failed to submit snapshot: %w", err)
	}

	return nil
}

// onJobLifecycleEvent handles job lifecycle events
func (s *HPCAccountingService) onJobLifecycleEvent(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
	ctx := context.Background()

	switch event {
	case HPCJobEventCompleted, HPCJobEventFailed, HPCJobEventCancelled, HPCJobEventTimeout:
		// Capture final snapshot and create accounting record
		if err := s.finalizeJobAccounting(ctx, job); err != nil {
			fmt.Printf("failed to finalize accounting for job %s: %v\n", job.VirtEngineJobID, err)
		}
	}
}

// finalizeJobAccounting creates the final accounting record for a completed job
func (s *HPCAccountingService) finalizeJobAccounting(ctx context.Context, job *HPCSchedulerJob) error {
	// Get final accounting data
	metrics, err := s.scheduler.GetJobAccounting(ctx, job.VirtEngineJobID)
	if err != nil {
		return err
	}

	s.mu.Lock()

	// Get aggregator
	aggregator, exists := s.jobAggregators[job.VirtEngineJobID]
	if exists && metrics != nil {
		aggregator.AddMetrics(metrics)
	}

	// Increment snapshot counter for final
	s.snapshotCounter[job.VirtEngineJobID]++
	seqNum := s.snapshotCounter[job.VirtEngineJobID]

	s.mu.Unlock()

	// Create final snapshot
	if aggregator != nil {
		finalMetrics := aggregator.GetAggregatedMetrics()
		detailedMetrics := s.convertToDetailedMetrics(finalMetrics, job)
		periodStart, periodEnd := aggregator.GetPeriod()

		snapshot := &hpctypes.HPCUsageSnapshot{
			JobID:             job.VirtEngineJobID,
			ClusterID:         s.clusterID,
			SchedulerType:     string(s.schedulerType),
			SchedulerJobID:    job.SchedulerJobID,
			SnapshotType:      hpctypes.SnapshotTypeFinal,
			SequenceNumber:    seqNum,
			ProviderAddress:   s.signer.GetProviderAddress(),
			CustomerAddress:   aggregator.CustomerAddress,
			Metrics:           detailedMetrics,
			CumulativeMetrics: detailedMetrics,
			JobState:          s.convertJobState(job.State),
			SnapshotTime:      time.Now(),
		}

		// Sign the snapshot
		hash := s.hashSnapshot(snapshot)
		sig, _ := s.signer.Sign(hash)
		snapshot.ProviderSignature = hex.EncodeToString(sig)
		snapshot.ContentHash = hex.EncodeToString(hash)

		// Submit final snapshot
		s.submitter.SubmitUsageSnapshot(ctx, snapshot)

		// Create accounting record
		record, err := s.createAccountingRecord(ctx, job, aggregator, periodStart, periodEnd, detailedMetrics)
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.pendingRecords = append(s.pendingRecords, record)
		s.mu.Unlock()
	}

	// Cleanup aggregator
	s.mu.Lock()
	delete(s.jobAggregators, job.VirtEngineJobID)
	delete(s.snapshotCounter, job.VirtEngineJobID)
	s.mu.Unlock()

	return nil
}

// createAccountingRecord creates an accounting record from aggregated metrics
func (s *HPCAccountingService) createAccountingRecord(
	ctx context.Context,
	job *HPCSchedulerJob,
	aggregator *HPCUsageAggregator,
	periodStart, periodEnd time.Time,
	metrics hpctypes.HPCDetailedMetrics,
) (*hpctypes.HPCAccountingRecord, error) {
	// Get billing rules
	rules, err := s.submitter.GetBillingRules(ctx, s.signer.GetProviderAddress())
	if err != nil {
		// Use defaults
		rules = &hpctypes.HPCBillingRules{}
		*rules = hpctypes.DefaultHPCBillingRules("uvirt")
	}

	// Calculate billing
	calculator := hpctypes.NewHPCBillingCalculator(*rules)
	breakdown, billable, err := calculator.CalculateBillableAmount(&metrics, nil, nil)
	if err != nil {
		return nil, err
	}

	providerReward := calculator.CalculateProviderReward(billable)
	platformFee := calculator.CalculatePlatformFee(billable)

	offeringID := ""
	if job.OriginalJob != nil {
		offeringID = job.OriginalJob.OfferingID
	}

	record := &hpctypes.HPCAccountingRecord{
		JobID:             job.VirtEngineJobID,
		ClusterID:         s.clusterID,
		ProviderAddress:   s.signer.GetProviderAddress(),
		CustomerAddress:   aggregator.CustomerAddress,
		OfferingID:        offeringID,
		SchedulerType:     string(s.schedulerType),
		SchedulerJobID:    job.SchedulerJobID,
		UsageMetrics:      metrics,
		BillableAmount:    billable,
		BillableBreakdown: *breakdown,
		ProviderReward:    providerReward,
		PlatformFee:       platformFee,
		Status:            hpctypes.AccountingStatusPending,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		FormulaVersion:    rules.FormulaVersion,
	}

	return record, nil
}

// submitPendingRecords submits pending accounting records
func (s *HPCAccountingService) submitPendingRecords(ctx context.Context) error {
	s.mu.Lock()
	if len(s.pendingRecords) == 0 {
		s.mu.Unlock()
		return nil
	}

	// Get batch
	batchSize := s.config.SubmissionBatchSize
	if batchSize > len(s.pendingRecords) {
		batchSize = len(s.pendingRecords)
	}
	batch := s.pendingRecords[:batchSize]
	s.pendingRecords = s.pendingRecords[batchSize:]
	s.mu.Unlock()

	// Submit batch
	var failedRecords []*hpctypes.HPCAccountingRecord
	for _, record := range batch {
		if err := s.submitter.SubmitAccountingRecord(ctx, record); err != nil {
			fmt.Printf("failed to submit accounting record: %v\n", err)
			failedRecords = append(failedRecords, record)
		}
	}

	// Re-queue failed records
	if len(failedRecords) > 0 {
		s.mu.Lock()
		s.pendingRecords = append(failedRecords, s.pendingRecords...)
		s.mu.Unlock()
	}

	return nil
}

// convertToDetailedMetrics converts HPCSchedulerMetrics to HPCDetailedMetrics
func (s *HPCAccountingService) convertToDetailedMetrics(metrics *HPCSchedulerMetrics, job *HPCSchedulerJob) hpctypes.HPCDetailedMetrics {
	if metrics == nil {
		return hpctypes.HPCDetailedMetrics{}
	}

	detailed := hpctypes.HPCDetailedMetrics{
		WallClockSeconds: metrics.WallClockSeconds,
		CPUCoreSeconds:   metrics.CPUCoreSeconds,
		CPUTimeSeconds:   metrics.CPUTimeSeconds,
		MemoryGBSeconds:  metrics.MemoryGBSeconds,
		MemoryBytesMax:   metrics.MemoryBytesMax,
		GPUSeconds:       metrics.GPUSeconds,
		StorageGBHours:   metrics.StorageGBHours,
		NetworkBytesIn:   metrics.NetworkBytesIn,
		NetworkBytesOut:  metrics.NetworkBytesOut,
		NodeHours:        sdkmath.LegacyNewDec(int64(metrics.NodeHours * 1000000)).Quo(sdkmath.LegacyNewDec(1000000)),
		NodesUsed:        metrics.NodesUsed,
		EnergyJoules:     metrics.EnergyJoules,
		SubmitTime:       job.SubmitTime,
	}

	if job.StartTime != nil {
		detailed.StartTime = job.StartTime
		// Calculate queue time
		detailed.QueueTimeSeconds = int64(job.StartTime.Sub(job.SubmitTime).Seconds())
	}

	if job.EndTime != nil {
		detailed.EndTime = job.EndTime
	}

	return detailed
}

// calculateDeltaMetrics calculates the delta between two metric snapshots
func (s *HPCAccountingService) calculateDeltaMetrics(prev, curr *HPCSchedulerMetrics) HPCSchedulerMetrics {
	if prev == nil || curr == nil {
		if curr != nil {
			return *curr
		}
		return HPCSchedulerMetrics{}
	}

	return HPCSchedulerMetrics{
		WallClockSeconds: curr.WallClockSeconds - prev.WallClockSeconds,
		CPUCoreSeconds:   curr.CPUCoreSeconds - prev.CPUCoreSeconds,
		CPUTimeSeconds:   curr.CPUTimeSeconds - prev.CPUTimeSeconds,
		MemoryGBSeconds:  curr.MemoryGBSeconds - prev.MemoryGBSeconds,
		GPUSeconds:       curr.GPUSeconds - prev.GPUSeconds,
		StorageGBHours:   curr.StorageGBHours - prev.StorageGBHours,
		NetworkBytesIn:   curr.NetworkBytesIn - prev.NetworkBytesIn,
		NetworkBytesOut:  curr.NetworkBytesOut - prev.NetworkBytesOut,
	}
}

// convertJobState converts HPCJobState to hpctypes.JobState
func (s *HPCAccountingService) convertJobState(state HPCJobState) hpctypes.JobState {
	switch state {
	case HPCJobStatePending:
		return hpctypes.JobStatePending
	case HPCJobStateQueued:
		return hpctypes.JobStateQueued
	case HPCJobStateRunning:
		return hpctypes.JobStateRunning
	case HPCJobStateCompleted:
		return hpctypes.JobStateCompleted
	case HPCJobStateFailed:
		return hpctypes.JobStateFailed
	case HPCJobStateCancelled:
		return hpctypes.JobStateCancelled
	case HPCJobStateTimeout:
		return hpctypes.JobStateTimeout
	default:
		return hpctypes.JobStatePending
	}
}

// hashSnapshot creates a hash of the snapshot for signing
func (s *HPCAccountingService) hashSnapshot(snapshot *hpctypes.HPCUsageSnapshot) []byte {
	data := struct {
		JobID          string `json:"job_id"`
		ClusterID      string `json:"cluster_id"`
		SequenceNumber uint32 `json:"sequence_number"`
		SnapshotType   string `json:"snapshot_type"`
		SnapshotTime   int64  `json:"snapshot_time"`
	}{
		JobID:          snapshot.JobID,
		ClusterID:      snapshot.ClusterID,
		SequenceNumber: snapshot.SequenceNumber,
		SnapshotType:   string(snapshot.SnapshotType),
		SnapshotTime:   snapshot.SnapshotTime.Unix(),
	}

	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// GetPendingRecordCount returns the number of pending records
func (s *HPCAccountingService) GetPendingRecordCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pendingRecords)
}

// GetActiveJobCount returns the number of jobs being tracked
func (s *HPCAccountingService) GetActiveJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobAggregators)
}
