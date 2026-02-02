// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-5A: HPC Reconciliation Service - compares scheduler logs with on-chain records
package provider_daemon

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCReconciliationConfig configures the reconciliation service
type HPCReconciliationConfig struct {
	// ReconciliationInterval is how often to run reconciliation
	ReconciliationInterval time.Duration `json:"reconciliation_interval"`

	// JobLookbackDuration is how far back to look for jobs to reconcile
	JobLookbackDuration time.Duration `json:"job_lookback_duration"`

	// Tolerances defines acceptable discrepancy tolerances
	Tolerances hpctypes.ReconciliationTolerances `json:"tolerances"`

	// AutoResolveMinorDiscrepancies automatically resolves minor discrepancies
	AutoResolveMinorDiscrepancies bool `json:"auto_resolve_minor_discrepancies"`

	// MinorDiscrepancyThresholdPercent is the threshold for minor discrepancies
	MinorDiscrepancyThresholdPercent float64 `json:"minor_discrepancy_threshold_percent"`

	// AlertOnCriticalDiscrepancy sends alerts for critical discrepancies
	AlertOnCriticalDiscrepancy bool `json:"alert_on_critical_discrepancy"`
}

// DefaultHPCReconciliationConfig returns default reconciliation configuration
func DefaultHPCReconciliationConfig() HPCReconciliationConfig {
	return HPCReconciliationConfig{
		ReconciliationInterval:           1 * time.Hour,
		JobLookbackDuration:              24 * time.Hour,
		Tolerances:                       hpctypes.DefaultReconciliationTolerances(),
		AutoResolveMinorDiscrepancies:    true,
		MinorDiscrepancyThresholdPercent: 1.0,
		AlertOnCriticalDiscrepancy:       true,
	}
}

// SchedulerAccountingSource provides scheduler accounting data
type SchedulerAccountingSource interface {
	// GetJobAccountingData gets accounting data for a job from the scheduler
	GetJobAccountingData(ctx context.Context, jobID string) (*SchedulerJobAccounting, error)

	// ListCompletedJobs lists completed jobs in a time range
	ListCompletedJobs(ctx context.Context, start, end time.Time) ([]string, error)

	// GetRawAccountingLogs gets raw accounting logs for a job
	GetRawAccountingLogs(ctx context.Context, jobID string) (string, error)
}

// OnChainAccountingSource provides on-chain accounting data
type OnChainAccountingSource interface {
	// GetAccountingRecord gets an accounting record from the chain
	GetAccountingRecord(ctx context.Context, jobID string) (*hpctypes.HPCAccountingRecord, error)

	// GetUsageSnapshots gets usage snapshots for a job
	GetUsageSnapshots(ctx context.Context, jobID string) ([]hpctypes.HPCUsageSnapshot, error)

	// SubmitReconciliationRecord submits a reconciliation record to the chain
	SubmitReconciliationRecord(ctx context.Context, record *hpctypes.HPCReconciliationRecord) error

	// CreateDispute creates a dispute for a discrepancy
	CreateDispute(ctx context.Context, jobID, reason, evidence string) error
}

// SchedulerJobAccounting contains accounting data from the scheduler
type SchedulerJobAccounting struct {
	JobID            string    `json:"job_id"`
	SchedulerJobID   string    `json:"scheduler_job_id"`
	SubmitTime       time.Time `json:"submit_time"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	WallClockSeconds int64     `json:"wall_clock_seconds"`
	CPUCoreSeconds   int64     `json:"cpu_core_seconds"`
	MemoryGBSeconds  int64     `json:"memory_gb_seconds"`
	GPUSeconds       int64     `json:"gpu_seconds"`
	NodesUsed        int32     `json:"nodes_used"`
	ExitCode         int32     `json:"exit_code"`
	RawData          string    `json:"raw_data"`
}

// HPCReconciliationService reconciles scheduler and on-chain accounting
type HPCReconciliationService struct {
	config          HPCReconciliationConfig
	clusterID       string
	schedulerSource SchedulerAccountingSource
	onChainSource   OnChainAccountingSource

	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
	wg          sync.WaitGroup
	lastRunTime time.Time
	stats       ReconciliationStats
}

// ReconciliationStats tracks reconciliation statistics
type ReconciliationStats struct {
	TotalReconciled       int64         `json:"total_reconciled"`
	TotalMatched          int64         `json:"total_matched"`
	TotalDiscrepancies    int64         `json:"total_discrepancies"`
	CriticalDiscrepancies int64         `json:"critical_discrepancies"`
	AutoResolved          int64         `json:"auto_resolved"`
	LastRunTime           time.Time     `json:"last_run_time"`
	LastRunDuration       time.Duration `json:"last_run_duration"`
}

// NewHPCReconciliationService creates a new reconciliation service
func NewHPCReconciliationService(
	config HPCReconciliationConfig,
	clusterID string,
	schedulerSource SchedulerAccountingSource,
	onChainSource OnChainAccountingSource,
) *HPCReconciliationService {
	return &HPCReconciliationService{
		config:          config,
		clusterID:       clusterID,
		schedulerSource: schedulerSource,
		onChainSource:   onChainSource,
		stopCh:          make(chan struct{}),
	}
}

// Start starts the reconciliation service
func (s *HPCReconciliationService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.reconciliationLoop(ctx)

	return nil
}

// Stop stops the reconciliation service
func (s *HPCReconciliationService) Stop() error {
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

// reconciliationLoop runs periodic reconciliation
func (s *HPCReconciliationService) reconciliationLoop(ctx context.Context) {
	defer s.wg.Done()

	// Run immediately on start
	s.runReconciliation(ctx)

	ticker := time.NewTicker(s.config.ReconciliationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.runReconciliation(ctx)
		}
	}
}

// runReconciliation performs a reconciliation run
func (s *HPCReconciliationService) runReconciliation(ctx context.Context) {
	startTime := time.Now()

	s.mu.Lock()
	s.lastRunTime = startTime
	s.mu.Unlock()

	// Get completed jobs in lookback window
	end := time.Now()
	start := end.Add(-s.config.JobLookbackDuration)

	jobIDs, err := s.schedulerSource.ListCompletedJobs(ctx, start, end)
	if err != nil {
		fmt.Printf("failed to list completed jobs: %v\n", err)
		return
	}

	var matched, discrepancies, critical, autoResolved int64

	for _, jobID := range jobIDs {
		result, err := s.reconcileJob(ctx, jobID)
		if err != nil {
			fmt.Printf("failed to reconcile job %s: %v\n", jobID, err)
			continue
		}

		switch result.Status {
		case hpctypes.ReconciliationStatusMatched:
			matched++
		case hpctypes.ReconciliationStatusDiscrepancy:
			discrepancies++
			if result.HasCriticalDiscrepancies() {
				critical++
			} else if s.config.AutoResolveMinorDiscrepancies {
				autoResolved++
			}
		case hpctypes.ReconciliationStatusResolved:
			autoResolved++
		}
	}

	duration := time.Since(startTime)

	s.mu.Lock()
	s.stats.TotalReconciled += int64(len(jobIDs))
	s.stats.TotalMatched += matched
	s.stats.TotalDiscrepancies += discrepancies
	s.stats.CriticalDiscrepancies += critical
	s.stats.AutoResolved += autoResolved
	s.stats.LastRunTime = startTime
	s.stats.LastRunDuration = duration
	s.mu.Unlock()

	fmt.Printf("Reconciliation complete: %d jobs, %d matched, %d discrepancies, %d critical\n",
		len(jobIDs), matched, discrepancies, critical)
}

// reconcileJob reconciles a single job
func (s *HPCReconciliationService) reconcileJob(ctx context.Context, jobID string) (*hpctypes.HPCReconciliationRecord, error) {
	// Get scheduler data
	schedulerData, err := s.schedulerSource.GetJobAccountingData(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduler data: %w", err)
	}

	// Get on-chain data
	onChainRecord, err := s.onChainSource.GetAccountingRecord(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get on-chain data: %w", err)
	}

	// Build reconciliation record
	record := &hpctypes.HPCReconciliationRecord{
		JobID:              jobID,
		ClusterID:          s.clusterID,
		ProviderAddress:    onChainRecord.ProviderAddress,
		ReconciliationTime: time.Now(),
		SchedulerSource: hpctypes.ReconciliationSource{
			SourceType:  "scheduler",
			SourceID:    schedulerData.SchedulerJobID,
			ExtractTime: time.Now(),
			Metrics:     s.convertSchedulerToDetailedMetrics(schedulerData),
		},
		OnChainSource: hpctypes.ReconciliationSource{
			SourceType:  "on_chain",
			SourceID:    onChainRecord.RecordID,
			ExtractTime: onChainRecord.CreatedAt,
			Metrics:     onChainRecord.UsageMetrics,
		},
	}

	// Compare metrics and find discrepancies
	discrepancies := s.compareMetrics(schedulerData, onChainRecord)
	record.Discrepancies = discrepancies

	if len(discrepancies) == 0 {
		record.Status = hpctypes.ReconciliationStatusMatched
	} else {
		record.Status = hpctypes.ReconciliationStatusDiscrepancy

		// Check if we should auto-resolve
		if s.config.AutoResolveMinorDiscrepancies && !record.HasDiscrepancies() {
			record.Status = hpctypes.ReconciliationStatusResolved
			record.Resolution = "Auto-resolved: all discrepancies within tolerance"
			record.ResolutionAction = "none"
		} else if record.HasDiscrepancies() {
			// Check for critical discrepancies
			criticalDiscrepancies := record.GetCriticalDiscrepancies()
			if len(criticalDiscrepancies) > 0 && s.config.AlertOnCriticalDiscrepancy {
				// Create dispute
				evidence, _ := s.schedulerSource.GetRawAccountingLogs(ctx, jobID)
				reason := fmt.Sprintf("Critical discrepancy detected in %d fields", len(criticalDiscrepancies))
				if err := s.onChainSource.CreateDispute(ctx, jobID, reason, evidence); err != nil {
					fmt.Printf("failed to create dispute: %v\n", err)
				}
			}
		}
	}

	// Submit reconciliation record
	if err := s.onChainSource.SubmitReconciliationRecord(ctx, record); err != nil {
		return record, fmt.Errorf("failed to submit reconciliation: %w", err)
	}

	return record, nil
}

// compareMetrics compares scheduler and on-chain metrics
func (s *HPCReconciliationService) compareMetrics(
	scheduler *SchedulerJobAccounting,
	onChain *hpctypes.HPCAccountingRecord,
) []hpctypes.ReconciliationDiscrepancy {
	var discrepancies []hpctypes.ReconciliationDiscrepancy
	tolerances := s.config.Tolerances

	// Compare wall clock seconds
	if d := s.checkDiscrepancy(
		"wall_clock_seconds",
		float64(scheduler.WallClockSeconds),
		float64(onChain.UsageMetrics.WallClockSeconds),
		tolerances.WallClockSecondsPercent,
	); d != nil {
		discrepancies = append(discrepancies, *d)
	}

	// Compare CPU core seconds
	if d := s.checkDiscrepancy(
		"cpu_core_seconds",
		float64(scheduler.CPUCoreSeconds),
		float64(onChain.UsageMetrics.CPUCoreSeconds),
		tolerances.CPUCoreSecondsPercent,
	); d != nil {
		discrepancies = append(discrepancies, *d)
	}

	// Compare memory GB seconds
	if d := s.checkDiscrepancy(
		"memory_gb_seconds",
		float64(scheduler.MemoryGBSeconds),
		float64(onChain.UsageMetrics.MemoryGBSeconds),
		tolerances.MemoryGBSecondsPercent,
	); d != nil {
		discrepancies = append(discrepancies, *d)
	}

	// Compare GPU seconds
	if scheduler.GPUSeconds > 0 || onChain.UsageMetrics.GPUSeconds > 0 {
		if d := s.checkDiscrepancy(
			"gpu_seconds",
			float64(scheduler.GPUSeconds),
			float64(onChain.UsageMetrics.GPUSeconds),
			tolerances.GPUSecondsPercent,
		); d != nil {
			discrepancies = append(discrepancies, *d)
		}
	}

	// Compare nodes used
	if scheduler.NodesUsed != onChain.UsageMetrics.NodesUsed {
		discrepancies = append(discrepancies, hpctypes.ReconciliationDiscrepancy{
			Field:             "nodes_used",
			SchedulerValue:    fmt.Sprintf("%d", scheduler.NodesUsed),
			OnChainValue:      fmt.Sprintf("%d", onChain.UsageMetrics.NodesUsed),
			DifferencePercent: "N/A",
			Severity:          "high",
			ToleranceExceeded: true,
		})
	}

	return discrepancies
}

// checkDiscrepancy checks for a discrepancy between two values
func (s *HPCReconciliationService) checkDiscrepancy(
	field string,
	schedulerValue, onChainValue float64,
	tolerancePercent float64,
) *hpctypes.ReconciliationDiscrepancy {
	if schedulerValue == 0 && onChainValue == 0 {
		return nil
	}

	// Calculate percentage difference
	var diffPercent float64
	if schedulerValue == 0 {
		diffPercent = 100.0
	} else {
		diffPercent = math.Abs((onChainValue-schedulerValue)/schedulerValue) * 100
	}

	if diffPercent <= tolerancePercent {
		return nil
	}

	// Determine severity
	severity := "low"
	if diffPercent > tolerancePercent*2 {
		severity = "medium"
	}
	if diffPercent > tolerancePercent*5 {
		severity = "high"
	}
	if diffPercent > tolerancePercent*10 {
		severity = "critical"
	}

	return &hpctypes.ReconciliationDiscrepancy{
		Field:             field,
		SchedulerValue:    fmt.Sprintf("%.2f", schedulerValue),
		OnChainValue:      fmt.Sprintf("%.2f", onChainValue),
		DifferencePercent: fmt.Sprintf("%.2f%%", diffPercent),
		Severity:          severity,
		ToleranceExceeded: true,
	}
}

// convertSchedulerToDetailedMetrics converts scheduler data to detailed metrics
func (s *HPCReconciliationService) convertSchedulerToDetailedMetrics(data *SchedulerJobAccounting) hpctypes.HPCDetailedMetrics {
	return hpctypes.HPCDetailedMetrics{
		WallClockSeconds: data.WallClockSeconds,
		CPUCoreSeconds:   data.CPUCoreSeconds,
		MemoryGBSeconds:  data.MemoryGBSeconds,
		GPUSeconds:       data.GPUSeconds,
		NodesUsed:        data.NodesUsed,
		SubmitTime:       data.SubmitTime,
		StartTime:        &data.StartTime,
		EndTime:          &data.EndTime,
	}
}

// GetStats returns reconciliation statistics
func (s *HPCReconciliationService) GetStats() ReconciliationStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

// RunManualReconciliation triggers a manual reconciliation run
func (s *HPCReconciliationService) RunManualReconciliation(ctx context.Context) error {
	s.runReconciliation(ctx)
	return nil
}

// ReconcileSingleJob reconciles a single job on demand
func (s *HPCReconciliationService) ReconcileSingleJob(ctx context.Context, jobID string) (*hpctypes.HPCReconciliationRecord, error) {
	return s.reconcileJob(ctx, jobID)
}
