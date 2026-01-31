// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ============================================================================
// Webhook Server
// ============================================================================

// WebhookServer handles incoming webhook requests from off-ramp providers.
type WebhookServer struct {
	service  Service
	config   WebhookConfig
	mux      *http.ServeMux
	providers map[ProviderType]Provider

	// Idempotency tracking
	processedEvents map[string]time.Time
	processedMu     sync.RWMutex
}

// NewWebhookServer creates a new webhook server.
func NewWebhookServer(service Service, config WebhookConfig, providers map[ProviderType]Provider) *WebhookServer {
	ws := &WebhookServer{
		service:         service,
		config:          config,
		mux:             http.NewServeMux(),
		providers:       providers,
		processedEvents: make(map[string]time.Time),
	}

	// Register endpoints
	ws.mux.HandleFunc(config.Path, ws.handleWebhook)
	ws.mux.HandleFunc(config.Path+"/paypal", ws.handlePayPalWebhook)
	ws.mux.HandleFunc(config.Path+"/ach", ws.handleACHWebhook)
	ws.mux.HandleFunc(config.Path+"/stripe", ws.handleACHWebhook) // Alias for Stripe Treasury

	return ws
}

// ServeHTTP implements http.Handler.
func (ws *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.mux.ServeHTTP(w, r)
}

// handleWebhook handles generic webhook requests.
func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Try to determine provider from headers or body
	var provider ProviderType
	if r.Header.Get("PayPal-Transmission-Id") != "" {
		provider = ProviderPayPal
	} else if r.Header.Get("Stripe-Signature") != "" {
		provider = ProviderACH
	} else {
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}

	ws.processWebhook(w, r, body, provider)
}

// handlePayPalWebhook handles PayPal-specific webhooks.
func (ws *WebhookServer) handlePayPalWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	ws.processWebhook(w, r, body, ProviderPayPal)
}

// handleACHWebhook handles ACH/Stripe-specific webhooks.
func (ws *WebhookServer) handleACHWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	ws.processWebhook(w, r, body, ProviderACH)
}

// processWebhook processes a webhook for a specific provider.
func (ws *WebhookServer) processWebhook(w http.ResponseWriter, r *http.Request, body []byte, providerType ProviderType) {
	ctx := r.Context()

	// Get provider
	provider, ok := ws.providers[providerType]
	if !ok {
		http.Error(w, "Provider not configured", http.StatusBadRequest)
		return
	}

	// Get signature based on provider
	var signature string
	switch providerType {
	case ProviderPayPal:
		signature = r.Header.Get("PayPal-Transmission-Sig")
	case ProviderACH:
		signature = r.Header.Get("Stripe-Signature")
	}

	// Validate signature
	if ws.config.SignatureVerification {
		if err := provider.ValidateWebhook(body, signature); err != nil {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse event
	event, err := provider.ParseWebhookEvent(body)
	if err != nil {
		http.Error(w, "Invalid event payload", http.StatusBadRequest)
		return
	}

	// Check idempotency
	if ws.isProcessed(event.ID) {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process event through service
	if err := ws.service.HandleWebhook(ctx, providerType, body, signature); err != nil {
		// Log error but return 200 to prevent retries for non-retriable errors
		w.WriteHeader(http.StatusOK)
		return
	}

	// Mark as processed
	ws.markProcessed(event.ID)

	w.WriteHeader(http.StatusOK)
}

// isProcessed checks if an event was already processed.
func (ws *WebhookServer) isProcessed(eventID string) bool {
	ws.processedMu.RLock()
	defer ws.processedMu.RUnlock()
	_, exists := ws.processedEvents[eventID]
	return exists
}

// markProcessed marks an event as processed.
func (ws *WebhookServer) markProcessed(eventID string) {
	ws.processedMu.Lock()
	defer ws.processedMu.Unlock()
	ws.processedEvents[eventID] = time.Now()

	// Clean up old entries (keep last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	for id, t := range ws.processedEvents {
		if t.Before(cutoff) {
			delete(ws.processedEvents, id)
		}
	}
}

// ============================================================================
// Default Webhook Handler
// ============================================================================

// DefaultWebhookHandler implements WebhookHandler.
type DefaultWebhookHandler struct {
	mu       sync.RWMutex
	handlers map[WebhookEventType][]EventHandler
	store    PayoutStore
}

// NewDefaultWebhookHandler creates a new webhook handler.
func NewDefaultWebhookHandler(store PayoutStore) *DefaultWebhookHandler {
	return &DefaultWebhookHandler{
		handlers: make(map[WebhookEventType][]EventHandler),
		store:    store,
	}
}

// HandleEvent processes a webhook event.
func (h *DefaultWebhookHandler) HandleEvent(ctx context.Context, event *WebhookEvent) error {
	// Update payout status first
	if err := h.updatePayoutFromEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to update payout: %w", err)
	}

	// Execute registered handlers
	h.mu.RLock()
	handlers, ok := h.handlers[event.Type]
	h.mu.RUnlock()

	if !ok || len(handlers) == 0 {
		return nil
	}

	var lastErr error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// updatePayoutFromEvent updates payout status based on webhook event.
func (h *DefaultWebhookHandler) updatePayoutFromEvent(ctx context.Context, event *WebhookEvent) error {
	// Try to find payout by provider ID first, then by our ID
	var payout *PayoutIntent
	var err error

	if event.ProviderPayoutID != "" {
		payout, err = h.store.GetByProviderPayoutID(ctx, event.ProviderPayoutID)
	}
	if err != nil || payout == nil {
		if event.PayoutID != "" {
			payout, err = h.store.GetByID(ctx, event.PayoutID)
		}
	}

	if err != nil || payout == nil {
		return fmt.Errorf("payout not found for event: %v", err)
	}

	// Update status
	oldStatus := payout.Status
	payout.Status = event.Status
	payout.UpdatedAt = time.Now()

	if event.FailureCode != "" {
		payout.FailureCode = event.FailureCode
		payout.FailureMessage = event.FailureMessage
	}

	if event.Status.IsTerminal() {
		now := time.Now()
		payout.CompletedAt = &now
	}

	payout.AddAuditEntry(
		"webhook_received",
		"webhook_handler",
		fmt.Sprintf("event=%s, old_status=%s, new_status=%s", event.Type, oldStatus, event.Status),
	)

	return h.store.Save(ctx, payout)
}

// RegisterHandler registers a handler for a specific event type.
func (h *DefaultWebhookHandler) RegisterHandler(eventType WebhookEventType, handler EventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[eventType] = append(h.handlers[eventType], handler)
}

// UnregisterHandler removes handlers for an event type.
func (h *DefaultWebhookHandler) UnregisterHandler(eventType WebhookEventType) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.handlers, eventType)
}

// ============================================================================
// Webhook Event Builder (for testing)
// ============================================================================

// WebhookEventBuilder helps build webhook events for testing.
type WebhookEventBuilder struct {
	event WebhookEvent
}

// NewWebhookEventBuilder creates a new event builder.
func NewWebhookEventBuilder() *WebhookEventBuilder {
	return &WebhookEventBuilder{
		event: WebhookEvent{
			ID:         fmt.Sprintf("evt_%d", time.Now().UnixNano()),
			Timestamp:  time.Now(),
			ReceivedAt: time.Now(),
		},
	}
}

// WithType sets the event type.
func (b *WebhookEventBuilder) WithType(t WebhookEventType) *WebhookEventBuilder {
	b.event.Type = t
	return b
}

// WithProvider sets the provider.
func (b *WebhookEventBuilder) WithProvider(p ProviderType) *WebhookEventBuilder {
	b.event.Provider = p
	return b
}

// WithPayoutID sets the payout ID.
func (b *WebhookEventBuilder) WithPayoutID(id string) *WebhookEventBuilder {
	b.event.PayoutID = id
	return b
}

// WithProviderPayoutID sets the provider payout ID.
func (b *WebhookEventBuilder) WithProviderPayoutID(id string) *WebhookEventBuilder {
	b.event.ProviderPayoutID = id
	return b
}

// WithStatus sets the status.
func (b *WebhookEventBuilder) WithStatus(s PayoutStatus) *WebhookEventBuilder {
	b.event.Status = s
	return b
}

// WithFailure sets failure details.
func (b *WebhookEventBuilder) WithFailure(code, message string) *WebhookEventBuilder {
	b.event.FailureCode = code
	b.event.FailureMessage = message
	return b
}

// Build returns the built event.
func (b *WebhookEventBuilder) Build() *WebhookEvent {
	return &b.event
}

// ============================================================================
// Mock Provider Webhook Payloads
// ============================================================================

// MockPayPalWebhookPayload creates a mock PayPal webhook payload.
func MockPayPalWebhookPayload(eventType string, payoutItemID string, senderItemID string, status string) []byte {
	payload := map[string]interface{}{
		"id":          fmt.Sprintf("WH-%d", time.Now().UnixNano()),
		"event_type":  eventType,
		"create_time": time.Now().Format(time.RFC3339),
		"resource": map[string]interface{}{
			"payout_item_id":     payoutItemID,
			"sender_item_id":     senderItemID,
			"transaction_status": status,
		},
	}

	data, _ := json.Marshal(payload)
	return data
}

// MockStripeWebhookPayload creates a mock Stripe webhook payload.
func MockStripeWebhookPayload(eventType string, outboundPaymentID string, payoutID string, status string) []byte {
	payload := map[string]interface{}{
		"id":      fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		"type":    eventType,
		"created": time.Now().Unix(),
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":     outboundPaymentID,
				"status": status,
				"metadata": map[string]interface{}{
					"payout_id": payoutID,
				},
			},
		},
	}

	data, _ := json.Marshal(payload)
	return data
}

// Ensure implementations satisfy interfaces
var (
	_ WebhookHandler = (*DefaultWebhookHandler)(nil)
)
