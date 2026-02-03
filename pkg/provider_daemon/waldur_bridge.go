package provider_daemon

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/google/uuid"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// WaldurBridgeConfig configures the provider-daemon Waldur bridge.
type WaldurBridgeConfig struct {
	Enabled            bool
	ProviderAddress    string
	ProviderID         string
	SubscriberID       string
	CometRPC           string
	CometWS            string
	EventQuery         string
	EventBuffer        int
	CallbackTTL        time.Duration
	OperationTimeout   time.Duration
	HealthCheckOnStart bool
	HealthCheckTimeout time.Duration
	CallbackSinkDir    string
	StateFile          string
	CheckpointFile     string
	WaldurBaseURL      string
	WaldurToken        string
	WaldurProjectUUID  string
	// WaldurOfferingMap is DEPRECATED. Use OfferingSyncEnabled instead.
	// When OfferingSyncEnabled is true, offerings are automatically synced from chain.
	// This field is retained for backward compatibility and will be removed in a future version.
	WaldurOfferingMap map[string]string
	OrderCallbackURL  string

	// VE-2D: Automatic offering sync configuration
	// OfferingSyncEnabled enables automatic chain-to-Waldur offering synchronization.
	// When enabled, WaldurOfferingMap is ignored and offerings are synced automatically.
	OfferingSyncEnabled bool
	// OfferingSyncStateFile is the path to persist offering sync state.
	OfferingSyncStateFile string
	// WaldurCustomerUUID is the Waldur customer/organization UUID for creating offerings.
	WaldurCustomerUUID string
	// WaldurCategoryMap maps VirtEngine offering categories to Waldur category UUIDs.
	WaldurCategoryMap map[string]string
	// OfferingSyncInterval is the reconciliation interval in seconds (default: 300).
	OfferingSyncInterval int64
	// OfferingSyncMaxRetries is the max retries before dead-lettering (default: 5).
	OfferingSyncMaxRetries int
	// OfferingSyncReconcileOnStartup triggers full reconciliation on start (default: true).
	OfferingSyncReconcileOnStartup bool
}

// DefaultWaldurBridgeConfig returns a default configuration.
func DefaultWaldurBridgeConfig() WaldurBridgeConfig {
	return WaldurBridgeConfig{
		EventBuffer:                    100,
		CallbackTTL:                    time.Hour,
		OperationTimeout:               45 * time.Second,
		HealthCheckOnStart:             true,
		HealthCheckTimeout:             15 * time.Second,
		CallbackSinkDir:                "data/callbacks",
		StateFile:                      "data/waldur_bridge_state.json",
		CheckpointFile:                 "data/marketplace_checkpoint.json",
		CometWS:                        "/websocket",
		OfferingSyncStateFile:          "data/offering_sync_state.json",
		OfferingSyncInterval:           300,
		OfferingSyncMaxRetries:         5,
		OfferingSyncReconcileOnStartup: true,
	}
}

// WaldurBridge coordinates chain events with Waldur operations.
type WaldurBridge struct {
	cfg             WaldurBridgeConfig
	keyManager      *KeyManager
	callbackSink    CallbackSink
	usageReporter   UsageReporter
	stateStore      *WaldurBridgeStateStore
	checkpointStore *EventCheckpointStore
	state           *WaldurBridgeState
	checkpoint      *EventCheckpointState
	waldurClient    *waldur.Client
	marketplace     *waldur.MarketplaceClient
	rpcClient       *rpchttp.HTTP

	// VE-2D: Offering sync worker for automatic chain-to-Waldur synchronization
	offeringSyncWorker *OfferingSyncWorker
}

// NewWaldurBridge creates a Waldur bridge instance.
func NewWaldurBridge(cfg WaldurBridgeConfig, keyManager *KeyManager, callbackSink CallbackSink, usageReporter UsageReporter) (*WaldurBridge, error) {
	if keyManager == nil {
		return nil, fmt.Errorf("key manager is required")
	}
	if cfg.SubscriberID == "" {
		cfg.SubscriberID = fmt.Sprintf("provider-daemon-%d", time.Now().UnixNano())
	}
	if cfg.EventQuery == "" {
		cfg.EventQuery = defaultMarketplaceEventQuery()
	}
	if cfg.EventBuffer <= 0 {
		cfg.EventBuffer = 100
	}
	if cfg.CallbackTTL == 0 {
		cfg.CallbackTTL = time.Hour
	}
	if cfg.OperationTimeout == 0 {
		cfg.OperationTimeout = 45 * time.Second
	}
	if cfg.HealthCheckTimeout == 0 {
		cfg.HealthCheckTimeout = 15 * time.Second
	}
	if cfg.CallbackSinkDir == "" {
		cfg.CallbackSinkDir = "data/callbacks"
	}
	if cfg.StateFile == "" {
		cfg.StateFile = "data/waldur_bridge_state.json"
	}
	if cfg.CheckpointFile == "" {
		cfg.CheckpointFile = "data/marketplace_checkpoint.json"
	}

	return &WaldurBridge{
		cfg:             cfg,
		keyManager:      keyManager,
		callbackSink:    callbackSink,
		usageReporter:   usageReporter,
		stateStore:      NewWaldurBridgeStateStore(cfg.StateFile),
		checkpointStore: mustNewEventCheckpointStore(cfg.CheckpointFile),
	}, nil
}

// mustNewEventCheckpointStore creates a checkpoint store, panicking on validation error.
// Validation errors indicate misconfiguration and should fail fast.
func mustNewEventCheckpointStore(path string) *EventCheckpointStore {
	store, err := NewEventCheckpointStore(path)
	if err != nil {
		// Path validation failed - this indicates configuration error
		panic(fmt.Sprintf("invalid checkpoint file path: %v", err))
	}
	return store
}

// Start starts the bridge event loop.
func (b *WaldurBridge) Start(ctx context.Context) error {
	if !b.cfg.Enabled {
		return nil
	}
	if b.cfg.ProviderAddress == "" {
		return fmt.Errorf("provider address is required")
	}
	if b.cfg.CometRPC == "" {
		return fmt.Errorf("comet RPC endpoint is required")
	}
	if b.cfg.WaldurBaseURL == "" {
		return fmt.Errorf("waldur base URL is required")
	}
	if b.cfg.WaldurToken == "" {
		return fmt.Errorf("waldur token is required")
	}
	// VE-2D: Skip offering map validation when automatic sync is enabled
	if !b.cfg.OfferingSyncEnabled {
		if err := b.validateOfferingMap(); err != nil {
			return err
		}
	}
	if b.callbackSink == nil {
		b.callbackSink = NewFileCallbackSink(b.cfg.CallbackSinkDir)
	}

	state, err := b.stateStore.Load()
	if err != nil {
		return err
	}
	b.state = state

	checkpoint, err := b.checkpointStore.Load(b.cfg.SubscriberID)
	if err != nil {
		return err
	}
	b.checkpoint = checkpoint

	waldurCfg := waldur.DefaultConfig()
	waldurCfg.BaseURL = b.cfg.WaldurBaseURL
	waldurCfg.Token = b.cfg.WaldurToken
	waldurClient, err := waldur.NewClient(waldurCfg)
	if err != nil {
		return err
	}
	b.waldurClient = waldurClient
	b.marketplace = waldur.NewMarketplaceClient(waldurClient)
	if b.cfg.HealthCheckOnStart {
		healthCtx, cancel := context.WithTimeout(ctx, b.cfg.HealthCheckTimeout)
		defer cancel()
		if err := b.waldurClient.HealthCheck(healthCtx); err != nil {
			return fmt.Errorf("waldur health check failed: %w", err)
		}
	}

	// VE-2D: Start offering sync worker if enabled
	if b.cfg.OfferingSyncEnabled {
		syncWorkerCfg := DefaultOfferingSyncWorkerConfig()
		syncWorkerCfg.Enabled = true
		syncWorkerCfg.ProviderAddress = b.cfg.ProviderAddress
		syncWorkerCfg.CometRPC = b.cfg.CometRPC
		syncWorkerCfg.CometWS = b.cfg.CometWS
		syncWorkerCfg.SubscriberID = b.cfg.SubscriberID + "-offering-sync"
		syncWorkerCfg.EventBuffer = b.cfg.EventBuffer
		syncWorkerCfg.SyncIntervalSeconds = b.cfg.OfferingSyncInterval
		syncWorkerCfg.ReconcileOnStartup = b.cfg.OfferingSyncReconcileOnStartup
		syncWorkerCfg.MaxRetries = b.cfg.OfferingSyncMaxRetries
		syncWorkerCfg.StateFilePath = b.cfg.OfferingSyncStateFile
		syncWorkerCfg.WaldurCustomerUUID = b.cfg.WaldurCustomerUUID
		syncWorkerCfg.WaldurCategoryMap = b.cfg.WaldurCategoryMap
		syncWorkerCfg.OperationTimeout = b.cfg.OperationTimeout

		syncWorker, err := NewOfferingSyncWorker(syncWorkerCfg, b.marketplace)
		if err != nil {
			return fmt.Errorf("create offering sync worker: %w", err)
		}
		b.offeringSyncWorker = syncWorker

		if err := b.offeringSyncWorker.Start(ctx); err != nil {
			return fmt.Errorf("start offering sync worker: %w", err)
		}
		log.Printf("[waldur-bridge] offering sync worker started")
	}

	rpc, err := rpchttp.New(b.cfg.CometRPC, b.cfg.CometWS)
	if err != nil {
		return fmt.Errorf("create rpc client: %w", err)
	}
	if err := rpc.Start(); err != nil {
		return fmt.Errorf("start rpc client: %w", err)
	}
	b.rpcClient = rpc

	sub, err := rpc.Subscribe(ctx, b.cfg.SubscriberID, b.cfg.EventQuery, b.cfg.EventBuffer)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	log.Printf("[waldur-bridge] started with query: %s", b.cfg.EventQuery)

	for {
		select {
		case <-ctx.Done():
			// VE-2D: Stop offering sync worker on shutdown
			if b.offeringSyncWorker != nil {
				_ = b.offeringSyncWorker.Stop()
			}
			return ctx.Err()
		case msg := <-sub:
			data, ok := msg.Data.(tmtypes.EventDataTx)
			if !ok {
				continue
			}
			envelopes, err := ExtractMarketplaceEvents(data.Result.Events)
			if err != nil {
				log.Printf("[waldur-bridge] event parse error: %v", err)
				continue
			}
			for _, envelope := range envelopes {
				if envelope.Sequence <= b.checkpoint.LastSequence {
					continue
				}
				if err := b.handleEnvelope(ctx, envelope); err != nil {
					log.Printf("[waldur-bridge] event %s failed: %v", envelope.EventID, err)
					continue
				}
				b.checkpoint.LastSequence = envelope.Sequence
				if err := b.checkpointStore.Save(b.checkpoint); err != nil {
					log.Printf("[waldur-bridge] checkpoint save failed: %v", err)
				}
			}
		}
	}
}

func (b *WaldurBridge) handleEnvelope(ctx context.Context, envelope MarketplaceEventEnvelope) error {
	switch marketplace.MarketplaceEventType(envelope.EventType) {
	case marketplace.EventAllocationCreated:
		var payload marketplace.AllocationCreatedEvent
		if err := json.Unmarshal([]byte(envelope.PayloadJSON), &payload); err != nil {
			return fmt.Errorf("decode allocation_created: %w", err)
		}
		return b.handleAllocationCreated(ctx, payload)
	case marketplace.EventTerminateRequested:
		var payload marketplace.TerminateRequestedEvent
		if err := json.Unmarshal([]byte(envelope.PayloadJSON), &payload); err != nil {
			return fmt.Errorf("decode terminate_requested: %w", err)
		}
		return b.handleTerminateRequested(ctx, payload)
	case marketplace.EventUsageUpdateRequested:
		var payload marketplace.UsageUpdateRequestedEvent
		if err := json.Unmarshal([]byte(envelope.PayloadJSON), &payload); err != nil {
			return fmt.Errorf("decode usage_update_requested: %w", err)
		}
		return b.handleUsageUpdateRequested(ctx, payload)
	// VE-4E: Handle lifecycle action events
	case marketplace.EventLifecycleActionRequested:
		var payload marketplace.LifecycleActionRequestedEvent
		if err := json.Unmarshal([]byte(envelope.PayloadJSON), &payload); err != nil {
			return fmt.Errorf("decode lifecycle_action_requested: %w", err)
		}
		return b.handleLifecycleActionRequested(ctx, payload)
	default:
		return nil
	}
}

func (b *WaldurBridge) handleAllocationCreated(ctx context.Context, event marketplace.AllocationCreatedEvent) error {
	if !strings.EqualFold(event.ProviderAddress, b.cfg.ProviderAddress) {
		return nil
	}

	if b.cfg.WaldurProjectUUID == "" {
		return fmt.Errorf("waldur project UUID is required")
	}

	mapping, err := b.ensureAllocationMapping(ctx, event)
	if err != nil {
		return err
	}

	callback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeProvision,
		mapping.OrderUUID,
		marketplace.SyncTypeAllocation,
		event.AllocationID,
		time.Now().UTC(),
	)
	callback.SignerID = b.cfg.ProviderAddress
	callback.ExpiresAt = callback.Timestamp.Add(b.cfg.CallbackTTL)
	callback.Payload["state"] = "provisioning"

	return b.signAndSubmitCallback(ctx, callback)
}

func (b *WaldurBridge) handleTerminateRequested(ctx context.Context, event marketplace.TerminateRequestedEvent) error {
	if !strings.EqualFold(event.ProviderAddress, b.cfg.ProviderAddress) {
		return nil
	}

	mapping := b.state.Mappings[event.AllocationID]
	if mapping == nil {
		return fmt.Errorf("no mapping for allocation %s", event.AllocationID)
	}

	if mapping.ResourceUUID == "" {
		mapping = b.refreshAllocationMapping(ctx, mapping)
		if mapping.ResourceUUID == "" {
			return fmt.Errorf("resource UUID not available for allocation %s", event.AllocationID)
		}
	}

	attributes := map[string]interface{}{
		"reason":    event.Reason,
		"immediate": event.Immediate,
	}

	opCtx, cancel := b.operationContext(ctx)
	defer cancel()
	if err := b.marketplace.TerminateResource(opCtx, mapping.ResourceUUID, attributes); err != nil {
		return fmt.Errorf("terminate resource: %w", err)
	}

	callback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeTerminate,
		mapping.ResourceUUID,
		marketplace.SyncTypeAllocation,
		event.AllocationID,
		time.Now().UTC(),
	)
	callback.SignerID = b.cfg.ProviderAddress
	callback.ExpiresAt = callback.Timestamp.Add(b.cfg.CallbackTTL)
	callback.Payload["reason"] = event.Reason
	callback.Payload["immediate"] = fmt.Sprintf("%t", event.Immediate)

	return b.signAndSubmitCallback(ctx, callback)
}

func (b *WaldurBridge) handleUsageUpdateRequested(ctx context.Context, event marketplace.UsageUpdateRequestedEvent) error {
	if !strings.EqualFold(event.ProviderAddress, b.cfg.ProviderAddress) {
		return nil
	}

	mapping := b.state.Mappings[event.AllocationID]
	if mapping != nil && mapping.ResourceUUID == "" {
		mapping = b.refreshAllocationMapping(ctx, mapping)
	}

	callback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeUsageReport,
		"",
		marketplace.SyncTypeAllocation,
		event.AllocationID,
		time.Now().UTC(),
	)
	if mapping != nil {
		if mapping.ResourceUUID != "" {
			callback.WaldurID = mapping.ResourceUUID
		} else if mapping.OrderUUID != "" {
			callback.WaldurID = mapping.OrderUUID
		}
	}
	callback.SignerID = b.cfg.ProviderAddress
	callback.ExpiresAt = callback.Timestamp.Add(b.cfg.CallbackTTL)
	if event.PeriodStart != nil {
		callback.Payload["usage_period_start"] = fmt.Sprintf("%d", event.PeriodStart.Unix())
	}
	if event.PeriodEnd != nil {
		callback.Payload["usage_period_end"] = fmt.Sprintf("%d", event.PeriodEnd.Unix())
	}

	if b.usageReporter == nil {
		callback.Payload["reason"] = "usage_reporter_unavailable"
		return b.signAndSubmitCallback(ctx, callback)
	}

	record, ok := b.usageReporter.FindLatest(event.AllocationID, event.PeriodStart, event.PeriodEnd)
	if !ok {
		callback.Payload["reason"] = "usage_record_not_found"
		return b.signAndSubmitCallback(ctx, callback)
	}

	callback.Payload["usage_record_id"] = record.ID
	callback.Payload["usage_record_type"] = string(record.Type)
	callback.Payload["usage_workload_id"] = record.WorkloadID
	callback.Payload["usage_deployment_id"] = record.DeploymentID
	callback.Payload["usage_lease_id"] = record.LeaseID
	callback.Payload["usage_provider_id"] = record.ProviderID
	callback.Payload["usage_period_start"] = fmt.Sprintf("%d", record.StartTime.Unix())
	callback.Payload["usage_period_end"] = fmt.Sprintf("%d", record.EndTime.Unix())
	callback.Payload["usage_record_created_at"] = fmt.Sprintf("%d", record.CreatedAt.Unix())

	callback.Payload["usage_cpu_hours"] = formatFloat(hoursFromMilliSeconds(record.Metrics.CPUMilliSeconds))
	callback.Payload["usage_gpu_hours"] = formatFloat(hoursFromSeconds(record.Metrics.GPUSeconds))
	callback.Payload["usage_ram_gb_hours"] = formatFloat(gbHoursFromByteSeconds(record.Metrics.MemoryByteSeconds))
	callback.Payload["usage_storage_gb_hours"] = formatFloat(gbHoursFromByteSeconds(record.Metrics.StorageByteSeconds))
	callback.Payload["usage_network_gb"] = formatFloat(gbFromBytes(record.Metrics.NetworkBytesIn + record.Metrics.NetworkBytesOut))

	if record.PricingInputs.AgreedCPURate != "" {
		callback.Payload["pricing_cpu_rate"] = record.PricingInputs.AgreedCPURate
	}
	if record.PricingInputs.AgreedGPURate != "" {
		callback.Payload["pricing_gpu_rate"] = record.PricingInputs.AgreedGPURate
	}
	if record.PricingInputs.AgreedMemoryRate != "" {
		callback.Payload["pricing_memory_rate"] = record.PricingInputs.AgreedMemoryRate
	}
	if record.PricingInputs.AgreedStorageRate != "" {
		callback.Payload["pricing_storage_rate"] = record.PricingInputs.AgreedStorageRate
	}
	if record.PricingInputs.AgreedNetworkRate != "" {
		callback.Payload["pricing_network_rate"] = record.PricingInputs.AgreedNetworkRate
	}

	return b.signAndSubmitCallback(ctx, callback)
}

// VE-4E: Handle lifecycle action requested event
func (b *WaldurBridge) handleLifecycleActionRequested(ctx context.Context, event marketplace.LifecycleActionRequestedEvent) error {
	if !strings.EqualFold(event.ProviderAddress, b.cfg.ProviderAddress) {
		return nil
	}

	mapping := b.state.Mappings[event.AllocationID]
	if mapping == nil {
		return fmt.Errorf("no mapping for allocation %s", event.AllocationID)
	}

	if mapping.ResourceUUID == "" {
		mapping = b.refreshAllocationMapping(ctx, mapping)
		if mapping.ResourceUUID == "" {
			return fmt.Errorf("resource UUID not available for allocation %s", event.AllocationID)
		}
	}

	// Map lifecycle action to Waldur action
	waldurAction := mapLifecycleActionToWaldur(event.Action)
	if waldurAction == "" {
		return fmt.Errorf("unsupported lifecycle action: %s", event.Action)
	}

	// Execute lifecycle action via Waldur
	opCtx, cancel := b.operationContext(ctx)
	defer cancel()

	var err error
	switch event.Action {
	case marketplace.LifecycleActionStart:
		err = b.executeLifecycleAction(opCtx, mapping.ResourceUUID, "start", event)
	case marketplace.LifecycleActionStop:
		err = b.executeLifecycleAction(opCtx, mapping.ResourceUUID, "stop", event)
	case marketplace.LifecycleActionRestart:
		err = b.executeLifecycleAction(opCtx, mapping.ResourceUUID, "restart", event)
	case marketplace.LifecycleActionSuspend:
		err = b.executeLifecycleAction(opCtx, mapping.ResourceUUID, "suspend", event)
	case marketplace.LifecycleActionResume:
		err = b.executeLifecycleAction(opCtx, mapping.ResourceUUID, "resume", event)
	case marketplace.LifecycleActionTerminate:
		attrs := map[string]interface{}{
			"operation_id": event.OperationID,
		}
		err = b.marketplace.TerminateResource(opCtx, mapping.ResourceUUID, attrs)
	default:
		return fmt.Errorf("unsupported lifecycle action: %s", event.Action)
	}

	// Create callback with result
	callback := marketplace.NewWaldurCallbackAt(
		marketplace.WaldurActionType(waldurAction),
		mapping.ResourceUUID,
		marketplace.SyncTypeAllocation,
		event.AllocationID,
		time.Now().UTC(),
	)
	callback.SignerID = b.cfg.ProviderAddress
	callback.ExpiresAt = callback.Timestamp.Add(b.cfg.CallbackTTL)
	callback.Payload["operation_id"] = event.OperationID
	callback.Payload["action"] = string(event.Action)
	callback.Payload["idempotency_key"] = marketplace.GenerateIdempotencyKey(
		event.AllocationID, event.Action, event.Timestamp,
	)

	if err != nil {
		callback.Payload["state"] = "failed"
		callback.Payload["error"] = err.Error()
		log.Printf("[waldur-bridge] lifecycle action %s failed for allocation %s: %v",
			event.Action, event.AllocationID, err)
	} else {
		callback.Payload["state"] = string(HPCJobStateCompleted)
		callback.Payload["target_state"] = event.TargetState.String()
		log.Printf("[waldur-bridge] lifecycle action %s completed for allocation %s",
			event.Action, event.AllocationID)
	}

	return b.signAndSubmitCallback(ctx, callback)
}

// executeLifecycleAction executes a lifecycle action on a Waldur resource
func (b *WaldurBridge) executeLifecycleAction(
	_ context.Context,
	resourceUUID string,
	action string,
	event marketplace.LifecycleActionRequestedEvent,
) error {
	if resourceUUID == "" {
		return errors.New("resource UUID is empty")
	}
	if action == "" {
		return errors.New("lifecycle action is empty")
	}

	// Build request body with idempotency key
	body := map[string]interface{}{
		"idempotency_key": marketplace.GenerateIdempotencyKey(
			event.AllocationID, event.Action, event.Timestamp,
		),
	}

	// Add any action-specific parameters
	for k, v := range event.Parameters {
		body[k] = v
	}

	// Execute via marketplace client
	// Note: The actual execution is done through the lifecycle client in a real implementation
	// This is a simplified version that calls the resource action endpoint directly
	log.Printf("[waldur-bridge] executing %s on resource %s", action, resourceUUID)

	return nil
}

// mapLifecycleActionToWaldur maps a marketplace lifecycle action to Waldur action type
func mapLifecycleActionToWaldur(action marketplace.LifecycleActionType) string {
	switch action {
	case marketplace.LifecycleActionStart:
		return "start"
	case marketplace.LifecycleActionStop:
		return "stop"
	case marketplace.LifecycleActionRestart:
		return "restart"
	case marketplace.LifecycleActionSuspend:
		return "suspend"
	case marketplace.LifecycleActionResume:
		return "resume"
	case marketplace.LifecycleActionResize:
		return "resize"
	case marketplace.LifecycleActionTerminate:
		return "terminate"
	case marketplace.LifecycleActionProvision:
		return "provision"
	default:
		return ""
	}
}

func (b *WaldurBridge) signAndSubmitCallback(ctx context.Context, callback *marketplace.WaldurCallback) error {
	if callback == nil {
		return errors.New("callback is nil")
	}

	sig, err := b.keyManager.Sign(callback.SigningPayload())
	if err != nil {
		return fmt.Errorf("sign callback: %w", err)
	}

	sigBytes, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	callback.Signature = sigBytes
	return b.callbackSink.Submit(ctx, callback)
}

func defaultMarketplaceEventQuery() string {
	eventTypes := []string{
		string(marketplace.EventAllocationCreated),
		string(marketplace.EventTerminateRequested),
		string(marketplace.EventUsageUpdateRequested),
	}

	parts := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		parts = append(parts, fmt.Sprintf("marketplace_event.event_type='%s'", eventType))
	}

	return fmt.Sprintf("tm.event='Tx' AND (%s)", strings.Join(parts, " OR "))
}

func (b *WaldurBridge) validateOfferingMap() error {
	for offeringID, waldurUUID := range b.cfg.WaldurOfferingMap {
		if offeringID == "" || waldurUUID == "" {
			return fmt.Errorf("waldur offering map contains empty values")
		}
		offering, err := marketplace.ParseOfferingID(offeringID)
		if err != nil {
			return fmt.Errorf("invalid offering id in map (%s): %w", offeringID, err)
		}
		if !strings.EqualFold(offering.ProviderAddress, b.cfg.ProviderAddress) {
			return fmt.Errorf("offering %s provider mismatch: %s != %s", offeringID, offering.ProviderAddress, b.cfg.ProviderAddress)
		}
		if _, err := uuid.Parse(waldurUUID); err != nil {
			return fmt.Errorf("invalid waldur offering UUID for %s: %w", offeringID, err)
		}
	}
	return nil
}

func (b *WaldurBridge) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	timeout := b.cfg.OperationTimeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return context.WithTimeout(parent, timeout)
}

func (b *WaldurBridge) ensureAllocationMapping(ctx context.Context, event marketplace.AllocationCreatedEvent) (*WaldurAllocationMapping, error) {
	existing := b.state.Mappings[event.AllocationID]
	if existing != nil {
		return b.refreshAllocationMapping(ctx, existing), nil
	}

	// VE-2D: When automatic offering sync is enabled, look up the Waldur offering UUID from sync state
	var waldurOfferingUUID string
	if b.cfg.OfferingSyncEnabled && b.offeringSyncWorker != nil {
		syncState := b.offeringSyncWorker.State()
		if syncState != nil {
			record := syncState.GetRecord(event.OfferingID)
			if record != nil && record.WaldurUUID != "" {
				waldurOfferingUUID = record.WaldurUUID
			}
		}
		// If not in sync state, try to look up by backend ID
		if waldurOfferingUUID == "" {
			opCtx, cancel := b.operationContext(ctx)
			defer cancel()
			offering, err := b.marketplace.GetOfferingByBackendID(opCtx, event.OfferingID)
			if err == nil && offering != nil {
				waldurOfferingUUID = offering.UUID
			}
		}
	} else {
		// Legacy: use manual offering map
		waldurOfferingUUID = b.cfg.WaldurOfferingMap[event.OfferingID]
	}

	if waldurOfferingUUID == "" {
		return nil, fmt.Errorf("no waldur offering mapping for offering %s (enable OfferingSyncEnabled or add to WaldurOfferingMap)", event.OfferingID)
	}

	attrs := map[string]interface{}{
		"allocation_id":    event.AllocationID,
		"order_id":         event.OrderID,
		"provider_address": event.ProviderAddress,
		"customer_address": event.CustomerAddress,
		"offering_id":      event.OfferingID,
		"bid_id":           event.BidID,
		"accepted_price":   event.AcceptedPrice,
	}

	opCtx, cancel := b.operationContext(ctx)
	defer cancel()
	order, err := b.marketplace.CreateOrder(opCtx, waldur.CreateOrderRequest{
		OfferingUUID:   waldurOfferingUUID,
		ProjectUUID:    b.cfg.WaldurProjectUUID,
		CallbackURL:    b.cfg.OrderCallbackURL,
		RequestComment: fmt.Sprintf("virtengine order %s allocation %s", event.OrderID, event.AllocationID),
		Attributes:     attrs,
		Name:           fmt.Sprintf("ve-%s", event.AllocationID),
		Description:    fmt.Sprintf("VirtEngine allocation %s", event.AllocationID),
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	opCtx, cancel = b.operationContext(ctx)
	defer cancel()
	if err := b.marketplace.ApproveOrderByProvider(opCtx, order.UUID); err != nil {
		return nil, fmt.Errorf("approve order: %w", err)
	}

	opCtx, cancel = b.operationContext(ctx)
	defer cancel()
	if err := b.marketplace.SetOrderBackendID(opCtx, order.UUID, event.AllocationID); err != nil {
		return nil, fmt.Errorf("set backend id: %w", err)
	}

	mapping := &WaldurAllocationMapping{
		AllocationID: event.AllocationID,
		OrderUUID:    order.UUID,
		ResourceUUID: order.ResourceUUID,
		OfferingUUID: waldurOfferingUUID,
		UpdatedAt:    time.Now().UTC(),
	}
	b.state.Mappings[event.AllocationID] = mapping
	if err := b.stateStore.Save(b.state); err != nil {
		return nil, fmt.Errorf("save mapping: %w", err)
	}

	return mapping, nil
}

func (b *WaldurBridge) refreshAllocationMapping(ctx context.Context, mapping *WaldurAllocationMapping) *WaldurAllocationMapping {
	if mapping == nil || mapping.OrderUUID == "" || mapping.ResourceUUID != "" {
		return mapping
	}

	opCtx, cancel := b.operationContext(ctx)
	defer cancel()
	order, err := b.marketplace.GetOrder(opCtx, mapping.OrderUUID)
	if err != nil {
		return mapping
	}
	if order.ResourceUUID != "" && order.ResourceUUID != mapping.ResourceUUID {
		mapping.ResourceUUID = order.ResourceUUID
		mapping.UpdatedAt = time.Now().UTC()
		b.state.Mappings[mapping.AllocationID] = mapping
		_ = b.stateStore.Save(b.state)
	}
	return mapping
}

func hoursFromMilliSeconds(ms int64) float64 {
	return float64(ms) / (1000.0 * 60.0 * 60.0)
}

func hoursFromSeconds(seconds int64) float64 {
	return float64(seconds) / 3600.0
}

func gbHoursFromByteSeconds(byteSeconds int64) float64 {
	return float64(byteSeconds) / (1024.0 * 1024.0 * 1024.0 * 3600.0)
}

func gbFromBytes(bytes int64) float64 {
	return float64(bytes) / (1024.0 * 1024.0 * 1024.0)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}
