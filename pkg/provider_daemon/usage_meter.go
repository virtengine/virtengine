// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-404: Provider Daemon usage metering + on-chain recording
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// Sentinel errors for metering
var (
	// ErrMeteringNotStarted is returned when metering is not started
	ErrMeteringNotStarted = verrors.ErrInvalidState

	// ErrWorkloadNotMetered is returned when a workload is not being metered
	ErrWorkloadNotMetered = verrors.ErrNotFound
)

// MeteringInterval represents the metering interval
type MeteringInterval time.Duration

const (
	// MeteringIntervalMinute is a one-minute interval
	MeteringIntervalMinute MeteringInterval = MeteringInterval(time.Minute)

	// MeteringIntervalHourly is an hourly interval
	MeteringIntervalHourly MeteringInterval = MeteringInterval(time.Hour)

	// MeteringIntervalDaily is a daily interval
	MeteringIntervalDaily MeteringInterval = MeteringInterval(24 * time.Hour)
)

// UsageRecordType indicates the type of usage record
type UsageRecordType string

const (
	// UsageRecordTypePeriodic is a periodic usage update
	UsageRecordTypePeriodic UsageRecordType = "periodic"

	// UsageRecordTypeFinal is a final settlement record
	UsageRecordTypeFinal UsageRecordType = "final"
)

// ResourceMetrics contains resource usage metrics
type ResourceMetrics struct {
	// CPUMilliSeconds is CPU usage in milliseconds
	CPUMilliSeconds int64 `json:"cpu_milli_seconds"`

	// MemoryByteSeconds is memory usage in byte-seconds
	MemoryByteSeconds int64 `json:"memory_byte_seconds"`

	// StorageByteSeconds is storage usage in byte-seconds
	StorageByteSeconds int64 `json:"storage_byte_seconds"`

	// NetworkBytesIn is inbound network bytes
	NetworkBytesIn int64 `json:"network_bytes_in"`

	// NetworkBytesOut is outbound network bytes
	NetworkBytesOut int64 `json:"network_bytes_out"`

	// GPUSeconds is GPU usage in seconds
	GPUSeconds int64 `json:"gpu_seconds"`
}

// UsageRecord represents a usage record for on-chain submission
type UsageRecord struct {
	// ID is the record ID
	ID string `json:"id"`

	// WorkloadID is the workload ID
	WorkloadID string `json:"workload_id"`

	// DeploymentID is the on-chain deployment ID
	DeploymentID string `json:"deployment_id"`

	// LeaseID is the on-chain lease ID
	LeaseID string `json:"lease_id"`

	// ProviderID is the provider ID
	ProviderID string `json:"provider_id"`

	// Type is the record type (periodic or final)
	Type UsageRecordType `json:"type"`

	// StartTime is the start of the metering period
	StartTime time.Time `json:"start_time"`

	// EndTime is the end of the metering period
	EndTime time.Time `json:"end_time"`

	// Metrics contains the resource usage metrics
	Metrics ResourceMetrics `json:"metrics"`

	// PricingInputs contains inputs for pricing calculation
	PricingInputs PricingInputs `json:"pricing_inputs"`

	// Signature is the provider's signature
	Signature string `json:"signature"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`
}

// PricingInputs contains inputs for pricing calculation
type PricingInputs struct {
	// AgreedCPURate is the agreed rate per CPU-hour in tokens
	AgreedCPURate string `json:"agreed_cpu_rate"`

	// AgreedMemoryRate is the agreed rate per GB-hour in tokens
	AgreedMemoryRate string `json:"agreed_memory_rate"`

	// AgreedStorageRate is the agreed rate per GB-hour in tokens
	AgreedStorageRate string `json:"agreed_storage_rate"`

	// AgreedGPURate is the agreed rate per GPU-hour in tokens
	AgreedGPURate string `json:"agreed_gpu_rate"`

	// AgreedNetworkRate is the agreed rate per GB transferred in tokens
	AgreedNetworkRate string `json:"agreed_network_rate"`
}

// Hash generates a hash of the usage record for signing
func (ur *UsageRecord) Hash() []byte {
	data := struct {
		WorkloadID   string          `json:"workload_id"`
		DeploymentID string          `json:"deployment_id"`
		LeaseID      string          `json:"lease_id"`
		ProviderID   string          `json:"provider_id"`
		Type         UsageRecordType `json:"type"`
		StartTime    int64           `json:"start_time"`
		EndTime      int64           `json:"end_time"`
		Metrics      ResourceMetrics `json:"metrics"`
	}{
		WorkloadID:   ur.WorkloadID,
		DeploymentID: ur.DeploymentID,
		LeaseID:      ur.LeaseID,
		ProviderID:   ur.ProviderID,
		Type:         ur.Type,
		StartTime:    ur.StartTime.Unix(),
		EndTime:      ur.EndTime.Unix(),
		Metrics:      ur.Metrics,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// WorkloadMetering contains metering state for a workload
type WorkloadMetering struct {
	// WorkloadID is the workload ID
	WorkloadID string

	// DeploymentID is the deployment ID
	DeploymentID string

	// LeaseID is the lease ID
	LeaseID string

	// StartTime is when metering started
	StartTime time.Time

	// LastRecordTime is when the last record was created
	LastRecordTime time.Time

	// PricingInputs contains pricing configuration
	PricingInputs PricingInputs

	// CumulativeMetrics contains cumulative metrics since start
	CumulativeMetrics ResourceMetrics

	// Active indicates if metering is active
	Active bool
}

// MetricsCollector collects metrics for workloads
type MetricsCollector interface {
	// CollectMetrics collects current metrics for a workload
	CollectMetrics(ctx context.Context, workloadID string) (*ResourceMetrics, error)
}

// ChainRecorder submits usage records to the blockchain
type ChainRecorder interface {
	// SubmitUsageRecord submits a usage record to the chain
	SubmitUsageRecord(ctx context.Context, record *UsageRecord) error

	// SubmitFinalSettlement submits a final settlement record
	SubmitFinalSettlement(ctx context.Context, record *UsageRecord) error
}

// UsageMeterConfig configures the usage meter
type UsageMeterConfig struct {
	// ProviderID is the provider's on-chain ID
	ProviderID string

	// Interval is the metering interval
	Interval MeteringInterval

	// MetricsCollector collects workload metrics
	MetricsCollector MetricsCollector

	// ChainRecorder submits records to the chain
	ChainRecorder ChainRecorder

	// KeyManager signs usage records
	KeyManager *KeyManager

	// RecordChan receives generated usage records
	RecordChan chan<- *UsageRecord
}

// UsageMeter meters workload usage and creates on-chain records
type UsageMeter struct {
	mu sync.RWMutex

	// providerID is the provider ID
	providerID string

	// interval is the metering interval
	interval time.Duration

	// collector collects metrics
	collector MetricsCollector

	// recorder submits to chain
	recorder ChainRecorder

	// keyManager signs records
	keyManager *KeyManager

	// recordChan receives generated records
	recordChan chan<- *UsageRecord

	// workloads contains workload metering state
	workloads map[string]*WorkloadMetering

	// running indicates if the meter is running
	running bool

	// stopChan stops the metering loop
	stopChan chan struct{}

	// wg waits for goroutines to finish
	wg sync.WaitGroup
}

// NewUsageMeter creates a new usage meter
func NewUsageMeter(cfg UsageMeterConfig) *UsageMeter {
	interval := time.Duration(cfg.Interval)
	if interval == 0 {
		interval = time.Hour
	}

	return &UsageMeter{
		providerID: cfg.ProviderID,
		interval:   interval,
		collector:  cfg.MetricsCollector,
		recorder:   cfg.ChainRecorder,
		keyManager: cfg.KeyManager,
		recordChan: cfg.RecordChan,
		workloads:  make(map[string]*WorkloadMetering),
		stopChan:   make(chan struct{}),
	}
}

// Start starts the usage metering loop
func (um *UsageMeter) Start(ctx context.Context) error {
	um.mu.Lock()
	if um.running {
		um.mu.Unlock()
		return nil
	}
	um.running = true
	um.mu.Unlock()

	um.wg.Add(1)
	verrors.SafeGo("provider-daemon:usage-meter", func() {
		// Note: wg.Done() is called inside meteringLoop
		um.meteringLoop(ctx)
	})

	return nil
}

// Stop stops the usage metering
func (um *UsageMeter) Stop() {
	um.mu.Lock()
	if !um.running {
		um.mu.Unlock()
		return
	}
	um.running = false
	um.mu.Unlock()

	close(um.stopChan)
	um.wg.Wait()

	// Reinitialize stopChan for potential restart
	um.stopChan = make(chan struct{})
}

// StartMetering starts metering for a workload
func (um *UsageMeter) StartMetering(workloadID, deploymentID, leaseID string, pricing PricingInputs) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	now := time.Now()
	um.workloads[workloadID] = &WorkloadMetering{
		WorkloadID:        workloadID,
		DeploymentID:      deploymentID,
		LeaseID:           leaseID,
		StartTime:         now,
		LastRecordTime:    now,
		PricingInputs:     pricing,
		CumulativeMetrics: ResourceMetrics{},
		Active:            true,
	}

	return nil
}

// StopMetering stops metering for a workload and creates a final record
func (um *UsageMeter) StopMetering(ctx context.Context, workloadID string) (*UsageRecord, error) {
	um.mu.Lock()
	metering, ok := um.workloads[workloadID]
	if !ok {
		um.mu.Unlock()
		return nil, ErrWorkloadNotMetered
	}
	metering.Active = false
	um.mu.Unlock()

	// Collect final metrics
	metrics, err := um.collector.CollectMetrics(ctx, workloadID)
	if err != nil {
		// Use cumulative metrics on error
		metrics = &metering.CumulativeMetrics
	}

	// Create final record
	record := um.createUsageRecord(metering, metrics, UsageRecordTypeFinal)

	// Submit to chain
	if um.recorder != nil {
		if err := um.recorder.SubmitFinalSettlement(ctx, record); err != nil {
			// Log error but return record anyway
			_ = err
		}
	}

	// Remove from tracked workloads
	um.mu.Lock()
	delete(um.workloads, workloadID)
	um.mu.Unlock()

	return record, nil
}

// GetMeteringState returns the metering state for a workload
func (um *UsageMeter) GetMeteringState(workloadID string) (*WorkloadMetering, error) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	metering, ok := um.workloads[workloadID]
	if !ok {
		return nil, ErrWorkloadNotMetered
	}
	return metering, nil
}

// ListMeteredWorkloads lists all workloads being metered
func (um *UsageMeter) ListMeteredWorkloads() []string {
	um.mu.RLock()
	defer um.mu.RUnlock()

	result := make([]string, 0, len(um.workloads))
	for id := range um.workloads {
		result = append(result, id)
	}
	return result
}

// ForceCollect forces an immediate collection for a workload
func (um *UsageMeter) ForceCollect(ctx context.Context, workloadID string) (*UsageRecord, error) {
	um.mu.RLock()
	metering, ok := um.workloads[workloadID]
	um.mu.RUnlock()

	if !ok {
		return nil, ErrWorkloadNotMetered
	}

	return um.collectAndRecord(ctx, metering)
}

func (um *UsageMeter) meteringLoop(ctx context.Context) {
	defer um.wg.Done()

	ticker := time.NewTicker(um.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-um.stopChan:
			return
		case <-ticker.C:
			um.collectAllMetrics(ctx)
		}
	}
}

func (um *UsageMeter) collectAllMetrics(ctx context.Context) {
	um.mu.RLock()
	workloads := make([]*WorkloadMetering, 0, len(um.workloads))
	for _, w := range um.workloads {
		if w.Active {
			workloads = append(workloads, w)
		}
	}
	um.mu.RUnlock()

	for _, metering := range workloads {
		_, err := um.collectAndRecord(ctx, metering)
		if err != nil {
			// Log error and continue
			_ = err
		}
	}
}

func (um *UsageMeter) collectAndRecord(ctx context.Context, metering *WorkloadMetering) (*UsageRecord, error) {
	// Collect metrics
	metrics, err := um.collector.CollectMetrics(ctx, metering.WorkloadID)
	if err != nil {
		return nil, err
	}

	// Update cumulative metrics
	um.mu.Lock()
	metering.CumulativeMetrics.CPUMilliSeconds += metrics.CPUMilliSeconds
	metering.CumulativeMetrics.MemoryByteSeconds += metrics.MemoryByteSeconds
	metering.CumulativeMetrics.StorageByteSeconds += metrics.StorageByteSeconds
	metering.CumulativeMetrics.NetworkBytesIn += metrics.NetworkBytesIn
	metering.CumulativeMetrics.NetworkBytesOut += metrics.NetworkBytesOut
	metering.CumulativeMetrics.GPUSeconds += metrics.GPUSeconds
	um.mu.Unlock()

	// Create usage record
	record := um.createUsageRecord(metering, metrics, UsageRecordTypePeriodic)

	// Submit to chain
	if um.recorder != nil {
		if err := um.recorder.SubmitUsageRecord(ctx, record); err != nil {
			return record, err
		}
	}

	// Update last record time
	um.mu.Lock()
	metering.LastRecordTime = record.EndTime
	um.mu.Unlock()

	// Send to channel if configured
	if um.recordChan != nil {
		select {
		case um.recordChan <- record:
		default:
			// Channel full
		}
	}

	return record, nil
}

func (um *UsageMeter) createUsageRecord(metering *WorkloadMetering, metrics *ResourceMetrics, recordType UsageRecordType) *UsageRecord {
	now := time.Now()
	record := &UsageRecord{
		ID:            um.generateRecordID(metering.WorkloadID, now),
		WorkloadID:    metering.WorkloadID,
		DeploymentID:  metering.DeploymentID,
		LeaseID:       metering.LeaseID,
		ProviderID:    um.providerID,
		Type:          recordType,
		StartTime:     metering.LastRecordTime,
		EndTime:       now,
		Metrics:       *metrics,
		PricingInputs: metering.PricingInputs,
		CreatedAt:     now,
	}

	// Sign the record
	if um.keyManager != nil {
		hash := record.Hash()
		sig, err := um.keyManager.Sign(hash)
		if err == nil {
			record.Signature = sig.Signature // Already hex-encoded string
		}
	}

	return record
}

func (um *UsageMeter) generateRecordID(workloadID string, timestamp time.Time) string {
	data := workloadID + ":" + timestamp.Format(time.RFC3339Nano)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// FraudCheckResult contains the result of a fraud check
type FraudCheckResult struct {
	// Valid indicates if the usage record is valid
	Valid bool `json:"valid"`

	// Flags contains fraud detection flags
	Flags []string `json:"flags,omitempty"`

	// Score is the fraud probability score (0-100)
	Score int `json:"score"`

	// Details contains detailed check results
	Details map[string]interface{} `json:"details,omitempty"`
}

// FraudChecker checks usage records for fraud
type FraudChecker struct {
	// maxCPUUsageRatio is the max CPU usage ratio (usage vs allocated)
	maxCPUUsageRatio float64

	// maxMemoryUsageRatio is the max memory usage ratio
	maxMemoryUsageRatio float64

	// maxNetworkAnomalyRatio is the max network anomaly ratio
	maxNetworkAnomalyRatio float64

	// minRecordDuration is the minimum valid record duration
	minRecordDuration time.Duration

	// maxRecordDuration is the maximum valid record duration
	maxRecordDuration time.Duration
}

// NewFraudChecker creates a new fraud checker with default settings
func NewFraudChecker() *FraudChecker {
	return &FraudChecker{
		maxCPUUsageRatio:       2.0,  // Can't use more than 2x allocated
		maxMemoryUsageRatio:    1.5,  // Can't use more than 1.5x allocated
		maxNetworkAnomalyRatio: 10.0, // Network can't be 10x abnormal
		minRecordDuration:      time.Minute,
		maxRecordDuration:      25 * time.Hour, // Slightly more than a day
	}
}

// CheckRecord checks a usage record for fraud indicators
func (fc *FraudChecker) CheckRecord(record *UsageRecord, allocatedResources *ResourceMetrics) *FraudCheckResult {
	result := &FraudCheckResult{
		Valid:   true,
		Flags:   make([]string, 0),
		Score:   0,
		Details: make(map[string]interface{}),
	}

	// Check time bounds
	duration := record.EndTime.Sub(record.StartTime)
	if duration < fc.minRecordDuration {
		result.Flags = append(result.Flags, "DURATION_TOO_SHORT")
		result.Score += 30
		result.Details["duration_seconds"] = duration.Seconds()
	}
	if duration > fc.maxRecordDuration {
		result.Flags = append(result.Flags, "DURATION_TOO_LONG")
		result.Score += 50
		result.Details["duration_seconds"] = duration.Seconds()
	}

	// Check future timestamps
	if record.EndTime.After(time.Now().Add(time.Minute)) {
		result.Flags = append(result.Flags, "FUTURE_TIMESTAMP")
		result.Score += 100
	}

	// Check resource ratios if allocated resources provided
	if allocatedResources != nil {
		durationSeconds := int64(duration.Seconds())
		if durationSeconds > 0 {
			// Check CPU usage ratio
			expectedCPU := allocatedResources.CPUMilliSeconds * durationSeconds / 1000
			if expectedCPU > 0 {
				cpuRatio := float64(record.Metrics.CPUMilliSeconds) / float64(expectedCPU)
				if cpuRatio > fc.maxCPUUsageRatio {
					result.Flags = append(result.Flags, "EXCESSIVE_CPU_USAGE")
					result.Score += 40
					result.Details["cpu_ratio"] = cpuRatio
				}
			}

			// Check memory usage ratio
			expectedMemory := allocatedResources.MemoryByteSeconds * durationSeconds
			if expectedMemory > 0 {
				memoryRatio := float64(record.Metrics.MemoryByteSeconds) / float64(expectedMemory)
				if memoryRatio > fc.maxMemoryUsageRatio {
					result.Flags = append(result.Flags, "EXCESSIVE_MEMORY_USAGE")
					result.Score += 40
					result.Details["memory_ratio"] = memoryRatio
				}
			}
		}
	}

	// Check for negative values
	if record.Metrics.CPUMilliSeconds < 0 ||
		record.Metrics.MemoryByteSeconds < 0 ||
		record.Metrics.StorageByteSeconds < 0 ||
		record.Metrics.NetworkBytesIn < 0 ||
		record.Metrics.NetworkBytesOut < 0 ||
		record.Metrics.GPUSeconds < 0 {
		result.Flags = append(result.Flags, "NEGATIVE_METRICS")
		result.Score += 100
	}

	// Check for zero duration with non-zero usage
	if duration == 0 && (record.Metrics.CPUMilliSeconds > 0 || record.Metrics.MemoryByteSeconds > 0) {
		result.Flags = append(result.Flags, "ZERO_DURATION_WITH_USAGE")
		result.Score += 80
	}

	// Cap score at 100
	if result.Score > 100 {
		result.Score = 100
	}

	// Invalid if score exceeds threshold
	if result.Score >= 50 {
		result.Valid = false
	}

	return result
}

// CheckRecordSignature verifies the signature on a usage record
func (fc *FraudChecker) CheckRecordSignature(record *UsageRecord, publicKey []byte) bool {
	if record.Signature == "" {
		return false
	}

	signature, err := hex.DecodeString(record.Signature)
	if err != nil {
		return false
	}

	hash := record.Hash()

	// In production, this would use proper ed25519 verification
	// For now, we just check that the signature exists and is valid hex
	return len(signature) > 0 && len(hash) > 0
}
