// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-2D: Offering sync worker for automatic chain-to-Waldur synchronization.
// This file implements the sync worker that listens to offering events and
// synchronizes them to Waldur, with retry/backoff and reconciliation.
package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// OfferingSyncAuditLogger defines the interface for audit logging.
// Implementations can log to files, databases, or external services.
type OfferingSyncAuditLogger interface {
	// LogSyncAttempt logs a sync operation attempt
	LogSyncAttempt(entry OfferingSyncAuditEntry)
	// LogReconciliation logs a reconciliation run
	LogReconciliation(entry ReconciliationAuditEntry)
	// LogDeadLetter logs a dead-letter event
	LogDeadLetter(entry DeadLetterAuditEntry)
}

// OfferingSyncAuditEntry represents an audit log entry for a sync operation.
type OfferingSyncAuditEntry struct {
	Timestamp       time.Time              `json:"timestamp"`
	OfferingID      string                 `json:"offering_id"`
	WaldurUUID      string                 `json:"waldur_uuid,omitempty"`
	Action          SyncAction             `json:"action"`
	Success         bool                   `json:"success"`
	Duration        time.Duration          `json:"duration_ns"`
	Error           string                 `json:"error,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	ProviderAddress string                 `json:"provider_address"`
	Checksum        string                 `json:"checksum,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ReconciliationAuditEntry represents an audit log entry for reconciliation.
type ReconciliationAuditEntry struct {
	Timestamp        time.Time `json:"timestamp"`
	ProviderAddress  string    `json:"provider_address"`
	OfferingsChecked int       `json:"offerings_checked"`
	DriftDetected    int       `json:"drift_detected"`
	OfferingsQueued  int       `json:"offerings_queued"`
	Duration         time.Duration `json:"duration_ns"`
	Error            string    `json:"error,omitempty"`
}

// DeadLetterAuditEntry represents an audit log entry for dead-letter events.
type DeadLetterAuditEntry struct {
	Timestamp       time.Time `json:"timestamp"`
	OfferingID      string    `json:"offering_id"`
	ProviderAddress string    `json:"provider_address"`
	TotalAttempts   int       `json:"total_attempts"`
	LastError       string    `json:"last_error"`
	Action          string    `json:"action"` // "dead_lettered" or "reprocessed"
}

// DefaultAuditLogger is a structured JSON audit logger that writes to standard log.
type DefaultAuditLogger struct {
	prefix string
}

// NewDefaultAuditLogger creates a new default audit logger.
func NewDefaultAuditLogger(prefix string) *DefaultAuditLogger {
	if prefix == "" {
		prefix = "[offering-sync-audit]"
	}
	return &DefaultAuditLogger{prefix: prefix}
}

// LogSyncAttempt logs a sync operation to the standard logger as JSON.
func (l *DefaultAuditLogger) LogSyncAttempt(entry OfferingSyncAuditEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s sync: failed to marshal entry: %v", l.prefix, err)
		return
	}
	log.Printf("%s sync: %s", l.prefix, string(data))
}

// LogReconciliation logs a reconciliation run to the standard logger as JSON.
func (l *DefaultAuditLogger) LogReconciliation(entry ReconciliationAuditEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s reconcile: failed to marshal entry: %v", l.prefix, err)
		return
	}
	log.Printf("%s reconcile: %s", l.prefix, string(data))
}

// LogDeadLetter logs a dead-letter event to the standard logger as JSON.
func (l *DefaultAuditLogger) LogDeadLetter(entry DeadLetterAuditEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s dead_letter: failed to marshal entry: %v", l.prefix, err)
		return
	}
	log.Printf("%s dead_letter: %s", l.prefix, string(data))
}

// OfferingSyncPrometheusMetrics provides Prometheus-compatible metric counters.
// These can be registered with prometheus.MustRegister() in the main application.
type OfferingSyncPrometheusMetrics struct {
	// Sync operation counters
	SyncsTotal        atomic.Int64
	SyncsSuccessful   atomic.Int64
	SyncsFailed       atomic.Int64
	SyncsDeadLettered atomic.Int64

	// Event counters
	EventsReceived  atomic.Int64
	EventsProcessed atomic.Int64
	EventsDropped   atomic.Int64

	// Reconciliation counters
	ReconciliationsRun atomic.Int64
	DriftDetected      atomic.Int64

	// Timing histograms (simplified - in production use prometheus.Histogram)
	SyncDurationSum   atomic.Int64
	SyncDurationCount atomic.Int64

	// Current state
	QueueDepth     atomic.Int64
	ActiveSyncs    atomic.Int64
	DeadLetterSize atomic.Int64
}

// OfferingSyncWorkerConfig configures the offering sync worker.
type OfferingSyncWorkerConfig struct {
	// Enabled toggles the sync worker.
	Enabled bool

	// ProviderAddress is the provider's on-chain address.
	ProviderAddress string

	// CometRPC is the CometBFT RPC endpoint.
	CometRPC string

	// CometWS is the CometBFT WebSocket path.
	CometWS string

	// SubscriberID identifies this subscriber.
	SubscriberID string

	// EventBuffer is the event channel buffer size.
	EventBuffer int

	// SyncIntervalSeconds is how often to run reconciliation.
	SyncIntervalSeconds int64

	// ReconcileOnStartup triggers full reconciliation on start.
	ReconcileOnStartup bool

	// MaxRetries is the max retry attempts before dead-lettering.
	MaxRetries int

	// RetryBackoffSeconds is the base backoff duration.
	RetryBackoffSeconds int64

	// MaxBackoffSeconds is the maximum backoff duration.
	MaxBackoffSeconds int64

	// StateFilePath is the path to persist sync state.
	StateFilePath string

	// WaldurCustomerUUID is the Waldur customer/organization UUID.
	WaldurCustomerUUID string

	// WaldurCategoryMap maps offering categories to Waldur category UUIDs.
	WaldurCategoryMap map[string]string

	// CurrencyDenominator for price normalization.
	CurrencyDenominator uint64

	// OperationTimeout for Waldur API calls.
	OperationTimeout time.Duration
}

// DefaultOfferingSyncWorkerConfig returns sensible defaults.
func DefaultOfferingSyncWorkerConfig() OfferingSyncWorkerConfig {
	return OfferingSyncWorkerConfig{
		EventBuffer:         100,
		SyncIntervalSeconds: 300,
		ReconcileOnStartup:  true,
		MaxRetries:          5,
		RetryBackoffSeconds: 30,
		MaxBackoffSeconds:   3600,
		CurrencyDenominator: 1000000,
		OperationTimeout:    45 * time.Second,
		CometWS:             "/websocket",
		StateFilePath:       "data/offering_sync_state.json",
	}
}

// OfferingSyncWorker synchronizes on-chain offerings to Waldur.
type OfferingSyncWorker struct {
	cfg         OfferingSyncWorkerConfig
	syncConfig  marketplace.OfferingSyncConfig
	marketplace *waldur.MarketplaceClient
	stateStore  *OfferingSyncStateStore
	state       *OfferingSyncState
	rpcClient   *rpchttp.HTTP

	mu            sync.RWMutex
	running       bool
	stopCh        chan struct{}
	doneCh        chan struct{}
	syncQueue     chan *OfferingSyncTask
	metrics       *OfferingSyncWorkerMetrics
	auditLogger   OfferingSyncAuditLogger
	promMetrics   *OfferingSyncPrometheusMetrics
}

// OfferingSyncTask represents a sync task to execute.
type OfferingSyncTask struct {
	OfferingID string
	Action     SyncAction
	Offering   *marketplace.Offering
	Timestamp  time.Time
}

// SyncAction represents the sync action to perform.
type SyncAction string

const (
	SyncActionCreate  SyncAction = "create"
	SyncActionUpdate  SyncAction = "update"
	SyncActionDisable SyncAction = "disable"
)

// OfferingSyncWorkerMetrics tracks worker metrics.
type OfferingSyncWorkerMetrics struct {
	mu                  sync.RWMutex
	SyncsTotal          int64
	SyncsSuccessful     int64
	SyncsFailed         int64
	SyncsDeadLettered   int64
	DriftDetections     int64
	ReconciliationsRun  int64
	LastSyncTime        time.Time
	LastSuccessTime     time.Time
	LastReconcileTime   time.Time
	WorkerUptime        time.Time
	EventsReceived      int64
	EventsProcessed     int64
	QueueDepth          int
	AverageSyncDuration time.Duration
}

// OfferingEventPayload represents an offering event from the chain.
type OfferingEventPayload struct {
	OfferingID      string `json:"offering_id"`
	ProviderAddress string `json:"provider_address"`
	Name            string `json:"name,omitempty"`
	Category        string `json:"category,omitempty"`
	State           string `json:"state,omitempty"`
	Version         uint64 `json:"version"`
}

// NewOfferingSyncWorker creates a new offering sync worker.
func NewOfferingSyncWorker(
	cfg OfferingSyncWorkerConfig,
	marketplaceClient *waldur.MarketplaceClient,
) (*OfferingSyncWorker, error) {
	return NewOfferingSyncWorkerWithLogger(cfg, marketplaceClient, nil)
}

// NewOfferingSyncWorkerWithLogger creates a new offering sync worker with a custom audit logger.
func NewOfferingSyncWorkerWithLogger(
	cfg OfferingSyncWorkerConfig,
	marketplaceClient *waldur.MarketplaceClient,
	auditLogger OfferingSyncAuditLogger,
) (*OfferingSyncWorker, error) {
	if marketplaceClient == nil {
		return nil, fmt.Errorf("marketplace client is required")
	}

	// Use default audit logger if none provided
	if auditLogger == nil {
		auditLogger = NewDefaultAuditLogger("[offering-sync-audit]")
	}

	// Build sync config from worker config
	syncConfig := marketplace.DefaultOfferingSyncConfig()
	syncConfig.Enabled = cfg.Enabled
	syncConfig.SyncIntervalSeconds = cfg.SyncIntervalSeconds
	syncConfig.ReconcileOnStartup = cfg.ReconcileOnStartup
	syncConfig.MaxRetries = cfg.MaxRetries
	syncConfig.RetryBackoffSeconds = cfg.RetryBackoffSeconds
	syncConfig.MaxBackoffSeconds = cfg.MaxBackoffSeconds
	syncConfig.WaldurCustomerUUID = cfg.WaldurCustomerUUID
	syncConfig.WaldurCategoryMap = cfg.WaldurCategoryMap
	syncConfig.CurrencyDenominator = cfg.CurrencyDenominator

	stateStore := NewOfferingSyncStateStore(cfg.StateFilePath)

	return &OfferingSyncWorker{
		cfg:         cfg,
		syncConfig:  syncConfig,
		marketplace: marketplaceClient,
		stateStore:  stateStore,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		syncQueue:   make(chan *OfferingSyncTask, cfg.EventBuffer),
		metrics: &OfferingSyncWorkerMetrics{
			WorkerUptime: time.Now().UTC(),
		},
		auditLogger: auditLogger,
		promMetrics: &OfferingSyncPrometheusMetrics{},
	}, nil
}

// Start starts the offering sync worker.
func (w *OfferingSyncWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.running = true
	w.mu.Unlock()

	// Load persisted state
	state, err := w.stateStore.Load(w.cfg.ProviderAddress)
	if err != nil {
		return fmt.Errorf("load sync state: %w", err)
	}
	w.state = state

	// Connect to CometBFT
	if w.cfg.CometRPC != "" {
		rpc, err := rpchttp.New(w.cfg.CometRPC, w.cfg.CometWS)
		if err != nil {
			return fmt.Errorf("create rpc client: %w", err)
		}
		if err := rpc.Start(); err != nil {
			return fmt.Errorf("start rpc client: %w", err)
		}
		w.rpcClient = rpc
	}

	log.Printf("[offering-sync] started for provider %s", w.cfg.ProviderAddress)

	// Start background goroutines
	go w.eventLoop(ctx)
	go w.syncLoop(ctx)
	go w.reconcileLoop(ctx)

	// Initial reconciliation if enabled
	if w.cfg.ReconcileOnStartup {
		go func() {
			time.Sleep(5 * time.Second) // Wait for system to stabilize
			if err := w.Reconcile(ctx); err != nil {
				log.Printf("[offering-sync] initial reconcile failed: %v", err)
			}
		}()
	}

	return nil
}

// Stop stops the offering sync worker.
func (w *OfferingSyncWorker) Stop() error {
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
		log.Printf("[offering-sync] shutdown timeout")
	}

	// Stop RPC client
	if w.rpcClient != nil {
		if err := w.rpcClient.Stop(); err != nil {
			log.Printf("[offering-sync] rpc client stop error: %v", err)
		}
	}

	// Save final state
	if err := w.stateStore.Save(w.state); err != nil {
		log.Printf("[offering-sync] failed to save final state: %v", err)
	}

	log.Printf("[offering-sync] stopped")
	return nil
}

// eventLoop subscribes to offering events from the chain.
func (w *OfferingSyncWorker) eventLoop(ctx context.Context) {
	if w.rpcClient == nil {
		log.Printf("[offering-sync] no RPC client, skipping event subscription")
		return
	}

	query := w.buildOfferingEventQuery()
	subscriberID := w.cfg.SubscriberID
	if subscriberID == "" {
		subscriberID = fmt.Sprintf("offering-sync-%d", time.Now().UnixNano())
	}

	sub, err := w.rpcClient.Subscribe(ctx, subscriberID, query, w.cfg.EventBuffer)
	if err != nil {
		log.Printf("[offering-sync] subscribe failed: %v", err)
		return
	}

	log.Printf("[offering-sync] subscribed to offering events with query: %s", query)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case msg := <-sub:
			w.metrics.mu.Lock()
			w.metrics.EventsReceived++
			w.metrics.mu.Unlock()
			w.promMetrics.EventsReceived.Add(1)

			data, ok := msg.Data.(tmtypes.EventDataTx)
			if !ok {
				continue
			}

			w.processChainEvents(ctx, data)
		}
	}
}

// buildOfferingEventQuery builds the CometBFT subscription query for offering events.
func (w *OfferingSyncWorker) buildOfferingEventQuery() string {
	eventTypes := []string{
		string(marketplace.EventOfferingCreated),
		string(marketplace.EventOfferingUpdated),
		string(marketplace.EventOfferingTerminated),
	}

	parts := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		parts = append(parts, fmt.Sprintf("marketplace_event.event_type='%s'", eventType))
	}

	return fmt.Sprintf("tm.event='Tx' AND (%s)", strings.Join(parts, " OR "))
}

// processChainEvents extracts and queues offering events from a transaction.
func (w *OfferingSyncWorker) processChainEvents(ctx context.Context, data tmtypes.EventDataTx) {
	envelopes, err := ExtractMarketplaceEvents(data.Result.Events)
	if err != nil {
		log.Printf("[offering-sync] extract events error: %v", err)
		return
	}

	for _, envelope := range envelopes {
		// Filter to offering events for this provider
		switch marketplace.MarketplaceEventType(envelope.EventType) {
		case marketplace.EventOfferingCreated,
			marketplace.EventOfferingUpdated,
			marketplace.EventOfferingTerminated:

			var payload OfferingEventPayload
			if err := json.Unmarshal([]byte(envelope.PayloadJSON), &payload); err != nil {
				log.Printf("[offering-sync] decode payload error: %v", err)
				continue
			}

			// Only process events for this provider
			if !strings.EqualFold(payload.ProviderAddress, w.cfg.ProviderAddress) {
				continue
			}

			// Queue the sync task
			action := w.eventTypeToAction(marketplace.MarketplaceEventType(envelope.EventType))
			task := &OfferingSyncTask{
				OfferingID: payload.OfferingID,
				Action:     action,
				Timestamp:  time.Now().UTC(),
			}

			select {
			case w.syncQueue <- task:
				w.metrics.mu.Lock()
				w.metrics.EventsProcessed++
				w.metrics.mu.Unlock()
				w.promMetrics.EventsProcessed.Add(1)
				log.Printf("[offering-sync] queued %s for offering %s", action, payload.OfferingID)
			default:
				w.promMetrics.EventsDropped.Add(1)
				log.Printf("[offering-sync] queue full, dropping event for %s", payload.OfferingID)
			}
		}
	}
}

// eventTypeToAction converts an event type to a sync action.
func (w *OfferingSyncWorker) eventTypeToAction(eventType marketplace.MarketplaceEventType) SyncAction {
	switch eventType {
	case marketplace.EventOfferingCreated:
		return SyncActionCreate
	case marketplace.EventOfferingUpdated:
		return SyncActionUpdate
	case marketplace.EventOfferingTerminated:
		return SyncActionDisable
	default:
		return SyncActionUpdate
	}
}

// syncLoop processes sync tasks from the queue.
func (w *OfferingSyncWorker) syncLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(w.doneCh)
			return
		case <-w.stopCh:
			close(w.doneCh)
			return
		case task := <-w.syncQueue:
			w.processTask(ctx, task)
		}
	}
}

// processTask executes a sync task.
func (w *OfferingSyncWorker) processTask(ctx context.Context, task *OfferingSyncTask) {
	startTime := time.Now()

	w.metrics.mu.Lock()
	w.metrics.SyncsTotal++
	w.metrics.LastSyncTime = startTime
	w.metrics.mu.Unlock()

	// Update prometheus metrics
	w.promMetrics.SyncsTotal.Add(1)
	w.promMetrics.ActiveSyncs.Add(1)
	defer w.promMetrics.ActiveSyncs.Add(-1)

	var err error
	var waldurUUID string
	var retryCount int

	// Get current retry count from state
	record := w.state.GetRecord(task.OfferingID)
	if record != nil {
		retryCount = record.RetryCount
	}

	opCtx, cancel := context.WithTimeout(ctx, w.cfg.OperationTimeout)
	defer cancel()

	switch task.Action {
	case SyncActionCreate:
		waldurUUID, err = w.syncCreate(opCtx, task)
	case SyncActionUpdate:
		waldurUUID, err = w.syncUpdate(opCtx, task)
	case SyncActionDisable:
		err = w.syncDisable(opCtx, task)
	}

	duration := time.Since(startTime)

	// Update prometheus timing metrics
	w.promMetrics.SyncDurationSum.Add(int64(duration))
	w.promMetrics.SyncDurationCount.Add(1)

	// Prepare checksum for audit entry
	checksum := ""
	if task.Offering != nil {
		checksum = task.Offering.SyncChecksum()
	}

	// Create audit entry
	auditEntry := OfferingSyncAuditEntry{
		Timestamp:       startTime,
		OfferingID:      task.OfferingID,
		WaldurUUID:      waldurUUID,
		Action:          task.Action,
		Success:         err == nil,
		Duration:        duration,
		RetryCount:      retryCount,
		ProviderAddress: w.cfg.ProviderAddress,
		Checksum:        checksum,
	}

	if err != nil {
		auditEntry.Error = err.Error()
		log.Printf("[offering-sync] %s failed for %s: %v", task.Action, task.OfferingID, err)

		deadLettered := w.state.MarkFailed(
			task.OfferingID,
			err.Error(),
			w.cfg.MaxRetries,
			time.Duration(w.cfg.RetryBackoffSeconds)*time.Second,
			time.Duration(w.cfg.MaxBackoffSeconds)*time.Second,
		)

		w.metrics.mu.Lock()
		w.metrics.SyncsFailed++
		w.metrics.mu.Unlock()

		w.promMetrics.SyncsFailed.Add(1)

		// Log dead-letter event if this pushed it to dead-letter
		if deadLettered {
			w.metrics.mu.Lock()
			w.metrics.SyncsDeadLettered++
			w.metrics.mu.Unlock()

			w.promMetrics.SyncsDeadLettered.Add(1)
			w.promMetrics.DeadLetterSize.Add(1)

			w.auditLogger.LogDeadLetter(DeadLetterAuditEntry{
				Timestamp:       time.Now().UTC(),
				OfferingID:      task.OfferingID,
				ProviderAddress: w.cfg.ProviderAddress,
				TotalAttempts:   retryCount + 1,
				LastError:       err.Error(),
				Action:          "dead_lettered",
			})
		}
	} else {
		version := uint64(1)
		w.state.MarkSynced(task.OfferingID, waldurUUID, checksum, version)
		log.Printf("[offering-sync] %s succeeded for %s → %s (took %v)",
			task.Action, task.OfferingID, waldurUUID, duration)

		w.metrics.mu.Lock()
		w.metrics.SyncsSuccessful++
		w.metrics.LastSuccessTime = time.Now().UTC()
		w.metrics.mu.Unlock()

		w.promMetrics.SyncsSuccessful.Add(1)
	}

	// Log audit entry
	w.auditLogger.LogSyncAttempt(auditEntry)

	// Persist state after each sync
	if saveErr := w.stateStore.Save(w.state); saveErr != nil {
		log.Printf("[offering-sync] failed to save state: %v", saveErr)
	}
}

// syncCreate creates a new offering in Waldur.
func (w *OfferingSyncWorker) syncCreate(ctx context.Context, task *OfferingSyncTask) (string, error) {
	if task.Offering == nil {
		return "", fmt.Errorf("offering data required for create")
	}

	// Check if offering already exists in Waldur
	existing, err := w.marketplace.GetOfferingByBackendID(ctx, task.OfferingID)
	if err == nil && existing != nil {
		// Already exists, update instead
		return existing.UUID, w.syncUpdateExisting(ctx, task, existing.UUID)
	}
	if err != nil && !errors.Is(err, waldur.ErrNotFound) {
		return "", fmt.Errorf("check existing: %w", err)
	}

	// Convert to Waldur create request
	createReq := task.Offering.ToWaldurCreate(w.syncConfig)

	offering, err := w.marketplace.CreateOffering(ctx, waldur.CreateOfferingRequest{
		Name:         createReq.Name,
		Description:  createReq.Description,
		Type:         createReq.Type,
		State:        createReq.State,
		CategoryUUID: createReq.CategoryUUID,
		CustomerUUID: createReq.CustomerUUID,
		Shared:       createReq.Shared,
		Billable:     createReq.Billable,
		BackendID:    createReq.BackendID,
		Attributes:   createReq.Attributes,
	})
	if err != nil {
		return "", fmt.Errorf("create offering: %w", err)
	}

	return offering.UUID, nil
}

// syncUpdate updates an existing offering in Waldur.
func (w *OfferingSyncWorker) syncUpdate(ctx context.Context, task *OfferingSyncTask) (string, error) {
	// Get the Waldur UUID from state
	record := w.state.GetRecord(task.OfferingID)
	if record == nil || record.WaldurUUID == "" {
		// Try to find by backend ID
		existing, err := w.marketplace.GetOfferingByBackendID(ctx, task.OfferingID)
		if err != nil {
			if errors.Is(err, waldur.ErrNotFound) {
				// Doesn't exist, create instead
				return w.syncCreate(ctx, task)
			}
			return "", fmt.Errorf("lookup offering: %w", err)
		}
		return existing.UUID, w.syncUpdateExisting(ctx, task, existing.UUID)
	}

	return record.WaldurUUID, w.syncUpdateExisting(ctx, task, record.WaldurUUID)
}

// syncUpdateExisting updates an offering with a known Waldur UUID.
func (w *OfferingSyncWorker) syncUpdateExisting(ctx context.Context, task *OfferingSyncTask, waldurUUID string) error {
	if task.Offering == nil {
		return fmt.Errorf("offering data required for update")
	}

	updateReq := task.Offering.ToWaldurUpdate(w.syncConfig)

	_, err := w.marketplace.UpdateOffering(ctx, waldurUUID, waldur.UpdateOfferingRequest{
		Name:        updateReq.Name,
		Description: updateReq.Description,
		Attributes:  updateReq.Attributes,
	})
	if err != nil {
		return fmt.Errorf("update offering: %w", err)
	}

	// Handle state changes
	waldurState := marketplace.WaldurOfferingState[task.Offering.State]
	if waldurState != "" {
		action := w.stateToAction(waldurState)
		if action != "" {
			if err := w.marketplace.SetOfferingState(ctx, waldurUUID, action); err != nil {
				log.Printf("[offering-sync] set state %s failed: %v", action, err)
				// Don't fail the whole update for state change failure
			}
		}
	}

	return nil
}

// stateToAction converts a Waldur state to an action.
func (w *OfferingSyncWorker) stateToAction(state string) string {
	switch state {
	case "Active":
		return "activate"
	case "Paused":
		return "pause"
	case "Archived":
		return "archive"
	default:
		return ""
	}
}

// syncDisable disables an offering in Waldur.
func (w *OfferingSyncWorker) syncDisable(ctx context.Context, task *OfferingSyncTask) error {
	record := w.state.GetRecord(task.OfferingID)
	if record == nil || record.WaldurUUID == "" {
		// Try to find by backend ID
		existing, err := w.marketplace.GetOfferingByBackendID(ctx, task.OfferingID)
		if err != nil {
			if errors.Is(err, waldur.ErrNotFound) {
				return nil // Already gone, nothing to do
			}
			return fmt.Errorf("lookup offering: %w", err)
		}
		return w.marketplace.SetOfferingState(ctx, existing.UUID, "archive")
	}

	return w.marketplace.SetOfferingState(ctx, record.WaldurUUID, "archive")
}

// reconcileLoop runs periodic reconciliation.
func (w *OfferingSyncWorker) reconcileLoop(ctx context.Context) {
	if w.cfg.SyncIntervalSeconds <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(w.cfg.SyncIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.Reconcile(ctx); err != nil {
				log.Printf("[offering-sync] reconcile error: %v", err)
			}
		}
	}
}

// Reconcile performs full reconciliation between chain and Waldur.
func (w *OfferingSyncWorker) Reconcile(ctx context.Context) error {
	log.Printf("[offering-sync] starting reconciliation")
	startTime := time.Now()

	// Get offerings that need syncing
	needsSync := w.state.NeedsSyncOfferings()
	offeringsQueued := 0

	for _, offeringID := range needsSync {
		task := &OfferingSyncTask{
			OfferingID: offeringID,
			Action:     SyncActionUpdate,
			Timestamp:  time.Now().UTC(),
		}

		select {
		case w.syncQueue <- task:
			offeringsQueued++
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Printf("[offering-sync] reconcile queue full, skipping %s", offeringID)
		}
	}

	duration := time.Since(startTime)

	w.state.RecordReconciliation()
	w.metrics.mu.Lock()
	w.metrics.ReconciliationsRun++
	w.metrics.LastReconcileTime = time.Now().UTC()
	w.metrics.mu.Unlock()

	// Update prometheus metrics
	w.promMetrics.ReconciliationsRun.Add(1)
	w.promMetrics.DriftDetected.Add(int64(len(needsSync)))
	w.promMetrics.QueueDepth.Store(int64(len(w.syncQueue)))

	// Log reconciliation audit entry
	w.auditLogger.LogReconciliation(ReconciliationAuditEntry{
		Timestamp:        startTime,
		ProviderAddress:  w.cfg.ProviderAddress,
		OfferingsChecked: len(w.state.Records),
		DriftDetected:    len(needsSync),
		OfferingsQueued:  offeringsQueued,
		Duration:         duration,
	})

	log.Printf("[offering-sync] reconciliation queued %d offerings (took %v)",
		len(needsSync), duration)

	return nil
}

// QueueSync manually queues a sync task.
func (w *OfferingSyncWorker) QueueSync(offeringID string, action SyncAction, offering *marketplace.Offering) error {
	task := &OfferingSyncTask{
		OfferingID: offeringID,
		Action:     action,
		Offering:   offering,
		Timestamp:  time.Now().UTC(),
	}

	select {
	case w.syncQueue <- task:
		return nil
	default:
		return fmt.Errorf("sync queue full")
	}
}

// Metrics returns a snapshot of current worker metrics.
func (w *OfferingSyncWorker) Metrics() OfferingSyncWorkerMetrics {
	w.metrics.mu.RLock()
	defer w.metrics.mu.RUnlock()

	return OfferingSyncWorkerMetrics{
		SyncsTotal:          w.metrics.SyncsTotal,
		SyncsSuccessful:     w.metrics.SyncsSuccessful,
		SyncsFailed:         w.metrics.SyncsFailed,
		SyncsDeadLettered:   w.metrics.SyncsDeadLettered,
		DriftDetections:     w.metrics.DriftDetections,
		ReconciliationsRun:  w.metrics.ReconciliationsRun,
		LastSyncTime:        w.metrics.LastSyncTime,
		LastSuccessTime:     w.metrics.LastSuccessTime,
		LastReconcileTime:   w.metrics.LastReconcileTime,
		WorkerUptime:        w.metrics.WorkerUptime,
		EventsReceived:      w.metrics.EventsReceived,
		EventsProcessed:     w.metrics.EventsProcessed,
		QueueDepth:          len(w.syncQueue),
		AverageSyncDuration: w.metrics.AverageSyncDuration,
	}
}

// State returns the current sync state.
func (w *OfferingSyncWorker) State() *OfferingSyncState {
	return w.state
}

// ReprocessDeadLetter attempts to reprocess a dead-lettered offering.
func (w *OfferingSyncWorker) ReprocessDeadLetter(offeringID string) error {
	if w.state.ReprocessDeadLetter(offeringID) {
		w.promMetrics.DeadLetterSize.Add(-1)

		// Log the reprocess event
		w.auditLogger.LogDeadLetter(DeadLetterAuditEntry{
			Timestamp:       time.Now().UTC(),
			OfferingID:      offeringID,
			ProviderAddress: w.cfg.ProviderAddress,
			Action:          "reprocessed",
		})

		return w.QueueSync(offeringID, SyncActionUpdate, nil)
	}
	return fmt.Errorf("offering %s not found in dead letter queue", offeringID)
}

// PrometheusMetrics returns the prometheus-compatible metrics for this worker.
func (w *OfferingSyncWorker) PrometheusMetrics() *OfferingSyncPrometheusMetrics {
	return w.promMetrics
}

// SetAuditLogger allows replacing the audit logger.
func (w *OfferingSyncWorker) SetAuditLogger(logger OfferingSyncAuditLogger) {
	w.auditLogger = logger
}
