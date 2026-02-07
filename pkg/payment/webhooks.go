// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
package payment

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Webhook Handler Implementation
// ============================================================================

// WebhookServer handles incoming webhook requests from payment gateways
type WebhookServer struct {
	service Service
	config  WebhookConfig
	mux     *http.ServeMux

	// Idempotency tracking
	processedEvents map[string]time.Time
	processedMu     sync.RWMutex
}

// NewWebhookServer creates a new webhook server
func NewWebhookServer(service Service, config WebhookConfig) *WebhookServer {
	ws := &WebhookServer{
		service:         service,
		config:          config,
		mux:             http.NewServeMux(),
		processedEvents: make(map[string]time.Time),
	}

	ws.mux.HandleFunc(config.Path, ws.handleWebhook)
	ws.mux.HandleFunc(config.Path+"/stripe", ws.handleStripeWebhook)
	ws.mux.HandleFunc(config.Path+"/adyen", ws.handleAdyenWebhook)

	return ws
}

// ServeHTTP implements http.Handler
func (ws *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.mux.ServeHTTP(w, r)
}

// handleWebhook handles generic webhook requests
func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(body); err != nil && r.ContentLength > 0 {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Determine gateway from headers or content
	var gateway GatewayType
	if strings.Contains(r.Header.Get("Stripe-Signature"), "v1=") {
		gateway = GatewayStripe
	} else if r.Header.Get("X-Adyen-Hmac-Signature") != "" {
		gateway = GatewayAdyen
	} else {
		http.Error(w, "Unknown gateway", http.StatusBadRequest)
		return
	}

	ws.processWebhook(w, r, body, gateway)
}

// handleStripeWebhook handles Stripe-specific webhooks
func (ws *WebhookServer) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(body); err != nil && r.ContentLength > 0 {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	ws.processWebhook(w, r, body, GatewayStripe)
}

// handleAdyenWebhook handles Adyen-specific webhooks
func (ws *WebhookServer) handleAdyenWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(body); err != nil && r.ContentLength > 0 {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	ws.processWebhook(w, r, body, GatewayAdyen)
}

// processWebhook processes a webhook for a specific gateway
func (ws *WebhookServer) processWebhook(w http.ResponseWriter, r *http.Request, body []byte, gateway GatewayType) {
	ctx := r.Context()

	// Get signature based on gateway
	var signature string
	switch gateway {
	case GatewayStripe:
		signature = r.Header.Get("Stripe-Signature")
	case GatewayAdyen:
		signature = r.Header.Get("X-Adyen-Hmac-Signature")
	}

	// Validate signature
	if ws.config.SignatureVerification {
		if err := ws.service.ValidateWebhook(body, signature); err != nil {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse event
	event, err := ws.service.ParseWebhookEvent(body)
	if err != nil {
		http.Error(w, "Invalid event payload", http.StatusBadRequest)
		return
	}

	// Check idempotency
	if ws.isProcessed(event.ID) {
		// Already processed - return success
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process event
	if err := ws.service.HandleEvent(ctx, event); err != nil {
		// Log error but return 200 to prevent retries for non-retriable errors
		// In production, distinguish between retriable and non-retriable errors
		w.WriteHeader(http.StatusOK)
		return
	}

	// Mark as processed
	ws.markProcessed(event.ID)

	// Return success
	// Adyen expects "[accepted]" response
	if gateway == GatewayAdyen {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[accepted]"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// isProcessed checks if an event was already processed
func (ws *WebhookServer) isProcessed(eventID string) bool {
	ws.processedMu.RLock()
	defer ws.processedMu.RUnlock()
	_, exists := ws.processedEvents[eventID]
	return exists
}

// markProcessed marks an event as processed
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
// Default Event Handlers
// ============================================================================

// PaymentSucceededHandler is a default handler for payment success events
type PaymentSucceededHandler struct {
	OnSuccess func(ctx context.Context, paymentIntent PaymentIntent) error
}

// Handle processes a payment succeeded event
func (h *PaymentSucceededHandler) Handle(ctx context.Context, event WebhookEvent) error {
	if event.Type != WebhookEventPaymentIntentSucceeded {
		return nil
	}

	// Parse payment intent from event data
	var intent PaymentIntent
	if data, ok := event.Data.(map[string]interface{}); ok {
		if obj, ok := data["object"].(map[string]interface{}); ok {
			// Extract payment intent details
			if id, ok := obj["id"].(string); ok {
				intent.ID = id
			}
			if status, ok := obj["status"].(string); ok {
				intent.Status = PaymentIntentStatus(status)
			}
		}
	}

	if h.OnSuccess != nil {
		return h.OnSuccess(ctx, intent)
	}

	return nil
}

// RefundHandler is a default handler for refund events
type RefundHandler struct {
	OnRefund func(ctx context.Context, refund Refund) error
}

// Handle processes a refund event
func (h *RefundHandler) Handle(ctx context.Context, event WebhookEvent) error {
	if event.Type != WebhookEventChargeRefunded {
		return nil
	}

	// Parse refund from event data
	var refund Refund
	if data, ok := event.Data.(map[string]interface{}); ok {
		if obj, ok := data["object"].(map[string]interface{}); ok {
			if id, ok := obj["id"].(string); ok {
				refund.ID = id
			}
			if status, ok := obj["status"].(string); ok {
				refund.Status = RefundStatus(status)
			}
		}
	}

	if h.OnRefund != nil {
		return h.OnRefund(ctx, refund)
	}

	return nil
}

// DisputeHandler is a default handler for dispute events
type DisputeEventHandler struct {
	OnDispute func(ctx context.Context, dispute Dispute) error
}

// Handle processes a dispute event
func (h *DisputeEventHandler) Handle(ctx context.Context, event WebhookEvent) error {
	if event.Type != WebhookEventChargeDisputeCreated && event.Type != WebhookEventChargeDisputeClosed {
		return nil
	}

	// Parse dispute from event data
	var dispute Dispute
	if data, ok := event.Data.(map[string]interface{}); ok {
		if obj, ok := data["object"].(map[string]interface{}); ok {
			if id, ok := obj["id"].(string); ok {
				dispute.ID = id
			}
			if status, ok := obj["status"].(string); ok {
				dispute.Status = DisputeStatus(status)
			}
		}
	}

	if h.OnDispute != nil {
		return h.OnDispute(ctx, dispute)
	}

	return nil
}

// ============================================================================
// Webhook Event Builder (for testing)
// ============================================================================

// WebhookEventBuilder helps build webhook events for testing
type WebhookEventBuilder struct {
	event WebhookEvent
}

// NewWebhookEventBuilder creates a new event builder
func NewWebhookEventBuilder() *WebhookEventBuilder {
	return &WebhookEventBuilder{
		event: WebhookEvent{
			ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
			Timestamp: time.Now(),
		},
	}
}

// WithType sets the event type
func (b *WebhookEventBuilder) WithType(t WebhookEventType) *WebhookEventBuilder {
	b.event.Type = t
	return b
}

// WithGateway sets the gateway
func (b *WebhookEventBuilder) WithGateway(g GatewayType) *WebhookEventBuilder {
	b.event.Gateway = g
	return b
}

// WithPaymentIntent sets payment intent data
func (b *WebhookEventBuilder) WithPaymentIntent(intent PaymentIntent) *WebhookEventBuilder {
	b.event.Data = map[string]interface{}{
		"object": map[string]interface{}{
			"id":     intent.ID,
			"status": string(intent.Status),
			"amount": intent.Amount.Value,
		},
	}
	return b
}

// WithRefund sets refund data
func (b *WebhookEventBuilder) WithRefund(refund Refund) *WebhookEventBuilder {
	b.event.Data = map[string]interface{}{
		"object": map[string]interface{}{
			"id":     refund.ID,
			"status": string(refund.Status),
			"amount": refund.Amount.Value,
		},
	}
	return b
}

// Build returns the built event
func (b *WebhookEventBuilder) Build() WebhookEvent {
	return b.event
}
