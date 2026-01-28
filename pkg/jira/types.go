// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
// This package implements Jira REST API client for enterprise support ticket management,
// including ticket creation, updates, comments, and SLA tracking.
//
// CRITICAL: Never log API tokens or sensitive ticket content.
package jira

import (
	"time"
)

// IssueType represents Jira issue types
type IssueType string

const (
	// IssueTypeServiceRequest represents a service request
	IssueTypeServiceRequest IssueType = "Service Request"

	// IssueTypeIncident represents an incident
	IssueTypeIncident IssueType = "Incident"

	// IssueTypeProblem represents a problem
	IssueTypeProblem IssueType = "Problem"

	// IssueTypeChange represents a change request
	IssueTypeChange IssueType = "Change"

	// IssueTypeTask represents a task
	IssueTypeTask IssueType = "Task"
)

// Priority represents Jira priority levels
type Priority string

const (
	// PriorityHighest represents highest priority
	PriorityHighest Priority = "Highest"

	// PriorityHigh represents high priority
	PriorityHigh Priority = "High"

	// PriorityMedium represents medium priority
	PriorityMedium Priority = "Medium"

	// PriorityLow represents low priority
	PriorityLow Priority = "Low"

	// PriorityLowest represents lowest priority
	PriorityLowest Priority = "Lowest"
)

// Status represents Jira issue status
type Status string

const (
	// StatusOpen represents open status
	StatusOpen Status = "Open"

	// StatusInProgress represents in progress status
	StatusInProgress Status = "In Progress"

	// StatusWaitingForCustomer represents waiting for customer status
	StatusWaitingForCustomer Status = "Waiting for Customer"

	// StatusWaitingForSupport represents waiting for support status
	StatusWaitingForSupport Status = "Waiting for Support"

	// StatusResolved represents resolved status
	StatusResolved Status = "Resolved"

	// StatusClosed represents closed status
	StatusClosed Status = "Closed"

	// StatusCanceled represents canceled status
	StatusCanceled Status = "Canceled"
)

// Issue represents a Jira issue
type Issue struct {
	// ID is the Jira issue ID
	ID string `json:"id"`

	// Key is the Jira issue key (e.g., "PROJ-123")
	Key string `json:"key"`

	// Self is the URL to this issue
	Self string `json:"self"`

	// Fields contains the issue fields
	Fields IssueFields `json:"fields"`
}

// IssueFields represents Jira issue fields
type IssueFields struct {
	// Summary is the issue title
	Summary string `json:"summary"`

	// Description is the issue description
	Description string `json:"description,omitempty"`

	// IssueType is the type of issue
	IssueType IssueTypeField `json:"issuetype"`

	// Project is the project containing the issue
	Project Project `json:"project"`

	// Priority is the issue priority
	Priority PriorityField `json:"priority,omitempty"`

	// Status is the current status
	Status StatusField `json:"status,omitempty"`

	// Reporter is the issue reporter
	Reporter *User `json:"reporter,omitempty"`

	// Assignee is the issue assignee
	Assignee *User `json:"assignee,omitempty"`

	// Created is the creation timestamp
	Created string `json:"created,omitempty"`

	// Updated is the last update timestamp
	Updated string `json:"updated,omitempty"`

	// ResolutionDate is when the issue was resolved
	ResolutionDate string `json:"resolutiondate,omitempty"`

	// Labels are issue labels
	Labels []string `json:"labels,omitempty"`

	// Components are issue components
	Components []Component `json:"components,omitempty"`

	// CustomFields holds custom field values
	// Key is the custom field ID (e.g., "customfield_10001")
	CustomFields map[string]interface{} `json:"-"`
}

// IssueTypeField represents the issue type field
type IssueTypeField struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Self string `json:"self,omitempty"`
}

// Project represents a Jira project
type Project struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// PriorityField represents the priority field
type PriorityField struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Self string `json:"self,omitempty"`
}

// StatusField represents the status field
type StatusField struct {
	ID             string         `json:"id,omitempty"`
	Name           string         `json:"name"`
	Self           string         `json:"self,omitempty"`
	StatusCategory StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory represents a status category
type StatusCategory struct {
	ID   int    `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

// User represents a Jira user
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Self         string `json:"self,omitempty"`
}

// Component represents a Jira component
type Component struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Self string `json:"self,omitempty"`
}

// Comment represents a Jira comment
type Comment struct {
	ID      string `json:"id,omitempty"`
	Self    string `json:"self,omitempty"`
	Author  *User  `json:"author,omitempty"`
	Body    string `json:"body"`
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// CommentResponse represents a paginated comment response
type CommentResponse struct {
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	Comments   []Comment `json:"comments"`
}

// Transition represents a Jira workflow transition
type Transition struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	To   StatusField `json:"to"`
}

// TransitionsResponse represents available transitions
type TransitionsResponse struct {
	Transitions []Transition `json:"transitions"`
}

// SearchResult represents Jira search results
type SearchResult struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

// CreateIssueRequest represents a request to create an issue
type CreateIssueRequest struct {
	// Fields contains the issue fields
	Fields CreateIssueFields `json:"fields"`
}

// CreateIssueFields contains fields for creating an issue
type CreateIssueFields struct {
	Project     Project        `json:"project"`
	Summary     string         `json:"summary"`
	Description string         `json:"description,omitempty"`
	IssueType   IssueTypeField `json:"issuetype"`
	Priority    *PriorityField `json:"priority,omitempty"`
	Assignee    *User          `json:"assignee,omitempty"`
	Reporter    *User          `json:"reporter,omitempty"`
	Labels      []string       `json:"labels,omitempty"`
	Components  []Component    `json:"components,omitempty"`

	// CustomFields for additional fields
	CustomFields map[string]interface{} `json:"-"`
}

// UpdateIssueRequest represents a request to update an issue
type UpdateIssueRequest struct {
	// Fields to set
	Fields map[string]interface{} `json:"fields,omitempty"`

	// Update operations
	Update map[string][]UpdateOperation `json:"update,omitempty"`
}

// UpdateOperation represents an update operation
type UpdateOperation struct {
	Set    interface{} `json:"set,omitempty"`
	Add    interface{} `json:"add,omitempty"`
	Remove interface{} `json:"remove,omitempty"`
}

// TransitionRequest represents a request to transition an issue
type TransitionRequest struct {
	Transition TransitionID           `json:"transition"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Update     map[string]interface{} `json:"update,omitempty"`
}

// TransitionID represents a transition ID
type TransitionID struct {
	ID string `json:"id"`
}

// AddCommentRequest represents a request to add a comment
type AddCommentRequest struct {
	Body string `json:"body"`
}

// CreateIssueResponse represents the response from creating an issue
type CreateIssueResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

// ErrorResponse represents a Jira API error response
type ErrorResponse struct {
	ErrorMessages []string          `json:"errorMessages,omitempty"`
	Errors        map[string]string `json:"errors,omitempty"`
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	if len(e.ErrorMessages) > 0 {
		return e.ErrorMessages[0]
	}
	for _, msg := range e.Errors {
		return msg
	}
	return "unknown Jira error"
}

// WebhookEvent represents a Jira webhook event
type WebhookEvent struct {
	// Timestamp is when the event occurred
	Timestamp int64 `json:"timestamp"`

	// WebhookEvent is the event type
	WebhookEvent string `json:"webhookEvent"`

	// IssueEventTypeName is the specific issue event type
	IssueEventTypeName string `json:"issue_event_type_name,omitempty"`

	// User who triggered the event
	User *User `json:"user,omitempty"`

	// Issue is the affected issue
	Issue *Issue `json:"issue,omitempty"`

	// Comment is the affected comment (for comment events)
	Comment *Comment `json:"comment,omitempty"`

	// Changelog contains field changes
	Changelog *Changelog `json:"changelog,omitempty"`
}

// Changelog represents issue field changes
type Changelog struct {
	ID    string          `json:"id"`
	Items []ChangelogItem `json:"items"`
}

// ChangelogItem represents a single field change
type ChangelogItem struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype"`
	FieldID    string `json:"fieldId,omitempty"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
}

// SLAInfo represents SLA information for a ticket
type SLAInfo struct {
	// TicketKey is the Jira issue key
	TicketKey string `json:"ticketKey"`

	// ResponseSLA is the response time SLA
	ResponseSLA *SLAMetric `json:"responseSLA,omitempty"`

	// ResolutionSLA is the resolution time SLA
	ResolutionSLA *SLAMetric `json:"resolutionSLA,omitempty"`

	// CustomSLAs are additional SLA metrics
	CustomSLAs map[string]*SLAMetric `json:"customSLAs,omitempty"`
}

// SLAMetric represents a single SLA metric
type SLAMetric struct {
	// Name is the SLA name
	Name string `json:"name"`

	// TargetDuration is the target duration in minutes
	TargetDuration int64 `json:"targetDuration"`

	// ElapsedDuration is the elapsed time in minutes
	ElapsedDuration int64 `json:"elapsedDuration"`

	// RemainingDuration is the remaining time in minutes (can be negative if breached)
	RemainingDuration int64 `json:"remainingDuration"`

	// Breached indicates if the SLA was breached
	Breached bool `json:"breached"`

	// Paused indicates if the SLA clock is paused
	Paused bool `json:"paused"`

	// CompletedAt is when the SLA was completed
	CompletedAt *time.Time `json:"completedAt,omitempty"`

	// BreachedAt is when the SLA was breached
	BreachedAt *time.Time `json:"breachedAt,omitempty"`
}

// VirtEngineSupportRequest represents a support request from VirtEngine
type VirtEngineSupportRequest struct {
	// TicketID is the on-chain ticket ID
	TicketID string `json:"ticketId"`

	// TicketNumber is the human-readable ticket number
	TicketNumber string `json:"ticketNumber"`

	// SubmitterAddress is the wallet address of the submitter
	SubmitterAddress string `json:"submitterAddress"`

	// Category is the ticket category
	Category string `json:"category"`

	// Priority is the ticket priority
	Priority string `json:"priority"`

	// Subject is the ticket subject
	Subject string `json:"subject"`

	// Description is the decrypted ticket description
	// CRITICAL: This should only be provided after secure decryption
	Description string `json:"description"`

	// RelatedEntity contains related entity information
	RelatedEntity *RelatedEntity `json:"relatedEntity,omitempty"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"createdAt"`
}

// RelatedEntity represents a related entity in a support request
type RelatedEntity struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// TicketBridgeConfig holds configuration for the ticket bridge
type TicketBridgeConfig struct {
	// ProjectKey is the Jira project key
	ProjectKey string `json:"projectKey"`

	// DefaultIssueType is the default issue type for new tickets
	DefaultIssueType IssueType `json:"defaultIssueType"`

	// CategoryToComponent maps VirtEngine categories to Jira components
	CategoryToComponent map[string]string `json:"categoryToComponent"`

	// PriorityMapping maps VirtEngine priorities to Jira priorities
	PriorityMapping map[string]Priority `json:"priorityMapping"`

	// CustomFieldMappings maps VirtEngine fields to Jira custom fields
	CustomFieldMappings map[string]string `json:"customFieldMappings"`

	// Labels are default labels to apply
	Labels []string `json:"labels"`
}

// DefaultTicketBridgeConfig returns default bridge configuration
func DefaultTicketBridgeConfig() TicketBridgeConfig {
	return TicketBridgeConfig{
		ProjectKey:       "VESUPPORT",
		DefaultIssueType: IssueTypeServiceRequest,
		CategoryToComponent: map[string]string{
			"account":     "Account Management",
			"identity":    "Identity & VEID",
			"billing":     "Billing & Payments",
			"provider":    "Provider Services",
			"marketplace": "Marketplace",
			"technical":   "Technical Support",
			"security":    "Security",
			"other":       "General",
		},
		PriorityMapping: map[string]Priority{
			"low":    PriorityLow,
			"medium": PriorityMedium,
			"high":   PriorityHigh,
			"urgent": PriorityHighest,
		},
		CustomFieldMappings: map[string]string{
			"ticketId":         "customfield_10100",
			"ticketNumber":     "customfield_10101",
			"submitterAddress": "customfield_10102",
			"relatedEntity":    "customfield_10103",
		},
		Labels: []string{"virtengine", "waldur"},
	}
}
