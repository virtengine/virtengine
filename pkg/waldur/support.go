// Package waldur provides Support Desk API for Waldur integration.
//
// VE-12B: Waldur Support Desk integration for ticket management.
// Waldur MasterMind has a built-in support module for issue tracking.
package waldur

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SupportClient provides support desk operations through Waldur
type SupportClient struct {
	client *Client
}

// NewSupportClient creates a new support client
func NewSupportClient(c *Client) *SupportClient {
	return &SupportClient{client: c}
}

// IssueType represents a Waldur support issue type
type IssueType string

const (
	// IssueTypeIncident is for service disruptions
	IssueTypeIncident IssueType = "Incident"

	// IssueTypeServiceRequest is for service requests
	IssueTypeServiceRequest IssueType = "Service Request"

	// IssueTypeInformational is for informational inquiries
	IssueTypeInformational IssueType = "Informational"

	// IssueTypeChange is for change requests
	IssueTypeChange IssueType = "Change Request"
)

// IssuePriority represents a Waldur support issue priority
type IssuePriority string

const (
	// PriorityLow is for low priority issues
	PriorityLow IssuePriority = "low"

	// PriorityNormal is for normal priority issues
	PriorityNormal IssuePriority = "normal"

	// PriorityHigh is for high priority issues
	PriorityHigh IssuePriority = "high"

	// PriorityCritical is for critical priority issues
	PriorityCritical IssuePriority = "critical"
)

// IssueState represents a Waldur support issue state
type IssueState string

const (
	// StateNew is for new issues
	StateNew IssueState = "new"

	// StateOpen is for open issues
	StateOpen IssueState = "open"

	// StateInProgress is for issues being worked on
	StateInProgress IssueState = "in_progress"

	// StateWaiting is for issues waiting on customer
	StateWaiting IssueState = "waiting"

	// StateResolved is for resolved issues
	StateResolved IssueState = "resolved"

	// StateClosed is for closed issues
	StateClosed IssueState = "closed"

	// StateCanceled is for canceled issues
	StateCanceled IssueState = "canceled"
)

// SupportIssue represents a Waldur support issue
type SupportIssue struct {
	// UUID is the Waldur issue UUID
	UUID string `json:"uuid"`

	// Key is the human-readable issue key (e.g., "SUPPORT-123")
	Key string `json:"key"`

	// Type is the issue type
	Type IssueType `json:"type"`

	// Priority is the issue priority
	Priority IssuePriority `json:"priority"`

	// State is the current state
	State IssueState `json:"status"`

	// Summary is the issue summary/title
	Summary string `json:"summary"`

	// Description is the detailed description
	Description string `json:"description"`

	// CallerUUID is the customer/caller UUID
	CallerUUID string `json:"caller_uuid,omitempty"`

	// CallerName is the customer/caller name
	CallerName string `json:"caller_name,omitempty"`

	// AssigneeUUID is the assigned support agent UUID
	AssigneeUUID string `json:"assignee_uuid,omitempty"`

	// AssigneeName is the assigned support agent name
	AssigneeName string `json:"assignee_name,omitempty"`

	// CustomerUUID is the Waldur customer UUID
	CustomerUUID string `json:"customer_uuid,omitempty"`

	// ProjectUUID is the Waldur project UUID
	ProjectUUID string `json:"project_uuid,omitempty"`

	// ResourceUUID is the related resource UUID
	ResourceUUID string `json:"resource_uuid,omitempty"`

	// BackendID is the external system ID (VirtEngine ticket ID)
	BackendID string `json:"backend_id,omitempty"`

	// Resolution is the resolution text
	Resolution string `json:"resolution,omitempty"`

	// CreatedAt is when the issue was created
	CreatedAt time.Time `json:"created"`

	// ModifiedAt is when the issue was last modified
	ModifiedAt time.Time `json:"modified,omitempty"`

	// ResolvedAt is when the issue was resolved
	ResolvedAt *time.Time `json:"resolved,omitempty"`
}

// SupportComment represents a Waldur support issue comment
type SupportComment struct {
	// UUID is the comment UUID
	UUID string `json:"uuid"`

	// IssueUUID is the parent issue UUID
	IssueUUID string `json:"issue_uuid"`

	// Author is the comment author
	Author string `json:"author"`

	// AuthorUUID is the comment author UUID
	AuthorUUID string `json:"author_uuid"`

	// Description is the comment text
	Description string `json:"description"`

	// IsPublic indicates if the comment is visible to the customer
	IsPublic bool `json:"is_public"`

	// CreatedAt is when the comment was created
	CreatedAt time.Time `json:"created"`
}

// CreateIssueRequest contains parameters for creating a support issue
type CreateIssueRequest struct {
	// Type is the issue type
	Type IssueType `json:"type"`

	// Priority is the issue priority
	Priority IssuePriority `json:"priority"`

	// Summary is the issue summary/title
	Summary string `json:"summary"`

	// Description is the detailed description
	Description string `json:"description"`

	// CallerUUID is the customer/caller UUID
	CallerUUID string `json:"caller,omitempty"`

	// CustomerUUID is the Waldur customer UUID
	CustomerUUID string `json:"customer,omitempty"`

	// ProjectUUID is the Waldur project UUID
	ProjectUUID string `json:"project,omitempty"`

	// ResourceUUID is the related resource UUID
	ResourceUUID string `json:"resource,omitempty"`

	// BackendID is the external system ID (VirtEngine ticket ID)
	BackendID string `json:"backend_id,omitempty"`
}

// UpdateIssueRequest contains parameters for updating a support issue
type UpdateIssueRequest struct {
	// Summary is the updated summary
	Summary string `json:"summary,omitempty"`

	// Description is the updated description
	Description string `json:"description,omitempty"`

	// Priority is the updated priority
	Priority IssuePriority `json:"priority,omitempty"`

	// AssigneeUUID is the assignee to set
	AssigneeUUID string `json:"assignee,omitempty"`
}

// ListIssuesParams contains parameters for listing support issues
type ListIssuesParams struct {
	// CustomerUUID filters by customer
	CustomerUUID string

	// ProjectUUID filters by project
	ProjectUUID string

	// State filters by state
	State IssueState

	// Priority filters by priority
	Priority IssuePriority

	// Type filters by type
	Type IssueType

	// CallerUUID filters by caller
	CallerUUID string

	// BackendID filters by backend ID
	BackendID string

	// Page is the page number
	Page int

	// PageSize is the page size
	PageSize int
}

// AddCommentRequest contains parameters for adding a comment
type AddCommentRequest struct {
	// Description is the comment text
	Description string `json:"description"`

	// IsPublic indicates if the comment is visible to the customer
	IsPublic bool `json:"is_public"`
}

// CreateIssue creates a new support issue in Waldur
func (s *SupportClient) CreateIssue(ctx context.Context, req CreateIssueRequest) (*SupportIssue, error) {
	if req.Summary == "" {
		return nil, fmt.Errorf("summary is required")
	}
	if req.Type == "" {
		req.Type = IssueTypeServiceRequest
	}
	if req.Priority == "" {
		req.Priority = PriorityNormal
	}

	var issue *SupportIssue

	err := s.client.doWithRetry(ctx, func() error {
		body := map[string]interface{}{
			"summary":     req.Summary,
			"description": req.Description,
			"type":        string(req.Type),
			"priority":    string(req.Priority),
		}

		if req.CallerUUID != "" {
			body["caller"] = req.CallerUUID
		}
		if req.CustomerUUID != "" {
			body["customer"] = req.CustomerUUID
		}
		if req.ProjectUUID != "" {
			body["project"] = req.ProjectUUID
		}
		if req.ResourceUUID != "" {
			body["resource"] = req.ResourceUUID
		}
		if req.BackendID != "" {
			body["backend_id"] = req.BackendID
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodPost, "/support-issues/", bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusCreated && statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &issue); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return issue, err
}

// GetIssue retrieves a support issue by UUID
func (s *SupportClient) GetIssue(ctx context.Context, issueUUID string) (*SupportIssue, error) {
	if issueUUID == "" {
		return nil, fmt.Errorf("issue UUID is required")
	}

	var issue *SupportIssue

	err := s.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/support-issues/%s/", issueUUID)
		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if statusCode == http.StatusNotFound {
			return ErrNotFound
		}
		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &issue); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return issue, err
}

// GetIssueByBackendID finds a Waldur issue by its backend ID (VirtEngine ticket ID)
func (s *SupportClient) GetIssueByBackendID(ctx context.Context, backendID string) (*SupportIssue, error) {
	if backendID == "" {
		return nil, fmt.Errorf("backend ID is required")
	}

	var issue *SupportIssue

	err := s.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/support-issues/?backend_id=%s", backendID)
		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		// Waldur returns an array
		var issues []SupportIssue
		if err := json.Unmarshal(respBody, &issues); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		if len(issues) == 0 {
			return ErrNotFound
		}

		issue = &issues[0]
		return nil
	})

	return issue, err
}

// ListIssues lists support issues with filtering
func (s *SupportClient) ListIssues(ctx context.Context, params ListIssuesParams) ([]SupportIssue, error) {
	var issues []SupportIssue

	err := s.client.doWithRetry(ctx, func() error {
		path := "/support-issues/?"

		if params.CustomerUUID != "" {
			path += fmt.Sprintf("customer_uuid=%s&", params.CustomerUUID)
		}
		if params.ProjectUUID != "" {
			path += fmt.Sprintf("project_uuid=%s&", params.ProjectUUID)
		}
		if params.State != "" {
			path += fmt.Sprintf("status=%s&", string(params.State))
		}
		if params.Priority != "" {
			path += fmt.Sprintf("priority=%s&", string(params.Priority))
		}
		if params.Type != "" {
			path += fmt.Sprintf("type=%s&", string(params.Type))
		}
		if params.CallerUUID != "" {
			path += fmt.Sprintf("caller_uuid=%s&", params.CallerUUID)
		}
		if params.BackendID != "" {
			path += fmt.Sprintf("backend_id=%s&", params.BackendID)
		}
		if params.Page > 0 {
			path += fmt.Sprintf("page=%d&", params.Page)
		}
		if params.PageSize > 0 {
			path += fmt.Sprintf("page_size=%d&", params.PageSize)
		}

		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &issues); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return issues, err
}

// UpdateIssue updates a support issue
func (s *SupportClient) UpdateIssue(ctx context.Context, issueUUID string, req UpdateIssueRequest) (*SupportIssue, error) {
	if issueUUID == "" {
		return nil, fmt.Errorf("issue UUID is required")
	}

	var issue *SupportIssue

	err := s.client.doWithRetry(ctx, func() error {
		body := make(map[string]interface{})

		if req.Summary != "" {
			body["summary"] = req.Summary
		}
		if req.Description != "" {
			body["description"] = req.Description
		}
		if req.Priority != "" {
			body["priority"] = string(req.Priority)
		}
		if req.AssigneeUUID != "" {
			body["assignee"] = req.AssigneeUUID
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		path := fmt.Sprintf("/support-issues/%s/", issueUUID)
		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodPatch, path, bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &issue); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return issue, err
}

// SetIssueState changes the state of a support issue
func (s *SupportClient) SetIssueState(ctx context.Context, issueUUID string, state IssueState, resolution string) error {
	if issueUUID == "" {
		return fmt.Errorf("issue UUID is required")
	}

	// Map state to action
	var action string
	switch state {
	case StateInProgress:
		action = "start_processing"
	case StateWaiting:
		action = "request_info"
	case StateResolved:
		action = "resolve"
	case StateClosed:
		action = "close"
	case StateCanceled:
		action = "cancel"
	default:
		return fmt.Errorf("unsupported state transition: %s", state)
	}

	return s.client.doWithRetry(ctx, func() error {
		var bodyBytes []byte
		var err error

		if resolution != "" && (state == StateResolved || state == StateClosed) {
			body := map[string]interface{}{
				"resolution": resolution,
			}
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return fmt.Errorf("marshal request: %w", err)
			}
		}

		path := fmt.Sprintf("/support-issues/%s/%s/", issueUUID, action)
		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodPost, path, bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusAccepted {
			return mapHTTPError(statusCode, respBody)
		}

		return nil
	})
}

// AddComment adds a comment to a support issue
func (s *SupportClient) AddComment(ctx context.Context, issueUUID string, req AddCommentRequest) (*SupportComment, error) {
	if issueUUID == "" {
		return nil, fmt.Errorf("issue UUID is required")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	var comment *SupportComment

	err := s.client.doWithRetry(ctx, func() error {
		body := map[string]interface{}{
			"description": req.Description,
			"is_public":   req.IsPublic,
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		path := fmt.Sprintf("/support-issues/%s/comment/", issueUUID)
		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodPost, path, bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusCreated && statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &comment); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return comment, err
}

// GetComments retrieves all comments for a support issue
func (s *SupportClient) GetComments(ctx context.Context, issueUUID string) ([]SupportComment, error) {
	if issueUUID == "" {
		return nil, fmt.Errorf("issue UUID is required")
	}

	var comments []SupportComment

	err := s.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/support-comments/?issue_uuid=%s", issueUUID)
		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &comments); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return comments, err
}

// DeleteIssue deletes a support issue (if allowed by Waldur)
func (s *SupportClient) DeleteIssue(ctx context.Context, issueUUID string) error {
	if issueUUID == "" {
		return fmt.Errorf("issue UUID is required")
	}

	return s.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/support-issues/%s/", issueUUID)
		respBody, statusCode, err := s.client.doRequest(ctx, http.MethodDelete, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusNoContent && statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		return nil
	})
}

// MapVirtEnginePriorityToWaldur maps VirtEngine priority to Waldur priority
func MapVirtEnginePriorityToWaldur(priority string) IssuePriority {
	switch priority {
	case "low":
		return PriorityLow
	case "normal", "medium":
		return PriorityNormal
	case "high":
		return PriorityHigh
	case "urgent", "critical":
		return PriorityCritical
	default:
		return PriorityNormal
	}
}

// MapWaldurPriorityToVirtEngine maps Waldur priority to VirtEngine priority
func MapWaldurPriorityToVirtEngine(priority IssuePriority) string {
	switch priority {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "urgent"
	default:
		return "normal"
	}
}

// MapVirtEngineStatusToWaldur maps VirtEngine status to Waldur state
func MapVirtEngineStatusToWaldur(status string) IssueState {
	switch status {
	case "open":
		return StateNew
	case "assigned":
		return StateOpen
	case "in_progress":
		return StateInProgress
	case "pending_customer":
		return StateWaiting
	case "resolved":
		return StateResolved
	case "closed":
		return StateClosed
	case "canceled":
		return StateCanceled
	default:
		return StateNew
	}
}

// MapWaldurStateToVirtEngine maps Waldur state to VirtEngine status
func MapWaldurStateToVirtEngine(state IssueState) string {
	switch state {
	case StateNew:
		return "open"
	case StateOpen:
		return "assigned"
	case StateInProgress:
		return "in_progress"
	case StateWaiting:
		return "pending_customer"
	case StateResolved:
		return "resolved"
	case StateClosed:
		return "closed"
	case StateCanceled:
		return "canceled"
	default:
		return "open"
	}
}

// MapVirtEngineCategoryToWaldurType maps VirtEngine category to Waldur issue type
func MapVirtEngineCategoryToWaldurType(category string) IssueType {
	switch category {
	case "incident", "outage", "security":
		return IssueTypeIncident
	case "change", "upgrade":
		return IssueTypeChange
	case "question", "inquiry":
		return IssueTypeInformational
	default:
		return IssueTypeServiceRequest
	}
}

