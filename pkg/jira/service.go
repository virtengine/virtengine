// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
// This file implements the main Jira service that coordinates all components.
//
// CRITICAL: Never log API tokens or sensitive ticket content.
package jira

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Service is the main Jira Service Desk integration service
type Service struct {
	client     IClient
	bridge     ITicketBridge
	slaTracker ISLATracker
	webhook    IWebhookHandler
	config     ServiceConfig
	mu         sync.RWMutex
	started    bool
	stopCh     chan struct{}
}

// ServiceConfig holds service configuration
type ServiceConfig struct {
	// Client configuration
	ClientConfig ClientConfig

	// Bridge configuration
	BridgeConfig TicketBridgeConfig

	// SLA configuration
	SLAConfig SLAConfig

	// Webhook configuration
	WebhookConfig WebhookConfig

	// SLACheckInterval is the interval for checking SLA breaches
	SLACheckInterval time.Duration

	// OnSLABreach is called when an SLA is breached
	OnSLABreach func(ctx context.Context, ticket *TicketSLA) error

	// OnStatusChange is called when a Jira status changes
	OnStatusChange func(ctx context.Context, issueKey, fromStatus, toStatus string) error

	// OnNewComment is called when a new comment is added in Jira
	OnNewComment func(ctx context.Context, issueKey string, comment *Comment) error
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		ClientConfig:     DefaultClientConfig(),
		BridgeConfig:     DefaultTicketBridgeConfig(),
		SLAConfig:        DefaultSLAConfig(),
		SLACheckInterval: 5 * time.Minute,
	}
}

// IService defines the main service interface
type IService interface {
	// Start starts the service
	Start(ctx context.Context) error

	// Stop stops the service
	Stop(ctx context.Context) error

	// CreateTicket creates a Jira ticket from a VirtEngine support request
	CreateTicket(ctx context.Context, req *VirtEngineSupportRequest) (*Issue, error)

	// UpdateTicket updates a Jira ticket
	UpdateTicket(ctx context.Context, ticketID string, req *VirtEngineSupportRequest) error

	// AddReply adds a reply to a ticket
	AddReply(ctx context.Context, ticketID string, message string, isAgent bool) (*Comment, error)

	// CloseTicket closes a ticket
	CloseTicket(ctx context.Context, ticketID string, resolution string) error

	// GetTicket retrieves a ticket by VirtEngine ID
	GetTicket(ctx context.Context, ticketID string) (*Issue, error)

	// GetSLAInfo retrieves SLA information for a ticket
	GetSLAInfo(ticketID string) (*SLAInfo, error)

	// GetSLAMetrics retrieves aggregated SLA metrics
	GetSLAMetrics() *SLAMetrics

	// GetWebhookHandler returns the webhook handler for HTTP integration
	GetWebhookHandler() IWebhookHandler

	// Client returns the underlying Jira client
	Client() IClient
}

// NewService creates a new Jira service
func NewService(config ServiceConfig) (*Service, error) {
	// Create client
	client, err := NewClient(config.ClientConfig)
	if err != nil {
		return nil, fmt.Errorf("jira service: failed to create client: %w", err)
	}

	// Create bridge
	bridge := NewTicketBridge(client, config.BridgeConfig)

	// Create SLA tracker
	slaTracker := NewSLATracker(config.SLAConfig)

	// Create webhook handler
	webhook := NewWebhookHandler(config.WebhookConfig)

	return &Service{
		client:     client,
		bridge:     bridge,
		slaTracker: slaTracker,
		webhook:    webhook,
		config:     config,
		stopCh:     make(chan struct{}),
	}, nil
}

// Ensure Service implements IService
var _ IService = (*Service)(nil)

// Start starts the service
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("jira service: already started")
	}

	// Register webhook handlers
	s.registerWebhookHandlers()

	// Start SLA monitoring
	go s.monitorSLAs(ctx)

	s.started = true
	return nil
}

// Stop stops the service
func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	close(s.stopCh)
	s.started = false
	return nil
}

// CreateTicket creates a Jira ticket from a VirtEngine support request
func (s *Service) CreateTicket(ctx context.Context, req *VirtEngineSupportRequest) (*Issue, error) {
	// Create the issue in Jira
	issue, err := s.bridge.CreateFromSupportRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("jira service: failed to create ticket: %w", err)
	}

	// Start SLA tracking (non-critical, errors are silently ignored)
	_ = s.slaTracker.StartTracking(req.TicketID, issue.Key, req.Priority, req.CreatedAt)

	return issue, nil
}

// UpdateTicket updates a Jira ticket
func (s *Service) UpdateTicket(ctx context.Context, ticketID string, req *VirtEngineSupportRequest) error {
	// Find the Jira issue
	issue, err := s.bridge.GetTicketByVirtEngineID(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("jira service: failed to find ticket: %w", err)
	}

	if issue == nil {
		return fmt.Errorf("jira service: ticket not found: %s", ticketID)
	}

	return s.bridge.UpdateFromSupportRequest(ctx, issue.Key, req)
}

// AddReply adds a reply to a ticket
func (s *Service) AddReply(ctx context.Context, ticketID string, message string, isAgent bool) (*Comment, error) {
	// Find the Jira issue
	issue, err := s.bridge.GetTicketByVirtEngineID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("jira service: failed to find ticket: %w", err)
	}

	if issue == nil {
		return nil, fmt.Errorf("jira service: ticket not found: %s", ticketID)
	}

	// Add the comment
	comment, err := s.bridge.AddReplyToTicket(ctx, issue.Key, message, isAgent)
	if err != nil {
		return nil, fmt.Errorf("jira service: failed to add reply: %w", err)
	}

	// Record first response if this is an agent reply
	if isAgent {
		_ = s.slaTracker.RecordFirstResponse(ticketID, time.Now())
	}

	return comment, nil
}

// CloseTicket closes a ticket
func (s *Service) CloseTicket(ctx context.Context, ticketID string, resolution string) error {
	// Find the Jira issue
	issue, err := s.bridge.GetTicketByVirtEngineID(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("jira service: failed to find ticket: %w", err)
	}

	if issue == nil {
		return fmt.Errorf("jira service: ticket not found: %s", ticketID)
	}

	// Close the ticket
	if err := s.bridge.CloseTicket(ctx, issue.Key, resolution); err != nil {
		return fmt.Errorf("jira service: failed to close ticket: %w", err)
	}

	// Record resolution
	_ = s.slaTracker.RecordResolution(ticketID, time.Now())

	return nil
}

// GetTicket retrieves a ticket by VirtEngine ID
func (s *Service) GetTicket(ctx context.Context, ticketID string) (*Issue, error) {
	issue, err := s.bridge.GetTicketByVirtEngineID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("jira service: failed to get ticket: %w", err)
	}

	return issue, nil
}

// GetSLAInfo retrieves SLA information for a ticket
func (s *Service) GetSLAInfo(ticketID string) (*SLAInfo, error) {
	return s.slaTracker.GetSLAInfo(ticketID)
}

// GetSLAMetrics retrieves aggregated SLA metrics
func (s *Service) GetSLAMetrics() *SLAMetrics {
	return s.slaTracker.GetMetrics()
}

// GetWebhookHandler returns the webhook handler
func (s *Service) GetWebhookHandler() IWebhookHandler {
	return s.webhook
}

// Client returns the underlying Jira client
func (s *Service) Client() IClient {
	return s.client
}

// registerWebhookHandlers registers webhook event handlers
func (s *Service) registerWebhookHandlers() {
	// Status change handler
	if s.config.OnStatusChange != nil {
		s.webhook.RegisterHandler(WebhookEventIssueUpdated, StatusChangeHandler(func(ctx context.Context, issueKey, fromStatus, toStatus string) error {
			// Handle SLA pausing/resuming based on status
			s.handleStatusChangeForSLA(issueKey, fromStatus, toStatus)

			// Call user callback
			return s.config.OnStatusChange(ctx, issueKey, fromStatus, toStatus)
		}))
	} else {
		// Still register for SLA handling
		s.webhook.RegisterHandler(WebhookEventIssueUpdated, StatusChangeHandler(func(ctx context.Context, issueKey, fromStatus, toStatus string) error {
			s.handleStatusChangeForSLA(issueKey, fromStatus, toStatus)
			return nil
		}))
	}

	// Comment handler
	if s.config.OnNewComment != nil {
		s.webhook.RegisterHandler(WebhookEventCommentCreated, CommentHandler(func(ctx context.Context, issueKey string, comment *Comment) error {
			return s.config.OnNewComment(ctx, issueKey, comment)
		}))
	}
}

// handleStatusChangeForSLA handles SLA pausing/resuming based on status
func (s *Service) handleStatusChangeForSLA(issueKey, fromStatus, toStatus string) {
	// Find ticket by Jira key
	// This is a simplified approach - in production, you'd want a reverse mapping
	for _, info := range s.slaTracker.GetAllSLAInfo() {
		if info.TicketKey == issueKey {
			ticketID := "" // Would need to be extracted from custom field

			// Pause SLA when waiting for customer
			if toStatus == "Waiting for Customer" || toStatus == "Waiting for customer" {
				_ = s.slaTracker.PauseSLA(ticketID)
			}

			// Resume SLA when no longer waiting
			if fromStatus == "Waiting for Customer" || fromStatus == "Waiting for customer" {
				_ = s.slaTracker.ResumeSLA(ticketID)
			}

			break
		}
	}
}

// monitorSLAs periodically checks for SLA breaches
func (s *Service) monitorSLAs(ctx context.Context) {
	interval := s.config.SLACheckInterval
	if interval == 0 {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			breached, err := s.slaTracker.CheckSLABreaches(ctx)
			if err != nil {
				continue
			}

			// Call breach callback for each breached ticket
			if s.config.OnSLABreach != nil {
				for _, ticket := range breached {
					_ = s.config.OnSLABreach(ctx, ticket)
				}
			}
		}
	}
}

// SyncVirtEngineTicket synchronizes a VirtEngine ticket status to Jira
func (s *Service) SyncVirtEngineTicket(ctx context.Context, ticketID, status string) error {
	// Find the Jira issue
	issue, err := s.bridge.GetTicketByVirtEngineID(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("jira service: failed to find ticket: %w", err)
	}

	if issue == nil {
		return fmt.Errorf("jira service: ticket not found: %s", ticketID)
	}

	return s.bridge.SyncStatus(ctx, issue.Key, status)
}

// RecordAgentResponse records that an agent has responded to a ticket
func (s *Service) RecordAgentResponse(ticketID string) error {
	return s.slaTracker.RecordFirstResponse(ticketID, time.Now())
}

// PauseTicketSLA pauses SLA tracking for a ticket
func (s *Service) PauseTicketSLA(ticketID string) error {
	return s.slaTracker.PauseSLA(ticketID)
}

// ResumeTicketSLA resumes SLA tracking for a ticket
func (s *Service) ResumeTicketSLA(ticketID string) error {
	return s.slaTracker.ResumeSLA(ticketID)
}

