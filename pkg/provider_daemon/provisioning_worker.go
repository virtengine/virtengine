package provider_daemon

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// ProvisioningConfig configures the provisioning worker.
type ProvisioningConfig struct {
	Enabled          bool
	ProviderAddress  string
	SubscriberID     string
	CometRPC         string
	CometWS          string
	EventQuery       string
	EventBuffer      int
	CallbackTTL      time.Duration
	StateFile        string
	CheckpointFile   string
	MaxRetries       int
	RetryBackoff     time.Duration
	MaxBackoff       time.Duration
	PollInterval     time.Duration
	RetryOnFailure   bool
	ProgressInterval time.Duration
}

// DefaultProvisioningConfig returns default provisioning config.
func DefaultProvisioningConfig() ProvisioningConfig {
	return ProvisioningConfig{
		EventBuffer:      100,
		CallbackTTL:      time.Hour,
		MaxRetries:       5,
		RetryBackoff:     10 * time.Second,
		MaxBackoff:       5 * time.Minute,
		PollInterval:     15 * time.Second,
		RetryOnFailure:   true,
		ProgressInterval: 2 * time.Minute,
	}
}

// Provisioner executes provisioning for a service type.
type Provisioner interface {
	Name() string
	CanHandle(serviceType marketplace.ServiceType) bool
	Provision(ctx context.Context, req ProvisioningRequest) (*ProvisioningResult, error)
}

// ProvisioningRequest contains provisioning inputs.
type ProvisioningRequest struct {
	AllocationID       string
	OfferingID         string
	ProviderAddress    string
	ServiceType        marketplace.ServiceType
	Specifications     map[string]string
	EncryptedConfigRef string
	ResourceID         string
}

// ProvisioningResult contains provisioning results.
type ProvisioningResult struct {
	State      marketplace.AllocationState
	Phase      marketplace.ProvisioningPhase
	Progress   uint8
	ResourceID string
	Endpoints  map[string]string
	Message    string
	ErrorCode  string
}

// ProvisioningWorker handles provisioning events and retries.
type ProvisioningWorker struct {
	cfg             ProvisioningConfig
	keyManager      *KeyManager
	callbackSink    CallbackSink
	provisioners    []Provisioner
	stateStore      *ProvisioningStateStore
	state           *ProvisioningState
	checkpoint      *EventCheckpointStore
	eventSubscriber *CometEventSubscriber

	mu sync.Mutex
}

// NewProvisioningWorker creates a new provisioning worker.
func NewProvisioningWorker(cfg ProvisioningConfig, keyManager *KeyManager, callbackSink CallbackSink, provisioners ...Provisioner) (*ProvisioningWorker, error) {
	if !cfg.Enabled {
		return &ProvisioningWorker{cfg: cfg}, nil
	}
	if cfg.ProviderAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}
	if keyManager == nil {
		return nil, fmt.Errorf("key manager is required")
	}
	if callbackSink == nil {
		return nil, fmt.Errorf("callback sink is required")
	}
	if len(provisioners) == 0 {
		return nil, fmt.Errorf("at least one provisioner is required")
	}
	if cfg.SubscriberID == "" {
		cfg.SubscriberID = fmt.Sprintf("provisioner-%d", time.Now().UnixNano())
	}
	if cfg.EventQuery == "" {
		cfg.EventQuery = defaultProvisioningEventQuery()
	}
	if cfg.CallbackTTL == 0 {
		cfg.CallbackTTL = time.Hour
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 10 * time.Second
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 5 * time.Minute
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 15 * time.Second
	}

	worker := &ProvisioningWorker{
		cfg:          cfg,
		keyManager:   keyManager,
		callbackSink: callbackSink,
		provisioners: provisioners,
		stateStore:   NewProvisioningStateStore(cfg.StateFile),
		checkpoint:   mustNewEventCheckpointStore(cfg.CheckpointFile),
	}

	state, err := worker.stateStore.Load()
	if err != nil {
		return nil, err
	}
	state.ProviderAddress = cfg.ProviderAddress
	worker.state = state

	subCfg := DefaultEventSubscriberConfig()
	subCfg.CometRPC = cfg.CometRPC
	subCfg.CometWS = cfg.CometWS
	subCfg.SubscriberID = cfg.SubscriberID
	subCfg.EventBuffer = cfg.EventBuffer
	subCfg.CheckpointStore = worker.checkpoint

	subscriber, err := NewCometEventSubscriber(subCfg)
	if err != nil {
		return nil, err
	}
	worker.eventSubscriber = subscriber

	return worker, nil
}

// Start begins the provisioning loop.
func (w *ProvisioningWorker) Start(ctx context.Context) error {
	if w == nil || !w.cfg.Enabled {
		return nil
	}
	if w.eventSubscriber == nil {
		return errors.New("event subscriber not configured")
	}

	events, err := w.eventSubscriber.Subscribe(ctx, w.cfg.SubscriberID, w.cfg.EventQuery)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = w.eventSubscriber.Close()
			return ctx.Err()
		case <-ticker.C:
			w.processDueTasks(ctx)
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if event.Sequence <= w.eventSubscriber.LastCheckpoint() {
				continue
			}
			if err := w.handleMarketplaceEvent(ctx, event); err != nil {
				log.Printf("[provisioning] event %s failed: %v", event.ID, err)
				continue
			}
			w.eventSubscriber.SetCheckpoint(event.Sequence)
		}
	}
}

func (w *ProvisioningWorker) handleMarketplaceEvent(ctx context.Context, event MarketplaceEvent) error {
	if event.Type != EventType(marketplace.EventProvisionRequested) {
		return nil
	}

	payload := marketplace.ProvisionRequestedEvent{}
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("marshal provision request: %w", err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("decode provision request: %w", err)
	}
	return w.handleProvisionRequested(ctx, payload)
}

func (w *ProvisioningWorker) handleProvisionRequested(ctx context.Context, event marketplace.ProvisionRequestedEvent) error {
	if !strings.EqualFold(event.ProviderAddress, w.cfg.ProviderAddress) {
		return nil
	}

	serviceType := marketplace.ServiceType(strings.ToLower(event.ServiceType))
	if serviceType == marketplace.ServiceTypeUnknown {
		serviceType = marketplace.ServiceTypeFromSpecs(event.Specifications)
	}

	task := w.getOrCreateTask(event, serviceType)
	w.saveState()

	return w.processTask(ctx, task)
}

func (w *ProvisioningWorker) getOrCreateTask(event marketplace.ProvisionRequestedEvent, serviceType marketplace.ServiceType) *ProvisioningTask {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state.Tasks == nil {
		w.state.Tasks = map[string]*ProvisioningTask{}
	}

	task := w.state.Tasks[event.AllocationID]
	if task == nil {
		now := time.Now().UTC()
		task = &ProvisioningTask{
			AllocationID:       event.AllocationID,
			OfferingID:         event.OfferingID,
			ProviderAddress:    event.ProviderAddress,
			ServiceType:        string(serviceType),
			EncryptedConfigRef: event.EncryptedConfigRef,
			Specifications:     event.Specifications,
			State:              marketplace.AllocationStateProvisioning.String(),
			Phase:              string(marketplace.ProvisioningPhaseRequested),
			Progress:           0,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		w.state.Tasks[event.AllocationID] = task
	} else {
		task.OfferingID = event.OfferingID
		task.EncryptedConfigRef = event.EncryptedConfigRef
		task.ProviderAddress = event.ProviderAddress
		if task.Specifications == nil && len(event.Specifications) > 0 {
			task.Specifications = event.Specifications
		}
		if task.ServiceType == "" {
			task.ServiceType = string(serviceType)
		}
		task.UpdatedAt = time.Now().UTC()
	}

	return task
}

func (w *ProvisioningWorker) processDueTasks(ctx context.Context) {
	w.mu.Lock()
	tasks := make([]*ProvisioningTask, 0, len(w.state.Tasks))
	for _, task := range w.state.Tasks {
		if task.CompletedAt != nil {
			continue
		}
		if task.NextAttemptAt != nil && time.Now().UTC().Before(*task.NextAttemptAt) {
			continue
		}
		tasks = append(tasks, task)
	}
	w.mu.Unlock()

	for _, task := range tasks {
		if err := w.processTask(ctx, task); err != nil {
			log.Printf("[provisioning] task %s failed: %v", task.AllocationID, err)
		}
	}
}

func (w *ProvisioningWorker) processTask(ctx context.Context, task *ProvisioningTask) error {
	if task == nil {
		return nil
	}

	provisioner := w.selectProvisioner(marketplace.ServiceType(task.ServiceType))
	if provisioner == nil {
		task.LastError = "no provisioner for service type"
		return w.markFailed(ctx, task, "no_provisioner")
	}

	req := ProvisioningRequest{
		AllocationID:       task.AllocationID,
		OfferingID:         task.OfferingID,
		ProviderAddress:    task.ProviderAddress,
		ServiceType:        marketplace.ServiceType(task.ServiceType),
		Specifications:     task.Specifications,
		EncryptedConfigRef: task.EncryptedConfigRef,
		ResourceID:         task.ResourceID,
	}

	result, err := provisioner.Provision(ctx, req)
	if err != nil {
		task.Attempts++
		task.LastError = err.Error()
		task.UpdatedAt = time.Now().UTC()
		if task.Attempts >= w.cfg.MaxRetries {
			return w.markFailed(ctx, task, "max_retries_exceeded")
		}
		task.NextAttemptAt = pointerToTime(time.Now().UTC().Add(backoffDuration(task.Attempts, w.cfg.RetryBackoff, w.cfg.MaxBackoff)))
		task.Phase = string(marketplace.ProvisioningPhaseProvisioning)
		task.State = marketplace.AllocationStateProvisioning.String()
		w.saveState()
		return w.sendProvisioningCallback(ctx, task, marketplace.AllocationStateProvisioning, marketplace.ProvisioningPhaseProvisioning, 0, err.Error(), "retrying")
	}

	if result == nil {
		return nil
	}

	task.ResourceID = result.ResourceID
	task.Endpoints = result.Endpoints
	task.Progress = result.Progress
	task.UpdatedAt = time.Now().UTC()

	switch result.State {
	case marketplace.AllocationStateActive:
		task.State = result.State.String()
		task.Phase = string(result.Phase)
		task.CompletedAt = pointerToTime(time.Now().UTC())
		task.NextAttemptAt = nil
		w.saveState()
		return w.sendProvisioningCallback(ctx, task, result.State, result.Phase, result.Progress, result.Message, result.ErrorCode)
	case marketplace.AllocationStateFailed, marketplace.AllocationStateTerminated:
		task.Attempts++
		task.LastError = result.Message
		if w.cfg.RetryOnFailure && task.Attempts < w.cfg.MaxRetries {
			task.State = marketplace.AllocationStateProvisioning.String()
			task.Phase = string(marketplace.ProvisioningPhaseProvisioning)
			task.NextAttemptAt = pointerToTime(time.Now().UTC().Add(backoffDuration(task.Attempts, w.cfg.RetryBackoff, w.cfg.MaxBackoff)))
			w.saveState()
			return w.sendProvisioningCallback(ctx, task, marketplace.AllocationStateProvisioning, marketplace.ProvisioningPhaseProvisioning, result.Progress, result.Message, result.ErrorCode)
		}
		task.State = result.State.String()
		task.Phase = string(result.Phase)
		task.CompletedAt = pointerToTime(time.Now().UTC())
		w.saveState()
		return w.sendProvisioningCallback(ctx, task, result.State, result.Phase, result.Progress, result.Message, result.ErrorCode)
	default:
		task.State = marketplace.AllocationStateProvisioning.String()
		task.Phase = string(result.Phase)
		task.NextAttemptAt = pointerToTime(time.Now().UTC().Add(w.cfg.PollInterval))
		w.saveState()
		return w.sendProvisioningCallback(ctx, task, marketplace.AllocationStateProvisioning, result.Phase, result.Progress, result.Message, result.ErrorCode)
	}
}

func (w *ProvisioningWorker) selectProvisioner(serviceType marketplace.ServiceType) Provisioner {
	for _, provisioner := range w.provisioners {
		if provisioner.CanHandle(serviceType) {
			return provisioner
		}
	}
	return nil
}

func (w *ProvisioningWorker) markFailed(ctx context.Context, task *ProvisioningTask, errorCode string) error {
	task.State = marketplace.AllocationStateFailed.String()
	task.Phase = string(marketplace.ProvisioningPhaseFailed)
	task.CompletedAt = pointerToTime(time.Now().UTC())
	w.saveState()
	return w.sendProvisioningCallback(ctx, task, marketplace.AllocationStateFailed, marketplace.ProvisioningPhaseFailed, task.Progress, task.LastError, errorCode)
}

func (w *ProvisioningWorker) sendProvisioningCallback(
	ctx context.Context,
	task *ProvisioningTask,
	state marketplace.AllocationState,
	phase marketplace.ProvisioningPhase,
	progress uint8,
	message string,
	errorCode string,
) error {
	callback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeProvision,
		task.ResourceID,
		marketplace.SyncTypeAllocation,
		task.AllocationID,
		time.Now().UTC(),
	)
	callback.SignerID = w.cfg.ProviderAddress
	callback.ExpiresAt = callback.Timestamp.Add(w.cfg.CallbackTTL)
	callback.Payload["state"] = state.String()
	callback.Payload["phase"] = string(phase)
	if progress > 0 {
		callback.Payload["progress"] = fmt.Sprintf("%d", progress)
	}
	if message != "" {
		callback.Payload["message"] = message
	}
	if errorCode != "" {
		callback.Payload["error_code"] = errorCode
	}
	if task.ResourceID != "" {
		callback.Payload["resource_id"] = task.ResourceID
	}
	if task.ServiceType != "" {
		callback.Payload["service_type"] = task.ServiceType
	}
	for key, value := range task.Endpoints {
		callback.Payload["endpoint_"+key] = value
	}

	sig, err := w.keyManager.Sign(callback.SigningPayload())
	if err != nil {
		return fmt.Errorf("sign callback: %w", err)
	}
	sigBytes, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	callback.Signature = sigBytes

	return w.callbackSink.Submit(ctx, callback)
}

func (w *ProvisioningWorker) saveState() {
	if w.stateStore == nil || w.state == nil {
		return
	}
	if err := w.stateStore.Save(w.state); err != nil {
		log.Printf("[provisioning] failed to save state: %v", err)
	}
}

func pointerToTime(value time.Time) *time.Time {
	return &value
}

func backoffDuration(attempt int, base, max time.Duration) time.Duration {
	if attempt <= 0 {
		return base
	}
	backoff := base
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if backoff >= max {
			return max
		}
	}
	if backoff > max {
		return max
	}
	return backoff
}

func defaultProvisioningEventQuery() string {
	eventTypes := []string{
		string(marketplace.EventProvisionRequested),
	}
	parts := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		parts = append(parts, fmt.Sprintf("marketplace_event.event_type='%s'", eventType))
	}
	return fmt.Sprintf("tm.event='Tx' AND (%s)", strings.Join(parts, " OR "))
}
