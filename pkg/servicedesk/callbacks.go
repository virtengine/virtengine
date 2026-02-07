package servicedesk

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
	"time"

	"cosmossdk.io/log"

	"github.com/virtengine/virtengine/pkg/jira"
)

// CallbackServer handles webhooks from external service desks
type CallbackServer struct {
	bridge *Bridge
	config *Config
	logger log.Logger

	server *http.Server
	mux    *http.ServeMux

	// Nonce tracking to prevent replay attacks
	mu     sync.RWMutex
	nonces map[string]time.Time

	// Jira webhook handler
	jiraHandler jira.IWebhookHandler
}

// NewCallbackServer creates a new callback server
func NewCallbackServer(bridge *Bridge, config *Config, logger log.Logger) *CallbackServer {
	s := &CallbackServer{
		bridge: bridge,
		config: config,
		logger: logger.With("component", "callback_server"),
		nonces: make(map[string]time.Time),
	}

	// Create HTTP mux
	s.mux = http.NewServeMux()
	s.setupRoutes()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         config.WebhookConfig.ListenAddr,
		Handler:      s.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Initialize Jira webhook handler if Jira is configured
	if config.JiraConfig != nil {
		s.jiraHandler = jira.NewWebhookHandler(jira.WebhookConfig{
			Secret:           config.JiraConfig.WebhookSecret,
			RequireSignature: config.WebhookConfig.RequireSignature,
		})
		s.registerJiraHandlers()
	}

	return s
}

// setupRoutes sets up HTTP routes
func (s *CallbackServer) setupRoutes() {
	prefix := s.config.WebhookConfig.PathPrefix

	// Health check
	s.mux.HandleFunc(prefix+"/health", s.handleHealth)

	// Jira webhooks
	s.mux.HandleFunc(prefix+"/jira", s.handleJiraWebhook)

	// Waldur webhooks
	s.mux.HandleFunc(prefix+"/waldur", s.handleWaldurWebhook)

	// Generic signed callback endpoint
	s.mux.HandleFunc(prefix+"/callback", s.handleSignedCallback)
}

// Start starts the callback server
func (s *CallbackServer) Start(ctx context.Context) error {
	s.logger.Info("starting callback server", "addr", s.config.WebhookConfig.ListenAddr)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("callback server error", "error", err)
		}
	}()

	// Start nonce cleanup goroutine
	go s.cleanupNonces(ctx)

	return nil
}

// Stop stops the callback server
func (s *CallbackServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping callback server")
	return s.server.Shutdown(ctx)
}

// handleHealth handles health check requests
func (s *CallbackServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handleJiraWebhook handles Jira webhook requests
func (s *CallbackServer) handleJiraWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check IP allowlist if configured
	if !s.isAllowedIP(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delegate to Jira webhook handler
	if s.jiraHandler != nil {
		s.jiraHandler.HandleHTTP(w, r)
		return
	}

	http.Error(w, "Jira not configured", http.StatusServiceUnavailable)
}

// handleWaldurWebhook handles Waldur webhook requests
func (s *CallbackServer) handleWaldurWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check IP allowlist if configured
	if !s.isAllowedIP(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Verify signature if required
	if s.config.WebhookConfig.RequireSignature && s.config.WaldurConfig != nil {
		signature := r.Header.Get("X-Waldur-Signature")
		if !s.verifySignature(body, signature, s.config.WaldurConfig.WebhookSecret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse Waldur event
	var event WaldurWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert to callback payload
	payload := s.waldurEventToCallback(&event)
	if payload == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process callback
	ctx := r.Context()
	if err := s.bridge.HandleExternalCallback(ctx, payload); err != nil {
		s.logger.Error("failed to process waldur callback", "error", err)
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handleSignedCallback handles generic signed callback requests
func (s *CallbackServer) handleSignedCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse payload
	var payload CallbackPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate payload
	if err := payload.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check nonce (prevent replay)
	if !s.checkAndRecordNonce(payload.Nonce) {
		http.Error(w, "Duplicate nonce", http.StatusConflict)
		return
	}

	// Verify signature
	secret := s.getSecretForServiceDesk(payload.ServiceDeskType)
	if s.config.WebhookConfig.RequireSignature && secret != "" {
		if !s.verifyCallbackSignature(&payload, secret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Process callback
	ctx := r.Context()
	if err := s.bridge.HandleExternalCallback(ctx, &payload); err != nil {
		s.logger.Error("failed to process callback", "error", err)
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// registerJiraHandlers registers handlers for Jira webhook events
func (s *CallbackServer) registerJiraHandlers() {
	// Handle status changes
	s.jiraHandler.RegisterHandler(jira.WebhookEventIssueUpdated,
		jira.StatusChangeHandler(func(ctx context.Context, issueKey, fromStatus, toStatus string) error {
			s.logger.Debug("jira status change",
				"issue_key", issueKey,
				"from", fromStatus,
				"to", toStatus,
			)

			// Find the on-chain ticket ID from the Jira issue
			ticketID := s.findTicketIDFromJiraKey(ctx, issueKey)
			if ticketID == "" {
				s.logger.Warn("could not find ticket for jira issue", "issue_key", issueKey)
				return nil
			}

			// Create callback payload
			payload := &CallbackPayload{
				EventType:       "status_changed",
				ServiceDeskType: ServiceDeskJira,
				ExternalID:      issueKey,
				OnChainTicketID: ticketID,
				Changes: map[string]interface{}{
					"status":      s.config.MappingSchema.MapJiraStatusToOnChain(toStatus),
					"from_status": fromStatus,
					"to_status":   toStatus,
				},
				Timestamp: time.Now(),
				Nonce:     fmt.Sprintf("jira-%s-%d", issueKey, time.Now().UnixNano()),
			}

			return s.bridge.HandleExternalCallback(ctx, payload)
		}),
	)

	// Handle comments
	s.jiraHandler.RegisterHandler(jira.WebhookEventCommentCreated,
		jira.CommentHandler(func(ctx context.Context, issueKey string, comment *jira.Comment) error {
			s.logger.Debug("jira comment created",
				"issue_key", issueKey,
				"comment_id", comment.ID,
			)

			ticketID := s.findTicketIDFromJiraKey(ctx, issueKey)
			if ticketID == "" {
				return nil
			}

			payload := &CallbackPayload{
				EventType:       "comment_added",
				ServiceDeskType: ServiceDeskJira,
				ExternalID:      issueKey,
				OnChainTicketID: ticketID,
				Changes: map[string]interface{}{
					"comment_id":       comment.ID,
					"comment_body":     comment.PlainText(), // Note: should be encrypted before storing
					"comment_internal": comment.IsInternal(),
				},
				Timestamp: time.Now(),
				Nonce:     fmt.Sprintf("jira-comment-%s-%d", comment.ID, time.Now().UnixNano()),
			}

			return s.bridge.HandleExternalCallback(ctx, payload)
		}),
	)

	// Handle assignee changes
	s.jiraHandler.RegisterHandler(jira.WebhookEventIssueUpdated,
		jira.AssigneeChangeHandler(func(ctx context.Context, issueKey, fromAssignee, toAssignee string) error {
			s.logger.Debug("jira assignee change",
				"issue_key", issueKey,
				"from", fromAssignee,
				"to", toAssignee,
			)

			ticketID := s.findTicketIDFromJiraKey(ctx, issueKey)
			if ticketID == "" {
				return nil
			}

			payload := &CallbackPayload{
				EventType:       "assignee_changed",
				ServiceDeskType: ServiceDeskJira,
				ExternalID:      issueKey,
				OnChainTicketID: ticketID,
				Changes: map[string]interface{}{
					"from_assignee": fromAssignee,
					"to_assignee":   toAssignee,
				},
				Timestamp: time.Now(),
				Nonce:     fmt.Sprintf("jira-assign-%s-%d", issueKey, time.Now().UnixNano()),
			}

			return s.bridge.HandleExternalCallback(ctx, payload)
		}),
	)
}

// findTicketIDFromJiraKey finds the on-chain ticket ID from a Jira issue key
func (s *CallbackServer) findTicketIDFromJiraKey(ctx context.Context, jiraKey string) string {
	// In a full implementation, this would look up the mapping in persistent storage
	// For now, we'll try to extract it from the Jira issue custom fields

	if s.bridge.jiraClient == nil {
		return ""
	}

	issue, err := s.bridge.jiraClient.GetIssue(ctx, jiraKey)
	if err != nil {
		s.logger.Error("failed to get jira issue", "key", jiraKey, "error", err)
		return ""
	}

	// Look for the ticket ID in custom fields
	if cfID, ok := s.config.MappingSchema.CustomFieldMappings["ticket_id"]; ok {
		if ticketID, ok := issue.Fields.CustomFields[cfID].(string); ok {
			return ticketID
		}
	}

	// Try to extract from summary if it contains the ticket ID pattern
	if strings.HasPrefix(issue.Fields.Summary, "[TKT-") {
		end := strings.Index(issue.Fields.Summary, "]")
		if end > 1 {
			return issue.Fields.Summary[1:end]
		}
	}

	return ""
}

// WaldurWebhookEvent represents a Waldur webhook event
type WaldurWebhookEvent struct {
	EventType string                 `json:"event_type"`
	UUID      string                 `json:"uuid"`
	ObjectID  string                 `json:"object_id"`
	Hook      string                 `json:"hook"`
	Timestamp string                 `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// waldurEventToCallback converts a Waldur event to a callback payload
func (s *CallbackServer) waldurEventToCallback(event *WaldurWebhookEvent) *CallbackPayload {
	// In a full implementation, this would map Waldur event types to our callback format
	// and look up the on-chain ticket ID

	// Placeholder implementation
	ticketID := ""
	if id, ok := event.Payload["virtengine_ticket_id"].(string); ok {
		ticketID = id
	}
	if ticketID == "" {
		return nil
	}

	return &CallbackPayload{
		EventType:       event.EventType,
		ServiceDeskType: ServiceDeskWaldur,
		ExternalID:      event.ObjectID,
		OnChainTicketID: ticketID,
		Changes:         event.Payload,
		Timestamp:       time.Now(),
		Nonce:           fmt.Sprintf("waldur-%s-%d", event.UUID, time.Now().UnixNano()),
	}
}

// verifySignature verifies an HMAC signature
func (s *CallbackServer) verifySignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return true
	}
	if signature == "" {
		return false
	}

	// Remove prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// verifyCallbackSignature verifies the signature in a callback payload
func (s *CallbackServer) verifyCallbackSignature(payload *CallbackPayload, secret string) bool {
	// Compute signature over relevant fields
	toSign := fmt.Sprintf("%s:%s:%s:%s:%d",
		payload.EventType,
		payload.ServiceDeskType,
		payload.ExternalID,
		payload.OnChainTicketID,
		payload.Timestamp.Unix(),
	)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(toSign))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(payload.Signature), []byte(expectedSig))
}

// getSecretForServiceDesk returns the webhook secret for a service desk type
func (s *CallbackServer) getSecretForServiceDesk(t ServiceDeskType) string {
	switch t {
	case ServiceDeskJira:
		if s.config.JiraConfig != nil {
			return s.config.JiraConfig.WebhookSecret
		}
	case ServiceDeskWaldur:
		if s.config.WaldurConfig != nil {
			return s.config.WaldurConfig.WebhookSecret
		}
	}
	return ""
}

// checkAndRecordNonce checks if a nonce has been used and records it
func (s *CallbackServer) checkAndRecordNonce(nonce string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.nonces[nonce]; exists {
		return false // Already used
	}

	s.nonces[nonce] = time.Now()
	return true
}

// cleanupNonces removes old nonces
func (s *CallbackServer) cleanupNonces(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			cutoff := time.Now().Add(-1 * time.Hour)
			for nonce, timestamp := range s.nonces {
				if timestamp.Before(cutoff) {
					delete(s.nonces, nonce)
				}
			}
			s.mu.Unlock()
		}
	}
}

// isAllowedIP checks if the request IP is allowed
func (s *CallbackServer) isAllowedIP(r *http.Request) bool {
	if len(s.config.WebhookConfig.AllowedIPs) == 0 {
		return true // No allowlist configured
	}

	// Get client IP
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}
	clientIP = strings.TrimSpace(clientIP)

	// Remove port if present
	if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
		clientIP = clientIP[:idx]
	}

	for _, allowed := range s.config.WebhookConfig.AllowedIPs {
		if clientIP == allowed {
			return true
		}
	}

	return false
}
