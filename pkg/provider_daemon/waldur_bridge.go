package provider_daemon

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
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
	WaldurOfferingMap  map[string]string
	OrderCallbackURL   string
}

// DefaultWaldurBridgeConfig returns a default configuration.
func DefaultWaldurBridgeConfig() WaldurBridgeConfig {
	return WaldurBridgeConfig{
		EventBuffer:        100,
		CallbackTTL:        time.Hour,
		OperationTimeout:   45 * time.Second,
		HealthCheckOnStart: true,
		HealthCheckTimeout: 15 * time.Second,
		CallbackSinkDir:    "data/callbacks",
		StateFile:          "data/waldur_bridge_state.json",
		CheckpointFile:     "data/marketplace_checkpoint.json",
		CometWS:            "/websocket",
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
		checkpointStore: NewEventCheckpointStore(cfg.CheckpointFile),
	}, nil
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
	if err := b.validateOfferingMap(); err != nil {
		return err
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

	waldurOfferingUUID := b.cfg.WaldurOfferingMap[event.OfferingID]
	if waldurOfferingUUID == "" {
		return nil, fmt.Errorf("no waldur offering mapping for offering %s", event.OfferingID)
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
