// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-3D: Waldur ingestion worker for automatic Waldur-to-chain synchronization.
// This file implements the ingestion worker that fetches Waldur offerings and
// ingests them on-chain, with retry/backoff, rate limiting, and reconciliation.
package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
	"golang.org/x/time/rate"
)

// WaldurIngestAuditLogger defines the interface for audit logging.
type WaldurIngestAuditLogger interface {
	// LogIngestAttempt logs an ingestion operation attempt.
	LogIngestAttempt(entry IngestAuditEntry)
	// LogReconciliation logs a reconciliation run.
	LogReconciliation(entry IngestReconciliationAuditEntry)
	// LogDeadLetter logs a dead-letter event.
	LogDeadLetter(entry IngestDeadLetterAuditEntry)
}

// IngestAuditEntry represents an audit log entry for an ingestion operation.
type IngestAuditEntry struct {
	Timestamp       time.Time              `json:"timestamp"`
	WaldurUUID      string                 `json:"waldur_uuid"`
	OfferingName    string                 `json:"offering_name"`
	ChainOfferingID string                 `json:"chain_offering_id,omitempty"`
	Action          string                 `json:"action"`
	Success         bool                   `json:"success"`
	Duration        time.Duration          `json:"duration_ns"`
	Error           string                 `json:"error,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	ProviderAddress string                 `json:"provider_address"`
	Checksum        string                 `json:"checksum,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// IngestReconciliationAuditEntry represents an audit log entry for reconciliation.
type IngestReconciliationAuditEntry struct {
	Timestamp         time.Time     `json:"timestamp"`
	WaldurCustomer    string        `json:"waldur_customer"`
	ProviderAddress   string        `json:"provider_address"`
	OfferingsChecked  int           `json:"offerings_checked"`
	DriftDetected     int           `json:"drift_detected"`
	OfferingsQueued   int           `json:"offerings_queued"`
	NewOfferingsFound int           `json:"new_offerings_found"`
	Duration          time.Duration `json:"duration_ns"`
	Error             string        `json:"error,omitempty"`
}

// IngestDeadLetterAuditEntry represents an audit log entry for dead-letter events.
type IngestDeadLetterAuditEntry struct {
	Timestamp       time.Time `json:"timestamp"`
	WaldurUUID      string    `json:"waldur_uuid"`
	OfferingName    string    `json:"offering_name"`
	ProviderAddress string    `json:"provider_address"`
	TotalAttempts   int       `json:"total_attempts"`
	LastError       string    `json:"last_error"`
	Action          string    `json:"action"` // "dead_lettered" or "reprocessed"
}

// DefaultIngestAuditLogger logs to standard log in JSON format.
type DefaultIngestAuditLogger struct {
	prefix string
}

// NewDefaultIngestAuditLogger creates a new default audit logger.
func NewDefaultIngestAuditLogger(prefix string) *DefaultIngestAuditLogger {
	if prefix == "" {
		prefix = "[waldur-ingest-audit]"
	}
	return &DefaultIngestAuditLogger{prefix: prefix}
}

// LogIngestAttempt logs an ingestion operation.
func (l *DefaultIngestAuditLogger) LogIngestAttempt(entry IngestAuditEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s ingest: failed to marshal entry: %v", l.prefix, err)
		return
	}
	log.Printf("%s ingest: %s", l.prefix, string(data))
}

// LogReconciliation logs a reconciliation run.
func (l *DefaultIngestAuditLogger) LogReconciliation(entry IngestReconciliationAuditEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s reconcile: failed to marshal entry: %v", l.prefix, err)
		return
	}
	log.Printf("%s reconcile: %s", l.prefix, string(data))
}

// LogDeadLetter logs a dead-letter event.
func (l *DefaultIngestAuditLogger) LogDeadLetter(entry IngestDeadLetterAuditEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s dead_letter: failed to marshal entry: %v", l.prefix, err)
		return
	}
	log.Printf("%s dead_letter: %s", l.prefix, string(data))
}

// WaldurIngestPrometheusMetrics provides Prometheus-compatible metric counters.
type WaldurIngestPrometheusMetrics struct {
	// Ingestion operation counters
	IngestsTotal        atomic.Int64
	IngestsSuccessful   atomic.Int64
	IngestsFailed       atomic.Int64
	IngestsDeadLettered atomic.Int64
	IngestsSkipped      atomic.Int64

	// Offering counters
	OfferingsCreated    atomic.Int64
	OfferingsUpdated    atomic.Int64
	OfferingsDeprecated atomic.Int64

	// Fetch counters
	FetchesTotal      atomic.Int64
	FetchesSuccessful atomic.Int64
	FetchesFailed     atomic.Int64

	// Reconciliation counters
	ReconciliationsRun atomic.Int64
	DriftDetected      atomic.Int64

	// Rate limiting
	RateLimitHits atomic.Int64

	// Timing histograms (simplified - in production use prometheus.Histogram)
	IngestDurationSum   atomic.Int64
	IngestDurationCount atomic.Int64

	// Current state
	QueueDepth     atomic.Int64
	ActiveIngests  atomic.Int64
	DeadLetterSize atomic.Int64
}

// WaldurIngestWorkerConfig configures the ingestion worker.
type WaldurIngestWorkerConfig struct {
	// Enabled toggles the ingestion worker.
	Enabled bool

	// ProviderAddress is the provider's on-chain address.
	ProviderAddress string

	// WaldurCustomerUUID is the Waldur customer UUID to ingest from.
	WaldurCustomerUUID string

	// WaldurCategoryUUIDs filters offerings to specific categories (empty = all).
	WaldurCategoryUUIDs []string

	// WaldurOfferingTypes filters offerings to specific types (empty = all).
	WaldurOfferingTypes []string

	// IngestIntervalSeconds is how often to run full ingestion.
	IngestIntervalSeconds int64

	// ReconcileIntervalSeconds is how often to run reconciliation.
	ReconcileIntervalSeconds int64

	// ReconcileOnStartup triggers full reconciliation on start.
	ReconcileOnStartup bool

	// PageSize is the number of offerings to fetch per page.
	PageSize int

	// MaxRetries is the max retry attempts before dead-lettering.
	MaxRetries int

	// RetryBackoffSeconds is the base backoff duration.
	RetryBackoffSeconds int64

	// MaxBackoffSeconds is the maximum backoff duration.
	MaxBackoffSeconds int64

	// StateFilePath is the path to persist ingestion state.
	StateFilePath string

	// RateLimitPerSecond limits Waldur API calls per second.
	RateLimitPerSecond float64

	// RateLimitBurst is the burst limit for rate limiting.
	RateLimitBurst int

	// OperationTimeout for individual ingestion operations.
	OperationTimeout time.Duration

	// IngestConfig contains field mapping configuration.
	IngestConfig marketplace.IngestConfig

	// SkipSharedOfferings skips offerings not marked as shared.
	SkipSharedOfferings bool

	// SkipNonBillableOfferings skips non-billable offerings.
	SkipNonBillableOfferings bool

	// OnlyActiveOfferings only ingests Active state offerings.
	OnlyActiveOfferings bool
}

// DefaultWaldurIngestWorkerConfig returns sensible defaults.
func DefaultWaldurIngestWorkerConfig() WaldurIngestWorkerConfig {
	return WaldurIngestWorkerConfig{
		IngestIntervalSeconds:    3600, // 1 hour
		ReconcileIntervalSeconds: 300,  // 5 minutes
		ReconcileOnStartup:       true,
		PageSize:                 50,
		MaxRetries:               5,
		RetryBackoffSeconds:      30,
		MaxBackoffSeconds:        3600,
		RateLimitPerSecond:       2.0,
		RateLimitBurst:           5,
		OperationTimeout:         60 * time.Second,
		StateFilePath:            "data/waldur_ingest_state.json",
		IngestConfig:             marketplace.DefaultIngestConfig(),
		SkipSharedOfferings:      false,
		SkipNonBillableOfferings: false,
		OnlyActiveOfferings:      true,
	}
}

// OfferingSubmitter defines the interface for submitting offerings on-chain.
type OfferingSubmitter interface {
	// CreateOffering creates a new offering on-chain.
	CreateOffering(ctx context.Context, offering *marketplace.Offering) (string, error)
	// UpdateOffering updates an existing offering on-chain.
	UpdateOffering(ctx context.Context, offeringID string, offering *marketplace.Offering) error
	// DeprecateOffering deprecates an offering on-chain.
	DeprecateOffering(ctx context.Context, offeringID string) error
	// GetNextOfferingSequence returns the next sequence number for offerings.
	GetNextOfferingSequence(ctx context.Context, providerAddress string) (uint64, error)
	// ValidateProviderVEID validates the provider's VEID score.
	ValidateProviderVEID(ctx context.Context, providerAddress string, minScore uint32) error
}

// WaldurIngestWorker ingests Waldur offerings onto the chain.
type WaldurIngestWorker struct {
	cfg         WaldurIngestWorkerConfig
	marketplace *waldur.MarketplaceClient
	submitter   OfferingSubmitter
	stateStore  *WaldurIngestStateStore
	state       *WaldurIngestState
	rateLimiter *rate.Limiter

	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
	doneCh      chan struct{}
	ingestQueue chan *WaldurIngestTask

	metrics     *WaldurIngestWorkerMetrics
	auditLogger WaldurIngestAuditLogger
	promMetrics *WaldurIngestPrometheusMetrics
}

// WaldurIngestTask represents an ingestion task to execute.
type WaldurIngestTask struct {
	WaldurUUID string
	Offering   *marketplace.WaldurOfferingImport
	Action     marketplace.IngestAction
	Timestamp  time.Time
}

// WaldurIngestWorkerMetrics tracks worker metrics.
type WaldurIngestWorkerMetrics struct {
	mu                    sync.RWMutex
	IngestsTotal          int64
	IngestsSuccessful     int64
	IngestsFailed         int64
	IngestsDeadLettered   int64
	IngestsSkipped        int64
	DriftDetections       int64
	ReconciliationsRun    int64
	LastIngestTime        time.Time
	LastSuccessTime       time.Time
	LastReconcileTime     time.Time
	WorkerUptime          time.Time
	OfferingsCreated      int64
	OfferingsUpdated      int64
	OfferingsDeprecated   int64
	QueueDepth            int
	AverageIngestDuration time.Duration
}

// NewWaldurIngestWorker creates a new ingestion worker.
func NewWaldurIngestWorker(
	cfg WaldurIngestWorkerConfig,
	marketplaceClient *waldur.MarketplaceClient,
	submitter OfferingSubmitter,
) (*WaldurIngestWorker, error) {
	return NewWaldurIngestWorkerWithLogger(cfg, marketplaceClient, submitter, nil)
}

// NewWaldurIngestWorkerWithLogger creates a new ingestion worker with a custom audit logger.
func NewWaldurIngestWorkerWithLogger(
	cfg WaldurIngestWorkerConfig,
	marketplaceClient *waldur.MarketplaceClient,
	submitter OfferingSubmitter,
	auditLogger WaldurIngestAuditLogger,
) (*WaldurIngestWorker, error) {
	if marketplaceClient == nil {
		return nil, fmt.Errorf("marketplace client is required")
	}
	if submitter == nil {
		return nil, fmt.Errorf("offering submitter is required")
	}

	if auditLogger == nil {
		auditLogger = NewDefaultIngestAuditLogger("[waldur-ingest-audit]")
	}

	stateStore := NewWaldurIngestStateStore(cfg.StateFilePath)

	// Create rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(cfg.RateLimitPerSecond), cfg.RateLimitBurst)

	return &WaldurIngestWorker{
		cfg:         cfg,
		marketplace: marketplaceClient,
		submitter:   submitter,
		stateStore:  stateStore,
		rateLimiter: rateLimiter,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		ingestQueue: make(chan *WaldurIngestTask, 100),
		metrics: &WaldurIngestWorkerMetrics{
			WorkerUptime: time.Now().UTC(),
		},
		auditLogger: auditLogger,
		promMetrics: &WaldurIngestPrometheusMetrics{},
	}, nil
}

// Start starts the ingestion worker.
func (w *WaldurIngestWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.running = true
	w.mu.Unlock()

	// Load persisted state
	state, err := w.stateStore.Load(w.cfg.WaldurCustomerUUID, w.cfg.ProviderAddress)
	if err != nil {
		return fmt.Errorf("load ingest state: %w", err)
	}
	w.state = state

	log.Printf("[waldur-ingest] started for provider %s (customer %s)",
		w.cfg.ProviderAddress, w.cfg.WaldurCustomerUUID)

	// Start background goroutines
	go w.ingestLoop(ctx)
	go w.processLoop(ctx)
	go w.reconcileLoop(ctx)

	// Initial reconciliation if enabled
	if w.cfg.ReconcileOnStartup {
		go func() {
			time.Sleep(5 * time.Second) // Wait for system to stabilize
			if err := w.Reconcile(ctx); err != nil {
				log.Printf("[waldur-ingest] initial reconcile failed: %v", err)
			}
		}()
	}

	return nil
}

// Stop stops the ingestion worker.
func (w *WaldurIngestWorker) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = false
	w.mu.Unlock()

	close(w.stopCh)

	// Wait for goroutines to finish with timeout
	select {
	case <-w.doneCh:
	case <-time.After(10 * time.Second):
		log.Printf("[waldur-ingest] shutdown timeout")
	}

	// Save final state
	if err := w.stateStore.Save(w.state); err != nil {
		log.Printf("[waldur-ingest] failed to save final state: %v", err)
	}

	log.Printf("[waldur-ingest] stopped")
	return nil
}

// ingestLoop periodically fetches and queues Waldur offerings.
func (w *WaldurIngestWorker) ingestLoop(ctx context.Context) {
	if w.cfg.IngestIntervalSeconds <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(w.cfg.IngestIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.FetchAndQueueOfferings(ctx); err != nil {
				log.Printf("[waldur-ingest] fetch error: %v", err)
			}
		}
	}
}

// FetchAndQueueOfferings fetches offerings from Waldur and queues them for ingestion.
func (w *WaldurIngestWorker) FetchAndQueueOfferings(ctx context.Context) error {
	log.Printf("[waldur-ingest] starting offering fetch")
	startTime := time.Now()

	page := 1
	totalQueued := 0

	for {
		// Rate limit
		if err := w.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit wait: %w", err)
		}

		w.promMetrics.FetchesTotal.Add(1)

		// Build list params with filters
		params := waldur.ListOfferingsParams{
			CustomerUUID: w.cfg.WaldurCustomerUUID,
			Page:         page,
			PageSize:     w.cfg.PageSize,
		}

		if w.cfg.OnlyActiveOfferings {
			params.State = "Active"
		}

		offerings, err := w.marketplace.ListOfferings(ctx, params)
		if err != nil {
			w.promMetrics.FetchesFailed.Add(1)
			return fmt.Errorf("list offerings page %d: %w", page, err)
		}

		w.promMetrics.FetchesSuccessful.Add(1)

		if len(offerings) == 0 {
			break
		}

		for _, o := range offerings {
			// Apply filters
			if w.shouldSkipOffering(&o) {
				continue
			}

			// Convert to import format
			importOffering := w.convertToImport(&o)

			// Validate
			validation := importOffering.Validate(w.cfg.IngestConfig)
			if !validation.Valid {
				log.Printf("[waldur-ingest] skipping %s: %v", o.UUID, validation.Errors)
				w.state.MarkSkipped(o.UUID, fmt.Sprintf("validation failed: %v", validation.Errors))
				continue
			}

			// Determine action
			action := w.determineAction(importOffering)

			task := &WaldurIngestTask{
				WaldurUUID: o.UUID,
				Offering:   importOffering,
				Action:     action,
				Timestamp:  time.Now().UTC(),
			}

			select {
			case w.ingestQueue <- task:
				totalQueued++
			default:
				log.Printf("[waldur-ingest] queue full, skipping %s", o.UUID)
			}
		}

		// Update cursor
		w.state.UpdateCursor(&IngestCursor{
			Page:           page,
			PageSize:       w.cfg.PageSize,
			ProcessedCount: totalQueued,
			StartedAt:      startTime,
		})

		if len(offerings) < w.cfg.PageSize {
			break // Last page
		}
		page++
	}

	w.state.ResetCursor()
	log.Printf("[waldur-ingest] fetch complete: queued %d offerings (took %v)",
		totalQueued, time.Since(startTime))

	return nil
}

// shouldSkipOffering returns true if the offering should be skipped based on filters.
func (w *WaldurIngestWorker) shouldSkipOffering(o *waldur.Offering) bool {
	if w.cfg.SkipSharedOfferings && !o.Shared {
		return true
	}
	if w.cfg.SkipNonBillableOfferings && !o.Billable {
		return true
	}

	// Filter by category UUIDs
	if len(w.cfg.WaldurCategoryUUIDs) > 0 {
		found := false
		for _, catUUID := range w.cfg.WaldurCategoryUUIDs {
			if o.Category == catUUID {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Filter by offering types
	if len(w.cfg.WaldurOfferingTypes) > 0 {
		found := false
		for _, t := range w.cfg.WaldurOfferingTypes {
			if o.Type == t {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

// convertToImport converts a Waldur offering to the import format.
func (w *WaldurIngestWorker) convertToImport(o *waldur.Offering) *marketplace.WaldurOfferingImport {
	return &marketplace.WaldurOfferingImport{
		UUID:         o.UUID,
		Name:         o.Name,
		Description:  o.Description,
		Type:         o.Type,
		State:        o.State,
		CategoryUUID: o.Category,
		Shared:       o.Shared,
		Billable:     o.Billable,
		Created:      o.CreatedAt,
		Modified:     o.CreatedAt, // Use created as modified if not available
	}
}

// determineAction determines the ingestion action based on current state.
func (w *WaldurIngestWorker) determineAction(o *marketplace.WaldurOfferingImport) marketplace.IngestAction {
	record := w.state.GetRecord(o.UUID)
	if record == nil {
		return marketplace.IngestActionCreate
	}

	// Check for archived/deprecated
	if o.State == "Archived" {
		return marketplace.IngestActionDeprecate
	}

	// Check for drift
	newChecksum := o.IngestChecksum()
	if record.WaldurChecksum != newChecksum {
		return marketplace.IngestActionUpdate
	}

	return marketplace.IngestActionSkip
}

// processLoop processes ingestion tasks from the queue.
func (w *WaldurIngestWorker) processLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(w.doneCh)
			return
		case <-w.stopCh:
			close(w.doneCh)
			return
		case task := <-w.ingestQueue:
			w.processTask(ctx, task)
		}
	}
}

// processTask executes an ingestion task.
func (w *WaldurIngestWorker) processTask(ctx context.Context, task *WaldurIngestTask) {
	startTime := time.Now()

	w.metrics.mu.Lock()
	w.metrics.IngestsTotal++
	w.metrics.LastIngestTime = startTime
	w.metrics.mu.Unlock()

	w.promMetrics.IngestsTotal.Add(1)
	w.promMetrics.ActiveIngests.Add(1)
	defer w.promMetrics.ActiveIngests.Add(-1)

	var err error
	var chainOfferingID string
	var retryCount int

	// Get current retry count from state
	record := w.state.GetRecord(task.WaldurUUID)
	if record != nil {
		retryCount = record.RetryCount
	}

	opCtx, cancel := context.WithTimeout(ctx, w.cfg.OperationTimeout)
	defer cancel()

	switch task.Action {
	case marketplace.IngestActionCreate:
		chainOfferingID, err = w.ingestCreate(opCtx, task)
	case marketplace.IngestActionUpdate:
		chainOfferingID, err = w.ingestUpdate(opCtx, task)
	case marketplace.IngestActionDeprecate:
		err = w.ingestDeprecate(opCtx, task)
	case marketplace.IngestActionSkip:
		// No action needed
	}

	duration := time.Since(startTime)

	// Update prometheus timing metrics
	w.promMetrics.IngestDurationSum.Add(int64(duration))
	w.promMetrics.IngestDurationCount.Add(1)

	// Prepare checksum
	checksum := ""
	if task.Offering != nil {
		checksum = task.Offering.IngestChecksum()
	}

	// Create audit entry
	auditEntry := IngestAuditEntry{
		Timestamp:       startTime,
		WaldurUUID:      task.WaldurUUID,
		OfferingName:    task.Offering.Name,
		ChainOfferingID: chainOfferingID,
		Action:          string(task.Action),
		Success:         err == nil,
		Duration:        duration,
		RetryCount:      retryCount,
		ProviderAddress: w.cfg.ProviderAddress,
		Checksum:        checksum,
	}

	if err != nil {
		auditEntry.Error = err.Error()
		log.Printf("[waldur-ingest] %s failed for %s: %v", task.Action, task.WaldurUUID, err)

		deadLettered := w.state.MarkFailed(
			task.WaldurUUID,
			err.Error(),
			w.cfg.MaxRetries,
			time.Duration(w.cfg.RetryBackoffSeconds)*time.Second,
			time.Duration(w.cfg.MaxBackoffSeconds)*time.Second,
		)

		w.metrics.mu.Lock()
		w.metrics.IngestsFailed++
		w.metrics.mu.Unlock()

		w.promMetrics.IngestsFailed.Add(1)

		if deadLettered {
			w.metrics.mu.Lock()
			w.metrics.IngestsDeadLettered++
			w.metrics.mu.Unlock()

			w.promMetrics.IngestsDeadLettered.Add(1)
			w.promMetrics.DeadLetterSize.Add(1)

			w.auditLogger.LogDeadLetter(IngestDeadLetterAuditEntry{
				Timestamp:       time.Now().UTC(),
				WaldurUUID:      task.WaldurUUID,
				OfferingName:    task.Offering.Name,
				ProviderAddress: w.cfg.ProviderAddress,
				TotalAttempts:   retryCount + 1,
				LastError:       err.Error(),
				Action:          "dead_lettered",
			})
		}
	} else {
		version := uint64(1)
		if record != nil {
			version = record.ChainVersion + 1
		}

		w.state.MarkIngested(task.WaldurUUID, chainOfferingID, checksum, version)
		log.Printf("[waldur-ingest] %s succeeded for %s → %s (took %v)",
			task.Action, task.WaldurUUID, chainOfferingID, duration)

		w.metrics.mu.Lock()
		w.metrics.IngestsSuccessful++
		w.metrics.LastSuccessTime = time.Now().UTC()
		switch task.Action {
		case marketplace.IngestActionCreate:
			w.metrics.OfferingsCreated++
		case marketplace.IngestActionUpdate:
			w.metrics.OfferingsUpdated++
		case marketplace.IngestActionDeprecate:
			w.metrics.OfferingsDeprecated++
		}
		w.metrics.mu.Unlock()

		w.promMetrics.IngestsSuccessful.Add(1)
		switch task.Action {
		case marketplace.IngestActionCreate:
			w.promMetrics.OfferingsCreated.Add(1)
		case marketplace.IngestActionUpdate:
			w.promMetrics.OfferingsUpdated.Add(1)
		case marketplace.IngestActionDeprecate:
			w.promMetrics.OfferingsDeprecated.Add(1)
		}
	}

	// Log audit entry
	w.auditLogger.LogIngestAttempt(auditEntry)

	// Persist state
	if saveErr := w.stateStore.Save(w.state); saveErr != nil {
		log.Printf("[waldur-ingest] failed to save state: %v", saveErr)
	}
}

// ingestCreate creates a new on-chain offering from Waldur.
func (w *WaldurIngestWorker) ingestCreate(ctx context.Context, task *WaldurIngestTask) (string, error) {
	if task.Offering == nil {
		return "", fmt.Errorf("offering data required for create")
	}

	// Validate provider VEID if required
	if w.cfg.IngestConfig.MinIdentityScore > 0 {
		if err := w.submitter.ValidateProviderVEID(ctx, w.cfg.ProviderAddress, w.cfg.IngestConfig.MinIdentityScore); err != nil {
			return "", fmt.Errorf("provider VEID validation failed: %w", err)
		}
	}

	// Get next sequence number
	sequence, err := w.submitter.GetNextOfferingSequence(ctx, w.cfg.ProviderAddress)
	if err != nil {
		return "", fmt.Errorf("get next sequence: %w", err)
	}

	// Convert to on-chain offering
	offering := task.Offering.ToOffering(w.cfg.ProviderAddress, sequence, w.cfg.IngestConfig)

	// Validate offering
	if err := offering.Validate(); err != nil {
		return "", fmt.Errorf("invalid offering: %w", err)
	}

	// Submit to chain
	offeringID, err := w.submitter.CreateOffering(ctx, offering)
	if err != nil {
		return "", fmt.Errorf("create offering: %w", err)
	}

	return offeringID, nil
}

// ingestUpdate updates an existing on-chain offering.
func (w *WaldurIngestWorker) ingestUpdate(ctx context.Context, task *WaldurIngestTask) (string, error) {
	record := w.state.GetRecord(task.WaldurUUID)
	if record == nil || record.ChainOfferingID == "" {
		// No existing record, try create instead
		return w.ingestCreate(ctx, task)
	}

	if task.Offering == nil {
		return "", fmt.Errorf("offering data required for update")
	}

	// Convert to on-chain offering
	offering := task.Offering.ToOffering(w.cfg.ProviderAddress, record.ChainVersion, w.cfg.IngestConfig)

	// Update on chain
	if err := w.submitter.UpdateOffering(ctx, record.ChainOfferingID, offering); err != nil {
		return "", fmt.Errorf("update offering: %w", err)
	}

	return record.ChainOfferingID, nil
}

// ingestDeprecate deprecates an on-chain offering.
func (w *WaldurIngestWorker) ingestDeprecate(ctx context.Context, task *WaldurIngestTask) error {
	record := w.state.GetRecord(task.WaldurUUID)
	if record == nil || record.ChainOfferingID == "" {
		return nil // Nothing to deprecate
	}

	if err := w.submitter.DeprecateOffering(ctx, record.ChainOfferingID); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// Log but don't fail - offering might already be deprecated
		log.Printf("[waldur-ingest] deprecate warning for %s: %v", record.ChainOfferingID, err)
	}

	w.state.MarkDeprecated(task.WaldurUUID)
	return nil
}

// reconcileLoop runs periodic reconciliation.
func (w *WaldurIngestWorker) reconcileLoop(ctx context.Context) {
	if w.cfg.ReconcileIntervalSeconds <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(w.cfg.ReconcileIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.Reconcile(ctx); err != nil {
				log.Printf("[waldur-ingest] reconcile error: %v", err)
			}
		}
	}
}

// Reconcile performs full reconciliation between Waldur and chain state.
func (w *WaldurIngestWorker) Reconcile(ctx context.Context) error {
	log.Printf("[waldur-ingest] starting reconciliation")
	startTime := time.Now()

	// Get offerings that need re-ingestion
	needsIngest := w.state.NeedsIngestOfferings()
	offeringsQueued := 0
	driftDetected := 0

	for _, waldurUUID := range needsIngest {
		record := w.state.GetRecord(waldurUUID)
		if record == nil {
			continue
		}

		// Re-fetch from Waldur to check for drift
		if err := w.rateLimiter.Wait(ctx); err != nil {
			break
		}

		// Queue for re-ingestion
		task := &WaldurIngestTask{
			WaldurUUID: waldurUUID,
			Action:     marketplace.IngestActionUpdate,
			Timestamp:  time.Now().UTC(),
		}

		select {
		case w.ingestQueue <- task:
			offeringsQueued++
			if record.State == IngestRecordStateOutOfSync {
				driftDetected++
			}
		case <-ctx.Done():
			break
		default:
			log.Printf("[waldur-ingest] reconcile queue full, skipping %s", waldurUUID)
		}
	}

	duration := time.Since(startTime)

	w.state.RecordReconciliation()
	w.metrics.mu.Lock()
	w.metrics.ReconciliationsRun++
	w.metrics.LastReconcileTime = time.Now().UTC()
	w.metrics.DriftDetections += int64(driftDetected)
	w.metrics.mu.Unlock()

	w.promMetrics.ReconciliationsRun.Add(1)
	w.promMetrics.DriftDetected.Add(int64(driftDetected))
	w.promMetrics.QueueDepth.Store(int64(len(w.ingestQueue)))

	// Log reconciliation audit entry
	w.auditLogger.LogReconciliation(IngestReconciliationAuditEntry{
		Timestamp:        startTime,
		WaldurCustomer:   w.cfg.WaldurCustomerUUID,
		ProviderAddress:  w.cfg.ProviderAddress,
		OfferingsChecked: len(needsIngest),
		DriftDetected:    driftDetected,
		OfferingsQueued:  offeringsQueued,
		Duration:         duration,
	})

	log.Printf("[waldur-ingest] reconciliation queued %d offerings (drift: %d) (took %v)",
		offeringsQueued, driftDetected, duration)

	return nil
}

// QueueIngest manually queues an ingestion task.
func (w *WaldurIngestWorker) QueueIngest(waldurUUID string, offering *marketplace.WaldurOfferingImport, action marketplace.IngestAction) error {
	task := &WaldurIngestTask{
		WaldurUUID: waldurUUID,
		Offering:   offering,
		Action:     action,
		Timestamp:  time.Now().UTC(),
	}

	select {
	case w.ingestQueue <- task:
		return nil
	default:
		return fmt.Errorf("ingest queue full")
	}
}

// Metrics returns current worker metrics.
func (w *WaldurIngestWorker) Metrics() WaldurIngestWorkerMetrics {
	w.metrics.mu.RLock()
	defer w.metrics.mu.RUnlock()

	// Copy all fields except the mutex
	return WaldurIngestWorkerMetrics{
		IngestsTotal:        w.metrics.IngestsTotal,
		IngestsSuccessful:   w.metrics.IngestsSuccessful,
		IngestsFailed:       w.metrics.IngestsFailed,
		IngestsDeadLettered: w.metrics.IngestsDeadLettered,
		IngestsSkipped:      w.metrics.IngestsSkipped,
		DriftDetections:     w.metrics.DriftDetections,
		ReconciliationsRun:  w.metrics.ReconciliationsRun,
		LastIngestTime:      w.metrics.LastIngestTime,
		LastSuccessTime:     w.metrics.LastSuccessTime,
		LastReconcileTime:   w.metrics.LastReconcileTime,
		WorkerUptime:        w.metrics.WorkerUptime,
		OfferingsCreated:    w.metrics.OfferingsCreated,
		OfferingsUpdated:    w.metrics.OfferingsUpdated,
		OfferingsDeprecated: w.metrics.OfferingsDeprecated,
		QueueDepth:          len(w.ingestQueue),
	}
}

// State returns the current ingestion state.
func (w *WaldurIngestWorker) State() *WaldurIngestState {
	return w.state
}

// Stats returns summary statistics.
func (w *WaldurIngestWorker) Stats() IngestStats {
	return w.state.GetStats()
}

// ReprocessDeadLetter attempts to reprocess a dead-lettered offering.
func (w *WaldurIngestWorker) ReprocessDeadLetter(waldurUUID string) error {
	if w.state.ReprocessDeadLetter(waldurUUID) {
		w.promMetrics.DeadLetterSize.Add(-1)

		w.auditLogger.LogDeadLetter(IngestDeadLetterAuditEntry{
			Timestamp:       time.Now().UTC(),
			WaldurUUID:      waldurUUID,
			ProviderAddress: w.cfg.ProviderAddress,
			Action:          "reprocessed",
		})

		// Queue for re-ingestion
		return w.QueueIngest(waldurUUID, nil, marketplace.IngestActionUpdate)
	}
	return fmt.Errorf("offering %s not found in dead letter queue", waldurUUID)
}

// PrometheusMetrics returns the prometheus-compatible metrics.
func (w *WaldurIngestWorker) PrometheusMetrics() *WaldurIngestPrometheusMetrics {
	return w.promMetrics
}

// SetAuditLogger allows replacing the audit logger.
func (w *WaldurIngestWorker) SetAuditLogger(logger WaldurIngestAuditLogger) {
	w.auditLogger = logger
}
