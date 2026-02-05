package servicedesk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/log"

	"github.com/virtengine/virtengine/pkg/jira"
	"github.com/virtengine/virtengine/pkg/waldur"
)

// Bridge is the main service that orchestrates ticket sync between
// VirtEngine on-chain tickets and external service desk systems.
type Bridge struct {
	config *Config
	logger log.Logger

	// Adapters
	jiraClient    jira.IClient
	jiraBridge    jira.ITicketBridge
	waldurClient  *waldur.Client
	waldurSupport *waldur.SupportClient

	// Internal components
	syncManager    *SyncManager
	auditLogger    *AuditLogger
	callbackServer *CallbackServer
	retryQueue     *RetryQueue
	decryptor      PayloadDecryptor

	// State
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup

	// Event channel for on-chain events
	eventCh chan *SyncEvent

	// Optional handler for external reference updates
	externalRefHandler ExternalRefHandler
}

// IBridge defines the bridge service interface
type IBridge interface {
	// Start starts the bridge service
	Start(ctx context.Context) error

	// Stop stops the bridge service
	Stop(ctx context.Context) error

	// HandleTicketCreated handles a ticket created event
	HandleTicketCreated(ctx context.Context, event *TicketCreatedEvent) error

	// HandleTicketUpdated handles a ticket updated event
	HandleTicketUpdated(ctx context.Context, event *TicketUpdatedEvent) error

	// HandleTicketClosed handles a ticket closed event
	HandleTicketClosed(ctx context.Context, event *TicketClosedEvent) error

	// HandleTicketResponseAdded handles a ticket response event
	HandleTicketResponseAdded(ctx context.Context, event *TicketResponseAddedEvent) error

	// HandleExternalCallback handles a callback from external service desk
	HandleExternalCallback(ctx context.Context, payload *CallbackPayload) error

	// GetSyncStatus returns the sync status for a ticket
	GetSyncStatus(ctx context.Context, ticketID string) (*TicketSyncRecord, error)

	// SyncTicket manually triggers sync for a ticket
	SyncTicket(ctx context.Context, ticketID string, direction SyncDirection) error

	// Health returns the health status
	Health(ctx context.Context) (*HealthStatus, error)

	// Decryptor returns the payload decryptor
	Decryptor() PayloadDecryptor

	// SetInboundHandler sets a handler for inbound updates
	SetInboundHandler(handler InboundUpdateHandler)

	// GetTicketRecipients returns recipient key IDs for a ticket
	GetTicketRecipients(ticketID string) []string

	// SetExternalRefHandler sets the handler for external reference updates
	SetExternalRefHandler(handler ExternalRefHandler)
}

// ExternalRefHandler handles external ticket references created by the bridge.
type ExternalRefHandler interface {
	HandleExternalRef(ctx context.Context, ticketID string, ref ExternalTicketRef) error
}

// Ensure Bridge implements IBridge
var _ IBridge = (*Bridge)(nil)

// TicketCreatedEvent represents an on-chain ticket creation event
type TicketCreatedEvent struct {
	TicketID        string         `json:"ticket_id"`
	TicketNumber    string         `json:"ticket_number,omitempty"`
	CustomerAddress string         `json:"customer_address"`
	ProviderAddress string         `json:"provider_address,omitempty"`
	Category        string         `json:"category"`
	Priority        string         `json:"priority"`
	Subject         string         `json:"subject"`
	Description     string         `json:"description"`
	Recipients      []string       `json:"recipients,omitempty"`
	RelatedEntity   *RelatedEntity `json:"related_entity,omitempty"`
	BlockHeight     int64          `json:"block_height"`
	Timestamp       time.Time      `json:"timestamp"`
	TxHash          string         `json:"tx_hash"`
}

// RelatedEntity represents a related on-chain entity
type RelatedEntity struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// TicketUpdatedEvent represents an on-chain ticket update event
type TicketUpdatedEvent struct {
	TicketID    string                 `json:"ticket_id"`
	Changes     map[string]interface{} `json:"changes"`
	UpdatedBy   string                 `json:"updated_by"`
	BlockHeight int64                  `json:"block_height"`
	Timestamp   time.Time              `json:"timestamp"`
	TxHash      string                 `json:"tx_hash"`
}

// TicketClosedEvent represents an on-chain ticket closed event
type TicketClosedEvent struct {
	TicketID    string    `json:"ticket_id"`
	ClosedBy    string    `json:"closed_by"`
	Resolution  string    `json:"resolution,omitempty"`
	BlockHeight int64     `json:"block_height"`
	Timestamp   time.Time `json:"timestamp"`
	TxHash      string    `json:"tx_hash"`
}

// TicketResponseAddedEvent represents a response added to a ticket.
type TicketResponseAddedEvent struct {
	TicketID    string    `json:"ticket_id"`
	Author      string    `json:"author"`
	IsAgent     bool      `json:"is_agent"`
	Message     string    `json:"message"`
	BlockHeight int64     `json:"block_height"`
	Timestamp   time.Time `json:"timestamp"`
	TxHash      string    `json:"tx_hash"`
}

// HealthStatus represents the bridge health status
type HealthStatus struct {
	Healthy       bool      `json:"healthy"`
	JiraStatus    string    `json:"jira_status,omitempty"`
	WaldurStatus  string    `json:"waldur_status,omitempty"`
	LastSync      time.Time `json:"last_sync"`
	PendingEvents int       `json:"pending_events"`
	FailedEvents  int       `json:"failed_events"`
}

// NewBridge creates a new bridge service
func NewBridge(config *Config, logger log.Logger) (*Bridge, error) {
	if err := config.Validate(); err != nil {
		return nil, ErrConfigInvalid.Wrap(err.Error())
	}

	bridge := &Bridge{
		config:  config,
		logger:  logger.With("module", "servicedesk"),
		stopCh:  make(chan struct{}),
		eventCh: make(chan *SyncEvent, 1000),
	}

	decryptor, err := NewPayloadDecryptor(config.Decryption)
	if err != nil {
		return nil, fmt.Errorf("failed to init decryptor: %w", err)
	}
	bridge.decryptor = decryptor

	// Initialize Jira client if configured
	if config.JiraConfig != nil {
		jiraClient, err := jira.NewClient(jira.ClientConfig{
			BaseURL: config.JiraConfig.BaseURL,
			Auth: jira.AuthConfig{
				Type:     jira.AuthTypeBasic,
				Username: config.JiraConfig.Username,
				APIToken: config.JiraConfig.APIToken,
			},
			Timeout: config.JiraConfig.Timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create jira client: %w", err)
		}
		bridge.jiraClient = jiraClient
		bridge.jiraBridge = jira.NewTicketBridge(jiraClient, jira.TicketBridgeConfig{
			ProjectKey:       config.JiraConfig.ProjectKey,
			DefaultIssueType: jira.IssueType(config.JiraConfig.IssueType),
			PriorityMapping:  mapPriorityToJira(config.MappingSchema),
		})
	}

	// Initialize Waldur client if configured
	if config.WaldurConfig != nil {
		waldurClient, err := waldur.NewClient(waldur.Config{
			BaseURL: config.WaldurConfig.BaseURL,
			Token:   config.WaldurConfig.Token,
			Timeout: config.WaldurConfig.Timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create waldur client: %w", err)
		}
		bridge.waldurClient = waldurClient
		bridge.waldurSupport = waldur.NewSupportClient(waldurClient)
	}

	// Initialize internal components
	bridge.syncManager = NewSyncManager(bridge, config.SyncConfig)
	bridge.auditLogger = NewAuditLogger(config.AuditConfig, logger)
	bridge.retryQueue = NewRetryQueue(config.RetryConfig, logger)

	return bridge, nil
}

// mapPriorityToJira creates the priority mapping for Jira
func mapPriorityToJira(schema *MappingSchema) map[string]jira.Priority {
	mapping := make(map[string]jira.Priority)
	if schema == nil {
		return mapping
	}
	for _, m := range schema.PriorityMappings {
		mapping[m.OnChainPriority] = jira.Priority(m.JiraPriority)
	}
	return mapping
}

// Start starts the bridge service
func (b *Bridge) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = true
	b.mu.Unlock()

	b.logger.Info("starting service desk bridge")

	// Start internal components
	b.wg.Add(1)
	go b.eventProcessor(ctx)

	b.wg.Add(1)
	go b.syncLoop(ctx)

	// Start callback server if enabled
	if b.config.WebhookConfig.Enabled {
		b.callbackServer = NewCallbackServer(b, b.config, b.logger)
		if err := b.callbackServer.Start(ctx); err != nil {
			return fmt.Errorf("failed to start callback server: %w", err)
		}
	}

	// Start retry queue processor
	b.wg.Add(1)
	go b.retryQueue.Start(ctx, b.processRetryEvent)

	b.logger.Info("service desk bridge started")
	return nil
}

// Stop stops the bridge service
func (b *Bridge) Stop(ctx context.Context) error {
	b.mu.Lock()
	if !b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = false
	b.mu.Unlock()

	b.logger.Info("stopping service desk bridge")

	// Signal stop
	close(b.stopCh)

	// Stop callback server
	if b.callbackServer != nil {
		if err := b.callbackServer.Stop(ctx); err != nil {
			b.logger.Error("failed to stop callback server", "error", err)
		}
	}

	// Wait for goroutines
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	b.logger.Info("service desk bridge stopped")
	return nil
}

// HandleTicketCreated handles a ticket created event
func (b *Bridge) HandleTicketCreated(ctx context.Context, event *TicketCreatedEvent) error {
	b.logger.Debug("handling ticket created event", "ticket_id", event.TicketID)

	// Create audit entry
	b.auditLogger.LogEvent(ctx, AuditEventTicketCreate, map[string]interface{}{
		"ticket_id":    event.TicketID,
		"customer":     event.CustomerAddress,
		"category":     event.Category,
		"priority":     event.Priority,
		"block_height": event.BlockHeight,
	})

	// Create sync event
	syncEvent := &SyncEvent{
		ID:        fmt.Sprintf("create-%s-%d", event.TicketID, event.BlockHeight),
		Type:      "ticket_created",
		TicketID:  event.TicketID,
		Direction: SyncDirectionOutbound,
		Payload: map[string]interface{}{
			"customer_address": event.CustomerAddress,
			"provider_address": event.ProviderAddress,
			"category":         event.Category,
			"priority":         event.Priority,
			"subject":          event.Subject,
			"description":      event.Description,
			"ticket_number":    event.TicketNumber,
			"related_entity":   event.RelatedEntity,
			"tx_hash":          event.TxHash,
		},
		Timestamp:   event.Timestamp,
		BlockHeight: event.BlockHeight,
		MaxRetries:  b.config.RetryConfig.MaxRetries,
		Status:      SyncStatusPending,
	}

	if len(event.Recipients) > 0 {
		b.syncManager.SetTicketRecipients(event.TicketID, event.Recipients)
	}

	// Queue for processing
	select {
	case b.eventCh <- syncEvent:
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue full, add to retry queue
		b.retryQueue.Add(syncEvent)
	}

	return nil
}

// HandleTicketUpdated handles a ticket updated event
func (b *Bridge) HandleTicketUpdated(ctx context.Context, event *TicketUpdatedEvent) error {
	b.logger.Debug("handling ticket updated event", "ticket_id", event.TicketID)

	// Create audit entry
	b.auditLogger.LogEvent(ctx, AuditEventTicketUpdate, map[string]interface{}{
		"ticket_id":    event.TicketID,
		"updated_by":   event.UpdatedBy,
		"changes":      event.Changes,
		"block_height": event.BlockHeight,
	})

	// Create sync event
	syncEvent := &SyncEvent{
		ID:          fmt.Sprintf("update-%s-%d", event.TicketID, event.BlockHeight),
		Type:        "ticket_updated",
		TicketID:    event.TicketID,
		Direction:   SyncDirectionOutbound,
		Payload:     event.Changes,
		Timestamp:   event.Timestamp,
		BlockHeight: event.BlockHeight,
		MaxRetries:  b.config.RetryConfig.MaxRetries,
		Status:      SyncStatusPending,
	}

	// Queue for processing
	select {
	case b.eventCh <- syncEvent:
	case <-ctx.Done():
		return ctx.Err()
	default:
		b.retryQueue.Add(syncEvent)
	}

	return nil
}

// HandleTicketClosed handles a ticket closed event
func (b *Bridge) HandleTicketClosed(ctx context.Context, event *TicketClosedEvent) error {
	b.logger.Debug("handling ticket closed event", "ticket_id", event.TicketID)

	// Create audit entry
	b.auditLogger.LogEvent(ctx, AuditEventTicketClose, map[string]interface{}{
		"ticket_id":    event.TicketID,
		"closed_by":    event.ClosedBy,
		"resolution":   event.Resolution,
		"block_height": event.BlockHeight,
	})

	// Create sync event
	syncEvent := &SyncEvent{
		ID:        fmt.Sprintf("close-%s-%d", event.TicketID, event.BlockHeight),
		Type:      "ticket_closed",
		TicketID:  event.TicketID,
		Direction: SyncDirectionOutbound,
		Payload: map[string]interface{}{
			"closed_by":  event.ClosedBy,
			"resolution": event.Resolution,
			"tx_hash":    event.TxHash,
		},
		Timestamp:   event.Timestamp,
		BlockHeight: event.BlockHeight,
		MaxRetries:  b.config.RetryConfig.MaxRetries,
		Status:      SyncStatusPending,
	}

	// Queue for processing
	select {
	case b.eventCh <- syncEvent:
	case <-ctx.Done():
		return ctx.Err()
	default:
		b.retryQueue.Add(syncEvent)
	}

	return nil
}

// HandleTicketResponseAdded handles a ticket response event.
func (b *Bridge) HandleTicketResponseAdded(ctx context.Context, event *TicketResponseAddedEvent) error {
	b.logger.Debug("handling ticket response event", "ticket_id", event.TicketID)

	b.auditLogger.LogEvent(ctx, AuditEventTicketUpdate, map[string]interface{}{
		"ticket_id":    event.TicketID,
		"author":       event.Author,
		"is_agent":     event.IsAgent,
		"block_height": event.BlockHeight,
	})

	syncEvent := &SyncEvent{
		ID:        fmt.Sprintf("response-%s-%d", event.TicketID, event.BlockHeight),
		Type:      "ticket_response_added",
		TicketID:  event.TicketID,
		Direction: SyncDirectionOutbound,
		Payload: map[string]interface{}{
			"author":   event.Author,
			"is_agent": event.IsAgent,
			"message":  event.Message,
			"tx_hash":  event.TxHash,
		},
		Timestamp:   event.Timestamp,
		BlockHeight: event.BlockHeight,
		MaxRetries:  b.config.RetryConfig.MaxRetries,
		Status:      SyncStatusPending,
	}

	select {
	case b.eventCh <- syncEvent:
	case <-ctx.Done():
		return ctx.Err()
	default:
		b.retryQueue.Add(syncEvent)
	}

	return nil
}

// HandleExternalCallback handles a callback from external service desk
func (b *Bridge) HandleExternalCallback(ctx context.Context, payload *CallbackPayload) error {
	if err := payload.Validate(); err != nil {
		return ErrSignatureInvalid.Wrap(err.Error())
	}

	b.logger.Debug("handling external callback",
		"event_type", payload.EventType,
		"external_id", payload.ExternalID,
		"ticket_id", payload.OnChainTicketID,
	)

	// Create audit entry
	b.auditLogger.LogEvent(ctx, AuditEventExternalCallback, map[string]interface{}{
		"event_type":   payload.EventType,
		"service_desk": payload.ServiceDeskType,
		"external_id":  payload.ExternalID,
		"ticket_id":    payload.OnChainTicketID,
		"changes":      payload.Changes,
	})

	// Create sync event for inbound sync
	syncEvent := &SyncEvent{
		ID:         fmt.Sprintf("callback-%s-%s", payload.ExternalID, payload.Nonce),
		Type:       payload.EventType,
		TicketID:   payload.OnChainTicketID,
		Direction:  SyncDirectionInbound,
		Payload:    payload.Changes,
		Timestamp:  payload.Timestamp,
		MaxRetries: b.config.RetryConfig.MaxRetries,
		Status:     SyncStatusPending,
	}

	// Queue for processing
	select {
	case b.eventCh <- syncEvent:
	case <-ctx.Done():
		return ctx.Err()
	default:
		b.retryQueue.Add(syncEvent)
	}

	return nil
}

// GetSyncStatus returns the sync status for a ticket
func (b *Bridge) GetSyncStatus(ctx context.Context, ticketID string) (*TicketSyncRecord, error) {
	return b.syncManager.GetSyncRecord(ctx, ticketID)
}

// SyncTicket manually triggers sync for a ticket
func (b *Bridge) SyncTicket(ctx context.Context, ticketID string, direction SyncDirection) error {
	b.logger.Info("manual sync triggered", "ticket_id", ticketID, "direction", direction)

	b.auditLogger.LogEvent(ctx, AuditEventManualSync, map[string]interface{}{
		"ticket_id": ticketID,
		"direction": direction,
	})

	return b.syncManager.SyncTicket(ctx, ticketID, direction)
}

// Health returns the health status
func (b *Bridge) Health(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:       true,
		LastSync:      b.syncManager.LastSyncTime(),
		PendingEvents: len(b.eventCh),
		FailedEvents:  b.retryQueue.FailedCount(),
	}

	// Check Jira health
	if b.jiraClient != nil {
		_, err := b.jiraClient.GetServiceDeskInfo(ctx)
		if err != nil {
			status.JiraStatus = "unhealthy: " + err.Error()
			status.Healthy = false
		} else {
			status.JiraStatus = "healthy"
		}
	}

	// Check Waldur health
	if b.waldurClient != nil {
		err := b.waldurClient.HealthCheck(ctx)
		if err != nil {
			status.WaldurStatus = "unhealthy: " + err.Error()
			status.Healthy = false
		} else {
			status.WaldurStatus = "healthy"
		}
	}

	return status, nil
}

// Decryptor returns the configured payload decryptor.
func (b *Bridge) Decryptor() PayloadDecryptor {
	return b.decryptor
}

// SetInboundHandler sets the handler used for inbound sync updates.
func (b *Bridge) SetInboundHandler(handler InboundUpdateHandler) {
	if b.syncManager != nil {
		b.syncManager.SetInboundHandler(handler)
	}
}

// GetTicketRecipients returns cached recipients for a ticket.
func (b *Bridge) GetTicketRecipients(ticketID string) []string {
	if b.syncManager == nil {
		return nil
	}
	return b.syncManager.GetTicketRecipients(ticketID)
}

// SetExternalRefHandler registers a handler for external ticket references.
func (b *Bridge) SetExternalRefHandler(handler ExternalRefHandler) {
	b.externalRefHandler = handler
}

// eventProcessor processes sync events from the queue
func (b *Bridge) eventProcessor(ctx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ctx.Done():
			return
		case event := <-b.eventCh:
			if err := b.processEvent(ctx, event); err != nil {
				b.logger.Error("failed to process event",
					"event_id", event.ID,
					"error", err,
				)
				event.Error = err.Error()
				event.RetryCount++
				if event.CanRetry() {
					b.retryQueue.Add(event)
				} else {
					event.Status = SyncStatusFailed
					b.auditLogger.LogEvent(ctx, AuditEventSyncFailed, map[string]interface{}{
						"event_id":    event.ID,
						"ticket_id":   event.TicketID,
						"error":       err.Error(),
						"retry_count": event.RetryCount,
					})
				}
			}
		}
	}
}

// processEvent processes a single sync event
func (b *Bridge) processEvent(ctx context.Context, event *SyncEvent) error {
	switch event.Direction {
	case SyncDirectionOutbound:
		return b.processOutboundEvent(ctx, event)
	case SyncDirectionInbound:
		return b.processInboundEvent(ctx, event)
	default:
		return fmt.Errorf("unknown sync direction: %s", event.Direction)
	}
}

// processOutboundEvent processes an outbound sync event (on-chain to external)
func (b *Bridge) processOutboundEvent(ctx context.Context, event *SyncEvent) error {
	if !b.config.SyncConfig.EnableOutbound {
		return nil
	}

	switch event.Type {
	case "ticket_created":
		return b.syncTicketCreated(ctx, event)
	case "ticket_updated":
		return b.syncTicketUpdated(ctx, event)
	case "ticket_closed":
		return b.syncTicketClosed(ctx, event)
	case "ticket_response_added":
		return b.syncTicketResponseAdded(ctx, event)
	default:
		b.logger.Warn("unknown outbound event type", "type", event.Type)
		return nil
	}
}

// processInboundEvent processes an inbound sync event (external to on-chain)
func (b *Bridge) processInboundEvent(ctx context.Context, event *SyncEvent) error {
	if !b.config.SyncConfig.EnableInbound {
		return nil
	}

	// Check for conflicts
	conflict, err := b.syncManager.CheckConflict(ctx, event)
	if err != nil {
		return err
	}
	if conflict != nil {
		return b.handleConflict(ctx, event, conflict)
	}

	// Process inbound update
	return b.syncManager.ProcessInboundUpdate(ctx, event)
}

// syncTicketCreated syncs a new ticket to external systems
func (b *Bridge) syncTicketCreated(ctx context.Context, event *SyncEvent) error {
	// Create in Jira
	if b.jiraBridge != nil {
		ticketNumber, _ := event.Payload["ticket_number"].(string)
		if ticketNumber == "" {
			ticketNumber = event.TicketID
		}
		req := &jira.VirtEngineSupportRequest{
			TicketID:         event.TicketID,
			TicketNumber:     ticketNumber,
			SubmitterAddress: event.Payload["customer_address"].(string),
			Category:         event.Payload["category"].(string),
			Priority:         event.Payload["priority"].(string),
			Subject:          event.Payload["subject"].(string),
			Description:      event.Payload["description"].(string),
			CreatedAt:        event.Timestamp,
		}

		if entity, ok := event.Payload["related_entity"].(*RelatedEntity); ok && entity != nil {
			req.RelatedEntity = &jira.RelatedEntity{
				Type: entity.Type,
				ID:   entity.ID,
			}
		}

		issue, err := b.jiraBridge.CreateFromSupportRequest(ctx, req)
		if err != nil {
			return ErrExternalAPIError.Wrapf("jira create failed: %v", err)
		}

		ref := ExternalTicketRef{
			Type:        ServiceDeskJira,
			ExternalID:  issue.Key,
			ExternalURL: issue.Self,
			SyncStatus:  SyncStatusSynced,
			CreatedAt:   time.Now(),
		}
		// Update sync record
		b.syncManager.UpdateExternalRef(ctx, event.TicketID, ref)
		if b.externalRefHandler != nil {
			if err := b.externalRefHandler.HandleExternalRef(ctx, event.TicketID, ref); err != nil {
				b.logger.Error("external ref handler failed", "error", err, "ticket_id", event.TicketID, "service", "jira")
			}
		}

		b.auditLogger.LogEvent(ctx, AuditEventSyncSuccess, map[string]interface{}{
			"ticket_id":   event.TicketID,
			"external_id": issue.Key,
			"service":     "jira",
		})
	}

	// Create in Waldur support system
	if b.waldurSupport != nil {
		category := ""
		priority := ""
		if c, ok := event.Payload["category"].(string); ok {
			category = c
		}
		if p, ok := event.Payload["priority"].(string); ok {
			priority = p
		}

		req := waldur.CreateIssueRequest{
			Type:        waldur.MapVirtEngineCategoryToWaldurType(category),
			Priority:    waldur.MapVirtEnginePriorityToWaldur(priority),
			Summary:     event.Payload["subject"].(string),
			Description: event.Payload["description"].(string),
			BackendID:   event.TicketID, // Store VirtEngine ticket ID
		}

		// Set customer/project if available in config
		if b.config.WaldurConfig.OrganizationUUID != "" {
			req.CustomerUUID = b.config.WaldurConfig.OrganizationUUID
		}
		if b.config.WaldurConfig.ProjectUUID != "" {
			req.ProjectUUID = b.config.WaldurConfig.ProjectUUID
		}

		issue, err := b.waldurSupport.CreateIssue(ctx, req)
		if err != nil {
			b.logger.Error("waldur issue creation failed", "error", err, "ticket_id", event.TicketID)
			// Don't fail the whole sync if Waldur fails - log and continue
			b.auditLogger.LogEvent(ctx, AuditEventSyncFailed, map[string]interface{}{
				"ticket_id": event.TicketID,
				"service":   "waldur",
				"error":     err.Error(),
			})
		} else {
			// Build external URL for the Waldur issue
			externalURL := fmt.Sprintf("%s/support/%s/", b.config.WaldurConfig.BaseURL, issue.UUID)

			ref := ExternalTicketRef{
				Type:        ServiceDeskWaldur,
				ExternalID:  issue.UUID,
				ExternalURL: externalURL,
				SyncStatus:  SyncStatusSynced,
				CreatedAt:   time.Now(),
			}
			// Update sync record
			b.syncManager.UpdateExternalRef(ctx, event.TicketID, ref)
			if b.externalRefHandler != nil {
				if err := b.externalRefHandler.HandleExternalRef(ctx, event.TicketID, ref); err != nil {
					b.logger.Error("external ref handler failed", "error", err, "ticket_id", event.TicketID, "service", "waldur")
				}
			}

			b.auditLogger.LogEvent(ctx, AuditEventSyncSuccess, map[string]interface{}{
				"ticket_id":   event.TicketID,
				"external_id": issue.UUID,
				"issue_key":   issue.Key,
				"service":     "waldur",
			})
		}
	}

	return nil
}

// syncTicketUpdated syncs ticket updates to external systems
func (b *Bridge) syncTicketUpdated(ctx context.Context, event *SyncEvent) error {
	record, err := b.syncManager.GetSyncRecord(ctx, event.TicketID)
	if err != nil {
		return err
	}
	if record == nil {
		// No sync record, ticket not yet synced
		return nil
	}

	// Update in Jira
	if jiraRef := record.GetExternalRef(ServiceDeskJira); jiraRef != nil && b.jiraBridge != nil {
		// Sync status changes
		if status, ok := event.Payload["status"].(string); ok {
			if err := b.jiraBridge.SyncStatus(ctx, jiraRef.ExternalID, status); err != nil {
				return ErrExternalAPIError.Wrapf("jira status sync failed: %v", err)
			}
		}

		b.auditLogger.LogEvent(ctx, AuditEventSyncSuccess, map[string]interface{}{
			"ticket_id":   event.TicketID,
			"external_id": jiraRef.ExternalID,
			"service":     "jira",
			"changes":     event.Payload,
		})
	}

	// Update in Waldur
	if waldurRef := record.GetExternalRef(ServiceDeskWaldur); waldurRef != nil && b.waldurSupport != nil {
		// Sync status changes
		if status, ok := event.Payload["status"].(string); ok {
			waldurState := waldur.MapVirtEngineStatusToWaldur(status)
			if err := b.waldurSupport.SetIssueState(ctx, waldurRef.ExternalID, waldurState, ""); err != nil {
				b.logger.Error("waldur status sync failed", "error", err, "ticket_id", event.TicketID)
				// Don't fail the whole sync
			}
		}

		// Sync priority changes
		if priority, ok := event.Payload["priority"].(string); ok {
			req := waldur.UpdateIssueRequest{
				Priority: waldur.MapVirtEnginePriorityToWaldur(priority),
			}
			if _, err := b.waldurSupport.UpdateIssue(ctx, waldurRef.ExternalID, req); err != nil {
				b.logger.Error("waldur priority sync failed", "error", err, "ticket_id", event.TicketID)
			}
		}

		b.auditLogger.LogEvent(ctx, AuditEventSyncSuccess, map[string]interface{}{
			"ticket_id":   event.TicketID,
			"external_id": waldurRef.ExternalID,
			"service":     "waldur",
			"changes":     event.Payload,
		})
	}

	return nil
}

// syncTicketClosed syncs ticket closure to external systems
func (b *Bridge) syncTicketClosed(ctx context.Context, event *SyncEvent) error {
	record, err := b.syncManager.GetSyncRecord(ctx, event.TicketID)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}

	// Close in Jira
	if jiraRef := record.GetExternalRef(ServiceDeskJira); jiraRef != nil && b.jiraBridge != nil {
		resolution := ""
		if r, ok := event.Payload["resolution"].(string); ok {
			resolution = r
		}
		if err := b.jiraBridge.CloseTicket(ctx, jiraRef.ExternalID, resolution); err != nil {
			return ErrExternalAPIError.Wrapf("jira close failed: %v", err)
		}

		// Update sync record
		now := time.Now()
		jiraRef.SyncStatus = SyncStatusSynced
		jiraRef.LastSyncAt = &now
		b.syncManager.UpdateExternalRef(ctx, event.TicketID, *jiraRef)

		b.auditLogger.LogEvent(ctx, AuditEventSyncSuccess, map[string]interface{}{
			"ticket_id":   event.TicketID,
			"external_id": jiraRef.ExternalID,
			"service":     "jira",
			"action":      "close",
		})
	}

	// Close in Waldur
	if waldurRef := record.GetExternalRef(ServiceDeskWaldur); waldurRef != nil && b.waldurSupport != nil {
		resolution := ""
		if r, ok := event.Payload["resolution"].(string); ok {
			resolution = r
		}
		if err := b.waldurSupport.SetIssueState(ctx, waldurRef.ExternalID, waldur.StateClosed, resolution); err != nil {
			b.logger.Error("waldur close failed", "error", err, "ticket_id", event.TicketID)
			// Don't fail the whole sync
		} else {
			// Update sync record
			now := time.Now()
			waldurRef.SyncStatus = SyncStatusSynced
			waldurRef.LastSyncAt = &now
			b.syncManager.UpdateExternalRef(ctx, event.TicketID, *waldurRef)

			b.auditLogger.LogEvent(ctx, AuditEventSyncSuccess, map[string]interface{}{
				"ticket_id":   event.TicketID,
				"external_id": waldurRef.ExternalID,
				"service":     "waldur",
				"action":      "close",
			})
		}
	}

	return nil
}

// syncTicketResponseAdded syncs a response as a comment in external systems
func (b *Bridge) syncTicketResponseAdded(ctx context.Context, event *SyncEvent) error {
	record, err := b.syncManager.GetSyncRecord(ctx, event.TicketID)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}

	message, _ := event.Payload["message"].(string)
	isAgent, _ := event.Payload["is_agent"].(bool)

	if jiraRef := record.GetExternalRef(ServiceDeskJira); jiraRef != nil && b.jiraBridge != nil {
		if _, err := b.jiraBridge.AddReplyToTicket(ctx, jiraRef.ExternalID, message, isAgent); err != nil {
			return ErrExternalAPIError.Wrapf("jira comment failed: %v", err)
		}
	}

	if waldurRef := record.GetExternalRef(ServiceDeskWaldur); waldurRef != nil && b.waldurSupport != nil {
		_, err := b.waldurSupport.AddComment(ctx, waldurRef.ExternalID, waldur.AddCommentRequest{
			Description: message,
			IsPublic:    true,
		})
		if err != nil {
			b.logger.Error("waldur comment failed", "error", err, "ticket_id", event.TicketID)
		}
	}

	return nil
}

// handleConflict handles a sync conflict
func (b *Bridge) handleConflict(ctx context.Context, event *SyncEvent, conflict *Conflict) error {
	b.auditLogger.LogEvent(ctx, AuditEventConflictDetected, map[string]interface{}{
		"ticket_id":  event.TicketID,
		"event_id":   event.ID,
		"conflict":   conflict,
		"resolution": b.config.SyncConfig.ConflictResolution,
	})

	switch b.config.SyncConfig.ConflictResolution {
	case ConflictResolutionOnChainWins:
		// Ignore inbound update, on-chain data wins
		b.logger.Info("conflict resolved: on-chain wins", "ticket_id", event.TicketID)
		return nil

	case ConflictResolutionExternalWins:
		// Process the inbound update
		return b.syncManager.ProcessInboundUpdate(ctx, event)

	case ConflictResolutionNewestWins:
		if event.Timestamp.After(conflict.OnChainTimestamp) {
			return b.syncManager.ProcessInboundUpdate(ctx, event)
		}
		return nil

	case ConflictResolutionManual:
		event.Status = SyncStatusConflict
		return ErrConflict.Wrapf("manual resolution required for ticket %s", event.TicketID)

	default:
		return fmt.Errorf("unknown conflict resolution strategy: %s", b.config.SyncConfig.ConflictResolution)
	}
}

// syncLoop runs the periodic sync loop
func (b *Bridge) syncLoop(ctx context.Context) {
	defer b.wg.Done()

	ticker := time.NewTicker(b.config.SyncConfig.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := b.syncManager.RunSync(ctx); err != nil {
				b.logger.Error("sync cycle failed", "error", err)
			}
		}
	}
}

// processRetryEvent processes a retry event
func (b *Bridge) processRetryEvent(ctx context.Context, event *SyncEvent) error {
	return b.processEvent(ctx, event)
}
