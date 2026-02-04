// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-21C: HPC Settlement Pipeline - batches and submits job accounting to chain
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// SettlementRecordStatus represents the status of a settlement record
type SettlementRecordStatus string

const (
	// SettlementRecordStatusPending indicates the record is pending submission
	SettlementRecordStatusPending SettlementRecordStatus = "pending"

	// SettlementRecordStatusSubmitted indicates the record has been submitted to chain
	SettlementRecordStatusSubmitted SettlementRecordStatus = "submitted"

	// SettlementRecordStatusConfirmed indicates the record has been confirmed on-chain
	SettlementRecordStatusConfirmed SettlementRecordStatus = "confirmed"

	// SettlementRecordStatusFailed indicates the record failed to settle
	SettlementRecordStatusFailed SettlementRecordStatus = "failed"
)

// IsValid checks if the settlement status is valid
func (s SettlementRecordStatus) IsValid() bool {
	switch s {
	case SettlementRecordStatusPending, SettlementRecordStatusSubmitted,
		SettlementRecordStatusConfirmed, SettlementRecordStatusFailed:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the settlement status is terminal
func (s SettlementRecordStatus) IsTerminal() bool {
	return s == SettlementRecordStatusConfirmed || s == SettlementRecordStatusFailed
}

// HPCBatchSettlementConfig configures the batch settlement pipeline
type HPCBatchSettlementConfig struct {
	// Enabled indicates if the pipeline is enabled
	Enabled bool

	// BatchSize is the maximum number of records to batch together
	BatchSize int

	// BatchInterval is the interval between batch submissions
	BatchInterval time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryBackoff is the initial backoff duration for retries
	RetryBackoff time.Duration

	// MaxPendingRecords is the max number of pending records before forcing submission
	MaxPendingRecords int
}

// DefaultHPCBatchSettlementConfig returns default configuration
func DefaultHPCBatchSettlementConfig() HPCBatchSettlementConfig {
	return HPCBatchSettlementConfig{
		Enabled:           true,
		BatchSize:         50,
		BatchInterval:     time.Minute * 5,
		MaxRetries:        3,
		RetryBackoff:      time.Second * 5,
		MaxPendingRecords: 100,
	}
}

// Validate validates the settlement config
func (c *HPCBatchSettlementConfig) Validate() error {
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive")
	}
	if c.BatchInterval <= 0 {
		return fmt.Errorf("batch_interval must be positive")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	if c.RetryBackoff <= 0 {
		return fmt.Errorf("retry_backoff must be positive")
	}
	if c.MaxPendingRecords <= 0 {
		return fmt.Errorf("max_pending_records must be positive")
	}
	return nil
}

// HPCSettlementConfig is an alias for HPCBatchSettlementConfig for compatibility
type HPCSettlementConfig = HPCBatchSettlementConfig

// DefaultHPCSettlementConfig returns the default settlement config
func DefaultHPCSettlementConfig() HPCSettlementConfig {
	return DefaultHPCBatchSettlementConfig()
}

// HPCSettlementPipeline is an alias for HPCBatchSettlementPipeline for compatibility
type HPCSettlementPipeline = HPCBatchSettlementPipeline

// NewHPCSettlementPipeline creates a new HPC settlement pipeline
func NewHPCSettlementPipeline(
	config HPCSettlementConfig,
	reporter HPCOnChainReporter,
	signer HPCSchedulerSigner,
) *HPCSettlementPipeline {
	return NewHPCBatchSettlementPipeline(config, reporter, signer)
}

// HPCSettlementRecord represents a record queued for settlement
type HPCSettlementRecord struct {
	// JobID is the VirtEngine job ID
	JobID string `json:"job_id"`

	// ClusterID is the HPC cluster ID
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer address
	CustomerAddress string `json:"customer_address"`

	// UsageMetrics contains the usage metrics from the scheduler
	UsageMetrics *HPCSchedulerMetrics `json:"usage_metrics"`

	// Status is the settlement status
	Status SettlementRecordStatus `json:"status"`

	// Attempts is the number of submission attempts
	Attempts int `json:"attempts"`

	// LastAttempt is the time of the last submission attempt
	LastAttempt time.Time `json:"last_attempt"`

	// LastError is the last error encountered
	LastError error `json:"-"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// ConfirmedAt is when the record was confirmed on-chain
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`

	// UsageRecordIDs are the source usage record IDs
	UsageRecordIDs []string `json:"usage_record_ids,omitempty"`

	// IsFinal indicates if this is the final settlement for the job
	IsFinal bool `json:"is_final"`

	// TxHash is the transaction hash if submitted
	TxHash string `json:"tx_hash,omitempty"`
}

// Hash generates a hash of the settlement record for deduplication.
// Uses job identity fields for deduplication, not mutable fields like CreatedAt.
func (r *HPCSettlementRecord) Hash() string {
	data := fmt.Sprintf("%s:%s:%s:%s",
		r.JobID,
		r.ClusterID,
		r.ProviderAddress,
		r.CustomerAddress,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Validate validates the settlement record
func (r *HPCSettlementRecord) Validate() error {
	if r.JobID == "" {
		return fmt.Errorf("job_id cannot be empty")
	}
	if r.ClusterID == "" {
		return fmt.Errorf("cluster_id cannot be empty")
	}
	if r.ProviderAddress == "" {
		return fmt.Errorf("provider_address cannot be empty")
	}
	if r.CustomerAddress == "" {
		return fmt.Errorf("customer_address cannot be empty")
	}
	if r.UsageMetrics == nil {
		return fmt.Errorf("usage_metrics cannot be nil")
	}
	if !r.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	return nil
}

// CanRetry checks if the record can be retried
func (r *HPCSettlementRecord) CanRetry(maxRetries int) bool {
	return r.Status == SettlementRecordStatusFailed && r.Attempts < maxRetries
}

// HPCSettlementStats contains statistics for the settlement pipeline
type HPCSettlementStats struct {
	// TotalQueued is the total number of records ever queued
	TotalQueued int64 `json:"total_queued"`

	// TotalSubmitted is the total number of records submitted
	TotalSubmitted int64 `json:"total_submitted"`

	// TotalConfirmed is the total number of confirmed settlements
	TotalConfirmed int64 `json:"total_confirmed"`

	// TotalFailed is the total number of failed settlements
	TotalFailed int64 `json:"total_failed"`

	// TotalRetried is the total number of retried submissions
	TotalRetried int64 `json:"total_retried"`

	// PendingCount is the current number of pending records
	PendingCount int `json:"pending_count"`

	// SubmittedCount is the current number of submitted (awaiting confirmation) records
	SubmittedCount int `json:"submitted_count"`

	// ConfirmedCount is the current number of confirmed records in cache
	ConfirmedCount int `json:"confirmed_count"`

	// FailedCount is the current number of failed records
	FailedCount int `json:"failed_count"`

	// LastSubmissionTime is the time of the last batch submission
	LastSubmissionTime *time.Time `json:"last_submission_time,omitempty"`

	// LastError is the last error encountered
	LastError string `json:"last_error,omitempty"`

	// BatchesSubmitted is the total number of batches submitted
	BatchesSubmitted int64 `json:"batches_submitted"`

	// AverageConfirmationTime is the average time to confirmation
	AverageConfirmationTime time.Duration `json:"average_confirmation_time"`
}

// HPCBatchSettlementPipeline batches and submits HPC job accounting to the blockchain
type HPCBatchSettlementPipeline struct {
	config   HPCBatchSettlementConfig
	reporter HPCOnChainReporter
	signer   HPCSchedulerSigner

	mu sync.RWMutex

	// pending contains records waiting to be submitted
	pending map[string]*HPCSettlementRecord

	// submitted contains records waiting for confirmation
	submitted map[string]*HPCSettlementRecord

	// confirmed contains recently confirmed records (for stats)
	confirmed map[string]*HPCSettlementRecord

	// failed contains failed records
	failed map[string]*HPCSettlementRecord

	// stats tracks pipeline statistics
	stats HPCSettlementStats

	// running indicates if the pipeline is running
	running bool

	// stopCh signals the pipeline to stop
	stopCh chan struct{}

	// wg waits for goroutines to finish
	wg sync.WaitGroup

	// confirmationTimes tracks time to confirmation for averaging
	confirmationTimes []time.Duration
}

// NewHPCBatchSettlementPipeline creates a new HPC batch settlement pipeline
func NewHPCBatchSettlementPipeline(
	config HPCBatchSettlementConfig,
	reporter HPCOnChainReporter,
	signer HPCSchedulerSigner,
) *HPCBatchSettlementPipeline {
	return &HPCBatchSettlementPipeline{
		config:            config,
		reporter:          reporter,
		signer:            signer,
		pending:           make(map[string]*HPCSettlementRecord),
		submitted:         make(map[string]*HPCSettlementRecord),
		confirmed:         make(map[string]*HPCSettlementRecord),
		failed:            make(map[string]*HPCSettlementRecord),
		stopCh:            make(chan struct{}),
		confirmationTimes: make([]time.Duration, 0, 100),
	}
}

// Start starts the settlement pipeline
func (p *HPCBatchSettlementPipeline) Start(ctx context.Context) error {
	if !p.config.Enabled {
		return nil
	}

	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = true
	p.stopCh = make(chan struct{})
	p.mu.Unlock()

	p.wg.Add(1)
	verrors.SafeGo("provider-daemon:hpc-batch-settlement-pipeline", func() {
		defer p.wg.Done()
		p.runLoop(ctx)
	})

	log.Printf("[hpc-batch-settlement-pipeline] started with batch_size=%d, interval=%v",
		p.config.BatchSize, p.config.BatchInterval)

	return nil
}

// Stop stops the settlement pipeline
func (p *HPCBatchSettlementPipeline) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = false
	close(p.stopCh)
	p.mu.Unlock()

	p.wg.Wait()

	log.Printf("[hpc-batch-settlement-pipeline] stopped")
	return nil
}

// QueueSettlement queues a settlement record for batch submission
func (p *HPCBatchSettlementPipeline) QueueSettlement(record *HPCSettlementRecord) error {
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	if err := record.Validate(); err != nil {
		return fmt.Errorf("invalid record: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already pending or submitted
	recordHash := record.Hash()
	if _, exists := p.pending[recordHash]; exists {
		return nil // Already queued, skip
	}
	if _, exists := p.submitted[recordHash]; exists {
		return nil // Already submitted, skip
	}

	// Check max pending limit
	if len(p.pending) >= p.config.MaxPendingRecords {
		return fmt.Errorf("max pending records reached (%d)", p.config.MaxPendingRecords)
	}

	// Set initial status
	record.Status = SettlementRecordStatusPending
	record.CreatedAt = time.Now()
	record.Attempts = 0

	p.pending[recordHash] = record
	p.stats.TotalQueued++
	p.stats.PendingCount = len(p.pending)

	return nil
}

// GetPendingCount returns the number of pending records
func (p *HPCBatchSettlementPipeline) GetPendingCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pending)
}

// GetStats returns the current pipeline statistics
func (p *HPCBatchSettlementPipeline) GetStats() *HPCSettlementStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.stats
	stats.PendingCount = len(p.pending)
	stats.SubmittedCount = len(p.submitted)
	stats.ConfirmedCount = len(p.confirmed)
	stats.FailedCount = len(p.failed)

	return &stats
}

// GetPendingRecords returns a copy of pending records
func (p *HPCBatchSettlementPipeline) GetPendingRecords() []*HPCSettlementRecord {
	p.mu.RLock()
	defer p.mu.RUnlock()

	records := make([]*HPCSettlementRecord, 0, len(p.pending))
	for _, r := range p.pending {
		records = append(records, r)
	}
	return records
}

// GetSubmittedRecords returns a copy of submitted records awaiting confirmation
func (p *HPCBatchSettlementPipeline) GetSubmittedRecords() []*HPCSettlementRecord {
	p.mu.RLock()
	defer p.mu.RUnlock()

	records := make([]*HPCSettlementRecord, 0, len(p.submitted))
	for _, r := range p.submitted {
		records = append(records, r)
	}
	return records
}

// GetFailedRecords returns a copy of failed records
func (p *HPCBatchSettlementPipeline) GetFailedRecords() []*HPCSettlementRecord {
	p.mu.RLock()
	defer p.mu.RUnlock()

	records := make([]*HPCSettlementRecord, 0, len(p.failed))
	for _, r := range p.failed {
		records = append(records, r)
	}
	return records
}

// ConfirmSettlement marks a submitted record as confirmed
func (p *HPCBatchSettlementPipeline) ConfirmSettlement(jobID string, txHash string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Find the record in submitted
	var recordHash string
	var record *HPCSettlementRecord
	for hash, r := range p.submitted {
		if r.JobID == jobID && (txHash == "" || r.TxHash == txHash) {
			recordHash = hash
			record = r
			break
		}
	}

	if record == nil {
		return fmt.Errorf("submitted record not found for job: %s", jobID)
	}

	now := time.Now()
	record.Status = SettlementRecordStatusConfirmed
	record.ConfirmedAt = &now

	// Track confirmation time
	confirmTime := now.Sub(record.LastAttempt)
	p.confirmationTimes = append(p.confirmationTimes, confirmTime)
	if len(p.confirmationTimes) > 100 {
		p.confirmationTimes = p.confirmationTimes[1:]
	}

	// Calculate average confirmation time
	var totalTime time.Duration
	for _, t := range p.confirmationTimes {
		totalTime += t
	}
	if len(p.confirmationTimes) > 0 {
		p.stats.AverageConfirmationTime = totalTime / time.Duration(len(p.confirmationTimes))
	}

	// Move to confirmed
	delete(p.submitted, recordHash)
	p.confirmed[recordHash] = record
	p.stats.TotalConfirmed++
	p.stats.SubmittedCount = len(p.submitted)

	log.Printf("[hpc-batch-settlement-pipeline] confirmed settlement for job %s (tx: %s)", jobID, txHash)

	return nil
}

// RetryFailed retries all failed records that haven't exceeded max retries
func (p *HPCBatchSettlementPipeline) RetryFailed() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	retried := 0
	for hash, record := range p.failed {
		if record.CanRetry(p.config.MaxRetries) {
			record.Status = SettlementRecordStatusPending
			p.pending[hash] = record
			delete(p.failed, hash)
			retried++
			p.stats.TotalRetried++
		}
	}

	p.stats.PendingCount = len(p.pending)
	return retried
}

// runLoop runs the settlement pipeline loop
func (p *HPCBatchSettlementPipeline) runLoop(ctx context.Context) {
	batchTicker := time.NewTicker(p.config.BatchInterval)
	defer batchTicker.Stop()

	retryTicker := time.NewTicker(p.config.RetryBackoff * 10) // Retry check every 10x backoff
	defer retryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-batchTicker.C:
			p.processBatch(ctx)
		case <-retryTicker.C:
			p.processRetries(ctx)
		}
	}
}

// processBatch processes a batch of pending records
func (p *HPCBatchSettlementPipeline) processBatch(ctx context.Context) {
	p.mu.Lock()

	if len(p.pending) == 0 {
		p.mu.Unlock()
		return
	}

	// Select up to BatchSize records
	batch := make([]*HPCSettlementRecord, 0, p.config.BatchSize)
	batchHashes := make([]string, 0, p.config.BatchSize)

	for hash, record := range p.pending {
		if len(batch) >= p.config.BatchSize {
			break
		}
		batch = append(batch, record)
		batchHashes = append(batchHashes, hash)
	}

	p.mu.Unlock()

	if len(batch) == 0 {
		return
	}

	log.Printf("[hpc-batch-settlement-pipeline] processing batch of %d records", len(batch))

	// Submit each record in the batch
	successCount := 0
	failCount := 0

	for i, record := range batch {
		if err := p.submitRecord(ctx, record); err != nil {
			log.Printf("[hpc-batch-settlement-pipeline] failed to submit job %s: %v", record.JobID, err)
			failCount++

			p.mu.Lock()
			record.Status = SettlementRecordStatusFailed
			record.LastError = err
			record.Attempts++
			record.LastAttempt = time.Now()

			hash := batchHashes[i]
			if record.Attempts >= p.config.MaxRetries {
				p.failed[hash] = record
				delete(p.pending, hash)
				p.stats.TotalFailed++
			}
			p.stats.LastError = err.Error()
			p.mu.Unlock()
		} else {
			successCount++

			p.mu.Lock()
			record.Status = SettlementRecordStatusSubmitted
			record.LastAttempt = time.Now()
			record.Attempts++

			hash := batchHashes[i]
			p.submitted[hash] = record
			delete(p.pending, hash)
			p.stats.TotalSubmitted++
			p.mu.Unlock()
		}
	}

	p.mu.Lock()
	now := time.Now()
	p.stats.LastSubmissionTime = &now
	p.stats.BatchesSubmitted++
	p.stats.PendingCount = len(p.pending)
	p.stats.SubmittedCount = len(p.submitted)
	p.mu.Unlock()

	log.Printf("[hpc-batch-settlement-pipeline] batch complete: %d submitted, %d failed", successCount, failCount)
}

// submitRecord submits a single record to the chain
func (p *HPCBatchSettlementPipeline) submitRecord(ctx context.Context, record *HPCSettlementRecord) error {
	if p.reporter == nil {
		return fmt.Errorf("reporter not configured")
	}

	if record.UsageMetrics == nil {
		return fmt.Errorf("usage metrics is nil")
	}

	// Submit to chain via reporter
	return p.reporter.ReportJobAccounting(ctx, record.JobID, record.UsageMetrics)
}

// processRetries checks for failed records that can be retried
func (p *HPCBatchSettlementPipeline) processRetries(_ context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	retryCount := 0

	for hash, record := range p.failed {
		if !record.CanRetry(p.config.MaxRetries) {
			continue
		}

		// Calculate backoff: backoff * 2^attempts
		backoff := p.config.RetryBackoff * time.Duration(1<<uint(record.Attempts))
		if now.Sub(record.LastAttempt) < backoff {
			continue
		}

		// Move back to pending for retry
		record.Status = SettlementRecordStatusPending
		p.pending[hash] = record
		delete(p.failed, hash)
		retryCount++
		p.stats.TotalRetried++
	}

	if retryCount > 0 {
		p.stats.PendingCount = len(p.pending)
		log.Printf("[hpc-batch-settlement-pipeline] queued %d records for retry", retryCount)
	}
}

// ClearConfirmed clears the confirmed records cache
func (p *HPCBatchSettlementPipeline) ClearConfirmed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.confirmed = make(map[string]*HPCSettlementRecord)
}

// ClearFailed clears all failed records
func (p *HPCBatchSettlementPipeline) ClearFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failed = make(map[string]*HPCSettlementRecord)
}

// IsRunning returns whether the pipeline is running
func (p *HPCBatchSettlementPipeline) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// ForceFlush forces immediate submission of all pending records
func (p *HPCBatchSettlementPipeline) ForceFlush(ctx context.Context) error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return fmt.Errorf("pipeline not running")
	}
	p.mu.Unlock()

	// Process all pending in batches
	for {
		p.mu.RLock()
		pendingCount := len(p.pending)
		p.mu.RUnlock()

		if pendingCount == 0 {
			break
		}

		p.processBatch(ctx)

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

// CreateSettlementRecordFromUsage creates a settlement record from an HPCUsageRecord
func CreateSettlementRecordFromUsage(usage *HPCUsageRecord) *HPCSettlementRecord {
	if usage == nil {
		return nil
	}
	return &HPCSettlementRecord{
		JobID:           usage.JobID,
		ClusterID:       usage.ClusterID,
		ProviderAddress: usage.ProviderAddress,
		CustomerAddress: usage.CustomerAddress,
		UsageMetrics:    usage.Metrics,
		Status:          SettlementRecordStatusPending,
		UsageRecordIDs:  []string{usage.RecordID},
		IsFinal:         usage.IsFinal,
		CreatedAt:       time.Now(),
	}
}
