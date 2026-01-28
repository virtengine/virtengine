// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
// This file implements the ticket bridge between VirtEngine support requests
// and Jira Service Desk issues.
//
// CRITICAL: Never log API tokens or sensitive ticket content.
package jira

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TicketBridge handles ticket synchronization between VirtEngine and Jira
type TicketBridge struct {
	client IClient
	config TicketBridgeConfig
}

// ITicketBridge defines the ticket bridge interface
type ITicketBridge interface {
	// CreateFromSupportRequest creates a Jira issue from a VirtEngine support request
	CreateFromSupportRequest(ctx context.Context, req *VirtEngineSupportRequest) (*Issue, error)

	// UpdateFromSupportRequest updates a Jira issue from a VirtEngine support request
	UpdateFromSupportRequest(ctx context.Context, jiraKey string, req *VirtEngineSupportRequest) error

	// AddReplyToTicket adds a reply/comment to a Jira issue
	AddReplyToTicket(ctx context.Context, jiraKey string, message string, isAgent bool) (*Comment, error)

	// CloseTicket closes a Jira issue
	CloseTicket(ctx context.Context, jiraKey string, resolution string) error

	// GetTicketByVirtEngineID finds a Jira issue by VirtEngine ticket ID
	GetTicketByVirtEngineID(ctx context.Context, ticketID string) (*Issue, error)

	// SyncStatus synchronizes status from VirtEngine to Jira
	SyncStatus(ctx context.Context, jiraKey string, status string) error
}

// Ensure TicketBridge implements ITicketBridge
var _ ITicketBridge = (*TicketBridge)(nil)

// NewTicketBridge creates a new ticket bridge
func NewTicketBridge(client IClient, config TicketBridgeConfig) *TicketBridge {
	return &TicketBridge{
		client: client,
		config: config,
	}
}

// CreateFromSupportRequest creates a Jira issue from a VirtEngine support request
func (b *TicketBridge) CreateFromSupportRequest(ctx context.Context, req *VirtEngineSupportRequest) (*Issue, error) {
	if req == nil {
		return nil, fmt.Errorf("jira bridge: support request is required")
	}

	if req.Subject == "" {
		return nil, fmt.Errorf("jira bridge: subject is required")
	}

	// Map priority
	priority := b.mapPriority(req.Priority)

	// Build description with metadata
	description := b.buildDescription(req)

	// Build custom fields
	customFields := make(map[string]interface{})
	if cfID, ok := b.config.CustomFieldMappings["ticketId"]; ok {
		customFields[cfID] = req.TicketID
	}
	if cfID, ok := b.config.CustomFieldMappings["ticketNumber"]; ok {
		customFields[cfID] = req.TicketNumber
	}
	if cfID, ok := b.config.CustomFieldMappings["submitterAddress"]; ok {
		// Store truncated address for security
		customFields[cfID] = truncateAddress(req.SubmitterAddress)
	}
	if req.RelatedEntity != nil {
		if cfID, ok := b.config.CustomFieldMappings["relatedEntity"]; ok {
			customFields[cfID] = fmt.Sprintf("%s:%s", req.RelatedEntity.Type, req.RelatedEntity.ID)
		}
	}

	// Build components
	var components []Component
	if componentName, ok := b.config.CategoryToComponent[req.Category]; ok {
		components = append(components, Component{Name: componentName})
	}

	// Build labels
	labels := make([]string, 0, len(b.config.Labels)+1)
	labels = append(labels, b.config.Labels...)
	labels = append(labels, "category:"+req.Category)

	// Create issue request
	createReq := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project: Project{
				Key: b.config.ProjectKey,
			},
			Summary:     fmt.Sprintf("[%s] %s", req.TicketNumber, req.Subject),
			Description: description,
			IssueType: IssueTypeField{
				Name: string(b.config.DefaultIssueType),
			},
			Priority: &PriorityField{
				Name: string(priority),
			},
			Labels:       labels,
			Components:   components,
			CustomFields: customFields,
		},
	}

	// Create the issue
	resp, err := b.client.CreateIssue(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("jira bridge: failed to create issue: %w", err)
	}

	// Retrieve the created issue
	issue, err := b.client.GetIssue(ctx, resp.Key)
	if err != nil {
		// Return partial result if we can't get the full issue
		return &Issue{
			ID:   resp.ID,
			Key:  resp.Key,
			Self: resp.Self,
		}, nil
	}

	return issue, nil
}

// UpdateFromSupportRequest updates a Jira issue from a VirtEngine support request
func (b *TicketBridge) UpdateFromSupportRequest(ctx context.Context, jiraKey string, req *VirtEngineSupportRequest) error {
	if jiraKey == "" {
		return fmt.Errorf("jira bridge: jira key is required")
	}

	if req == nil {
		return fmt.Errorf("jira bridge: support request is required")
	}

	// Build update request
	updateReq := &UpdateIssueRequest{
		Fields: make(map[string]interface{}),
	}

	// Update priority if changed
	if req.Priority != "" {
		priority := b.mapPriority(req.Priority)
		updateReq.Fields["priority"] = map[string]string{"name": string(priority)}
	}

	// Update summary if changed
	if req.Subject != "" {
		updateReq.Fields["summary"] = fmt.Sprintf("[%s] %s", req.TicketNumber, req.Subject)
	}

	return b.client.UpdateIssue(ctx, jiraKey, updateReq)
}

// AddReplyToTicket adds a reply/comment to a Jira issue
func (b *TicketBridge) AddReplyToTicket(ctx context.Context, jiraKey string, message string, isAgent bool) (*Comment, error) {
	if jiraKey == "" {
		return nil, fmt.Errorf("jira bridge: jira key is required")
	}

	if message == "" {
		return nil, fmt.Errorf("jira bridge: message is required")
	}

	// Prefix message with source indicator
	var prefix string
	if isAgent {
		prefix = "[VirtEngine Support Agent]\n\n"
	} else {
		prefix = "[VirtEngine Customer]\n\n"
	}

	return b.client.AddComment(ctx, jiraKey, &AddCommentRequest{
		Body: prefix + message,
	})
}

// CloseTicket closes a Jira issue
func (b *TicketBridge) CloseTicket(ctx context.Context, jiraKey string, resolution string) error {
	if jiraKey == "" {
		return fmt.Errorf("jira bridge: jira key is required")
	}

	// Get available transitions
	transitions, err := b.client.GetTransitions(ctx, jiraKey)
	if err != nil {
		return fmt.Errorf("jira bridge: failed to get transitions: %w", err)
	}

	// Find close/resolve transition
	var closeTransition *Transition
	for _, t := range transitions.Transitions {
		name := strings.ToLower(t.Name)
		if name == "close" || name == "resolve" || name == "done" {
			closeTransition = &t
			break
		}
	}

	if closeTransition == nil {
		return fmt.Errorf("jira bridge: no close transition available for issue %s", jiraKey)
	}

	// Add resolution comment
	if resolution != "" {
		_, _ = b.client.AddComment(ctx, jiraKey, &AddCommentRequest{
			Body: fmt.Sprintf("[VirtEngine Resolution]\n\n%s", resolution),
		})
	}

	// Transition the issue
	return b.client.TransitionIssue(ctx, jiraKey, &TransitionRequest{
		Transition: TransitionID{ID: closeTransition.ID},
	})
}

// GetTicketByVirtEngineID finds a Jira issue by VirtEngine ticket ID
func (b *TicketBridge) GetTicketByVirtEngineID(ctx context.Context, ticketID string) (*Issue, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("jira bridge: ticket ID is required")
	}

	cfID, ok := b.config.CustomFieldMappings["ticketId"]
	if !ok {
		return nil, fmt.Errorf("jira bridge: ticketId custom field not configured")
	}

	// Search for the issue
	jql := fmt.Sprintf(`project = "%s" AND "%s" ~ "%s"`, b.config.ProjectKey, cfID, ticketID)
	result, err := b.client.SearchIssues(ctx, jql, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("jira bridge: search failed: %w", err)
	}

	if len(result.Issues) == 0 {
		return nil, nil
	}

	return &result.Issues[0], nil
}

// SyncStatus synchronizes status from VirtEngine to Jira
func (b *TicketBridge) SyncStatus(ctx context.Context, jiraKey string, status string) error {
	if jiraKey == "" {
		return fmt.Errorf("jira bridge: jira key is required")
	}

	// Map VirtEngine status to Jira status
	targetStatus := b.mapStatus(status)

	// Get current issue to check status
	issue, err := b.client.GetIssue(ctx, jiraKey)
	if err != nil {
		return fmt.Errorf("jira bridge: failed to get issue: %w", err)
	}

	// Skip if already in target status
	if issue.Fields.Status.Name == targetStatus {
		return nil
	}

	// Get available transitions
	transitions, err := b.client.GetTransitions(ctx, jiraKey)
	if err != nil {
		return fmt.Errorf("jira bridge: failed to get transitions: %w", err)
	}

	// Find matching transition
	var matchingTransition *Transition
	for _, t := range transitions.Transitions {
		if strings.EqualFold(t.To.Name, targetStatus) {
			matchingTransition = &t
			break
		}
	}

	if matchingTransition == nil {
		// No direct transition available, log but don't fail
		return nil
	}

	// Transition the issue
	return b.client.TransitionIssue(ctx, jiraKey, &TransitionRequest{
		Transition: TransitionID{ID: matchingTransition.ID},
	})
}

// mapPriority maps VirtEngine priority to Jira priority
func (b *TicketBridge) mapPriority(priority string) Priority {
	if mapped, ok := b.config.PriorityMapping[priority]; ok {
		return mapped
	}
	return PriorityMedium
}

// mapStatus maps VirtEngine status to Jira status
func (b *TicketBridge) mapStatus(status string) string {
	statusMap := map[string]string{
		"open":             "Open",
		"in_progress":      "In Progress",
		"waiting_customer": "Waiting for Customer",
		"waiting_internal": "Waiting for Support",
		"resolved":         "Resolved",
		"closed":           "Closed",
	}

	if mapped, ok := statusMap[status]; ok {
		return mapped
	}
	return "Open"
}

// buildDescription builds a Jira description from a VirtEngine support request
func (b *TicketBridge) buildDescription(req *VirtEngineSupportRequest) string {
	var sb strings.Builder

	sb.WriteString("h2. Support Request Details\n\n")

	// Add metadata
	sb.WriteString(fmt.Sprintf("*VirtEngine Ticket ID:* %s\n", req.TicketID))
	sb.WriteString(fmt.Sprintf("*Ticket Number:* %s\n", req.TicketNumber))
	sb.WriteString(fmt.Sprintf("*Category:* %s\n", req.Category))
	sb.WriteString(fmt.Sprintf("*Priority:* %s\n", req.Priority))
	sb.WriteString(fmt.Sprintf("*Submitted:* %s\n", req.CreatedAt.Format(time.RFC3339)))

	// Add related entity if present
	if req.RelatedEntity != nil {
		sb.WriteString(fmt.Sprintf("*Related Entity:* %s (%s)\n", req.RelatedEntity.ID, req.RelatedEntity.Type))
	}

	sb.WriteString("\n----\n\n")

	// Add description
	sb.WriteString("h2. Description\n\n")
	sb.WriteString(req.Description)

	sb.WriteString("\n\n----\n\n")
	sb.WriteString("_This ticket was automatically created from VirtEngine Support System via Waldur integration._\n")

	return sb.String()
}

// truncateAddress truncates a wallet address for display
func truncateAddress(address string) string {
	if len(address) <= 16 {
		return address
	}
	return address[:8] + "..." + address[len(address)-8:]
}
