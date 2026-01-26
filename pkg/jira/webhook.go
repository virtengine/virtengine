// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
// This file implements webhook handling for Jira status changes.
//
// CRITICAL: Never log API tokens or sensitive ticket content.
package jira

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// WebhookHandler handles Jira webhooks
type WebhookHandler struct {
	// secret is the webhook secret for signature verification
	// CRITICAL: Never log this value
	secret string

	// handlers are registered event handlers
	handlers map[string][]WebhookEventHandler

	// mu protects handlers map
	mu sync.RWMutex

	// config holds webhook configuration
	config WebhookConfig
}

// WebhookConfig holds webhook handler configuration
type WebhookConfig struct {
	// Secret is the webhook secret for HMAC verification
	// CRITICAL: Never log this value
	Secret string

	// RequireSignature requires webhook signature verification
	RequireSignature bool

	// AllowedIPs are allowed IP addresses (optional)
	AllowedIPs []string
}

// WebhookEventHandler is a callback for webhook events
type WebhookEventHandler func(ctx context.Context, event *WebhookEvent) error

// WebhookEventType represents webhook event types
type WebhookEventType string

const (
	// WebhookEventIssueCreated is fired when an issue is created
	WebhookEventIssueCreated WebhookEventType = "jira:issue_created"

	// WebhookEventIssueUpdated is fired when an issue is updated
	WebhookEventIssueUpdated WebhookEventType = "jira:issue_updated"

	// WebhookEventIssueDeleted is fired when an issue is deleted
	WebhookEventIssueDeleted WebhookEventType = "jira:issue_deleted"

	// WebhookEventCommentCreated is fired when a comment is created
	WebhookEventCommentCreated WebhookEventType = "comment_created"

	// WebhookEventCommentUpdated is fired when a comment is updated
	WebhookEventCommentUpdated WebhookEventType = "comment_updated"

	// WebhookEventCommentDeleted is fired when a comment is deleted
	WebhookEventCommentDeleted WebhookEventType = "comment_deleted"

	// WebhookEventWorklogCreated is fired when a worklog is created
	WebhookEventWorklogCreated WebhookEventType = "worklog_created"

	// WebhookEventWorklogUpdated is fired when a worklog is updated
	WebhookEventWorklogUpdated WebhookEventType = "worklog_updated"

	// WebhookEventWorklogDeleted is fired when a worklog is deleted
	WebhookEventWorklogDeleted WebhookEventType = "worklog_deleted"

	// WebhookEventAll matches all event types
	WebhookEventAll WebhookEventType = "*"
)

// IWebhookHandler defines the webhook handler interface
type IWebhookHandler interface {
	// HandleHTTP handles an HTTP webhook request
	HandleHTTP(w http.ResponseWriter, r *http.Request)

	// HandleEvent processes a webhook event
	HandleEvent(ctx context.Context, event *WebhookEvent) error

	// RegisterHandler registers a handler for an event type
	RegisterHandler(eventType WebhookEventType, handler WebhookEventHandler)

	// UnregisterHandlers unregisters all handlers for an event type
	UnregisterHandlers(eventType WebhookEventType)

	// ParseEvent parses a webhook event from bytes
	ParseEvent(data []byte) (*WebhookEvent, error)

	// VerifySignature verifies the webhook signature
	VerifySignature(payload []byte, signature string) bool
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(config WebhookConfig) *WebhookHandler {
	return &WebhookHandler{
		secret:   config.Secret,
		handlers: make(map[string][]WebhookEventHandler),
		config:   config,
	}
}

// Ensure WebhookHandler implements IWebhookHandler
var _ IWebhookHandler = (*WebhookHandler)(nil)

// HandleHTTP handles an HTTP webhook request
func (h *WebhookHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature if required
	if h.config.RequireSignature {
		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			signature = r.Header.Get("X-Atlassian-Webhook-Signature")
		}

		if !h.VerifySignature(body, signature) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse event
	event, err := h.ParseEvent(body)
	if err != nil {
		http.Error(w, "Failed to parse webhook event", http.StatusBadRequest)
		return
	}

	// Handle event
	ctx := r.Context()
	if err := h.HandleEvent(ctx, event); err != nil {
		// Log error but return success to Jira (to avoid retries for handler errors)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// HandleEvent processes a webhook event
func (h *WebhookHandler) HandleEvent(ctx context.Context, event *WebhookEvent) error {
	if event == nil {
		return fmt.Errorf("webhook: event is nil")
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	var handlers []WebhookEventHandler

	// Get specific handlers
	if specific, ok := h.handlers[event.WebhookEvent]; ok {
		handlers = append(handlers, specific...)
	}

	// Get wildcard handlers
	if wildcard, ok := h.handlers[string(WebhookEventAll)]; ok {
		handlers = append(handlers, wildcard...)
	}

	// Execute handlers
	var errs []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("webhook: %d handler(s) failed: %v", len(errs), errs[0])
	}

	return nil
}

// RegisterHandler registers a handler for an event type
func (h *WebhookHandler) RegisterHandler(eventType WebhookEventType, handler WebhookEventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers[string(eventType)] = append(h.handlers[string(eventType)], handler)
}

// UnregisterHandlers unregisters all handlers for an event type
func (h *WebhookHandler) UnregisterHandlers(eventType WebhookEventType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.handlers, string(eventType))
}

// ParseEvent parses a webhook event from bytes
func (h *WebhookHandler) ParseEvent(data []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("webhook: failed to parse event: %w", err)
	}

	return &event, nil
}

// VerifySignature verifies the webhook signature
func (h *WebhookHandler) VerifySignature(payload []byte, signature string) bool {
	if h.secret == "" {
		return true // No secret configured, skip verification
	}

	if signature == "" {
		return false
	}

	// Remove prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// StatusChangeHandler creates a handler for status changes
func StatusChangeHandler(callback func(ctx context.Context, issueKey, fromStatus, toStatus string) error) WebhookEventHandler {
	return func(ctx context.Context, event *WebhookEvent) error {
		if event.Issue == nil || event.Changelog == nil {
			return nil
		}

		for _, item := range event.Changelog.Items {
			if item.Field == "status" {
				return callback(ctx, event.Issue.Key, item.FromString, item.ToString)
			}
		}

		return nil
	}
}

// CommentHandler creates a handler for new comments
func CommentHandler(callback func(ctx context.Context, issueKey string, comment *Comment) error) WebhookEventHandler {
	return func(ctx context.Context, event *WebhookEvent) error {
		if event.Issue == nil || event.Comment == nil {
			return nil
		}

		return callback(ctx, event.Issue.Key, event.Comment)
	}
}

// AssigneeChangeHandler creates a handler for assignee changes
func AssigneeChangeHandler(callback func(ctx context.Context, issueKey, fromAssignee, toAssignee string) error) WebhookEventHandler {
	return func(ctx context.Context, event *WebhookEvent) error {
		if event.Issue == nil || event.Changelog == nil {
			return nil
		}

		for _, item := range event.Changelog.Items {
			if item.Field == "assignee" {
				return callback(ctx, event.Issue.Key, item.FromString, item.ToString)
			}
		}

		return nil
	}
}

// PriorityChangeHandler creates a handler for priority changes
func PriorityChangeHandler(callback func(ctx context.Context, issueKey, fromPriority, toPriority string) error) WebhookEventHandler {
	return func(ctx context.Context, event *WebhookEvent) error {
		if event.Issue == nil || event.Changelog == nil {
			return nil
		}

		for _, item := range event.Changelog.Items {
			if item.Field == "priority" {
				return callback(ctx, event.Issue.Key, item.FromString, item.ToString)
			}
		}

		return nil
	}
}

// WebhookRouter provides HTTP routing for webhooks
type WebhookRouter struct {
	handler *WebhookHandler
	path    string
}

// NewWebhookRouter creates a new webhook router
func NewWebhookRouter(handler *WebhookHandler, path string) *WebhookRouter {
	return &WebhookRouter{
		handler: handler,
		path:    path,
	}
}

// ServeHTTP implements http.Handler
func (r *WebhookRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == r.path {
		r.handler.HandleHTTP(w, req)
		return
	}

	http.NotFound(w, req)
}

// Mount registers the webhook router with an HTTP mux
func (r *WebhookRouter) Mount(mux *http.ServeMux) {
	mux.Handle(r.path, r)
}
