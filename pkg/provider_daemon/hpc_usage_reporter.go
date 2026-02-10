// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: HPC Usage Reporter - reports job usage metrics on-chain
package provider_daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCUsageRecord represents a signed usage record for on-chain submission
type HPCUsageRecord struct {
	// RecordID is a unique identifier for this record
	RecordID string `json:"record_id"`

	// JobID is the VirtEngine job ID
	JobID string `json:"job_id"`

	// ClusterID is the HPC cluster ID
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer address
	CustomerAddress string `json:"customer_address"`

	// PeriodStart is the start of the usage period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the usage period
	PeriodEnd time.Time `json:"period_end"`

	// Metrics contains the usage metrics
	Metrics *HPCSchedulerMetrics `json:"metrics"`

	// IsFinal indicates if this is the final usage record
	IsFinal bool `json:"is_final"`

	// JobState is the current job state
	JobState HPCJobState `json:"job_state"`

	// Timestamp is when the record was created
	Timestamp time.Time `json:"timestamp"`

	// Signature is the provider's signature (hex encoded)
	Signature string `json:"signature"`
}

// Hash generates a hash of the usage record for signing
func (r *HPCUsageRecord) Hash() []byte {
	data := struct {
		RecordID        string `json:"record_id"`
		JobID           string `json:"job_id"`
		ClusterID       string `json:"cluster_id"`
		ProviderAddress string `json:"provider_address"`
		PeriodStart     int64  `json:"period_start"`
		PeriodEnd       int64  `json:"period_end"`
		IsFinal         bool   `json:"is_final"`
		Timestamp       int64  `json:"timestamp"`
	}{
		RecordID:        r.RecordID,
		JobID:           r.JobID,
		ClusterID:       r.ClusterID,
		ProviderAddress: r.ProviderAddress,
		PeriodStart:     r.PeriodStart.Unix(),
		PeriodEnd:       r.PeriodEnd.Unix(),
		IsFinal:         r.IsFinal,
		Timestamp:       r.Timestamp.Unix(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	hash := sha256.Sum256(bytes)
	return hash[:]
}

// ToChainAccounting converts to x/hpc job accounting
func (r *HPCUsageRecord) ToChainAccounting() *hpctypes.JobAccounting {
	return &hpctypes.JobAccounting{
		JobID:                r.JobID,
		ClusterID:            r.ClusterID,
		ProviderAddress:      r.ProviderAddress,
		CustomerAddress:      r.CustomerAddress,
		UsageMetrics:         r.Metrics.ToChainMetrics(),
		SignedUsageRecordIDs: []string{r.RecordID},
		JobCompletionStatus:  r.JobState.ToChainState(),
		CreatedAt:            r.Timestamp,
	}
}

// HPCUsageReporter collects and reports job usage metrics
type HPCUsageReporter struct {
	config    HPCUsageReportingConfig
	clusterID string
	signer    HPCSchedulerSigner

	mu             sync.RWMutex
	running        bool
	stopCh         chan struct{}
	wg             sync.WaitGroup
	pendingRecords []*HPCUsageRecord
	lastReportTime map[string]time.Time // job ID -> last report time
	recordCounter  uint64
}

// NewHPCUsageReporter creates a new usage reporter
func NewHPCUsageReporter(
	config HPCUsageReportingConfig,
	clusterID string,
	signer HPCSchedulerSigner,
) *HPCUsageReporter {
	return &HPCUsageReporter{
		config:         config,
		clusterID:      clusterID,
		signer:         signer,
		stopCh:         make(chan struct{}),
		pendingRecords: make([]*HPCUsageRecord, 0),
		lastReportTime: make(map[string]time.Time),
	}
}

// Start starts the usage reporter
func (r *HPCUsageReporter) Start() error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = true
	r.stopCh = make(chan struct{})
	r.mu.Unlock()

	return nil
}

// Stop stops the usage reporter
func (r *HPCUsageReporter) Stop() error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = false
	close(r.stopCh)
	r.mu.Unlock()

	r.wg.Wait()
	return nil
}

// CreateUsageRecord creates a signed usage record for a job
func (r *HPCUsageReporter) CreateUsageRecord(
	job *HPCSchedulerJob,
	customerAddress string,
	periodStart, periodEnd time.Time,
	isFinal bool,
) (*HPCUsageRecord, error) {
	r.mu.Lock()
	r.recordCounter++
	recordID := fmt.Sprintf("%s-%s-%d", r.clusterID, job.VirtEngineJobID, r.recordCounter)
	r.mu.Unlock()

	record := &HPCUsageRecord{
		RecordID:        recordID,
		JobID:           job.VirtEngineJobID,
		ClusterID:       r.clusterID,
		ProviderAddress: r.signer.GetProviderAddress(),
		CustomerAddress: customerAddress,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		Metrics:         job.Metrics,
		IsFinal:         isFinal,
		JobState:        job.State,
		Timestamp:       time.Now(),
	}

	// Sign the record
	hash := record.Hash()
	sig, err := r.signer.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign usage record: %w", err)
	}
	record.Signature = hex.EncodeToString(sig)

	return record, nil
}

// QueueRecord queues a usage record for batch submission
func (r *HPCUsageReporter) QueueRecord(record *HPCUsageRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pendingRecords = append(r.pendingRecords, record)
	r.lastReportTime[record.JobID] = record.Timestamp
}

// GetPendingRecords returns pending records up to the batch size
func (r *HPCUsageReporter) GetPendingRecords() []*HPCUsageRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := len(r.pendingRecords)
	if count > r.config.BatchSize {
		count = r.config.BatchSize
	}

	records := make([]*HPCUsageRecord, count)
	copy(records, r.pendingRecords[:count])
	return records
}

// AcknowledgeRecords removes acknowledged records from the queue
func (r *HPCUsageReporter) AcknowledgeRecords(recordIDs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ackSet := make(map[string]bool)
	for _, id := range recordIDs {
		ackSet[id] = true
	}

	remaining := make([]*HPCUsageRecord, 0, len(r.pendingRecords))
	for _, record := range r.pendingRecords {
		if !ackSet[record.RecordID] {
			remaining = append(remaining, record)
		}
	}
	r.pendingRecords = remaining
}

// ShouldReportUsage checks if it's time to report usage for a job
func (r *HPCUsageReporter) ShouldReportUsage(jobID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lastReport, exists := r.lastReportTime[jobID]
	if !exists {
		return true
	}

	return time.Since(lastReport) >= r.config.ReportInterval
}

// GetPendingCount returns the number of pending records
func (r *HPCUsageReporter) GetPendingCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.pendingRecords)
}

// HPCUsageAggregator aggregates usage metrics across multiple records
type HPCUsageAggregator struct {
	JobID           string
	ClusterID       string
	CustomerAddress string
	ProviderAddress string

	mu           sync.Mutex
	startTime    time.Time
	lastUpdate   time.Time
	totalMetrics HPCSchedulerMetrics
	recordCount  int
	peakMetrics  HPCSchedulerMetrics
}

// NewHPCUsageAggregator creates a new usage aggregator
func NewHPCUsageAggregator(jobID, clusterID, customerAddress, providerAddress string) *HPCUsageAggregator {
	now := time.Now()
	return &HPCUsageAggregator{
		JobID:           jobID,
		ClusterID:       clusterID,
		CustomerAddress: customerAddress,
		ProviderAddress: providerAddress,
		startTime:       now,
		lastUpdate:      now,
	}
}

// AddMetrics adds a metrics sample to the aggregator
func (a *HPCUsageAggregator) AddMetrics(metrics *HPCSchedulerMetrics) {
	if metrics == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.recordCount++
	a.lastUpdate = time.Now()

	// Aggregate cumulative metrics
	a.totalMetrics.WallClockSeconds = metrics.WallClockSeconds // Wall clock is absolute
	a.totalMetrics.CPUTimeSeconds += metrics.CPUTimeSeconds
	a.totalMetrics.CPUCoreSeconds += metrics.CPUCoreSeconds
	a.totalMetrics.MemoryGBSeconds += metrics.MemoryGBSeconds
	a.totalMetrics.GPUSeconds += metrics.GPUSeconds
	a.totalMetrics.StorageGBHours += metrics.StorageGBHours
	a.totalMetrics.NetworkBytesIn += metrics.NetworkBytesIn
	a.totalMetrics.NetworkBytesOut += metrics.NetworkBytesOut
	a.totalMetrics.EnergyJoules += metrics.EnergyJoules
	a.totalMetrics.NodeHours = metrics.NodeHours // Absolute value
	a.totalMetrics.NodesUsed = metrics.NodesUsed

	// Track peak values
	if metrics.MemoryBytesMax > a.peakMetrics.MemoryBytesMax {
		a.peakMetrics.MemoryBytesMax = metrics.MemoryBytesMax
	}
}

// GetAggregatedMetrics returns the aggregated metrics
func (a *HPCUsageAggregator) GetAggregatedMetrics() *HPCSchedulerMetrics {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := a.totalMetrics
	result.MemoryBytesMax = a.peakMetrics.MemoryBytesMax
	return &result
}

// GetPeriod returns the aggregation period
func (a *HPCUsageAggregator) GetPeriod() (start, end time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.startTime, a.lastUpdate
}

// Reset resets the aggregator for a new period
func (a *HPCUsageAggregator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.startTime = time.Now()
	a.lastUpdate = a.startTime
	a.totalMetrics = HPCSchedulerMetrics{}
	a.peakMetrics = HPCSchedulerMetrics{}
	a.recordCount = 0
}

// HPCBillingCalculator calculates billing based on usage metrics
type HPCBillingCalculator struct {
	// Rate per CPU core-hour
	CPUCoreHourRate float64

	// Rate per GPU-hour
	GPUHourRate float64

	// Rate per GB-hour of memory
	MemoryGBHourRate float64

	// Rate per node-hour
	NodeHourRate float64

	// Rate per GB of storage per hour
	StorageGBHourRate float64

	// Rate per GB of network transfer
	NetworkGBRate float64
}

// DefaultHPCBillingCalculator returns a calculator with default rates
func DefaultHPCBillingCalculator() *HPCBillingCalculator {
	return &HPCBillingCalculator{
		CPUCoreHourRate:   0.05,  // $0.05 per core-hour
		GPUHourRate:       1.00,  // $1.00 per GPU-hour
		MemoryGBHourRate:  0.01,  // $0.01 per GB-hour
		NodeHourRate:      0.50,  // $0.50 per node-hour
		StorageGBHourRate: 0.001, // $0.001 per GB-hour
		NetworkGBRate:     0.10,  // $0.10 per GB
	}
}

// CalculateCost calculates the cost for given usage metrics
func (c *HPCBillingCalculator) CalculateCost(metrics *HPCSchedulerMetrics) float64 {
	if metrics == nil {
		return 0
	}

	cost := 0.0

	// CPU cost (convert core-seconds to core-hours)
	cpuCoreHours := float64(metrics.CPUCoreSeconds) / 3600.0
	cost += cpuCoreHours * c.CPUCoreHourRate

	// GPU cost (convert GPU-seconds to GPU-hours)
	gpuHours := float64(metrics.GPUSeconds) / 3600.0
	cost += gpuHours * c.GPUHourRate

	// Memory cost (convert GB-seconds to GB-hours)
	memoryGBHours := float64(metrics.MemoryGBSeconds) / 3600.0
	cost += memoryGBHours * c.MemoryGBHourRate

	// Node cost
	cost += metrics.NodeHours * c.NodeHourRate

	// Storage cost
	cost += float64(metrics.StorageGBHours) * c.StorageGBHourRate

	// Network cost (convert bytes to GB)
	networkGB := float64(metrics.NetworkBytesIn+metrics.NetworkBytesOut) / (1024 * 1024 * 1024)
	cost += networkGB * c.NetworkGBRate

	return cost
}

// CalculateCostBreakdown returns a detailed cost breakdown
func (c *HPCBillingCalculator) CalculateCostBreakdown(metrics *HPCSchedulerMetrics) map[string]float64 {
	if metrics == nil {
		return nil
	}

	breakdown := make(map[string]float64)

	cpuCoreHours := float64(metrics.CPUCoreSeconds) / 3600.0
	breakdown["cpu"] = cpuCoreHours * c.CPUCoreHourRate

	gpuHours := float64(metrics.GPUSeconds) / 3600.0
	breakdown["gpu"] = gpuHours * c.GPUHourRate

	memoryGBHours := float64(metrics.MemoryGBSeconds) / 3600.0
	breakdown["memory"] = memoryGBHours * c.MemoryGBHourRate

	breakdown["node"] = metrics.NodeHours * c.NodeHourRate

	breakdown["storage"] = float64(metrics.StorageGBHours) * c.StorageGBHourRate

	networkGB := float64(metrics.NetworkBytesIn+metrics.NetworkBytesOut) / (1024 * 1024 * 1024)
	breakdown["network"] = networkGB * c.NetworkGBRate

	breakdown["total"] = breakdown["cpu"] + breakdown["gpu"] + breakdown["memory"] +
		breakdown["node"] + breakdown["storage"] + breakdown["network"]

	return breakdown
}
